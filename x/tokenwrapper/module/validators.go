package tokenwrapper

import (
	"fmt"

	types "zigchain/x/tokenwrapper/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
)

// validateChannel checks if the channel exists and is open
func (im IBCModule) validateChannel(ctx sdk.Context, portID, channelID string) error {
	channel, found := im.channelKeeper.GetChannel(ctx, portID, channelID)
	if !found {
		return types.ErrChannelNotFound.Wrapf("%s/%s", portID, channelID)
	}
	if channel.State != channeltypes.OPEN {
		return types.ErrChannelNotOpen.Wrapf("%s/%s, state: %s", portID, channelID, channel.State)
	}
	return nil
}

// validateChannel checks if the channel exists and is open
func (w ICS4Wrapper) validateChannel(ctx sdk.Context, portID, channelID string) error {
	channel, found := w.channelKeeper.GetChannel(ctx, portID, channelID)
	if !found {
		return types.ErrChannelNotFound.Wrapf("%s/%s", portID, channelID)
	}
	if channel.State != channeltypes.OPEN {
		return types.ErrChannelNotOpen.Wrapf("%s/%s, state: %s", portID, channelID, channel.State)
	}
	return nil
}

// checkCounterypartyChannelMatchesIBCSettings checks if the counterparty channel matches the expected IBC settings
func (im IBCModule) checkCounterypartyChannelMatchesIBCSettings(ctx sdk.Context, portID, channelID string) error {
	// Get the expected counterparty port ID from the module keeper
	expectedCounterpartyPortId := im.keeper.GetCounterpartyPort(ctx)
	// Get the expected counterparty channel ID from the module keeper
	expectedCounterpartyChannelId := im.keeper.GetCounterpartyChannel(ctx)

	// If the expected counterparty port ID is not set, return an error
	if expectedCounterpartyPortId == "" {
		return fmt.Errorf("counterparty port ID not set")
	}
	// If the expected counterparty channel ID is not set, return an error
	if expectedCounterpartyChannelId == "" {
		return fmt.Errorf("counterparty channel ID not set")
	}

	channel, found := im.channelKeeper.GetChannel(ctx, portID, channelID)
	if !found {
		return types.ErrChannelNotFound.Wrapf("%s/%s", portID, channelID)
	}
	if channel.State != channeltypes.OPEN {
		return types.ErrChannelNotOpen.Wrapf("%s/%s, state: %s", portID, channelID, channel.State)
	}
	if channel.Counterparty.PortId != expectedCounterpartyPortId {
		return fmt.Errorf("counterparty port ID mismatch: expected %s, got %s", expectedCounterpartyPortId, channel.Counterparty.PortId)
	}
	if channel.Counterparty.ChannelId != expectedCounterpartyChannelId {
		return fmt.Errorf("counterparty channel ID mismatch: expected %s, got %s", expectedCounterpartyChannelId, channel.Counterparty.ChannelId)
	}
	return nil
}

// checkCounterypartyChannelMatchesIBCSettings checks if the counterparty channel matches the expected IBC settings
func (w ICS4Wrapper) checkCounterypartyChannelMatchesIBCSettings(ctx sdk.Context, portID, channelID string) error {
	// Get the expected counterparty port ID from the module keeper
	expectedCounterpartyPortId := w.keeper.GetCounterpartyPort(ctx)
	// Get the expected counterparty channel ID from the module keeper
	expectedCounterpartyChannelId := w.keeper.GetCounterpartyChannel(ctx)

	// If the expected counterparty port ID is not set, return an error
	if expectedCounterpartyPortId == "" {
		return fmt.Errorf("counterparty port ID not set")
	}
	// If the expected counterparty channel ID is not set, return an error
	if expectedCounterpartyChannelId == "" {
		return fmt.Errorf("counterparty channel ID not set")
	}

	channel, found := w.channelKeeper.GetChannel(ctx, portID, channelID)
	if !found {
		return types.ErrChannelNotFound.Wrapf("%s/%s", portID, channelID)
	}
	if channel.State != channeltypes.OPEN {
		return types.ErrChannelNotOpen.Wrapf("%s/%s, state: %s", portID, channelID, channel.State)
	}
	if channel.Counterparty.PortId != expectedCounterpartyPortId {
		return fmt.Errorf("counterparty port ID mismatch: expected %s, got %s", expectedCounterpartyPortId, channel.Counterparty.PortId)
	}
	if channel.Counterparty.ChannelId != expectedCounterpartyChannelId {
		return fmt.Errorf("counterparty channel ID mismatch: expected %s, got %s", expectedCounterpartyChannelId, channel.Counterparty.ChannelId)
	}
	return nil
}

