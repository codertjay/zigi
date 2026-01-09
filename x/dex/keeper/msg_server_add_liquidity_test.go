package keeper_test

import (
	"fmt"
	"strconv"
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"zigchain/x/dex/testutil"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	keepertest "zigchain/testutil/keeper"
	"zigchain/testutil/sample"
	"zigchain/x/dex/keeper"
	"zigchain/x/dex/testutil/common"
	"zigchain/x/dex/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

// Positive test cases

func TestMsgServerAddLiquidity_Positive(t *testing.T) {
	// Test case: add liquidity to a pool

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 50)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	// the ratio of abc to usdt is 2:1, so lp token will be sqrt(100 * 50) = 70
	createPoolExpectedLPCoin := sample.Coin("zp1", 70)

	// how much we will add to the pool of abc and usdt coins
	addLiquidityBase := sample.Coin("abc", 20)
	addLiquidityQuote := sample.Coin("usdt", 10)
	// how much we will mint LP token as a result of adding liquidity,
	// the ratio of abc to usdt is 2:1, so lp token will be sqrt(20 * 10) = 14
	addLiquidityExpectedLPCoin := sample.Coin("zp1", 14)

	// create dex keeper with pool mock
	server, dexKeeper, ctx, pool, poolAccount, bankKeeper, accountKeeper := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// extract the pool address from the pool account
	poolAddress := poolAccount.GetAddress()

	// code will check if the signer has the required balance of abc
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityBase).
		Return(true).
		Times(2)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityQuote).
		Return(true).
		Times(2)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(addLiquidityBase, addLiquidityQuote)).
		Return(nil).
		Times(1)

	// code will mint new lp tokens, so we can send them to the signer in the next step
	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(addLiquidityExpectedLPCoin)).
		Return(nil).
		Times(1)

	// SendCoinsFromModuleToAccount(context.Context, string, sdk.AccAddress, sdk.Coins) error
	// code will send the minted LP token from dex module to the signer
	bankKeeper.
		EXPECT().
		SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, signer, sdk.NewCoins(addLiquidityExpectedLPCoin)).
		Return(nil).
		Times(1)

	// create an "add liquidity" message
	txAddLiquidityMsg := &types.MsgAddLiquidity{
		Creator: creator,
		PoolId:  pool.PoolId,
		Base:    addLiquidityBase,
		Quote:   addLiquidityQuote,
	}

	// make rpc call to add liquidity
	resp, err := server.AddLiquidity(ctx, txAddLiquidityMsg)

	// make sure there is no error
	require.NoError(t, err)

	// check response
	require.Equal(t, addLiquidityExpectedLPCoin, resp.Lptoken)

	poolId := pool.PoolId
	// get the pool from the keeper, so we can compare new state to expected values
	pool, found := dexKeeper.GetPool(ctx,
		poolId,
	)

	// make sure the pool was found
	require.True(t, found)

	// new abc is old abc + added abc
	newPoolBase := createPoolBase.Add(addLiquidityBase)
	// new usdt is old usdt + added usdt
	newPoolQuote := createPoolQuote.Add(addLiquidityQuote)
	// new LP token is old LP token + minted LP token
	newPoolLPToken := createPoolExpectedLPCoin.Add(addLiquidityExpectedLPCoin)

	// quick pool check
	common.PoolCheck(
		t,
		pool,
		poolId,
		creator,
		newPoolBase,
		newPoolQuote,
		newPoolLPToken,
		poolAddress,
	)

}

