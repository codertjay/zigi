package keeper

import (
	"context"

	"zigchain/x/dex/types"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
)

// SetPoolsMeta set poolsMeta in the store
func (k Keeper) SetPoolsMeta(ctx context.Context, poolsMeta types.PoolsMeta) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.PoolsMetaKey))
	b := k.cdc.MustMarshal(&poolsMeta)
	store.Set([]byte{0}, b)
}

// GetPoolsMeta returns poolsMeta
func (k Keeper) GetPoolsMeta(ctx context.Context) (val types.PoolsMeta, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.PoolsMetaKey))

	b := store.Get([]byte{0})
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemovePoolsMeta removes poolsMeta from the store
func (k Keeper) RemovePoolsMeta(ctx context.Context) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.PoolsMetaKey))
	store.Delete([]byte{0})
}
