package keeper_test

import (
	"fmt"
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	keepertest "zigchain/testutil/keeper"
	"zigchain/testutil/sample"
	"zigchain/x/dex/keeper"
	"zigchain/x/dex/testutil"
	"zigchain/x/dex/testutil/common"
	"zigchain/x/dex/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

// Positive test cases

func TestMsgServerRemoveLiquidity_Positive(t *testing.T) {
	// Test case: remove liquidity from a pool

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

	// how much LP token we will remove
	removeLiquidityLPTokens := sample.Coin("zp1", 40)

	// how much abc and usdt tokens we expect to get back
	removeLiquidityExpectedBase := sample.Coin("abc", 40)
	removeLiquidityExpectedQuote := sample.Coin("usdt", 40)

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

	// code will check if the signer has the required balance of LP token
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, removeLiquidityLPTokens).
		Return(true).
		Times(1)

	// SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error
	// code will transfer LP tokens from the signer to the module
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(removeLiquidityLPTokens)).
		Return(nil).
		Times(1)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	// code will check if the signer has the required balance of base
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), poolAddress, removeLiquidityExpectedBase).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of quote
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), poolAddress, removeLiquidityExpectedQuote).
		Return(true).
		Times(1)

	// code will send abc and usdt coins from the module to the signer
	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), poolAddress, signer, sdk.NewCoins(removeLiquidityExpectedBase, removeLiquidityExpectedQuote)).
		Return(nil).
		Times(1)

	// BurnCoins(context.Context, string, sdk.Coins) error
	// code will burn the LP tokens that were removed
	bankKeeper.
		EXPECT().
		BurnCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(removeLiquidityLPTokens)).
		Return(nil).
		Times(1)

	// create a "remove liquidity" message
	txRemoveLiquidityMsg := &types.MsgRemoveLiquidity{
		Creator: creator,
		Lptoken: removeLiquidityLPTokens,
	}

	// make rpc call to remove liquidity
	resp, err := server.RemoveLiquidity(ctx, txRemoveLiquidityMsg)

	// make sure there is no error
	require.NoError(t, err)

	// check the response
	require.NotNil(t, resp)
	require.Equal(t, removeLiquidityExpectedBase, resp.Base)
	require.Equal(t, removeLiquidityExpectedQuote, resp.Quote)

	poolId := pool.PoolId
	// get the pool from the keeper, so we can compare new state to expected values
	pool, found := dexKeeper.GetPool(ctx, poolId)

	// make sure the pool was found
	require.True(t, found)

	// calculate new abc, usdt and LP token amounts after removing liquidity
	newPoolBase := createPoolBase.Sub(removeLiquidityExpectedBase)
	newPoolQuote := createPoolQuote.Sub(removeLiquidityExpectedQuote)
	newPoolLPToken := createPoolExpectedLPCoin.Sub(removeLiquidityLPTokens)

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

func TestMsgServerRemoveTotalLiquidity_Positive(t *testing.T) {
	// Test case: remove total liquidity from a pool

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // Assert that all expectations are met

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

	// how much LP token we will remove
	removeLiquidityLPTokens := sample.Coin("zp1", 100)

	// how much abc and usdt tokens we expect to get back
	removeLiquidityExpectedBase := sample.Coin("abc", 100)
	removeLiquidityExpectedQuote := sample.Coin("usdt", 100)

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

	// code will check if the signer has the required balance of LP token
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, removeLiquidityLPTokens).
		Return(true).
		Times(1)

	// SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error
	// code will transfer LP tokens from the signer to the module
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(removeLiquidityLPTokens)).
		Return(nil).
		Times(1)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	// code will check if the signer has the required balance of base
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), poolAddress, removeLiquidityExpectedBase).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of quote
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), poolAddress, removeLiquidityExpectedQuote).
		Return(true).
		Times(1)

	// code will send abc and usdt coins from the module to the signer
	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), poolAddress, signer, sdk.NewCoins(removeLiquidityExpectedBase, removeLiquidityExpectedQuote)).
		Return(nil).
		Times(1)

	// BurnCoins(context.Context, string, sdk.Coins) error
	// code will burn the LP tokens that were removed
	bankKeeper.
		EXPECT().
		BurnCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(removeLiquidityLPTokens)).
		Return(nil).
		Times(1)

	// create a "remove liquidity" message
	txRemoveLiquidityMsg := &types.MsgRemoveLiquidity{
		Creator: creator,
		Lptoken: removeLiquidityLPTokens,
	}

	// make rpc call to remove liquidity
	_, err := server.RemoveLiquidity(ctx, txRemoveLiquidityMsg)

	// make sure there is no error
	require.NoError(t, err)

	poolId := pool.PoolId
	// get the pool from the keeper, so we can compare new state to expected values
	pool, found := dexKeeper.GetPool(ctx, poolId)

	// make sure the pool was found
	require.True(t, found)

	// calculate new abc, usdt and LP token amounts after removing liquidity
	newPoolBaseAmount := createPoolBase.Amount.Sub(removeLiquidityExpectedBase.Amount)
	newPoolBase := sdk.NewInt64Coin("abc", newPoolBaseAmount.Int64())
	newPoolBaseAmount = createPoolQuote.Amount.Sub(removeLiquidityExpectedQuote.Amount)
	newPoolQuote := sdk.NewInt64Coin("usdt", newPoolBaseAmount.Int64())
	newPoolLPTokenAmount := createPoolExpectedLPCoin.Amount.Sub(removeLiquidityLPTokens.Amount)
	newPoolLPToken := sdk.NewInt64Coin("zp1", newPoolLPTokenAmount.Int64())

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

