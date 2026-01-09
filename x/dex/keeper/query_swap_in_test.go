package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "zigchain/testutil/keeper"
	"zigchain/testutil/sample"
	"zigchain/x/dex/keeper"
	"zigchain/x/dex/testutil"
	"zigchain/x/dex/types"
	"zigchain/zutils/constants"
)

// Positive test cases

func TestQuerySwapIn_Valid(t *testing.T) {
	// Test case: querying the swap-in amount for a valid pool and coin

	// create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// sample signer address
	creator := sample.AccAddress()
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 1000)
	creationFee := sample.Coin("uzig", 100000000)
	// square root of a * b
	createPoolExpectedLPCoin := sample.Coin("zp1", 316)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	// create mock account keeper
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, accountKeeper)

	// set minimal lock to 0
	params := k.GetParams(ctx)
	params.MinimalLiquidityLock = 0
	k.SetParams(ctx, params)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// get id of next pool
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

	// create a sample message
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

	// create query server
	qs := keeper.NewQueryServerImpl(k)

	// how much we want to swap
	incoming := sample.Coin("abc", 10)

	outUsdt := sample.Coin("usdt", 83)

	// perform SwapIn query
	resp, err := qs.SwapIn(ctx, &types.QuerySwapInRequest{
		PoolId: pool.PoolId,
		CoinIn: incoming.String(),
	})

	require.NoError(t, err)

	// verify the response
	require.Equal(t, "usdt", resp.CoinOut.Denom)
	require.Equal(t, outUsdt.Amount.String(), resp.CoinOut.Amount.String())
	require.Equal(t, "abc", resp.Fee.Denom)
	require.Equal(t, "1", resp.Fee.Amount.String())
}

// Negative test cases

func TestQuerySwapIn_InvalidRequest(t *testing.T) {
	// Test case: querying the swap-in amount with a nil request

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	qs := keeper.NewQueryServerImpl(k)

	_, err := qs.SwapIn(ctx, nil)

	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	require.Contains(t, err.Error(), "invalid request")
}

func TestQuerySwapIn_EmptyPool(t *testing.T) {
	// Test case: querying the swap-in amount sending an empty pool id

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	qs := keeper.NewQueryServerImpl(k)

	_, err := qs.SwapIn(ctx, &types.QuerySwapInRequest{
		PoolId: "",
		CoinIn: "abc10",
	})

	require.Error(t, err)

	// check an error message
	require.Equal(t, "Invalid pool id: pool id is empty: invalid coins", err.Error())
}

func TestQuerySwapIn_PoolIdTooShort(t *testing.T) {
	// Test case: querying the swap-in amount sending a pool id that is too short

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	qs := keeper.NewQueryServerImpl(k)

	_, err := qs.SwapIn(ctx, &types.QuerySwapInRequest{
		PoolId: "z",
		CoinIn: "abc10",
	})

	require.Error(t, err)

	// check an error message
	require.Equal(t,
		fmt.Sprintf(
			"Invalid pool id: 'z' pool id is too short, minimum %d characters: invalid coins",
			constants.MinSubDenomLength,
		),
		err.Error(),
	)
}

func TestQuerySwapIn_PoolIdTooLong(t *testing.T) {
	// Test case: querying the swap-in amount sending a pool id that is too long

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	qs := keeper.NewQueryServerImpl(k)

	poolId := "zp1112socdjdfjdcjdskfjdkfjdskfjkdsfjedsalfjdsjfdskvdskvkfjsd"
	_, err := qs.SwapIn(ctx, &types.QuerySwapInRequest{
		PoolId: poolId,
		CoinIn: "abc10",
	})

	require.Error(t, err)

	// check an error message
	require.Equal(t,
		fmt.Sprintf(
			"Invalid pool id: '%s' pool id is too long (60), maximum %d characters: invalid coins",
			poolId,
			constants.MaxSubDenomLength,
		),
		err.Error(),
	)
}

func TestQuerySwapIn_PoolIdBadPrefix(t *testing.T) {
	// Test case: querying the swap-in amount sending a pool id that has a bad prefix

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	qs := keeper.NewQueryServerImpl(k)

	poolId := "bla1"
	_, err := qs.SwapIn(ctx, &types.QuerySwapInRequest{
		PoolId: poolId,
		CoinIn: "abc10",
	})

	require.Error(t, err)

	// check an error message
	require.Equal(t,
		fmt.Sprintf(
			"Invalid pool id: '%s', pool id has to start with '%s' followed by numbers e.g. %s123: invalid coins",
			poolId,
			constants.PoolPrefix,
			constants.PoolPrefix,
		),
		err.Error(),
	)
}

func TestQuerySwapIn_PoolIdBadChars(t *testing.T) {
	// Test case: querying the swap-in amount sending a pool id that has a bad characters

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	qs := keeper.NewQueryServerImpl(k)

	poolId := "zp1!!!invalid-id"
	_, err := qs.SwapIn(ctx, &types.QuerySwapInRequest{
		PoolId: poolId,
		CoinIn: "abc10",
	})

	require.Error(t, err)

	// check an error message
	require.Equal(t,
		fmt.Sprintf(
			"Invalid pool id: '%s', pool id has to start with '%s' followed by numbers e.g. %s123: invalid coins",
			poolId,
			constants.PoolPrefix,
			constants.PoolPrefix,
		),
		err.Error(),
	)
}

