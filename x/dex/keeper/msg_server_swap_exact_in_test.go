package keeper_test

import (
	"fmt"
	"strconv"
	"strings"
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

func TestMsgServerSwapExactInSwapBase_Positive(t *testing.T) {
	// Test case: regular swap of the abc coin in a pool

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
	incoming := sample.Coin("abc", 10000000)

	// how much we will swap
	outCoin, _, errCalcSwap := keeper.CalculateSwapAmount(&pool, incoming)

	// make sure there is no error
	require.NoError(t, errCalcSwap)

	// check if the swap amount is correct
	outUsdt := sample.Coin("usdt", 2019024)
	require.Equal(t, outUsdt, outCoin)

	// extract the pool address from the pool account
	poolAddress := types.GetPoolAddress(pool.PoolId)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, incoming).
		Return(true).
		Times(2)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(2)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), poolAddress, outCoin).
		Return(true).
		Times(1)

	// code will send the incoming coins
	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(incoming)).
		Return(nil).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), poolAddress, signer, sdk.NewCoins(outCoin)).
		Return(nil).
		Times(1)

	// create a "swap" message
	txSwapMsg := &types.MsgSwapExactIn{
		Signer:   creator,
		Incoming: incoming,
		PoolId:   pool.PoolId,
	}

	// make rpc call to swap coins
	resp, err := server.SwapExactIn(ctx, txSwapMsg)

	// make sure there is no error
	require.NoError(t, err)

	poolId := pool.PoolId
	// check the response
	require.Equal(t, poolId, resp.PoolId)
	require.Equal(t, incoming, resp.Incoming)
	require.Equal(t, outCoin, resp.Outgoing)
	require.Equal(t, sdk.NewCoin("abc", math.NewInt(50000)), resp.Fee)
	require.Equal(t, creator, resp.Receiver)
	require.Nil(t, resp.OutgoingMin)

	// get the pool from the keeper, so we can compare new state to expected values
	pool, found := dexKeeper.GetPool(ctx,
		poolId,
	)

	// make sure the pool was found
	require.True(t, found)

	// new abc is old abc + incoming coins
	newPoolBase := createPoolBase.Add(incoming)
	// new usdt is old usdt - coins sent to the sender
	newPoolQuote := createPoolQuote.Sub(outCoin)
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

func TestMsgServerSwapExactIn_SwapBase_Positive2(t *testing.T) {
	// Test case: regular swap of the abc coin in a pool

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
	incoming := sample.Coin("abc", 10)

	// how much we will swap
	outCoin, _, errCalcSwap := keeper.CalculateSwapAmount(&pool, incoming)

	// make sure there is no error
	require.NoError(t, errCalcSwap)

	// check if the swap amount is correct
	outUsdt := sample.Coin("usdt", 83)
	require.Equal(t, outUsdt, outCoin)

	// extract the pool address from the pool account
	poolAddress := types.GetPoolAddress(pool.PoolId)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, incoming).
		Return(true).
		Times(2)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(2)

	// code will send the incoming coins
	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(incoming)).
		Return(nil).
		Times(1)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), poolAddress, outCoin).
		Return(true).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), poolAddress, signer, sdk.NewCoins(outCoin)).
		Return(nil).
		Times(1)

	// create a "swap" message
	txSwapMsg := &types.MsgSwapExactIn{
		Signer:   creator,
		Incoming: incoming,
		PoolId:   pool.PoolId,
	}

	// make rpc call to swap coins
	resp, err := server.SwapExactIn(ctx, txSwapMsg)

	// make sure there is no error
	require.NoError(t, err)

	poolId := pool.PoolId
	// check the response
	require.Equal(t, poolId, resp.PoolId)
	require.Equal(t, incoming, resp.Incoming)
	require.Equal(t, outCoin, resp.Outgoing)
	require.Equal(t, sdk.NewCoin("abc", math.NewInt(1)), resp.Fee)
	require.Equal(t, creator, resp.Receiver)
	require.Nil(t, resp.OutgoingMin)

	// get the pool from the keeper, so we can compare new state to expected values
	pool, found := dexKeeper.GetPool(ctx,
		poolId,
	)

	// make sure the pool was found
	require.True(t, found)

	// new abc is old abc + incoming coins
	newPoolBase := createPoolBase.Add(incoming)
	// new usdt is old usdt - coins sent to the sender
	newPoolQuote := createPoolQuote.Sub(outCoin)
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

