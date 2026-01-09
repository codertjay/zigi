package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "zigchain/testutil/keeper"
	"zigchain/x/tokenwrapper/types"
)

// Positive test cases

func TestGetParams(t *testing.T) {
	// Test case: Get default parameters

	k, ctx := keepertest.TokenwrapperKeeper(t, nil)
	params := types.DefaultParams()

	require.NoError(t, k.SetParams(ctx, params))
	require.EqualValues(t, params, k.GetParams(ctx))
}
