package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v10/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"
)

// ChannelKeeper defines the expected interface for the IBC channel keeper
type ChannelKeeper interface {
	GetChannel(ctx sdk.Context, portID, channelID string) (channeltypes.Channel, bool)
}

// ConnectionKeeper defines the expected interface for the IBC connection keeper
type ConnectionKeeper interface {
	GetConnection(ctx sdk.Context, connectionID string) (connectiontypes.ConnectionEnd, bool)
}

// TransferKeeper defines the expected interface for the Transfer module.
type TransferKeeper interface {
	GetTotalEscrowForDenom(ctx sdk.Context, denom string) sdk.Coin
	SetTotalEscrowForDenom(ctx sdk.Context, coin sdk.Coin)
}

type CallbacksCompatibleModule interface {
	OnRecvPacket(
		ctx sdk.Context,
		channelVersion string,
		packet channeltypes.Packet,
		relayer sdk.AccAddress,
	) ibcexported.Acknowledgement

	OnAcknowledgementPacket(
		ctx sdk.Context,
		channelVersion string,
		packet channeltypes.Packet,
		acknowledgement []byte,
		relayer sdk.AccAddress,
	) error

	OnChanCloseConfirm(
		ctx sdk.Context,
		portID,
		channelID string,
	) error

	OnChanCloseInit(
		ctx sdk.Context,
		portID,
		channelID string,
	) error

	OnChanOpenAck(
		ctx sdk.Context,
		portID,
		channelID string,
		counterpartyChannelID string,
		counterpartyVersion string,
	) error

	OnChanOpenConfirm(
		ctx sdk.Context,
		portID,
		channelID string,
	) error

	OnChanOpenInit(
		ctx sdk.Context,
		order channeltypes.Order,
		connectionHops []string,
		portID string,
		channelID string,
		counterparty channeltypes.Counterparty,
		version string,
	) (string, error)

	OnChanOpenTry(
		ctx sdk.Context,
		order channeltypes.Order,
		connectionHops []string,
		portID,
		channelID string,
		counterparty channeltypes.Counterparty,
		counterpartyVersion string,
	) (version string, err error)

	OnTimeoutPacket(
		ctx sdk.Context,
		channelVersion string,
		packet channeltypes.Packet,
		relayer sdk.AccAddress,
	) error

	UnmarshalPacketData(ctx sdk.Context, portID string, channelID string, bz []byte) (interface{}, string, error)
}

type ICS4Wrapper interface {
	GetAppVersion(
		ctx sdk.Context,
		portID,
		channelID string,
	) (string, bool)

	SendPacket(
		ctx sdk.Context,
		sourcePort string,
		sourceChannel string,
		timeoutHeight clienttypes.Height,
		timeoutTimestamp uint64,
		data []byte,
	) (sequence uint64, err error)

	WriteAcknowledgement(
		ctx sdk.Context,
		packet ibcexported.PacketI,
		ack ibcexported.Acknowledgement,
	) error
}

type TokenwrapperKeeper interface {
	Logger() log.Logger
	IsEnabled(ctx sdk.Context) bool
	GetNativePort(ctx sdk.Context) string
	GetCounterpartyPort(ctx sdk.Context) string
	GetNativeChannel(ctx sdk.Context) string
	GetCounterpartyChannel(ctx sdk.Context) string
	GetDenom(ctx sdk.Context) string
	GetCounterpartyClientId(ctx sdk.Context) string
	GetNativeClientId(ctx sdk.Context) string
	GetDecimalDifference(ctx sdk.Context) uint32
	GetDecimalConversionFactor(ctx sdk.Context) sdkmath.Int
	LockTokens(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) error
	UnlockTokens(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) error
	AddToTotalTransferredIn(ctx sdk.Context, amount sdkmath.Int)
	AddToTotalTransferredOut(ctx sdk.Context, amount sdkmath.Int)
	BurnIbcTokens(ctx sdk.Context, amt sdk.Coins) error
	CheckAccountBalance(ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) error
	CheckModuleBalance(ctx sdk.Context, coins sdk.Coins) error
	GetIBCRecvDenom(ctx sdk.Context, denom string) string
	ScaleDownTokenPrecision(ctx sdk.Context, amount sdkmath.Int) (sdkmath.Int, error)
	ScaleUpTokenPrecision(ctx sdk.Context, amount sdkmath.Int) (sdkmath.Int, error)
	CheckBalances(ctx sdk.Context, receiver sdk.AccAddress, amount sdkmath.Int, recvDenom string, convertedAmount sdkmath.Int) error
	LockIBCTokens(ctx sdk.Context, receiver sdk.AccAddress, amount sdkmath.Int, recvDenom string) (sdk.Coins, error)
	UnlockNativeTokens(ctx sdk.Context, receiver sdk.AccAddress, amount sdkmath.Int, ibcCoins sdk.Coins) (sdk.Coins, error)
}
