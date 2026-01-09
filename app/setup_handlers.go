package app

import (
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/app/upgrades"
	v1 "zigchain/app/upgrades/v1"
	v2 "zigchain/app/upgrades/v2"
)

var (
	Upgrades = []upgrades.Upgrade{
		v1.Upgrade,
		v2.Upgrade,
	}
	Forks = []upgrades.Fork{}
)

func (app *App) setupUpgradeHandlers() {
	for _, upgrade := range Upgrades {
		app.UpgradeKeeper.SetUpgradeHandler(
			upgrade.UpgradeName,
			upgrade.CreateUpgradeHandler(
				app.ModuleManager,
				app.Configurator(),
				&app.AppKeepers,
			),
		)
	}
}

// configure store loader that checks if version == upgradeHeight and applies store upgrades
func (app *App) setupUpgradeStoreLoaders() {
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Sprintf("failed to read upgrade info from disk %s", err))
	}

	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	// Use Debug level for logs that shouldn't appear in genesis file
	app.Logger().Debug("Upgrade info", "info", upgradeInfo)

	for _, upgrade := range Upgrades {
		upgrade := upgrade
		if upgradeInfo.Name == upgrade.UpgradeName {
			storeUpgrades := upgrade.StoreUpgrades

			// Use Debug level instead of Info to prevent inclusion in genesis file
			app.Logger().Debug(fmt.Sprintf("Setting store loader with height %d and store upgrades: %+v", upgradeInfo.Height, storeUpgrades))

			// Use upgrade store loader for the initial loading of all stores when app starts,
			// it checks if version == upgradeHeight and applies store upgrades before loading the stores,
			// so that new stores start with the correct version (the current height of chain),
			// instead the default which is the latest version that store last committed i.e 0 for new stores.
			app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
		} else {
			app.Logger().Debug("No need to load upgrade store.")
		}
	}
}

// PreBlockForks is intended to be ran in a chain upgrade.
func PreBlockForks(ctx sdk.Context, app *App) {
	for _, fork := range Forks {
		if ctx.BlockHeight() == fork.UpgradeHeight {
			fork.PreBlockForkLogic(ctx, &app.AppKeepers)
			return
		}
	}
}
