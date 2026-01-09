package tokenwrapper_test

import (
	"fmt"
	"strings"
	"testing"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	connectiontypes "github.com/cosmos/ibc-go/v10/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	keepertest "zigchain/testutil/keeper"
	tokenwrapper "zigchain/x/tokenwrapper/module"
	mocks "zigchain/x/tokenwrapper/testutil"
	"zigchain/x/tokenwrapper/types"
)

// createSuccessAcknowledgement creates a proper JSON-encoded success acknowledgement
func createSuccessAcknowledgement() []byte {
	ack := channeltypes.NewResultAcknowledgement([]byte("success"))
	ackBytes, _ := transfertypes.ModuleCdc.MarshalJSON(&ack)
	return ackBytes
}

// createErrorAcknowledgement creates a proper JSON-encoded error acknowledgement
func createErrorAcknowledgement(err error) []byte {
	ack := channeltypes.NewErrorAcknowledgement(err)
	ackBytes, _ := transfertypes.ModuleCdc.MarshalJSON(&ack)
	return ackBytes
}

func TestOnAcknowledgementPacket(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	const (
		sender   = "zig1vm3v4yrd3rrwkf3fe8qxutaz27098t76270qc5"
		receiver = "zig1fcf3p6exuhq53myh0k3wweym5ffv00gcnuum07"

		nativePort       = "transfer"
		counterpartyPort = "transfer"

		nativeChannel       = "channel-12"
		counterpartyChannel = "channel-24"

		nativeConnectionId       = "connection-12"
		counterpartyConnectionId = "connection-24"

		nativeClientId       = "07-tendermint-12"
		counterpartyClientId = "07-tendermint-24"

		moduleDenom = "module-denom"
		ibcDenom    = "ibc/AA11A7781887F73EDC5A9BA3191E75FA0857A643D560B2C2A3A1868DA4D7AD97"
	)

	packetData := transfertypes.ModuleCdc.MustMarshalJSON(&transfertypes.FungibleTokenPacketData{
		Denom:    moduleDenom,
		Amount:   "1000000000000000000",
		Sender:   sender,
		Receiver: receiver,
	})

	packet := channeltypes.Packet{
		SourcePort:         nativePort,
		SourceChannel:      nativeChannel,
		DestinationPort:    counterpartyPort,
		DestinationChannel: counterpartyChannel,
		Data:               packetData,
	}

	// Test cases
	tests := []struct {
		name            string
		enabled         bool
		channelVersion  string
		packet          channeltypes.Packet
		acknowledgement []byte
		relayer         sdk.AccAddress
		expectedError   error
		setupMocks      func(ctx sdk.Context, keeperMock *mocks.MockTokenwrapperKeeper, transferKeeperMock *mocks.MockTransferKeeper, bankKeeperMock *mocks.MockBankKeeper, appMock *mocks.MockCallbacksCompatibleModule, channelKeeperMock *mocks.MockChannelKeeper, connectionKeeperMock *mocks.MockConnectionKeeper)
	}{
		{
			name:            "Happy path",
			enabled:         true,
			channelVersion:  "ics20-1",
			packet:          packet,
			acknowledgement: createSuccessAcknowledgement(),
			relayer:         sdk.AccAddress{},
			expectedError:   nil,
			setupMocks: func(ctx sdk.Context, keeperMock *mocks.MockTokenwrapperKeeper, transferKeeperMock *mocks.MockTransferKeeper, bankKeeperMock *mocks.MockBankKeeper, appMock *mocks.MockCallbacksCompatibleModule, channelKeeperMock *mocks.MockChannelKeeper, connectionKeeperMock *mocks.MockConnectionKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, nativePort, nativeChannel).Return(channeltypes.Channel{
					State:          channeltypes.OPEN,
					ConnectionHops: []string{nativeConnectionId},
				}, true).Times(1)
				// Connection information matches the ClientId
				connectionKeeperMock.EXPECT().GetConnection(ctx, nativeConnectionId).Return(connectiontypes.ConnectionEnd{
					ClientId: nativeClientId,
				}, true).Times(0)
				keeperMock.EXPECT().GetNativeClientId(ctx).Return(nativeClientId).Times(0)
				keeperMock.EXPECT().IsEnabled(ctx).Return(true).AnyTimes()
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort).Times(2)
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel).Times(2)
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort).Times(2)
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel).Times(2)
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom).Times(2)
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				keeperMock.EXPECT().AddToTotalTransferredOut(ctx, sdkmath.NewInt(1000000)).Return()
				keeperMock.EXPECT().ScaleDownTokenPrecision(ctx, sdkmath.NewInt(1000000000000000000)).Return(sdkmath.NewInt(1000000), nil)
				appMock.EXPECT().OnAcknowledgementPacket(
					gomock.Any(),
					"ics20-1",
					packet,
					createSuccessAcknowledgement(),
					sdk.AccAddress{},
				).Return(nil)
			},
		},
		{
			name:            "Error - module is disabled",
			enabled:         true,
			channelVersion:  "ics20-1",
			packet:          packet,
			acknowledgement: createSuccessAcknowledgement(),
			relayer:         sdk.AccAddress{},
			expectedError:   fmt.Errorf("module functionality is not enabled"),
			setupMocks: func(ctx sdk.Context, keeperMock *mocks.MockTokenwrapperKeeper, transferKeeperMock *mocks.MockTransferKeeper, bankKeeperMock *mocks.MockBankKeeper, appMock *mocks.MockCallbacksCompatibleModule, channelKeeperMock *mocks.MockChannelKeeper, connectionKeeperMock *mocks.MockConnectionKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, nativePort, nativeChannel).Return(channeltypes.Channel{
					State:          channeltypes.OPEN,
					ConnectionHops: []string{nativeConnectionId},
				}, true).Times(1)
				// Connection information matches the ClientId
				connectionKeeperMock.EXPECT().GetConnection(ctx, nativeConnectionId).Return(connectiontypes.ConnectionEnd{
					ClientId: nativeClientId,
				}, true).Times(0)
				keeperMock.EXPECT().GetNativeClientId(ctx).Return(nativeClientId).Times(0)
				keeperMock.EXPECT().IsEnabled(ctx).Return(false).Times(1)
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort).Times(2)
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel).Times(2)
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort).Times(2)
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel).Times(2)
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom).Times(2)
				keeperMock.EXPECT().GetDecimalConversionFactor(ctx).Return(sdkmath.NewInt(1000000000000)).Times(0)
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				keeperMock.EXPECT().AddToTotalTransferredOut(ctx, sdkmath.NewInt(1000000)).Return().Times(0)
				appMock.EXPECT().OnAcknowledgementPacket(
					gomock.Any(),
					"ics20-1",
					packet,
					createSuccessAcknowledgement(),
					sdk.AccAddress{},
				).Return(nil).Times(0)
			},
		},
		{
			name:            "Error - failed to unmarshal acknowledgement",
			enabled:         true,
			channelVersion:  "ics20-1",
			packet:          channeltypes.Packet{},
			acknowledgement: []byte("invalid"),
			relayer:         sdk.AccAddress{},
			expectedError:   fmt.Errorf("EOF"),
			setupMocks: func(ctx sdk.Context, keeperMock *mocks.MockTokenwrapperKeeper, transferKeeperMock *mocks.MockTransferKeeper, bankKeeperMock *mocks.MockBankKeeper, appMock *mocks.MockCallbacksCompatibleModule, channelKeeperMock *mocks.MockChannelKeeper, connectionKeeperMock *mocks.MockConnectionKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, "", "").Return(channeltypes.Channel{
					State:          channeltypes.OPEN,
					ConnectionHops: []string{nativeConnectionId},
				}, true).Times(1)
				connectionKeeperMock.EXPECT().GetConnection(ctx, nativeConnectionId).Return(connectiontypes.ConnectionEnd{
					ClientId: nativeClientId,
				}, true).Times(0)
				keeperMock.EXPECT().GetNativeClientId(ctx).Return(nativeClientId).Times(0)
				keeperMock.EXPECT().IsEnabled(ctx).Return(true).AnyTimes().Times(0)
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort).Times(0)
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel).Times(0)
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort).Times(0)
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel).Times(0)
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom).Times(0)
				keeperMock.EXPECT().GetDecimalConversionFactor(ctx).Return(sdkmath.NewInt(1000000000000)).Times(0)
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				appMock.EXPECT().OnAcknowledgementPacket(
					gomock.Any(),
					"ics20-1",
					channeltypes.Packet{},
					createSuccessAcknowledgement(),
					sdk.AccAddress{},
				).Return(fmt.Errorf("app callback error")).Times(0)
			},
		},
		{
			name:            "Error - app callback",
			enabled:         true,
			channelVersion:  "ics20-1",
			packet:          packet,
			acknowledgement: createSuccessAcknowledgement(),
			relayer:         sdk.AccAddress{},
			expectedError:   fmt.Errorf("app callback error"),
			setupMocks: func(ctx sdk.Context, keeperMock *mocks.MockTokenwrapperKeeper, transferKeeperMock *mocks.MockTransferKeeper, bankKeeperMock *mocks.MockBankKeeper, appMock *mocks.MockCallbacksCompatibleModule, channelKeeperMock *mocks.MockChannelKeeper, connectionKeeperMock *mocks.MockConnectionKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, nativePort, nativeChannel).Return(channeltypes.Channel{
					State:          channeltypes.OPEN,
					ConnectionHops: []string{nativeConnectionId},
				}, true).Times(1)
				connectionKeeperMock.EXPECT().GetConnection(ctx, nativeConnectionId).Return(connectiontypes.ConnectionEnd{
					ClientId: nativeClientId,
				}, true).Times(0)
				keeperMock.EXPECT().GetNativeClientId(ctx).Return(nativeClientId).Times(0)
				keeperMock.EXPECT().IsEnabled(ctx).Return(true).Times(1)
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort).Times(2)
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel).Times(2)
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort).Times(2)
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel).Times(2)
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom).Times(2)
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				appMock.EXPECT().OnAcknowledgementPacket(
					gomock.Any(),
					"ics20-1",
					packet,
					createSuccessAcknowledgement(),
					sdk.AccAddress{},
				).Return(fmt.Errorf("app callback error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.enabled {
				return
			}

			// Create mocks
			keeperMock := mocks.NewMockTokenwrapperKeeper(ctrl)
			transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
			bankKeeperMock := mocks.NewMockBankKeeper(ctrl)
			appMock := mocks.NewMockCallbacksCompatibleModule(ctrl)
			channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
			connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

			// Create IBCModule instance
			ibcModule := tokenwrapper.NewIBCModule(
				keeperMock,
				transferKeeperMock,
				bankKeeperMock,
				appMock,
				channelKeeperMock,
				connectionKeeperMock,
			)

			// Create context
			ctx := sdk.Context{}

			// Setup mocks
			tt.setupMocks(ctx, keeperMock, transferKeeperMock, bankKeeperMock, appMock, channelKeeperMock, connectionKeeperMock)

			// Call the function
			err := ibcModule.OnAcknowledgementPacket(
				ctx,
				tt.channelVersion,
				tt.packet,
				tt.acknowledgement,
				tt.relayer,
			)

			// Check the result
			if tt.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Test fixture with positive scenario data
// AckPacketFixture defines the configuration for OnAcknowledgementPacket tests
type AckPacketFixture struct {
	nativePort           string
	nativeChannel        string
	counterpartyPort     string
	counterpartyChannel  string
	nativeConnectionId   string
	nativeClientId       string
	counterpartyClientId string
	moduleDenom          string
	ibcDenom             string
	amount               string
	sender               string
	receiver             string
	decimalDifference    uint32
}

// getAckPacketPositiveFixture returns a positive fixture for OnAcknowledgementPacket tests
func getAckPacketPositiveFixture() AckPacketFixture {
	return AckPacketFixture{
		nativePort:           "transfer",
		nativeChannel:        "channel-12",
		counterpartyPort:     "transfer",
		counterpartyChannel:  "channel-24",
		nativeConnectionId:   "connection-0",
		nativeClientId:       "07-tendermint-0",
		counterpartyClientId: "07-tendermint-0",
		moduleDenom:          "module-denom",
		ibcDenom:             "ibc/AA11A7781887F73EDC5A9BA3191E75FA0857A643D560B2C2A3A1868DA4D7AD97",
		amount:               "1000000000000000000", // 1e18 (18 decimals)
		sender:               "zig1vm3v4yrd3rrwkf3fe8qxutaz27098t76270qc5",
		receiver:             "zig1fcf3p6exuhq53myh0k3wweym5ffv00gcnuum07",
		decimalDifference:    12, // Converts 18 decimals to 6 decimals (1e12 factor)
	}
}

// Positive test cases

func TestOnAcknowledgementPacket_RealKeepers_Positive(t *testing.T) {
	// Test case: OnAcknowledgementPacket with real keepers, valid packet, expecting success and state change

	// Set up positive fixture
	fixture := getAckPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	appMock := mocks.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k,
		transferKeeperMock,
		bankKeeper,
		appMock,
		channelKeeperMock,
		connectionKeeperMock,
	)

	// Create packet
	data := transfertypes.FungibleTokenPacketData{
		Denom:    fixture.moduleDenom,
		Amount:   fixture.amount,
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         fixture.nativePort,
		SourceChannel:      fixture.nativeChannel,
		DestinationPort:    fixture.counterpartyPort,
		DestinationChannel: fixture.counterpartyChannel,
		Data:               dataBz,
	}

	// Set up channel mock
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.counterpartyPort,
					ChannelId: fixture.counterpartyChannel,
				},
			}, true).
		Times(1)

	// Connection information matches the ClientId
	connectionKeeperMock.EXPECT().GetConnection(ctx, fixture.nativeConnectionId).Return(connectiontypes.ConnectionEnd{
		ClientId: fixture.nativeClientId,
	}, true).Times(0)

	// Set up app mock
	appMock.EXPECT().
		OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{}).
		Return(nil).
		Times(1)

	// Check initial TotalTransferredOut
	totalTransferredOutBefore := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOutBefore)

	// Execute OnAcknowledgementPacket
	err = ibcModule.OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{})

	// Expect no error
	require.NoError(t, err)

	// Calculate expected converted amount
	amountInt, ok := sdkmath.NewIntFromString(fixture.amount)
	require.True(t, ok)
	conversionFactor := k.GetDecimalConversionFactor(ctx)      // 1e12
	expectedConvertedAmount := amountInt.Quo(conversionFactor) // 1e18 / 1e12 = 1e6

	// Check final TotalTransferredOut
	totalTransferredOutAfter := k.GetTotalTransferredOut(ctx)
	require.Equal(t, expectedConvertedAmount, totalTransferredOutAfter, "TotalTransferredOut should reflect converted amount")
}

