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

func TestParamsQuery(t *testing.T) {
	// Test case: query params successfully

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	qs := keeper.NewQueryServerImpl(k)
	params := types.DefaultParams()
	require.NoError(t, k.SetParams(ctx, params))

	response, err := qs.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}

func TestQueryParams_ValidRequest(t *testing.T) {
	// Test case: query params with valid params

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	qs := keeper.NewQueryServerImpl(k)

	params := types.Params{
		NewPoolFeePct: 600,
		CreationFee:   200_000_000,
	}
	require.NoError(t, k.SetParams(ctx, params))

	response, err := qs.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}

// Negative test cases

func TestQueryParams_InvalidRequest_Nil(t *testing.T) {
	// Test case: query params with nil request

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	qs := keeper.NewQueryServerImpl(k)

	_, err := qs.Params(ctx, nil)
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument.String(), status.Code(err).String())
	require.Contains(t, err.Error(), "invalid request")

}
