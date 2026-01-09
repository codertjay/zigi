package keeper_test

import (
	"fmt"
	"strconv"
	"testing"

	"cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"zigchain/x/dex/keeper"
	"zigchain/x/dex/testutil"

	keepertest "zigchain/testutil/keeper"
	"zigchain/testutil/sample"
	"zigchain/x/dex/testutil/common"
	"zigchain/x/dex/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

// Positive test cases

func TestMsgServerSwapExactOut_SwapBase_Positive(t *testing.T) {
	// Test case: swap exact number of coins in a pool

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
	createPoolBase := sample.Coin("abc", 542000000)
	createPoolQuote := sample.Coin("usdt", 112000000)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 246381817)

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

	// how much we will swap
	outgoing := sample.Coin("usdt", 2019024)

	// how much we will swap
	inCoin, _, errCalcSwap := keeper.CalculateSwapExactOutAmount(&pool, outgoing)

	// make sure there is no error
	require.NoError(t, errCalcSwap)

	// check if the swap amount is correct
	inAbc := sample.Coin("abc", 10000002)
	require.Equal(t, inAbc, inCoin)

	// extract the pool address from the pool account
	poolAddress := types.GetPoolAddress(pool.PoolId)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, inAbc).
		Return(true).
		Times(1)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(2)

	// code will send the incoming coins
	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(inAbc)).
		Return(nil).
		Times(1)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), poolAddress, outgoing).
		Return(true).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), poolAddress, signer, sdk.NewCoins(outgoing)).
		Return(nil).
		Times(1)

	// create a "swap-exact-out" message
	txSwapExactOutMsg := &types.MsgSwapExactOut{
		Signer:   creator,
		Outgoing: outgoing,
		PoolId:   pool.PoolId,
	}

	// make rpc call to swap coins
	resp, err := server.SwapExactOut(ctx, txSwapExactOutMsg)

	// make sure there is no error
	require.NoError(t, err)

	poolId := pool.PoolId
	// check the response
	require.Equal(t, poolId, resp.PoolId)
	require.Equal(t, outgoing, resp.Outgoing)
	require.Equal(t, inCoin, resp.Incoming)
	require.Equal(t, sdk.NewCoin("abc", math.NewInt(50000)), resp.Fee)
	require.Equal(t, creator, resp.Receiver)
	require.Nil(t, resp.IncomingMax)

	// get the pool from the keeper, so we can compare new state to expected values
	pool, found := dexKeeper.GetPool(ctx,
		poolId,
	)

	// make sure the pool was found
	require.True(t, found)

	// new base is old base + inCoin
	newPoolBase := createPoolBase.Add(inCoin)
	// new quote is old quote - coins sent to the sender
	newPoolQuote := createPoolQuote.Sub(outgoing)
	// LP tokens doesn't change
	newPoolLPToken := createPoolExpectedLPCoin

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

func TestMsgServerSwapExactOut_SwapBase_Positive2(t *testing.T) {
	// Test case: swap the exact number of coins in a pool

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
	createPoolBase := sample.Coin("abc", 100000000)
	createPoolQuote := sample.Coin("usdt", 1000000000)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 316227766)
	// how much we will swap
	outgoing := sample.Coin("usdt", 90495679)

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

	// how much we will swap
	inCoin, fee, errCalcSwap := keeper.CalculateSwapExactOutAmount(&pool, outgoing)

	// make sure there is no error
	require.NoError(t, errCalcSwap)

	require.Equal(t, fee.Amount.String(), "49999")

	// check if the swap amount is correct
	inAbc := sample.Coin("abc", 9999998)
	require.Equal(t, inAbc.Amount.String(), inCoin.Amount.String())

	// extract the pool address from the pool account
	poolAddress := types.GetPoolAddress(pool.PoolId)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, inAbc).
		Return(true).
		Times(1)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(2)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), poolAddress, outgoing).
		Return(true).
		Times(1)

	// code will send the incoming coins
	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(inAbc)).
		Return(nil).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), poolAddress, signer, sdk.NewCoins(outgoing)).
		Return(nil).
		Times(1)

	// create a "swap-exact-out" message
	txSwapExactOutMsg := &types.MsgSwapExactOut{
		Signer:   creator,
		Outgoing: outgoing,
		PoolId:   pool.PoolId,
	}

	// make rpc call to swap coins
	resp, err := server.SwapExactOut(ctx, txSwapExactOutMsg)

	// make sure there is no error
	require.NoError(t, err)

	poolId := pool.PoolId
	// check the response
	require.Equal(t, poolId, resp.PoolId)
	require.Equal(t, outgoing, resp.Outgoing)
	require.Equal(t, inCoin, resp.Incoming)
	require.Equal(t, sdk.NewCoin("abc", math.NewInt(49999)), resp.Fee)
	require.Equal(t, creator, resp.Receiver)
	require.Nil(t, resp.IncomingMax)

	// get the pool from the keeper, so we can compare new state to expected values
	pool, found := dexKeeper.GetPool(ctx,
		poolId,
	)

	// make sure the pool was found
	require.True(t, found)

	// new base is old base + inCoin
	newPoolBase := createPoolBase.Add(inCoin)
	// the new quote is old quote - coins sent to the sender
	newPoolQuote := createPoolQuote.Sub(outgoing)
	// LP tokens doesn't change
	newPoolLPToken := createPoolExpectedLPCoin

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

