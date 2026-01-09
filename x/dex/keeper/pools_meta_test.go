package keeper_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "zigchain/testutil/keeper"
	"zigchain/testutil/nullify"
	"zigchain/x/dex/keeper"
	"zigchain/x/dex/types"
)

func createTestPoolsMeta(keeper keeper.Keeper, ctx context.Context) types.PoolsMeta {
	item := types.PoolsMeta{}
	keeper.SetPoolsMeta(ctx, item)
	return item
}

// Positive test cases

func TestPoolsMetaGet(t *testing.T) {
	// Test case: get pool meta

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	item := createTestPoolsMeta(k, ctx)
	rst, found := k.GetPoolsMeta(ctx)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&item),
		nullify.Fill(&rst),
	)
}

func TestPoolsMetaRemove(t *testing.T) {
	// Test case: remove pool meta

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	createTestPoolsMeta(k, ctx)
	k.RemovePoolsMeta(ctx)
	_, found := k.GetPoolsMeta(ctx)
	require.False(t, found)
}

// Negative test cases

func TestPoolsMetaGet_NotFound(t *testing.T) {
	// Test case: get pool meta not found

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)

	// No SetPoolsMeta called
	_, found := k.GetPoolsMeta(ctx)
	require.False(t, found)
}
