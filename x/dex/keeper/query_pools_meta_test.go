package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "zigchain/testutil/keeper"
	"zigchain/x/dex/keeper"
	"zigchain/x/dex/types"
)

// Positive test cases

func TestQueryPoolsMeta_GetPoolsMeta_Positive(t *testing.T) {
	// Test case: get pool meta

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	qs := keeper.NewQueryServerImpl(k)

	// set a sample PoolsMeta
	expected := types.PoolsMeta{NextPoolId: 42}
	k.SetPoolsMeta(ctx, expected)

	// query the PoolsMeta
	resp, err := qs.GetPoolsMeta(ctx, &types.QueryGetPoolsMetaRequest{})
	require.NoError(t, err)
	require.Equal(t, expected, resp.PoolsMeta)
}

// Negative test cases

func TestQueryPoolsMeta_GetPoolsMeta_InvalidRequest(t *testing.T) {
	// Test case: get pool meta with invalid request

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	qs := keeper.NewQueryServerImpl(k)

	// query the PoolsMeta with an invalid request
	_, err := qs.GetPoolsMeta(ctx, nil)
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument.String(), status.Code(err).String())
	require.Contains(t, err.Error(), "invalid request")
}

func TestQueryPoolsMeta_GetPoolsMeta_PoolNotFound(t *testing.T) {
	// Test case: get pool meta

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	qs := keeper.NewQueryServerImpl(k)

	// query the PoolsMeta with an invalid request
	_, err := qs.GetPoolsMeta(ctx, &types.QueryGetPoolsMetaRequest{})
	require.Error(t, err)
	require.Equal(t, codes.NotFound.String(), status.Code(err).String())
}