func TestMsgServerSwapExactOut_SwapQuote(t *testing.T) {
	// Test case: swap the exact number of coins in a pool

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
	createPoolQuote := sample.Coin("usdt", 1000)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 316)

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

	// how much we will swap
	outgoing := sample.Coin("abc", 1)

	// how much we will swap
	inCoin, _, errCalcSwap := keeper.CalculateSwapExactOutAmount(&pool, outgoing)

	// make sure there is no error
	require.NoError(t, errCalcSwap)

	// check if the swap amount is correct
	inAbc := sample.Coin("usdt", 10)
	require.Equal(t, inAbc, inCoin)

	// extract the pool address from the pool account
	poolAddress := types.GetPoolAddress(pool.PoolId)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, inAbc).
		Return(true).
		Times(1)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(2)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), poolAddress, outgoing).
		Return(true).
		Times(1)

	// code will send the incoming coins
	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(inAbc)).
		Return(nil).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), poolAddress, signer, sdk.NewCoins(outgoing)).
		Return(nil).
		Times(1)

	// create a "swap-exact-out" message
	txSwapExactOutMsg := &types.MsgSwapExactOut{
		Signer:   creator,
		Outgoing: outgoing,
		PoolId:   pool.PoolId,
	}

	// make rpc call to swap coins
	resp, err := server.SwapExactOut(ctx, txSwapExactOutMsg)

	// make sure there is no error
	require.NoError(t, err)

	poolId := pool.PoolId
	// check the response
	require.Equal(t, poolId, resp.PoolId)
	require.Equal(t, outgoing, resp.Outgoing)
	require.Equal(t, inCoin, resp.Incoming)
	require.Equal(t, sdk.NewCoin("usdt", math.NewInt(1)), resp.Fee)
	require.Equal(t, creator, resp.Receiver)
	require.Nil(t, resp.IncomingMax)

	// get the pool from the keeper, so we can compare new state to expected values
	pool, found := dexKeeper.GetPool(ctx,
		poolId,
	)

	// make sure the pool was found
	require.True(t, found)

	// new abc is old abc - coins send to the sender
	newPoolBase := createPoolBase.Sub(outgoing)
	// new usdt is old usdt + incoming coins
	newPoolQuote := createPoolQuote.Add(inCoin)
	// LP Tokens doesn't change
	newPoolLPToken := createPoolExpectedLPCoin

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

func TestMsgServerSwapExactOut_EmptyReceiverUsesSender(t *testing.T) {
	// Test case: swap exact out with empty receiver, should use sender as receiver

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
	createPoolBase := sample.Coin("abc", 542000000)
	createPoolQuote := sample.Coin("usdt", 112000000)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 246381817)

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

	// how much we will swap
	outgoing := sample.Coin("usdt", 2019024)

	// how much we will swap
	inCoin, _, errCalcSwap := keeper.CalculateSwapExactOutAmount(&pool, outgoing)

	// make sure there is no error
	require.NoError(t, errCalcSwap)

	// check if the swap amount is correct
	inAbc := sample.Coin("abc", 10000002)
	require.Equal(t, inAbc, inCoin)

	// extract the pool address from the pool account
	poolAddress := types.GetPoolAddress(pool.PoolId)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, inAbc).
		Return(true).
		Times(1)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(2)

	// code will send the incoming coins
	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(inAbc)).
		Return(nil).
		Times(1)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), poolAddress, outgoing).
		Return(true).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), poolAddress, signer, sdk.NewCoins(outgoing)).
		Return(nil).
		Times(1)

	// create a "swap-exact-out" message with empty receiver
	txSwapExactOutMsg := &types.MsgSwapExactOut{
		Signer:   creator,
		Outgoing: outgoing,
		PoolId:   pool.PoolId,
		Receiver: "", // Empty, should use sender
	}

	// make rpc call to swap coins
	resp, err := server.SwapExactOut(ctx, txSwapExactOutMsg)

	// make sure there is no error
	require.NoError(t, err)

	poolId := pool.PoolId
	// check the response, receiver should be sender
	require.Equal(t, poolId, resp.PoolId)
	require.Equal(t, outgoing, resp.Outgoing)
	require.Equal(t, inCoin, resp.Incoming)
	require.Equal(t, sdk.NewCoin("abc", math.NewInt(50000)), resp.Fee)
	require.Equal(t, creator, resp.Receiver)
	require.Nil(t, resp.IncomingMax)

	// get the pool from the keeper, so we can compare new state to expected values
	pool, found := dexKeeper.GetPool(ctx,
		poolId,
	)

	// make sure the pool was found
	require.True(t, found)

	// new base is old base + inCoin
	newPoolBase := createPoolBase.Add(inCoin)
	// new quote is old quote - coins sent to the sender
	newPoolQuote := createPoolQuote.Sub(outgoing)
	// LP tokens doesn't change
	newPoolLPToken := createPoolExpectedLPCoin

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

