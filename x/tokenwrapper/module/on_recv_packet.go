package tokenwrapper

import (
	"fmt"

	types "zigchain/x/tokenwrapper/types"
	"zigchain/zutils/constants"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"
)

// OnRecvPacket is called by the routing module when a packet addressed to this module has been received.
//
// Note:
// Assuming the native port and channels are transfer/channel-0 and the counterparty chain port and channels are transfer/channel-500
// then OnRecvPacket sourcePort and sourceChannel are transfer/channel-500 and destPort and destChannel are transfer/channel-0
func (im IBCModule) OnRecvPacket(
	ctx sdk.Context,
	channelVersion string,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	im.keeper.Logger().Info(fmt.Sprintf("OnRecvPacket (tokenwrapper): Sequence: %d, Source: %s, %s; Destination: %s, %s",
		packet.Sequence, packet.SourcePort, packet.SourceChannel, packet.DestinationPort, packet.DestinationChannel))

	// Validate channel is open and exists
	if err := im.validateChannel(ctx, packet.GetDestPort(), packet.GetDestChannel()); err != nil {
		errMsg := fmt.Errorf("channel validation failed: %v", err)
		types.EmitTokenWrapperErrorEvent(ctx, errMsg)
		im.keeper.Logger().Error(errMsg.Error())
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// Parse packet data
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		err := fmt.Errorf("failed to unmarshal packet data: %v", err)
		types.EmitTokenWrapperErrorEvent(ctx, err)
		im.keeper.Logger().Error(err.Error())
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// Get sender address
	sender := data.Sender
	if sender == "" {
		err := fmt.Errorf("sender address is empty")
		types.EmitTokenWrapperErrorEvent(ctx, err)
		im.keeper.Logger().Error(fmt.Sprintf("sender address is empty: %v", err))
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// Get receiver address
	receiver, err := sdk.AccAddressFromBech32(data.Receiver)
	if err != nil {
		types.EmitTokenWrapperErrorEvent(ctx, err)
		im.keeper.Logger().Error(fmt.Sprintf("invalid receiver address: %v", err))
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// Parse amount
	amount, ok := sdkmath.NewIntFromString(data.Amount)
	if !ok {
		err := fmt.Errorf("invalid amount: %s", data.Amount)
		types.EmitTokenWrapperErrorEvent(ctx, err)
		im.keeper.Logger().Error(fmt.Sprintf("invalid amount: %v", err))
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// Check if amount is positive otherwise return an error
	if amount.IsZero() || amount.IsNegative() {
		err := fmt.Errorf("amount is zero or negative: %s", data.Amount)
		types.EmitTokenWrapperErrorEvent(ctx, err)
		im.keeper.Logger().Error(fmt.Sprintf("amount is zero or negative: %v", err))
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// Process packet through middleware stack
	ack := im.app.OnRecvPacket(ctx, channelVersion, packet, relayer)
	if !ack.Success() {
		err := fmt.Errorf("packet processing failed")
		types.EmitTokenWrapperErrorEvent(ctx, err)
		im.keeper.Logger().Error(err.Error())
		return ack
	}

	// Note: by the time we reach here given no errors occured, the IBC vouchers are received in the receiver address

	// Check if the counterparty channel matches the expected IBC settings, if not, skip wrapping
	if err := im.checkCounterypartyChannelMatchesIBCSettings(ctx, packet.GetDestPort(), packet.GetDestChannel()); err != nil {
		err := fmt.Errorf("counterpartychannel matches IBC settings failed: %v", err)
		types.EmitTokenWrapperErrorEvent(ctx, err)
		im.keeper.Logger().Error(err.Error())
		return ack
	}

	// Check if native and counterparty client ids are valid, if not, skip wrapping
	if err := im.validateConnectionClientId(ctx, packet.GetDestPort(), packet.GetDestChannel()); err != nil {
		err := fmt.Errorf("client state validation failed: %v", err)
		types.EmitTokenWrapperErrorEvent(ctx, err)
		im.keeper.Logger().Error(err.Error())
		return ack
	}

	// Validate IBC settings exist, if not, skip wrapping
	if !im.validateIBCSettingsExist(ctx) {
		info := fmt.Sprintf("IBC settings validation failed: %v, skipping wrapping", types.ErrIBCSettingsNotSet)
		types.EmitTokenWrapperInfoEvent(ctx, info)
		im.keeper.Logger().Info(info)
		return ack
	}

	// Skip wrapping if packet denom is not the module denom
	if !im.validateIBCDenomIsModuleDenom(ctx, data.Denom) {
		info := fmt.Sprintf("packet denom is not the module denom, skipping wrapping: %s", data.Denom)
		types.EmitTokenWrapperInfoEvent(ctx, info)
		im.keeper.Logger().Info(info)
		return ack
	}

	// If the sender chain is not the source chain, no wrapping is needed (denom has the source port and source channel as a prefix)
	if !im.validateReceiverChainIsNotSourceChain(data.Denom, packet) {
		info := "sender chain is not the source chain, skipping wrapping"
		types.EmitTokenWrapperInfoEvent(ctx, info)
		im.keeper.Logger().Info(info)
		return ack
	}

	// Check if module is enabled
	if !im.keeper.IsEnabled(ctx) {
		err := fmt.Errorf("module disabled: %v", types.ErrModuleDisabled)
		types.EmitTokenWrapperErrorEvent(ctx, err)
		im.keeper.Logger().Error(err.Error())
		return ack
	}

	// Check if IBC settings match the expected values
	if !im.validateIBCSettingsMatchOnRecvPacket(ctx, packet) {
		err := fmt.Errorf("%v, failed with packet: %v", types.ErrIBCSettingsMismatch, packet)
		types.EmitTokenWrapperErrorEvent(ctx, err)
		im.keeper.Logger().Error(err.Error())
		return ack
	}

	// Get IBC denom for the received token
	recvDenom := im.keeper.GetIBCRecvDenom(ctx, data.Denom)

	// Handle token conversion
	convertedAmount, err := im.keeper.ScaleDownTokenPrecision(ctx, amount)
	if err != nil {
		types.EmitTokenWrapperErrorEvent(ctx, err)
		im.keeper.Logger().Error(fmt.Sprintf("token conversion failed: %v", err))
		return ack
	}

	// Check balances
	if err := im.keeper.CheckBalances(ctx, receiver, amount, recvDenom, convertedAmount); err != nil {
		types.EmitTokenWrapperErrorEvent(ctx, err)
		im.keeper.Logger().Error(fmt.Sprintf("balances check failed: %v", err))
		// Note: return a successful acknowlegment in order to keep the IBC vouchers in the receiver address
		// the receiver address will be able to convert the IBC vouchers with native ZIG through the recovery process
		return ack
	}

	// Lock IBC tokens
	ibcCoins, err := im.keeper.LockIBCTokens(ctx, receiver, amount, recvDenom)
	if err != nil {
		types.EmitTokenWrapperErrorEvent(ctx, err)
		im.keeper.Logger().Error(fmt.Sprintf("failed to lock IBC tokens: %v", err))
		return ack
	}

	// Unlock native tokens
	_, err = im.keeper.UnlockNativeTokens(ctx, receiver, convertedAmount, ibcCoins)
	if err != nil {
		types.EmitTokenWrapperErrorEvent(ctx, err)
		im.keeper.Logger().Error(fmt.Sprintf("failed to unlock native tokens: %v", err))
		return ack
	}

	// Track transferred amount
	im.keeper.AddToTotalTransferredIn(ctx, convertedAmount)

	// Emit success event
	types.EmitTokenWrapperPacketEvent(
		ctx,
		types.EventTypeTokenWrapperOnRecvPacket,
		sender,
		receiver.String(),
		convertedAmount,
		constants.BondDenom,
		packet.GetSourcePort(),
		packet.GetSourceChannel(),
		packet.GetDestPort(),
		packet.GetDestChannel(),
	)

	return ack
}