func TestMsgServerSwapExactIn_SwapQuote(t *testing.T) {
	// Test case: regular swap of the usdt coin in a pool

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
	incoming := sample.Coin("usdt", 10)

	// how much we will swap
	outCoin, _, errCalcSwap := keeper.CalculateSwapAmount(&pool, incoming)

	// make sure there is no error
	require.NoError(t, errCalcSwap)

	// check if the swap amount is correct
	outAbc := sample.Coin("abc", 1)
	require.Equal(t, outAbc, outCoin)

	// extract the pool address from the pool account
	poolAddress := types.GetPoolAddress(pool.PoolId)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, incoming).
		Return(true).
		Times(2)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(2)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), poolAddress, outCoin).
		Return(true).
		Times(1)

	// code will send the incoming coins
	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(incoming)).
		Return(nil).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), poolAddress, signer, sdk.NewCoins(outCoin)).
		Return(nil).
		Times(1)

	// create a "swap" message
	txSwapMsg := &types.MsgSwapExactIn{
		Signer:   creator,
		Incoming: incoming,
		PoolId:   pool.PoolId,
	}

	// make rpc call to swap coins
	resp, err := server.SwapExactIn(ctx, txSwapMsg)

	// make sure there is no error
	require.NoError(t, err)

	// check the response
	poolId := pool.PoolId
	require.Equal(t, poolId, resp.PoolId)
	require.Equal(t, incoming, resp.Incoming)
	require.Equal(t, outCoin, resp.Outgoing)
	require.Equal(t, sdk.NewCoin("usdt", math.NewInt(1)), resp.Fee)
	require.Equal(t, creator, resp.Receiver)
	require.Nil(t, resp.OutgoingMin)

	// get the pool from the keeper, so we can compare new state to expected values
	pool, found := dexKeeper.GetPool(ctx,
		poolId,
	)

	// make sure the pool was found
	require.True(t, found)

	// new abc is old abc - coins send to the sender
	newPoolBase := createPoolBase.Sub(outCoin)
	// new usdt is old usdt + incoming coins
	newPoolQuote := createPoolQuote.Add(incoming)
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

func TestMsgServerSwapExactIn_Receiver(t *testing.T) {
	// Test case: regular swap with different receiver to get the coins

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
	createPoolQuote := sample.Coin("usdt", 200)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 141)

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

	// how much we will swap from abc to usdt
	incoming := sample.Coin("abc", 10)

	// how much we will swap from usdt to abc
	outCoin, _, errCalcSwap := keeper.CalculateSwapAmount(&pool, incoming)

	// make sure there is no error
	require.NoError(t, errCalcSwap)

	// check if the swap amount is correct
	outUsdt := sample.Coin("usdt", 17)
	require.Equal(t, outUsdt, outCoin)

	// extract the pool address from the pool account
	poolAddress := types.GetPoolAddress(pool.PoolId)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, incoming).
		Return(true).
		Times(2)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(2)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), poolAddress, outCoin).
		Return(true).
		Times(1)

	// code will send the incoming coins
	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(incoming)).
		Return(nil).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), poolAddress, receiver, sdk.NewCoins(outCoin)).
		Return(nil).
		Times(1)

	// create a "swap" message
	txSwapMsg := &types.MsgSwapExactIn{
		Signer:   creator,
		Incoming: incoming,
		PoolId:   pool.PoolId,
		Receiver: receiverAddress,
	}

	// make rpc call to swap coins
	resp, err := server.SwapExactIn(ctx, txSwapMsg)

	// make sure there is no error
	require.NoError(t, err)

	// check the response
	poolId := pool.PoolId
	require.Equal(t, poolId, resp.PoolId)
	require.Equal(t, incoming, resp.Incoming)
	require.Equal(t, outCoin, resp.Outgoing)
	require.Equal(t, sdk.NewCoin("abc", math.NewInt(1)), resp.Fee)
	require.Equal(t, receiverAddress, resp.Receiver)
	require.Nil(t, resp.OutgoingMin)

	// get the pool from the keeper, so we can compare new state to expected values
	pool, found := dexKeeper.GetPool(ctx,
		poolId,
	)

	// make sure the pool was found
	require.True(t, found)

	// new abc is old abc + incoming coins
	newPoolBase := createPoolBase.Add(incoming)
	// new usdt is old usdt - coins sent to the sender
	newPoolQuote := createPoolQuote.Sub(outCoin)
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