func TestMsgServerSwapExactOut_Receiver(t *testing.T) {
	// Test case: swap exact number of coins in a pool with a receiver

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create a sample recipient address
	receiverAddress := sample.AccAddress()
	receiver := sdk.MustAccAddressFromBech32(receiverAddress)

	// create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 1000)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 316)

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

	// how much we will swap
	outgoing := sample.Coin("usdt", 91)

	// how much we will swap
	inCoin, _, errCalcSwap := keeper.CalculateSwapExactOutAmount(&pool, outgoing)

	// make sure there is no error
	require.NoError(t, errCalcSwap)

	// check if the swap amount is correct
	inAbc := sample.Coin("abc", 10)
	require.Equal(t, inAbc, inCoin)

	// extract the pool address from the pool account
	poolAddress := types.GetPoolAddress(pool.PoolId)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, inAbc).
		Return(true).
		Times(1)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(2)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), poolAddress, outgoing).
		Return(true).
		Times(1)

	// code will send the incoming coins
	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(inAbc)).
		Return(nil).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), poolAddress, receiver, sdk.NewCoins(outgoing)).
		Return(nil).
		Times(1)

	// create a "swap-exact-out" message
	txSwapExactOutMsg := &types.MsgSwapExactOut{
		Signer:   creator,
		Outgoing: outgoing,
		PoolId:   pool.PoolId,
		Receiver: receiverAddress,
	}

	// make rpc call to swap coins
	resp, err := server.SwapExactOut(ctx, txSwapExactOutMsg)

	// make sure there is no error
	require.NoError(t, err)

	poolId := pool.PoolId
	// check the response
	require.Equal(t, poolId, resp.PoolId)
	require.Equal(t, outgoing, resp.Outgoing)
	require.Equal(t, inCoin, resp.Incoming)
	require.Equal(t, sdk.NewCoin("abc", math.NewInt(1)), resp.Fee)
	require.Equal(t, receiverAddress, resp.Receiver)
	require.Nil(t, resp.IncomingMax)

	// get the pool from the keeper, so we can compare new state to expected values
	pool, found := dexKeeper.GetPool(ctx,
		poolId,
	)

	// make sure the pool was found
	require.True(t, found)

	// new base is old base + inCoin
	newPoolBase := createPoolBase.Add(inCoin)
	// new quote is old quote - coins sent to the sender
	newPoolQuote := createPoolQuote.Sub(outgoing)
	// LP tokens doesn't change
	newPoolLPToken := createPoolExpectedLPCoin

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

func TestMsgServerSwapExactOut_CalculateSwapExactOutAmount(t *testing.T) {
	// Create a very small pool to make it easier to trigger minimum fee errors
	pool := types.Pool{
		LpToken: sdk.NewInt64Coin("zp1", 246),
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 542),
			sdk.NewInt64Coin("usdt", 112),
		},
		Fee: 50000,
	}

	t.Run("outgoing_that_would_require_less_than_MinSwapFee_as_fee", func(t *testing.T) {
		// Use a tiny outgoing amount that will result in an incoming amount less than MinSwapFee
		outgoing := sdk.NewInt64Coin("usdt", 1)
		_, _, err := keeper.CalculateSwapExactOutAmount(&pool, outgoing)
		require.NoError(t, err)
	})

	t.Run("outgoing_that_results_in_fee_==_MinSwapFee", func(t *testing.T) {
		// Use an amount that should result in exactly MinSwapFee
		outgoing := sdk.NewInt64Coin("usdt", 2)
		inCoin, fee, err := keeper.CalculateSwapExactOutAmount(&pool, outgoing)
		require.NoError(t, err)
		// The fee is calculated based on the incoming amount and pool fee rate
		// For this pool with fee rate 50000 (0.5%), the fee will be 10204
		require.Equal(t, sdk.NewInt64Coin("abc", 9), fee)
		require.True(t, inCoin.Amount.GT(math.ZeroInt()))
	})

	t.Run("outgoing_just_above_MinSwapFee", func(t *testing.T) {
		// Use an amount that should result in a fee just above MinSwapFee
		outgoing := sdk.NewInt64Coin("usdt", 3)
		inCoin, fee, err := keeper.CalculateSwapExactOutAmount(&pool, outgoing)
		require.NoError(t, err)
		// The fee should be greater than MinSwapFee (1) due to the pool's fee rate
		require.True(t, fee.Amount.GT(math.NewInt(types.MinSwapFee)))
		require.True(t, inCoin.Amount.GT(math.ZeroInt()))
	})

	t.Run("large_outgoing_amount", func(t *testing.T) {
		// Use a large but reasonable amount that won't exceed pool liquidity
		outgoing := sdk.NewInt64Coin("usdt", 5)
		inCoin, fee, err := keeper.CalculateSwapExactOutAmount(&pool, outgoing)
		require.NoError(t, err)
		// The fee should be significantly larger than MinSwapFee
		require.True(t, fee.Amount.GT(math.NewInt(types.MinSwapFee)))
		require.True(t, inCoin.Amount.GT(math.ZeroInt()))
	})
}

