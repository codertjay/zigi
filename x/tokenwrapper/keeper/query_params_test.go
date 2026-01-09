package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "zigchain/testutil/keeper"
	"zigchain/x/tokenwrapper/types"
)

// Positive test cases

func TestParamsQuery_Valid(t *testing.T) {
	// Test case: query params with valid request

	k, ctx := keepertest.TokenwrapperKeeper(t, nil)
	params := types.DefaultParams()
	require.NoError(t, k.SetParams(ctx, params))

	response, err := k.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}

// Negative test cases

func TestParamsQuery_InvalidRequest(t *testing.T) {
	// Test case: query params with invalid request

	k, ctx := keepertest.TokenwrapperKeeper(t, nil)

	response, err := k.Params(ctx, nil)
	require.Error(t, err)
	require.Nil(t, response)

	statusErr, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, statusErr.Code())
	require.Equal(t, "invalid request", statusErr.Message())
}
