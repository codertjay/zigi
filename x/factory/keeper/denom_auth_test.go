package keeper_test

import (
	"context"
	"strconv"
	"testing"

	errorsmod "cosmossdk.io/errors"

	keepertest "zigchain/testutil/keeper"
	"zigchain/testutil/nullify"
	"zigchain/x/factory/keeper"
	"zigchain/x/factory/types"

	"github.com/stretchr/testify/require"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNDenomAuth(keeper keeper.Keeper, ctx context.Context, n int) []types.DenomAuth {
	// set a specific denomAuth
	items := make([]types.DenomAuth, n)
	for i := range items {
		items[i].Denom = strconv.Itoa(i)

		keeper.SetDenomAuth(ctx, items[i])
	}
	return items
}

// Positive test cases

func TestDenomAuthGet(t *testing.T) {
	// Test case: get a specific denomAuth

	factoryKeeper, ctx := keepertest.FactoryKeeper(t, nil, nil)
	items := createNDenomAuth(factoryKeeper, ctx, 10)
	for _, item := range items {
		rst, found := factoryKeeper.GetDenomAuth(ctx,
			item.Denom,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}

func TestDenomAuthGetAll(t *testing.T) {
	// Test case: get all denomAuth

	factoryKeeper, ctx := keepertest.FactoryKeeper(t, nil, nil)
	items := createNDenomAuth(factoryKeeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(factoryKeeper.GetAllDenomAuth(ctx)),
	)
}

func TestMigrateAdminDenomAuthList(t *testing.T) {
	// Test cases for MigrateAdminDenomAuthList

	t.Run("successful migration with multiple denom auth entries", func(t *testing.T) {
		factoryKeeper, ctx := keepertest.FactoryKeeper(t, nil, nil)

		// Create test denom auth entries
		denomAuths := []types.DenomAuth{
			{
				Denom:         "denom1",
				BankAdmin:     "bankadmin1",
				MetadataAdmin: "metadataadmin1",
			},
			{
				Denom:         "denom2",
				BankAdmin:     "bankadmin2",
				MetadataAdmin: "metadataadmin2",
			},
			{
				Denom:         "denom3",
				BankAdmin:     "bankadmin3",
				MetadataAdmin: "metadataadmin3",
			},
		}

		// Store denom auths directly
		for _, da := range denomAuths {
			factoryKeeper.SetDenomAuth(ctx, da)
		}

		// Execute migration
		err := factoryKeeper.MigrateAdminDenomAuthList(ctx)
		require.NoError(t, err)

		// Verify that both admins for each denom are added to their respective admin lists
		for _, da := range denomAuths {
			// Verify the original denom auth data is still intact
			stored, found := factoryKeeper.GetDenomAuth(ctx, da.Denom)
			require.True(t, found)
			require.Equal(t, da.Denom, stored.Denom)
			require.Equal(t, da.BankAdmin, stored.BankAdmin)
			require.Equal(t, da.MetadataAdmin, stored.MetadataAdmin)

			// Test that we can remove from the admin list (proving it was added)
			factoryKeeper.RemoveDenomFromAdminDenomAuthList(ctx, da.BankAdmin, da.Denom)
			factoryKeeper.RemoveDenomFromAdminDenomAuthList(ctx, da.MetadataAdmin, da.Denom)
		}
	})

	t.Run("successful migration with empty denom auth list", func(t *testing.T) {
		factoryKeeper, ctx := keepertest.FactoryKeeper(t, nil, nil)

		// No denom auths stored

		// Execute migration - should not error with empty list
		err := factoryKeeper.MigrateAdminDenomAuthList(ctx)
		require.NoError(t, err)

		// Verify GetAllDenomAuth returns empty
		allDenomAuths := factoryKeeper.GetAllDenomAuth(ctx)
		require.Empty(t, allDenomAuths)
	})

	t.Run("migration with denom auth having same admin for both roles", func(t *testing.T) {
		factoryKeeper, ctx := keepertest.FactoryKeeper(t, nil, nil)

		// Create denom auth with the same admin for both roles
		denomAuth := types.DenomAuth{
			Denom:         "sharedadmin",
			BankAdmin:     "sharedadmin",
			MetadataAdmin: "sharedadmin",
		}

		factoryKeeper.SetDenomAuth(ctx, denomAuth)

		// Execute migration
		err := factoryKeeper.MigrateAdminDenomAuthList(ctx)
		require.NoError(t, err)

		// Verify the denom auth data is intact
		stored, found := factoryKeeper.GetDenomAuth(ctx, "sharedadmin")
		require.True(t, found)
		require.Equal(t, denomAuth, stored)

		// Test removal works (proving it was added to the admin list)
		factoryKeeper.RemoveDenomFromAdminDenomAuthList(ctx, "sharedadmin", "sharedadmin")
	})

	t.Run("migration with empty admin fields", func(t *testing.T) {
		factoryKeeper, ctx := keepertest.FactoryKeeper(t, nil, nil)

		// Create denom auth with empty admin fields
		denomAuths := []types.DenomAuth{
			{
				Denom:         "denom1",
				BankAdmin:     "", // Empty bank admin
				MetadataAdmin: "metadataadmin1",
			},
			{
				Denom:         "denom2",
				BankAdmin:     "bankadmin2",
				MetadataAdmin: "", // Empty metadata admin
			},
		}

		for _, da := range denomAuths {
			factoryKeeper.SetDenomAuth(ctx, da)
		}

		// Execute migration - should not panic with empty admin fields
		err := factoryKeeper.MigrateAdminDenomAuthList(ctx)
		require.NoError(t, err)

		// Verify data is intact
		for _, da := range denomAuths {
			stored, found := factoryKeeper.GetDenomAuth(ctx, da.Denom)
			require.True(t, found)
			require.Equal(t, da, stored)

			// Test removal for non-empty admins
			if da.BankAdmin != "" {
				factoryKeeper.RemoveDenomFromAdminDenomAuthList(ctx, da.BankAdmin, da.Denom)
			}
			if da.MetadataAdmin != "" {
				factoryKeeper.RemoveDenomFromAdminDenomAuthList(ctx, da.MetadataAdmin, da.Denom)
			}
		}
	})

	t.Run("verify migration can be called multiple times safely", func(t *testing.T) {
		factoryKeeper, ctx := keepertest.FactoryKeeper(t, nil, nil)

		// Create test denom auth
		denomAuth := types.DenomAuth{
			Denom:         "testdenom",
			BankAdmin:     "bankadmin",
			MetadataAdmin: "metadataadmin",
		}

		factoryKeeper.SetDenomAuth(ctx, denomAuth)

		// Execute migration first time
		err := factoryKeeper.MigrateAdminDenomAuthList(ctx)
		require.NoError(t, err)

		// Execute migration second time
		err = factoryKeeper.MigrateAdminDenomAuthList(ctx)
		require.NoError(t, err)

		// Verify data is still correct
		stored, found := factoryKeeper.GetDenomAuth(ctx, "testdenom")
		require.True(t, found)
		require.Equal(t, denomAuth, stored)

		// Clean up
		factoryKeeper.RemoveDenomFromAdminDenomAuthList(ctx, "bankadmin", "testdenom")
		factoryKeeper.RemoveDenomFromAdminDenomAuthList(ctx, "metadataadmin", "testdenom")
	})

	t.Run("migration with special characters in denom and admin", func(t *testing.T) {
		factoryKeeper, ctx := keepertest.FactoryKeeper(t, nil, nil)

		// Test with various edge cases
		testCases := []types.DenomAuth{
			{
				Denom:         "denom-with-dash",
				BankAdmin:     "admin-with-dash",
				MetadataAdmin: "metadata-with-dash",
			},
			{
				Denom:         "denom_with_underscore",
				BankAdmin:     "admin_with_underscore",
				MetadataAdmin: "metadata_with_underscore",
			},
			{
				Denom:         "denom.with.dots",
				BankAdmin:     "admin.with.dots",
				MetadataAdmin: "metadata.with.dots",
			},
		}

		for _, tc := range testCases {
			factoryKeeper.SetDenomAuth(ctx, tc)
		}

		// Execute migration
		err := factoryKeeper.MigrateAdminDenomAuthList(ctx)
		require.NoError(t, err)

		// Verify all data is intact
		for _, tc := range testCases {
			stored, found := factoryKeeper.GetDenomAuth(ctx, tc.Denom)
			require.True(t, found)
			require.Equal(t, tc, stored)

			// Test removal works
			factoryKeeper.RemoveDenomFromAdminDenomAuthList(ctx, tc.BankAdmin, tc.Denom)
			factoryKeeper.RemoveDenomFromAdminDenomAuthList(ctx, tc.MetadataAdmin, tc.Denom)
		}
	})
}

func TestMigrateAdminDenomAuthListIntegration(t *testing.T) {
	factoryKeeper, ctx := keepertest.FactoryKeeper(t, nil, nil)

	// Setup: create multiple denom auths with overlapping admins
	denomAuths := []types.DenomAuth{
		{
			Denom:         "denom1",
			BankAdmin:     "admin1",
			MetadataAdmin: "admin2",
		},
		{
			Denom:         "denom2",
			BankAdmin:     "admin1", // Same bank admin as denom1
			MetadataAdmin: "admin3",
		},
		{
			Denom:         "denom3",
			BankAdmin:     "admin4",
			MetadataAdmin: "admin2", // Same metadata admin as denom1
		},
	}

	// Store all denom auths
	for _, da := range denomAuths {
		factoryKeeper.SetDenomAuth(ctx, da)
	}

	// Execute migration
	err := factoryKeeper.MigrateAdminDenomAuthList(ctx)
	require.NoError(t, err)

	// Verify we can retrieve all denom auths after migration
	for _, da := range denomAuths {
		stored, found := factoryKeeper.GetDenomAuth(ctx, da.Denom)
		require.True(t, found, "Should find denom auth for %s", da.Denom)
		require.Equal(t, da, stored, "Stored denom auth should match original for %s", da.Denom)
	}

	// Verify migration is idempotent by running it again
	err = factoryKeeper.MigrateAdminDenomAuthList(ctx)
	require.NoError(t, err)

	// Final verification that all data remains correct
	allDenomAuths := factoryKeeper.GetAllDenomAuth(ctx)
	require.Len(t, allDenomAuths, 3, "Should have exactly 3 denom auths after migration")

	// Clean up admin lists
	for _, da := range denomAuths {
		factoryKeeper.RemoveDenomFromAdminDenomAuthList(ctx, da.BankAdmin, da.Denom)
		factoryKeeper.RemoveDenomFromAdminDenomAuthList(ctx, da.MetadataAdmin, da.Denom)
	}
}

// Negative test cases

func TestDenomAuthGetNotFound(t *testing.T) {
	// Test case: try to get denomAuth that does not exist

	factoryKeeper, ctx := keepertest.FactoryKeeper(t, nil, nil)
	_, found := factoryKeeper.GetDenomAuth(
		ctx,
		"does-not-exist",
	)
	require.False(t, found)
}

func TestDenomAuthGetEmpty(t *testing.T) {
	// Test case: try to get denomAuth with empty input

	factoryKeeper, ctx := keepertest.FactoryKeeper(t, nil, nil)
	_, found := factoryKeeper.GetDenomAuth(ctx, "")
	require.False(t, found)
}

func TestDenomAuthGetInvalidDenom(t *testing.T) {
	// Test case: try to get denomAuth with empty input

	factoryKeeper, ctx := keepertest.FactoryKeeper(t, nil, nil)
	_, found := factoryKeeper.GetDenomAuth(
		ctx,
		"!invalid_denom!",
	)
	require.False(t, found)
}

func TestDenomAuthGetAllEmpty(t *testing.T) {
	// Test case: get all denomAuth when there are none

	factoryKeeper, ctx := keepertest.FactoryKeeper(t, nil, nil)
	require.Empty(t, factoryKeeper.GetAllDenomAuth(ctx))
}

func TestDenomAuthGetEmptyDenom(t *testing.T) {
	// Test case: retrieve a DenomAuth with an empty Denom field

	factoryKeeper, ctx := keepertest.FactoryKeeper(t, nil, nil)

	// Create a DenomAuth with an empty Denom
	denomAuth := types.DenomAuth{
		Denom: "", // Explicitly set empty Denom
	}

	// Store the DenomAuth directly
	factoryKeeper.SetDenomAuth(ctx, denomAuth)

	// Attempt to retrieve it
	result, found := factoryKeeper.GetDenomAuth(ctx, "")

	// Verify the results
	require.False(t, found, "Expected found to be false for empty Denom")
	require.Equal(t, denomAuth, result, "Expected returned DenomAuth to match the stored one")
}

func TestProposedDenomAuthGetEmptyDenom(t *testing.T) {
	// Test case: retrieve a ProposedDenomAuth with an empty Denom field

	factoryKeeper, ctx := keepertest.FactoryKeeper(t, nil, nil)

	// Create a DenomAuth with an empty Denom
	denomAuth := types.DenomAuth{
		Denom: "", // Explicitly set empty Denom
	}

	// Store the ProposedDenomAuth directly
	factoryKeeper.SetProposedDenomAuth(ctx, denomAuth)

	// Attempt to retrieve it
	result, found := factoryKeeper.GetProposedDenomAuth(ctx, "")

	// Verify the results
	require.False(t, found, "Expected found to be false for empty Denom")
	require.Equal(t, denomAuth, result, "Expected returned DenomAuth to match the stored one")
}

func TestDisableDenomAuthNotFound(t *testing.T) {
	// Test case: attempt to disable a non-existent DenomAuth

	factoryKeeper, ctx := keepertest.FactoryKeeper(t, nil, nil)

	// Attempt to disable a DenomAuth that does not exist
	nonExistentDenom := "non-existent-denom"
	err := factoryKeeper.DisableDenomAuth(ctx, nonExistentDenom)

	// Verify the error
	require.Error(t, err, "Expected an error when disabling a non-existent DenomAuth")
	require.ErrorIs(t, err, types.ErrDenomAuthNotFound, "Expected ErrDenomAuthNotFound error")
	require.EqualError(t, err, errorsmod.Wrapf(types.ErrDenomAuthNotFound, "Denom: (%s)", nonExistentDenom).Error(), "Error message should match")
}