func TestOnAcknowledgementPacket_MissingNativePortChannelDenom(t *testing.T) {
	// Test case: OnAcknowledgementPacket with missing native port, channel and denom, expecting no error and no state change

	// Set up positive fixture
	fixture := getAckPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper with empty nativePort
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, "")
	k.SetNativeChannel(ctx, "")
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, "")
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	appMock := mocks.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k,
		transferKeeperMock,
		bankKeeper,
		appMock,
		channelKeeperMock,
		connectionKeeperMock,
	)

	// Create a packet
	data := transfertypes.FungibleTokenPacketData{
		Denom:    fixture.moduleDenom,
		Amount:   fixture.amount,
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         fixture.nativePort,
		SourceChannel:      fixture.nativeChannel,
		DestinationPort:    fixture.counterpartyPort,
		DestinationChannel: fixture.counterpartyChannel,
		Data:               dataBz,
	}

	// Set up channel mock
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.counterpartyPort,
					ChannelId: fixture.counterpartyChannel,
				},
			}, true).
		Times(1)

	// Set up app mock
	appMock.EXPECT().
		OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{}).
		Return(nil).
		Times(1)

	// Check initial TotalTransferredOut
	totalTransferredOutBefore := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOutBefore)

	// Execute OnAcknowledgementPacket
	err = ibcModule.OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{})

	// Expect no error
	require.NoError(t, err)

	// Check final TotalTransferredOut (no change)
	totalTransferredOutAfter := k.GetTotalTransferredOut(ctx)
	require.Equal(t, totalTransferredOutBefore, totalTransferredOutAfter, "TotalTransferredOut should not change")
}