func TestMsgServerRemoveTotalLiquidity_Positive2(t *testing.T) {
	// Test case: remove total liquidity from a pool

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // Assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 1000)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 316)
	// how much LP token we will remove
	removeLiquidityLPTokens := sample.Coin("zp1", 50)

	// how much abc and usdt tokens we expect to get back
	removeLiquidityExpectedBase := sample.Coin("abc", 15)
	removeLiquidityExpectedQuote := sample.Coin("usdt", 158)

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

	// code will check if the signer has the required balance of LP token
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, removeLiquidityLPTokens).
		Return(true).
		Times(1)

	// SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error
	// code will transfer LP tokens from the signer to the module
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(removeLiquidityLPTokens)).
		Return(nil).
		Times(1)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	// code will check if the signer has the required balance of base
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), poolAddress, removeLiquidityExpectedBase).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of quote
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), poolAddress, removeLiquidityExpectedQuote).
		Return(true).
		Times(1)

	// code will send abc and usdt coins from the module to the signer
	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), poolAddress, signer, sdk.NewCoins(removeLiquidityExpectedBase, removeLiquidityExpectedQuote)).
		Return(nil).
		Times(1)

	// BurnCoins(context.Context, string, sdk.Coins) error
	// code will burn the LP tokens that were removed
	bankKeeper.
		EXPECT().
		BurnCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(removeLiquidityLPTokens)).
		Return(nil).
		Times(1)

	// create a "remove liquidity" message
	txRemoveLiquidityMsg := &types.MsgRemoveLiquidity{
		Creator: creator,
		Lptoken: removeLiquidityLPTokens,
	}

	// make rpc call to remove liquidity
	_, err := server.RemoveLiquidity(ctx, txRemoveLiquidityMsg)

	// make sure there is no error
	require.NoError(t, err)

	poolId := pool.PoolId
	// get the pool from the keeper, so we can compare new state to expected values
	pool, found := dexKeeper.GetPool(ctx, poolId)

	// make sure the pool was found
	require.True(t, found)

	// calculate new abc, usdt and LP token amounts after removing liquidity
	newPoolBaseAmount := createPoolBase.Amount.Sub(removeLiquidityExpectedBase.Amount)
	newPoolBase := sdk.NewInt64Coin("abc", newPoolBaseAmount.Int64())
	newPoolBaseAmount = createPoolQuote.Amount.Sub(removeLiquidityExpectedQuote.Amount)
	newPoolQuote := sdk.NewInt64Coin("usdt", newPoolBaseAmount.Int64())
	newPoolLPTokenAmount := createPoolExpectedLPCoin.Amount.Sub(removeLiquidityLPTokens.Amount)
	newPoolLPToken := sdk.NewInt64Coin("zp1", newPoolLPTokenAmount.Int64())

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