// validateClientId verifies the connection matches the expected native and counterparty client ID
func (im IBCModule) validateConnectionClientId(ctx sdk.Context, portID, channelID string) error {
	// Get the expected client ID from the module keeper
	expectedNativeClientId := im.keeper.GetNativeClientId(ctx)
	expectedCounterpartyClientId := im.keeper.GetCounterpartyClientId(ctx)

	// If the expected native client ID is not set, return an error
	if expectedNativeClientId == "" {
		return fmt.Errorf("native client ID not set")
	}
	// If the expected counterparty client ID is not set, return an error
	if expectedCounterpartyClientId == "" {
		return fmt.Errorf("counterparty client ID not set")
	}

	// Get the channel information for the packet using the native chain port and channel id
	// Note 1: the chain can only retrieve its own IBC connection info and not the counterparty IBC connections
	// Note 2: dest port and channel are equal to native port and channel on a `OnRecvPacket` callback
	// Note 3: src port and channel are equal to native port and channel on a `OnAcknowledgementPacket` callback
	channel, _ := im.channelKeeper.GetChannel(ctx, portID, channelID)

	// check connection hops contains at least one connection hop
	if len(channel.ConnectionHops) == 0 {
		return fmt.Errorf("no connection hops found")
	}

	connectionEnd, found := im.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return fmt.Errorf("connection not found")
	}
	if connectionEnd.ClientId != expectedNativeClientId {
		return fmt.Errorf("native client ID mismatch: expected %s, got %s", expectedNativeClientId, connectionEnd.ClientId)
	}
	if connectionEnd.Counterparty.ClientId != expectedCounterpartyClientId {
		return fmt.Errorf("counterparty client ID mismatch: expected %s, got %s", expectedCounterpartyClientId, connectionEnd.Counterparty.ClientId)
	}
	return nil
}

// validateClientId verifies the connection matches the expected native and counterparty client ID
func (w ICS4Wrapper) validateConnectionClientId(ctx sdk.Context, portID, channelID string) error {
	// Get the expected client ID from the module keeper
	expectedNativeClientId := w.keeper.GetNativeClientId(ctx)
	expectedCounterpartyClientId := w.keeper.GetCounterpartyClientId(ctx)

	// If the expected native client ID is not set, return an error
	if expectedNativeClientId == "" {
		return fmt.Errorf("native client ID not set")
	}
	// If the expected counterparty client ID is not set, return an error
	if expectedCounterpartyClientId == "" {
		return fmt.Errorf("counterparty client ID not set")
	}

	// Get the channel information for the packet using the native chain port and channel id
	// Note 1: the chain can only retrieve its own IBC connection info and not the counterparty IBC connections
	// Note 2: dest port and channel are equal to native port and channel on a `OnRecvPacket` callback
	// Note 3: src port and channel are equal to native port and channel on a `OnAcknowledgementPacket` callback
	channel, _ := w.channelKeeper.GetChannel(ctx, portID, channelID)

	// check connection hops contains at least one connection hop
	if len(channel.ConnectionHops) == 0 {
		return fmt.Errorf("no connection hops found")
	}

	connectionEnd, found := w.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return fmt.Errorf("connection not found")
	}
	if connectionEnd.ClientId != expectedNativeClientId {
		return fmt.Errorf("native client ID mismatch: expected %s, got %s", expectedNativeClientId, connectionEnd.ClientId)
	}
	if connectionEnd.Counterparty.ClientId != expectedCounterpartyClientId {
		return fmt.Errorf("counterparty client ID mismatch: expected %s, got %s", expectedCounterpartyClientId, connectionEnd.Counterparty.ClientId)
	}
	return nil
}