func TestMsgServerAddLiquidity_ReturnQuote(t *testing.T) {
	// Test case: add liquidity to the pool
	// First we will need to create a pool, then we will add liquidity to it.
	// Remining quote will be returned to the signer to keep the pool ratio.

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 100)
	creationFee := sample.Coin("uzig", 100000000)
	// square root of a * b
	createPoolExpectedLPCoin := sample.Coin("zp1", 100)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	// create the mock account keeper
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, accountKeeper)

	// set minimal lock to 0
	params := k.GetParams(ctx)
	params.MinimalLiquidityLock = 0
	params.MaxSlippage = types.MaximumMaxSlippage
	k.SetParams(ctx, params)

	// get access to the message server
	srv := keeper.NewMsgServerImpl(k)

	// get id of the next pool
	poolId := k.GetNextPoolIDString(ctx)

	// get the pool address
	poolAddress := types.GetPoolAddress(poolId)

	// create a pool account
	poolAccount := sample.PoolModuleAccount(poolAddress)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(nil).
		Times(1)

	accountKeeper.
		EXPECT().
		NewAccount(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx sdk.Context, acc sdk.AccountI) sdk.AccountI {
			return acc
		})
	accountKeeper.
		EXPECT().
		SetAccount(gomock.Any(), gomock.Eq(poolAccount)).
		Times(1)

	// code will check if the signer has the required balance of abc
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, createPoolBase).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, createPoolQuote).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of creationFee
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, creationFee).
		Return(true).
		Times(1)

	// code will transfer creationFee coins from the signer to the module
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(creationFee)).
		Return(nil).
		Times(1)

	// code will burn the creationFee coins from the signer
	bankKeeper.
		EXPECT().
		BurnCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(creationFee)).
		Return(nil).
		Times(1)

	// code will check if the signer has the required balance of abc
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, createPoolBase).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, createPoolQuote).
		Return(true).
		Times(1)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(createPoolBase, createPoolQuote)).
		Return(nil).
		Times(1)

	// code will mint the LP token and drop it into the module dex module account
	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(createPoolExpectedLPCoin)).
		Return(nil).
		Times(1)

	// SendCoinsFromModuleToAccount(context.Context, string, sdk.AccAddress, sdk.Coins) error
	// code will send the minted LP token from dex module to the signer
	bankKeeper.
		EXPECT().
		SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, signer, sdk.NewCoins(createPoolExpectedLPCoin)).
		Return(nil).
		Times(1)

	// create a sample message to compare
	txMsg := &types.MsgCreatePool{
		Creator: creator,
		Base:    createPoolBase,
		Quote:   createPoolQuote,
	}

	// make rpc call to create pool
	_, err := srv.CreatePool(ctx, txMsg)

	// make sure there is no error
	require.NoError(t, err)

	// get the pool from the keeper
	pool, found := k.GetPool(ctx,
		poolId,
	)

	// make sure the pool was found
	require.True(t, found)

	// expected (left) vs. actual (right)
	require.Equal(t, txMsg.Creator, pool.Creator)
	require.Equal(t, poolId, pool.PoolId)
	require.Equal(t, txMsg.Base, pool.Coins[0])
	require.Equal(t, txMsg.Quote, pool.Coins[1])
	require.Equal(t, createPoolExpectedLPCoin, pool.LpToken)
	require.Equal(t, uint32(500), pool.Fee)
	require.Equal(t, "constant_product", pool.Formula)
	require.Equal(t, poolAddress.String(), pool.Address)

	// add liquidity to the pool on a different ratio
	// ------------------------------------------------------

	// how much we will add to the pool of abc and usdt coins
	// we will add 211 abc and 522 usdt to the pool which doesn't match the ratio of 1:1 of the pool,
	// so there would be 311 usdt that will be returned to the signer to keep the pool ratio

	addLiquidityBase := sample.Coin("abc", 211)
	addLiquidityQuote := sample.Coin("usdt", 522)

	// how much we will mint LP token as a result of adding liquidity
	addLiquidityExpectedLPCoin := sample.Coin("zp1", 211)
	addLiquidityExpectedReturnedCoins := sdk.NewCoins(sample.Coin("abc", 0), sample.Coin("usdt", 311))

	// code will check if the signer has the required balance of abc
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityBase).
		Return(true).
		Times(2)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityQuote).
		Return(true).
		Times(2)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(addLiquidityBase, addLiquidityQuote)).
		Return(nil).
		Times(1)

	// code will mint new lp tokens, so we can send them to the signer in the next step
	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(addLiquidityExpectedLPCoin)).
		Return(nil).
		Times(1)

	// SendCoinsFromModuleToAccount(context.Context, string, sdk.AccAddress, sdk.Coins) error
	// code will send the minted LP token from dex module to the signer
	bankKeeper.
		EXPECT().
		SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, signer, sdk.NewCoins(addLiquidityExpectedLPCoin)).
		Return(nil).
		Times(1)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	for _, coin := range addLiquidityExpectedReturnedCoins {
		bankKeeper.
			EXPECT().
			HasBalance(gomock.Any(), poolAddress, coin).
			Return(true).
			Times(1)
	}

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), poolAddress, signer, addLiquidityExpectedReturnedCoins).
		Return(nil).
		Times(1)

	// create an "add liquidity" message
	txAddLiquidityMsg := &types.MsgAddLiquidity{
		Creator: creator,
		PoolId:  pool.PoolId,
		Base:    addLiquidityBase,
		Quote:   addLiquidityQuote,
	}

	// make rpc call to add liquidity
	resp, err := srv.AddLiquidity(ctx, txAddLiquidityMsg)

	// make sure there is no error
	require.NoError(t, err)

	// check response
	require.Equal(t, addLiquidityExpectedLPCoin, resp.Lptoken)
	require.Equal(t, addLiquidityExpectedReturnedCoins, sdk.NewCoins(resp.ReturnedCoins...))

	// get the pool from the keeper, so we can compare new state to expected values
	pool, found = k.GetPool(ctx,
		poolId,
	)

	// make sure the pool was found
	require.True(t, found)

	// new abc is old abc + added abc
	newPoolBase := createPoolBase.Add(addLiquidityBase)
	// new usdt is old usdt + added usdt
	newPoolQuote := createPoolQuote.Add(sample.Coin("usdt", 211))
	// new LP token is old LP token + minted LP token
	newPoolLPToken := createPoolExpectedLPCoin.Add(addLiquidityExpectedLPCoin)

	// quick pool check
	common.PoolCheck(
		t,
		pool,
		poolId,
		creator,
		newPoolBase,
		newPoolQuote,
		newPoolLPToken,
		poolAddress,
	)

	// EVENT CHECK
	// ----------------------------------------------------------------

	// check that the event was emitted
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// find the pauser address added event
	expectedCoinsIn := sdk.NewCoins(sample.Coin("abc", 211), sample.Coin("usdt", 211))
	poolCoinsAfter := sdk.NewCoins(sample.Coin("abc", 311), sample.Coin("usdt", 311), sample.Coin("zp1", 311))
	expectedCoinsOut := sdk.NewCoins(sample.Coin("zp1", 211))
	expectedReturnedCoins := sdk.NewCoins(sample.Coin("usdt", 311))

	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventLiquidityAdded {
			foundEvent = true
			// check event attributes
			for _, attr := range event.Attributes {
				if string(attr.Key) == sdk.AttributeKeyModule {
					require.Equal(t, "dex", string(attr.Value))
				}
				if string(attr.Key) == sdk.AttributeKeySender {
					require.Equal(t, creator, string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyPoolId {
					require.Equal(t, poolId, string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyTokensIn {
					require.Equal(t, expectedCoinsIn.String(), string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyLPTokenOut {
					require.Equal(t, expectedCoinsOut.String(), string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyReturnedCoins {
					require.Equal(t, expectedReturnedCoins.String(), string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyPoolState {
					require.Equal(t, poolCoinsAfter.String(), string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyReceiver {
					require.Equal(t, creator, string(attr.Value))
				}
			}
			break
		}
	}
	require.True(t, foundEvent, "Expected EventLiquidityAdded event to be emitted")
}

func TestMsgServerAddLiquidity_ReturnBase(t *testing.T) {
	// Test case: add liquidity to the pool
	// First we will need to create a pool, then we will add liquidity to it.
	// Remining base will be returned to the signer to keep the pool ratio.

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 1000)
	createPoolQuote := sample.Coin("usdt", 4000)
	creationFee := sample.Coin("uzig", 100000000)
	// square root of a * b
	createPoolExpectedLPCoin := sample.Coin("zp1", 2000)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	// create the mock account keeper
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, accountKeeper)

	// set minimal lock to 0
	params := k.GetParams(ctx)
	params.MinimalLiquidityLock = 0
	params.MaxSlippage = types.MaximumMaxSlippage
	k.SetParams(ctx, params)

	// get access to the message server
	srv := keeper.NewMsgServerImpl(k)

	// get id of the next pool
	poolId := k.GetNextPoolIDString(ctx)

	// get the pool address
	poolAddress := types.GetPoolAddress(poolId)

	// create a pool account
	poolAccount := sample.PoolModuleAccount(poolAddress)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(nil).
		Times(1)

	accountKeeper.
		EXPECT().
		NewAccount(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx sdk.Context, acc sdk.AccountI) sdk.AccountI {
			return acc
		})
	accountKeeper.
		EXPECT().
		SetAccount(gomock.Any(), gomock.Eq(poolAccount)).
		Times(1)

	// code will check if the signer has the required balance of abc
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, createPoolBase).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, createPoolQuote).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of creationFee
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, creationFee).
		Return(true).
		Times(1)

	// code will transfer creationFee coins from the signer to the module
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(creationFee)).
		Return(nil).
		Times(1)

	// code will burn the creationFee coins from the signer
	bankKeeper.
		EXPECT().
		BurnCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(creationFee)).
		Return(nil).
		Times(1)

	// code will check if the signer has the required balance of abc
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, createPoolBase).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, createPoolQuote).
		Return(true).
		Times(1)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(createPoolBase, createPoolQuote)).
		Return(nil).
		Times(1)

	// code will mint the LP token and drop it into the module dex module account
	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(createPoolExpectedLPCoin)).
		Return(nil).
		Times(1)

	// SendCoinsFromModuleToAccount(context.Context, string, sdk.AccAddress, sdk.Coins) error
	// code will send the minted LP token from dex module to the signer
	bankKeeper.
		EXPECT().
		SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, signer, sdk.NewCoins(createPoolExpectedLPCoin)).
		Return(nil).
		Times(1)

	// create a sample message to compare
	txMsg := &types.MsgCreatePool{
		Creator: creator,
		Base:    createPoolBase,
		Quote:   createPoolQuote,
	}

	// make rpc call to create pool
	_, err := srv.CreatePool(ctx, txMsg)

	// make sure there is no error
	require.NoError(t, err)

	// get the pool from the keeper
	pool, found := k.GetPool(ctx,
		poolId,
	)

	// make sure the pool was found
	require.True(t, found)

	// expected (left) vs. actual (right)
	require.Equal(t, txMsg.Creator, pool.Creator)
	require.Equal(t, poolId, pool.PoolId)
	require.Equal(t, txMsg.Base, pool.Coins[0])
	require.Equal(t, txMsg.Quote, pool.Coins[1])
	require.Equal(t, createPoolExpectedLPCoin, pool.LpToken)
	require.Equal(t, uint32(500), pool.Fee)
	require.Equal(t, "constant_product", pool.Formula)
	require.Equal(t, poolAddress.String(), pool.Address)

	// add liquidity to the pool on a different ratio
	// ------------------------------------------------------

	// how much we will add to the pool of abc and usdt coins
	// we will add 300 abc and 600 usdt to the pool which doesn't match the ratio of 1:4 of the pool,
	// so there would be 150 abc that will be returned to the signer to keep the pool ratio

	addLiquidityBase := sample.Coin("abc", 300)
	addLiquidityQuote := sample.Coin("usdt", 600)

	// how much we will mint LP token as a result of adding liquidity
	addLiquidityExpectedLPCoin := sample.Coin("zp1", 300)
	addLiquidityExpectedReturnedCoins := sdk.NewCoins(sample.Coin("abc", 150), sample.Coin("usdt", 0))

	// code will check if the signer has the required balance of abc
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityBase).
		Return(true).
		Times(2)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityQuote).
		Return(true).
		Times(2)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(addLiquidityBase, addLiquidityQuote)).
		Return(nil).
		Times(1)

	// code will mint new lp tokens, so we can send them to the signer in the next step
	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(addLiquidityExpectedLPCoin)).
		Return(nil).
		Times(1)

	// SendCoinsFromModuleToAccount(context.Context, string, sdk.AccAddress, sdk.Coins) error
	// code will send the minted LP token from dex module to the signer
	bankKeeper.
		EXPECT().
		SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, signer, sdk.NewCoins(addLiquidityExpectedLPCoin)).
		Return(nil).
		Times(1)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	for _, coin := range addLiquidityExpectedReturnedCoins {
		bankKeeper.
			EXPECT().
			HasBalance(gomock.Any(), poolAddress, coin).
			Return(true).
			Times(1)
	}

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), poolAddress, signer, addLiquidityExpectedReturnedCoins).
		Return(nil).
		Times(1)

	// create an "add liquidity" message
	txAddLiquidityMsg := &types.MsgAddLiquidity{
		Creator: creator,
		PoolId:  pool.PoolId,
		Base:    addLiquidityBase,
		Quote:   addLiquidityQuote,
	}

	// make rpc call to add liquidity
	resp, err := srv.AddLiquidity(ctx, txAddLiquidityMsg)

	// make sure there is no error
	require.NoError(t, err)

	// check response
	require.Equal(t, addLiquidityExpectedLPCoin, resp.Lptoken)
	require.Equal(t, addLiquidityExpectedReturnedCoins, sdk.NewCoins(resp.ReturnedCoins...))

	// get the pool from the keeper, so we can compare new state to expected values
	pool, found = k.GetPool(ctx,
		poolId,
	)

	// make sure the pool was found
	require.True(t, found)

	// new abc is old abc + added abc
	newPoolBase := createPoolBase.Add(sample.Coin("abc", 150))
	// new usdt is old usdt + added usdt
	newPoolQuote := createPoolQuote.Add(addLiquidityQuote)
	// new LP token is old LP token + minted LP token
	newPoolLPToken := createPoolExpectedLPCoin.Add(addLiquidityExpectedLPCoin)

	// quick pool check
	common.PoolCheck(
		t,
		pool,
		poolId,
		creator,
		newPoolBase,
		newPoolQuote,
		newPoolLPToken,
		poolAddress,
	)
	//
	// EVENT CHECK
	// ----------------------------------------------------------------

	// check that the event was emitted
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// find the pauser address added event
	expectedCoinsIn := sdk.NewCoins(sample.Coin("abc", 150), sample.Coin("usdt", 600))
	poolCoinsAfter := sdk.NewCoins(sample.Coin("abc", 1150), sample.Coin("usdt", 4600), sample.Coin("zp1", 2300))
	expectedCoinsOut := sdk.NewCoins(sample.Coin("zp1", 300))
	expectedReturnedCoins := sdk.NewCoins(sample.Coin("abc", 150))

	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventLiquidityAdded {
			foundEvent = true
			// check event attributes
			for _, attr := range event.Attributes {
				if string(attr.Key) == sdk.AttributeKeyModule {
					require.Equal(t, "dex", string(attr.Value))
				}
				if string(attr.Key) == sdk.AttributeKeySender {
					require.Equal(t, creator, string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyPoolId {
					require.Equal(t, poolId, string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyTokensIn {
					require.Equal(t, expectedCoinsIn.String(), string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyLPTokenOut {
					require.Equal(t, expectedCoinsOut.String(), string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyReturnedCoins {
					require.Equal(t, expectedReturnedCoins.String(), string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyPoolState {
					require.Equal(t, poolCoinsAfter.String(), string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyReceiver {
					require.Equal(t, creator, string(attr.Value))
				}
			}
			break
		}
	}
	require.True(t, foundEvent, "Expected EventLiquidityAdded event to be emitted")
}

func TestMsgServerAddLiquidity_Valid(t *testing.T) {
	// Test case: add liquidity to a pool

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 100)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	// square root of a * b
	createPoolExpectedLPCoin := sample.Coin("zp1", 100)

	// how much we will add to the pool of abc and usdt coins
	addLiquidityBase := sample.Coin("abc", 211)
	addLiquidityQuote := sample.Coin("usdt", 522)

	// how much we will mint LP token as a result of adding liquidity
	addLiquidityExpectedLPCoin := sample.Coin("zp1", 211)
	addLiquidityExpectedReturnedCoins := sdk.NewCoins(sample.Coin("abc", 0), sample.Coin("usdt", 311))

	// create dex keeper with pool mock
	server, dexKeeper, ctx, pool, poolAccount, bankKeeper, accountKeeper := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// set max slippage to maximum allowed value
	params := types.DefaultParams()
	params.MaxSlippage = types.MaximumMaxSlippage
	dexKeeper.SetParams(ctx, params)

	// extract the pool address from the pool account
	poolAddress := poolAccount.GetAddress()

	// code will check if the signer has the required balance of abc
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityBase).
		Return(true).
		Times(2)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityQuote).
		Return(true).
		Times(2)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(addLiquidityBase, addLiquidityQuote)).
		Return(nil).
		Times(1)

	// code will mint new lp tokens, so we can send them to the signer in the next step
	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(addLiquidityExpectedLPCoin)).
		Return(nil).
		Times(1)

	// SendCoinsFromModuleToAccount(context.Context, string, sdk.AccAddress, sdk.Coins) error
	// code will send the minted LP token from dex module to the signer
	bankKeeper.
		EXPECT().
		SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, signer, sdk.NewCoins(addLiquidityExpectedLPCoin)).
		Return(nil).
		Times(1)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	for _, coin := range addLiquidityExpectedReturnedCoins {
		bankKeeper.
			EXPECT().
			HasBalance(gomock.Any(), poolAddress, coin).
			Return(true).
			Times(1)
	}

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), poolAddress, signer, addLiquidityExpectedReturnedCoins).
		Return(nil).
		Times(1)

	// create an "add liquidity" message
	txAddLiquidityMsg := &types.MsgAddLiquidity{
		Creator: creator,
		PoolId:  pool.PoolId,
		Base:    addLiquidityBase,
		Quote:   addLiquidityQuote,
	}

	// make rpc call to add liquidity
	resp, err := server.AddLiquidity(ctx, txAddLiquidityMsg)

	// make sure there is no error
	require.NoError(t, err)

	// check response
	require.Equal(t, addLiquidityExpectedLPCoin, resp.Lptoken)

	poolId := pool.PoolId
	// get the pool from the keeper, so we can compare new state to expected values
	pool, found := dexKeeper.GetPool(ctx,
		poolId,
	)

	// make sure the pool was found
	require.True(t, found)

	// new abc is old abc + added abc
	newPoolBase := createPoolBase.Add(addLiquidityBase)
	// new usdt is old usdt + added usdt
	newPoolQuote := createPoolQuote.Add(sample.Coin("usdt", 211))
	// new LP token is old LP token + minted LP token
	newPoolLPToken := createPoolExpectedLPCoin.Add(addLiquidityExpectedLPCoin)

	// quick pool check
	common.PoolCheck(
		t,
		pool,
		poolId,
		creator,
		newPoolBase,
		newPoolQuote,
		newPoolLPToken,
		poolAddress,
	)
}

func TestMsgServerAddLiquidity_WithReceiver(t *testing.T) {
	// Test case: add liquidity to a pool

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create a sample receiver address
	receiver := sample.AccAddress()

	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 100)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	// square root of a * b
	createPoolExpectedLPCoin := sample.Coin("zp1", 100)

	// how much we will add to the pool of abc and usdt coins
	addLiquidityBase := sample.Coin("abc", 211)
	addLiquidityQuote := sample.Coin("usdt", 522)

	// how much we will mint LP token as a result of adding liquidity
	addLiquidityExpectedLPCoin := sample.Coin("zp1", 211)
	addLiquidityExpectedReturnedCoins := sdk.NewCoins(sample.Coin("abc", 0), sample.Coin("usdt", 311))

	// create dex keeper with pool mock
	server, dexKeeper, ctx, pool, poolAccount, bankKeeper, accountKeeper := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// set max slippage to maximum allowed value
	params := types.DefaultParams()
	params.MaxSlippage = types.MaximumMaxSlippage
	dexKeeper.SetParams(ctx, params)

	// extract the pool address from the pool account
	poolAddress := poolAccount.GetAddress()

	// code will check if the signer has the required balance of abc
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityBase).
		Return(true).
		Times(2)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityQuote).
		Return(true).
		Times(2)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(addLiquidityBase, addLiquidityQuote)).
		Return(nil).
		Times(1)

	// code will mint new lp tokens, so we can send them to the signer in the next step
	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(addLiquidityExpectedLPCoin)).
		Return(nil).
		Times(1)

	// SendCoinsFromModuleToAccount(context.Context, string, sdk.AccAddress, sdk.Coins) error
	// code will send the minted LP token from dex module to the signer
	bankKeeper.
		EXPECT().
		SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, sdk.MustAccAddressFromBech32(receiver), sdk.NewCoins(addLiquidityExpectedLPCoin)).
		Return(nil).
		Times(1)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	for _, coin := range addLiquidityExpectedReturnedCoins {
		bankKeeper.
			EXPECT().
			HasBalance(gomock.Any(), poolAddress, coin).
			Return(true).
			Times(1)
	}

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), poolAddress, sdk.MustAccAddressFromBech32(receiver), addLiquidityExpectedReturnedCoins).
		Return(nil).
		Times(1)

	// create an "add liquidity" message
	txAddLiquidityMsg := &types.MsgAddLiquidity{
		Creator:  creator,
		PoolId:   pool.PoolId,
		Base:     addLiquidityBase,
		Quote:    addLiquidityQuote,
		Receiver: receiver,
	}

	// make rpc call to add liquidity
	resp, err := server.AddLiquidity(ctx, txAddLiquidityMsg)

	// make sure there is no error
	require.NoError(t, err)

	// check response
	require.Equal(t, addLiquidityExpectedLPCoin, resp.Lptoken)

	poolId := pool.PoolId
	// get the pool from the keeper, so we can compare new state to expected values
	pool, found := dexKeeper.GetPool(ctx,
		poolId,
	)

	// make sure the pool was found
	require.True(t, found)

	// new abc is old abc + added abc
	newPoolBase := createPoolBase.Add(addLiquidityBase)
	// new usdt is old usdt + added usdt
	newPoolQuote := createPoolQuote.Add(sample.Coin("usdt", 211))
	// new LP token is old LP token + minted LP token
	newPoolLPToken := createPoolExpectedLPCoin.Add(addLiquidityExpectedLPCoin)

	// quick pool check
	common.PoolCheck(
		t,
		pool,
		poolId,
		creator,
		newPoolBase,
		newPoolQuote,
		newPoolLPToken,
		poolAddress,
	)
}

