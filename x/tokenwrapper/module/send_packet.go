package tokenwrapper

import (
	"fmt"

	types "zigchain/x/tokenwrapper/types"
	"zigchain/zutils/constants"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
)

// SendPacket implements the ICS4Wrapper interface
//
// Note:
// Assuming the native port and channels are transfer/channel-0 and the counterparty chain port and channels are transfer/channel-500
// then SendPacket sourcePort and sourceChannel are transfer/channel-0 and destPort and destChannel are transfer/channel-500
func (w *ICS4Wrapper) SendPacket(
	ctx sdk.Context,
	sourcePort string,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	packetData []byte,
) (uint64, error) {
	w.keeper.Logger().Info(fmt.Sprintf("SendPacket (tokenwrapper): Source: %s, %s", sourcePort, sourceChannel))

	// Validate channel is open and exists
	if err := w.validateChannel(ctx, sourcePort, sourceChannel); err != nil {
		errMsg := fmt.Errorf("channel validation failed: %v", err)
		types.EmitTokenWrapperErrorEvent(ctx, errMsg)
		w.keeper.Logger().Error(errMsg.Error())
		return 0, err
	}

	// Parse the packet data
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packetData, &data); err != nil {
		types.EmitTokenWrapperErrorEvent(ctx, err)
		w.keeper.Logger().Error(fmt.Sprintf("failed to unmarshal packet data: %v", err))
		return 0, err
	}

	// Get module params
	moduleNativePort := w.keeper.GetNativePort(ctx)
	moduleNativeChannel := w.keeper.GetNativeChannel(ctx)
	moduleDenom := w.keeper.GetDenom(ctx)

	// If the denom is the constants.BondDenom, then override the denom to the module denom
	if data.Denom == constants.BondDenom {
		data.Denom = moduleDenom
	}

	// If the module functionality is disabled and the denom, source port and source channel are the ibc settings then stop the packet from being processed
	if !w.keeper.IsEnabled(ctx) &&
		data.Denom == moduleDenom &&
		sourcePort == moduleNativePort &&
		sourceChannel == moduleNativeChannel {

		types.EmitTokenWrapperErrorEvent(ctx, fmt.Errorf("module functionality is not enabled"))
		w.keeper.Logger().Error("module functionality is not enabled")
		return 0, types.ErrModuleDisabled
	}

	// If the sender chain is not the source chain, no wrapping is needed
	if transfertypes.ExtractDenomFromPath(data.Denom).HasPrefix(sourcePort, sourceChannel) {
		types.EmitTokenWrapperInfoEvent(ctx, "sender chain is not the source chain")
		w.keeper.Logger().Info("sender chain is not the source chain")
		return w.ics4Wrapper.SendPacket(ctx, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, packetData)
	}

	// If any of the module params are not set, no wrapping is needed
	if moduleNativePort == "" || moduleNativeChannel == "" || moduleDenom == "" {
		types.EmitTokenWrapperInfoEvent(ctx, "module params are not set")
		w.keeper.Logger().Info("module params are not set")
		return w.ics4Wrapper.SendPacket(ctx, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, packetData)
	}

	// If the denom, native port and native channel don't match the module IBC settings, no wrapping is needed
	if data.Denom != moduleDenom || sourcePort != moduleNativePort || sourceChannel != moduleNativeChannel {
		types.EmitTokenWrapperInfoEvent(ctx, "denom, native port and native channel don't match the module IBC settings, no wrapping is needed")
		w.keeper.Logger().Info("denom, native port and native channel don't match the module IBC settings, no wrapping is needed")
		return w.ics4Wrapper.SendPacket(ctx, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, packetData)
	}

	// User wants to send native tokens to Axelar

	// Check if the counterparty channel matches the expected IBC settings, if not, return an error
	if err := w.checkCounterypartyChannelMatchesIBCSettings(ctx, sourcePort, sourceChannel); err != nil {
		err := fmt.Errorf("counterparty channel matches IBC settings failed: %v", err)
		types.EmitTokenWrapperErrorEvent(ctx, err)
		w.keeper.Logger().Error(err.Error())
		return 0, err
	}

	// Check if native and counterparty client ids are valid, if not, return an error
	if err := w.validateConnectionClientId(ctx, sourcePort, sourceChannel); err != nil {
		err := fmt.Errorf("client state validation failed: %v", err)
		types.EmitTokenWrapperErrorEvent(ctx, err)
		w.keeper.Logger().Error(err.Error())
		return 0, err
	}

	// Get the sender address
	sender, err := sdk.AccAddressFromBech32(data.Sender)
	if err != nil {
		types.EmitTokenWrapperErrorEvent(ctx, err)
		w.keeper.Logger().Error(fmt.Sprintf("invalid sender address: %v", err))
		return 0, err
	}

	// Get the receiver address
	receiver := data.Receiver
	if receiver == "" {
		err := fmt.Errorf("receiver address is empty")
		types.EmitTokenWrapperErrorEvent(ctx, err)
		w.keeper.Logger().Error(fmt.Sprintf("receiver address is empty: %v", err))
		return 0, err
	}

	// Parse amount as sdkmath.Int
	amount, ok := sdkmath.NewIntFromString(data.Amount)
	if !ok {
		err := fmt.Errorf("invalid amount: %s", data.Amount)
		types.EmitTokenWrapperErrorEvent(ctx, err)
		w.keeper.Logger().Error(fmt.Sprintf("invalid amount: %v", err))
		return 0, err
	}

	// Make sure amount is positive
	if amount.IsZero() || amount.IsNegative() {
		err := fmt.Errorf("amount is zero or negative: %s", amount.String())
		types.EmitTokenWrapperErrorEvent(ctx, err)
		w.keeper.Logger().Error(fmt.Sprintf("amount is zero or negative: %v", err))
		return 0, err
	}

	// Create native coins
	nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, amount))

	// Note: escrow is not needed since we're dealing with IBC tokens rather than native ones

	// Obtain the escrow address for the source channel end
	escrowAddress := transfertypes.GetEscrowAddress(sourcePort, sourceChannel)

	// Unescrow the tokens
	if err := w.handleUnescrowToken(ctx, escrowAddress, sender, nativeCoins[0]); err != nil {
		types.EmitTokenWrapperErrorEvent(ctx, err)
		w.keeper.Logger().Error(fmt.Sprintf("failed to unescrow token: %v", err))
		return 0, err
	}

	// Check if sender has received the unescrowed tokens
	if !w.bankKeeper.HasBalance(ctx, sender, nativeCoins[0]) {
		err := fmt.Errorf("sender does not have the unescrowed tokens: %s", nativeCoins[0].String())
		types.EmitTokenWrapperErrorEvent(ctx, err)
		w.keeper.Logger().Error(fmt.Sprintf("sender does not have the unescrowed tokens: %v", err))
		return 0, err
	}

	// Get the sender denom
	sourcePrefix := transfertypes.NewHop(moduleNativePort, moduleNativeChannel)
	prefixedDenom := fmt.Sprintf("%s/%s", sourcePrefix.String(), data.Denom)
	sendDenom := transfertypes.ExtractDenomFromPath(prefixedDenom).IBCDenom()

	// module address
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)

	// Convert from 6 decimals to 18 decimals
	// Multiply by conversion factor to convert from 6 decimals to 18 decimals
	convertedAmount, err := w.keeper.ScaleUpTokenPrecision(ctx, amount)

	// Check if the converted amount is zero or negative
	if err != nil {
		types.EmitTokenWrapperErrorEvent(ctx, err)
		w.keeper.Logger().Error(fmt.Sprintf("converted amount is zero or negative: %v", err))
		return 0, err
	}

	// Check if the sender has enough balance to lock tokens
	if err := w.keeper.CheckAccountBalance(ctx, sender, nativeCoins); err != nil {
		types.EmitTokenWrapperErrorEvent(ctx, err)
		w.keeper.Logger().Error(fmt.Sprintf("insufficient balance: %v", err))
		return 0, err
	}

	// Create ibc coins with converted amount
	ibcCoins := sdk.NewCoins(sdk.NewCoin(sendDenom, convertedAmount))

	// Check if the module has enough balance to burn tokens
	if err := w.keeper.CheckModuleBalance(ctx, ibcCoins); err != nil {
		types.EmitTokenWrapperErrorEvent(ctx, err)
		w.keeper.Logger().Error(fmt.Sprintf("insufficient balance: %v", err))
		return 0, err
	}

	// Modify the packet data by replacing the denom with an IBC denom and updating the amount
	data.Denom = prefixedDenom
	data.Amount = convertedAmount.String()
	bz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	if err != nil {
		types.EmitTokenWrapperErrorEvent(ctx, err)
		w.keeper.Logger().Error(fmt.Sprintf("failed to marshal packet data: %v", err))
		return 0, err
	}

	// Lock native tokens
	if err := w.keeper.LockTokens(ctx, sender, nativeCoins); err != nil {
		types.EmitTokenWrapperErrorEvent(ctx, err)
		w.keeper.Logger().Error(fmt.Sprintf("failed to lock tokens: %v", err))
		return 0, err
	}

	w.keeper.Logger().Info(fmt.Sprintf("Locked %s tokens for module address: %s, sender: %s", nativeCoins.String(), moduleAddr.String(), sender.String()))

	// Burn ibc tokens
	// Note: we burn the ibc vouchers as the packet instructs the other chain to unescrow the source tokens
	if err := w.keeper.BurnIbcTokens(ctx, ibcCoins); err != nil {
		// In case of burn failure, we must unlock the previously locked tokens to maintain state consistency
		if err := w.keeper.UnlockTokens(ctx, sender, nativeCoins); err != nil {
			types.EmitTokenWrapperErrorEvent(ctx, err)
			w.keeper.Logger().Error(fmt.Sprintf("failed to unlock previously locked native tokens: %v", err))
		}
		types.EmitTokenWrapperErrorEvent(ctx, err)
		w.keeper.Logger().Error(fmt.Sprintf("failed to burn ibc tokens: %v", err))
		return 0, err
	}

	w.keeper.Logger().Info(fmt.Sprintf("Burned %s tokens for module address: %s, sender: %s", ibcCoins.String(), moduleAddr.String(), sender.String()))

	// Send the packet with modified data
	sequence, err := w.ics4Wrapper.SendPacket(ctx, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, bz)
	if err != nil {
		types.EmitTokenWrapperErrorEvent(ctx, err)
		w.keeper.Logger().Error(fmt.Sprintf("failed to send packet: %v", err))
		return 0, err
	}

	// Emit success event for wrapped token packet
	types.EmitTokenWrapperPacketEvent(
		ctx,
		types.EventTypeTokenWrapperSendPacket,
		sender.String(),
		receiver,
		convertedAmount,
		sendDenom,
		sourcePort,
		sourceChannel,
		"",
		"",
	)

	return sequence, nil
}
