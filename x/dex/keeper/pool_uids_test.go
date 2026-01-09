package keeper_test

import (
	"context"
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	keepertest "zigchain/testutil/keeper"
	"zigchain/testutil/nullify"
	"zigchain/x/dex/keeper"
	"zigchain/x/dex/types"

	"github.com/stretchr/testify/require"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNPoolUids(keeper keeper.Keeper, ctx context.Context, n int) []types.PoolUids {
	items := make([]types.PoolUids, n)
	for i := range items {
		items[i].PoolUid = strconv.Itoa(i)

		keeper.SetPoolUids(ctx, items[i])
	}
	return items
}

// Positive test cases

func TestPoolUidsGet(t *testing.T) {
	// Test case: get pool Uids

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	items := createNPoolUids(k, ctx, 10)
	for _, item := range items {
		rst, found := k.GetPoolUids(ctx,
			item.PoolUid,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}

func TestPoolUidsRemove(t *testing.T) {
	// Test case: remove pool Uids

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	items := createNPoolUids(k, ctx, 10)
	for _, item := range items {
		k.RemovePoolUids(ctx,
			item.PoolUid,
		)
		_, found := k.GetPoolUids(ctx,
			item.PoolUid,
		)
		require.False(t, found)
	}
}

func TestPoolUidsGetAll(t *testing.T) {
	// Test case: get all pool Uids

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	items := createNPoolUids(k, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(k.GetAllPoolUids(ctx)),
	)
}

func TestGetPoolUidString(t *testing.T) {
	// Test case: get pool uid string

	pool := types.Pool{
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 100),
			sdk.NewInt64Coin("usdt", 200),
		},
	}

	expected := "abc-usdt"
	require.Equal(t, expected, types.GetPoolUidString(pool))
}

func TestSetPoolUidFromPool(t *testing.T) {
	// Test case: set pool uid

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)

	pool := types.Pool{
		PoolId: "zp1",
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 100),
			sdk.NewInt64Coin("usdt", 200),
		},
	}

	k.SetPoolUidFromPool(ctx, pool)

	uid := types.GetPoolUidString(pool)
	stored, found := k.GetPoolUids(ctx, uid)
	require.True(t, found)
	require.Equal(t, pool.PoolId, stored.PoolId)
	require.Equal(t, uid, stored.PoolUid)
}

func TestGetPoolUidsFromCoins(t *testing.T) {
	// Test case: get pool uids from coins

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)

	pool := types.Pool{
		PoolId: "zp2",
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("abc", 100),
			sdk.NewInt64Coin("usdt", 200),
		},
	}

	k.SetPoolUidFromPool(ctx, pool)

	// correct input
	coins := sdk.NewCoins(
		sdk.NewInt64Coin("abc", 1),
		sdk.NewInt64Coin("usdt", 1),
	)

	foundUid, found := k.GetPoolUidsFromCoins(ctx, coins)
	require.True(t, found)
	require.Equal(t, "abc-usdt", foundUid.PoolUid)
	require.Equal(t, pool.PoolId, foundUid.PoolId)

	// invalid input (1 coin)
	invalidCoins := sdk.NewCoins(sdk.NewInt64Coin("abc", 1))
	_, found = k.GetPoolUidsFromCoins(ctx, invalidCoins)
	require.False(t, found)
}

func TestGetPoolUidsFromPool(t *testing.T) {
	// Test case: get pool uids from pool

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)

	pool := types.Pool{
		PoolId: "zp3",
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("eth", 100),
			sdk.NewInt64Coin("usdt", 100),
		},
	}

	k.SetPoolUidFromPool(ctx, pool)

	uid, found := k.GetPoolUidsFromPool(ctx, pool)
	require.True(t, found)
	require.Equal(t, pool.PoolId, uid.PoolId)
	require.Equal(t, "eth-usdt", uid.PoolUid)
}

// Negative test cases

func TestGetPoolUidsFromCoins_InvalidCoinLength(t *testing.T) {
	// Test case: get pool uids from coins with invalid length

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)

	coins := sdk.NewCoins(sdk.NewInt64Coin("abc", 100)) // only 1 coin
	_, found := k.GetPoolUidsFromCoins(ctx, coins)
	require.False(t, found)
}

func TestGetPoolUids_NotFound(t *testing.T) {
	// Test case: get pool uids that does not exist

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)

	_, found := k.GetPoolUids(ctx, "non-existent-uid")
	require.False(t, found)
}

func TestGetPoolUidsFromPool_NotFound(t *testing.T) {
	// Test case: get pool uids from pool that does not exist

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)

	pool := types.Pool{
		PoolId: "zp404",
		Coins: []sdk.Coin{
			sdk.NewInt64Coin("ghost", 100),
			sdk.NewInt64Coin("phantom", 200),
		},
	}

	_, found := k.GetPoolUidsFromPool(ctx, pool)
	require.False(t, found)
}

func TestGetPoolUidString_PanicsOnEmptyCoins(t *testing.T) {
	// Test case: get pool uid string panics on empty coins

	pool := types.Pool{
		Coins: []sdk.Coin{}, // invalid
	}

	require.Panics(t, func() {
		types.GetPoolUidString(pool)
	})
}
