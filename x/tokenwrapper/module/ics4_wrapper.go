package tokenwrapper

import (
	types "zigchain/x/tokenwrapper/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	porttypes "github.com/cosmos/ibc-go/v10/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"
)

type ICS4Wrapper struct {
	ics4Wrapper      porttypes.ICS4Wrapper
	bankKeeper       types.BankKeeper
	transferKeeper   types.TransferKeeper
	channelKeeper    types.ChannelKeeper
	connectionKeeper types.ConnectionKeeper
	keeper           types.TokenwrapperKeeper
}

func NewICS4Wrapper(app porttypes.ICS4Wrapper, bankKeeper types.BankKeeper, transferKeeper types.TransferKeeper, channelKeeper types.ChannelKeeper, connectionKeeper types.ConnectionKeeper, tokenwrapperKeeper types.TokenwrapperKeeper) *ICS4Wrapper {
	return &ICS4Wrapper{
		ics4Wrapper:      app,
		bankKeeper:       bankKeeper,
		transferKeeper:   transferKeeper,
		channelKeeper:    channelKeeper,
		connectionKeeper: connectionKeeper,
		keeper:           tokenwrapperKeeper,
	}
}

// GetAppVersion implements the ICS4Wrapper interface
func (w *ICS4Wrapper) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	return w.ics4Wrapper.GetAppVersion(ctx, portID, channelID)
}

// WriteAcknowledgement implements the ICS4Wrapper interface
func (w *ICS4Wrapper) WriteAcknowledgement(ctx sdk.Context, packet ibcexported.PacketI, ack ibcexported.Acknowledgement) error {
	return w.ics4Wrapper.WriteAcknowledgement(ctx, packet, ack)
}

// SetTransferKeeper sets the transfer keeper
func (w *ICS4Wrapper) SetTransferKeeper(transferKeeper types.TransferKeeper) {
	w.transferKeeper = transferKeeper
}