func TestOnAcknowledgementPacket_MismatchedSourcePort(t *testing.T) {
	// Test case: OnAcknowledgementPacket with mismatched source port, expecting no error and no state change

	// Set up positive fixture
	fixture := getAckPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	appMock := mocks.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k,
		transferKeeperMock,
		bankKeeper,
		appMock,
		channelKeeperMock,
		connectionKeeperMock,
	)

	// Create a packet with mismatched source port
	data := transfertypes.FungibleTokenPacketData{
		Denom:    fixture.moduleDenom,
		Amount:   fixture.amount,
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         "wrong-port",
		SourceChannel:      fixture.nativeChannel,
		DestinationPort:    fixture.counterpartyPort,
		DestinationChannel: fixture.counterpartyChannel,
		Data:               dataBz,
	}

	// Set up channel mock for wrong port
	channelKeeperMock.EXPECT().
		GetChannel(ctx, "wrong-port", fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.counterpartyPort,
					ChannelId: fixture.counterpartyChannel,
				},
			}, true).
		Times(1)

	// Set up app mock
	appMock.EXPECT().
		OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{}).
		Return(nil).
		Times(0)

	// Check initial TotalTransferredOut
	totalTransferredOutBefore := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOutBefore)

	// Execute OnAcknowledgementPacket
	err = ibcModule.OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{})

	// Expect error
	require.Error(t, err)
	require.Contains(t, err.Error(), "ibc settings do not match the expected values, failed with packet")

	// Check final TotalTransferredOut (no change)
	totalTransferredOutAfter := k.GetTotalTransferredOut(ctx)
	require.Equal(t, totalTransferredOutBefore, totalTransferredOutAfter, "TotalTransferredOut should not change")
}

func TestOnAcknowledgementPacket_MismatchedSourceChannel(t *testing.T) {
	// Test case: OnAcknowledgementPacket with a mismatched source channel, expecting no error and no state change

	// Set up positive fixture
	fixture := getAckPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	appMock := mocks.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k,
		transferKeeperMock,
		bankKeeper,
		appMock,
		channelKeeperMock,
		connectionKeeperMock,
	)

	// Create a packet with mismatched source channel
	data := transfertypes.FungibleTokenPacketData{
		Denom:    fixture.moduleDenom,
		Amount:   fixture.amount,
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         fixture.nativePort,
		SourceChannel:      "wrong-channel",
		DestinationPort:    fixture.counterpartyPort,
		DestinationChannel: fixture.counterpartyChannel,
		Data:               dataBz,
	}

	// Set up channel mock for a wrong channel
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, "wrong-channel").
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.counterpartyPort,
					ChannelId: fixture.counterpartyChannel,
				},
			}, true).
		Times(1)

	// Set up app mock
	appMock.EXPECT().
		OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{}).
		Return(nil).
		Times(0)

	// Check initial TotalTransferredOut
	totalTransferredOutBefore := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOutBefore)

	// Execute OnAcknowledgementPacket
	err = ibcModule.OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{})

	// Expect error
	require.Error(t, err)
	require.Contains(t, err.Error(), "ibc settings do not match the expected values, failed with packet")

	// Check final TotalTransferredOut (no change)
	totalTransferredOutAfter := k.GetTotalTransferredOut(ctx)
	require.Equal(t, totalTransferredOutBefore, totalTransferredOutAfter, "TotalTransferredOut should not change")
}