func TestMsgServerSwapExactOut_SwapBase_IncomingMax_Valid(t *testing.T) {
	// Test case: swap the exact number of coins in a pool with incoming max

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
	createPoolQuote := sample.Coin("usdt", 1000)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 316)

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

	// how much we will swap
	outgoing := sample.Coin("usdt", 91)

	// how much we will swap
	inCoin, _, errCalcSwap := keeper.CalculateSwapExactOutAmount(&pool, outgoing)

	// make sure there is no error
	require.NoError(t, errCalcSwap)

	// check if the swap amount is correct
	inAbc := sample.Coin("abc", 10)
	require.Equal(t, inAbc, inCoin)

	// set incoming max
	incomingMax := sample.Coin("abc", 10)

	// extract the pool address from the pool account
	poolAddress := types.GetPoolAddress(pool.PoolId)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, inAbc).
		Return(true).
		Times(1)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(2)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), poolAddress, outgoing).
		Return(true).
		Times(1)

	// code will send the incoming coins
	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(inAbc)).
		Return(nil).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), poolAddress, signer, sdk.NewCoins(outgoing)).
		Return(nil).
		Times(1)

	// create a "swap-exact-out" message
	txSwapExactOutMsg := &types.MsgSwapExactOut{
		Signer:      creator,
		Outgoing:    outgoing,
		PoolId:      pool.PoolId,
		IncomingMax: &incomingMax,
	}

	// make rpc call to swap coins
	resp, err := server.SwapExactOut(ctx, txSwapExactOutMsg)

	// make sure there is no error
	require.NoError(t, err)

	poolId := pool.PoolId
	// check the response
	require.Equal(t, poolId, resp.PoolId)
	require.Equal(t, outgoing, resp.Outgoing)
	require.Equal(t, inCoin, resp.Incoming)
	require.Equal(t, sdk.NewCoin("abc", math.NewInt(1)), resp.Fee)
	require.Equal(t, creator, resp.Receiver)
	require.Equal(t, incomingMax, *resp.IncomingMax)

	// get the pool from the keeper, so we can compare new state to expected values
	pool, found := dexKeeper.GetPool(ctx,
		poolId,
	)

	// make sure the pool was found
	require.True(t, found)

	// new base is old base + inCoin
	newPoolBase := createPoolBase.Add(inCoin)
	// new quote is old quote - coins sent to the sender
	newPoolQuote := createPoolQuote.Sub(outgoing)
	// LP tokens doesn't change
	newPoolLPToken := createPoolExpectedLPCoin

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

func TestMsgServerSwapExactOut_SwapBase_IncomingMax_GreaterThanInCoin(t *testing.T) {
	// Test case: swap exact number of coins in a pool with incoming max greater than inCoin

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
	createPoolQuote := sample.Coin("usdt", 1000)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 316)

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

	// how much we will swap
	outgoing := sample.Coin("usdt", 91)

	// how much we will swap
	inCoin, _, errCalcSwap := keeper.CalculateSwapExactOutAmount(&pool, outgoing)

	// make sure there is no error
	require.NoError(t, errCalcSwap)

	// check if the swap amount is correct
	inAbc := sample.Coin("abc", 10)
	require.Equal(t, inAbc, inCoin)

	// set incoming max
	incomingMax := sample.Coin("abc", 30)

	// extract the pool address from the pool account
	poolAddress := types.GetPoolAddress(pool.PoolId)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, inAbc).
		Return(true).
		Times(1)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(2)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), poolAddress, outgoing).
		Return(true).
		Times(1)

	// code will send the incoming coins
	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(inAbc)).
		Return(nil).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), poolAddress, signer, sdk.NewCoins(outgoing)).
		Return(nil).
		Times(1)

	// create a "swap-exact-out" message
	txSwapExactOutMsg := &types.MsgSwapExactOut{
		Signer:      creator,
		Outgoing:    outgoing,
		PoolId:      pool.PoolId,
		IncomingMax: &incomingMax,
	}

	// make rpc call to swap coins
	resp, err := server.SwapExactOut(ctx, txSwapExactOutMsg)

	// make sure there is no error
	require.NoError(t, err)

	poolId := pool.PoolId
	// check the response
	require.Equal(t, poolId, resp.PoolId)
	require.Equal(t, outgoing, resp.Outgoing)
	require.Equal(t, inCoin, resp.Incoming)
	require.Equal(t, sdk.NewCoin("abc", math.NewInt(1)), resp.Fee)
	require.Equal(t, creator, resp.Receiver)
	require.Equal(t, incomingMax, *resp.IncomingMax)

	// get the pool from the keeper, so we can compare new state to expected values
	pool, found := dexKeeper.GetPool(ctx,
		poolId,
	)

	// make sure the pool was found
	require.True(t, found)

	// new base is old base + inCoin
	newPoolBase := createPoolBase.Add(inCoin)
	// new quote is old quote - coins sent to the sender
	newPoolQuote := createPoolQuote.Sub(outgoing)
	// LP tokens doesn't change
	newPoolLPToken := createPoolExpectedLPCoin

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

