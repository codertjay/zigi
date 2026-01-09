package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "zigchain/testutil/keeper"
	"zigchain/x/factory/types"
)

// Positive test cases

func TestGetParams(t *testing.T) {
	// Test case: set and get all parameters

	k, ctx := keepertest.FactoryKeeper(t, nil, nil)
	// get defaults params
	params := types.DefaultParams()

	// make sure there is no error while setting the params
	require.NoError(t, k.SetParams(ctx, params))
	// make sure the params are the same that store returns
	require.EqualValues(t, params, k.GetParams(ctx))
}

func TestGetParams_Defaults(t *testing.T) {
	// Test case: get default parameters when none are set

	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	// Get params and compare to expected struct
	expected := types.Params{
		CreateFeeDenom:  "uzig",
		CreateFeeAmount: 1000,
		Beneficiary:     "",
	}

	require.Equal(t, expected, k.GetParams(ctx))
}