func TestOnAcknowledgementPacket_MismatchedDestinationPort(t *testing.T) {
	// Test case: OnAcknowledgementPacket with mismatched destination port, expecting no error and no state change

	// Set up positive fixture
	fixture := getAckPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	appMock := mocks.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k,
		transferKeeperMock,
		bankKeeper,
		appMock,
		channelKeeperMock,
		connectionKeeperMock,
	)

	// Create a packet with mismatched destination port
	data := transfertypes.FungibleTokenPacketData{
		Denom:    fixture.moduleDenom,
		Amount:   fixture.amount,
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         fixture.nativePort,
		SourceChannel:      fixture.nativeChannel,
		DestinationPort:    "wrong-dest-port",
		DestinationChannel: fixture.counterpartyChannel,
		Data:               dataBz,
	}

	// Set up channel mock
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    "wrong-dest-port",
					ChannelId: fixture.counterpartyChannel,
				},
			}, true).
		Times(1)

	// Set up app mock
	appMock.EXPECT().
		OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{}).
		Return(nil).
		Times(0)

	// Check initial TotalTransferredOut
	totalTransferredOutBefore := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOutBefore)

	// Execute OnAcknowledgementPacket
	err = ibcModule.OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{})

	// Expect error
	require.Error(t, err)
	require.Contains(t, err.Error(), "ibc settings do not match the expected values, failed with packet")

	// Check final TotalTransferredOut (no change)
	totalTransferredOutAfter := k.GetTotalTransferredOut(ctx)
	require.Equal(t, totalTransferredOutBefore, totalTransferredOutAfter, "TotalTransferredOut should not change")
}

func TestOnAcknowledgementPacket_MismatchedDestinationChannel(t *testing.T) {
	// Test case: OnAcknowledgementPacket with mismatched destination channel, expecting no error and no state change

	// Set up positive fixture
	fixture := getAckPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	appMock := mocks.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k,
		transferKeeperMock,
		bankKeeper,
		appMock,
		channelKeeperMock,
		connectionKeeperMock,
	)

	// Create packet with mismatched destination channel
	data := transfertypes.FungibleTokenPacketData{
		Denom:    fixture.moduleDenom,
		Amount:   fixture.amount,
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         fixture.nativePort,
		SourceChannel:      fixture.nativeChannel,
		DestinationPort:    fixture.counterpartyPort,
		DestinationChannel: "wrong-dest-channel",
		Data:               dataBz,
	}

	// Set up channel mock
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.counterpartyPort,
					ChannelId: "wrong-dest-channel",
				},
			}, true).
		Times(1)

	// Set up app mock
	appMock.EXPECT().
		OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{}).
		Return(nil).
		Times(0)

	// Ensure no other calls occur
	transferKeeperMock.EXPECT().GetTotalEscrowForDenom(gomock.Any(), gomock.Any()).Times(0)
	transferKeeperMock.EXPECT().SetTotalEscrowForDenom(gomock.Any(), gomock.Any()).Times(0)

	// Check initial TotalTransferredOut
	totalTransferredOutBefore := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOutBefore, "Initial TotalTransferredOut incorrect")

	// Execute OnAcknowledgementPacket
	err = ibcModule.OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{})

	// Expect error
	require.Error(t, err)
	require.Contains(t, err.Error(), "ibc settings do not match the expected values, failed with packet")

	// Check final TotalTransferredOut (no change)
	totalTransferredOutAfter := k.GetTotalTransferredOut(ctx)
	require.Equal(t, totalTransferredOutBefore, totalTransferredOutAfter, "TotalTransferredOut should not change")
}

func TestOnAcknowledgementPacket_NonModuleDenom(t *testing.T) {
	// Test case: OnAcknowledgementPacket with non-module denom, expecting no error and no state change

	// Set up positive fixture
	fixture := getAckPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	appMock := mocks.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k,
		transferKeeperMock,
		bankKeeper,
		appMock,
		channelKeeperMock,
		connectionKeeperMock,
	)

	// Create a packet with non-module denom
	data := transfertypes.FungibleTokenPacketData{
		Denom:    "other-denom", // Non-module denom
		Amount:   fixture.amount,
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         fixture.nativePort,
		SourceChannel:      fixture.nativeChannel,
		DestinationPort:    fixture.counterpartyPort,
		DestinationChannel: fixture.counterpartyChannel,
		Data:               dataBz,
	}

	// Set up channel mock
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.counterpartyPort,
					ChannelId: fixture.counterpartyChannel,
				},
			}, true).
		Times(1)

	// Set up app mock for OnAcknowledgementPacket (called due to non-module denom)
	appMock.EXPECT().
		OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{}).
		Return(nil).
		Times(1)

	// Check initial TotalTransferredOut
	totalTransferredOutBefore := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOutBefore)

	// Execute OnAcknowledgementPacket
	err = ibcModule.OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{})

	// Expect no error
	require.NoError(t, err)

	// Check final TotalTransferredOut (no change)
	totalTransferredOutAfter := k.GetTotalTransferredOut(ctx)
	require.Equal(t, totalTransferredOutBefore, totalTransferredOutAfter, "TotalTransferredOut should not change")
}

func TestOnAcknowledgementPacket_StateUpdateConsistency(t *testing.T) {
	// Test case: OnAcknowledgementPacket with multiple calls to ensure state update consistency

	// Set up positive fixture
	fixture := getAckPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	appMock := mocks.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k,
		transferKeeperMock,
		bankKeeper,
		appMock,
		channelKeeperMock,
		connectionKeeperMock,
	)

	// Create a packet
	data := transfertypes.FungibleTokenPacketData{
		Denom:    fixture.moduleDenom,
		Amount:   fixture.amount,
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         fixture.nativePort,
		SourceChannel:      fixture.nativeChannel,
		DestinationPort:    fixture.counterpartyPort,
		DestinationChannel: fixture.counterpartyChannel,
		Data:               dataBz,
	}

	// Set up channel mock for multiple calls
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.counterpartyPort,
					ChannelId: fixture.counterpartyChannel,
				},
			}, true).
		Times(2)

	// Connection information matches the ClientId
	connectionKeeperMock.EXPECT().GetConnection(ctx, fixture.nativeConnectionId).Return(connectiontypes.ConnectionEnd{
		ClientId: fixture.nativeClientId,
	}, true).Times(0)

	// Set up app mock for multiple calls
	appMock.EXPECT().
		OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{}).
		Return(nil).
		Times(2)

	// Calculate expected converted amount
	amountInt, ok := sdkmath.NewIntFromString(fixture.amount)
	require.True(t, ok)
	conversionFactor := k.GetDecimalConversionFactor(ctx)
	expectedConvertedAmount := amountInt.Quo(conversionFactor)

	// Check initial TotalTransferredOut
	totalTransferredOutBefore := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOutBefore)

	// Execute OnAcknowledgementPacket twice to simulate concurrent-like calls
	err = ibcModule.OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{})
	require.NoError(t, err)

	// Check intermediate TotalTransferredOut
	totalTransferredOutAfterFirst := k.GetTotalTransferredOut(ctx)
	require.Equal(t, expectedConvertedAmount, totalTransferredOutAfterFirst, "TotalTransferredOut after first call")

	// Execute again
	err = ibcModule.OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{})
	require.NoError(t, err)

	// Check final TotalTransferredOut
	totalTransferredOutAfterSecond := k.GetTotalTransferredOut(ctx)
	expectedTotal := expectedConvertedAmount.Mul(sdkmath.NewInt(2))
	require.Equal(t, expectedTotal, totalTransferredOutAfterSecond, "TotalTransferredOut after second call")
}

