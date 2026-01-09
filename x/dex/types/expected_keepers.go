package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MintKeeper defines the expected interface for the Mint module.
type MintKeeper interface {
	MintCoins(context.Context, sdk.Coins) error
	// Methods imported from an account should be defined here
}

// AccountKeeper defines the expected interface for the Account module.
type AccountKeeper interface {
	NewAccount(context.Context, sdk.AccountI) sdk.AccountI

	// GetAccount Retrieve an account from the store
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI
	// SetAccount Save an account to the store
	SetAccount(context.Context, sdk.AccountI)

	GetModuleAddress(string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
	// Methods imported from an account should be defined here
}

// BankKeeper defines the expected interface for the Bank module.
type BankKeeper interface {
	SpendableCoins(context.Context, sdk.AccAddress) sdk.Coins
	SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error
	SendCoinsFromModuleToAccount(context.Context, string, sdk.AccAddress, sdk.Coins) error
	MintCoins(context.Context, string, sdk.Coins) error
	BurnCoins(context.Context, string, sdk.Coins) error
	HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	GetAllBalances(context.Context, sdk.AccAddress) sdk.Coins
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
}

// ParamSubspace defines the expected Subspace interface for parameters.
type ParamSubspace interface {
	Get(context.Context, []byte, interface{})
	Set(context.Context, []byte, interface{})
}