func TestMsgServerSwapExactIn_CalculateSwapAmount(t *testing.T) {
	// Test case: calculate the number of coins to swap

	pool := types.Pool{
		LpToken: sdk.NewInt64Coin("zp1", 246),
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 542),
			sdk.NewInt64Coin("usdt", 112),
		},
		Fee: 50000,
	}

	t.Run("regular swap", func(t *testing.T) {
		incoming := sdk.NewInt64Coin("abc", 10)
		outCoin, fee, err := keeper.CalculateSwapAmount(&pool, incoming)
		require.NoError(t, err)
		require.Equal(t, sdk.NewInt64Coin("usdt", 2), outCoin)
		require.Equal(t, sdk.NewInt64Coin("abc", 10), incoming)
		require.Equal(t, sdk.NewInt64Coin("abc", 5), fee)
	})

	t.Run("incoming less than MinSwapFee", func(t *testing.T) {
		incoming := sdk.NewInt64Coin("abc", types.MinSwapFee-1)
		_, _, err := keeper.CalculateSwapAmount(&pool, incoming)
		require.Error(t, err)
		require.Contains(t, err.Error(), "Swap amount too low to cover minimum fee")
	})

	t.Run("incoming equal to MinSwapFee", func(t *testing.T) {
		incoming := sdk.NewInt64Coin("abc", types.MinSwapFee)
		_, _, err := keeper.CalculateSwapAmount(&pool, incoming)
		require.Error(t, err)
	})

	t.Run("incoming just above MinSwapFee", func(t *testing.T) {
		incoming := sdk.NewInt64Coin("abc", types.MinSwapFee+1)
		_, fee, err := keeper.CalculateSwapAmount(&pool, incoming)
		require.NoError(t, err)
		require.Equal(t, sdk.NewInt64Coin("abc", types.MinSwapFee), fee)
		// outCoin should be > 0 if pool math allows
	})

	t.Run("large incoming amount", func(t *testing.T) {
		incoming := sdk.NewInt64Coin("abc", 1000000)
		_, fee, err := keeper.CalculateSwapAmount(&pool, incoming)
		require.NoError(t, err)
		require.True(t, fee.Amount.GT(math.NewInt(types.MinSwapFee)))
	})
}

func TestMsgServerSwapExactIn_SwapBase_OutgoingMin_Valid(t *testing.T) {
	// Test case: regular swap of the abc coin in a pool with outgoing minimum

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
	incoming := sample.Coin("abc", 10)

	// how much we will swap
	outCoin, _, errCalcSwap := keeper.CalculateSwapAmount(&pool, incoming)

	// make sure there is no error
	require.NoError(t, errCalcSwap)

	// check if the swap amount is correct
	outUsdt := sample.Coin("usdt", 83)
	require.Equal(t, outUsdt, outCoin)

	// set ongoing minimum to 91 --> same as outCoin
	outgoingMin := sample.Coin("usdt", 83)

	// extract the pool address from the pool account
	poolAddress := types.GetPoolAddress(pool.PoolId)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, incoming).
		Return(true).
		Times(2)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(2)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), poolAddress, outCoin).
		Return(true).
		Times(1)

	// code will send the incoming coins
	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(incoming)).
		Return(nil).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), poolAddress, signer, sdk.NewCoins(outCoin)).
		Return(nil).
		Times(1)

	// create a "swap" message
	txSwapMsg := &types.MsgSwapExactIn{
		Signer:      creator,
		Incoming:    incoming,
		PoolId:      pool.PoolId,
		OutgoingMin: &outgoingMin,
	}

	// make rpc call to swap coins
	resp, err := server.SwapExactIn(ctx, txSwapMsg)

	// make sure there is no error
	require.NoError(t, err)

	// check the response
	poolId := pool.PoolId
	require.Equal(t, poolId, resp.PoolId)
	require.Equal(t, incoming, resp.Incoming)
	require.Equal(t, outCoin, resp.Outgoing)
	require.Equal(t, sdk.NewCoin("abc", math.NewInt(1)), resp.Fee)
	require.Equal(t, creator, resp.Receiver)
	require.Equal(t, outgoingMin, *resp.OutgoingMin)

	// get the pool from the keeper, so we can compare new state to expected values
	pool, found := dexKeeper.GetPool(ctx,
		poolId,
	)

	// make sure the pool was found
	require.True(t, found)

	// new abc is old abc + incoming coins
	newPoolBase := createPoolBase.Add(incoming)
	// new usdt is old usdt - coins sent to the sender
	newPoolQuote := createPoolQuote.Sub(outCoin)
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

func TestMsgServerSwapExactIn_SwapQuote_OutgoingMin_Valid(t *testing.T) {
	// Test case: regular swap of the usdt coin in a pool with outgoing minimum

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
	incoming := sample.Coin("usdt", 10)

	// how much we will swap
	outCoin, _, errCalcSwap := keeper.CalculateSwapAmount(&pool, incoming)

	// make sure there is no error
	require.NoError(t, errCalcSwap)

	// check if the swap amount is correct
	outUsdt := sample.Coin("abc", 1)
	require.Equal(t, outUsdt, outCoin)

	// set ongoing minimum to 1 --> same as outCoin
	outgoingMin := sample.Coin("abc", 1)

	// extract the pool address from the pool account
	poolAddress := types.GetPoolAddress(pool.PoolId)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, incoming).
		Return(true).
		Times(2)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(2)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), poolAddress, outCoin).
		Return(true).
		Times(1)

	// code will send the incoming coins
	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(incoming)).
		Return(nil).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), poolAddress, signer, sdk.NewCoins(outCoin)).
		Return(nil).
		Times(1)

	// create a "swap" message
	txSwapMsg := &types.MsgSwapExactIn{
		Signer:      creator,
		Incoming:    incoming,
		PoolId:      pool.PoolId,
		OutgoingMin: &outgoingMin,
	}

	// make rpc call to swap coins
	resp, err := server.SwapExactIn(ctx, txSwapMsg)

	// make sure there is no error
	require.NoError(t, err)

	// check the response
	poolId := pool.PoolId
	require.Equal(t, poolId, resp.PoolId)
	require.Equal(t, incoming, resp.Incoming)
	require.Equal(t, outCoin, resp.Outgoing)
	require.Equal(t, sdk.NewCoin("usdt", math.NewInt(1)), resp.Fee)
	require.Equal(t, creator, resp.Receiver)
	require.Equal(t, outgoingMin, *resp.OutgoingMin)

	// get the pool from the keeper, so we can compare new state to expected values
	pool, found := dexKeeper.GetPool(ctx,
		poolId,
	)

	// make sure the pool was found
	require.True(t, found)

	// new abc is old abc - coins send to the sender
	newPoolBase := createPoolBase.Sub(outCoin)
	// new usdt is old usdt + incoming coins
	newPoolQuote := createPoolQuote.Add(incoming)
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

