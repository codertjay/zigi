package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IBC events
const (
	EventTypeTimeout = "timeout"
	// this line is used by starport scaffolding # ibc/packet/event

	AttributeKeyAckSuccess = "success"
	AttributeKeyAck        = "acknowledgement"
	AttributeKeyAckError   = "error"

	// Tokenwrapper events
	EventTypeTokenWrapperOnRecvPacket = "tokenwrapper_on_recv_packet"
	EventTypeTokenWrapperSendPacket   = "tokenwrapper_send_packet"
	EventTypeTokenWrapperError        = "tokenwrapper_error"
	EventTypeTokenWrapperInfo         = "tokenwrapper_info"
	EventTypeTokenWrapperRefund       = "tokenwrapper_refund"
	EventTypePauserAddressAdded       = "pauser_address_added"
	EventTypePauserAddressRemoved     = "pauser_address_removed"
	EventTypeOperatorAddressProposed  = "operator_address_proposed"
	EventTypeOperatorAddressClaimed   = "operator_address_claimed"
	EventTypeIbcSettingsUpdated       = "ibc_settings_updated"
	EventTypeTokenWrapperEnabled      = "tokenwrapper_enabled"
	EventTypeTokenWrapperDisabled     = "tokenwrapper_disabled"
	EventTypeModuleWalletFunded       = "module_wallet_funded"
	EventTypeModuleWalletWithdrawn    = "module_wallet_withdrawn"
	EventTypeParamsUpdated            = "params_updated"
	EventTypeAddressZigRecovered      = "address_zig_recovered"

	// Tokenwrapper attributes
	AttributeKeySender               = "sender"
	AttributeKeyReceiver             = "receiver"
	AttributeKeyAmount               = "amount"
	AttributeKeyDenom                = "denom"
	AttributeKeySourcePort           = "source_port"
	AttributeKeySourceChannel        = "source_channel"
	AttributeKeyDestPort             = "dest_port"
	AttributeKeyDestChannel          = "dest_channel"
	AttributeKeyError                = "error"
	AttributeKeyInfo                 = "info"
	AttributeKeyPauserAddress        = "pauser_address"
	AttributeKeyOldOperator          = "old_operator"
	AttributeKeyNewOperator          = "new_operator"
	AttributeKeySigner               = "signer"
	AttributeKeyModuleAddress        = "module_address"
	AttributeKeyBalances             = "balances"
	AttributeKeyNativeClientId       = "native_client_id"
	AttributeKeyCounterpartyClientId = "counterparty_client_id"
	AttributeKeyNativePort           = "native_port"
	AttributeKeyCounterpartyPort     = "counterparty_port"
	AttributeKeyNativeChannel        = "native_channel"
	AttributeKeyCounterpartyChannel  = "counterparty_channel"
	AttributeKeyDecimalDifference    = "decimal_difference"
	AttributeKeyAuthority            = "authority"
	AttributeKeyRefundAmount         = "refund_amount"
	AttributeKeyRefundDenom          = "refund_denom"
	AttributeKeyConvertedAmount      = "converted_amount"
	AttributeKeyConvertedDenom       = "converted_denom"
	AttributeKeyAddress              = "address"
	AttributeKeyLockedIbcAmount      = "locked_ibc_amount"
	AttributeKeyUnlockedNativeAmount = "unlocked_native_amount"
)

// EmitTokenWrapperErrorEvent emits an error event for token wrapper operations
func EmitTokenWrapperErrorEvent(ctx sdk.Context, err error) {
	// skip if event manager not available
	if ctx.EventManager() == nil {
		return
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypeTokenWrapperError,
			sdk.NewAttribute(AttributeKeyError, err.Error()),
		),
	)
}

// EmitTokenWrapperInfoEvent emits an info event for token wrapper operations
func EmitTokenWrapperInfoEvent(ctx sdk.Context, info string) {
	// skip if event manager not available
	if ctx.EventManager() == nil {
		return
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypeTokenWrapperInfo,
			sdk.NewAttribute(AttributeKeyInfo, info),
		),
	)
}