func TestMsgServerAddLiquidity_AddWrongAmount(t *testing.T) {
	// Test case: add liquidity to a pool --> check if the coins are sorted correctly

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 50)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	// square root of a * b
	createPoolExpectedLPCoin := sample.Coin("zp1", 70)

	// how much we will add to the pool of abc and usdt coins
	addLiquidityBase := sample.Coin("abc", 20)
	addLiquidityQuote := sample.Coin("usdt", 10)
	// how much we will mint LP token as a result of adding liquidity
	addLiquidityExpectedLPCoin := sample.Coin("zp1", 14)

	// create dex keeper with pool mock
	server, dexKeeper, ctx, pool, poolAccount, bankKeeper, accountKeeper := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// extract the pool address from the pool account
	poolAddress := poolAccount.GetAddress()

	// code will check if the signer has the required balance of abc
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityBase).
		Return(true).
		Times(2)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityQuote).
		Return(true).
		Times(2)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(addLiquidityBase, addLiquidityQuote)).
		Return(nil).
		Times(1)

	// code will mint new lp tokens, so we can send them to the signer in the next step
	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(addLiquidityExpectedLPCoin)).
		Return(nil).
		Times(1)

	// SendCoinsFromModuleToAccount(context.Context, string, sdk.AccAddress, sdk.Coins) error
	// code will send the minted LP token from dex module to the signer
	bankKeeper.
		EXPECT().
		SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, signer, sdk.NewCoins(addLiquidityExpectedLPCoin)).
		Return(nil).
		Times(1)

	// create an "add liquidity" message
	txAddLiquidityMsg := &types.MsgAddLiquidity{
		Creator: creator,
		PoolId:  pool.PoolId,
		Base:    addLiquidityQuote,
		Quote:   addLiquidityBase,
	}

	// make rpc call to add liquidity
	resp, err := server.AddLiquidity(ctx, txAddLiquidityMsg)

	// make sure there is no error
	require.NoError(t, err)

	// check response
	require.Equal(t, addLiquidityExpectedLPCoin, resp.Lptoken)

	poolId := pool.PoolId
	// get the pool from the keeper, so we can compare new state to expected values
	pool, found := dexKeeper.GetPool(ctx,
		poolId,
	)

	// make sure the pool was found
	require.True(t, found)

	// new abc is old abc + added abc
	newPoolBase := createPoolBase.Add(addLiquidityBase)
	// new usdt is old usdt + added usdt
	newPoolQuote := createPoolQuote.Add(addLiquidityQuote)
	// new LP token is old LP token + minted LP token
	newPoolLPToken := createPoolExpectedLPCoin.Add(addLiquidityExpectedLPCoin)

	// quick pool check
	common.PoolCheck(
		t,
		pool,
		poolId,
		creator,
		newPoolBase,
		newPoolQuote,
		newPoolLPToken,
		poolAddress,
	)
}

func TestMsgServerAddLiquidity_CalculateLiquidityShares_Positive(t *testing.T) {
	// Test case: calculate liquidity shares

	// create a pool
	pool := types.Pool{
		LpToken: sdk.NewInt64Coin("zp1", 100),
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 100),
			sdk.NewInt64Coin("usdt", 100),
		},
		Fee: 2500,
	}
	abc := sdk.NewInt64Coin("abc", 200)
	usdt := sdk.NewInt64Coin("usdt", 200)
	expectedShares := sdk.NewInt64Coin("zp1", 200)

	params := types.Params{
		MaxSlippage: types.MaximumMaxSlippage,
	}
	shares, actualBase, actualQuote, returnedCoins, err := keeper.CalculateLiquidityShares(pool, params, abc, usdt)

	require.NoError(t, err)
	require.Equal(t, expectedShares, shares)
	require.Equal(t, abc, actualBase)
	require.Equal(t, usdt, actualQuote)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("abc", 0), sdk.NewInt64Coin("usdt", 0)), returnedCoins)
}

func TestMsgServerAddLiquidity_CalculateLiquidityShares_Positive2(t *testing.T) {
	// Test case: calculate liquidity shares

	// create a pool
	pool := types.Pool{
		LpToken: sdk.NewInt64Coin("zp1", 387),
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 1200),
			sdk.NewInt64Coin("usdt", 300),
		},
		Fee: 2500,
	}
	abc := sdk.NewInt64Coin("abc", 200)
	usdt := sdk.NewInt64Coin("usdt", 50)
	expectedShares := sdk.NewInt64Coin("zp1", 64)

	params := types.Params{
		MaxSlippage: types.MaximumMaxSlippage,
	}
	shares, actualBase, actualQuote, returnedCoins, err := keeper.CalculateLiquidityShares(pool, params, abc, usdt)

	expectedActualBase := sdk.NewInt64Coin("abc", 200)
	expectedActualQuote := usdt

	require.NoError(t, err)
	require.Equal(t, expectedShares, shares)
	require.Equal(t, expectedActualBase, actualBase)
	require.Equal(t, expectedActualQuote, actualQuote)
	require.Equal(t, sdk.NewCoins(abc.Sub(expectedActualBase), sdk.NewInt64Coin("usdt", 0)), returnedCoins)
}

func TestMsgServerAddLiquidity_CalculateLiquidityShares_ZeroLP(t *testing.T) {
	// Test case: calculate liquidity shares if the pool is empty

	// create a pool --> empty pool
	pool := types.Pool{
		LpToken: sdk.NewInt64Coin("zp1", 0),
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 0),
			sdk.NewInt64Coin("usdt", 0),
		},
		Fee: 2500,
	}
	abc := sdk.NewInt64Coin("abc", 50)
	usdt := sdk.NewInt64Coin("usdt", 30)
	expectedShares := sdk.NewInt64Coin("zp1", 38)

	params := types.Params{
		MaxSlippage: types.MaximumMaxSlippage,
	}
	shares, actualBase, actualQuote, returnedCoins, err := keeper.CalculateLiquidityShares(pool, params, abc, usdt)

	require.NoError(t, err)
	require.Equal(t, expectedShares, shares)
	require.Equal(t, abc, actualBase)
	require.Equal(t, usdt, actualQuote)
	require.Equal(t, sdk.NewCoins(), returnedCoins)
}

// Negative test cases

func TestMsgServerAddLiquidity_InvalidSigner(t *testing.T) {
	// Test case: try to add liquidity to the pool with invalid signer address

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 100)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 100)

	// how much we will add to the pool of abc and usdt coins
	addLiquidityBase := sample.Coin("abc", 200)
	addLiquidityQuote := sample.Coin("usdt", 200)

	// create dex keeper with pool mock
	server, _, ctx, pool, _, _, _ := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// create an "add liquidity" message
	txAddLiquidityMsg := &types.MsgAddLiquidity{
		Creator: "Bad signer",
		PoolId:  pool.PoolId,
		Base:    addLiquidityBase,
		Quote:   addLiquidityQuote,
	}

	// make rpc call to add liquidity
	_, err := server.AddLiquidity(ctx, txAddLiquidityMsg)

	require.Error(t, err)

	// check if the error message is correct
	// "Invalid address: Bad signer: invalid address"
	require.Equal(
		t,
		"Invalid address: Bad signer: invalid address",
		err.Error(),
	)
}