// validateIBCSettingsExist checks if the IBC settings are properly configured
func (im IBCModule) validateIBCSettingsExist(ctx sdk.Context) bool {
	moduleNativePort := im.keeper.GetNativePort(ctx)
	moduleNativeChannel := im.keeper.GetNativeChannel(ctx)
	moduleCounterpartyPort := im.keeper.GetCounterpartyPort(ctx)
	moduleCounterpartyChannel := im.keeper.GetCounterpartyChannel(ctx)
	moduleDenom := im.keeper.GetDenom(ctx)

	if moduleCounterpartyPort == "" || moduleCounterpartyChannel == "" ||
		moduleNativePort == "" || moduleNativeChannel == "" || moduleDenom == "" {
		return false
	}

	return true
}

// validateIBCDenomIsModuleDenom checks if the IBC denom is the module denom
func (im IBCModule) validateIBCDenomIsModuleDenom(ctx sdk.Context, denom string) bool {
	moduleDenom := im.keeper.GetDenom(ctx)
	return denom == moduleDenom
}

// validateReceiverChainIsNotSourceChain verifies that the token's denom path does not include the source chain's port/channel prefix
// this validation ensures tokens are only wrapped when they originate from the sender chain, not the source chain
func (im IBCModule) validateReceiverChainIsNotSourceChain(denom string, packet channeltypes.Packet) bool {
	return !transfertypes.ExtractDenomFromPath(denom).HasPrefix(packet.GetSourcePort(), packet.GetSourceChannel())
}

// validateIBCSettingsMatchOnRecvPacket checks if the IBC settings match the expected values for OnRecvPacket
func (im IBCModule) validateIBCSettingsMatchOnRecvPacket(ctx sdk.Context, packet channeltypes.Packet) bool {
	moduleNativePort := im.keeper.GetNativePort(ctx)
	moduleNativeChannel := im.keeper.GetNativeChannel(ctx)
	moduleCounterpartyPort := im.keeper.GetCounterpartyPort(ctx)
	moduleCounterpartyChannel := im.keeper.GetCounterpartyChannel(ctx)

	if packet.GetSourcePort() == moduleCounterpartyPort &&
		packet.GetSourceChannel() == moduleCounterpartyChannel &&
		packet.GetDestPort() == moduleNativePort &&
		packet.GetDestChannel() == moduleNativeChannel {
		return true
	}

	return false
}

// validateIBCSettingsMatchOnSendPacket checks if the IBC settings match the expected values for SendPacket, OnAcknowledgementPacket and OnTimeoutPacket
func (im IBCModule) validateIBCSettingsMatchOnSendPacket(ctx sdk.Context, packet channeltypes.Packet) bool {
	moduleNativePort := im.keeper.GetNativePort(ctx)
	moduleNativeChannel := im.keeper.GetNativeChannel(ctx)
	moduleCounterpartyPort := im.keeper.GetCounterpartyPort(ctx)
	moduleCounterpartyChannel := im.keeper.GetCounterpartyChannel(ctx)

	if packet.GetSourcePort() == moduleNativePort &&
		packet.GetSourceChannel() == moduleNativeChannel &&
		packet.GetDestPort() == moduleCounterpartyPort &&
		packet.GetDestChannel() == moduleCounterpartyChannel {
		return true
	}

	return false
}
