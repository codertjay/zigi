package keeper_test

import (
	"context"
	"testing"

	keepertest "zigchain/testutil/keeper"
	"zigchain/x/factory/keeper"
	"zigchain/x/factory/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

func setupAuthTest(t *testing.T) (keeper.Keeper, context.Context, types.DenomAuth) {
	factoryKeeper, ctx := keepertest.FactoryKeeper(t, nil, nil)

	// Create a test denom auth with both bank and metadata admins
	denomAuth := types.DenomAuth{
		Denom:         "test-denom",
		BankAdmin:     "bank-admin-address",
		MetadataAdmin: "metadata-admin-address",
	}

	factoryKeeper.SetDenomAuth(ctx, denomAuth)

	return factoryKeeper, ctx, denomAuth
}

// Positive test cases

func TestAuthBankAdmin_Positive(t *testing.T) {
	// Test case: bank admin auth with a valid denom and admin address

	factoryKeeper, ctx, denomAuth := setupAuthTest(t)

	err := factoryKeeper.Auth(ctx, denomAuth.Denom, "bank", denomAuth.BankAdmin)
	require.NoError(t, err)
}

func TestAuthMetadataAdmin_Positive(t *testing.T) {
	// Test case: metadata admin auth with a valid denom and admin address

	factoryKeeper, ctx, denomAuth := setupAuthTest(t)

	// Test successful metadata admin auth
	err := factoryKeeper.Auth(ctx, denomAuth.Denom, "metadata", denomAuth.MetadataAdmin)
	require.NoError(t, err)
}

// Negative test cases

func TestAuthBankAdmin_DenomUnauthorized(t *testing.T) {
	// Test case: bank admin auth with an unauthorized address

	factoryKeeper, ctx, denomAuth := setupAuthTest(t)

	err := factoryKeeper.Auth(ctx, denomAuth.Denom, "bank", "unauthorized-address")
	require.Error(t, err)
	require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)
}

func TestAuthBankAdmin_DenomLocked(t *testing.T) {
	// Test case: bank admin auth with a locked denom

	factoryKeeper, ctx, denomAuth := setupAuthTest(t)

	denomAuth.BankAdmin = ""
	factoryKeeper.SetDenomAuth(ctx, denomAuth)
	err := factoryKeeper.Auth(ctx, denomAuth.Denom, "bank", "any-address")
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrDenomLocked)
}

func TestAuthMetadataAdmin(t *testing.T) {
	// Test case: metadata admin auth

	factoryKeeper, ctx, denomAuth := setupAuthTest(t)

	// Test successful metadata admin auth
	err := factoryKeeper.Auth(ctx, denomAuth.Denom, "metadata", denomAuth.MetadataAdmin)
	require.NoError(t, err)

	// Test successful bank admin auth for metadata operations
	err = factoryKeeper.Auth(ctx, denomAuth.Denom, "metadata", denomAuth.BankAdmin)
	require.NoError(t, err)

	// Test unauthorized metadata admin
	err = factoryKeeper.Auth(ctx, denomAuth.Denom, "metadata", "unauthorized-address")
	require.Error(t, err)
	require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)

	// Test metadata admin with disabled metadata admin (bank admin should still work)
	denomAuth.MetadataAdmin = ""
	factoryKeeper.SetDenomAuth(ctx, denomAuth)

	// Bank admin should still have access
	err = factoryKeeper.Auth(ctx, denomAuth.Denom, "metadata", denomAuth.BankAdmin)
	require.NoError(t, err)

	// Other addresses should be denied
	err = factoryKeeper.Auth(ctx, denomAuth.Denom, "metadata", "any-address")
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrDenomLocked)
}

func TestAuthInvalidCases_DenomAuthNotFound(t *testing.T) {
	// Test case: invalid cases for Auth function with non-existent denom

	factoryKeeper, ctx, _ := setupAuthTest(t)

	err := factoryKeeper.Auth(ctx, "non-existent-denom", "bank", "any-address")
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrDenomAuthNotFound)
}

func TestAuthInvalidCases_InvalidRequest(t *testing.T) {
	// Test case: invalid cases for Auth function with invalid admin type

	factoryKeeper, ctx, denomAuth := setupAuthTest(t)

	err := factoryKeeper.Auth(ctx, denomAuth.Denom, "invalid-type", "any-address")
	require.Error(t, err)
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
}

func TestAuthInvalidCases_EmptyDenom(t *testing.T) {
	// Test case: invalid cases for Auth function with empty denom

	factoryKeeper, ctx, _ := setupAuthTest(t)

	err := factoryKeeper.Auth(ctx, "", "bank", "any-address")
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrDenomAuthNotFound)
}