func TestMsgServerRemoveTotalLiquidity_WithReceiver(t *testing.T) {
	// Test case: remove total liquidity from a pool

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // Assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create a sample receiver address
	receiver := sample.AccAddress()

	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 1000)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 316)
	// how much LP token we will remove
	removeLiquidityLPTokens := sample.Coin("zp1", 50)

	// how much abc and usdt tokens we expect to get back
	removeLiquidityExpectedBase := sample.Coin("abc", 15)
	removeLiquidityExpectedQuote := sample.Coin("usdt", 158)

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

	// code will check if the signer has the required balance of LP token
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, removeLiquidityLPTokens).
		Return(true).
		Times(1)

	// SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error
	// code will transfer LP tokens from the signer to the module
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(removeLiquidityLPTokens)).
		Return(nil).
		Times(1)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	// code will check if the signer has the required balance of base
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), poolAddress, removeLiquidityExpectedBase).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of quote
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), poolAddress, removeLiquidityExpectedQuote).
		Return(true).
		Times(1)

	// code will send abc and usdt coins from the module to the signer
	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), poolAddress, sdk.MustAccAddressFromBech32(receiver), sdk.NewCoins(removeLiquidityExpectedBase, removeLiquidityExpectedQuote)).
		Return(nil).
		Times(1)

	// BurnCoins(context.Context, string, sdk.Coins) error
	// code will burn the LP tokens that were removed
	bankKeeper.
		EXPECT().
		BurnCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(removeLiquidityLPTokens)).
		Return(nil).
		Times(1)

	// create a "remove liquidity" message
	txRemoveLiquidityMsg := &types.MsgRemoveLiquidity{
		Creator:  creator,
		Lptoken:  removeLiquidityLPTokens,
		Receiver: receiver,
	}

	// make rpc call to remove liquidity
	_, err := server.RemoveLiquidity(ctx, txRemoveLiquidityMsg)

	// make sure there is no error
	require.NoError(t, err)

	poolId := pool.PoolId
	// get the pool from the keeper, so we can compare new state to expected values
	pool, found := dexKeeper.GetPool(ctx, poolId)

	// make sure the pool was found
	require.True(t, found)

	// calculate new abc, usdt and LP token amounts after removing liquidity
	newPoolBaseAmount := createPoolBase.Amount.Sub(removeLiquidityExpectedBase.Amount)
	newPoolBase := sdk.NewInt64Coin("abc", newPoolBaseAmount.Int64())
	newPoolBaseAmount = createPoolQuote.Amount.Sub(removeLiquidityExpectedQuote.Amount)
	newPoolQuote := sdk.NewInt64Coin("usdt", newPoolBaseAmount.Int64())
	newPoolLPTokenAmount := createPoolExpectedLPCoin.Amount.Sub(removeLiquidityLPTokens.Amount)
	newPoolLPToken := sdk.NewInt64Coin("zp1", newPoolLPTokenAmount.Int64())

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

func TestCoinsToRemove_Positive(t *testing.T) {
	// Test case: calculate the amount of base and quote tokens to remove

	pool := types.Pool{
		LpToken: sdk.NewInt64Coin("zp1", 387),
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 300),
			sdk.NewInt64Coin("usdt", 500),
		},
		Fee: 2500,
	}

	lpToken := sdk.NewInt64Coin("lp1", 50)
	coins, err := keeper.CoinsToRemove(pool, lpToken)
	require.NoError(t, err)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("abc", 38), sdk.NewInt64Coin("usdt", 64)), coins)
}

func TestCoinsToRemove_Positive2(t *testing.T) {
	// Test case: calculate the amount of base and quote tokens to remove

	pool := types.Pool{
		LpToken: sdk.NewInt64Coin("zp1", 100),
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 100),
			sdk.NewInt64Coin("usdt", 100),
		},
		Fee: 2500,
	}

	lpToken := sdk.NewInt64Coin("lp1", 50)
	coins, err := keeper.CoinsToRemove(pool, lpToken)
	require.NoError(t, err)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("abc", 50), sdk.NewInt64Coin("usdt", 50)), coins)
}

// Negative test cases

func TestMsgServerRemoveLiquidity_InvalidSigner(t *testing.T) {
	// Test case: try to remove liquidity from the pool with invalid signer address

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

	// how much LP token we will remove
	removeLiquidityLPTokens := sample.Coin("zp1", 100)

	// create dex keeper with pool mock
	server, _, ctx, _, _, _, _ := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// create a "remove liquidity" message
	txRemoveLiquidityMsg := &types.MsgRemoveLiquidity{
		Creator: "invalid address",
		Lptoken: removeLiquidityLPTokens,
	}

	// make rpc call to remove liquidity
	_, err := server.RemoveLiquidity(ctx, txRemoveLiquidityMsg)

	// make sure there is an error
	require.Error(t, err)

	// check an error message
	require.Equal(
		t,
		"Invalid address: invalid address: invalid address",
		err.Error(),
	)
}

func TestMsgServerRemoveLiquidity_PoolNotFound(t *testing.T) {
	// Test case: try to remove liquidity from a pool that does not exist

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, nil)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// how much LP token we will remove
	removeLiquidityLPTokens := sample.Coin("zp1", 100)

	// create a "remove liquidity" message
	txRemoveLiquidityMsg := &types.MsgRemoveLiquidity{
		Creator: creator,
		Lptoken: removeLiquidityLPTokens,
	}

	// make rpc call to remove liquidity
	_, err := srv.RemoveLiquidity(ctx, txRemoveLiquidityMsg)

	// make sure there is an error
	require.Error(t, err)

	// check if the error is correct
	require.ErrorIs(t, err, types.ErrPoolNotFound)

	// check an error message
	require.Equal(
		t,
		"pool zp1 not found: pool not found",
		err.Error(),
	)
}

