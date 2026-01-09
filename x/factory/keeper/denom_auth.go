package keeper

import (
	"context"
	"zigchain/x/factory/types"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
)

// SetDenomAuth set a specific denomAuth in the store from its index
func (k Keeper) SetDenomAuth(ctx context.Context, denomAuth types.DenomAuth) {
	// Store the original denom auth
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	denomStore := prefix.NewStore(store, types.KeyPrefix(types.DenomAuthKeyPrefix))
	bz := k.cdc.MustMarshal(&denomAuth)
	denomStore.Set(types.DenomAuthKey(denomAuth.Denom), bz)
}

// GetDenomAuth returns a denomAuth from its index
func (k Keeper) GetDenomAuth(
	ctx context.Context,
	denom string,
) (val types.DenomAuth, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.DenomAuthKeyPrefix))

	b := store.Get(types.DenomAuthKey(
		denom,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)

	if val.Denom == "" {
		return val, false
	}
	return val, true
}

// SetProposedDenomAuth sets the proposed denom auth
func (k Keeper) SetProposedDenomAuth(ctx context.Context, denomAuth types.DenomAuth) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.ProposedDenomAuthKeyPrefix))
	b := k.cdc.MustMarshal(&denomAuth)
	store.Set(types.DenomAuthKey(
		denomAuth.Denom,
	), b)
}

// DeleteProposedDenomAuth deletes the proposed denom auth
func (k Keeper) DeleteProposedDenomAuth(ctx context.Context, denom string) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.ProposedDenomAuthKeyPrefix))
	store.Delete(types.DenomAuthKey(denom))
}

// GetProposedDenomAuth returns the proposed denom auth
func (k Keeper) GetProposedDenomAuth(
	ctx context.Context,
	denom string,
) (val types.DenomAuth, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.ProposedDenomAuthKeyPrefix))

	b := store.Get(types.DenomAuthKey(
		denom,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)

	if val.Denom == "" {
		return val, false
	}
	return val, true
}

// GetAllDenomAuth returns all denomAuth
func (k Keeper) GetAllDenomAuth(ctx context.Context) (list []types.DenomAuth) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.DenomAuthKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer func(iterator storetypes.Iterator) {
		err := iterator.Close()
		if err != nil {
			k.Logger().Error("Error closing iterator", "error", err)
		}
	}(iterator)

	for ; iterator.Valid(); iterator.Next() {
		var val types.DenomAuth
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// DisableDenomAuth disables the denom admin role
func (k Keeper) DisableDenomAuth(ctx context.Context, denom string) error {
	denomAuth, isFound := k.GetDenomAuth(
		ctx,
		denom,
	)
	if !isFound {
		return errorsmod.Wrapf(
			types.ErrDenomAuthNotFound,
			"Denom: (%s)",
			denom,
		)
	}

	// Disable the denom admin
	denomAuth.BankAdmin = ""

	// Store the updated denom auth
	k.SetDenomAuth(ctx, denomAuth)

	return nil
}

// AddDenomToAdminDenomAuthList adds a denom to an admin's index
func (k Keeper) AddDenomToAdminDenomAuthList(ctx context.Context, admin string, denom string) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.AdminDenomAuthListKey(admin))

	store.Set(types.DenomAuthNameKey(denom), []byte{})
}

// RemoveDenomFromAdminDenomAuthList removes a denom from an admin's index
func (k Keeper) RemoveDenomFromAdminDenomAuthList(ctx context.Context, admin string, denom string) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.AdminDenomAuthListKey(admin))

	store.Delete(types.DenomAuthNameKey(denom))
}

// MigrateAdminDenomAuthList creates the admin index for existing denom auth entries
func (k Keeper) MigrateAdminDenomAuthList(ctx context.Context) error {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.DenomAuthKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer func(iterator storetypes.Iterator) {
		err := iterator.Close()
		if err != nil {
			k.Logger().Error("Error closing iterator", "error", err)
		}
	}(iterator)

	for ; iterator.Valid(); iterator.Next() {
		var denomAuth types.DenomAuth
		if err := k.cdc.Unmarshal(iterator.Value(), &denomAuth); err != nil {
			return err
		}

		// Add both admins to admin index
		k.AddDenomToAdminDenomAuthList(ctx, denomAuth.BankAdmin, denomAuth.Denom)
		k.AddDenomToAdminDenomAuthList(ctx, denomAuth.MetadataAdmin, denomAuth.Denom)
	}

	return nil
}