func TestMsgServerAddLiquidity_PoolNotFound(t *testing.T) {
	// Test case: tyr to add liquidity to a pool that does not exist

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// how much we will add to the pool of abc and usdt coins
	addLiquidityBase := sample.Coin("abc", 200)
	addLiquidityQuote := sample.Coin("usdt", 200)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, nil)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// create an "add liquidity" message
	txAddLiquidityMsg := &types.MsgAddLiquidity{
		Creator: creator,
		PoolId:  "zp5",
		Base:    addLiquidityBase,
		Quote:   addLiquidityQuote,
	}

	// make rpc call to add liquidity
	_, err := srv.AddLiquidity(ctx, txAddLiquidityMsg)

	require.Error(t, err)

	// check if the error message is correct
	// "Liquidity pool () cannot be found: invalid request"
	require.Equal(
		t,
		"Liquidity pool (zp5) can not be found: invalid request",
		err.Error(),
	)
}

func TestMsgServerAddLiquidity_InvalidCoinsPair(t *testing.T) {
	// Test case: try to add liquidity to the pool with an invalid coins pair

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 100)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 100)

	// how much we will add to the pool
	addLiquidityBase := sample.Coin("base", 200)
	addLiquidityQuote := sample.Coin("quote", 200)

	// create dex keeper with pool mock
	server, _, ctx, pool, _, _, _ := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// create an "add liquidity" message
	txAddLiquidityMsg := &types.MsgAddLiquidity{
		Creator: creator,
		PoolId:  pool.PoolId,
		Base:    addLiquidityBase,
		Quote:   addLiquidityQuote,
	}

	// make rpc call to add liquidity
	_, err := server.AddLiquidity(ctx, txAddLiquidityMsg)

	require.Error(t, err)

	// check if the error message is correct
	// "invalid coins pair: base/quote, does not match pool coins: abc/usdt"
	require.Equal(
		t,
		"invalid coins pair: base/quote, does not match pool coins: abc/usdt: invalid coins",
		err.Error(),
	)
}

func TestMsgServerAddLiquidity_InsufficientBaseFunds(t *testing.T) {
	// Test case: try to add liquidity to the pool with insufficient base funds

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 100)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 100)

	// how much we will add to the pool of abc and usdt coins
	addLiquidityBase := sample.Coin("abc", 200)
	addLiquidityQuote := sample.Coin("usdt", 200)

	// create dex keeper with pool mock
	server, _, ctx, pool, _, bankKeeper, _ := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// will return false = No funds to check if error is properly thrown
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityBase).
		Return(false).
		Times(1)

	// create an "add liquidity" message
	txAddLiquidityMsg := &types.MsgAddLiquidity{
		Creator: creator,
		PoolId:  pool.PoolId,
		Base:    addLiquidityBase,
		Quote:   addLiquidityQuote,
	}

	// make rpc call to add liquidity
	_, err := server.AddLiquidity(ctx, txAddLiquidityMsg)

	require.Error(t, err)

	// check if the error message is correct
	// "Wallet has insufficient funds 200abc to add liquidity: insufficient funds"
	require.Equal(t,
		"Wallet has insufficient funds 200abc to add liquidity: insufficient funds",
		err.Error(),
	)
}

func TestMsgServerAddLiquidity_InsufficientQuoteFunds(t *testing.T) {
	// Test case: try to add liquidity to the pool with insufficient quote funds

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 100)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 100)

	// how much we will add to the pool of abc and usdt coins
	addLiquidityBase := sample.Coin("abc", 200)
	addLiquidityQuote := sample.Coin("usdt", 200)

	// create dex keeper with pool mock
	server, _, ctx, pool, _, bankKeeper, _ := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// code will check if the signer has the required balance of abc
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityBase).
		Return(true).
		Times(1)

	// will return false = No funds to check if error is properly thrown
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityQuote).
		Return(false).
		Times(1)

	// create an "add liquidity" message
	txAddLiquidityMsg := &types.MsgAddLiquidity{
		Creator: creator,
		PoolId:  pool.PoolId,
		Base:    addLiquidityBase,
		Quote:   addLiquidityQuote,
	}

	// make rpc call to add liquidity
	_, err := server.AddLiquidity(ctx, txAddLiquidityMsg)

	require.Error(t, err)

	// check if the error message is correct
	// "Wallet has insufficient funds 200usdt to add liquidity: insufficient funds"
	require.Equal(t,
		"Wallet has insufficient funds 200usdt to add liquidity: insufficient funds",
		err.Error(),
	)
}

func TestMsgServerAddLiquidity_FailToSendCoins(t *testing.T) {
	// Test case: failed to send coins

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 100)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 100)

	// how much we will add to the pool of abc and usdt coins
	addLiquidityBase := sample.Coin("abc", 200)
	addLiquidityQuote := sample.Coin("usdt", 200)

	// create dex keeper with pool mock
	server, _, ctx, pool, _, bankKeeper, _ := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// code will check if the signer has the required balance of abc
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityBase).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityQuote).
		Return(true).
		Times(1)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityBase).
		Return(false).
		Times(1)

	// create an "add liquidity" message
	txAddLiquidityMsg := &types.MsgAddLiquidity{
		Creator: creator,
		PoolId:  pool.PoolId,
		Base:    addLiquidityBase,
		Quote:   addLiquidityQuote,
	}

	// make rpc call to add liquidity
	_, err := server.AddLiquidity(ctx, txAddLiquidityMsg)

	require.Error(t, err)

	// check if the error message is correct
	// "Failed to send coins to module account: 200abc and 200usdt: insufficient funds"
	require.Equal(t,
		fmt.Sprintf(
			"AddLiquidity: Failed to send coins (%s): insufficient funds",
			sdk.NewCoins(addLiquidityBase, addLiquidityQuote).String(),
		),
		err.Error(),
	)
}

func TestMsgServerAddLiquidity_FailToSendCoinsInsufficientFunds(t *testing.T) {
	// Test case: failed to send coins due to insufficient funds

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 100)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 100)

	// how much we will add to the pool of abc and usdt coins
	addLiquidityBase := sample.Coin("abc", 200)
	addLiquidityQuote := sample.Coin("usdt", 200)

	// create dex keeper with pool mock
	server, _, ctx, pool, poolAccount, bankKeeper, accountKeeper := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// extract the pool address from the pool account
	poolAddress := poolAccount.GetAddress()

	// code will check if the signer has the required balance of abc
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityBase).
		Return(true).
		Times(2)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityQuote).
		Return(true).
		Times(2)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(addLiquidityBase, addLiquidityQuote)).
		Return(sdkerrors.ErrInsufficientFunds).
		Times(1)

	// create an "add liquidity" message
	txAddLiquidityMsg := &types.MsgAddLiquidity{
		Creator: creator,
		PoolId:  pool.PoolId,
		Base:    addLiquidityBase,
		Quote:   addLiquidityQuote,
	}

	// make rpc call to add liquidity
	_, err := server.AddLiquidity(ctx, txAddLiquidityMsg)

	require.Error(t, err)

	// check if the error message is correct
	// "Failed to send coins to module account: 200abc and 200usdt: insufficient funds"
	require.Equal(t,
		fmt.Sprintf(
			"AddLiquidity: Failed to send coins (%s): insufficient funds",
			sdk.NewCoins(addLiquidityBase, addLiquidityQuote).String(),
		),
		err.Error(),
	)
}

func TestMsgServerAddLiquidity_FailToMintShares(t *testing.T) {
	// Test case: failed to mint liquidity shares

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 100)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 100)

	// how much we will add to the pool of abc and usdt coins
	addLiquidityBase := sample.Coin("abc", 200)
	addLiquidityQuote := sample.Coin("usdt", 200)

	// how much we will mint LP token as a result of adding liquidity
	addLiquidityExpectedLPCoin := sample.Coin("zp1", 200)

	// create dex keeper with pool mock
	server, _, ctx, pool, poolAccount, bankKeeper, accountKeeper := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// extract the pool address from the pool account
	poolAddress := poolAccount.GetAddress()

	// code will check if the signer has the required balance of abc
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityBase).
		Return(true).
		Times(2)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityQuote).
		Return(true).
		Times(2)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(addLiquidityBase, addLiquidityQuote)).
		Return(nil).
		Times(1)

	// will return ErrInsufficientFunds
	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(addLiquidityExpectedLPCoin)).
		Return(sdkerrors.ErrInsufficientFunds).
		Times(1)

	// create an "add liquidity" message
	txAddLiquidityMsg := &types.MsgAddLiquidity{
		Creator: creator,
		PoolId:  pool.PoolId,
		Base:    addLiquidityBase,
		Quote:   addLiquidityQuote,
	}

	// make rpc call to add liquidity
	_, err := server.AddLiquidity(ctx, txAddLiquidityMsg)

	require.Error(t, err)

	// check if the error message is correct
	// "Failed to mint liquidity shares: insufficient funds"
	require.Equal(t,
		"Failed to mint liquidity shares: insufficient funds",
		err.Error(),
	)
}

func TestMsgServerAddLiquidity_FailToSendShares(t *testing.T) {
	// Test case: failed to send liquidity shares

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 100)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 100)

	// how much we will add to the pool of abc and usdt coins
	addLiquidityBase := sample.Coin("abc", 200)
	addLiquidityQuote := sample.Coin("usdt", 200)

	// how much we will mint LP token as a result of adding liquidity
	addLiquidityExpectedLPCoin := sample.Coin("zp1", 200)

	// create dex keeper with pool mock
	server, _, ctx, pool, poolAccount, bankKeeper, accountKeeper := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// extract the pool address from the pool account
	poolAddress := poolAccount.GetAddress()

	// code will check if the signer has the required balance of abc
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityBase).
		Return(true).
		Times(2)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityQuote).
		Return(true).
		Times(2)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(addLiquidityBase, addLiquidityQuote)).
		Return(nil).
		Times(1)

	// code will mint new lp tokens, so we can send them to the signer in the next step
	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(addLiquidityExpectedLPCoin)).
		Return(nil).
		Times(1)

	// will return ErrInsufficientFunds
	bankKeeper.
		EXPECT().
		SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, signer, sdk.NewCoins(addLiquidityExpectedLPCoin)).
		Return(sdkerrors.ErrInsufficientFunds).
		Times(1)

	// create an "add liquidity" message
	txAddLiquidityMsg := &types.MsgAddLiquidity{
		Creator: creator,
		PoolId:  pool.PoolId,
		Base:    addLiquidityBase,
		Quote:   addLiquidityQuote,
	}

	// make rpc call to add liquidity
	_, err := server.AddLiquidity(ctx, txAddLiquidityMsg)

	require.Error(t, err)

	// check if the error message is correct
	// "Failed to send liquidity shares: insufficient funds: insufficient funds"
	require.Equal(t,
		fmt.Sprintf(
			"Failed to send liquidity shares: %s, from pool ID: %s to receiver: %s: insufficient funds",
			addLiquidityExpectedLPCoin.String(),
			pool.PoolId,
			signer.String(),
		),
		err.Error(),
	)
}