func TestMsgServerRemoveLiquidity_InsufficientFunds(t *testing.T) {
	// Test case: try to remove liquidity from a pool if insufficient funds to remove liquidity

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

	// how much LP token we will remove
	removeLiquidityLPTokens := sample.Coin("zp1", 1000)

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

	// check if the pool was created
	require.NotNil(t, pool)

	// code will check if the signer has the required balance of LP token
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, removeLiquidityLPTokens).
		Return(false).
		Times(1)

	// create a "remove liquidity" message
	txRemoveLiquidityMsg := &types.MsgRemoveLiquidity{
		Creator: creator,
		Lptoken: removeLiquidityLPTokens,
	}

	// make rpc call to remove liquidity
	_, err := server.RemoveLiquidity(ctx, txRemoveLiquidityMsg)

	// make sure there is an error
	require.Error(t, err)

	// check if the error is correct
	require.ErrorIs(t, err, sdkerrors.ErrInsufficientFunds)

	// check an error message
	require.Equal(
		t,
		fmt.Sprintf(
			"Wallet has insufficient funds %s, can not remove liquidity: insufficient funds",
			removeLiquidityLPTokens.String(),
		),
		err.Error(),
	)
}

func TestMsgServerRemoveLiquidity_FailedToSendCoinsToModule(t *testing.T) {
	// Test case: failed to send coins to a dex account

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

	// how much LP token we will remove
	removeLiquidityLPTokens := sample.Coin("zp1", 50)

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

	// check if the pool was created
	require.NotNil(t, pool)

	// ensure this order of calls for remove liquidity
	gomock.InOrder(
		// code will check if the signer has the required balance of LP token
		bankKeeper.
			EXPECT().
			HasBalance(gomock.Any(), signer, removeLiquidityLPTokens).
			Return(true).
			Times(1),

		// SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error
		// code will return an error ErrInsufficientFunds
		bankKeeper.
			EXPECT().
			SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(removeLiquidityLPTokens)).
			Return(sdkerrors.ErrInsufficientFunds).
			Times(1),
	)

	// create a "remove liquidity" message
	txRemoveLiquidityMsg := &types.MsgRemoveLiquidity{
		Creator: creator,
		Lptoken: removeLiquidityLPTokens,
	}

	// make rpc call to remove liquidity
	_, err := server.RemoveLiquidity(ctx, txRemoveLiquidityMsg)

	// make sure there is an error
	require.Error(t, err)

	// check if the error is correct
	require.ErrorIs(t, err, sdkerrors.ErrInsufficientFunds)

	// check an error message
	require.Equal(
		t,
		fmt.Sprintf(
			"RemoveLiquidity: Failed to send %s coins, from signer %s to module %s: insufficient funds",
			removeLiquidityLPTokens.String(),
			signer.String(),
			types.ModuleName,
		),
		err.Error(),
	)
}

func TestMsgServerRemoveLiquidity_FailedToSendCoins(t *testing.T) {
	// Test case: failed to send coins from pool to address

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

	// how much LP token we will remove
	removeLiquidityLPTokens := sample.Coin("zp1", 50)

	// how much abc and usdt tokens we expect to get back
	removeLiquidityExpectedBase := sample.Coin("abc", 50)
	removeLiquidityExpectedQuote := sample.Coin("usdt", 50)

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

	// check if the pool was created
	require.NotNil(t, pool)

	// extract the pool address from the pool account
	poolAddress := poolAccount.GetAddress()

	// code will check if the signer has the required balance of LP token
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, removeLiquidityLPTokens).
		Return(true).
		Times(1)

	// SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error
	// code will transfer LP tokens from the signer to the module
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(removeLiquidityLPTokens)).
		Return(nil).
		Times(1)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	// code will check if the signer has the required balance of base
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), poolAddress, removeLiquidityExpectedBase).
		Return(false).
		Times(1)

	// create a "remove liquidity" message
	txRemoveLiquidityMsg := &types.MsgRemoveLiquidity{
		Creator: creator,
		Lptoken: removeLiquidityLPTokens,
	}

	// make rpc call to remove liquidity
	_, err := server.RemoveLiquidity(ctx, txRemoveLiquidityMsg)

	// make sure there is an error
	require.Error(t, err)

	// check if the error is correct
	require.ErrorIs(t, err, sdkerrors.ErrInsufficientFunds)

	// check an error message
	require.Equal(
		t,
		fmt.Sprintf(
			"RemoveLiquidity: Failed to send coins %s and %s: SendFromPoolToAddress: Insufficient funds in poolID %s (address: %s) to send %s coins to receiver %s: insufficient funds",
			removeLiquidityExpectedBase.String(),
			removeLiquidityExpectedQuote.String(),
			pool.PoolId,
			poolAddress.String(),
			removeLiquidityExpectedBase.String(),
			signer.String(),
		),
		err.Error(),
	)
}