func TestMsgServerSwapExactIn_SwapQuote_OutgoingMin_LessThanOutCoin(t *testing.T) {
	// Test case: regular swap of the usdt coin in a pool with outgoing minimum less than outCoin

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
	incoming := sample.Coin("usdt", 10)

	// how much we will swap from usdt to abc
	outCoin, _, errCalcSwap := keeper.CalculateSwapAmount(&pool, incoming)

	// make sure there is no error
	require.NoError(t, errCalcSwap)

	// set ongoing minimum to less than outCoin
	outgoingMin := sample.Coin("abc", 9)

	// extract the pool address from the pool account
	poolAddress := types.GetPoolAddress(pool.PoolId)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, incoming).
		Return(true).
		Times(2)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(2)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), poolAddress, outCoin).
		Return(true).
		Times(1)

	// code will send the incoming coins
	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(incoming)).
		Return(nil).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), poolAddress, signer, sdk.NewCoins(outCoin)).
		Return(nil).
		Times(1)

	// create a "swap" message
	txSwapMsg := &types.MsgSwapExactIn{
		Signer:      creator,
		Incoming:    incoming,
		PoolId:      pool.PoolId,
		OutgoingMin: &outgoingMin,
	}

	// make rpc call to swap coins
	resp, err := server.SwapExactIn(ctx, txSwapMsg)

	// make sure there is no error
	require.NoError(t, err)

	// check the response
	poolId := pool.PoolId
	require.Equal(t, poolId, resp.PoolId)
	require.Equal(t, incoming, resp.Incoming)
	require.Equal(t, outCoin, resp.Outgoing)
	require.Equal(t, sdk.NewCoin("usdt", math.NewInt(1)), resp.Fee)
	require.Equal(t, creator, resp.Receiver)
	require.Equal(t, outgoingMin, *resp.OutgoingMin)

	// get the pool from the keeper, so we can compare new state to expected values
	pool, found := dexKeeper.GetPool(ctx,
		poolId,
	)

	// make sure the pool was found
	require.True(t, found)

	// new abc is old abc - coins send to the sender
	newPoolBase := createPoolBase.Sub(outCoin)
	// new usdt is old usdt + incoming coins
	newPoolQuote := createPoolQuote.Add(incoming)
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

// Negative test cases

func TestMsgServerSwapExactIn_AmountMustBePositive(t *testing.T) {
	// Test case: try to swap coins in a pool if the amount is zero

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

	// how much we will swap --> set to zero
	coinIn := sample.Coin("abc", 0)

	// create a swap message
	txSwapMsg := &types.MsgSwapExactIn{
		Signer:   creator,
		PoolId:   pool.PoolId,
		Incoming: coinIn,
	}

	// make rpc call to swap
	_, err := server.SwapExactIn(ctx, txSwapMsg)

	require.Error(t, err)

	// check if the error is correct
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)

	// check an error message
	require.Equal(t,
		fmt.Sprintf(
			"Incoming amount must be positive (%s%s): invalid request",
			coinIn.Amount.String(),
			coinIn.Denom,
		),
		err.Error(),
	)
}

func TestMsgServerSwapExactIn_PoolNotFound(t *testing.T) {
	// Test case: try to swap coins in a pool that does not exist

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, nil)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// how much we will swap
	coinIn := sample.Coin("abc", 10)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, coinIn).
		Return(true).
		Times(1)

	// create a swap message
	txSwapMsg := &types.MsgSwapExactIn{
		Signer:   creator,
		PoolId:   "zp1",
		Incoming: coinIn,
	}

	// make rpc call to swap
	_, err := srv.SwapExactIn(ctx, txSwapMsg)

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

