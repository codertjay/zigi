package keeper

import (
	"zigchain/x/factory/types"
)

var _ types.QueryServer = Keeper{}