func TestMsgServerRemoveLiquidity_CouldNotBurnCoins(t *testing.T) {
	// Test case: failed to burn LP tokens

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

	// how much LP token we will remove
	removeLiquidityLPTokens := sample.Coin("zp1", 50)

	// how much abc and usdt tokens we expect to get back
	removeLiquidityExpectedBase := sample.Coin("abc", 50)
	removeLiquidityExpectedQuote := sample.Coin("usdt", 50)

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

	// check if the pool was created
	require.NotNil(t, pool)

	// extract the pool address from the pool account
	poolAddress := poolAccount.GetAddress()

	// code will check if the signer has the required balance of LP token
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, removeLiquidityLPTokens).
		Return(true).
		Times(1)

	// SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error
	// code will transfer LP tokens from the signer to the module
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(removeLiquidityLPTokens)).
		Return(nil).
		Times(1)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	// code will check if the signer has the required balance of base
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), poolAddress, removeLiquidityExpectedBase).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of quote
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), poolAddress, removeLiquidityExpectedQuote).
		Return(true).
		Times(1)

	// code will send abc and usdt coins from the module to the signer
	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), poolAddress, signer, sdk.NewCoins(removeLiquidityExpectedBase, removeLiquidityExpectedQuote)).
		Return(nil).
		Times(1)

	// BurnCoins(context.Context, string, sdk.Coins) error
	// code will return an error ErrInsufficientBalance
	bankKeeper.
		EXPECT().
		BurnCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(removeLiquidityLPTokens)).
		Return(types.ErrInsufficientBalance).
		Times(1)

	// create a "remove liquidity" message
	txRemoveLiquidityMsg := &types.MsgRemoveLiquidity{
		Creator: creator,
		Lptoken: removeLiquidityLPTokens,
	}

	// make rpc call to remove liquidity
	_, err := server.RemoveLiquidity(ctx, txRemoveLiquidityMsg)

	// make sure there is an error
	require.Error(t, err)

	// check if the error is correct
	require.ErrorIs(t, err, types.ErrInsufficientBalance)

	// check an error message
	require.Equal(
		t,
		fmt.Sprintf(
			"BurnCoins: Could not burn liquidity pool token: %s: insufficient balance",
			removeLiquidityLPTokens.String(),
		),
		err.Error(),
	)
}

func TestMsgServerRemoveLiquidity_InsufficientBalance(t *testing.T) {
	// Test case: try to remove liquidity from a pool with insufficient balance

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

	// how much LP token we will remove
	removeLiquidityLPTokens := sample.Coin("zp1", 200)

	// create dex keeper with pool mock
	server, dexKeeper, ctx, pool, _, bankKeeper, _ := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	poolId := pool.PoolId
	// get the pool from the keeper, so we can compare new state to expected values
	pool, found := dexKeeper.GetPool(ctx, poolId)

	// make sure the pool was found
	require.True(t, found)

	// code will check if the signer has the required balance of LP token
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, removeLiquidityLPTokens).
		Return(true).
		Times(1)

	// SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error
	// code will transfer LP tokens from the signer to the module
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(removeLiquidityLPTokens)).
		Return(nil).
		Times(1)

	// create a "remove liquidity" message
	txRemoveLiquidityMsg := &types.MsgRemoveLiquidity{
		Creator:  creator,
		Lptoken:  removeLiquidityLPTokens,
		Receiver: sample.AccAddress(),
	}

	// make rpc call to remove liquidity
	_, err := server.RemoveLiquidity(ctx, txRemoveLiquidityMsg)

	require.Error(t, err)

	// check if the error msg is correct
	require.Equal(
		t,
		fmt.Sprintf(
			"Insufficient balance: %s lptoken: 100zp1: insufficient balance",
			removeLiquidityLPTokens.String(),
		),
		err.Error(),
	)
}

func TestMsgServerRemoveLiquidity_InvalidReceiver_BadAddress(t *testing.T) {
	// Test case: try to remove liquidity from a pool with invalid receiver address

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

	// how much LP token we will remove
	removeLiquidityLPTokens := sample.Coin("zp1", 100)

	// create dex keeper with pool mock
	server, _, ctx, _, _, bankKeeper, _ := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// code will check if the signer has the required balance of LP token
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, removeLiquidityLPTokens).
		Return(true).
		Times(1)

	// SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error
	// code will transfer LP tokens from the signer to the module
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(removeLiquidityLPTokens)).
		Return(nil).
		Times(1)

	// create a "remove liquidity" message
	txRemoveLiquidityMsg := &types.MsgRemoveLiquidity{
		Creator:  creator,
		Lptoken:  removeLiquidityLPTokens,
		Receiver: "bad_receiver",
	}

	// make rpc call to remove liquidity
	_, err := server.RemoveLiquidity(ctx, txRemoveLiquidityMsg)

	require.Error(t, err)

	// check if the error msg is correct
	require.Equal(t,
		"decoding bech32 failed: invalid separator index -1",
		err.Error(),
	)
}