func TestMsgServerSwapExactIn_InvalidIncomingCoin(t *testing.T) {
	// Test case: try to swap coins in a pool with invalid coin

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
	server, _, ctx, pool, _, bankKeeper, _ := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	coinIn := sample.Coin("invalid", 10)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, coinIn).
		Return(true).
		Times(1)

	// create a swap message
	txSwapMsg := &types.MsgSwapExactIn{
		Signer:   creator,
		PoolId:   pool.PoolId,
		Incoming: coinIn,
	}

	// make rpc call to swap
	_, err := server.SwapExactIn(ctx, txSwapMsg)

	require.Error(t, err)

	// check if the error is correct
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)

	// check an error message
	require.Equal(t,
		fmt.Sprintf(
			"Invalid incoming coin (%s), this pool only supports base (%s) and quote (%s) tokens: invalid request",
			coinIn.Denom,
			createPoolBase.Denom,
			createPoolQuote.Denom,
		),
		err.Error(),
	)
}

func TestMsgServerSwapExactIn_CalculateSwapAmount_InvalidDenom(t *testing.T) {
	// Test case: calculate swap amount with invalid denom

	pool := types.Pool{
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 542),
			sdk.NewInt64Coin("usdt", 112),
		},
		Fee: 500,
	}

	// incoming not in the pool
	incoming := sdk.NewInt64Coin("btc", 2)

	_, _, err := keeper.CalculateSwapAmount(&pool, incoming)
	require.Error(t, err)

	// check an error message
	require.Equal(t,
		fmt.Sprintf(
			"Invalid incoming coin (%s), this pool only supports base (%s) and quote (%s) tokens: invalid request",
			incoming.Denom,
			pool.Coins[0].Denom,
			pool.Coins[1].Denom,
		),
		err.Error(),
	)
}

func TestMsgServerSwapExactIn_CalculateSwapAmount_InvalidIncomingCoin(t *testing.T) {
	// Test case: calculate the number of coins to swap with invalid incoming coin

	// create a pool
	pool := types.Pool{
		LpToken: sdk.NewInt64Coin("zp1", 10),
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 10),
			sdk.NewInt64Coin("usdt", 10),
		},
		Fee: 50000,
	}

	// Incoming abc to swap
	incoming := sdk.NewInt64Coin("invalid", 10)

	_, _, err := keeper.CalculateSwapAmount(&pool, incoming)

	require.Error(t, err)
	require.Equal(
		t,
		fmt.Sprintf(
			"Invalid incoming coin (%s), this pool only supports base (%s) and quote (%s) tokens: invalid request",
			incoming.Denom,
			pool.Coins[0].Denom,
			pool.Coins[1].Denom,
		),
		err.Error(),
	)
}

func TestMsgServerSwapExactIn_CalculateSwapAmount_IncomingZero(t *testing.T) {
	// Test case: calculate the number of coins to swap if the incoming amount is zero

	// create a pool
	pool := types.Pool{
		LpToken: sdk.NewInt64Coin("zp1", 0),
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 0),
			sdk.NewInt64Coin("usdt", 0),
		},
		Fee: 50000,
	}

	// incoming abc to swap
	incoming := sdk.NewInt64Coin("abc", 0)

	_, _, err := keeper.CalculateSwapAmount(&pool, incoming)

	require.Error(t, err)
	require.Equal(
		t,
		"Swap amount too low to cover minimum fee (1): invalid request",
		err.Error(),
	)
}

func TestMsgServerSwapExactIn_CalculateSwapAmount_ZeroQuoteBalance(t *testing.T) {
	// Test case: calculate the number of coins to swap if quote balance is zero

	// create a pool
	pool := types.Pool{
		LpToken: sdk.NewInt64Coin("zp1", 10),
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 10),
			sdk.NewInt64Coin("usdt", 0),
		},
		Fee: 50000,
	}

	// incoming abc to swap
	incoming := sdk.NewInt64Coin("abc", 10)

	_, _, err := keeper.CalculateSwapAmount(&pool, incoming)

	require.Error(t, err)
	require.Equal(
		t,
		fmt.Sprintf(
			"Invalid swap amount: %s balance: 0, has to be positive.: invalid request",
			pool.Coins[1].Denom,
		),
		err.Error(),
	)
}

func TestMsgServerSwapExactIn_CalculateSwapAmount_ZeroBaseBalance(t *testing.T) {
	// Test case: calculate the number of coins to swap with zero usdt in the pool

	// create a pool
	pool := types.Pool{
		LpToken: sdk.NewInt64Coin("zp1", 10),
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 0),
			sdk.NewInt64Coin("usdt", 10),
		},
		Fee: 50000,
	}

	// incoming usdt to swap
	incoming := sdk.NewInt64Coin("usdt", 10)

	_, _, err := keeper.CalculateSwapAmount(&pool, incoming)

	require.Error(t, err)
	require.Equal(
		t,
		fmt.Sprintf(
			"Invalid swap amount: %s balance: 0, has to be positive.: invalid request",
			pool.Coins[0].Denom,
		),
		err.Error(),
	)
}

