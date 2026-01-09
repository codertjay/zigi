package wasmbinding

import (
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	dexkeeper "zigchain/x/dex/keeper"
	factorykeeper "zigchain/x/factory/keeper"
)

// NewQueryPlugin returns a reference to a new QueryPlugin.
func NewQueryPlugin(
	bk *bankkeeper.BaseKeeper,
	fk *factorykeeper.Keeper,
	dk *dexkeeper.Keeper,
) *QueryPlugin {
	return &QueryPlugin{
		bankKeeper:    bk,
		factoryKeeper: fk,
		dexKeeper:     dk,
	}
}

type QueryPlugin struct {
	bankKeeper    *bankkeeper.BaseKeeper
	factoryKeeper *factorykeeper.Keeper
	dexKeeper     *dexkeeper.Keeper
}
