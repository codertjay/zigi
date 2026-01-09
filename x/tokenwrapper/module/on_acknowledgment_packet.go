package tokenwrapper

import (
	"fmt"

	types "zigchain/x/tokenwrapper/types"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
)

// OnAcknowledgePacket is called by the routing module when a packet sent by this module has been acknowledged.
func (im IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	channelVersion string,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	im.keeper.Logger().Info(fmt.Sprintf("OnAcknowledgementPacket (tokenwrapper): %v", packet))

	// Validate channel is open and exists
	if err := im.validateChannel(ctx, packet.GetSourcePort(), packet.GetSourceChannel()); err != nil {
		errMsg := fmt.Errorf("channel validation failed: %v", err)
		types.EmitTokenWrapperErrorEvent(ctx, errMsg)
		im.keeper.Logger().Error(errMsg.Error())
		return err
	}

	// Parse the packet data to get the amount
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		types.EmitTokenWrapperErrorEvent(ctx, err)
		im.keeper.Logger().Error(fmt.Sprintf("failed to unmarshal packet data: %v", err))
		return err
	}

	// Validate IBC settings exist, if not, skip wrapping
	if !im.validateIBCSettingsExist(ctx) {
		info := fmt.Sprintf("IBC settings validation failed: %v, skipping wrapping", types.ErrIBCSettingsNotSet)
		types.EmitTokenWrapperInfoEvent(ctx, info)
		im.keeper.Logger().Info(info)
		return im.app.OnAcknowledgementPacket(ctx, channelVersion, packet, acknowledgement, relayer)
	}

	// If the denom, source port and source channel are not the module IBC settings, no tracking is needed
	// Note:
	// Assuming the native port and channels are transfer/channel-0 and the counterparty chain port and channels are transfer/channel-500
	// then OnAcknowledgementPacket sourcePort and sourceChannel are transfer/channel-0 and destPort and destChannel are transfer/channel-500

	// Get base denom from the packet data
	baseDenom := transfertypes.ExtractDenomFromPath(data.Denom).Base

	// Skip tracking if packet denom is not the module denom
	if !im.validateIBCDenomIsModuleDenom(ctx, baseDenom) {
		info := fmt.Sprintf("packet denom is not the module denom, skipping tracking: %s", baseDenom)
		types.EmitTokenWrapperInfoEvent(ctx, info)
		im.keeper.Logger().Info(info)
		return im.app.OnAcknowledgementPacket(ctx, channelVersion, packet, acknowledgement, relayer)
	}

	// Packet denom is the module denom, we need to track the transfer

	// Check if IBC settings match the expected values
	if !im.validateIBCSettingsMatchOnSendPacket(ctx, packet) {
		err := fmt.Errorf("%v, failed with packet: %v", types.ErrIBCSettingsMismatch, packet)
		types.EmitTokenWrapperErrorEvent(ctx, err)
		im.keeper.Logger().Error(err.Error())
		// we should return an error to prevent misuse of unit-zig, "unit-zig" IBC vouchers can only be recovered
		return err
	}

	// Check if the module functionality is enabled only if IBC settings match
	if !im.keeper.IsEnabled(ctx) {
		err := fmt.Errorf("module functionality is not enabled")
		types.EmitTokenWrapperErrorEvent(ctx, err)
		im.keeper.Logger().Error(fmt.Sprintf("module functionality is not enabled: %v", err))
		return err
	}

	// Parse amount as sdkmath.Int
	amount, ok := sdkmath.NewIntFromString(data.Amount)
	if !ok {
		err := fmt.Errorf("invalid amount: %s", data.Amount)
		types.EmitTokenWrapperErrorEvent(ctx, err)
		im.keeper.Logger().Error(fmt.Sprintf("invalid amount: %v", err))
		return err
	}

	// Get the sender address
	sender, err := sdk.AccAddressFromBech32(data.Sender)
	if err != nil {
		types.EmitTokenWrapperErrorEvent(ctx, err)
		im.keeper.Logger().Error(fmt.Sprintf("invalid sender address: %v", err))
		return err
	}

	// First call the underlying OnAcknowledgementPacket
	if err := im.app.OnAcknowledgementPacket(ctx, channelVersion, packet, acknowledgement, relayer); err != nil {
		types.EmitTokenWrapperErrorEvent(ctx, err)
		im.keeper.Logger().Error(fmt.Sprintf("failed to call underlying OnAcknowledgementPacket: %v", err))
		return err
	}

	var ack channeltypes.Acknowledgement
	if err := transfertypes.ModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		types.EmitTokenWrapperErrorEvent(ctx, err)
		im.keeper.Logger().Error(fmt.Sprintf("failed to unmarshal acknowledgement: %v", err))
		return err
	}

	// log the acknowledgement status
	im.keeper.Logger().Info(fmt.Sprintf("acknowledgement status: %v", ack.Success()))

	// If the acknowledgement does not indicate success, handle the refund
	if !ack.Success() {
		im.keeper.Logger().Info("acknowledgement failed, handling refund")
		// Handle refund for failed acknowledgement
		if err := im.handleRefund(ctx, sender, amount, data.Denom); err != nil {
			types.EmitTokenWrapperErrorEvent(ctx, err)
			im.keeper.Logger().Error(fmt.Sprintf("failed to handle refund: %v", err))
			// return nil to apply the default behavior of the underlying OnAcknowledgementPacket
			return nil
		}

		// Return nil to indicate that the refund was successful, no need to track the transfer
		return nil
	}

	// Note 1: we need to revert the conversion done in SendPacket as we use 6 decimals for the stats recording
	// Convert from 18 decimals to 6 decimals
	// Note 2: we do not handle refund here because handleRefund will fail as it depends on scaling down the token precision,
	// and this logic must only be done when the acknowledgement is not successful.
	convertedAmount, err := im.keeper.ScaleDownTokenPrecision(ctx, amount)
	if err != nil {
		types.EmitTokenWrapperErrorEvent(ctx, err)
		im.keeper.Logger().Error(fmt.Sprintf("converted amount is zero or negative: %v", err))
		return err
	}

	// Track the amount of ZIG tokens transferred out (using the original amount with 6 decimals)
	im.keeper.AddToTotalTransferredOut(ctx, convertedAmount)

	return nil
}