func TestMsgServerRemoveLiquidity_InvalidReceiver_TooShort(t *testing.T) {
	// Test case: try to remove liquidity from a pool with invalid receiver address --> too short

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

	// how much LP token we will remove
	removeLiquidityLPTokens := sample.Coin("zp1", 100)

	// create dex keeper with pool mock
	server, _, ctx, _, _, bankKeeper, _ := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// code will check if the signer has the required balance of LP token
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, removeLiquidityLPTokens).
		Return(true).
		Times(1)

	// SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error
	// code will transfer LP tokens from the signer to the module
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(removeLiquidityLPTokens)).
		Return(nil).
		Times(1)

	// create a "remove liquidity" message
	txRemoveLiquidityMsg := &types.MsgRemoveLiquidity{
		Creator:  creator,
		Lptoken:  removeLiquidityLPTokens,
		Receiver: "zig123",
	}

	// make rpc call to remove liquidity
	_, err := server.RemoveLiquidity(ctx, txRemoveLiquidityMsg)

	require.Error(t, err)

	// check if the error msg is correct
	require.Equal(t,
		"decoding bech32 failed: invalid bech32 string length 6",
		err.Error(),
	)
}

func TestMsgServerRemoveLiquidity_InvalidReceiver_TooLong(t *testing.T) {
	// Test case: try to remove liquidity from a pool with invalid receiver address --> too long

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

	// how much LP token we will remove
	removeLiquidityLPTokens := sample.Coin("zp1", 100)

	// create dex keeper with pool mock
	server, _, ctx, _, _, bankKeeper, _ := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// code will check if the signer has the required balance of LP token
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, removeLiquidityLPTokens).
		Return(true).
		Times(1)

	// SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error
	// code will transfer LP tokens from the signer to the module
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(removeLiquidityLPTokens)).
		Return(nil).
		Times(1)

	// create a "remove liquidity" message
	txRemoveLiquidityMsg := &types.MsgRemoveLiquidity{
		Creator:  creator,
		Lptoken:  removeLiquidityLPTokens,
		Receiver: "zig1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
	}

	// make rpc call to remove liquidity
	_, err := server.RemoveLiquidity(ctx, txRemoveLiquidityMsg)

	require.Error(t, err)

	// check if the error msg is correct
	require.Equal(t,
		"decoding bech32 failed: invalid checksum (expected yurny3 got 567890)",
		err.Error(),
	)
}

func TestMsgServerRemoveLiquidity_InvalidReceiver_NotLowercaseOrUppercase(t *testing.T) {
	// Test case: try to remove liquidity from a pool with invalid receiver address --> not all lowercase or uppercase

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

	// how much LP token we will remove
	removeLiquidityLPTokens := sample.Coin("zp1", 100)

	// create dex keeper with pool mock
	server, _, ctx, _, _, bankKeeper, _ := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// code will check if the signer has the required balance of LP token
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, removeLiquidityLPTokens).
		Return(true).
		Times(1)

	// SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error
	// code will transfer LP tokens from the signer to the module
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(removeLiquidityLPTokens)).
		Return(nil).
		Times(1)

	// create a "remove liquidity" message
	txRemoveLiquidityMsg := &types.MsgRemoveLiquidity{
		Creator:  creator,
		Lptoken:  removeLiquidityLPTokens,
		Receiver: "BAd Address",
	}

	// make rpc call to remove liquidity
	_, err := server.RemoveLiquidity(ctx, txRemoveLiquidityMsg)

	require.Error(t, err)

	// check if the error msg is correct
	require.Equal(t,
		"decoding bech32 failed: string not all lowercase or all uppercase",
		err.Error(),
	)
}

func TestMsgServerRemoveLiquidity_InvalidReceiver_BadChars(t *testing.T) {
	// Test case: try to remove liquidity from a pool with invalid receiver address --> bad characters

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

	// how much LP token we will remove
	removeLiquidityLPTokens := sample.Coin("zp1", 100)

	// create dex keeper with pool mock
	server, _, ctx, _, _, bankKeeper, _ := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// code will check if the signer has the required balance of LP token
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, removeLiquidityLPTokens).
		Return(true).
		Times(1)

	// SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error
	// code will transfer LP tokens from the signer to the module
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(removeLiquidityLPTokens)).
		Return(nil).
		Times(1)

	// create a "remove liquidity" message
	txRemoveLiquidityMsg := &types.MsgRemoveLiquidity{
		Creator:  creator,
		Lptoken:  removeLiquidityLPTokens,
		Receiver: "zi!12z7zhe2d03wn03e8lkw8at6zjxa7y65q5xnp5%",
	}

	// make rpc call to remove liquidity
	_, err := server.RemoveLiquidity(ctx, txRemoveLiquidityMsg)

	require.Error(t, err)

	// check if the error msg is correct
	require.Equal(t,
		"decoding bech32 failed: invalid character not part of charset: 37",
		err.Error(),
	)
}

