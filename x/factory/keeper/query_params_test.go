package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "zigchain/testutil/keeper"
	"zigchain/x/factory/types"
)

// Positive test cases

func TestQueryParams_Positive(t *testing.T) {
	// Test case: query params successfully

	keeper, ctx := keepertest.FactoryKeeper(t, nil, nil)
	params := types.DefaultParams()
	require.NoError(t, keeper.SetParams(ctx, params))

	response, err := keeper.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}

// Negative test cases

func TestQueryParams_InvalidRequest(t *testing.T) {
	// Test case: query params with invalid request

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	_, err := k.Params(ctx, nil)
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument.String(), status.Code(err).String())
	require.Contains(t, err.Error(), "invalid request")
}