func TestQuerySwapIn_CoinInParseError(t *testing.T) {
	// Test case: querying the swap-in amount sending a malformed coin

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	qs := keeper.NewQueryServerImpl(k)

	_, err := qs.SwapIn(ctx, &types.QuerySwapInRequest{
		PoolId: "zp1",
		CoinIn: "bad*denom10", // malformed format
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "Invalid token format: failed to parse incoming token")
}

func TestQuerySwapIn_FailToParseToken(t *testing.T) {
	// Test case: querying the swap-in amount sending a coin with a bad denom

	// create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// sample signer address
	creator := sample.AccAddress()
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 1000)
	creationFee := sample.Coin("uzig", 100000000)
	// square root of a * b
	createPoolExpectedLPCoin := sample.Coin("zp1", 316)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	// create mock account keeper
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, accountKeeper)

	// set minimal lock to 0
	params := k.GetParams(ctx)
	params.MinimalLiquidityLock = 0
	k.SetParams(ctx, params)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// get id of next pool
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

	// create a sample message
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

	// create query server
	qs := keeper.NewQueryServerImpl(k)

	_, err = qs.SwapIn(ctx, &types.QuerySwapInRequest{
		PoolId: pool.PoolId,
		CoinIn: "10abc_", // bad denom
	})

	require.Error(t, err)
	require.Error(t, err)
	require.Equal(t, "Invalid token format: "+
		"failed to parse incoming token: "+
		"10abc_: invalid decimal coin expression: 10abc_",
		err.Error(),
	)
}

func TestQuerySwapIn_PoolNotFound(t *testing.T) {
	// Test case: querying the swap-in amount for a pool that does not exist

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	qs := keeper.NewQueryServerImpl(k)

	_, err := qs.SwapIn(ctx, &types.QuerySwapInRequest{
		PoolId: "zp1",
		CoinIn: "10abc", // bad denom
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "Liquidity pool (zp1) can not be found")
}

func TestQuerySwapIn_InvalidIncomingToken(t *testing.T) {
	// Test case: querying the swap-in amount with an invalid incoming token

	// create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// sample signer address
	creator := sample.AccAddress()
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 1000)
	creationFee := sample.Coin("uzig", 100000000)
	// square root of a * b
	createPoolExpectedLPCoin := sample.Coin("zp1", 316)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	// create mock account keeper
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, accountKeeper)

	// set minimal lock to 0
	params := k.GetParams(ctx)
	params.MinimalLiquidityLock = 0
	k.SetParams(ctx, params)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// get id of next pool
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

	// create a sample message
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

	// create query server
	qs := keeper.NewQueryServerImpl(k)

	// how much we want to swap
	incoming := sample.Coin("abc-", 10)

	// perform SwapIn query
	_, err = qs.SwapIn(ctx, &types.QuerySwapInRequest{
		PoolId: pool.PoolId,
		CoinIn: incoming.String(),
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "Invalid incoming coin (abc-), this pool only supports base (abc) and quote (usdt) tokens: invalid request")
}

func TestQuerySwapIn_InvalidTokenFormat(t *testing.T) {
	// Test case: querying the swap-in amount with an invalid token format

	// create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// sample signer address
	creator := sample.AccAddress()
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 1000)
	creationFee := sample.Coin("uzig", 100000000)
	// square root of a * b
	createPoolExpectedLPCoin := sample.Coin("zp1", 316)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	// create mock account keeper
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, accountKeeper)

	// set minimal lock to 0
	params := k.GetParams(ctx)
	params.MinimalLiquidityLock = 0
	k.SetParams(ctx, params)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// get id of next pool
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

	// create a sample message
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

	// create query server
	qs := keeper.NewQueryServerImpl(k)

	// perform SwapIn query
	_, err = qs.SwapIn(ctx, &types.QuerySwapInRequest{
		PoolId: pool.PoolId,
		CoinIn: "#110_invalid",
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "Invalid token format: failed to parse incoming token: #110_invalid")
}

func TestQuerySwapIn_CheckCoinDenomValidation(t *testing.T) {
	// Test case: trigger the CheckCoinDenom validation error

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, accountKeeper)

	// Create a simple pool directly without going through message server
	pool := types.Pool{
		PoolId:  "zp1",
		Creator: sample.AccAddress(),
		Coins: []sdk.Coin{
			sample.Coin("abc", 100),
			sample.Coin("usdt", 1000),
		},
		LpToken: sample.Coin("zp1", 100),
		Fee:     500,
		Formula: "constant_product",
	}
	k.SetPool(ctx, pool)

	qs := keeper.NewQueryServerImpl(k)

	invalidCharCoin := "10abc-def"

	_, err := qs.SwapIn(ctx, &types.QuerySwapInRequest{
		PoolId: "zp1",
		CoinIn: invalidCharCoin,
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "Invalid incoming coin (abc-def), this pool only supports base (abc) and quote (usdt) tokens: invalid request")
}

func TestQuerySwapIn_CheckCoinDenomShortDenom(t *testing.T) {
	// Test case: denom too short validation

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, accountKeeper)

	// Create a pool
	pool := types.Pool{
		PoolId:  "zp1",
		Creator: sample.AccAddress(),
		Coins: []sdk.Coin{
			sample.Coin("abc", 100),
			sample.Coin("usdt", 1000),
		},
		LpToken: sample.Coin("zp1", 100),
		Fee:     500,
		Formula: "constant_product",
	}
	k.SetPool(ctx, pool)

	qs := keeper.NewQueryServerImpl(k)

	// Try a short denom
	veryShortCoin := "10ab"

	_, err := qs.SwapIn(ctx, &types.QuerySwapInRequest{
		PoolId: "zp1",
		CoinIn: veryShortCoin,
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "Invalid token format: failed to validate incoming token")
	require.Contains(t, err.Error(), veryShortCoin)
	require.Contains(t, err.Error(), "denom name is too short")
}