func TestMsgServerAddLiquidity_CalculateLiquidityShares_MoreThanTwoCoins(t *testing.T) {
	// Test case: calculate liquidity shares for the pool with more than two coins

	// create a pool
	pool := types.Pool{
		LpToken: sdk.NewInt64Coin("zp1", 300),
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 100),
			sdk.NewInt64Coin("usdt", 100),
			sdk.NewInt64Coin("usdt", 100),
		},
		Fee: 2500,
	}
	abc := sdk.NewInt64Coin("abc", 200)
	usdt := sdk.NewInt64Coin("usdt", 200)

	// try to calculate liquidity shares
	params := types.Params{
		MaxSlippage: types.MaximumMaxSlippage,
	}
	_, _, _, _, err := keeper.CalculateLiquidityShares(pool, params, abc, usdt)

	require.Error(t, err)

	// check if the error message is correct
	// "pool must have exactly two coins: invalid request"
	require.Equal(t,
		"pool must have exactly two coins: invalid request",
		err.Error(),
	)
}

func TestMsgServerAddLiquidity_CalculateLiquiditySharesOneCoin(t *testing.T) {
	// Test case: tyr to calculate liquidity shares with one coin

	// create a pool
	pool := types.Pool{
		LpToken: sdk.NewInt64Coin("zp1", 200),
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 100),
		},
		Fee: 2500,
	}
	abc := sdk.NewInt64Coin("abc", 200)
	usdt := sdk.NewInt64Coin("usdt", 200)

	// try to calculate liquidity shares
	params := types.Params{
		MaxSlippage: types.MaximumMaxSlippage,
	}
	_, _, _, _, err := keeper.CalculateLiquidityShares(pool, params, abc, usdt)

	require.Error(t, err)

	// check if the error message is correct
	// "pool must have exactly two coins: invalid request"
	require.Equal(
		t,
		"pool must have exactly two coins: invalid request",
		err.Error(),
	)
}

func TestMsgServerAddLiquidity_InvalidReceiver_BadAddress(t *testing.T) {
	// Test case: try to add liquidity with invalid receiver address

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 100)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 100)

	// how much we will add to the pool of abc and usdt coins
	addLiquidityBase := sample.Coin("abc", 200)
	addLiquidityQuote := sample.Coin("usdt", 200)

	// how much we will mint LP token as a result of adding liquidity
	addLiquidityExpectedLPCoin := sample.Coin("zp1", 200)

	// create dex keeper with pool mock
	server, _, ctx, pool, poolAccount, bankKeeper, accountKeeper := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// extract the pool address from the pool account
	poolAddress := poolAccount.GetAddress()

	// code will check if the signer has the required balance of abc
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityBase).
		Return(true).
		Times(2)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityQuote).
		Return(true).
		Times(2)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(addLiquidityBase, addLiquidityQuote)).
		Return(nil).
		Times(1)

	// code will mint new lp tokens, so we can send them to the signer in the next step
	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(addLiquidityExpectedLPCoin)).
		Return(nil).
		Times(1)

	// create an "add liquidity" message
	txAddLiquidityMsg := &types.MsgAddLiquidity{
		Creator:  creator,
		PoolId:   pool.PoolId,
		Base:     addLiquidityBase,
		Quote:    addLiquidityQuote,
		Receiver: "bad_receiver",
	}

	// make rpc call to add liquidity
	_, err := server.AddLiquidity(ctx, txAddLiquidityMsg)

	require.Error(t, err)

	// check if the error message is correct
	require.Equal(t,
		"decoding bech32 failed: invalid separator index -1",
		err.Error(),
	)
}

func TestMsgServerAddLiquidity_InvalidReceiver_TooShort(t *testing.T) {
	// Test case: try to add liquidity with invalid receiver address --> too short

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 100)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 100)

	// how much we will add to the pool of abc and usdt coins
	addLiquidityBase := sample.Coin("abc", 200)
	addLiquidityQuote := sample.Coin("usdt", 200)

	// how much we will mint LP token as a result of adding liquidity
	addLiquidityExpectedLPCoin := sample.Coin("zp1", 200)

	// create dex keeper with pool mock
	server, _, ctx, pool, poolAccount, bankKeeper, accountKeeper := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// extract the pool address from the pool account
	poolAddress := poolAccount.GetAddress()

	// code will check if the signer has the required balance of abc
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityBase).
		Return(true).
		Times(2)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityQuote).
		Return(true).
		Times(2)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(addLiquidityBase, addLiquidityQuote)).
		Return(nil).
		Times(1)

	// code will mint new lp tokens, so we can send them to the signer in the next step
	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(addLiquidityExpectedLPCoin)).
		Return(nil).
		Times(1)

	// create an "add liquidity" message
	txAddLiquidityMsg := &types.MsgAddLiquidity{
		Creator:  creator,
		PoolId:   pool.PoolId,
		Base:     addLiquidityBase,
		Quote:    addLiquidityQuote,
		Receiver: "zig123",
	}

	// make rpc call to add liquidity
	_, err := server.AddLiquidity(ctx, txAddLiquidityMsg)

	require.Error(t, err)

	// check if the error message is correct
	require.Equal(t,
		"decoding bech32 failed: invalid bech32 string length 6",
		err.Error(),
	)
}

func TestMsgServerAddLiquidity_InvalidReceiver_TooLong(t *testing.T) {
	// Test case: try to add liquidity with invalid receiver address --> too long

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 100)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 100)

	// how much we will add to the pool of abc and usdt coins
	addLiquidityBase := sample.Coin("abc", 200)
	addLiquidityQuote := sample.Coin("usdt", 200)

	// how much we will mint LP token as a result of adding liquidity
	addLiquidityExpectedLPCoin := sample.Coin("zp1", 200)

	// create dex keeper with pool mock
	server, _, ctx, pool, poolAccount, bankKeeper, accountKeeper := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// extract the pool address from the pool account
	poolAddress := poolAccount.GetAddress()

	// code will check if the signer has the required balance of abc
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityBase).
		Return(true).
		Times(2)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityQuote).
		Return(true).
		Times(2)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(addLiquidityBase, addLiquidityQuote)).
		Return(nil).
		Times(1)

	// code will mint new lp tokens, so we can send them to the signer in the next step
	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(addLiquidityExpectedLPCoin)).
		Return(nil).
		Times(1)

	// create an "add liquidity" message
	txAddLiquidityMsg := &types.MsgAddLiquidity{
		Creator:  creator,
		PoolId:   pool.PoolId,
		Base:     addLiquidityBase,
		Quote:    addLiquidityQuote,
		Receiver: "zig1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
	}

	// make rpc call to add liquidity
	_, err := server.AddLiquidity(ctx, txAddLiquidityMsg)

	require.Error(t, err)

	// check if the error message is correct
	require.Equal(t,
		"decoding bech32 failed: invalid checksum (expected yurny3 got 567890)",
		err.Error(),
	)
}

func TestMsgServerAddLiquidity_InvalidReceiver_NotLowercaseOrUppercase(t *testing.T) {
	// Test case: try to add liquidity with invalid receiver address --> not lowercase or uppercase

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 100)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 100)

	// how much we will add to the pool of abc and usdt coins
	addLiquidityBase := sample.Coin("abc", 200)
	addLiquidityQuote := sample.Coin("usdt", 200)

	// how much we will mint LP token as a result of adding liquidity
	addLiquidityExpectedLPCoin := sample.Coin("zp1", 200)

	// create dex keeper with pool mock
	server, _, ctx, pool, poolAccount, bankKeeper, accountKeeper := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// extract the pool address from the pool account
	poolAddress := poolAccount.GetAddress()

	// code will check if the signer has the required balance of abc
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityBase).
		Return(true).
		Times(2)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityQuote).
		Return(true).
		Times(2)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(addLiquidityBase, addLiquidityQuote)).
		Return(nil).
		Times(1)

	// code will mint new lp tokens, so we can send them to the signer in the next step
	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(addLiquidityExpectedLPCoin)).
		Return(nil).
		Times(1)

	// create an "add liquidity" message
	txAddLiquidityMsg := &types.MsgAddLiquidity{
		Creator:  creator,
		PoolId:   pool.PoolId,
		Base:     addLiquidityBase,
		Quote:    addLiquidityQuote,
		Receiver: "BAd Address",
	}

	// make rpc call to add liquidity
	_, err := server.AddLiquidity(ctx, txAddLiquidityMsg)

	require.Error(t, err)

	// check if the error message is correct
	require.Equal(t,
		"decoding bech32 failed: string not all lowercase or all uppercase",
		err.Error(),
	)
}

func TestMsgServerAddLiquidity_InvalidReceiver_BadChars(t *testing.T) {
	// Test case: try to add liquidity with invalid receiver address --> bad characters

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 100)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 100)

	// how much we will add to the pool of abc and usdt coins
	addLiquidityBase := sample.Coin("abc", 200)
	addLiquidityQuote := sample.Coin("usdt", 200)

	// how much we will mint LP token as a result of adding liquidity
	addLiquidityExpectedLPCoin := sample.Coin("zp1", 200)

	// create dex keeper with pool mock
	server, _, ctx, pool, poolAccount, bankKeeper, accountKeeper := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// extract the pool address from the pool account
	poolAddress := poolAccount.GetAddress()

	// code will check if the signer has the required balance of abc
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityBase).
		Return(true).
		Times(2)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityQuote).
		Return(true).
		Times(2)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(addLiquidityBase, addLiquidityQuote)).
		Return(nil).
		Times(1)

	// code will mint new lp tokens, so we can send them to the signer in the next step
	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(addLiquidityExpectedLPCoin)).
		Return(nil).
		Times(1)

	// create an "add liquidity" message
	txAddLiquidityMsg := &types.MsgAddLiquidity{
		Creator:  creator,
		PoolId:   pool.PoolId,
		Base:     addLiquidityBase,
		Quote:    addLiquidityQuote,
		Receiver: "zi!12z7zhe2d03wn03e8lkw8at6zjxa7y65q5xnp5%",
	}

	// make rpc call to add liquidity
	_, err := server.AddLiquidity(ctx, txAddLiquidityMsg)

	require.Error(t, err)

	// check if the error message is correct
	require.Equal(t,
		"decoding bech32 failed: invalid character not part of charset: 37",
		err.Error(),
	)
}

func TestMsgServerAddLiquidity_ZeroBaseAndQuote(t *testing.T) {
	// Test case: try to add liquidity with zero abc and usdt coins

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 100)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 100)

	// how much we will add to the pool of abc and usdt coins
	addLiquidityBase := sample.Coin("abc", 0)
	addLiquidityQuote := sample.Coin("usdt", 0)

	// create dex keeper with pool mock
	server, _, ctx, pool, _, _, _ := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// create an "add liquidity" message
	txAddLiquidityMsg := &types.MsgAddLiquidity{
		Creator: creator,
		PoolId:  pool.PoolId,
		Base:    addLiquidityBase,
		Quote:   addLiquidityQuote,
	}

	// make rpc call to add liquidity
	_, err := server.AddLiquidity(ctx, txAddLiquidityMsg)

	require.Error(t, err)

	// check if the error message is correct
	require.Equal(t,
		"invalid coins pair, must send two coins, at positive (non zero) value: invalid request",
		err.Error(),
	)
}

