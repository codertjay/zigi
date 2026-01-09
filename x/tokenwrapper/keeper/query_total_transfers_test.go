package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "zigchain/testutil/keeper"
	"zigchain/x/tokenwrapper/types"
)

// Positive test cases

func TestTotalTransfersQuery_Valid(t *testing.T) {
	// Test case: query total transfers with valid parameters

	keeper, ctx := keepertest.TokenwrapperKeeper(t, nil)

	keeper.SetTotalTransferredIn(ctx, sdkmath.NewInt(100))
	keeper.SetTotalTransferredOut(ctx, sdkmath.NewInt(50))

	response, err := keeper.TotalTransfers(ctx, &types.QueryTotalTransfersRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryTotalTransfersResponse{
		TotalTransferredIn:  sdkmath.NewInt(100),
		TotalTransferredOut: sdkmath.NewInt(50),
	}, response)
}

// Negative test cases

func TestTotalTransfersQuery_InvalidRequest(t *testing.T) {
	// Test case: request is nil

	keeper, ctx := keepertest.TokenwrapperKeeper(t, nil)

	resp, err := keeper.TotalTransfers(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
	require.Equal(t, "invalid request", st.Message())
}