// Negative test cases

func TestMsgServerSwapExactOut_AmountMustBePositive(t *testing.T) {
	// Test case: try to swap exact out with zero amounts

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

	outgoing := sample.Coin("abc", 0)

	// create a "swap-exact-out" message
	txSwapExactOutMsg := &types.MsgSwapExactOut{
		Signer:   creator,
		Outgoing: outgoing,
		PoolId:   pool.PoolId,
	}

	// make rpc call to swap
	_, err := server.SwapExactOut(ctx, txSwapExactOutMsg)

	require.Error(t, err)

	// check if the error is correct
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)

	// check an error message
	require.Equal(t,
		fmt.Sprintf(
			"Outgoing amount must be positive (%s%s): invalid request",
			outgoing.Amount.String(),
			outgoing.Denom,
		),
		err.Error(),
	)
}

func TestMsgServerSwapExactOut_PoolNotFound(t *testing.T) {
	// Test case: try to swap exact out in a pool that does not exist

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

	outgoing := sample.Coin("abc", 10)

	// create a "swap-exact-out" message
	txSwapExactOutMsg := &types.MsgSwapExactOut{
		Signer:   creator,
		Outgoing: outgoing,
		PoolId:   "zp1",
	}

	// make rpc call to swap
	_, err := srv.SwapExactOut(ctx, txSwapExactOutMsg)

	require.Error(t, err)

	// check if the error is correct
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)

	// check an error message
	require.Equal(
		t,
		"Liquidity pool (zp1) can not be found: invalid request",
		err.Error(),
	)
}

func TestMsgServerSwapExactOut_InvalidOutgoingCoin(t *testing.T) {
	// Test case: try to swap exact out in a pool with invalid coin

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

	outgoing := sample.Coin("invalid", 10)

	// create a "swap-exact-out" message
	txSwapExactOutMsg := &types.MsgSwapExactOut{
		Signer:   creator,
		Outgoing: outgoing,
		PoolId:   pool.PoolId,
	}

	// make rpc call to swap
	_, err := server.SwapExactOut(ctx, txSwapExactOutMsg)

	require.Error(t, err)

	// check if the error is correct
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)

	// check an error message
	require.Equal(t,
		fmt.Sprintf(
			"Invalid outgoing coin (%s), this pool only supports base (%s) and quote (%s) tokens: invalid request",
			outgoing.Denom,
			createPoolBase.Denom,
			createPoolQuote.Denom,
		),
		err.Error(),
	)
}

func TestCalculateSwapExactOutAmount_InvalidDenom(t *testing.T) {
	// Test case: calculate swap exact out if outgoing is not in pool

	pool := types.Pool{
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 542),
			sdk.NewInt64Coin("usdt", 112),
		},
		Fee: 500,
	}

	// outgoing not in the pool
	outgoing := sdk.NewInt64Coin("btc", 2)

	_, _, err := keeper.CalculateSwapExactOutAmount(&pool, outgoing)
	require.Error(t, err)

	// check an error message
	require.Equal(t,
		fmt.Sprintf(
			"Invalid outgoing coin (%s), this pool only supports (%s) and (%s) tokens: invalid request",
			outgoing.Denom,
			pool.Coins[0].Denom,
			pool.Coins[1].Denom,
		),
		err.Error(),
	)
}

func TestMsgServerSwapExactOut_CalculateSwapExactOutAmount_OutgoingZero(t *testing.T) {
	// Test case: calculate swap exact out if outgoing is zero

	// create a pool
	pool := types.Pool{
		LpToken: sdk.NewInt64Coin("zp1", 10),
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 10),
			sdk.NewInt64Coin("usdt", 10),
		},
		Fee: 50000,
	}

	// outgoing quote to swap
	outgoing := sdk.NewInt64Coin("usdt", 0)

	_, _, err := keeper.CalculateSwapExactOutAmount(&pool, outgoing)

	require.Error(t, err)

	// check if the error is correct
	require.Equal(
		t,
		"Invalid new abc token balance: 0, has to be positive, after fee.: invalid request",
		err.Error(),
	)
}

