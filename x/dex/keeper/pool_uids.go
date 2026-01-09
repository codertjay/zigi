package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/x/dex/types"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
)

// SetPoolUids set a specific poolUids in the store from its index
func (k Keeper) SetPoolUids(ctx context.Context, poolUids types.PoolUids) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.PoolUidsKeyPrefix))
	b := k.cdc.MustMarshal(&poolUids)
	store.Set(types.PoolUidsKey(
		poolUids.PoolUid,
	), b)
}

// SetPoolUidFromPool proxy method that sets unique poolUids from a pool
func (k Keeper) SetPoolUidFromPool(ctx context.Context, pool types.Pool) {

	poolUids := types.PoolUids{
		PoolUid: types.GetPoolUidString(pool),
		PoolId:  pool.PoolId,
	}
	k.SetPoolUids(ctx, poolUids)

}

// GetPoolUidsFromCoins returns a poolUids from its denoms
func (k Keeper) GetPoolUidsFromCoins(
	ctx context.Context,
	coins sdk.Coins,
) (
	poolUid types.PoolUids,
	found bool,
) {

	if len(coins) != 2 {
		return poolUid, false
	}
	poolUidString := coins[0].Denom + types.PoolUidSeparator + coins[1].Denom
	return k.GetPoolUids(ctx, poolUidString)
}

// GetPoolUids returns a poolUids from its index
func (k Keeper) GetPoolUids(
	ctx context.Context,
	poolUid string,

) (val types.PoolUids, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.PoolUidsKeyPrefix))

	b := store.Get(types.PoolUidsKey(
		poolUid,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// GetPoolUidsFromPool proxy method that gets unique poolUids from a pool
func (k Keeper) GetPoolUidsFromPool(ctx context.Context, pool types.Pool) (val types.PoolUids, found bool) {

	return k.GetPoolUids(ctx, types.GetPoolUidString(pool))

}

// RemovePoolUids removes a poolUids from the store
func (k Keeper) RemovePoolUids(
	ctx context.Context,
	poolUid string,

) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.PoolUidsKeyPrefix))
	store.Delete(types.PoolUidsKey(
		poolUid,
	))
}

// GetAllPoolUids returns all poolUids
func (k Keeper) GetAllPoolUids(ctx context.Context) (list []types.PoolUids) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.PoolUidsKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer func(iterator storetypes.Iterator) {
		err := iterator.Close()
		if err != nil {
			k.logger.Error("failed to close iterator", "error", err)
		}
	}(iterator)

	for ; iterator.Valid(); iterator.Next() {
		var val types.PoolUids
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