// Allowed event packet types
var (
	AllowedPacketTypes = map[string]struct{}{
		EventTypeTokenWrapperOnRecvPacket: {},
		EventTypeTokenWrapperSendPacket:   {},
	}
)

func IsValidPacketType(packetType string) bool {
	_, ok := AllowedPacketTypes[packetType]
	return ok
}

// EmitTokenWrapperPacketEvent emits a success event for token wrapper operations
func EmitTokenWrapperPacketEvent(
	ctx sdk.Context,
	packetType string,
	sender string,
	receiver string,
	amount sdkmath.Int,
	denom string,
	sourcePort string,
	sourceChannel string,
	destPort string,
	destChannel string,
) {
	// skip if event manager not available
	if ctx.EventManager() == nil {
		return
	}

	if !IsValidPacketType(packetType) {
		return
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			packetType,
			sdk.NewAttribute(AttributeKeySender, sender),
			sdk.NewAttribute(AttributeKeyReceiver, receiver),
			sdk.NewAttribute(AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(AttributeKeyDenom, denom),
			sdk.NewAttribute(AttributeKeySourcePort, sourcePort),
			sdk.NewAttribute(AttributeKeySourceChannel, sourceChannel),
			sdk.NewAttribute(AttributeKeyDestPort, destPort),
			sdk.NewAttribute(AttributeKeyDestChannel, destChannel),
		),
	)
}

// EmitTokenWrapperRefundEvent emits a refund event for token wrapper operations
func EmitTokenWrapperRefundEvent(
	ctx sdk.Context,
	sender string,
	refundAmount sdkmath.Int,
	refundDenom string,
	convertedAmount sdkmath.Int,
	convertedDenom string,
) {
	// skip if event manager not available
	if ctx.EventManager() == nil {
		return
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypeTokenWrapperRefund,
			sdk.NewAttribute(AttributeKeySender, sender),
			sdk.NewAttribute(AttributeKeyRefundAmount, refundAmount.String()),
			sdk.NewAttribute(AttributeKeyRefundDenom, refundDenom),
			sdk.NewAttribute(AttributeKeyConvertedAmount, convertedAmount.String()),
			sdk.NewAttribute(AttributeKeyConvertedDenom, convertedDenom),
		),
	)
}

// EmitPauserAddressAddedEvent emits an event when a pauser address is added
func EmitPauserAddressAddedEvent(ctx sdk.Context, pauserAddress string) {
	// skip if event manager not available
	if ctx.EventManager() == nil {
		return
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypePauserAddressAdded,
			sdk.NewAttribute(AttributeKeyPauserAddress, pauserAddress),
		),
	)
}

// EmitPauserAddressRemovedEvent emits an event when a pauser address is removed
func EmitPauserAddressRemovedEvent(ctx sdk.Context, pauserAddress string) {
	// skip if event manager not available
	if ctx.EventManager() == nil {
		return
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypePauserAddressRemoved,
			sdk.NewAttribute(AttributeKeyPauserAddress, pauserAddress),
		),
	)
}

// EmitOperatorAddressProposedEvent emits an event when the operator address is proposed
func EmitOperatorAddressProposedEvent(ctx sdk.Context, oldOperator, newOperator string) {
	if ctx.EventManager() == nil {
		return
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypeOperatorAddressProposed,
			sdk.NewAttribute(AttributeKeyOldOperator, oldOperator),
			sdk.NewAttribute(AttributeKeyNewOperator, newOperator),
		),
	)
}

// EmitOperatorAddressClaimedEvent emits an event when the operator address is claimed
func EmitOperatorAddressClaimedEvent(ctx sdk.Context, oldOperator, newOperator string) {
	if ctx.EventManager() == nil {
		return
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypeOperatorAddressClaimed,
			sdk.NewAttribute(AttributeKeyOldOperator, oldOperator),
			sdk.NewAttribute(AttributeKeyNewOperator, newOperator),
		),
	)
}