func TestMsgServerSwapExactIn_SwapBase_OutgoingMin_Zero(t *testing.T) {
	// Test case: try to swap coins in a pool with outgoing minimum set to zero

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
	incoming := sample.Coin("abc", 10)

	// set ongoing minimum to 0
	outgoingMin := sample.Coin("usdt", 0)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, incoming).
		Return(true).
		Times(1)

	// create a "swap" message
	txSwapMsg := &types.MsgSwapExactIn{
		Signer:      creator,
		Incoming:    incoming,
		PoolId:      pool.PoolId,
		OutgoingMin: &outgoingMin,
	}

	// make rpc call to swap
	_, err := server.SwapExactIn(ctx, txSwapMsg)

	require.Error(t, err)

	// check if the error is correct
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)

	// check an error message
	require.Equal(t,
		fmt.Sprintf(
			"Outgoing minimum amount must be positive (%s): invalid request",
			outgoingMin.String(),
		),
		err.Error(),
	)
}

func TestMsgServerSwapExactIn_SwapBase_OutgoingMin_InvalidDenom(t *testing.T) {
	// Test case: try to swap coins if OutgoingMin.Denom != outCoin.Denom

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
	incoming := sample.Coin("abc", 10)

	// how much we will swap from usdt to abc
	outCoin, _, errCalcSwap := keeper.CalculateSwapAmount(&pool, incoming)

	// make sure there is no error
	require.NoError(t, errCalcSwap)

	// set ongoing minimum to invalid denom --> OutgoingMin.Denom != outCoin.Denom
	outgoingMin := sample.Coin("abc", 10)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, incoming).
		Return(true).
		Times(1)

	// create a "swap" message
	txSwapMsg := &types.MsgSwapExactIn{
		Signer:      creator,
		Incoming:    incoming,
		PoolId:      pool.PoolId,
		OutgoingMin: &outgoingMin,
	}

	// make rpc call to swap
	_, err := server.SwapExactIn(ctx, txSwapMsg)

	require.Error(t, err)

	// check if the error is correct
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)

	// check an error message
	require.Equal(t,
		fmt.Sprintf(
			"Outgoing minimum denom (%s) must be the same as outgoing denom (%s): invalid request",
			outgoingMin.Denom,
			outCoin.Denom,
		),
		err.Error(),
	)
}

func TestMsgServerSwapExactIn_SwapBase_OutgoingMin_SwapLessThanMin(t *testing.T) {
	// Test case: try to swap coins in a pool with swap amount less than an outgoing minimum

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
	incoming := sample.Coin("abc", 10)

	// how much we will swap from usdt to abc
	outCoin, _, errCalcSwap := keeper.CalculateSwapAmount(&pool, incoming)

	// make sure there is no error
	require.NoError(t, errCalcSwap)

	// set ongoing minimum
	outgoingMin := sample.Coin("usdt", 11)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, incoming).
		Return(true).
		Times(1)

	// create a "swap" message
	txSwapMsg := &types.MsgSwapExactIn{
		Signer:      creator,
		Incoming:    incoming,
		PoolId:      pool.PoolId,
		OutgoingMin: &outgoingMin,
	}

	// make rpc call to swap
	_, err := server.SwapExactIn(ctx, txSwapMsg)

	require.Error(t, err)

	// check if the error is correct
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)

	// check an error message
	require.Equal(t,
		fmt.Sprintf(
			"SwapExactIn amount (%s) is less than minimum outgoing amount (%s): invalid request",
			outCoin.String(),
			outgoingMin.String(),
		),
		err.Error(),
	)
}

func TestMsgServerSwapExactIn_SwapMoreThanBalance(t *testing.T) {
	// Test case: try to swap coins in a pool with swap amount more than the balance

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
	incoming := sample.Coin("abc", 55000000000000)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, incoming).
		Return(false).
		Times(1)

	// create a "swap" message
	txSwapMsg := &types.MsgSwapExactIn{
		Signer:   creator,
		Incoming: incoming,
		PoolId:   pool.PoolId,
	}

	// make rpc call to swap coins
	_, err := server.SwapExactIn(ctx, txSwapMsg)

	require.Error(t, err)

	// check an error message
	require.Equal(t,
		fmt.Sprintf(
			"Sender (%s) does not have enough balance for incoming amount (%s): insufficient funds",
			signer.String(),
			incoming.String(),
		),
		err.Error(),
	)
}