func TestMsgServerSwapExactOut_CalculateSwapExactOutAmount_OutgoingTooLarge(t *testing.T) {
	// Test case: calculate swap exact out if outgoing is too large

	pool := types.Pool{
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 542),
			sdk.NewInt64Coin("usdt", 112),
		},
		Fee: 500,
	}

	// outgoing too large
	outgoing := sdk.NewInt64Coin("usdt", 1112)

	_, _, err := keeper.CalculateSwapExactOutAmount(&pool, outgoing)
	require.Error(t, err)

	// check if the error is correct
	require.Equal(
		t,
		"Outgoing amount (1112) must be less than pool's quote coin amount (112): invalid request",
		err.Error(),
	)
}

func TestMsgServerSwapExactOut_CalculateSwapExactOutAmount_InvalidPoolBalances(t *testing.T) {
	// Test case: calculate swap exact out if pool balances are invalid

	pool := types.Pool{
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 112),
			sdk.NewInt64Coin("usdt", 0),
		},
		Fee: 500,
	}

	outgoing := sdk.NewInt64Coin("usdt", 1)

	_, _, err := keeper.CalculateSwapExactOutAmount(&pool, outgoing)
	require.Error(t, err)

	// check if the error is correct
	require.Equal(
		t,
		"Outgoing amount (1) must be less than pool's quote coin amount (0): invalid request",
		err.Error(),
	)
}

func TestMsgServerSwapExactOut_CalculateSwapAmount_InvalidIncomingCoin(t *testing.T) {
	// Test case: calculate swap amount if incoming coin is invalid

	// create a pool
	pool := types.Pool{
		LpToken: sdk.NewInt64Coin("zp1", 10),
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("base", 10),
			sdk.NewInt64Coin("quote", 10),
		},
		Fee: 50000,
	}

	// outgoing coin to swap
	outgoing := sdk.NewInt64Coin("invalid", 10)

	_, _, err := keeper.CalculateSwapExactOutAmount(&pool, outgoing)

	require.Error(t, err)
	require.Equal(
		t,
		fmt.Sprintf(
			"Invalid outgoing coin (%s), this pool only supports (%s) and (%s) tokens: invalid request",
			outgoing.Denom,
			pool.Coins[0].Denom,
			pool.Coins[1].Denom,
		),
		err.Error(),
	)
}

func TestMsgServerSwapExactOut_SwapBase_OutgoingMin_Zero(t *testing.T) {
	// Test case: try to swap exact out if the outgoing minimum is set to zero

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

	// how much we will swap
	outgoing := sample.Coin("usdt", 10)

	// set incoming max to zero
	incomingMax := sample.Coin("abc", 0)

	// create a "swap-exact-out" message
	txSwapExactOutMsg := &types.MsgSwapExactOut{
		Signer:      creator,
		Outgoing:    outgoing,
		PoolId:      pool.PoolId,
		IncomingMax: &incomingMax,
	}

	// make rpc call to swap
	_, err := server.SwapExactOut(ctx, txSwapExactOutMsg)

	require.Error(t, err)

	// check if the error is correct
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)

	// check an error message
	require.Equal(t,
		fmt.Sprintf(
			"Incoming maximum amount must be positive (%s): invalid request",
			incomingMax.String(),
		),
		err.Error(),
	)
}

func TestMsgServerSwapExactOut_SwapBase_OutgoingMin_InvalidDenom(t *testing.T) {
	// Test case: try to swap exact out if IncomingMax.Denom != inCoin.Denom

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

	// how much we will swap
	outgoing := sample.Coin("usdt", 10)

	// how much we will swap
	inCoin, _, errCalcSwap := keeper.CalculateSwapExactOutAmount(&pool, outgoing)

	// make sure there is no error
	require.NoError(t, errCalcSwap)

	// set incoming max to a different denom than incoming
	incomingMax := sample.Coin("usdt", 10)

	// create a "swap-exact-out" message
	txSwapExactOutMsg := &types.MsgSwapExactOut{
		Signer:      creator,
		Outgoing:    outgoing,
		PoolId:      pool.PoolId,
		IncomingMax: &incomingMax,
	}

	// make rpc call to swap
	_, err := server.SwapExactOut(ctx, txSwapExactOutMsg)

	require.Error(t, err)

	// check if the error is correct
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)

	// check an error message
	require.Equal(t,
		fmt.Sprintf(
			"Incoming maximum denom (%s) must be the same as incoming denom (%s): invalid request",
			incomingMax.Denom,
			inCoin.Denom,
		),
		err.Error(),
	)
}

