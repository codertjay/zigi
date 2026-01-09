package keeper_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "zigchain/testutil/keeper"
	"zigchain/testutil/nullify"
	"zigchain/x/factory/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

// Positive test case

func TestDenomGetAll(t *testing.T) {
	// Test case: get all denoms

	keeper, ctx := keepertest.FactoryKeeper(t, nil, nil)
	items := createNDenom(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllDenom(ctx)),
	)
}

func TestDenomGetAllNil(t *testing.T) {
	// Test case: get all denoms when empty

	keeper, ctx := keepertest.FactoryKeeper(t, nil, nil)
	require.Empty(t, keeper.GetAllDenom(ctx))
}

func TestGetAllLegacyDenom_ReturnType(t *testing.T) {
	// Test case: verify return type of GetAllLegacyDenom
	factoryKeeper, ctx := keepertest.FactoryKeeper(t, nil, nil)

	result := factoryKeeper.GetAllLegacyDenom(ctx)

	// Verify it returns the correct type
	var expectedType []types.LegacyDenom
	require.IsType(t, expectedType, result)
}

func TestGetAllLegacyDenom_Idempotent(t *testing.T) {
	// Test case: verify idempotency of GetAllLegacyDenom
	factoryKeeper, ctx := keepertest.FactoryKeeper(t, nil, nil)

	// First call
	result1 := factoryKeeper.GetAllLegacyDenom(ctx)

	// The Second call should work the same way
	result2 := factoryKeeper.GetAllLegacyDenom(ctx)

	// Results should be equivalent
	require.Equal(t, len(result1), len(result2))
	require.ElementsMatch(t, result1, result2)
}
