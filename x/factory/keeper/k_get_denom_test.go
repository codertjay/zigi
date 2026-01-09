package keeper_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "zigchain/testutil/keeper"
	"zigchain/testutil/nullify"
)

// Prevent strconv unused error
var _ = strconv.IntSize

// Positive test cases

func TestDenomGet(t *testing.T) {
	// Test case: get a denom

	// get access to the keeper and context
	keeper, ctx := keepertest.FactoryKeeper(t, nil, nil)

	// create a slice of 10 Denom
	items := createNDenom(keeper, ctx, 10)

	// loop over the slice and get each Denom
	for _, item := range items {
		// get the Denom from the keeper using the Denom field of the item (as a store key)
		denom, found := keeper.GetDenom(
			ctx,
			item.Denom,
		)

		// make sure the Denom was found
		require.True(t, found)

		// make sure the Denom is the same as the item that was set to the store
		// nullify.Fill will initialize the empty fields of the item with the zero value of the field type
		// int and float will be 0, string will be "", etc.
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&denom),
		)
	}
}

// Negative test cases

func TestDenomGetNonExisting(t *testing.T) {
	// Test case: try to get a non-existing denom

	keeper, ctx := keepertest.FactoryKeeper(t, nil, nil)
	_, found := keeper.GetDenom(
		ctx,
		"does-not-exist",
	)
	require.False(t, found)
}

func TestDenomGetEmpty(t *testing.T) {
	// Test case: try to get an empty denom

	keeper, ctx := keepertest.FactoryKeeper(t, nil, nil)
	_, found := keeper.GetDenom(
		ctx,
		"",
	)
	require.False(t, found)
}