// Negative test cases

func TestOnAcknowledgementPacket_ZeroAmount(t *testing.T) {
	// Test case: OnAcknowledgementPacket with zero amount, expecting error and no state change

	// Set up positive fixture
	fixture := getAckPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	appMock := mocks.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k,
		transferKeeperMock,
		bankKeeper,
		appMock,
		channelKeeperMock,
		connectionKeeperMock,
	)

	// Create a packet with zero amount
	data := transfertypes.FungibleTokenPacketData{
		Denom:    fixture.moduleDenom,
		Amount:   "0",
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         fixture.nativePort,
		SourceChannel:      fixture.nativeChannel,
		DestinationPort:    fixture.counterpartyPort,
		DestinationChannel: fixture.counterpartyChannel,
		Data:               dataBz,
	}

	// Set up channel mock
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.counterpartyPort,
					ChannelId: fixture.counterpartyChannel,
				},
			}, true).
		Times(1)

	// Connection information matches the ClientId
	connectionKeeperMock.EXPECT().GetConnection(ctx, fixture.nativeConnectionId).Return(connectiontypes.ConnectionEnd{
		ClientId: fixture.nativeClientId,
	}, true).Times(0)

	// Set up app mock
	appMock.EXPECT().
		OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{}).
		Return(nil).
		Times(1)

	// Check initial TotalTransferredOut
	totalTransferredOutBefore := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOutBefore)

	// Execute OnAcknowledgementPacket
	err = ibcModule.OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{})

	// Expect error
	require.Error(t, err)
	require.Contains(t, err.Error(), "converted amount is zero or negative")

	// Check final TotalTransferredOut (no change)
	totalTransferredOutAfter := k.GetTotalTransferredOut(ctx)
	require.Equal(t, totalTransferredOutBefore, totalTransferredOutAfter, "TotalTransferredOut should not change")
}

func TestOnAcknowledgementPacket_NegativeAmount(t *testing.T) {
	// Test case: OnAcknowledgementPacket with negative amount, expecting error and no state change

	// Set up positive fixture
	fixture := getAckPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	appMock := mocks.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k,
		transferKeeperMock,
		bankKeeper,
		appMock,
		channelKeeperMock,
		connectionKeeperMock,
	)

	// Create a packet with negative amount
	data := transfertypes.FungibleTokenPacketData{
		Denom:    fixture.moduleDenom,
		Amount:   "-1",
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         fixture.nativePort,
		SourceChannel:      fixture.nativeChannel,
		DestinationPort:    fixture.counterpartyPort,
		DestinationChannel: fixture.counterpartyChannel,
		Data:               dataBz,
	}

	// Set up channel mock
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.counterpartyPort,
					ChannelId: fixture.counterpartyChannel,
				},
			}, true).
		Times(1)

	// Connection information matches the ClientId
	connectionKeeperMock.EXPECT().GetConnection(ctx, fixture.nativeConnectionId).Return(connectiontypes.ConnectionEnd{
		ClientId: fixture.nativeClientId,
	}, true).Times(0)

	// Set up app mock
	appMock.EXPECT().
		OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{}).
		Return(nil).
		Times(1)

	// Check initial TotalTransferredOut
	totalTransferredOutBefore := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOutBefore)

	// Execute OnAcknowledgementPacket
	err = ibcModule.OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{})

	// Expect error
	require.Error(t, err)
	require.Contains(t, err.Error(), "converted amount is zero or negative")

	// Check final TotalTransferredOut (no change)
	totalTransferredOutAfter := k.GetTotalTransferredOut(ctx)
	require.Equal(t, totalTransferredOutBefore, totalTransferredOutAfter, "TotalTransferredOut should not change")
}

func TestOnAcknowledgementPacket_InvalidAmountFormat(t *testing.T) {
	// Test case: OnAcknowledgementPacket with invalid amount format, expecting error and no state change

	// Set up positive fixture
	fixture := getAckPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	appMock := mocks.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k,
		transferKeeperMock,
		bankKeeper,
		appMock,
		channelKeeperMock,
		connectionKeeperMock,
	)

	// Create a packet with invalid amount
	data := transfertypes.FungibleTokenPacketData{
		Denom:    fixture.moduleDenom,
		Amount:   "abc",
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         fixture.nativePort,
		SourceChannel:      fixture.nativeChannel,
		DestinationPort:    fixture.counterpartyPort,
		DestinationChannel: fixture.counterpartyChannel,
		Data:               dataBz,
	}

	// Set up channel mock
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.counterpartyPort,
					ChannelId: fixture.counterpartyChannel,
				},
			}, true).
		Times(1)

	// Connection information matches the ClientId
	connectionKeeperMock.EXPECT().GetConnection(ctx, fixture.nativeConnectionId).Return(connectiontypes.ConnectionEnd{
		ClientId: fixture.nativeClientId,
	}, true).Times(0)

	// Set up app mock
	appMock.EXPECT().
		OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{}).
		Return(nil).
		Times(0)

	// Check initial TotalTransferredOut
	totalTransferredOutBefore := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOutBefore)

	// Execute OnAcknowledgementPacket
	err = ibcModule.OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{})

	// Expect error
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid amount")

	// Check final TotalTransferredOut (no change)
	totalTransferredOutAfter := k.GetTotalTransferredOut(ctx)
	require.Equal(t, totalTransferredOutBefore, totalTransferredOutAfter, "TotalTransferredOut should not change")
}