func TestCoinsToRemove_InsufficientBalance(t *testing.T) {
	// Test case: insufficient balance to remove liquidity

	pool := types.Pool{
		LpToken: sdk.NewInt64Coin("zp1", 100),
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 100),
			sdk.NewInt64Coin("usdt", 100),
		},
		Fee: 2500,
	}
	lpToken := sdk.NewInt64Coin("lp1", 1000)

	_, err := keeper.CoinsToRemove(pool, lpToken)
	require.Error(t, err)
	require.Equal(
		t,
		fmt.Sprintf(
			"Insufficient balance: %s lptoken: %s: insufficient balance",
			lpToken.String(),
			pool.LpToken.String(),
		),
		err.Error(),
	)
}

func TestCoinsToRemove_ScaleFactorBenefits(t *testing.T) {
	// Test cases demonstrating the benefits of scale factor in precision handling
	tests := []struct {
		name            string
		poolLPTokens    int64
		poolBaseTokens  int64
		poolQuoteTokens int64
		removeLPTokens  int64
		expectedBase    int64
		expectedQuote   int64
		description     string
	}{
		{
			name:            "Small liquidity removal with scale factor",
			poolLPTokens:    1000000, // 1M LP tokens
			poolBaseTokens:  1000000, // 1M base tokens
			poolQuoteTokens: 1000000, // 1M quote tokens
			removeLPTokens:  1,       // Remove just 1 LP token
			expectedBase:    1,       // Should get 1 base token
			expectedQuote:   1,       // Should get 1 quote token
			description:     "Demonstrates scale factor preserves precision for small removals",
		},
		{
			name:            "Very small liquidity removal",
			poolLPTokens:    1000000000, // 1B LP tokens
			poolBaseTokens:  1000000000, // 1B base tokens
			poolQuoteTokens: 1000000000, // 1B quote tokens
			removeLPTokens:  100,        // Remove 100 LP tokens
			expectedBase:    100,        // Should get 100 base tokens
			expectedQuote:   100,        // Should get 100 quote tokens
			description:     "Demonstrates scale factor works with very large pools",
		},
		{
			name:            "Fractional liquidity removal",
			poolLPTokens:    1000, // 1K LP tokens
			poolBaseTokens:  1000, // 1K base tokens
			poolQuoteTokens: 1000, // 1K quote tokens
			removeLPTokens:  333,  // Remove 333 LP tokens (33.3%)
			expectedBase:    333,  // Should get 333 base tokens
			expectedQuote:   333,  // Should get 333 quote tokens
			description:     "Demonstrates scale factor handles fractional percentages correctly",
		},
		{
			name:            "Single LP token removal from large pool",
			poolLPTokens:    1000000000000, // 1T LP tokens
			poolBaseTokens:  1000000000000, // 1T base tokens
			poolQuoteTokens: 1000000000000, // 1T quote tokens
			removeLPTokens:  1,             // Remove 1 LP token
			expectedBase:    1,             // Should get 1 base token
			expectedQuote:   1,             // Should get 1 quote token
			description:     "Demonstrates scale factor works with trillion-sized pools",
		},
		{
			name:            "Single LP token removal from large pool with 1K base and 1T quote",
			poolLPTokens:    1000000000000, // 1T LP tokens
			poolBaseTokens:  1000,          // 1K base tokens
			poolQuoteTokens: 1000000000000, // 1T quote tokens
			removeLPTokens:  1,             // Remove 1 LP token
			expectedBase:    0,             // Should get 0 base token
			expectedQuote:   1,             // Should get 1 quote token
			description:     "Demonstrates scale factor works with trillion-sized pools",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := types.Pool{
				LpToken: sdk.NewInt64Coin("zp1", tt.poolLPTokens),
				Coins: []sdk.Coin{
					sdk.NewInt64Coin("abc", tt.poolBaseTokens),
					sdk.NewInt64Coin("usdt", tt.poolQuoteTokens),
				},
				Fee: 2500,
			}
			lpToken := sdk.NewInt64Coin("zp1", tt.removeLPTokens)
			coins, err := keeper.CoinsToRemove(pool, lpToken)
			require.NoError(t, err)
			require.Equal(t, sdk.NewCoins(
				sdk.NewInt64Coin("abc", tt.expectedBase),
				sdk.NewInt64Coin("usdt", tt.expectedQuote),
			), coins)
		})
	}
}