func TestMsgServerSwapExactOut_SwapBase_OutgoingMin_RequiredMoreThanIncoming(t *testing.T) {
	// Test case: try to swap exact out in a pool with required amount more than incoming

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

	// how much we will swap
	outgoing := sample.Coin("usdt", 10)

	// how much we will swap
	inCoin, _, errCalcSwap := keeper.CalculateSwapExactOutAmount(&pool, outgoing)

	// make sure there is no error
	require.NoError(t, errCalcSwap)

	// set incoming max
	incomingMax := sample.Coin("abc", 1)

	// create a "swap-exact-out" message
	txSwapExactOutMsg := &types.MsgSwapExactOut{
		Signer:      creator,
		Outgoing:    outgoing,
		PoolId:      pool.PoolId,
		IncomingMax: &incomingMax,
	}

	// make rpc call to swap
	_, err := server.SwapExactOut(ctx, txSwapExactOutMsg)

	require.Error(t, err)

	// check if the error is correct
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)

	// check an error message
	require.Equal(t,
		fmt.Sprintf(
			"Required amount (%s) is more than incoming maximum amount (%s): invalid request",
			inCoin.String(),
			incomingMax.String(),
		),
		err.Error(),
	)
}

func TestMsgServerSwapExactOut_SendFromAddressToPoolFail(t *testing.T) {
	// Test case: try to swap coins in a pool if the transfer from the sender to the pool fails

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
	createPoolBase := sample.Coin("abc", 542000000)
	createPoolQuote := sample.Coin("usdt", 112000000)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 246381817)

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

	// how much we will swap
	outgoing := sample.Coin("usdt", 2019024)

	// how much we will swap
	inCoin, _, errCalcSwap := keeper.CalculateSwapExactOutAmount(&pool, outgoing)

	// make sure there is no error
	require.NoError(t, errCalcSwap)

	// check if the swap amount is correct
	inAbc := sample.Coin("abc", 10000002)
	require.Equal(t, inAbc, inCoin)

	// extract the pool address from the pool account
	poolAddress := types.GetPoolAddress(pool.PoolId)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, inAbc).
		Return(false).
		Times(1)

	// create a "swap" message
	txSwapMsg := &types.MsgSwapExactOut{
		Signer:   creator,
		Outgoing: outgoing,
		PoolId:   pool.PoolId,
	}

	// make rpc call to swap coins
	_, err := server.SwapExactOut(ctx, txSwapMsg)

	require.Error(t, err)

	// check an error message
	require.Equal(t,
		fmt.Sprintf(
			"SwapExactOut: Failed to send %s coins: SendFromAddressToPool: Insufficient funds in sender %s to send %s coins to poolID %s (address: %s): insufficient funds",
			inAbc,
			signer.String(),
			inAbc,
			pool.PoolId,
			poolAddress.String(),
		),
		err.Error(),
	)
}

func TestMsgServerSwapExactOut_SendFromPoolToAddressFail(t *testing.T) {
	// Test case: try to swap coins in a pool if the transfer from the pool to the sender fails

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
	createPoolBase := sample.Coin("abc", 542000000)
	createPoolQuote := sample.Coin("usdt", 112000000)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 246381817)

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

	// how much we will swap
	outgoing := sample.Coin("usdt", 2019024)

	// how much we will swap
	inCoin, _, errCalcSwap := keeper.CalculateSwapExactOutAmount(&pool, outgoing)

	// make sure there is no error
	require.NoError(t, errCalcSwap)

	// check if the swap amount is correct
	inAbc := sample.Coin("abc", 10000002)
	require.Equal(t, inAbc, inCoin)

	// extract the pool address from the pool account
	poolAddress := types.GetPoolAddress(pool.PoolId)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, inAbc).
		Return(true).
		Times(1)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(2)

	// code will send the incoming coins
	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(inAbc)).
		Return(nil).
		Times(1)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), poolAddress, outgoing).
		Return(false).
		Times(1)

	// create a "swap" message
	txSwapMsg := &types.MsgSwapExactOut{
		Signer:   creator,
		Outgoing: outgoing,
		PoolId:   pool.PoolId,
	}

	// make rpc call to swap coins
	_, err := server.SwapExactOut(ctx, txSwapMsg)

	require.Error(t, err)

	// check an error message
	require.Equal(t,
		fmt.Sprintf(
			"SwapExactOut: Failed to send coins %s: SendFromPoolToAddress: Insufficient funds in poolID %s (address: %s) to send %s coins to receiver %s: insufficient funds",
			outgoing,
			pool.PoolId,
			poolAddress.String(),
			outgoing,
			signer.String(),
		),
		err.Error(),
	)
}

