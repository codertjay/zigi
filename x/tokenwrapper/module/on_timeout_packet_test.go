package tokenwrapper_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"

	keepertest "zigchain/testutil/keeper"
	tokenwrapper "zigchain/x/tokenwrapper/module"
	mocks "zigchain/x/tokenwrapper/testutil"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestOnTimeoutPacket(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	const (
		sender   = "zig1vm3v4yrd3rrwkf3fe8qxutaz27098t76270qc5"
		receiver = "zig1fcf3p6exuhq53myh0k3wweym5ffv00gcnuum07"

		nativePort       = "transfer"
		counterpartyPort = "transfer"

		nativeChannel       = "channel-12"
		counterpartyChannel = "channel-24"

		moduleDenom = "moduledenom"
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
		name           string
		enabled        bool
		channelVersion string
		packet         channeltypes.Packet
		relayer        sdk.AccAddress
		expectedError  error
		setupMocks     func(ctx sdk.Context, keeperMock *mocks.MockTokenwrapperKeeper, transferKeeperMock *mocks.MockTransferKeeper, bankKeeperMock *mocks.MockBankKeeper, appMock *mocks.MockCallbacksCompatibleModule, channelKeeperMock *mocks.MockChannelKeeper)
	}{
		{
			name:           "Happy path - timeout with refund",
			enabled:        true,
			channelVersion: "ics20-1",
			packet:         packet,
			relayer:        sdk.AccAddress{},
			expectedError:  nil,
			setupMocks: func(ctx sdk.Context, keeperMock *mocks.MockTokenwrapperKeeper, transferKeeperMock *mocks.MockTransferKeeper, bankKeeperMock *mocks.MockBankKeeper, appMock *mocks.MockCallbacksCompatibleModule, channelKeeperMock *mocks.MockChannelKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, nativePort, nativeChannel).Return(channeltypes.Channel{
					State: channeltypes.OPEN,
				}, true)
				keeperMock.EXPECT().IsEnabled(ctx).Return(true).Times(1)
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort).Times(2)
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel).Times(2)
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort).Times(2)
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel).Times(2)
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom).Times(2)
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				keeperMock.EXPECT().CheckAccountBalance(ctx, gomock.Any(), gomock.Any()).Return(nil).Times(1)
				keeperMock.EXPECT().CheckModuleBalance(ctx, gomock.Any()).Return(nil).Times(1)
				keeperMock.EXPECT().LockTokens(ctx, gomock.Any(), gomock.Any()).Return(nil).Times(1)
				keeperMock.EXPECT().UnlockTokens(ctx, gomock.Any(), gomock.Any()).Return(nil).Times(1)
				keeperMock.EXPECT().ScaleDownTokenPrecision(ctx, sdkmath.NewInt(1000000000000000000)).Return(sdkmath.NewInt(1000000), nil)
				appMock.EXPECT().OnTimeoutPacket(
					gomock.Any(),
					"ics20-1",
					packet,
					sdk.AccAddress{},
				).Return(nil).Times(1)
			},
		},
		{
			name:           "Error - module is disabled",
			enabled:        true,
			channelVersion: "ics20-1",
			packet:         packet,
			relayer:        sdk.AccAddress{},
			expectedError:  nil, // Should pass through to underlying app
			setupMocks: func(ctx sdk.Context, keeperMock *mocks.MockTokenwrapperKeeper, transferKeeperMock *mocks.MockTransferKeeper, bankKeeperMock *mocks.MockBankKeeper, appMock *mocks.MockCallbacksCompatibleModule, channelKeeperMock *mocks.MockChannelKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, nativePort, nativeChannel).Return(channeltypes.Channel{
					State: channeltypes.OPEN,
				}, true)
				keeperMock.EXPECT().IsEnabled(ctx).Return(false).Times(1)
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort).Times(2)
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel).Times(2)
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort).Times(2)
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel).Times(2)
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom).Times(2)
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				appMock.EXPECT().OnTimeoutPacket(
					gomock.Any(),
					"ics20-1",
					packet,
					sdk.AccAddress{},
				).Return(nil).Times(1)
			},
		},
		{
			name:           "Error - failed to unmarshal packet data",
			enabled:        true,
			channelVersion: "ics20-1",
			packet:         channeltypes.Packet{Data: []byte("invalid")},
			relayer:        sdk.AccAddress{},
			expectedError:  nil, // Should pass through to underlying app
			setupMocks: func(ctx sdk.Context, keeperMock *mocks.MockTokenwrapperKeeper, transferKeeperMock *mocks.MockTransferKeeper, bankKeeperMock *mocks.MockBankKeeper, appMock *mocks.MockCallbacksCompatibleModule, channelKeeperMock *mocks.MockChannelKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, "", "").Return(channeltypes.Channel{
					State: channeltypes.OPEN,
				}, true).Times(1)
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				appMock.EXPECT().OnTimeoutPacket(
					gomock.Any(),
					"ics20-1",
					channeltypes.Packet{Data: []byte("invalid")},
					sdk.AccAddress{},
				).Return(nil).Times(1)
			},
		},
		{
			name:           "Error - app callback error",
			enabled:        true,
			channelVersion: "ics20-1",
			packet:         packet,
			relayer:        sdk.AccAddress{},
			expectedError:  fmt.Errorf("app callback error"),
			setupMocks: func(ctx sdk.Context, keeperMock *mocks.MockTokenwrapperKeeper, transferKeeperMock *mocks.MockTransferKeeper, bankKeeperMock *mocks.MockBankKeeper, appMock *mocks.MockCallbacksCompatibleModule, channelKeeperMock *mocks.MockChannelKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, nativePort, nativeChannel).Return(channeltypes.Channel{
					State: channeltypes.OPEN,
				}, true).Times(1)
				keeperMock.EXPECT().IsEnabled(ctx).Return(true).Times(1)
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort).Times(2)
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel).Times(2)
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort).Times(2)
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel).Times(2)
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom).Times(2)
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				appMock.EXPECT().OnTimeoutPacket(
					gomock.Any(),
					"ics20-1",
					packet,
					sdk.AccAddress{},
				).Return(fmt.Errorf("app callback error")).Times(1)
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
			tt.setupMocks(ctx, keeperMock, transferKeeperMock, bankKeeperMock, appMock, channelKeeperMock)

			// Call the function
			err := ibcModule.OnTimeoutPacket(
				ctx,
				tt.channelVersion,
				tt.packet,
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
// TimeoutPacketFixture defines the configuration for OnTimeoutPacket tests
type TimeoutPacketFixture struct {
	nativePort           string
	nativeChannel        string
	counterpartyPort     string
	counterpartyChannel  string
	nativeConnectionId   string
	counterpartyClientId string
	moduleDenom          string
	ibcDenom             string
	amount               string
	sender               string
	receiver             string
	decimalDifference    uint32
}

// getTimeoutPacketPositiveFixture returns a positive fixture for OnTimeoutPacket tests
func getTimeoutPacketPositiveFixture() TimeoutPacketFixture {
	return TimeoutPacketFixture{
		nativePort:           "transfer",
		nativeChannel:        "channel-12",
		counterpartyPort:     "transfer",
		counterpartyChannel:  "channel-24",
		nativeConnectionId:   "connection-0",
		counterpartyClientId: "07-tendermint-0",
		moduleDenom:          "moduledenom",
		ibcDenom:             "ibc/AA11A7781887F73EDC5A9BA3191E75FA0857A643D560B2C2A3A1868DA4D7AD97",
		amount:               "1000000000000000000", // 1e18 (18 decimals)
		sender:               "zig1vm3v4yrd3rrwkf3fe8qxutaz27098t76270qc5",
		receiver:             "zig1fcf3p6exuhq53myh0k3wweym5ffv00gcnuum07",
		decimalDifference:    12, // Converts 18 decimals to 6 decimals (1e12 factor)
	}
}

// Positive test cases

func TestOnTimeoutPacket_RealKeepers_Positive(t *testing.T) {
	// Test case: OnTimeoutPacket with real keepers, valid packet, expecting success and refund

	// Set up positive fixture
	fixture := getTimeoutPacketPositiveFixture()

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

	// Add balance to the sender account for the IBC token
	senderAddr, err := sdk.AccAddressFromBech32(fixture.sender)
	require.NoError(t, err)

	// Mint IBC tokens to the sender account
	amount, ok := sdkmath.NewIntFromString(fixture.amount)
	require.True(t, ok, "failed to parse amount")
	ibcCoins := sdk.NewCoins(sdk.NewCoin(fixture.ibcDenom, amount))
	require.NoError(t, bankKeeper.MintCoins(ctx, "mint", ibcCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, "mint", senderAddr, ibcCoins))

	// Fund the module account with native tokens for the refund
	// The refund converts from 18 decimals to 6 decimals, so we need to fund with the converted amount
	conversionFactor := sdkmath.NewInt(1000000000000) // 1e12
	convertedAmount := amount.Quo(conversionFactor)
	nativeCoins := sdk.NewCoins(sdk.NewCoin("uzig", convertedAmount))
	require.NoError(t, bankKeeper.MintCoins(ctx, "mint", nativeCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToModule(ctx, "mint", "tokenwrapper", nativeCoins))

	// Set up app mock
	appMock.EXPECT().
		OnTimeoutPacket(ctx, "ics20-1", packet, sdk.AccAddress{}).
		Return(nil).
		Times(1)

	// Execute OnTimeoutPacket
	err = ibcModule.OnTimeoutPacket(ctx, "ics20-1", packet, sdk.AccAddress{})

	// Expect no error
	require.NoError(t, err)
}

func TestOnTimeoutPacket_MissingNativePortChannelDenom(t *testing.T) {
	// Test case: OnTimeoutPacket with missing native port, channel and denom, expecting no error and pass through

	// Set up positive fixture
	fixture := getTimeoutPacketPositiveFixture()

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
		OnTimeoutPacket(ctx, "ics20-1", packet, sdk.AccAddress{}).
		Return(nil).
		Times(1)

	// Execute OnTimeoutPacket
	err = ibcModule.OnTimeoutPacket(ctx, "ics20-1", packet, sdk.AccAddress{})

	// Expect no error
	require.NoError(t, err)
}

func TestOnTimeoutPacket_MismatchedSourcePort(t *testing.T) {
	// Test case: OnTimeoutPacket with mismatched source port, expecting no error and pass through

	// Set up positive fixture
	fixture := getTimeoutPacketPositiveFixture()

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
		OnTimeoutPacket(ctx, "ics20-1", packet, sdk.AccAddress{}).
		Return(nil).
		Times(1)

	// Execute OnTimeoutPacket
	err = ibcModule.OnTimeoutPacket(ctx, "ics20-1", packet, sdk.AccAddress{})

	// Expect no error
	require.NoError(t, err)
}

func TestOnTimeoutPacket_MismatchedSourceChannel(t *testing.T) {
	// Test case: OnTimeoutPacket with a mismatched source channel, expecting no error and pass through

	// Set up positive fixture
	fixture := getTimeoutPacketPositiveFixture()

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
		OnTimeoutPacket(ctx, "ics20-1", packet, sdk.AccAddress{}).
		Return(nil).
		Times(1)

	// Execute OnTimeoutPacket
	err = ibcModule.OnTimeoutPacket(ctx, "ics20-1", packet, sdk.AccAddress{})

	// Expect no error
	require.NoError(t, err)
}

// Negative test cases

func TestOnTimeoutPacket_ZeroAmount(t *testing.T) {
	// Test case: OnTimeoutPacket with zero amount, expecting error and pass through

	// Set up positive fixture
	fixture := getTimeoutPacketPositiveFixture()

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

	// Set up app mock
	appMock.EXPECT().
		OnTimeoutPacket(ctx, "ics20-1", packet, sdk.AccAddress{}).
		Return(nil).
		Times(1)

	// Execute OnTimeoutPacket
	err = ibcModule.OnTimeoutPacket(ctx, "ics20-1", packet, sdk.AccAddress{})

	// Expect no error (should pass through)
	require.NoError(t, err)
}

func TestOnTimeoutPacket_NegativeAmount(t *testing.T) {
	// Test case: OnTimeoutPacket with negative amount, expecting error and pass through

	// Set up positive fixture
	fixture := getTimeoutPacketPositiveFixture()

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

	// Set up app mock
	appMock.EXPECT().
		OnTimeoutPacket(ctx, "ics20-1", packet, sdk.AccAddress{}).
		Return(nil).
		Times(1)

	// Execute OnTimeoutPacket
	err = ibcModule.OnTimeoutPacket(ctx, "ics20-1", packet, sdk.AccAddress{})

	// Expect no error (should pass through)
	require.NoError(t, err)
}

func TestOnTimeoutPacket_InvalidAmountFormat(t *testing.T) {
	// Test case: OnTimeoutPacket with invalid amount format, expecting error and pass through

	// Set up positive fixture
	fixture := getTimeoutPacketPositiveFixture()

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

	// Set up app mock
	appMock.EXPECT().
		OnTimeoutPacket(ctx, "ics20-1", packet, sdk.AccAddress{}).
		Return(nil).
		Times(1)

	// Execute OnTimeoutPacket
	err = ibcModule.OnTimeoutPacket(ctx, "ics20-1", packet, sdk.AccAddress{})

	// Expect no error (should pass through)
	require.NoError(t, err)
}

func TestOnTimeoutPacket_ChannelNotFound(t *testing.T) {
	// Test case: OnTimeoutPacket with a channel not found, expecting error and pass through

	// Set up positive fixture
	fixture := getTimeoutPacketPositiveFixture()

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

	// Set up app mock
	appMock.EXPECT().
		OnTimeoutPacket(ctx, "ics20-1", packet, sdk.AccAddress{}).
		Return(nil).
		Times(1)

	// Execute OnTimeoutPacket
	err = ibcModule.OnTimeoutPacket(ctx, "ics20-1", packet, sdk.AccAddress{})

	// Expect no error (should pass through)
	require.NoError(t, err)
}

func TestOnTimeoutPacket_ChannelNotOpen(t *testing.T) {
	// Test case: OnTimeoutPacket with channel not open, expecting error and pass through

	// Set up positive fixture
	fixture := getTimeoutPacketPositiveFixture()

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

	// Set up app mock
	appMock.EXPECT().
		OnTimeoutPacket(ctx, "ics20-1", packet, sdk.AccAddress{}).
		Return(nil).
		Times(1)

	// Execute OnTimeoutPacket
	err = ibcModule.OnTimeoutPacket(ctx, "ics20-1", packet, sdk.AccAddress{})

	// Expect no error (should pass through)
	require.NoError(t, err)
}

func TestOnTimeoutPacket_InvalidSenderAddress(t *testing.T) {
	// Test case: OnTimeoutPacket with an invalid sender address, expecting error and pass through

	fixture := getTimeoutPacketPositiveFixture()

	// Use invalid Bech32 sender
	invalidSender := "not-a-valid-bech32"

	// Create keeper with bank
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	appMock := mocks.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	ibcModule := tokenwrapper.NewIBCModule(
		k,
		transferKeeperMock,
		bankKeeper,
		appMock,
		channelKeeperMock,
		connectionKeeperMock,
	)

	// Create packet with invalid sender
	data := transfertypes.FungibleTokenPacketData{
		Denom:    fixture.moduleDenom,
		Amount:   fixture.amount,
		Sender:   invalidSender,
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

	// Expect fallback call to app.OnTimeoutPacket
	appMock.EXPECT().
		OnTimeoutPacket(ctx, "ics20-1", packet, sdk.AccAddress{}).
		Return(nil).
		Times(1)

	err = ibcModule.OnTimeoutPacket(ctx, "ics20-1", packet, sdk.AccAddress{})
	require.NoError(t, err)

	require.NoError(t, err)

	// Check that an error event was emitted
	events := ctx.EventManager().Events()
	require.Equal(t, 1, len(events))
	require.Equal(t, "tokenwrapper_error", events[0].Type)
	require.Equal(t, 1, len(events[0].Attributes))

	msg := string(events[0].Attributes[0].Value)
	require.Contains(t, msg, "decoding bech32 failed: invalid separator index -1")
}

func TestOnTimeoutPacket_PacketDenomNotModuleDenom(t *testing.T) {
	// Test case: OnTimeoutPacket with packet denom not matching module denom, expecting fallback to app.OnTimeoutPacket

	fixture := getTimeoutPacketPositiveFixture()

	// Override denom so it will not match module denom
	packetDenom := "ibc/OTHER_IBC_HASH"
	baseDenom := transfertypes.ExtractDenomFromPath(packetDenom).Base
	require.NotEqual(t, fixture.moduleDenom, baseDenom)

	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom) // Different from baseDenom
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	appMock := mocks.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	ibcModule := tokenwrapper.NewIBCModule(
		k,
		transferKeeperMock,
		bankKeeper,
		appMock,
		channelKeeperMock,
		connectionKeeperMock,
	)

	// Packet with mismatching denom
	data := transfertypes.FungibleTokenPacketData{
		Denom:    packetDenom,
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

	// Expect fallback call to app.OnTimeoutPacket
	appMock.EXPECT().
		OnTimeoutPacket(ctx, "ics20-1", packet, sdk.AccAddress{}).
		Return(nil).
		Times(1)

	err = ibcModule.OnTimeoutPacket(ctx, "ics20-1", packet, sdk.AccAddress{})
	require.NoError(t, err)

	// Check that an info event was emitted
	events := ctx.EventManager().Events()
	require.Equal(t, 1, len(events))
	require.Equal(t, "tokenwrapper_info", events[0].Type)
	require.Equal(t, 1, len(events[0].Attributes))
	require.Equal(t, "packet denom is not the module denom, skipping refunding: ibc/OTHER_IBC_HASH", string(events[0].Attributes[0].Value))
}
