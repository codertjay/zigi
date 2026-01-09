package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "zigchain/testutil/keeper"
	"zigchain/x/dex/types"
)

// Positive test cases

func TestGetSetParams(t *testing.T) {
	// Test case: get and set parameters in the keeper
	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	params := types.DefaultParams()

	require.NoError(t, k.SetParams(ctx, params))
	require.EqualValues(t, params, k.GetParams(ctx))
}

func TestGetParams_EmptyStore(t *testing.T) {
	// Test case: get parameters from an empty store
	// If the keeper always initializes with defaults, test that behavior instead
	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)

	params := k.GetParams(ctx)

	// Verify we get the expected default params that the keeper initializes with
	expectedDefaultParams := types.Params{
		NewPoolFeePct:        500,
		CreationFee:          100000000,
		Beneficiary:          "",
		MinimalLiquidityLock: 1000,
		MaxSlippage:          0,
	}
	require.EqualValues(t, expectedDefaultParams, params)
}

func TestSetParams_NilParams(t *testing.T) {
	// Test case: set nil parameters (should work with zero values)
	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)

	// Create zero-value params (effectively "nil" in terms of content)
	params := types.Params{}

	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	// Verify we can retrieve the zero-value params
	retrieved := k.GetParams(ctx)
	require.EqualValues(t, params, retrieved)
}

func TestGetParams_AfterMultipleSets(t *testing.T) {
	// Test case: verify parameters can be updated multiple times
	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)

	// Set initial params
	params1 := types.DefaultParams()
	require.NoError(t, k.SetParams(ctx, params1))
	require.EqualValues(t, params1, k.GetParams(ctx))

	// Update params with new values
	params2 := types.Params{
		MinimalLiquidityLock: 1000,
		MaxSlippage:          500,
	}
	require.NoError(t, k.SetParams(ctx, params2))
	require.EqualValues(t, params2, k.GetParams(ctx))

	// Update again
	params3 := types.Params{
		MinimalLiquidityLock: 2000,
		MaxSlippage:          1000,
	}
	require.NoError(t, k.SetParams(ctx, params3))
	require.EqualValues(t, params3, k.GetParams(ctx))
}

func TestGetParams_AfterParamsDeleted(t *testing.T) {
	// Test case: test GetParams after params have been set and then "deleted"
	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)

	// First set some params
	initialParams := types.DefaultParams()
	require.NoError(t, k.SetParams(ctx, initialParams))

	// Verify they are set
	require.Equal(t, initialParams, k.GetParams(ctx))

	// Now "delete" by setting zero-value params (simulates deletion for coverage)
	zeroParams := types.Params{}
	require.NoError(t, k.SetParams(ctx, zeroParams))

	// Get should return the zero values
	retrieved := k.GetParams(ctx)
	require.Equal(t, zeroParams, retrieved)
}

func TestV2Migration(t *testing.T) {
	// Test case: V2 migration updates parameters correctly
	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)

	// Set initial params with non-default values
	initialParams := types.Params{
		NewPoolFeePct:        500,
		CreationFee:          100000000,
		Beneficiary:          "",
		MinimalLiquidityLock: 1000, // Non-zero value
		MaxSlippage:          2000, // Non-default value
	}
	require.NoError(t, k.SetParams(ctx, initialParams))

	// Execute migration
	err := k.V2Migration(ctx)
	require.NoError(t, err)

	// Verify params were updated correctly
	updatedParams := k.GetParams(ctx)
	require.Equal(t, uint32(0), updatedParams.MinimalLiquidityLock)
	require.Equal(t, types.DefaultMaxSlippage, updatedParams.MaxSlippage)

	// Verify other fields remain unchanged
	require.Equal(t, initialParams.NewPoolFeePct, updatedParams.NewPoolFeePct)
	require.Equal(t, initialParams.CreationFee, updatedParams.CreationFee)
	require.Equal(t, initialParams.Beneficiary, updatedParams.Beneficiary)
}

func TestV2Migration_FromEmptyStore(t *testing.T) {
	// Test case: V2 migration works when starting from empty store
	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)

	// Clear any existing params to simulate empty store
	require.NoError(t, k.SetParams(ctx, types.Params{}))

	// Store is empty initially, GetParams returns zero values
	initialParams := k.GetParams(ctx)
	require.Equal(t, uint32(0), initialParams.MinimalLiquidityLock)
	require.Equal(t, uint32(0), initialParams.MaxSlippage)

	// Execute migration
	err := k.V2Migration(ctx)
	require.NoError(t, err)

	// Verify params were set to expected values
	updatedParams := k.GetParams(ctx)
	require.Equal(t, uint32(0), updatedParams.MinimalLiquidityLock)
	require.Equal(t, types.DefaultMaxSlippage, updatedParams.MaxSlippage)
}