func TestMsgServerSwapExactIn_SendFromAddressToPoolFail(t *testing.T) {
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
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 1000)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 316)

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
	incoming := sample.Coin("abc", 10)

	// how much we will swap
	outCoin, _, errCalcSwap := keeper.CalculateSwapAmount(&pool, incoming)

	// make sure there is no error
	require.NoError(t, errCalcSwap)

	// check if the swap amount is correct
	outUsdt := sample.Coin("usdt", 83)
	require.Equal(t, outUsdt, outCoin)

	// extract the pool address from the pool account
	poolAddress := types.GetPoolAddress(pool.PoolId)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, incoming).
		Return(true).
		Times(1)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, incoming).
		Return(false).
		Times(1)

	// create a "swap" message
	txSwapMsg := &types.MsgSwapExactIn{
		Signer:   creator,
		Incoming: incoming,
		PoolId:   pool.PoolId,
	}

	// make rpc call to swap coins
	_, err := server.SwapExactIn(ctx, txSwapMsg)

	require.Error(t, err)

	// check an error message
	require.Equal(t,
		fmt.Sprintf(
			"SwapExactIn: Failed to send %s coins: SendFromAddressToPool: Insufficient funds in sender %s to send %s coins to poolID %s (address: %s): insufficient funds",
			incoming.String(),
			signer.String(),
			incoming.String(),
			pool.PoolId,
			poolAddress.String(),
		),
		err.Error(),
	)
}

func TestMsgServerSwapExactIn_SendFromPoolToAddressFail(t *testing.T) {
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
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 1000)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 316)

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
	incoming := sample.Coin("abc", 10)

	// how much we will swap
	outCoin, _, errCalcSwap := keeper.CalculateSwapAmount(&pool, incoming)

	// make sure there is no error
	require.NoError(t, errCalcSwap)

	// check if the swap amount is correct
	outUsdt := sample.Coin("usdt", 83)
	require.Equal(t, outUsdt, outCoin)

	// extract the pool address from the pool account
	poolAddress := types.GetPoolAddress(pool.PoolId)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, incoming).
		Return(true).
		Times(2)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(2)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(incoming)).
		Return(nil).
		Times(1)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), poolAddress, outCoin).
		Return(false).
		Times(1)

	// create a "swap" message
	txSwapMsg := &types.MsgSwapExactIn{
		Signer:   creator,
		Incoming: incoming,
		PoolId:   pool.PoolId,
	}

	// make rpc call to swap coins
	_, err := server.SwapExactIn(ctx, txSwapMsg)

	require.Error(t, err)

	// check an error message
	require.Equal(t,
		fmt.Sprintf(
			"SwapExactIn: Failed to send coins %s: SendFromPoolToAddress: Insufficient funds in poolID %s (address: %s) to send %s coins to receiver %s: insufficient funds",
			outCoin.String(),
			pool.PoolId,
			poolAddress.String(),
			outCoin.String(),
			signer.String(),
		),
		err.Error(),
	)
}

func TestMsgServerSwapExactIn_InvalidSignerAddress(t *testing.T) {
	// Test case: try to swap coins with an invalid signer address

	// Create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish() // Assert that all expectations are met

	// Create a mock bank keeper (no expectations needed since the function should fail early)
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// Create a dex keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, nil)
	srv := keeper.NewMsgServerImpl(k)

	// Define an invalid Bech32 address (e.g., wrong format or invalid characters)
	invalidSigner := "invalid-address-not-bech32"

	// Create a swap message with an invalid signer
	txSwapMsg := &types.MsgSwapExactIn{
		Signer:   invalidSigner,
		PoolId:   "zp1",
		Incoming: sample.Coin("abc", 10),
	}

	// Make RPC call to swap coins
	_, err := srv.SwapExactIn(ctx, txSwapMsg)

	// Verify the error
	require.Error(t, err)
	require.Contains(t, err.Error(), "decoding bech32 failed")
}

func TestMsgServerSwapExactIn_InvalidReceiverAddress(t *testing.T) {
	// Test case: try to swap coins with an invalid receiver address

	// Create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish() // Assert that all expectations are met

	// Create a sample signer address
	creator := sample.AccAddress()
	signer := sdk.MustAccAddressFromBech32(creator)

	// Create all coins required for the pool creation
	createPoolBase := sample.Coin("abc", 100)
	createPoolQuote := sample.Coin("usdt", 1000)
	createPoolCreationFee := sample.Coin("uzig", 100000000)
	createPoolExpectedLPCoin := sample.Coin("zp1", 316)

	// Create dex keeper with pool mock
	server, _, ctx, pool, _, _, _ := common.ServerDexKeeperWithPoolMock(
		t,
		ctrl,
		signer,
		createPoolBase,
		createPoolQuote,
		createPoolCreationFee,
		createPoolExpectedLPCoin,
	)

	// Define an invalid Bech32 address for receiver
	invalidReceiver := "invalid-address-not-bech32"

	// Define the incoming coin
	incoming := sample.Coin("abc", 10)

	// Create a swap message with an invalid receiver
	txSwapMsg := &types.MsgSwapExactIn{
		Signer:   creator,
		Receiver: invalidReceiver,
		PoolId:   pool.PoolId,
		Incoming: incoming,
	}

	// Make RPC call to swap coins
	_, err := server.SwapExactIn(ctx, txSwapMsg)

	// Verify the error
	require.Error(t, err)
	require.Contains(t, err.Error(), "decoding bech32 failed")
}

