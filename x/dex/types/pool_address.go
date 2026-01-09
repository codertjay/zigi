package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// GetPoolAddress Derive the pool account address from the pool ID
func GetPoolAddress(poolIDString string) sdk.AccAddress {
	return authtypes.NewModuleAddress(poolIDString)
}
