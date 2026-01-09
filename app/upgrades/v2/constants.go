package v2 //nolint:revive

import (
	storetypes "cosmossdk.io/store/types"

	"zigchain/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v2"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		// Added: []string{},
		// Renamed: []storetypes.StoreRename{},
		Deleted: []string{"crisis"},
	},
}