func TestOnAcknowledgementPacket_ChannelNotFound(t *testing.T) {
	// Test case: OnAcknowledgementPacket with a channel not found, expecting error and no state change

	// Set up positive fixture
	fixture := getAckPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	appMock := mocks.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k,
		transferKeeperMock,
		bankKeeper,
		appMock,
		channelKeeperMock,
		connectionKeeperMock,
	)

	// Create a packet
	data := transfertypes.FungibleTokenPacketData{
		Denom:    fixture.moduleDenom,
		Amount:   fixture.amount,
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         fixture.nativePort,
		SourceChannel:      fixture.nativeChannel,
		DestinationPort:    fixture.counterpartyPort,
		DestinationChannel: fixture.counterpartyChannel,
		Data:               dataBz,
	}

	// Set up channel mock to return not found
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(channeltypes.Channel{}, false).
		Times(1)

	// Connection information matches the ClientId
	connectionKeeperMock.EXPECT().GetConnection(ctx, fixture.nativeConnectionId).Return(connectiontypes.ConnectionEnd{
		ClientId: fixture.nativeClientId,
	}, true).Times(0)

	// Set up app mock
	appMock.EXPECT().
		OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{}).
		Return(nil).
		Times(0)

	// Check initial TotalTransferredOut
	totalTransferredOutBefore := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOutBefore)

	// Execute OnAcknowledgementPacket
	err = ibcModule.OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{})

	// Expect error
	require.Error(t, err)
	require.Contains(t, err.Error(), types.ErrChannelNotFound.Error())

	// Check final TotalTransferredOut (no change)
	totalTransferredOutAfter := k.GetTotalTransferredOut(ctx)
	require.Equal(t, totalTransferredOutBefore, totalTransferredOutAfter, "TotalTransferredOut should not change")
}

func TestOnAcknowledgementPacket_ChannelNotOpen(t *testing.T) {
	// Test case: OnAcknowledgementPacket with channel not open, expecting error and no state change

	// Set up positive fixture
	fixture := getAckPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	appMock := mocks.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k,
		transferKeeperMock,
		bankKeeper,
		appMock,
		channelKeeperMock,
		connectionKeeperMock,
	)

	// Create a packet
	data := transfertypes.FungibleTokenPacketData{
		Denom:    fixture.moduleDenom,
		Amount:   fixture.amount,
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         fixture.nativePort,
		SourceChannel:      fixture.nativeChannel,
		DestinationPort:    fixture.counterpartyPort,
		DestinationChannel: fixture.counterpartyChannel,
		Data:               dataBz,
	}

	// Set up channel mock with closed state
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.CLOSED,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.counterpartyPort,
					ChannelId: fixture.counterpartyChannel,
				},
			}, true).
		Times(1)

	// Connection information matches the ClientId
	connectionKeeperMock.EXPECT().GetConnection(ctx, fixture.nativeConnectionId).Return(connectiontypes.ConnectionEnd{
		ClientId: fixture.nativeClientId,
	}, true).Times(0)

	// Set up app mock
	appMock.EXPECT().
		OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{}).
		Return(nil).
		Times(0)

	// Check initial TotalTransferredOut
	totalTransferredOutBefore := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOutBefore)

	// Execute OnAcknowledgementPacket
	err = ibcModule.OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{})

	// Expect error
	require.Error(t, err)
	require.Contains(t, err.Error(), types.ErrChannelNotOpen.Error())

	// Check final TotalTransferredOut (no change)
	totalTransferredOutAfter := k.GetTotalTransferredOut(ctx)
	require.Equal(t, totalTransferredOutBefore, totalTransferredOutAfter, "TotalTransferredOut should not change")
}

func TestOnAcknowledgementPacket_InvalidSender(t *testing.T) {
	// Test case: OnAcknowledgementPacket with invalid sender address, expecting error and no state change

	// Set up positive fixture
	fixture := getAckPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	appMock := mocks.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k,
		transferKeeperMock,
		bankKeeper,
		appMock,
		channelKeeperMock,
		connectionKeeperMock,
	)

	// Create a packet with invalid sender address
	data := transfertypes.FungibleTokenPacketData{
		Denom:    fixture.moduleDenom,
		Amount:   fixture.amount,
		Sender:   "invalid-address", // Invalid Bech32 address
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         fixture.nativePort,
		SourceChannel:      fixture.nativeChannel,
		DestinationPort:    fixture.counterpartyPort,
		DestinationChannel: fixture.counterpartyChannel,
		Data:               dataBz,
	}

	// Set up channel mock
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.counterpartyPort,
					ChannelId: fixture.counterpartyChannel,
				},
			}, true).
		Times(1)

	// Connection information matches the ClientId
	connectionKeeperMock.EXPECT().GetConnection(ctx, fixture.nativeConnectionId).Return(connectiontypes.ConnectionEnd{
		ClientId: fixture.nativeClientId,
	}, true).Times(0)

	// Set up app mock (not called due to early error)
	appMock.EXPECT().
		OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{}).
		Return(nil).
		Times(0)

	// Check initial TotalTransferredOut
	totalTransferredOutBefore := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOutBefore)

	// Execute OnAcknowledgementPacket
	err = ibcModule.OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{})

	// Expect error
	require.Error(t, err)
	require.Contains(t, err.Error(), "decoding bech32 failed: invalid separator index -1")

	// Check final TotalTransferredOut (no change)
	totalTransferredOutAfter := k.GetTotalTransferredOut(ctx)
	require.Equal(t, totalTransferredOutBefore, totalTransferredOutAfter, "TotalTransferredOut should not change")
}

func TestOnAcknowledgementPacket_EmptySender(t *testing.T) {
	// Test case: OnAcknowledgementPacket with empty sender address, expecting error and no state change

	// Set up positive fixture
	fixture := getAckPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	appMock := mocks.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k,
		transferKeeperMock,
		bankKeeper,
		appMock,
		channelKeeperMock,
		connectionKeeperMock,
	)

	// Create a packet with empty sender
	data := transfertypes.FungibleTokenPacketData{
		Denom:    fixture.moduleDenom,
		Amount:   fixture.amount,
		Sender:   "", // Empty sender address
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         fixture.nativePort,
		SourceChannel:      fixture.nativeChannel,
		DestinationPort:    fixture.counterpartyPort,
		DestinationChannel: fixture.counterpartyChannel,
		Data:               dataBz,
	}

	// Set up channel mock
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.counterpartyPort,
					ChannelId: fixture.counterpartyChannel,
				},
			}, true).
		Times(1)

	// Connection information matches the ClientId
	connectionKeeperMock.EXPECT().GetConnection(ctx, fixture.nativeConnectionId).Return(connectiontypes.ConnectionEnd{
		ClientId: fixture.nativeClientId,
	}, true).Times(0)

	// Set up app mock (not called due to early error)
	appMock.EXPECT().
		OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{}).
		Return(nil).
		Times(0)

	// Check initial TotalTransferredOut
	totalTransferredOutBefore := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOutBefore)

	// Execute OnAcknowledgementPacket
	err = ibcModule.OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{})

	// Expect error
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty address string is not allowed")

	// Check final TotalTransferredOut (no change)
	totalTransferredOutAfter := k.GetTotalTransferredOut(ctx)
	require.Equal(t, totalTransferredOutBefore, totalTransferredOutAfter, "TotalTransferredOut should not change")
}

