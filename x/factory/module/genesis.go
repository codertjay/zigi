package factory

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/x/factory/keeper"
	"zigchain/x/factory/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set all the denom
	for _, elem := range genState.DenomList {
		k.SetDenom(ctx, elem)
	}
	// Set all the denomAuth
	for _, elem := range genState.DenomAuthList {
		k.SetDenomAuth(ctx, elem)

		// Add the metadata admin to the admin denom auth list
		k.AddDenomToAdminDenomAuthList(ctx, elem.MetadataAdmin, elem.Denom)

		// Add the bank admin to the admin denom auth list
		k.AddDenomToAdminDenomAuthList(ctx, elem.BankAdmin, elem.Denom)
	}
	// this line is used by starport scaffolding # genesis/module/init
	err := k.SetParams(ctx, genState.Params)
	if err != nil {
		panic(err)
	}
}

// ExportGenesis returns the module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	genesis.DenomList = k.GetAllDenom(ctx)
	genesis.DenomAuthList = k.GetAllDenomAuth(ctx)
	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