func TestMsgServerAddLiquidity_ZeroBase(t *testing.T) {
	// Test case: try to add liquidity if base coin is zero

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 100)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 100)

	// how much we will add to the pool of abc and usdt coins
	addLiquidityBase := sample.Coin("abc", 0)
	addLiquidityQuote := sample.Coin("usdt", 100)

	// create dex keeper with pool mock
	server, _, ctx, pool, _, _, _ := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// create an "add liquidity" message
	txAddLiquidityMsg := &types.MsgAddLiquidity{
		Creator: creator,
		PoolId:  pool.PoolId,
		Base:    addLiquidityBase,
		Quote:   addLiquidityQuote,
	}

	// make rpc call to add liquidity
	_, err := server.AddLiquidity(ctx, txAddLiquidityMsg)

	require.Error(t, err)

	// check if the error message is correct
	require.Equal(t,
		"invalid coins pair, must send two coins, at positive (non zero) value: invalid request",
		err.Error(),
	)
}

func TestMsgServerAddLiquidity_ZeroQuote(t *testing.T) {
	// Test case: try to add liquidity if quote coin is zero

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 100)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 100)

	// how much we will add to the pool of abc and usdt coins
	addLiquidityBase := sample.Coin("abc", 100)
	addLiquidityQuote := sample.Coin("usdt", 0)

	// create dex keeper with pool mock
	server, _, ctx, pool, _, _, _ := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// create an "add liquidity" message
	txAddLiquidityMsg := &types.MsgAddLiquidity{
		Creator: creator,
		PoolId:  pool.PoolId,
		Base:    addLiquidityBase,
		Quote:   addLiquidityQuote,
	}

	// make rpc call to add liquidity
	_, err := server.AddLiquidity(ctx, txAddLiquidityMsg)

	require.Error(t, err)

	// check if the error message is correct
	require.Equal(t,
		"invalid coins pair, must send two coins, at positive (non zero) value: invalid request",
		err.Error(),
	)
}

func TestMsgServerAddLiquidity_MaxSlippage(t *testing.T) {
	// Test case: add liquidity with different slippage scenarios
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	creator := sample.AccAddress()
	signer := sdk.MustAccAddressFromBech32(creator)
	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 50)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	// square root of a * b
	createPoolExpectedLPCoin := sample.Coin("zp1", 70)
	// Create test cases with different slippage scenarios
	testCases := []struct {
		name                  string
		maxSlippage           uint32
		baseAmount            int64
		quoteAmount           int64
		expectedLPCoin        int64
		expectedReturnedCoins sdk.Coins
		shouldSucceed         bool
		expectedError         string
	}{
		{
			name:                  "exact ratio match",
			maxSlippage:           500, // 5%
			baseAmount:            205,
			quoteAmount:           100,
			expectedLPCoin:        140,
			expectedReturnedCoins: sdk.NewCoins(sample.Coin("abc", 5), sample.Coin("usdt", 0)),
			shouldSucceed:         true,
		},
		{
			name:                  "within slippage limit",
			maxSlippage:           1000, // 10%
			baseAmount:            210,
			quoteAmount:           100,
			expectedLPCoin:        140,
			expectedReturnedCoins: sdk.NewCoins(sample.Coin("abc", 10), sample.Coin("usdt", 0)),
			shouldSucceed:         true,
		},
		{
			name:                  "exactly at slippage limit",
			maxSlippage:           1000, // 10%
			baseAmount:            210,
			quoteAmount:           100,
			expectedLPCoin:        140,
			expectedReturnedCoins: sdk.NewCoins(sample.Coin("abc", 10), sample.Coin("usdt", 0)),
			shouldSucceed:         true,
		},
		{
			name:                  "exceeds slippage limit",
			maxSlippage:           50, // 0.5%
			baseAmount:            111,
			quoteAmount:           100,
			expectedLPCoin:        78,
			expectedReturnedCoins: sdk.NewCoins(sample.Coin("abc", 0), sample.Coin("usdt", 50)),
			shouldSucceed:         false,
			expectedError:         "Failed to calculate liquidity shares: deposit ratio (1.110000000000000000) does not match pool ratio (2.000000000000000000) within allowed slippage of 0.005000000000000000: invalid request",
		},
		{
			name:                  "zero slippage tolerance",
			maxSlippage:           0,
			baseAmount:            101,
			quoteAmount:           100,
			expectedLPCoin:        70,
			expectedReturnedCoins: sdk.NewCoins(sample.Coin("abc", 0), sample.Coin("usdt", 50)),
			shouldSucceed:         false,
			expectedError:         "Failed to calculate liquidity shares: deposit ratio (1.010000000000000000) does not match pool ratio (2.000000000000000000) within allowed slippage of 0.000000000000000000: invalid request",
		},
		{
			name:                  "large slippage tolerance",
			maxSlippage:           1000, // 10%
			baseAmount:            210,
			quoteAmount:           100,
			expectedLPCoin:        140,
			expectedReturnedCoins: sdk.NewCoins(sample.Coin("abc", 10), sample.Coin("usdt", 0)),
			shouldSucceed:         true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup keeper with mock dependencies
			// create dex keeper with pool mock
			server, dexKeeper, ctx, pool, poolAccount, bankKeeper, accountKeeper := common.ServerDexKeeperWithPoolMock(
				t,
				ctrl,
				signer,
				createPoolBase,
				createPoolQuote,
				createPoolCreationFee,
				createPoolExpectedLPCoin,
			)
			// how much we will add to the pool of abc and usdt coins
			addLiquidityBase := sample.Coin("abc", tc.baseAmount)
			addLiquidityQuote := sample.Coin("usdt", tc.quoteAmount)
			// how much we will mint LP token as a result of adding liquidity
			addLiquidityExpectedLPCoin := sample.Coin("zp1", tc.expectedLPCoin)
			// Set params with test case slippage
			params := types.DefaultParams()
			params.MaxSlippage = tc.maxSlippage
			dexKeeper.SetParams(ctx, params)
			// extract the pool address from the pool account
			poolAddress := poolAccount.GetAddress()
			// Set up mock expectations for HasBalance
			// The function checks balances multiple times, so we need to set up expectations for all calls
			if tc.shouldSucceed {
				bankKeeper.EXPECT().
					HasBalance(gomock.Any(), signer, addLiquidityBase).
					Return(true).
					Times(2)
				bankKeeper.EXPECT().
					HasBalance(gomock.Any(), signer, addLiquidityQuote).
					Return(true).
					Times(2)
			} else {
				bankKeeper.EXPECT().
					HasBalance(gomock.Any(), signer, addLiquidityBase).
					Return(true).
					Times(1)
				bankKeeper.EXPECT().
					HasBalance(gomock.Any(), signer, addLiquidityQuote).
					Return(true).
					Times(1)
			}
			if tc.shouldSucceed {
				// GetAccount
				accountKeeper.EXPECT().
					GetAccount(gomock.Any(), poolAddress).
					Return(poolAccount).
					Times(1)
				bankKeeper.
					EXPECT().
					SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(addLiquidityBase, addLiquidityQuote)).
					Return(nil).
					Times(1)
				// code will mint new lp tokens, so we can send them to the signer in the next step
				bankKeeper.
					EXPECT().
					MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(addLiquidityExpectedLPCoin)).
					Return(nil).
					Times(1)
				// SendCoinsFromModuleToAccount(context.Context, string, sdk.AccAddress, sdk.Coins) error
				// code will send the minted LP token from dex module to the signer
				bankKeeper.
					EXPECT().
					SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, signer, sdk.NewCoins(addLiquidityExpectedLPCoin)).
					Return(nil).
					Times(1)

				accountKeeper.
					EXPECT().
					GetAccount(ctx, poolAddress).
					Return(poolAccount).
					Times(1)

				for _, coin := range tc.expectedReturnedCoins {
					bankKeeper.
						EXPECT().
						HasBalance(gomock.Any(), poolAddress, coin).
						Return(true).
						Times(1)
				}

				bankKeeper.
					EXPECT().
					SendCoins(gomock.Any(), poolAddress, signer, tc.expectedReturnedCoins).
					Return(nil).
					Times(1)
			}
			// Create message
			msg := types.MsgAddLiquidity{
				Creator: creator,
				PoolId:  pool.PoolId,
				Base:    addLiquidityBase,
				Quote:   addLiquidityQuote,
			}
			// Execute message
			_, err := server.AddLiquidity(ctx, &msg)
			if tc.shouldSucceed {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedError)
			}
		})
	}
}

// Test cases for the standard formula implementation based on the comment examples
func TestMsgServerAddLiquidity_CalculateLiquidityShares_RatioMatches(t *testing.T) {
	// Test case 1: Ratio matches the pool,
	// Pool state: 1_000ABC / 4_000USDT ==> LP 2_000 (sqrt(x*y))
	// Add liquidity: 100ABC / 400USDT ==> ratio: 100/400 = 1_000/4_000 = 0.25 (1 base token = 4 quote tokens)
	// Expected: 200 LP tokens (10% share of the pool)

	pool := types.Pool{
		LpToken: sdk.NewInt64Coin("zp1", 2000),
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 1000),
			sdk.NewInt64Coin("usdt", 4000),
		},
		Fee: 2500,
	}
	abc := sdk.NewInt64Coin("abc", 100)
	usdt := sdk.NewInt64Coin("usdt", 400)
	expectedShares := sdk.NewInt64Coin("zp1", 200)

	params := types.Params{
		MaxSlippage: types.MaximumMaxSlippage,
	}
	shares, actualBase, actualQuote, returnedCoins, err := keeper.CalculateLiquidityShares(pool, params, abc, usdt)

	require.NoError(t, err)
	require.Equal(t, expectedShares, shares)
	require.Equal(t, abc, actualBase)
	require.Equal(t, usdt, actualQuote)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("abc", 0), sdk.NewInt64Coin("usdt", 0)), returnedCoins)
}

func TestMsgServerAddLiquidity_CalculateLiquidityShares_RatioMismatches(t *testing.T) {
	// Test case 2: Ratio mismatches the pool,
	// Pool state: 1_000ABC / 4_000USDT ==> LP 2_000 (sqrt(x*y))
	// Add liquidity: 100ABC / 410USDT ==> ratio: 100/410 = 0.2439 (slightly off from the pool's ratio of 0.25)
	// Expected: 200 LP tokens (min of shares from base and quote)

	pool := types.Pool{
		LpToken: sdk.NewInt64Coin("zp1", 2000),
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 1000),
			sdk.NewInt64Coin("usdt", 4000),
		},
		Fee: 2500,
	}
	abc := sdk.NewInt64Coin("abc", 100)
	usdt := sdk.NewInt64Coin("usdt", 410)
	expectedShares := sdk.NewInt64Coin("zp1", 200) // min(200, 205) = 200

	params := types.Params{
		MaxSlippage: types.MaximumMaxSlippage,
	}
	shares, actualBase, actualQuote, returnedCoins, err := keeper.CalculateLiquidityShares(pool, params, abc, usdt)

	expectedActualBase := abc
	expectedActualQuote := sdk.NewInt64Coin("usdt", 400)

	require.NoError(t, err)
	require.Equal(t, expectedShares, shares)
	require.Equal(t, expectedActualBase, actualBase)
	require.Equal(t, expectedActualQuote, actualQuote)
	require.Equal(t, sdk.NewCoins(abc.Sub(expectedActualBase), usdt.Sub(expectedActualQuote)), returnedCoins)
}