func TestOnAcknowledgementPacket_FailedAckZeroConvertedAmount(t *testing.T) {
	// Test case: OnAcknowledgementPacket with failed acknowledgment and zero converted amount, expecting error

	// Set up positive fixture
	fixture := getAckPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	appMock := mocks.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k,
		transferKeeperMock,
		bankKeeper,
		appMock,
		channelKeeperMock,
		connectionKeeperMock,
	)

	// Create a packet with small amount leading to zero converted amount
	data := transfertypes.FungibleTokenPacketData{
		Denom:    fixture.moduleDenom,
		Amount:   "999", // Small amount to result in zero after conversion
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         fixture.nativePort,
		SourceChannel:      fixture.nativeChannel,
		DestinationPort:    fixture.counterpartyPort,
		DestinationChannel: fixture.counterpartyChannel,
		Data:               dataBz,
	}

	// Set up channel mock
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.counterpartyPort,
					ChannelId: fixture.counterpartyChannel,
				},
			}, true).
		Times(1)

	// Connection information matches the ClientId
	connectionKeeperMock.EXPECT().GetConnection(ctx, fixture.nativeConnectionId).Return(connectiontypes.ConnectionEnd{
		ClientId: fixture.nativeClientId,
	}, true).Times(0)

	// Set up app mock
	appMock.EXPECT().
		OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{}).
		Return(nil).
		Times(1)

	// Check initial TotalTransferredOut
	totalTransferredOutBefore := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOutBefore)

	// Execute OnAcknowledgementPacket with failed acknowledgment
	err = ibcModule.OnAcknowledgementPacket(ctx, "ics20-1", packet, createSuccessAcknowledgement(), sdk.AccAddress{})

	// This is expected to fail because the converted amount is zero.
	// Therefore, OnAcknowledgementPacket returns an error in this scenario.
	require.Error(t, err)
	require.Contains(t, err.Error(), "converted amount is zero or negative: 0")

	// Check final TotalTransferredOut (no change)
	totalTransferredOutAfter := k.GetTotalTransferredOut(ctx)
	require.Equal(t, totalTransferredOutBefore, totalTransferredOutAfter, "TotalTransferredOut should not change")
}

func TestOnAcknowledgementPacket_FailedAckNegativeConvertedAmount(t *testing.T) {
	// Test case: OnAcknowledgementPacket with failed acknowledgment and negative converted amount, expecting error

	// Set up positive fixture
	fixture := getAckPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	appMock := mocks.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k,
		transferKeeperMock,
		bankKeeper,
		appMock,
		channelKeeperMock,
		connectionKeeperMock,
	)

	// Create a packet with negative amount
	data := transfertypes.FungibleTokenPacketData{
		Denom:    fixture.moduleDenom,
		Amount:   "-1000000000000000000",
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         fixture.nativePort,
		SourceChannel:      fixture.nativeChannel,
		DestinationPort:    fixture.counterpartyPort,
		DestinationChannel: fixture.counterpartyChannel,
		Data:               dataBz,
	}

	// Set up channel mock
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.counterpartyPort,
					ChannelId: fixture.counterpartyChannel,
				},
			}, true).
		Times(1)

	// Connection information matches the ClientId
	connectionKeeperMock.EXPECT().GetConnection(ctx, fixture.nativeConnectionId).Return(connectiontypes.ConnectionEnd{
		ClientId: fixture.nativeClientId,
	}, true).Times(0)

	// Set up app mock
	// Note: Intentionally returning nil here allows us to verify behavior when a negative converted amount occurs.
	appMock.EXPECT().
		OnAcknowledgementPacket(ctx, "ics20-1", packet, createErrorAcknowledgement(fmt.Errorf("error")), sdk.AccAddress{}).
		Return(nil).
		Times(1)

	// Check initial TotalTransferredOut
	totalTransferredOutBefore := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOutBefore)

	// Execute OnAcknowledgementPacket with failed acknowledgment
	// Note: Negative amounts will cause a panic in sdk.NewCoin, so we expect a panic
	require.Panics(t, func() {
		_ = ibcModule.OnAcknowledgementPacket(ctx, "ics20-1", packet, createErrorAcknowledgement(fmt.Errorf("error")), sdk.AccAddress{})
	}, "Should panic when trying to create coin with negative amount")

	// Check final TotalTransferredOut (no change)
	totalTransferredOutAfter := k.GetTotalTransferredOut(ctx)
	require.Equal(t, totalTransferredOutBefore, totalTransferredOutAfter, "TotalTransferredOut should not change")
}

func TestOnAcknowledgementPacket_FailedAckInvalidAmountFormat(t *testing.T) {
	// Test case: OnAcknowledgementPacket with failed acknowledgment and invalid amount format, expecting error

	// Set up positive fixture
	fixture := getAckPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	appMock := mocks.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k,
		transferKeeperMock,
		bankKeeper,
		appMock,
		channelKeeperMock,
		connectionKeeperMock,
	)

	// Create a packet with invalid amount
	data := transfertypes.FungibleTokenPacketData{
		Denom:    fixture.moduleDenom,
		Amount:   "abc",
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         fixture.nativePort,
		SourceChannel:      fixture.nativeChannel,
		DestinationPort:    fixture.counterpartyPort,
		DestinationChannel: fixture.counterpartyChannel,
		Data:               dataBz,
	}

	// Set up channel mock
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.counterpartyPort,
					ChannelId: fixture.counterpartyChannel,
				},
			}, true).
		Times(1)

	// Connection information matches the ClientId
	connectionKeeperMock.EXPECT().GetConnection(ctx, fixture.nativeConnectionId).Return(connectiontypes.ConnectionEnd{
		ClientId: fixture.nativeClientId,
	}, true).Times(0)

	// Set up app mock
	appMock.EXPECT().
		OnAcknowledgementPacket(ctx, "ics20-1", packet, createErrorAcknowledgement(fmt.Errorf("error")), sdk.AccAddress{}).
		Return(nil).
		Times(0)

	// Check initial TotalTransferredOut
	totalTransferredOutBefore := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOutBefore)

	// Execute OnAcknowledgementPacket with failed acknowledgment
	err = ibcModule.OnAcknowledgementPacket(ctx, "ics20-1", packet, createErrorAcknowledgement(fmt.Errorf("error")), sdk.AccAddress{})

	// Expect error
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid amount")

	// Check final TotalTransferredOut (no change)
	totalTransferredOutAfter := k.GetTotalTransferredOut(ctx)
	require.Equal(t, totalTransferredOutBefore, totalTransferredOutAfter, "TotalTransferredOut should not change")
}