func TestV2Migration_AlreadyAtTargetValues(t *testing.T) {
	// Test case: V2 migration when params are already at target values
	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)

	// Set params to exactly what migration would set
	initialParams := types.Params{
		NewPoolFeePct:        500,
		CreationFee:          100000000,
		Beneficiary:          "",
		MinimalLiquidityLock: 0,
		MaxSlippage:          types.DefaultMaxSlippage,
	}
	require.NoError(t, k.SetParams(ctx, initialParams))

	// Execute migration
	err := k.V2Migration(ctx)
	require.NoError(t, err)

	// Verify params remain unchanged
	updatedParams := k.GetParams(ctx)
	require.Equal(t, uint32(0), updatedParams.MinimalLiquidityLock)
	require.Equal(t, types.DefaultMaxSlippage, updatedParams.MaxSlippage)
	require.Equal(t, initialParams, updatedParams) // The entire struct should be identical
}

func TestParams_CompleteWorkflow(t *testing.T) {
	// Test case: complete workflow of setting, getting, and migrating params

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)

	// 1. Start with default params (as initialized by DexKeeper)
	initial := k.GetParams(ctx)
	defaultParams := types.DefaultParams()
	require.Equal(t, defaultParams, initial)

	// 2. Update to custom params
	customParams := types.Params{
		NewPoolFeePct:        1000,
		CreationFee:          200000000,
		Beneficiary:          "test-beneficiary",
		MinimalLiquidityLock: 5000,
		MaxSlippage:          100,
	}
	require.NoError(t, k.SetParams(ctx, customParams))

	// 3. Verify retrieval
	retrieved := k.GetParams(ctx)
	require.Equal(t, customParams, retrieved)

	// 4. Run migration
	require.NoError(t, k.V2Migration(ctx))

	// 5. Verify migration changes - only specific fields should change
	migrated := k.GetParams(ctx)
	require.Equal(t, uint32(0), migrated.MinimalLiquidityLock)
	require.Equal(t, types.DefaultMaxSlippage, migrated.MaxSlippage)

	// Other fields should remain unchanged
	require.Equal(t, customParams.NewPoolFeePct, migrated.NewPoolFeePct)
	require.Equal(t, customParams.CreationFee, migrated.CreationFee)
	require.Equal(t, customParams.Beneficiary, migrated.Beneficiary)
}

func TestParams_CompleteWorkflow_2(t *testing.T) {
	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)

	// Clear any existing params to start from empty
	require.NoError(t, k.SetParams(ctx, types.Params{}))

	// 1. Start with empty params
	initial := k.GetParams(ctx)
	require.Equal(t, types.Params{}, initial)

	// 2. Set default params
	defaultParams := types.DefaultParams()
	require.NoError(t, k.SetParams(ctx, defaultParams))

	// 3. Verify retrieval
	retrieved := k.GetParams(ctx)
	require.Equal(t, defaultParams, retrieved)

	// 4. Run migration
	require.NoError(t, k.V2Migration(ctx))

	// 5. Verify migration changes
	migrated := k.GetParams(ctx)
	require.Equal(t, uint32(0), migrated.MinimalLiquidityLock)
	require.Equal(t, types.DefaultMaxSlippage, migrated.MaxSlippage)

	// Other fields should remain as they were in default params
	require.Equal(t, defaultParams.NewPoolFeePct, migrated.NewPoolFeePct)
	require.Equal(t, defaultParams.CreationFee, migrated.CreationFee)
	require.Equal(t, defaultParams.Beneficiary, migrated.Beneficiary)
}

func TestSetParams_ErrorHandling(t *testing.T) {
	// Test case: verify SetParams handles various scenarios without panicking
	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)

	// Test 1: Normal params should not error
	normalParams := types.DefaultParams()
	err := k.SetParams(ctx, normalParams)
	require.NoError(t, err)

	// Test 2: Zero-value params should not error
	zeroParams := types.Params{}
	err = k.SetParams(ctx, zeroParams)
	require.NoError(t, err)

	// Test 3: Params with extreme values should not error
	extremeParams := types.Params{
		NewPoolFeePct:        ^uint32(0), // max uint32
		CreationFee:          ^uint32(0),
		Beneficiary:          "very-long-beneficiary-address-string-that-is-extremely-long",
		MinimalLiquidityLock: ^uint32(0),
		MaxSlippage:          ^uint32(0),
	}
	err = k.SetParams(ctx, extremeParams)
	require.NoError(t, err)

	// Test 4: Verify the extreme params were stored correctly
	retrieved := k.GetParams(ctx)
	require.Equal(t, extremeParams, retrieved)
}
