package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "zigchain/testutil/keeper"
	"zigchain/testutil/nullify"
	"zigchain/testutil/sample"
	"zigchain/x/dex/keeper"
	"zigchain/x/dex/testutil"
	"zigchain/x/dex/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

// Positive test cases

func TestQueryPool_GetPool(t *testing.T) {
	// Test case: get pool data

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
	createPoolQuote := sample.Coin("usdt", 50)
	creationFee := sample.Coin("uzig", 100000000)
	// square root of a * b
	createPoolExpectedLPCoin := sample.Coin("zp1", 70)

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

	// get the pool from the query server
	qs := keeper.NewQueryServerImpl(k)
	response, err := qs.GetPool(ctx,
		&types.QueryGetPoolRequest{
			PoolId: poolId,
		},
	)

	require.NoError(t, err)

	// check the response
	require.NotNil(t, response)
	require.Equal(t, pool.Creator, response.Pool.Creator)
	require.Equal(t, poolId, response.Pool.PoolId)
	require.Equal(t, createPoolExpectedLPCoin, response.Pool.LpToken)
	require.Equal(t, createPoolBase, pool.Coins[0])
	require.Equal(t, createPoolQuote, pool.Coins[1])
	require.Equal(t, uint32(500), pool.Fee)
	require.Equal(t, "constant_product", pool.Formula)
}

func TestQueryPool_GetPool2(t *testing.T) {
	// Test case: get pool data

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	qs := keeper.NewQueryServerImpl(k)

	// create 2 pools
	pools := createNPool(k, ctx, 2)

	// query the first pool
	expected1 := pools[0]
	resp, err := qs.GetPool(ctx, &types.QueryGetPoolRequest{PoolId: expected1.PoolId})
	require.NoError(t, err)
	require.Equal(t, nullify.Fill(&expected1), nullify.Fill(&resp.Pool))

	// query the second pool
	expected2 := pools[1]
	resp, err = qs.GetPool(ctx, &types.QueryGetPoolRequest{PoolId: expected2.PoolId})
	require.NoError(t, err)
	require.Equal(t, nullify.Fill(&expected2), nullify.Fill(&resp.Pool))
}

func TestQueryPool_ListPool(t *testing.T) {
	// Test case: list pools data

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	srv := keeper.NewQueryServerImpl(k)

	// create multiple pools
	for i := 0; i < 5; i++ {
		pool := types.Pool{
			PoolId:  "zp" + strconv.Itoa(i+1),
			Creator: sample.AccAddress(),
			Coins: []sdk.Coin{
				sample.Coin("abc", 100),
				sample.Coin("usdt", 200),
			},
			LpToken: sample.Coin("zp", 50),
			Fee:     500,
			Formula: "constant_product",
		}
		k.SetPool(ctx, pool)
	}

	req := &types.QueryAllPoolRequest{
		Pagination: &query.PageRequest{Limit: 3},
	}

	res, err := srv.ListPool(ctx, req)
	require.NoError(t, err)
	require.Len(t, res.Pool, 3)
	require.NotNil(t, res.Pagination)
}

func TestQueryPool_ListPools_ReturnsAllPools(t *testing.T) {
	// Test case: list all pools

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	qs := keeper.NewQueryServerImpl(k)

	// create 3 pools with predictable IDs
	pools := []types.Pool{
		{
			PoolId:  "zp1",
			Creator: sample.AccAddress(),
			Coins: []sdk.Coin{
				sample.Coin("abc", 100),
				sample.Coin("usdt", 200),
			},
			LpToken: sample.Coin("zp1", 100),
			Fee:     500,
			Formula: "constant_product",
		},
		{
			PoolId:  "zp2",
			Creator: sample.AccAddress(),
			Coins: []sdk.Coin{
				sample.Coin("eth", 150),
				sample.Coin("usdc", 300),
			},
			LpToken: sample.Coin("zp2", 120),
			Fee:     500,
			Formula: "constant_product",
		},
		{
			PoolId:  "zp3",
			Creator: sample.AccAddress(),
			Coins: []sdk.Coin{
				sample.Coin("btc", 200),
				sample.Coin("usdt", 400),
			},
			LpToken: sample.Coin("zp3", 140),
			Fee:     500,
			Formula: "constant_product",
		},
	}

	// store all pools
	for _, pool := range pools {
		k.SetPool(ctx, pool)
	}

	// perform the query
	resp, err := qs.ListPool(ctx, &types.QueryAllPoolRequest{
		Pagination: &query.PageRequest{Limit: 100}, // ensure all are included
	})
	require.NoError(t, err)
	require.Len(t, resp.Pool, len(pools))

	// compare without zero-value noise
	require.ElementsMatch(t,
		nullify.Fill(pools),
		nullify.Fill(resp.Pool),
	)
}

