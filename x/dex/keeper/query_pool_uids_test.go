package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "zigchain/testutil/keeper"
	"zigchain/testutil/nullify"
	"zigchain/testutil/sample"
	"zigchain/x/dex/keeper"
	"zigchain/x/dex/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

// Positive test cases

func TestQueryPoolUids_GetPoolUid(t *testing.T) {
	// Test case: querying a pool UID by base and quote pair

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	qs := keeper.NewQueryServerImpl(k)

	// create and store a pool
	pool := types.Pool{
		PoolId: "zp1",
		Coins: []sdk.Coin{
			sample.Coin("abc", 100),
			sample.Coin("usdt", 200),
		},
	}
	k.SetPool(ctx, pool)
	k.SetPoolUidFromPool(ctx, pool)

	// query using a base/quote pair
	resp, err := qs.GetPoolUid(ctx, &types.QueryGetPoolUidRequest{
		Base:  "abc",
		Quote: "usdt",
	})
	require.NoError(t, err)
	require.Equal(t, "abc-usdt", resp.PoolUids.PoolUid)
	require.Equal(t, pool.PoolId, resp.PoolUids.PoolId)
}

func TestQueryPoolUids_ListPoolUids(t *testing.T) {
	// Test case: querying all pool UIDs

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	qs := keeper.NewQueryServerImpl(k)

	// set several pool uids
	pools := []types.Pool{
		{
			PoolId: "zp1",
			Coins:  []sdk.Coin{sample.Coin("abc", 100), sample.Coin("usdt", 200)},
		},
		{
			PoolId: "zp2",
			Coins:  []sdk.Coin{sample.Coin("btc", 300), sample.Coin("usdt", 400)},
		},
	}

	for _, p := range pools {
		k.SetPoolUidFromPool(ctx, p)
	}

	resp, err := qs.ListPoolUids(ctx, &types.QueryAllPoolUidsRequest{
		Pagination: &query.PageRequest{Limit: 10},
	})
	require.NoError(t, err)
	require.Len(t, resp.PoolUids, len(pools))
}

func TestQueryPoolUids_Paginated(t *testing.T) {
	// Test case: querying paginated results from the PoolUids table

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	qs := keeper.NewQueryServerImpl(k)
	msgs := createNPoolUids(k, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllPoolUidsRequest {
		return &types.QueryAllPoolUidsRequest{
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
			resp, err := qs.ListPoolUids(ctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.PoolUids), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.PoolUids),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := qs.ListPoolUids(ctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.PoolUids), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.PoolUids),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := qs.ListPoolUids(ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.PoolUids),
		)
	})
}

// Negative test cases

func TestQueryPoolUids_GetPoolUid_InvalidRequest(t *testing.T) {
	// Test case: querying with invalid request parameters

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	qs := keeper.NewQueryServerImpl(k)

	resp, err := qs.GetPoolUid(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	require.Contains(t, err.Error(), "invalid request")
}

func TestQueryPoolUids_GetPoolUid_NotFound(t *testing.T) {
	// Test case: querying a non-existent pool UID

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	qs := keeper.NewQueryServerImpl(k)

	resp, err := qs.GetPoolUid(ctx, &types.QueryGetPoolUidRequest{
		Base:  "btc",
		Quote: "usdt",
	})
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "Pool btc-usdt with base: btc and quote: usdt not found")
}

func TestQueryPoolUids_List_InvalidRequest(t *testing.T) {
	// Test case: querying with invalid request parameters

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	qs := keeper.NewQueryServerImpl(k)

	resp, err := qs.ListPoolUids(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	require.Contains(t, err.Error(), "invalid request")
}

func TestQueryPool_ListPoolUids_PaginateError(t *testing.T) {
	// Test case: querying with both key and offset in pagination

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	qs := keeper.NewQueryServerImpl(k)

	req := &types.QueryAllPoolUidsRequest{
		Pagination: &query.PageRequest{
			Key:    []byte("invalid"),
			Offset: 1,
		},
	}

	resp, err := qs.ListPoolUids(ctx, req)
	require.Nil(t, resp)
	require.Error(t, err)
	require.Equal(t, codes.Internal, status.Code(err))
	require.Contains(t, err.Error(), "either offset or key is expected, got both")
}
