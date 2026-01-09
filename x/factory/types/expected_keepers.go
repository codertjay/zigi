package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// AccountKeeper defines the expected interface for the Account module.
type AccountKeeper interface {
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI // only used for simulation
	GetModuleAddress(name string) sdk.AccAddress
	SetModuleAccount(context.Context, sdk.ModuleAccountI)
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
	// Methods imported from an account should be defined here
}

// BankKeeper defines the expected interface for the Bank module.
type BankKeeper interface {
	SendCoins(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx context.Context, name string, amt sdk.Coins) error
	SpendableCoins(context.Context, sdk.AccAddress) sdk.Coins

	HasBalance(ctx context.Context, addr sdk.AccAddress, amt sdk.Coin) bool
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	HasSupply(ctx context.Context, denom string) bool
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	GetSupply(ctx context.Context, denom string) sdk.Coin

	// Meta data
	GetDenomMetaData(ctx context.Context, denom string) (banktypes.Metadata, bool)
	HasDenomMetaData(ctx context.Context, denom string) bool
	SetDenomMetaData(ctx context.Context, denomMetaData banktypes.Metadata)
	IterateAllDenomMetaData(ctx context.Context, cb func(banktypes.Metadata) bool)
	//GetModuleAddress(moduleName string) sdk.AccAddress
	GetAllBalances(context.Context, sdk.AccAddress) sdk.Coins
}

type StakingKeeper interface {
	BondDenom(ctx context.Context) (string, error)
}

type ModuleAccount interface {
	GetName() string // name of the module; used to get the address
}

// ParamSubspace defines the expected Subspace interface for parameters.
type ParamSubspace interface {
	Get(context.Context, []byte, interface{})
	Set(context.Context, []byte, interface{})
}
