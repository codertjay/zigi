// This file help with low level operation on the store for the denom module
package keeper

import (
	"context"

	storetypes "cosmossdk.io/store/types"

	"zigchain/x/factory/types"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
)

// GetAllDenom returns all denom
func (k Keeper) GetAllDenom(ctx context.Context) (list []types.Denom) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.DenomKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer func(iterator storetypes.Iterator) {
		err := iterator.Close()
		if err != nil {
			k.Logger().Error("Error closing iterator", "error", err)
		}
	}(iterator)

	for ; iterator.Valid(); iterator.Next() {
		var val types.Denom
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// GetAllLegacyDenom returns all legacy denoms
func (k Keeper) GetAllLegacyDenom(ctx context.Context) (list []types.LegacyDenom) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.DenomKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer func(iterator storetypes.Iterator) {
		err := iterator.Close()
		if err != nil {
			k.Logger().Error("Error closing iterator", "error", err)
		}
	}(iterator)

	for ; iterator.Valid(); iterator.Next() {
		var val types.LegacyDenom
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
