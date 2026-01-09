package tokenwrapper

import (
	types "zigchain/x/tokenwrapper/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibccallbackstypes "github.com/cosmos/ibc-go/v10/modules/apps/callbacks/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
)

// IBCModule implements the ICS26 interface for interchain accounts host chains
type IBCModule struct {
	keeper           types.TokenwrapperKeeper
	transferKeeper   types.TransferKeeper
	bankKeeper       types.BankKeeper
	app              ibccallbackstypes.CallbacksCompatibleModule
	channelKeeper    types.ChannelKeeper
	connectionKeeper types.ConnectionKeeper
}

// NewIBCModule creates a new IBCModule given the associated keeper
func NewIBCModule(k types.TokenwrapperKeeper, transferKeeper types.TransferKeeper, bankKeeper types.BankKeeper, app ibccallbackstypes.CallbacksCompatibleModule, channelKeeper types.ChannelKeeper, connectionKeeper types.ConnectionKeeper) IBCModule {
	return IBCModule{
		keeper:           k,
		transferKeeper:   transferKeeper,
		bankKeeper:       bankKeeper,
		app:              app,
		channelKeeper:    channelKeeper,
		connectionKeeper: connectionKeeper,
	}
}

// OnChanOpenInit implements the IBCModule interface
func (im IBCModule) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	return im.app.OnChanOpenInit(
		ctx,
		order,
		connectionHops,
		portID,
		channelID,
		counterparty,
		version,
	)
}

// OnChanOpenTry implements the IBCModule interface
func (im IBCModule) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (string, error) {
	return im.app.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, counterparty, counterpartyVersion)
}

// OnChanOpenAck implements the IBCModule interface
func (im IBCModule) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyChannelId string,
	counterpartyVersion string,
) error {
	return im.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelId, counterpartyVersion)
}

// OnChanOpenConfirm implements the IBCModule interface
func (im IBCModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return im.app.OnChanOpenConfirm(ctx, portID, channelID)
}

// OnChanCloseInit implements the IBCModule interface
func (im IBCModule) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// Disallow user-initiated channel closing for channels
	return im.app.OnChanCloseInit(ctx, portID, channelID)
}

// OnChanCloseConfirm implements the IBCModule interface
func (im IBCModule) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return im.app.OnChanCloseConfirm(ctx, portID, channelID)
}
