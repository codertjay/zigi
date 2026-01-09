package keeper

import (
	"zigchain/x/tokenwrapper/types"
)

var _ types.QueryServer = Keeper{}