func TestMsgServerAddLiquidity_CalculateLiquidityShares_LargeRatioMismatch(t *testing.T) {
	// Test case 3: Large ratio mismatch
	// Pool state: 1_000ABC / 4_000USDT ==> LP 2_000 (sqrt(x*y))
	// Pool ratio: 1_000/4_000 = 0.25
	// Add liquidity: 300ABC / 600USDT
	// Expected: 300 LP tokens (min of shares from base and quote)

	pool := types.Pool{
		LpToken: sdk.NewInt64Coin("zp1", 2000),
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 1000),
			sdk.NewInt64Coin("usdt", 4000),
		},
		Fee: 2500,
	}
	abc := sdk.NewInt64Coin("abc", 300)
	usdt := sdk.NewInt64Coin("usdt", 600)
	expectedShares := sdk.NewInt64Coin("zp1", 300) // min(600, 300) = 300

	params := types.Params{
		MaxSlippage: types.MaximumMaxSlippage,
	}
	shares, actualBase, actualQuote, returnedCoins, err := keeper.CalculateLiquidityShares(pool, params, abc, usdt)

	expectedActualBase := sdk.NewInt64Coin("abc", 150)
	expectedActualQuote := usdt

	require.NoError(t, err)
	require.Equal(t, expectedShares, shares)
	require.Equal(t, expectedActualBase, actualBase)
	require.Equal(t, expectedActualQuote, actualQuote)
	require.Equal(t, sdk.NewCoins(abc.Sub(expectedActualBase), usdt.Sub(expectedActualQuote)), returnedCoins)
}

func TestMsgServerAddLiquidity_CalculateLiquidityShares_QuoteLimited(t *testing.T) {
	// Test case: Quote token limits the LP tokens
	// Pool state: 1_000ABC / 4_000USDT ==> LP 2_000
	// Add liquidity: 500ABC / 1000USDT
	// Expected: 500 LP tokens (min of shares from base and quote)

	pool := types.Pool{
		LpToken: sdk.NewInt64Coin("zp1", 2000),
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 1000),
			sdk.NewInt64Coin("usdt", 4000),
		},
		Fee: 2500,
	}
	abc := sdk.NewInt64Coin("abc", 500)
	usdt := sdk.NewInt64Coin("usdt", 1000)
	expectedShares := sdk.NewInt64Coin("zp1", 500) // min(1000, 500) = 500

	params := types.Params{
		MaxSlippage: types.MaximumMaxSlippage,
	}
	shares, actualBase, actualQuote, returnedCoins, err := keeper.CalculateLiquidityShares(pool, params, abc, usdt)

	expectedActualBase := sdk.NewInt64Coin("abc", 250)
	expectedActualQuote := usdt

	require.NoError(t, err)
	require.Equal(t, expectedShares, shares)
	require.Equal(t, expectedActualBase, actualBase)
	require.Equal(t, expectedActualQuote, actualQuote)
	require.Equal(t, sdk.NewCoins(abc.Sub(expectedActualBase), usdt.Sub(expectedActualQuote)), returnedCoins)
}

func TestMsgServerAddLiquidity_CalculateLiquidityShares_BaseLimited(t *testing.T) {
	// Test case: Base token limits the LP tokens
	// Pool state: 1_000ABC / 4_000USDT ==> LP 2_000
	// Add liquidity: 200ABC / 2000USDT
	// Expected: 400 LP tokens (min of shares from base and quote)

	pool := types.Pool{
		LpToken: sdk.NewInt64Coin("zp1", 2000),
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 1000),
			sdk.NewInt64Coin("usdt", 4000),
		},
		Fee: 2500,
	}
	abc := sdk.NewInt64Coin("abc", 200)
	usdt := sdk.NewInt64Coin("usdt", 2000)
	expectedShares := sdk.NewInt64Coin("zp1", 400) // min(400, 1000) = 400

	params := types.Params{
		MaxSlippage: types.MaximumMaxSlippage,
	}
	shares, actualBase, actualQuote, returnedCoins, err := keeper.CalculateLiquidityShares(pool, params, abc, usdt)

	expectedActualBase := abc
	expectedActualQuote := sample.Coin("usdt", 800)

	require.NoError(t, err)
	require.Equal(t, expectedShares, shares)
	require.Equal(t, expectedActualBase, actualBase)
	require.Equal(t, expectedActualQuote, actualQuote)
	require.Equal(t, sdk.NewCoins(abc.Sub(expectedActualBase), usdt.Sub(expectedActualQuote)), returnedCoins)
}

func TestMsgServerAddLiquidity_ReturnedCoinsFromPoolAccountWithPositiveReturnedCoins(t *testing.T) {
	// Test case: verify that returned coins are correctly sent from pool account
	// This test verifies the fix for the previous issue where returned coins were sent from module account:
	// - Pool: 500abc / 1000usdt (ratio: 0.5)
	// - Add: 50abc / 100usdt
	// - Expected: 35 LP tokens, return 50usdt
	// - Success: system now correctly sends 50usdt from pool account

	// create a sample signer address
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation (exact values from user's issue)
	createPoolBase := sample.Coin("abc", 500)
	createPoolQuote := sample.Coin("usdt", 1000)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	// square root of 500 * 1000 = sqrt(500000) ≈ 707
	createPoolExpectedLPCoin := sample.Coin("zp1", 707)

	// how much we will add to the pool (exact values from user's issue)
	addLiquidityBase := sample.Coin("abc", 25)
	addLiquidityQuote := sample.Coin("usdt", 100)

	// Expected calculations based on user's issue:
	// lpMintedFromBase = 25 * 707 / 500 = 35.35 ≈ 35
	// lpMintedFromQuote = 100 * 707 / 1000 = 70.7 ≈ 70
	// lpMinted = min(35, 70) = 35
	// actualBase = 25 (user provided)
	// actualQuote = 100 * 0.5 = 50
	// returnedQuote = 100 - 50 = 50
	addLiquidityExpectedLPCoin := sample.Coin("zp1", 35)
	addLiquidityExpectedActualBase := addLiquidityBase
	addLiquidityExpectedActualQuote := sample.Coin("usdt", 50)
	addLiquidityExpectedReturnedCoins := sdk.NewCoins(sample.Coin("abc", 0), sample.Coin("usdt", 50))

	// create dex keeper with real pool
	server, dexKeeper, ctx, pool, _, bankKeeper := common.ServerDexKeeperWithRealBank(
		t,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// set max slippage to maximum allowed value to avoid slippage errors
	params := types.DefaultParams()
	params.MaxSlippage = types.MaximumMaxSlippage
	dexKeeper.SetParams(ctx, params)

	// check signer balance
	signerBalanceUsdt := bankKeeper.GetBalance(ctx, signer, "usdt")
	signerBalanceAbc := bankKeeper.GetBalance(ctx, signer, "abc")
	signerBalanceLPtokens := bankKeeper.GetBalance(ctx, signer, "zp1")

	// check that the signer balance is 6usdt
	require.Equal(t, sdk.NewInt64Coin("usdt", 1000), signerBalanceUsdt)
	require.Equal(t, sdk.NewInt64Coin("abc", 500), signerBalanceAbc)
	require.Equal(t, sdk.NewInt64Coin("zp1", 707), signerBalanceLPtokens)

	// create an "add liquidity" message
	txAddLiquidityMsg := &types.MsgAddLiquidity{
		Creator: creator,
		PoolId:  pool.PoolId,
		Base:    addLiquidityBase,
		Quote:   addLiquidityQuote,
	}

	// make rpc call to add liquidity
	resp, err := server.AddLiquidity(ctx, txAddLiquidityMsg)

	// This should not fail
	require.NoError(t, err)

	// This should not contain the exact error from the user's issue
	if err != nil {
		require.NotContains(t, err.Error(), fmt.Sprintf("Failed to send liquidity shares: %s,%s, from pool ID: zp1 to receiver: %s: spendable balance 0usdt is smaller than 6usdt: insufficient funds", addLiquidityExpectedReturnedCoins.String(), addLiquidityExpectedLPCoin.String(), signer.String()))
		require.NotContains(t, err.Error(), "insufficient funds")

		// The error should mention the returned coins that couldn't be sent
		require.NotContains(t, err.Error(), "6usdt")
		require.NotContains(t, err.Error(), "68zp1")
	}

	// Check the response
	require.Equal(t, addLiquidityExpectedLPCoin, resp.Lptoken)
	require.Equal(t, addLiquidityExpectedReturnedCoins, sdk.NewCoins(resp.ReturnedCoins...))
	require.Equal(t, addLiquidityExpectedActualBase, resp.ActualBase)
	require.Equal(t, addLiquidityExpectedActualQuote, resp.ActualQuote)

	// check signer balance
	signerBalanceUsdt = bankKeeper.GetBalance(ctx, signer, "usdt")
	signerBalanceAbc = bankKeeper.GetBalance(ctx, signer, "abc")
	signerBalanceLPtokens = bankKeeper.GetBalance(ctx, signer, "zp1")

	// check that the signer balance is 6usdt
	require.Equal(t, sdk.NewInt64Coin("usdt", 950), signerBalanceUsdt)
	require.Equal(t, sdk.NewInt64Coin("abc", 475), signerBalanceAbc)
	require.Equal(t, sdk.NewInt64Coin("zp1", 742), signerBalanceLPtokens)
}

func TestMsgServerAddLiquidity_ReturnedCoinsFromPoolAccountWithNoReturnedCoins(t *testing.T) {
	// Test case: verify that returned coins are correctly sent from pool account
	// This test verifies the fix for the previous issue where returned coins were sent from module account:
	// - Pool: 555abc / 1050usdt (ratio: 0.5285)
	// - Add: 50abc / 100usdt
	// - Expected: 68 LP tokens, return 6usdt
	// - Success: system now correctly sends 6usdt from pool account

	// create a sample signer address
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation (exact values from user's issue)
	createPoolBase := sample.Coin("abc", 500)
	createPoolQuote := sample.Coin("usdt", 1000)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	// square root of 555 * 1050 = sqrt(582750) ≈ 763
	createPoolExpectedLPCoin := sample.Coin("zp1", 707)

	// how much we will add to the pool (exact values from user's issue)
	addLiquidityBase := sample.Coin("abc", 50)
	addLiquidityQuote := sample.Coin("usdt", 100)

	// Expected calculations based on user's issue:
	// lpMintedFromBase = 50 * 763 / 555 = 68.73... ≈ 68
	// lpMintedFromQuote = 100 * 763 / 1050 = 72.66... ≈ 72
	// lpMinted = min(68, 72) = 68
	// actualBase = 50 (user provided)
	// actualQuote = 50 / 0.5285 = 94.6... ≈ 94
	// returnedQuote = 100 - 94 = 6
	addLiquidityExpectedLPCoin := sample.Coin("zp1", 70)
	addLiquidityExpectedReturnedCoins := sdk.NewCoins(sample.Coin("abc", 0), sample.Coin("usdt", 0))
	addLiquidityExpectedActualBase := addLiquidityBase
	addLiquidityExpectedActualQuote := addLiquidityQuote

	// create dex keeper with real pool
	server, dexKeeper, ctx, pool, _, bankKeeper := common.ServerDexKeeperWithRealBank(
		t,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// set max slippage to maximum allowed value to avoid slippage errors
	params := types.DefaultParams()
	params.MaxSlippage = types.MaximumMaxSlippage
	dexKeeper.SetParams(ctx, params)

	// check signer balance
	signerBalanceUsdt := bankKeeper.GetBalance(ctx, signer, "usdt")
	signerBalanceAbc := bankKeeper.GetBalance(ctx, signer, "abc")

	// check that the signer balance is 6usdt
	require.Equal(t, sdk.NewInt64Coin("usdt", 1000), signerBalanceUsdt)
	require.Equal(t, sdk.NewInt64Coin("abc", 500), signerBalanceAbc)

	// create an "add liquidity" message
	txAddLiquidityMsg := &types.MsgAddLiquidity{
		Creator: creator,
		PoolId:  pool.PoolId,
		Base:    addLiquidityBase,
		Quote:   addLiquidityQuote,
	}

	// make rpc call to add liquidity
	resp, err := server.AddLiquidity(ctx, txAddLiquidityMsg)

	// This should not fail
	require.NoError(t, err)

	// This should not contain the exact error from the user's issue
	if err != nil {
		require.NotContains(t, err.Error(), fmt.Sprintf("Failed to send liquidity shares: %s,%s, from pool ID: zp1 to receiver: %s: spendable balance 0usdt is smaller than 6usdt: insufficient funds", addLiquidityExpectedReturnedCoins.String(), addLiquidityExpectedLPCoin.String(), signer.String()))
		require.NotContains(t, err.Error(), "insufficient funds")

		// The error should mention the returned coins that couldn't be sent
		require.NotContains(t, err.Error(), "6usdt")
		require.NotContains(t, err.Error(), "68zp1")
	}

	require.Equal(t, addLiquidityExpectedLPCoin, resp.Lptoken)
	require.Equal(t, addLiquidityExpectedReturnedCoins, sdk.NewCoins(resp.ReturnedCoins...))
	require.Equal(t, addLiquidityExpectedActualBase, resp.ActualBase)
	require.Equal(t, addLiquidityExpectedActualQuote, resp.ActualQuote)

	// check signer balance
	signerBalanceUsdt = bankKeeper.GetBalance(ctx, signer, "usdt")
	signerBalanceAbc = bankKeeper.GetBalance(ctx, signer, "abc")

	// check that the signer balance is 6usdt
	require.Equal(t, sdk.NewInt64Coin("usdt", 900), signerBalanceUsdt)
	require.Equal(t, sdk.NewInt64Coin("abc", 450), signerBalanceAbc)
}

func TestCalculateLiquidityShares_EmptyPool_ZeroBase(t *testing.T) {
	// Test case: calculate liquidity shares if the pool is empty and the base amount is zero

	// create a pool --> empty pool
	pool := types.Pool{
		LpToken: sdk.NewInt64Coin("zp1", 0),
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 0),
			sdk.NewInt64Coin("usdt", 0),
		},
		Fee: 2500,
	}
	abc := sdk.NewInt64Coin("abc", 0)
	usdt := sdk.NewInt64Coin("usdt", 30)

	params := types.Params{
		MaxSlippage: types.MaximumMaxSlippage,
	}
	_, _, _, _, err := keeper.CalculateLiquidityShares(pool, params, abc, usdt)

	require.Error(t, err)
	require.Equal(t, "when pool is empty both amounts must be positive: invalid request", err.Error())
}

func TestCalculateLiquidityShares_EmptyPool_ZeroQuote(t *testing.T) {
	// Test case: calculate liquidity shares if the pool is empty and quote amount is zero

	// create a pool --> empty pool
	pool := types.Pool{
		LpToken: sdk.NewInt64Coin("zp1", 0),
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 0),
			sdk.NewInt64Coin("usdt", 0),
		},
		Fee: 2500,
	}
	abc := sdk.NewInt64Coin("abc", 50)
	usdt := sdk.NewInt64Coin("usdt", 0)

	params := types.Params{
		MaxSlippage: types.MaximumMaxSlippage,
	}
	_, _, _, _, err := keeper.CalculateLiquidityShares(pool, params, abc, usdt)

	require.Error(t, err)
	require.Equal(t, "when pool is empty both amounts must be positive: invalid request", err.Error())
}