func TestMsgServerSwapExactOut_InvalidSignerAddress(t *testing.T) {
	// Test case: try to swap exact out with an invalid signer address

	// Create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a sample signer address (valid for pool setup, but we'll override msg.Signer)
	creator := sample.AccAddress()
	signer := sdk.MustAccAddressFromBech32(creator) // Used for pool setup

	// Create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 1000)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 316)

	// Create dex keeper with pool mock
	server, _, ctx, _, _, _, _ := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// Invalid Bech32 address
	invalidSigner := "invalid-address-not-bech32"

	// Create swap message with invalid signer
	txSwapMsg := &types.MsgSwapExactOut{
		Signer:   invalidSigner,
		PoolId:   "zp1", // Matches the pool created by ServerDexKeeperWithPoolMock
		Outgoing: sample.Coin("usdt", 10),
	}

	// Call SwapExactOut
	_, err := server.SwapExactOut(ctx, txSwapMsg)

	// Verify error
	require.Error(t, err)
	require.Contains(t, err.Error(), "decoding bech32 failed")
}

func TestMsgServerSwapExactOut_PoolZeroBalance(t *testing.T) {
	// Test case: try to swap exact out with a pool having a zero balance for one token

	// Create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a sample signer address
	creator := sample.AccAddress()

	// Create a pool with a zero balance for one token
	pool := types.Pool{
		PoolId:  "zp1",
		LpToken: sdk.NewInt64Coin("zp1", 100),
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 100),
			sdk.NewInt64Coin("usdt", 0), // Zero balance
		},
		Fee: 50000,
	}

	// Create dex keeper with a mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, nil)
	srv := keeper.NewMsgServerImpl(k)

	// Set the pool in the keeper
	k.SetPool(ctx, pool)

	// Define the outgoing coin
	outgoing := sample.Coin("usdt", 10)

	// Create a swap message
	txSwapMsg := &types.MsgSwapExactOut{
		Signer:   creator,
		PoolId:   pool.PoolId,
		Outgoing: outgoing,
	}

	// Call SwapExactOut
	_, err := srv.SwapExactOut(ctx, txSwapMsg)

	// Verify the error
	require.Error(t, err)
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
	require.Equal(t,
		fmt.Sprintf(
			"Invalid liquidity pool: %s token: %s balance: %s, has to be positive.: invalid request",
			pool.PoolId,
			"usdt",
			"0",
		),
		err.Error(),
	)
}

func TestMsgServerSwapExactOut_CalculateSwapExactOutAmount_NonPositiveQuoteBalance(t *testing.T) {
	// Test case: calculate swap exact out with outgoing amount equal to pool's quote balance

	// Create a pool
	pool := types.Pool{
		LpToken: sdk.NewInt64Coin("zp1", 246),
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 542),
			sdk.NewInt64Coin("usdt", 112),
		},
		Fee: 50000,
	}

	// Outgoing amount equal to the pool's quote balance
	outgoing := sdk.NewInt64Coin("usdt", 112)

	// Calculate swap
	_, _, err := keeper.CalculateSwapExactOutAmount(&pool, outgoing)
	require.Error(t, err)

	// Check the error message
	require.Equal(
		t,
		"Outgoing amount (112) must be less than pool's quote coin amount (112): invalid request",
		err.Error(),
	)
}

func TestMsgServerSwapExactOut_OutgoingTooLarge(t *testing.T) {
	// Test case: try to swap exact out with outgoing amount larger than pool balance

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
	createPoolQuote := sample.Coin("usdt", 1000)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 316)

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

	// outgoing larger than pool's quote coin
	outgoing := sample.Coin("usdt", 1001)

	// create a "swap-exact-out" message
	txSwapExactOutMsg := &types.MsgSwapExactOut{
		Signer:   creator,
		Outgoing: outgoing,
		PoolId:   pool.PoolId,
	}

	// make rpc call to swap
	_, err := server.SwapExactOut(ctx, txSwapExactOutMsg)

	require.Error(t, err)

	// check if the error is correct
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)

	// check an error message
	require.Equal(t,
		fmt.Sprintf(
			"Outgoing amount (%s) must be less than pool's quote coin amount (%s): invalid request",
			outgoing.Amount.String(),
			createPoolQuote.Amount.String(),
		),
		err.Error(),
	)
}

func TestMsgServerSwapExactOut_SmallOutgoingIncomingAfterFeeNotPositive(t *testing.T) {
	// Test case: try to swap small exact out amount leading to zero incoming after fee (due to integer truncation)

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation with imbalance (quote > base + 1)
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 200)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 141) // approx sqrt(100*200)

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

	// small outgoing
	outgoing := sample.Coin("usdt", 1)

	// create a "swap-exact-out" message
	txSwapExactOutMsg := &types.MsgSwapExactOut{
		Signer:   creator,
		Outgoing: outgoing,
		PoolId:   pool.PoolId,
	}

	// make rpc call to swap
	_, err := server.SwapExactOut(ctx, txSwapExactOutMsg)

	require.Error(t, err)

	// check if the error is correct
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)

	// check an error message
	require.Equal(t,
		fmt.Sprintf(
			"Invalid new %s token balance: %s, has to be positive, after fee.: invalid request",
			createPoolBase.Denom,
			"0",
		),
		err.Error(),
	)
}