func TestQueryPool_Paginated(t *testing.T) {
	// Test case: list pools data with pagination

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	qs := keeper.NewQueryServerImpl(k)
	msgs := createNPool(k, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllPoolRequest {
		return &types.QueryAllPoolRequest{
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}
	t.Run("ByOffset", func(t *testing.T) {
		step := 2
		for i := 0; i < len(msgs); i += step {
			resp, err := qs.ListPool(ctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Pool), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.Pool),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := qs.ListPool(ctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Pool), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.Pool),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := qs.ListPool(ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.Pool),
		)
	})
}

func TestQueryPool_GetPoolBalances(t *testing.T) {
	// Test case: get pool balances

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
	createPoolQuote := sample.Coin("usdt", 50)
	creationFee := sample.Coin("uzig", 100000000)
	// square root of a * b
	createPoolExpectedLPCoin := sample.Coin("zp1", 70)

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

	// get the pool balances from the query server
	qs := keeper.NewQueryServerImpl(k)

	// expect balances to be returned from bank keeper for pool address
	expectedBalances := sdk.NewCoins(createPoolBase, createPoolQuote)

	bankKeeper.EXPECT().
		GetAllBalances(ctx, poolAddress).
		Return(expectedBalances).
		Times(1)

	// get the pool balances from the query server
	response, err := qs.GetPoolBalances(ctx,
		&types.QueryGetPoolBalancesRequest{
			PoolId: poolId,
		},
	)

	// check the response
	require.NoError(t, err)
	require.Equal(t, poolId, response.Pool.PoolId)
	require.Equal(t, creator, response.Pool.Creator)
	require.Equal(t, uint32(500), response.Pool.Fee)
	require.Equal(t, "constant_product", response.Pool.Formula)
	require.Equal(t, poolAddress.String(), response.Pool.Address)
	require.Equal(t, createPoolExpectedLPCoin, response.Pool.LpToken)
	require.ElementsMatch(t, sdk.NewCoins(createPoolBase, createPoolQuote), response.Pool.Coins)
	require.ElementsMatch(t, sdk.NewCoins(createPoolBase, createPoolQuote), response.Balances)
}

// Negative test cases

func TestQueryPool_GetPool_InvalidRequest_Nil(t *testing.T) {
	// Test case: try to get pool data with nil request

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	srv := keeper.NewQueryServerImpl(k)

	_, err := srv.GetPool(ctx, nil)
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	require.Contains(t, err.Error(), "invalid request")
}

func TestQueryPool_GetPool_NotFound(t *testing.T) {
	// Test case: try to get pool data with non-existing pool id

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	srv := keeper.NewQueryServerImpl(k)

	_, err := srv.GetPool(ctx, &types.QueryGetPoolRequest{PoolId: "does-not-exist"})
	require.Error(t, err)
	require.Equal(t, codes.NotFound, status.Code(err))
	require.Contains(t, err.Error(), "not found")
}

func TestQueryPool_GetPool_KeyNotFound(t *testing.T) {
	// Test case: try to get pool data with non-existing pool id

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	qs := keeper.NewQueryServerImpl(k)

	_, err := qs.GetPool(ctx, &types.QueryGetPoolRequest{PoolId: strconv.Itoa(100000)})
	require.ErrorIs(t, err, status.Error(codes.NotFound, "not found"))
}

func TestQueryPool_ListPool_InvalidRequest_Nil(t *testing.T) {
	// Test case: try to list pools data with nil request

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	srv := keeper.NewQueryServerImpl(k)

	_, err := srv.ListPool(ctx, nil)
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	require.Contains(t, err.Error(), "invalid request")
}

func TestQueryPool_GetPoolBalances_InvalidRequest_Nil(t *testing.T) {
	// Test case: try to get pool balances with nil request

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	srv := keeper.NewQueryServerImpl(k)

	_, err := srv.GetPoolBalances(ctx, nil)
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	require.Contains(t, err.Error(), "invalid request")
}

func TestQueryPool_GetPoolBalances_NotFound(t *testing.T) {
	// Test case: try to get pool balances with non-existing pool id

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	srv := keeper.NewQueryServerImpl(k)

	_, err := srv.GetPoolBalances(ctx, &types.QueryGetPoolBalancesRequest{PoolId: "does-not-exist"})
	require.Error(t, err)
	require.Equal(t, codes.NotFound, status.Code(err))
	require.Contains(t, err.Error(), "not found")
}

func TestQueryPool_ListPool_PaginateError(t *testing.T) {
	// Test case: try to list pools data with invalid pagination request

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	qs := keeper.NewQueryServerImpl(k)

	req := &types.QueryAllPoolRequest{
		Pagination: &query.PageRequest{
			Key:    []byte("invalid"),
			Offset: 1,
		},
	}

	resp, err := qs.ListPool(ctx, req)
	require.Nil(t, resp)
	require.Error(t, err)
	require.Equal(t, codes.Internal, status.Code(err))
	require.Contains(t, err.Error(), "either offset or key is expected, got both")
}
