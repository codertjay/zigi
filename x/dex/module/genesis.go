package dex

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/x/dex/keeper"
	"zigchain/x/dex/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set all the pool
	for _, elem := range genState.PoolList {
		k.SetPool(ctx, elem)
	}
	// This will set data into memory if available in store
	if genState.PoolsMeta != nil {
		k.SetPoolsMeta(ctx, *genState.PoolsMeta)
	}
	// Set all the poolUids
	for _, elem := range genState.PoolUidsList {
		k.SetPoolUids(ctx, elem)
	}
	// this line is used by starport scaffolding # genesis/module/init
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}
}

// ExportGenesis returns the module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	genesis.PoolList = k.GetAllPool(ctx)
	// Get all poolsMeta
	poolsMeta, found := k.GetPoolsMeta(ctx)
	if found {
		genesis.PoolsMeta = &poolsMeta
	}
	genesis.PoolUidsList = k.GetAllPoolUids(ctx)
	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