func TestCoinsToRemove_ScaleFactorEdgeCases(t *testing.T) {
	// Test edge cases to ensure scale factor handles them correctly
	tests := []struct {
		name            string
		poolLPTokens    int64
		poolBaseTokens  int64
		poolQuoteTokens int64
		removeLPTokens  int64
		shouldSucceed   bool
		description     string
	}{
		{
			name:            "Remove exactly half of pool",
			poolLPTokens:    1000,
			poolBaseTokens:  1000,
			poolQuoteTokens: 1000,
			removeLPTokens:  500,
			shouldSucceed:   true,
			description:     "Should get exactly half of base and quote tokens",
		},
		{
			name:            "Remove all but one LP token",
			poolLPTokens:    1000,
			poolBaseTokens:  1000,
			poolQuoteTokens: 1000,
			removeLPTokens:  999,
			shouldSucceed:   true,
			description:     "Should get almost all base and quote tokens",
		},
		{
			name:            "Remove single LP token from small pool",
			poolLPTokens:    10,
			poolBaseTokens:  10,
			poolQuoteTokens: 10,
			removeLPTokens:  1,
			shouldSucceed:   true,
			description:     "Should get 1 base and 1 quote token",
		},
		{
			name:            "Remove more LP tokens than exist",
			poolLPTokens:    100,
			poolBaseTokens:  100,
			poolQuoteTokens: 100,
			removeLPTokens:  101,
			shouldSucceed:   false,
			description:     "Should fail with insufficient balance error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := types.Pool{
				LpToken: sdk.NewInt64Coin("zp1", tt.poolLPTokens),
				Coins: []sdk.Coin{
					sdk.NewInt64Coin("abc", tt.poolBaseTokens),
					sdk.NewInt64Coin("usdt", tt.poolQuoteTokens),
				},
				Fee: 2500,
			}
			lpToken := sdk.NewInt64Coin("zp1", tt.removeLPTokens)
			coins, err := keeper.CoinsToRemove(pool, lpToken)
			if tt.shouldSucceed {
				require.NoError(t, err)
				require.Len(t, coins, 2)
				// Verify proportional amounts
				expectedBase := tt.poolBaseTokens * tt.removeLPTokens / tt.poolLPTokens
				expectedQuote := tt.poolQuoteTokens * tt.removeLPTokens / tt.poolLPTokens
				require.Equal(t, expectedBase, coins[0].Amount.Int64())
				require.Equal(t, expectedQuote, coins[1].Amount.Int64())
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), "Insufficient balance")
			}
		})
	}
}

func TestMsgServerRemoveLiquidity_LpTokenDenomMismatch(t *testing.T) {
	// Test case: pool found but LP token denom does not match (sanity check failure)

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

	// how much LP token we will remove
	removeLiquidityLPTokens := sample.Coin("zp1", 50)

	// create dex keeper with pool mock
	server, dexKeeper, ctx, pool, _, _, _ := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	poolId := pool.PoolId

	// Retrieve the pool, modify the LP token denom to trigger mismatch, and set it back
	// Assuming SetPool uses pool.PoolId as the storage key, this keeps it retrievable under the original key
	pool, found := dexKeeper.GetPool(ctx, poolId)
	require.True(t, found)
	originalDenom := pool.LpToken.Denom
	pool.LpToken.Denom = "wrongzp1"
	dexKeeper.SetPool(ctx, pool)

	// create a "remove liquidity" message with original denom
	txRemoveLiquidityMsg := &types.MsgRemoveLiquidity{
		Creator: creator,
		Lptoken: removeLiquidityLPTokens,
	}

	// make rpc call to remove liquidity
	_, err := server.RemoveLiquidity(ctx, txRemoveLiquidityMsg)

	// make sure there is an error
	require.Error(t, err)

	// check if the error is correct
	require.ErrorIs(t, err, types.ErrPoolNotFound)

	// check an error message
	require.Equal(
		t,
		fmt.Sprintf(
			"pool lp token: wrongzp1 different then incoming lp token: %s: pool not found",
			originalDenom,
		),
		err.Error(),
	)
}

func TestCoinsToRemove_PoolLpTokenZero(t *testing.T) {
	// Test case: pool LP token amount is zero

	pool := types.Pool{
		LpToken: sdk.NewInt64Coin("zp1", 0),
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 100),
			sdk.NewInt64Coin("usdt", 100),
		},
		Fee: 2500,
	}

	lpToken := sdk.NewInt64Coin("zp1", 50)

	_, err := keeper.CoinsToRemove(pool, lpToken)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInsufficientBalance)
	require.Equal(
		t,
		fmt.Sprintf(
			"Insufficient balance: %s pool lp token: %s: insufficient balance",
			lpToken.String(),
			pool.LpToken.String(),
		),
		err.Error(),
	)
}