// EmitIbcSettingsUpdatedEvent emits an event when IBC settings are updated
func EmitIbcSettingsUpdatedEvent(ctx sdk.Context, signer string, msg interface{}) {
	if ctx.EventManager() == nil {
		return
	}
	settings, ok := msg.(map[string]string)
	if !ok {
		return
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypeIbcSettingsUpdated,
			sdk.NewAttribute(AttributeKeySigner, signer),
			sdk.NewAttribute(AttributeKeyNativeClientId, settings[AttributeKeyNativeClientId]),
			sdk.NewAttribute(AttributeKeyCounterpartyClientId, settings[AttributeKeyCounterpartyClientId]),
			sdk.NewAttribute(AttributeKeyNativePort, settings[AttributeKeyNativePort]),
			sdk.NewAttribute(AttributeKeyCounterpartyPort, settings[AttributeKeyCounterpartyPort]),
			sdk.NewAttribute(AttributeKeyNativeChannel, settings[AttributeKeyNativeChannel]),
			sdk.NewAttribute(AttributeKeyCounterpartyChannel, settings[AttributeKeyCounterpartyChannel]),
			sdk.NewAttribute(AttributeKeyDenom, settings[AttributeKeyDenom]),
			sdk.NewAttribute(AttributeKeyDecimalDifference, settings[AttributeKeyDecimalDifference]),
		),
	)
}

// EmitTokenWrapperEnabledEvent emits an event when the token wrapper is enabled
func EmitTokenWrapperEnabledEvent(ctx sdk.Context, signer string) {
	if ctx.EventManager() == nil {
		return
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypeTokenWrapperEnabled,
			sdk.NewAttribute(AttributeKeySigner, signer),
		),
	)
}

// EmitTokenWrapperDisabledEvent emits an event when the token wrapper is disabled
func EmitTokenWrapperDisabledEvent(ctx sdk.Context, signer string) {
	if ctx.EventManager() == nil {
		return
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypeTokenWrapperDisabled,
			sdk.NewAttribute(AttributeKeySigner, signer),
		),
	)
}

// EmitModuleWalletFundedEvent emits an event when the module wallet is funded
func EmitModuleWalletFundedEvent(ctx sdk.Context, signer, moduleAddress string, amount sdk.Coins, balances sdk.Coins) {
	if ctx.EventManager() == nil {
		return
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypeModuleWalletFunded,
			sdk.NewAttribute(AttributeKeySigner, signer),
			sdk.NewAttribute(AttributeKeyModuleAddress, moduleAddress),
			sdk.NewAttribute(AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(AttributeKeyBalances, balances.String()),
		),
	)
}

// EmitModuleWalletWithdrawnEvent emits an event when the module wallet is withdrawn from
func EmitModuleWalletWithdrawnEvent(ctx sdk.Context, signer, moduleAddress string, amount sdk.Coins, balances sdk.Coins) {
	if ctx.EventManager() == nil {
		return
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypeModuleWalletWithdrawn,
			sdk.NewAttribute(AttributeKeySigner, signer),
			sdk.NewAttribute(AttributeKeyModuleAddress, moduleAddress),
			sdk.NewAttribute(AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(AttributeKeyBalances, balances.String()),
		),
	)
}

// EmitParamsUpdatedEvent emits an event when module parameters are updated
func EmitParamsUpdatedEvent(ctx sdk.Context, authority string) {
	if ctx.EventManager() == nil {
		return
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypeParamsUpdated,
			sdk.NewAttribute(AttributeKeyAuthority, authority),
		),
	)
}

// EmitAddressZigRecovered emits an event when an address’s zig tokens are recovered
func EmitAddressZigRecovered(ctx sdk.Context, signer, address sdk.AccAddress, lockedIbcAmount, unlockedNativeAmount sdk.Coin) {
	// skip if event manager not available
	if ctx.EventManager() == nil {
		return
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypeAddressZigRecovered,
			sdk.NewAttribute(AttributeKeySigner, signer.String()),
			sdk.NewAttribute(AttributeKeyAddress, address.String()),
			sdk.NewAttribute(AttributeKeyLockedIbcAmount, lockedIbcAmount.String()),
			sdk.NewAttribute(AttributeKeyUnlockedNativeAmount, unlockedNativeAmount.String()),
		),
	)
}