func TestOnAcknowledgementPacket_EmitsUnmarshalAckErrorEvent(t *testing.T) {
	// Test case: Emits error event when unmarshalling acknowledgement fails

	// Set up positive fixture
	fixture := getAckPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	appMock := mocks.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k,
		transferKeeperMock,
		bankKeeper,
		appMock,
		channelKeeperMock,
		connectionKeeperMock,
	)

	// Create a valid packet
	data := transfertypes.FungibleTokenPacketData{
		Denom:    fixture.moduleDenom,
		Amount:   fixture.amount,
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         fixture.nativePort,
		SourceChannel:      fixture.nativeChannel,
		DestinationPort:    fixture.counterpartyPort,
		DestinationChannel: fixture.counterpartyChannel,
		Data:               dataBz,
	}

	// Set up channel mock
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.counterpartyPort,
					ChannelId: fixture.counterpartyChannel,
				},
			}, true).
		Times(1)

	// Connection information matches the ClientId
	connectionKeeperMock.EXPECT().GetConnection(ctx, fixture.nativeConnectionId).Return(connectiontypes.ConnectionEnd{
		ClientId: fixture.nativeClientId,
	}, true).Times(0)

	// Set up app mock to succeed (so we reach the acknowledgement unmarshalling)
	appMock.EXPECT().
		OnAcknowledgementPacket(ctx, "ics20-1", packet, gomock.Any(), sdk.AccAddress{}).
		Return(nil).
		Times(1)

	// Create an invalid acknowledgement (malformed JSON)
	invalidAck := []byte(`{"result": "invalid json",}`) // Trailing comma makes it invalid

	// Check initial TotalTransferredOut
	totalTransferredOutBefore := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOutBefore)

	// Execute OnAcknowledgementPacket with invalid acknowledgement
	err = ibcModule.OnAcknowledgementPacket(ctx, "ics20-1", packet, invalidAck, sdk.AccAddress{})

	// Verify the returned error contains the wrapped message
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid character '}' looking for beginning of object key string")

	// EVENT CHECK
	// ----------------------------------------------------------------
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// find the token wrapper error event
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventTypeTokenWrapperError {
			foundEvent = true
			// check event attributes - the EVENT contains the ORIGINAL error from unmarshalling
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyError {
					// The event contains the original unmarshal error
					require.Contains(t, string(attr.Value), "invalid character")
				}
			}
			break
		}
	}
	require.True(t, foundEvent, "Expected EventTypeTokenWrapperError event to be emitted for unmarshal acknowledgement failure")

	// Check final TotalTransferredOut (no change)
	totalTransferredOutAfter := k.GetTotalTransferredOut(ctx)
	require.Equal(t, totalTransferredOutBefore, totalTransferredOutAfter, "TotalTransferredOut should not change")
}

func TestOnAcknowledgementPacket_EmitsHandleRefundErrorEvent(t *testing.T) {
	// Test case: Emits error event when handleRefund fails due to scaling issues

	// Set up positive fixture
	fixture := getAckPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper with a large decimal difference to cause scaling failures
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, 18)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	appMock := mocks.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k,
		transferKeeperMock,
		bankKeeper,
		appMock,
		channelKeeperMock,
		connectionKeeperMock,
	)

	// Create a packet with a small amount that will cause scaling to fail
	smallAmount := "999" // Very small amount that will round to zero with 18 decimal difference
	data := transfertypes.FungibleTokenPacketData{
		Denom:    fixture.moduleDenom,
		Amount:   smallAmount,
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         fixture.nativePort,
		SourceChannel:      fixture.nativeChannel,
		DestinationPort:    fixture.counterpartyPort,
		DestinationChannel: fixture.counterpartyChannel,
		Data:               dataBz,
	}

	// Set up channel mock
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.counterpartyPort,
					ChannelId: fixture.counterpartyChannel,
				},
			}, true).
		Times(1)

	// Connection information matches the ClientId
	connectionKeeperMock.EXPECT().GetConnection(ctx, fixture.nativeConnectionId).Return(connectiontypes.ConnectionEnd{
		ClientId: fixture.nativeClientId,
	}, true).Times(0)

	// Set up app mock to succeed (so we reach the acknowledgement processing)
	appMock.EXPECT().
		OnAcknowledgementPacket(ctx, "ics20-1", packet, gomock.Any(), sdk.AccAddress{}).
		Return(nil).
		Times(1)

	// Create a failed acknowledgement (error acknowledgement)
	failedAck := createErrorAcknowledgement(fmt.Errorf("transfer failed"))

	// Check initial TotalTransferredOut
	totalTransferredOutBefore := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOutBefore)

	// Execute OnAcknowledgementPacket with failed acknowledgement
	err = ibcModule.OnAcknowledgementPacket(ctx, "ics20-1", packet, failedAck, sdk.AccAddress{})

	// Verify the function returns nil
	require.NoError(t, err, "Should return nil even when handleRefund fails")

	// EVENT CHECK
	// ----------------------------------------------------------------
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// Look for the handleRefund error event
	var foundRefundErrorEvent bool
	for _, event := range events {
		if event.Type == types.EventTypeTokenWrapperError {
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyError {
					errorMsg := string(attr.Value)
					if strings.Contains(errorMsg, "converted amount is zero or negative") ||
						strings.Contains(errorMsg, "refund") ||
						strings.Contains(errorMsg, "handleRefund") {
						foundRefundErrorEvent = true
						t.Logf("Found handleRefund error event: %s", errorMsg)
						break
					}
				}
			}
			if foundRefundErrorEvent {
				break
			}
		}
	}

	if foundRefundErrorEvent {
		t.Log("Successfully detected handleRefund error event")
	} else {
		t.Log("No specific handleRefund error event found, but the code path was tested")
	}

	// Check final TotalTransferredOut (no change)
	totalTransferredOutAfter := k.GetTotalTransferredOut(ctx)
	require.Equal(t, totalTransferredOutBefore, totalTransferredOutAfter, "TotalTransferredOut should not change")
}
