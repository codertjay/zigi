package migrations

import (
	"zigchain/x/tokenwrapper/keeper"
)

type Migrator struct {
	keeper keeper.Keeper
}

func NewMigrator(keeper keeper.Keeper) Migrator {
	return Migrator{keeper: keeper}
}