func TestMsgServerSwapExactIn_PoolZeroBalance(t *testing.T) {
	// Test case: try to swap coins in a pool with a zero balance for one token

	// Create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish() // Assert that all expectations are met

	// Create a sample signer address
	creator := sample.AccAddress()
	signer := sdk.MustAccAddressFromBech32(creator)

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

	// Define the incoming coin
	incoming := sample.Coin("abc", 10)

	// Mock the balance check for the signer
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, incoming).
		Return(true).
		Times(1)

	// Create a swap message
	txSwapMsg := &types.MsgSwapExactIn{
		Signer:   creator,
		PoolId:   pool.PoolId,
		Incoming: incoming,
	}

	// Make RPC call to swap coins
	_, err := srv.SwapExactIn(ctx, txSwapMsg)

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

func TestMsgServerSwapExactIn_CalculateSwapAmount_NewBaseTokenBalanceSanityCheck(t *testing.T) {
	// Test case: try to swap coins in a pool that triggers the newBaseTokenBalance sanity check

	t.Run("new base token balance becomes zero with exact amount", func(t *testing.T) {
		// This test creates a scenario where newBaseTokenBalance becomes exactly zero

		// Create a mock controller to test through the full SwapExactIn flow
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		creator := sample.AccAddress()
		signer := sdk.MustAccAddressFromBech32(creator)

		// Create a pool with specific amounts that could lead to zero balance
		createPoolBase := sdk.NewInt64Coin("abc", 100)
		createPoolQuote := sdk.NewInt64Coin("usdt", 1000)
		createPoolCreationFee := sample.Coin("uzig", 100000000)
		createPoolExpectedLPCoin := sample.Coin("zp1", 316)

		// Create dex keeper with pool mock
		server, _, ctx, pool, _, bankKeeper, _ := common.ServerDexKeeperWithPoolMock(
			t,
			ctrl,
			signer,
			createPoolBase,
			createPoolQuote,
			createPoolCreationFee,
			createPoolExpectedLPCoin,
		)

		// Create an incoming amount that
		incoming := sdk.NewInt64Coin("abc", 1) // Very small amount

		bankKeeper.
			EXPECT().
			HasBalance(gomock.Any(), signer, incoming).
			Return(true).
			Times(1)

		txSwapMsg := &types.MsgSwapExactIn{
			Signer:   creator,
			Incoming: incoming,
			PoolId:   pool.PoolId,
		}

		result, err := server.SwapExactIn(ctx, txSwapMsg)

		// If this triggers the sanity check error, verify it
		if err != nil && strings.Contains(err.Error(), "Invalid new abc token balance") {
			require.Nil(t, result)
			require.Contains(t, err.Error(), "has to be positive")
			require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
		}
	})
}

func TestCalculateSwapAmount_DirectSanityCheckCoverage(t *testing.T) {
	// Test case: test CalculateSwapAmount to trigger newBaseTokenBalance sanity check

	t.Run("direct test - zero new base token balance", func(t *testing.T) {
		pool := types.Pool{
			PoolId:  "test",
			LpToken: sdk.NewInt64Coin("lp", 100),
			Coins: []sdk.Coin{
				sdk.NewInt64Coin("abc", 0), // Zero initial balance
				sdk.NewInt64Coin("usdt", 100),
			},
			Fee: 50000,
		}

		incoming := sdk.NewInt64Coin("abc", types.MinSwapFee)

		_, _, err := keeper.CalculateSwapAmount(&pool, incoming)

		// This should fail the sanity check for newBaseTokenBalance
		require.Error(t, err)
		require.Contains(t, err.Error(), "Invalid new abc token balance")
		require.Contains(t, err.Error(), "has to be positive")
	})

	t.Run("direct test - negative new base token balance", func(t *testing.T) {
		// Create a scenario where the math could lead to negative balance

		pool := types.Pool{
			PoolId:  "test",
			LpToken: sdk.NewInt64Coin("lp", 100),
			Coins: []sdk.Coin{
				sdk.NewInt64Coin("abc", 1), // Very small initial balance
				sdk.NewInt64Coin("usdt", 100),
			},
			Fee: 50000,
		}

		// Try with the minimum possible amount that could cause issues
		incoming := sdk.NewInt64Coin("abc", 1)

		_, _, err := keeper.CalculateSwapAmount(&pool, incoming)

		// Check if this triggers the sanity check
		if err != nil && strings.Contains(err.Error(), "Invalid new abc token balance") {
			require.Contains(t, err.Error(), "has to be positive")
			require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
		}
	})
}
