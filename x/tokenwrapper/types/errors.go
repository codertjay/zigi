package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/tokenwrapper module sentinel errors
var (
	ErrInvalidSigner                       = sdkerrors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrSample                              = sdkerrors.Register(ModuleName, 1101, "sample error")
	ErrInvalidPacketTimeout                = sdkerrors.Register(ModuleName, 1500, "invalid packet timeout")
	ErrInvalidVersion                      = sdkerrors.Register(ModuleName, 1501, "invalid version")
	ErrModuleDisabled                      = sdkerrors.Register(ModuleName, 1502, "module functionality is not enabled")
	ErrChannelNotFound                     = sdkerrors.Register(ModuleName, 1503, "channel not found")
	ErrChannelNotOpen                      = sdkerrors.Register(ModuleName, 1504, "channel is not open")
	ErrIBCSettingsNotSet                   = sdkerrors.Register(ModuleName, 1505, "ibc settings are not set")
	ErrIBCSettingsMismatch                 = sdkerrors.Register(ModuleName, 1506, "ibc settings do not match the expected values")
	ErrNoIBCVouchersAvailableInAddress     = sdkerrors.Register(ModuleName, 1507, "no ibc vouchers available in address")
	ErrRecoveryNotAllowedOnOperatorAddress = sdkerrors.Register(ModuleName, 1508, "recovery not allowed on operator address")
)
