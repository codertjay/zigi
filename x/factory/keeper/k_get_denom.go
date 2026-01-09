// This file help with low level operation on the store for the denom module
package keeper

import (
	"context"

	"zigchain/x/factory/types"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
)

// GetDenom returns a denom from its index
func (k Keeper) GetDenom(
	ctx context.Context,
	denom string,

) (val types.Denom, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.DenomKeyPrefix))

	b := store.Get(types.DenomKey(
		denom,
	))
	if b == nil {

		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}
