package keeper_test

import (
	"context"
	"strconv"
	"testing"

	cosmosmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	keepertest "zigchain/testutil/keeper"
	"zigchain/testutil/sample"
	"zigchain/x/factory/keeper"
	"zigchain/x/factory/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

// createNDenom creates a slice of n Denom for testing purposes
func createNDenom(keeper keeper.Keeper, ctx context.Context, n int) []types.Denom {

	// create a slice of length n with empty structs of type types.Denom
	items := make([]types.Denom, n)
	auths := make([]types.DenomAuth, n)

	address := sample.AccAddress()
	// Loop over the slice and set the Denom field of each struct to the index of the slice
	for i := range items {
		fullDenom := "coin" + types.FactoryDenomDelimiterChar + address + types.FactoryDenomDelimiterChar + "abc-" + strconv.Itoa(i)
		// Set the Denom fields appending number on the end 0,1,2...
		items[i].Denom = fullDenom
		items[i].MintingCap = cosmosmath.NewUint(1000000).Add(cosmosmath.NewUint(uint64(i)))
		items[i].Minted = cosmosmath.NewUint(1000).Add(cosmosmath.NewUint(uint64(i)))
		// alternate true vs false
		items[i].CanChangeMintingCap = i%2 == 0
		items[i].Creator = address

		auths[i].Denom = fullDenom
		auths[i].BankAdmin = address
		auths[i].MetadataAdmin = address

		keeper.SetDenomAuth(ctx, auths[i])
		// Save the Denom to the store
		// Validators are not applied on this level, so it will save what you set it to save
		keeper.SetDenom(ctx, items[i])

	}

	// Return the slice
	return items
}

func TestV2Migration(t *testing.T) {
	// Test the V2Migration function of the keeper

	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	// Create some legacy denoms for migration
	legacyDenoms := []types.LegacyDenom{
		{
			Creator:            sample.AccAddress(),
			Denom:              "factory/creator1/abc",
			MaxSupply:          cosmosmath.NewUint(1000000),
			Minted:             cosmosmath.NewUint(50000),
			CanChangeMaxSupply: true,
		},
		{
			Creator:            sample.AccAddress(),
			Denom:              "factory/creator2/def",
			MaxSupply:          cosmosmath.NewUint(2000000),
			Minted:             cosmosmath.NewUint(100000),
			CanChangeMaxSupply: false,
		},
	}

	// Set legacy denoms
	for _, legacyDenom := range legacyDenoms {
		k.SetDenom(ctx, types.Denom{
			Creator:             legacyDenom.Creator,
			Denom:               legacyDenom.Denom,
			MintingCap:          legacyDenom.MaxSupply,
			Minted:              legacyDenom.Minted,
			CanChangeMintingCap: legacyDenom.CanChangeMaxSupply,
		})
	}

	// Create some existing denom auths to migrate
	denomAuths := []types.DenomAuth{
		{
			Denom:         "factory/creator1/abc",
			BankAdmin:     sample.AccAddress(),
			MetadataAdmin: sample.AccAddress(),
		},
		{
			Denom:         "factory/creator2/def",
			BankAdmin:     sample.AccAddress(),
			MetadataAdmin: sample.AccAddress(),
		},
	}

	for _, auth := range denomAuths {
		k.SetDenomAuth(ctx, auth)
	}

	// Execute migration
	err := k.V2Migration(ctx)
	require.NoError(t, err)

	// Verify legacy denoms were migrated to new denom structure
	for _, legacyDenom := range legacyDenoms {
		migratedDenom, found := k.GetDenom(ctx, legacyDenom.Denom)
		require.True(t, found)

		require.Equal(t, legacyDenom.Creator, migratedDenom.Creator)
		require.Equal(t, legacyDenom.Denom, migratedDenom.Denom)
		require.Equal(t, legacyDenom.MaxSupply, migratedDenom.MintingCap)
		require.Equal(t, legacyDenom.Minted, migratedDenom.Minted)
		require.Equal(t, legacyDenom.CanChangeMaxSupply, migratedDenom.CanChangeMintingCap)
	}
}

func TestV2Migration_EmptyLegacyDenoms(t *testing.T) {
	// Test the V2Migration function with no legacy denoms

	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	// Execute migration with no legacy denoms
	err := k.V2Migration(ctx)
	require.NoError(t, err)
}
