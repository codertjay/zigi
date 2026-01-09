package tokenwrapper

import (
	"fmt"
	"zigchain/x/tokenwrapper/types"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
)

// OnTimeoutPacket is called by the routing module when a packet sent by this module has timed-out (such that it will not be received on the destination chain).
func (im IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	channelVersion string,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	im.keeper.Logger().Error(fmt.Sprintf("OnTimeoutPacket (tokenwrapper): %v", packet))

	// Validate channel is open and exists
	if err := im.validateChannel(ctx, packet.GetSourcePort(), packet.GetSourceChannel()); err != nil {
		errMsg := fmt.Errorf("channel validation failed: %v", err)
		types.EmitTokenWrapperErrorEvent(ctx, errMsg)
		im.keeper.Logger().Error(errMsg.Error())
		return im.app.OnTimeoutPacket(ctx, channelVersion, packet, relayer)
	}

	// Parse the packet data to get the amount and sender
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		types.EmitTokenWrapperErrorEvent(ctx, err)
		im.keeper.Logger().Error(fmt.Sprintf("failed to unmarshal packet data: %v", err))
		return im.app.OnTimeoutPacket(ctx, channelVersion, packet, relayer)
	}

	// Get the sender address
	sender, err := sdk.AccAddressFromBech32(data.Sender)
	if err != nil {
		types.EmitTokenWrapperErrorEvent(ctx, err)
		im.keeper.Logger().Error(fmt.Sprintf("invalid sender address: %v", err))
		return im.app.OnTimeoutPacket(ctx, channelVersion, packet, relayer)
	}

	// Parse amount as sdkmath.Int
	amount, ok := sdkmath.NewIntFromString(data.Amount)
	if !ok {
		err := fmt.Errorf("invalid amount: %s", data.Amount)
		types.EmitTokenWrapperErrorEvent(ctx, err)
		im.keeper.Logger().Error(fmt.Sprintf("invalid amount: %v", err))
		return im.app.OnTimeoutPacket(ctx, channelVersion, packet, relayer)
	}

	// Check if amount is zero or negative
	if amount.IsZero() || amount.IsNegative() {
		err := fmt.Errorf("amount is zero or negative: %s", data.Amount)
		types.EmitTokenWrapperErrorEvent(ctx, err)
		im.keeper.Logger().Error(fmt.Sprintf("amount is zero or negative: %v", err))
		return im.app.OnTimeoutPacket(ctx, channelVersion, packet, relayer)
	}

	// Validate IBC settings exist, if not, skip refunding
	if !im.validateIBCSettingsExist(ctx) {
		info := fmt.Sprintf("IBC settings validation failed: %v, skipping refunding", types.ErrIBCSettingsNotSet)
		types.EmitTokenWrapperInfoEvent(ctx, info)
		im.keeper.Logger().Info(info)
		return im.app.OnTimeoutPacket(ctx, channelVersion, packet, relayer)
	}

	// Get base denom from the packet data
	baseDenom := transfertypes.ExtractDenomFromPath(data.Denom).Base

	// Skip refunding if packet denom is not the module denom
	if !im.validateIBCDenomIsModuleDenom(ctx, baseDenom) {
		info := fmt.Sprintf("packet denom is not the module denom, skipping refunding: %s", baseDenom)
		types.EmitTokenWrapperInfoEvent(ctx, info)
		im.keeper.Logger().Info(info)
		return im.app.OnTimeoutPacket(ctx, channelVersion, packet, relayer)
	}

	// Check if IBC settings match the expected values
	if !im.validateIBCSettingsMatchOnSendPacket(ctx, packet) {
		err := fmt.Errorf("%v, failed with packet: %v", types.ErrIBCSettingsMismatch, packet)
		types.EmitTokenWrapperErrorEvent(ctx, err)
		im.keeper.Logger().Error(err.Error())
		return im.app.OnTimeoutPacket(ctx, channelVersion, packet, relayer)
	}

	// Check if the module functionality is enabled
	if !im.keeper.IsEnabled(ctx) {
		err := fmt.Errorf("module functionality is not enabled")
		types.EmitTokenWrapperErrorEvent(ctx, err)
		im.keeper.Logger().Error(fmt.Sprintf("module functionality is not enabled: %v", err))
		return im.app.OnTimeoutPacket(ctx, channelVersion, packet, relayer)
	}

	// First call the underlying OnTimeoutPacket
	if err := im.app.OnTimeoutPacket(ctx, channelVersion, packet, relayer); err != nil {
		types.EmitTokenWrapperErrorEvent(ctx, err)
		im.keeper.Logger().Error(fmt.Sprintf("failed to call underlying OnTimeoutPacket: %v", err))
		return err
	}

	// Handle refund for timeout
	if err := im.handleRefund(ctx, sender, amount, data.Denom); err != nil {
		types.EmitTokenWrapperErrorEvent(ctx, err)
		im.keeper.Logger().Error(fmt.Sprintf("failed to handle refund: %v", err))
		// return nil to apply the default behavior of the underlying OnTimeoutPacket
		return nil
	}

	// Emit error event for timeout
	types.EmitTokenWrapperErrorEvent(ctx, fmt.Errorf("IBC packet timeout: %s/%s sequence %d", packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence()))

	return nil
}
