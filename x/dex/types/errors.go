package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/dex module sentinel errors
var (
	ErrInvalidSigner = sdkerrors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")

	ErrPoolNotFound              = sdkerrors.Register(ModuleName, 1502, "pool not found")
	ErrInsufficientBalance       = sdkerrors.Register(ModuleName, 1503, "insufficient balance")
	ErrorInvalidAmount           = sdkerrors.Register(ModuleName, 1504, "invalid amount")
	ErrorInvalidReceiver         = sdkerrors.Register(ModuleName, 1505, "invalid receiver")
	ErrPoolAccountNotFound       = sdkerrors.Register(ModuleName, 1506, "pool account not found")
	ErrNonPositiveAmounts        = sdkerrors.Register(ModuleName, 1507, "on initial pool creation amounts must be positive")
	ErrInsufficientLiquidityLock = sdkerrors.Register(ModuleName, 1508, "on initial pool creation lpAmount must be greater than minimal lock")
	ErrInvalidPoolAddress        = sdkerrors.Register(ModuleName, 1509, "pool address does not match expected module address")
)
