// This file help with low level operation on the store for the denom module
package keeper

import (
	"context"

	"zigchain/x/factory/types"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
)

// SetDenom set a specific denom in the store from its index
func (k Keeper) SetDenom(
	ctx context.Context,
	denom types.Denom,
) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.DenomKeyPrefix))
	b := k.cdc.MustMarshal(&denom)
	store.Set(types.DenomKey(
		denom.Denom,
	), b)
}

// Migrate legacy denom to new denom structure
func (k Keeper) V2Migration(ctx context.Context) error {
	// Migrate legacy denoms to new denom structure
	legacyDenoms := k.GetAllLegacyDenom(ctx)
	for _, legacyDenom := range legacyDenoms {
		denom := types.Denom{
			Creator:             legacyDenom.Creator,
			Denom:               legacyDenom.Denom,
			MintingCap:          legacyDenom.MaxSupply,
			Minted:              legacyDenom.Minted,
			CanChangeMintingCap: legacyDenom.CanChangeMaxSupply,
		}
		k.SetDenom(ctx, denom)
	}

	// Migrate existing denom auth to admin index
	// #nosec G104 -- migration logic, error handling not critical
	k.MigrateAdminDenomAuthList(ctx)

	return nil
}