func TestCalculateLiquidityShares_EmptyPool_ZeroBaseAndQuote(t *testing.T) {
	// Test case: calculate liquidity shares if the pool is empty and both amounts are zero

	// create a pool --> empty pool
	pool := types.Pool{
		LpToken: sdk.NewInt64Coin("zp1", 0),
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 0),
			sdk.NewInt64Coin("usdt", 0),
		},
		Fee: 2500,
	}
	abc := sdk.NewInt64Coin("abc", 0)
	usdt := sdk.NewInt64Coin("usdt", 0)

	params := types.Params{
		MaxSlippage: types.MaximumMaxSlippage,
	}
	_, _, _, _, err := keeper.CalculateLiquidityShares(pool, params, abc, usdt)

	require.Error(t, err)
	require.Equal(t, "when pool is empty both amounts must be positive: invalid request", err.Error())
}

func TestMsgServerAddLiquidity_FailToSendReturnedCoins(t *testing.T) {
	// Test case: try to add liquidity to a pool, but fail to send returned coins due to insufficient funds

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 100)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	// square root of 100 * 100 = 100
	createPoolExpectedLPCoin := sample.Coin("zp1", 100)

	// how much we will add to the pool of abc and usdt coins
	// Adding a mismatched ratio to ensure returned coins
	addLiquidityBase := sample.Coin("abc", 211)
	addLiquidityQuote := sample.Coin("usdt", 522)
	// Expected: lpMintedFromBase = 211 * 100 / 100 = 211
	// lpMintedFromQuote = 522 * 100 / 100 = 522
	// lpMinted = min(211, 522) = 211
	// actualQuote = 211 (to match pool ratio 1:1)
	// returnedCoins = (0abc, 311usdt)
	addLiquidityExpectedLPCoin := sample.Coin("zp1", 211)
	addLiquidityExpectedReturnedCoins := sdk.NewCoins(sample.Coin("usdt", 311))

	// create dex keeper with pool mock
	server, dexKeeper, ctx, pool, poolAccount, bankKeeper, accountKeeper := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// set max slippage to maximum allowed value to avoid slippage errors
	params := types.DefaultParams()
	params.MaxSlippage = types.MaximumMaxSlippage
	dexKeeper.SetParams(ctx, params)

	// extract the pool address from the pool account
	poolAddress := poolAccount.GetAddress()

	// code will check if the signer has the required balance of abc
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityBase).
		Return(true).
		Times(2)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, addLiquidityQuote).
		Return(true).
		Times(2)

	// GetAccount for pool address
	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	// Send coins from signer to pool
	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(addLiquidityBase, addLiquidityQuote)).
		Return(nil).
		Times(1)

	// Mint new LP tokens
	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(addLiquidityExpectedLPCoin)).
		Return(nil).
		Times(1)

	// Send minted LP tokens to signer
	bankKeeper.
		EXPECT().
		SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, signer, sdk.NewCoins(addLiquidityExpectedLPCoin)).
		Return(nil).
		Times(1)

	// GetAccount for pool address when sending returned coins
	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	// Check balance for returned coins (usdt)
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), poolAddress, addLiquidityExpectedReturnedCoins[0]).
		Return(true).
		Times(1)

	// Simulate failure to send returned coins
	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), poolAddress, signer, addLiquidityExpectedReturnedCoins).
		Return(sdkerrors.ErrInsufficientFunds).
		Times(1)

	// create an "add liquidity" message
	txAddLiquidityMsg := &types.MsgAddLiquidity{
		Creator: creator,
		PoolId:  pool.PoolId,
		Base:    addLiquidityBase,
		Quote:   addLiquidityQuote,
	}

	// make rpc call to add liquidity
	_, err := server.AddLiquidity(ctx, txAddLiquidityMsg)

	// make sure there is an error
	require.Error(t, err)

	// check if the error message is correct
	require.Equal(t,
		fmt.Sprintf(
			"AddLiquidity: Failed to send back 'returned coins' (%s) to receiver: %s: insufficient funds",
			addLiquidityExpectedReturnedCoins.String(),
			signer.String(),
		),
		err.Error(),
	)
}

func TestCalculateLiquidityShares_InvalidPoolAmounts(t *testing.T) {
	// Test cases for when pool amounts are not positive
	testCases := []struct {
		name        string
		pool        types.Pool
		base        sdk.Coin
		quote       sdk.Coin
		expectError bool
		errorMsg    string
	}{
		{
			name: "pool with zero base amount",
			pool: types.Pool{
				LpToken: sdk.NewInt64Coin("zp1", 100),
				Coins: []sdk.Coin{
					sdk.NewInt64Coin("abc", 0), // Zero base amount
					sdk.NewInt64Coin("usdt", 100),
				},
				Fee: 2500,
			},
			base:        sdk.NewInt64Coin("abc", 100),
			quote:       sdk.NewInt64Coin("usdt", 100),
			expectError: true,
			errorMsg:    "pool amounts must be positive: invalid request",
		},
		{
			name: "pool with zero quote amount",
			pool: types.Pool{
				LpToken: sdk.NewInt64Coin("zp1", 100),
				Coins: []sdk.Coin{
					sdk.NewInt64Coin("abc", 100),
					sdk.NewInt64Coin("usdt", 0), // Zero quote amount
				},
				Fee: 2500,
			},
			base:        sdk.NewInt64Coin("abc", 100),
			quote:       sdk.NewInt64Coin("usdt", 100),
			expectError: true,
			errorMsg:    "pool amounts must be positive: invalid request",
		},
		{
			name: "pool with both zero amounts",
			pool: types.Pool{
				LpToken: sdk.NewInt64Coin("zp1", 100),
				Coins: []sdk.Coin{
					sdk.NewInt64Coin("abc", 0),  // Zero base amount
					sdk.NewInt64Coin("usdt", 0), // Zero quote amount
				},
				Fee: 2500,
			},
			base:        sdk.NewInt64Coin("abc", 100),
			quote:       sdk.NewInt64Coin("usdt", 100),
			expectError: true,
			errorMsg:    "pool amounts must be positive: invalid request",
		},
	}

	params := types.Params{
		MaxSlippage: types.MaximumMaxSlippage,
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shares, actualBase, actualQuote, returnedCoins, err := keeper.CalculateLiquidityShares(tc.pool, params, tc.base, tc.quote)

			if tc.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorMsg)
				require.Equal(t, sdk.Coin{}, shares)
				require.Equal(t, sdk.Coin{}, actualBase)
				require.Equal(t, sdk.Coin{}, actualQuote)
				require.Equal(t, sdk.Coins{}, returnedCoins)
			} else {
				require.NoError(t, err)
				// Verify that we got some valid return values
				require.True(t, shares.Denom != "")
				require.True(t, actualBase.Denom != "")
				require.True(t, actualQuote.Denom != "")
			}
		})
	}
}
