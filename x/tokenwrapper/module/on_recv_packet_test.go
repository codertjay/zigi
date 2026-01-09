package tokenwrapper_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	tokenwrapper "zigchain/x/tokenwrapper/module"
	"zigchain/x/tokenwrapper/testutil"
	"zigchain/x/tokenwrapper/types"
	"zigchain/zutils/constants"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"

	keepertest "zigchain/testutil/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	connectiontypes "github.com/cosmos/ibc-go/v10/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// Test fixture with positive scenario data
type OnRecvPacketFixture struct {
	faucet                   string
	sender                   string
	receiver                 string
	nativePort               string
	counterpartyPort         string
	nativeChannel            string
	counterpartyChannel      string
	nativeConnectionId       string
	counterpartyConnectionId string
	nativeClientId           string
	counterpartyClientId     string
	moduleDenom              string
	ibcDenom                 string
	amount                   string
	decimalDifference        uint32
}

func getRecvPacketPositiveFixture() OnRecvPacketFixture {

	return OnRecvPacketFixture{
		faucet:                   "zig15yk64u7zc9g9k2yr2wmzeva5qgwxps6y8c2amk",
		sender:                   "axelar15yk64u7zc9g9k2yr2wmzeva5qgwxps6yzue8jl",
		receiver:                 "zig1fcf3p6exuhq53myh0k3wweym5ffv00gcnuum07",
		nativePort:               "transfer",
		counterpartyPort:         "transfer",
		nativeChannel:            "channel-12",
		counterpartyChannel:      "channel-24",
		nativeConnectionId:       "connection-12",
		counterpartyConnectionId: "connection-24",
		nativeClientId:           "07-tendermint-12",
		counterpartyClientId:     "07-tendermint-24",
		moduleDenom:              "axl",
		ibcDenom:                 "ibc/148AEF32AA7274DC6AFD912A5C1478AC10246B8AEE1C8DEA6D831B752000E89F",
		amount:                   "1000000000000000000", // 1 token with 18 decimals
		decimalDifference:        12,
	}
}

func TestOnRecvPacket_Positive(t *testing.T) {
	// Test case: OnRecvPacket successfully no mocking the TokenwrapperKeeper
	// but mocking the TransferKeeper, BankKeeper, and other dependencies

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// set up positive fixture
	fixture := getRecvPacketPositiveFixture()

	// Create mocks
	transferKeeperMock := testutil.NewMockTransferKeeper(ctrl)
	bankKeeperMock := testutil.NewMockBankKeeper(ctrl)
	appMock := testutil.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := testutil.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := testutil.NewMockConnectionKeeper(ctrl)

	// Create the TokenwrapperKeeper with the mocks
	k, ctx := keepertest.TokenwrapperKeeper(t, bankKeeperMock)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k, transferKeeperMock, bankKeeperMock, appMock, channelKeeperMock, connectionKeeperMock,
	)

	// Create the data and the packet
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
		SourcePort:         fixture.counterpartyPort,
		SourceChannel:      fixture.counterpartyChannel,
		DestinationPort:    fixture.nativePort,
		DestinationChannel: fixture.nativeChannel,
		Data:               dataBz,
	}

	// code will check that the channel is valid (open and connected)
	channelKeeperMock.
		EXPECT().
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
		Times(3)

	// code will check that the connection is valid and matches the ClientId
	connectionKeeperMock.
		EXPECT().
		GetConnection(ctx, fixture.nativeConnectionId).
		Return(
			connectiontypes.ConnectionEnd{
				ClientId: fixture.nativeClientId,
				Counterparty: connectiontypes.Counterparty{
					ClientId: fixture.counterpartyClientId,
				},
			}, true).
		Times(1)

	// code will check that App callback is called and successful
	appMock.
		EXPECT().
		OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).
		Return(channeltypes.NewResultAcknowledgement([]byte("success")))

	// code will check if the receiver has the ibc denom and also send it to the module
	amountInt, _ := sdkmath.NewIntFromString(fixture.amount)

	bankKeeperMock.
		EXPECT().
		HasBalance(ctx, sdk.MustAccAddressFromBech32(fixture.receiver), sdk.NewCoin(fixture.ibcDenom, amountInt)).
		Return(true).
		Times(2)

	bankKeeperMock.
		EXPECT().
		SendCoinsFromAccountToModule(ctx, sdk.MustAccAddressFromBech32(fixture.receiver), types.ModuleName, sdk.NewCoins(sdk.NewCoin(fixture.ibcDenom, amountInt))).
		Return(nil)

	// code will check if the module has the bond denom and also send it to the receiver
	bankKeeperMock.
		EXPECT().
		HasBalance(ctx, authtypes.NewModuleAddress(types.ModuleName), sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(1000000))).
		Return(true).
		Times(2)

	bankKeeperMock.
		EXPECT().
		SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.MustAccAddressFromBech32(fixture.receiver), sdk.NewCoins(sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(1000000)))).
		Return(nil)

	// Lock IBC tokens
	amount, _ := sdkmath.NewIntFromString(fixture.amount)

	// Converted amount to native tokens
	conversionFactor := k.GetDecimalConversionFactor(ctx)
	convertedAmount := amount.Quo(conversionFactor)
	require.True(t, convertedAmount.IsPositive(), "converted amount should be positive")

	// check TotalTransferredIn before the packet is received
	totalTransferredIn := k.GetTotalTransferredIn(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredIn,
		"TotalTransferredIn should be zero before OnRecvPacket")

	// check TotalTransferredOut before the packet is received
	totalTransferredOut := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOut,
		"TotalTransferredOut should be zero before OnRecvPacket")

	// Execute
	respOnRecvPacket := ibcModule.OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{})

	// Validate that no error occurred
	require.NotNil(t, respOnRecvPacket)
	require.True(t, respOnRecvPacket.Success())

	// Check sent amount has been added to TotalTransferredIn
	totalTransferredIn = k.GetTotalTransferredIn(ctx)
	require.Equal(t, convertedAmount, totalTransferredIn)

	// Check TotalTransferredOut is still zero
	totalTransferredOut = k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOut, "TotalTransferredOut should be zero after successful OnRecvPacket")

	// Check the acknowledgement event
	events := ctx.EventManager().Events()
	require.Equal(t, 1, len(events))
	require.Equal(t, types.EventTypeTokenWrapperOnRecvPacket, events[0].Type)
	// sender
	require.Equal(t, types.AttributeKeySender, events[0].Attributes[0].Key)
	require.Equal(t, fixture.sender, events[0].Attributes[0].Value)
	// receiver
	require.Equal(t, types.AttributeKeyReceiver, events[0].Attributes[1].Key)
	require.Equal(t, fixture.receiver, events[0].Attributes[1].Value)
	// amount
	require.Equal(t, types.AttributeKeyAmount, events[0].Attributes[2].Key)
	require.Equal(t, convertedAmount.String(), events[0].Attributes[2].Value)
	// denom
	require.Equal(t, types.AttributeKeyDenom, events[0].Attributes[3].Key)
	require.Equal(t, constants.BondDenom, events[0].Attributes[3].Value)
	// source port
	require.Equal(t, types.AttributeKeySourcePort, events[0].Attributes[4].Key)
	require.Equal(t, fixture.counterpartyPort, events[0].Attributes[4].Value)
	// source channel.
	require.Equal(t, types.AttributeKeySourceChannel, events[0].Attributes[5].Key)
	require.Equal(t, fixture.counterpartyChannel, events[0].Attributes[5].Value)
	// dest port
	require.Equal(t, types.AttributeKeyDestPort, events[0].Attributes[6].Key)
	require.Equal(t, fixture.nativePort, events[0].Attributes[6].Value)
	// dest channel
	require.Equal(t, types.AttributeKeyDestChannel, events[0].Attributes[7].Key)
	require.Equal(t, fixture.nativeChannel, events[0].Attributes[7].Value)
}

func TestOnRecvPacket_ChannelOff(t *testing.T) {
	// Test case: OnRecvPacket negative test because the channel is off
	// no mocking the TokenwrapperKeeper
	// but mocking the TransferKeeper, BankKeeper, and other dependencies

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// set up positive fixture
	fixture := getRecvPacketPositiveFixture()

	// Create mocks
	// keeperMock := testutil.NewMockTokenwrapperKeeper(ctrl)
	transferKeeperMock := testutil.NewMockTransferKeeper(ctrl)
	bankKeeperMock := testutil.NewMockBankKeeper(ctrl)
	appMock := testutil.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := testutil.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := testutil.NewMockConnectionKeeper(ctrl)

	// Create the TokenwrapperKeeper with the mocks
	k, ctx := keepertest.TokenwrapperKeeper(t, bankKeeperMock)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k, transferKeeperMock, bankKeeperMock, appMock, channelKeeperMock, connectionKeeperMock,
	)

	// Create the data and the packet
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
		SourcePort:         fixture.counterpartyPort,
		SourceChannel:      fixture.counterpartyChannel,
		DestinationPort:    fixture.nativePort,
		DestinationChannel: fixture.nativeChannel,
		Data:               dataBz,
	}

	// code will check that the channel is valid (open and connected)
	channelKeeperMock.
		EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.counterpartyPort,
					ChannelId: fixture.counterpartyChannel,
				},
			}, false).
		Times(1)

	// check TotalTransferredIn before the packet is received
	totalTransferredIn := k.GetTotalTransferredIn(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredIn,
		"TotalTransferredIn should be zero before OnRecvPacket")

	// check TotalTransferredOut before the packet is received
	totalTransferredOut := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOut,
		"TotalTransferredOut should be zero before OnRecvPacket")

	// Execute
	ack := ibcModule.OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{})

	// Validate that an error occurred
	require.NotNil(t, ack)
	require.False(t, ack.Success(), "OnRecvPacket should fail when channel is not open")

	// Verify it's an error acknowledgement and check the specific error message
	errAck, _ := ack.(channeltypes.Acknowledgement)
	expectedErr := "ABCI code: 1503: error handling packet: see events for details"
	actualErr := string(errAck.Acknowledgement())
	require.Contains(t, actualErr, expectedErr)

	// Check that totals have remained unchanged since processing failed
	totalTransferredInAfter := k.GetTotalTransferredIn(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredInAfter,
		"TotalTransferredIn should remain zero after failed OnRecvPacket")

	totalTransferredOutAfter := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOutAfter,
		"TotalTransferredOut should remain zero after failed OnRecvPacket")

	// Check that an error event was emitted
	events := ctx.EventManager().Events()
	require.Equal(t, 1, len(events))
	require.Equal(t, types.EventTypeTokenWrapperError, events[0].Type)

	// Check the error message in the event
	require.Contains(t, events[0].Attributes[0].Value, fmt.Sprintf("channel validation failed: %s/%s: channel not found", fixture.nativePort, fixture.nativeChannel))
}

func TestOnRecvPacket_PacketNotMatchChain(t *testing.T) {
	// Test case: OnRecvPacket the packet information doesn't match the chain information
	// no mocking the TokenwrapperKeeper
	// but mocking the TransferKeeper, BankKeeper, and other dependencies

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// set up positive fixture
	fixture := getRecvPacketPositiveFixture()

	// Create mocks
	transferKeeperMock := testutil.NewMockTransferKeeper(ctrl)
	bankKeeperMock := testutil.NewMockBankKeeper(ctrl)
	appMock := testutil.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := testutil.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := testutil.NewMockConnectionKeeper(ctrl)

	// Create the TokenwrapperKeeper with the mocks
	k, ctx := keepertest.TokenwrapperKeeper(t, bankKeeperMock)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k, transferKeeperMock, bankKeeperMock, appMock, channelKeeperMock, connectionKeeperMock,
	)

	// Create the data and the packet
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
		SourcePort:         fixture.counterpartyPort,
		SourceChannel:      fixture.counterpartyChannel,
		DestinationPort:    fixture.nativePort,
		DestinationChannel: fixture.nativeChannel,
		Data:               dataBz,
	}

	// code will check that the channel is valid (open and connected)
	channelKeeperMock.
		EXPECT().
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
		Times(3)

	// code will check that the connection is valid and matches the ClientId
	connectionKeeperMock.
		EXPECT().
		GetConnection(ctx, fixture.nativeConnectionId).
		Return(connectiontypes.ConnectionEnd{ClientId: "otherblockchain-1"}, true).
		Times(1)

	appMock.
		EXPECT().
		OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).
		Return(channeltypes.NewResultAcknowledgement([]byte("success")))

	// check TotalTransferredIn before the packet is received
	totalTransferredIn := k.GetTotalTransferredIn(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredIn,
		"TotalTransferredIn should be zero before OnRecvPacket")

	// check TotalTransferredOut before the packet is received
	totalTransferredOut := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOut,
		"TotalTransferredOut should be zero before OnRecvPacket")

	// Execute
	ack := ibcModule.OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{})

	// Validate that an error occurred
	require.NotNil(t, ack)
	require.True(t, ack.Success(), "OnRecvPacket should still succeed when client id is invalid")

	// Check that totals have remained unchanged since processing failed
	totalTransferredInAfter := k.GetTotalTransferredIn(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredInAfter,
		"TotalTransferredIn should remain zero after failed OnRecvPacket")

	totalTransferredOutAfter := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOutAfter,
		"TotalTransferredOut should remain zero after failed OnRecvPacket")

	// Check that an error event was emitted
	events := ctx.EventManager().Events()
	require.Equal(t, 1, len(events))
	require.Equal(t, types.EventTypeTokenWrapperError, events[0].Type)

	// Check the error message in the event
	require.Contains(t, events[0].Attributes[0].Value, fmt.Sprintf("client state validation failed: native client ID mismatch: expected %s, got %s", fixture.nativeClientId, "otherblockchain-1"))
}

func TestOnRecvPacket_RealKeepers_Positive(t *testing.T) {
	// Test case: OnRecvPacket with real keepers to verify balance changes

	// set up a positive fixture
	fixture := getRecvPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := testutil.NewMockTransferKeeper(ctrl)
	appMock := testutil.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := testutil.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := testutil.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k, transferKeeperMock, nil, appMock, channelKeeperMock, connectionKeeperMock,
	)

	// Create the data and the packet
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
		SourcePort:         fixture.counterpartyPort,
		SourceChannel:      fixture.counterpartyChannel,
		DestinationPort:    fixture.nativePort,
		DestinationChannel: fixture.nativeChannel,
		Data:               dataBz,
	}

	// Set up channel and connection mocks
	channelKeeperMock.
		EXPECT().
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
		Times(3)

	connectionKeeperMock.
		EXPECT().
		GetConnection(ctx, fixture.nativeConnectionId).
		Return(
			connectiontypes.ConnectionEnd{
				ClientId: fixture.nativeClientId,
				Counterparty: connectiontypes.Counterparty{
					ClientId: fixture.counterpartyClientId,
				},
			}, true).
		Times(1)

	appMock.
		EXPECT().
		OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).
		Return(channeltypes.NewResultAcknowledgement([]byte("success")))

	// Set up initial balances
	receiverAddr := sdk.MustAccAddressFromBech32(fixture.receiver)
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)

	// Mint IBC tokens to receiver
	amountInt, _ := sdkmath.NewIntFromString(fixture.amount)
	ibcCoins := sdk.NewCoins(sdk.NewCoin(fixture.ibcDenom, amountInt))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, ibcCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, receiverAddr, ibcCoins))

	// Mint native tokens to module
	conversionFactor := k.GetDecimalConversionFactor(ctx)
	convertedAmount := amountInt.Quo(conversionFactor)
	nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, convertedAmount))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, nativeCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, nativeCoins))

	// Check initial balances
	receiverNativeBalanceBefore := bankKeeper.GetBalance(ctx, receiverAddr, constants.BondDenom)
	require.Equal(t, sdkmath.ZeroInt(), receiverNativeBalanceBefore.Amount)

	receiverIBCBalanceBefore := bankKeeper.GetBalance(ctx, receiverAddr, fixture.ibcDenom)
	require.Equal(t, amountInt, receiverIBCBalanceBefore.Amount)

	moduleNativeBalanceBefore := bankKeeper.GetBalance(ctx, moduleAddr, constants.BondDenom)
	require.Equal(t, convertedAmount, moduleNativeBalanceBefore.Amount)

	moduleIBCBalanceBefore := bankKeeper.GetBalance(ctx, moduleAddr, fixture.ibcDenom)
	require.Equal(t, sdkmath.ZeroInt(), moduleIBCBalanceBefore.Amount)
	// Execute OnRecvPacket
	respOnRecvPacket := ibcModule.OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{})

	// Validate that no error occurred
	require.NotNil(t, respOnRecvPacket)
	require.True(t, respOnRecvPacket.Success())

	// Check final balances
	receiverNativeBalanceAfter := bankKeeper.GetBalance(ctx, receiverAddr, constants.BondDenom)
	require.Equal(t, convertedAmount, receiverNativeBalanceAfter.Amount, "Receiver should have received native tokens")

	receiverIBCBalanceAfter := bankKeeper.GetBalance(ctx, receiverAddr, fixture.ibcDenom)
	require.Equal(t, sdkmath.ZeroInt(), receiverIBCBalanceAfter.Amount, "Receiver should have no IBC tokens after wrapping")

	moduleNativeBalanceAfter := bankKeeper.GetBalance(ctx, moduleAddr, constants.BondDenom)
	require.Equal(t, sdkmath.ZeroInt(), moduleNativeBalanceAfter.Amount, "Module should have no native tokens after wrapping")

	moduleIBCBalanceAfter := bankKeeper.GetBalance(ctx, moduleAddr, fixture.ibcDenom)
	require.Equal(t, amountInt, moduleIBCBalanceAfter.Amount, "Module should have no IBC tokens after wrapping")

	// Check TotalTransferredIn
	totalTransferredIn := k.GetTotalTransferredIn(ctx)
	require.Equal(t, convertedAmount, totalTransferredIn, "TotalTransferredIn should match converted amount")

	// Check TotalTransferredOut
	totalTransferredOut := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOut, "TotalTransferredOut should be zero after successful OnRecvPacket")

	// Check the OnRecvPacket event
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// Find the OnRecvPacket event
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventTypeTokenWrapperOnRecvPacket {
			foundEvent = true
			// Check event attributes
			for _, attr := range event.Attributes {
				switch attr.Key {
				case types.AttributeKeySender:
					require.Equal(t, fixture.sender, attr.Value)
				case types.AttributeKeyReceiver:
					require.Equal(t, fixture.receiver, attr.Value)
				case types.AttributeKeyAmount:
					require.Equal(t, convertedAmount.String(), attr.Value)
				case types.AttributeKeyDenom:
					require.Equal(t, constants.BondDenom, attr.Value)
				case types.AttributeKeySourcePort:
					require.Equal(t, fixture.counterpartyPort, attr.Value)
				case types.AttributeKeySourceChannel:
					require.Equal(t, fixture.counterpartyChannel, attr.Value)
				case types.AttributeKeyDestPort:
					require.Equal(t, fixture.nativePort, attr.Value)
				case types.AttributeKeyDestChannel:
					require.Equal(t, fixture.nativeChannel, attr.Value)
				}
			}
			break
		}
	}

	require.True(t, foundEvent, "Expected EventTypeTokenWrapperOnRecvPacket event to be emitted")
}

func TestOnRecvPacket_RealKeepers_Positive_NoIBCSettings(t *testing.T) {
	// Test case: OnRecvPacket with real keepers to verify balance no change when one of the IBC settings are set

	// set up positive fixture
	fixture := getRecvPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	// Native channel is not set, so it should not be used
	// k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := testutil.NewMockTransferKeeper(ctrl)
	appMock := testutil.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := testutil.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := testutil.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k, transferKeeperMock, nil, appMock, channelKeeperMock, connectionKeeperMock,
	)

	// Create the data and the packet
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
		SourcePort:         fixture.counterpartyPort,
		SourceChannel:      fixture.counterpartyChannel,
		DestinationPort:    fixture.nativePort,
		DestinationChannel: fixture.nativeChannel,
		Data:               dataBz,
	}

	// Set up channel and connection mocks
	channelKeeperMock.
		EXPECT().
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
		Times(3)

	connectionKeeperMock.
		EXPECT().
		GetConnection(ctx, fixture.nativeConnectionId).
		Return(
			connectiontypes.ConnectionEnd{
				ClientId: fixture.nativeClientId,
				Counterparty: connectiontypes.Counterparty{
					ClientId: fixture.counterpartyClientId,
				},
			}, true).
		Times(1)

	appMock.
		EXPECT().
		OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).
		Return(channeltypes.NewResultAcknowledgement([]byte("success")))

	// Set up initial balances
	receiverAddr := sdk.MustAccAddressFromBech32(fixture.receiver)
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)

	// Mint IBC tokens to receiver
	amountInt, _ := sdkmath.NewIntFromString(fixture.amount)
	ibcCoins := sdk.NewCoins(sdk.NewCoin(fixture.ibcDenom, amountInt))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, ibcCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, receiverAddr, ibcCoins))

	// Mint native tokens to module
	conversionFactor := k.GetDecimalConversionFactor(ctx)
	convertedAmount := amountInt.Quo(conversionFactor)
	nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, convertedAmount))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, nativeCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, nativeCoins))

	// Check initial balances
	receiverBalanceBefore := bankKeeper.GetBalance(ctx, receiverAddr, fixture.ibcDenom)
	require.Equal(t, amountInt, receiverBalanceBefore.Amount)

	moduleBalanceBefore := bankKeeper.GetBalance(ctx, moduleAddr, constants.BondDenom)
	require.Equal(t, convertedAmount, moduleBalanceBefore.Amount)

	// Execute OnRecvPacket
	respOnRecvPacket := ibcModule.OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{})

	// Validate that no error occurred
	require.NotNil(t, respOnRecvPacket)
	require.True(t, respOnRecvPacket.Success())

	// Check final balances
	receiverNativeBalanceAfter := bankKeeper.GetBalance(ctx, receiverAddr, constants.BondDenom)
	require.Equal(t, sdkmath.ZeroInt(), receiverNativeBalanceAfter.Amount, "Receiver shouldn't have received native tokens as no wrapping occurred")

	receiverIBCBalanceAfter := bankKeeper.GetBalance(ctx, receiverAddr, fixture.ibcDenom)
	require.Equal(t, amountInt, receiverIBCBalanceAfter.Amount, "Receiver have the IBC tokens as no wrapping occurred")

	moduleBondBalanceAfter := bankKeeper.GetBalance(ctx, moduleAddr, constants.BondDenom)
	require.Equal(t, convertedAmount, moduleBondBalanceAfter.Amount, "Module should have the native tokens as no wrapping occurred")

	moduleIBCBalanceAfter := bankKeeper.GetBalance(ctx, moduleAddr, fixture.ibcDenom)
	require.Equal(t, sdkmath.ZeroInt(), moduleIBCBalanceAfter.Amount, "Module shouldn't have the IBC tokens as no wrapping occurred")

	// Check TotalTransferredIn
	totalTransferredIn := k.GetTotalTransferredIn(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredIn, "TotalTransferredIn should be zero as no wrapping occurred")

	// Check TotalTransferredOut
	totalTransferredOut := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOut, "TotalTransferredOut should be zero after successful OnRecvPacket")
}

func TestOnRecvPacket_RealKeepers_Positive_DifferentDenom(t *testing.T) {
	// Test case: OnRecvPacket with real keepers to verify balance no change when the denom sent is different from the module denom

	// set up positive fixture
	fixture := getRecvPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, "diff-denom")
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := testutil.NewMockTransferKeeper(ctrl)
	appMock := testutil.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := testutil.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := testutil.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k, transferKeeperMock, nil, appMock, channelKeeperMock, connectionKeeperMock,
	)

	// Create the data and the packet
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
		SourcePort:         fixture.counterpartyPort,
		SourceChannel:      fixture.counterpartyChannel,
		DestinationPort:    fixture.nativePort,
		DestinationChannel: fixture.nativeChannel,
		Data:               dataBz,
	}

	// Set up channel and connection mocks
	channelKeeperMock.
		EXPECT().
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
		Times(3)

	connectionKeeperMock.
		EXPECT().
		GetConnection(ctx, fixture.nativeConnectionId).
		Return(
			connectiontypes.ConnectionEnd{
				ClientId: fixture.nativeClientId,
				Counterparty: connectiontypes.Counterparty{
					ClientId: fixture.counterpartyClientId,
				},
			}, true).
		Times(1)

	appMock.
		EXPECT().
		OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).
		Return(channeltypes.NewResultAcknowledgement([]byte("success")))

	// Set up initial balances
	receiverAddr := sdk.MustAccAddressFromBech32(fixture.receiver)
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)

	// Mint IBC tokens to receiver
	amountInt, _ := sdkmath.NewIntFromString(fixture.amount)
	ibcCoins := sdk.NewCoins(sdk.NewCoin(fixture.ibcDenom, amountInt))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, ibcCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, receiverAddr, ibcCoins))

	// Mint native tokens to module
	conversionFactor := k.GetDecimalConversionFactor(ctx)
	convertedAmount := amountInt.Quo(conversionFactor)
	nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, convertedAmount))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, nativeCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, nativeCoins))

	// Check initial balances
	receiverBalanceBefore := bankKeeper.GetBalance(ctx, receiverAddr, fixture.ibcDenom)
	require.Equal(t, amountInt, receiverBalanceBefore.Amount)

	moduleBalanceBefore := bankKeeper.GetBalance(ctx, moduleAddr, constants.BondDenom)
	require.Equal(t, convertedAmount, moduleBalanceBefore.Amount)

	// Execute OnRecvPacket
	respOnRecvPacket := ibcModule.OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{})

	// Validate that no error occurred
	require.NotNil(t, respOnRecvPacket)
	require.True(t, respOnRecvPacket.Success())

	// Check final balances
	receiverBalanceAfter := bankKeeper.GetBalance(ctx, receiverAddr, fixture.ibcDenom)
	require.Equal(t, amountInt, receiverBalanceAfter.Amount, "Receiver have the IBC tokens as no wrapping occurred")

	receiverNativeBalance := bankKeeper.GetBalance(ctx, receiverAddr, constants.BondDenom)
	require.Equal(t, sdkmath.ZeroInt(), receiverNativeBalance.Amount, "Receiver shouldn't have received native tokens as no wrapping occurred")

	moduleBondBalanceAfter := bankKeeper.GetBalance(ctx, moduleAddr, constants.BondDenom)
	require.Equal(t, convertedAmount, moduleBondBalanceAfter.Amount, "Module should have the native tokens as no wrapping occurred")

	moduleIBCBalanceAfter := bankKeeper.GetBalance(ctx, moduleAddr, fixture.ibcDenom)
	require.Equal(t, sdkmath.ZeroInt(), moduleIBCBalanceAfter.Amount, "Module shouldn't have the IBC tokens as no wrapping occurred")

	// Check TotalTransferredIn
	totalTransferredIn := k.GetTotalTransferredIn(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredIn, "TotalTransferredIn should be zero as no wrapping occurred")

	// Check TotalTransferredOut
	totalTransferredOut := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOut, "TotalTransferredOut should be zero after successful OnRecvPacket")
}

func TestOnRecvPacket_RealKeepers_Positive_DisabledTW_DifferentDenom(t *testing.T) {
	// Test case: OnRecvPacket with real keepers to verify balance no change when
	// the denom sent is different from the module denom and the TokenwrapperKeeper is disabled

	// set up positive fixture
	fixture := getRecvPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, false)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, "diff-denom")
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := testutil.NewMockTransferKeeper(ctrl)
	appMock := testutil.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := testutil.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := testutil.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k, transferKeeperMock, nil, appMock, channelKeeperMock, connectionKeeperMock,
	)

	// Create the data and the packet
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
		SourcePort:         fixture.counterpartyPort,
		SourceChannel:      fixture.counterpartyChannel,
		DestinationPort:    fixture.nativePort,
		DestinationChannel: fixture.nativeChannel,
		Data:               dataBz,
	}

	// Set up channel and connection mocks
	channelKeeperMock.
		EXPECT().
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
		Times(3)

	connectionKeeperMock.
		EXPECT().
		GetConnection(ctx, fixture.nativeConnectionId).
		Return(
			connectiontypes.ConnectionEnd{
				ClientId: fixture.nativeClientId,
				Counterparty: connectiontypes.Counterparty{
					ClientId: fixture.counterpartyClientId,
				},
			}, true).
		Times(1)

	appMock.
		EXPECT().
		OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).
		Return(channeltypes.NewResultAcknowledgement([]byte("success")))

	// Set up initial balances
	receiverAddr := sdk.MustAccAddressFromBech32(fixture.receiver)
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)

	// Mint IBC tokens to receiver
	amountInt, _ := sdkmath.NewIntFromString(fixture.amount)
	ibcCoins := sdk.NewCoins(sdk.NewCoin(fixture.ibcDenom, amountInt))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, ibcCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, receiverAddr, ibcCoins))

	// Mint native tokens to module
	conversionFactor := k.GetDecimalConversionFactor(ctx)
	convertedAmount := amountInt.Quo(conversionFactor)
	nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, convertedAmount))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, nativeCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, nativeCoins))

	// Check initial balances
	receiverBalanceBefore := bankKeeper.GetBalance(ctx, receiverAddr, fixture.ibcDenom)
	require.Equal(t, amountInt, receiverBalanceBefore.Amount)

	moduleBalanceBefore := bankKeeper.GetBalance(ctx, moduleAddr, constants.BondDenom)
	require.Equal(t, convertedAmount, moduleBalanceBefore.Amount)

	// Execute OnRecvPacket
	respOnRecvPacket := ibcModule.OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{})

	// Validate that no error occurred
	require.NotNil(t, respOnRecvPacket)
	require.True(t, respOnRecvPacket.Success())

	// Check final balances
	receiverBalanceAfter := bankKeeper.GetBalance(ctx, receiverAddr, fixture.ibcDenom)
	require.Equal(t, amountInt, receiverBalanceAfter.Amount, "Receiver have the IBC tokens as no wrapping occurred")

	receiverNativeBalance := bankKeeper.GetBalance(ctx, receiverAddr, constants.BondDenom)
	require.Equal(t, sdkmath.ZeroInt(), receiverNativeBalance.Amount, "Receiver shouldn't have received native tokens as no wrapping occurred")

	moduleBondBalanceAfter := bankKeeper.GetBalance(ctx, moduleAddr, constants.BondDenom)
	require.Equal(t, convertedAmount, moduleBondBalanceAfter.Amount, "Module should have the native tokens as no wrapping occurred")

	moduleIBCBalanceAfter := bankKeeper.GetBalance(ctx, moduleAddr, fixture.ibcDenom)
	require.Equal(t, sdkmath.ZeroInt(), moduleIBCBalanceAfter.Amount, "Module shouldn't have the IBC tokens as no wrapping occurred")

	// Check TotalTransferredIn
	totalTransferredIn := k.GetTotalTransferredIn(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredIn, "TotalTransferredIn should be zero as no wrapping occurred")

	// Check TotalTransferredOut
	totalTransferredOut := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOut, "TotalTransferredOut should be zero after successful OnRecvPacket")
}

func TestOnRecvPacket_RealKeepers_Negative_DisabledTW(t *testing.T) {
	// Test case: OnRecvPacket with real keepers to verify balance no change when
	// the denom sent the denom from the module denom and the TokenwrapperKeeper is disabled

	// set up positive fixture
	fixture := getRecvPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, false)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := testutil.NewMockTransferKeeper(ctrl)
	appMock := testutil.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := testutil.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := testutil.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k, transferKeeperMock, nil, appMock, channelKeeperMock, connectionKeeperMock,
	)

	// Create the data and the packet
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
		SourcePort:         fixture.counterpartyPort,
		SourceChannel:      fixture.counterpartyChannel,
		DestinationPort:    fixture.nativePort,
		DestinationChannel: fixture.nativeChannel,
		Data:               dataBz,
	}

	// code will check that App callback is called and successful
	appMock.
		EXPECT().
		OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).
		Return(channeltypes.NewResultAcknowledgement([]byte("success")))

	// Set up channel and connection mocks
	channelKeeperMock.
		EXPECT().
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
		Times(3)

	connectionKeeperMock.
		EXPECT().
		GetConnection(ctx, fixture.nativeConnectionId).
		Return(
			connectiontypes.ConnectionEnd{
				ClientId: fixture.nativeClientId,
				Counterparty: connectiontypes.Counterparty{
					ClientId: fixture.counterpartyClientId,
				},
			}, true).
		Times(1)

	// Set up initial balances
	// ------------------------------------------------------

	faucetAddr := sdk.MustAccAddressFromBech32(fixture.faucet)
	receiverAddr := sdk.MustAccAddressFromBech32(fixture.receiver)
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)

	// Mint IBC tokens to sender
	amountInt, _ := sdkmath.NewIntFromString(fixture.amount)
	ibcCoins := sdk.NewCoins(sdk.NewCoin(fixture.ibcDenom, amountInt))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, ibcCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, faucetAddr, ibcCoins))

	// Mint native tokens to module
	conversionFactor := k.GetDecimalConversionFactor(ctx)
	convertedAmount := amountInt.Quo(conversionFactor)
	nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, convertedAmount))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, nativeCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, nativeCoins))

	// Check initial balances
	// ------------------------------------------------------

	receiverBalanceBefore := bankKeeper.GetBalance(ctx, receiverAddr, fixture.ibcDenom)
	require.Equal(t, sdkmath.ZeroInt(), receiverBalanceBefore.Amount)

	moduleBalanceBefore := bankKeeper.GetBalance(ctx, moduleAddr, constants.BondDenom)
	require.Equal(t, convertedAmount, moduleBalanceBefore.Amount)

	// Execute
	// ------------------------------------------------------

	ack := ibcModule.OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{})

	// Check error
	// ------------------------------------------------------

	// Validate that no error occurred with OnRecvPacket IBC stack
	require.NotNil(t, ack)
	require.True(t, ack.Success(), "OnRecvPacket should not fail when the TokenWrapper is disabled")

	// Check stats
	// ------------------------------------------------------

	// Check that totals have remained unchanged since processing failed
	totalTransferredInAfter := k.GetTotalTransferredIn(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredInAfter,
		"TotalTransferredIn should remain zero after failed OnRecvPacket")

	totalTransferredOutAfter := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOutAfter,
		"TotalTransferredOut should remain zero after failed OnRecvPacket")

	// Check events
	// ------------------------------------------------------
	events := ctx.EventManager().Events()

	// Check that the last event is the error event
	require.Equal(t, types.EventTypeTokenWrapperError, events[len(events)-1].Type)

	// Check the error message in the event
	require.Contains(t, events[len(events)-1].Attributes[0].Value, "module disabled: module functionality is not enabled")

	// Check final balances
	// ------------------------------------------------------

	receiverBalanceAfter := bankKeeper.GetBalance(ctx, receiverAddr, fixture.ibcDenom)
	require.Equal(t, sdkmath.ZeroInt(), receiverBalanceAfter.Amount, "Receiver have the IBC tokens as no wrapping occurred")

	receiverNativeBalance := bankKeeper.GetBalance(ctx, receiverAddr, constants.BondDenom)
	require.Equal(t, sdkmath.ZeroInt(), receiverNativeBalance.Amount, "Receiver shouldn't have received native tokens as no wrapping occurred")

	moduleBondBalanceAfter := bankKeeper.GetBalance(ctx, moduleAddr, constants.BondDenom)
	require.Equal(t, convertedAmount, moduleBondBalanceAfter.Amount, "Module should have the native tokens as no wrapping occurred")

	moduleIBCBalanceAfter := bankKeeper.GetBalance(ctx, moduleAddr, fixture.ibcDenom)
	require.Equal(t, sdkmath.ZeroInt(), moduleIBCBalanceAfter.Amount, "Module shouldn't have the IBC tokens as no wrapping occurred")

	// Check TotalTransferredIn
	totalTransferredIn := k.GetTotalTransferredIn(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredIn, "TotalTransferredIn should be zero as no wrapping occurred")

	// Check TotalTransferredOut
	totalTransferredOut := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOut, "TotalTransferredOut should be zero after successful OnRecvPacket")
}

func TestOnRecvPacket_RealKeepers_Negative_DiffCounterpartyChannel(t *testing.T) {
	// Test case: OnRecvPacket with real keepers
	// The counterpartyChannel is different from the one set in the TokenwrapperKeeper
	// so it shouldn't be processed by the TokenwrapperKeeper

	// set up positive fixture
	fixture := getRecvPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := testutil.NewMockTransferKeeper(ctrl)
	appMock := testutil.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := testutil.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := testutil.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k, transferKeeperMock, nil, appMock, channelKeeperMock, connectionKeeperMock,
	)

	// Create the data and the packet
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
		SourcePort:         fixture.counterpartyPort,
		SourceChannel:      fixture.counterpartyChannel,
		DestinationPort:    fixture.nativePort,
		DestinationChannel: "channel-2", // Channel on ZIGChain that is not the one set in the TokenwrapperKeeper
		Data:               dataBz,
	}

	// code will check that App callback is called and successful
	appMock.
		EXPECT().
		OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).
		Return(channeltypes.NewResultAcknowledgement([]byte("success")))

	// Set up channel and connection mocks
	// ------------------------------------------------------

	// code will check that the channel is valid for the nativePort and nativeChannel (open and connected)
	// but the channel returned is not the same that the destinationChannel in the packet
	channelKeeperMock.
		EXPECT().
		GetChannel(ctx, fixture.nativePort, "channel-2").
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.counterpartyPort,
					ChannelId: fixture.counterpartyChannel,
				},
			}, true).
		Times(3)

	connectionKeeperMock.
		EXPECT().
		GetConnection(ctx, fixture.nativeConnectionId).
		Return(connectiontypes.ConnectionEnd{ClientId: fixture.nativeClientId,
			Counterparty: connectiontypes.Counterparty{
				ClientId: fixture.counterpartyClientId,
			},
		}, true).
		Times(1)

	// Set up initial balances
	// ------------------------------------------------------

	faucetAddr := sdk.MustAccAddressFromBech32(fixture.faucet)
	receiverAddr := sdk.MustAccAddressFromBech32(fixture.receiver)
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)

	// Mint IBC tokens to sender
	amountInt, _ := sdkmath.NewIntFromString(fixture.amount)
	ibcCoins := sdk.NewCoins(sdk.NewCoin(fixture.ibcDenom, amountInt))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, ibcCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, faucetAddr, ibcCoins))

	// Mint native tokens to module
	conversionFactor := k.GetDecimalConversionFactor(ctx)
	convertedAmount := amountInt.Quo(conversionFactor)
	nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, convertedAmount))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, nativeCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, nativeCoins))

	// Check initial balances
	// ------------------------------------------------------

	receiverBalanceBefore := bankKeeper.GetBalance(ctx, receiverAddr, fixture.ibcDenom)
	require.Equal(t, sdkmath.ZeroInt(), receiverBalanceBefore.Amount)

	moduleBalanceBefore := bankKeeper.GetBalance(ctx, moduleAddr, constants.BondDenom)
	require.Equal(t, convertedAmount, moduleBalanceBefore.Amount)

	// Execute
	// ------------------------------------------------------

	ack := ibcModule.OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{})

	// Check error
	// ------------------------------------------------------

	// Validate that no error occurred in the IBC OnRecvPacket middlestack
	require.NotNil(t, ack)
	require.True(t, ack.Success(), "OnRecvPacket should not fail when the IBC settings mismatch")

	// Check stats
	// ------------------------------------------------------

	// Check that totals have remained unchanged since processing failed
	totalTransferredInAfter := k.GetTotalTransferredIn(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredInAfter,
		"TotalTransferredIn should remain zero after failed OnRecvPacket")

	totalTransferredOutAfter := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOutAfter,
		"TotalTransferredOut should remain zero after failed OnRecvPacket")

	// Check events
	// ------------------------------------------------------
	events := ctx.EventManager().Events()

	// Check that the last event is the error event
	require.Equal(t, types.EventTypeTokenWrapperError, events[len(events)-1].Type)

	// Check the error message in the event
	require.Contains(t, events[len(events)-1].Attributes[0].Value, fmt.Sprintf("ibc settings do not match the expected values, failed with packet: {1 %s %s %s %s", fixture.counterpartyPort, fixture.counterpartyChannel, fixture.nativePort, "channel-2"))

	// Check final balances
	// ------------------------------------------------------

	receiverBalanceAfter := bankKeeper.GetBalance(ctx, receiverAddr, fixture.ibcDenom)
	require.Equal(t, sdkmath.ZeroInt(), receiverBalanceAfter.Amount, "Receiver have the IBC tokens as no wrapping occurred")

	receiverNativeBalance := bankKeeper.GetBalance(ctx, receiverAddr, constants.BondDenom)
	require.Equal(t, sdkmath.ZeroInt(), receiverNativeBalance.Amount, "Receiver shouldn't have received native tokens as no wrapping occurred")

	moduleBondBalanceAfter := bankKeeper.GetBalance(ctx, moduleAddr, constants.BondDenom)
	require.Equal(t, convertedAmount, moduleBondBalanceAfter.Amount, "Module should have the native tokens as no wrapping occurred")

	moduleIBCBalanceAfter := bankKeeper.GetBalance(ctx, moduleAddr, fixture.ibcDenom)
	require.Equal(t, sdkmath.ZeroInt(), moduleIBCBalanceAfter.Amount, "Module shouldn't have the IBC tokens as no wrapping occurred")
}

func TestOnRecvPacket_RealKeepers_Negative_InsufficientFunds(t *testing.T) {
	// Test case: OnRecvPacket with real keepers when there are not enough funds in the module to wrap the IBC tokens

	// set up a positive fixture
	fixture := getRecvPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := testutil.NewMockTransferKeeper(ctrl)
	appMock := testutil.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := testutil.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := testutil.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k, transferKeeperMock, nil, appMock, channelKeeperMock, connectionKeeperMock,
	)

	// Create the data and the packet
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
		SourcePort:         fixture.counterpartyPort,
		SourceChannel:      fixture.counterpartyChannel,
		DestinationPort:    fixture.nativePort,
		DestinationChannel: fixture.nativeChannel,
		Data:               dataBz,
	}

	// Set up channel and connection mocks
	channelKeeperMock.
		EXPECT().
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
		Times(3)

	connectionKeeperMock.
		EXPECT().
		GetConnection(ctx, fixture.nativeConnectionId).
		Return(connectiontypes.ConnectionEnd{ClientId: fixture.nativeClientId,
			Counterparty: connectiontypes.Counterparty{
				ClientId: fixture.counterpartyClientId,
			},
		}, true).
		Times(1)

	appMock.
		EXPECT().
		OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).
		Return(channeltypes.NewResultAcknowledgement([]byte("success")))

	// Set up initial balances
	receiverAddr := sdk.MustAccAddressFromBech32(fixture.receiver)
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)

	// Mint IBC tokens to receiver
	amountInt, _ := sdkmath.NewIntFromString(fixture.amount)
	ibcCoins := sdk.NewCoins(sdk.NewCoin(fixture.ibcDenom, amountInt))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, ibcCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, receiverAddr, ibcCoins))

	// Mint native tokens to module
	conversionFactor := k.GetDecimalConversionFactor(ctx)
	convertedAmount := amountInt.Quo(conversionFactor)
	insufficientAmount := convertedAmount.Quo(sdkmath.NewInt(2))
	nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, insufficientAmount))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, nativeCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, nativeCoins))

	// Check initial balances
	receiverNativeBalanceBefore := bankKeeper.GetBalance(ctx, receiverAddr, constants.BondDenom)
	require.Equal(t, sdkmath.ZeroInt(), receiverNativeBalanceBefore.Amount)

	receiverIBCBalanceBefore := bankKeeper.GetBalance(ctx, receiverAddr, fixture.ibcDenom)
	require.Equal(t, amountInt, receiverIBCBalanceBefore.Amount)

	moduleNativeBalanceBefore := bankKeeper.GetBalance(ctx, moduleAddr, constants.BondDenom)
	require.Equal(t, insufficientAmount, moduleNativeBalanceBefore.Amount)

	moduleIBCBalanceBefore := bankKeeper.GetBalance(ctx, moduleAddr, fixture.ibcDenom)
	require.Equal(t, sdkmath.ZeroInt(), moduleIBCBalanceBefore.Amount)
	// Execute OnRecvPacket
	respOnRecvPacket := ibcModule.OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{})

	// Validate that ERROR occurred
	require.NotNil(t, respOnRecvPacket)
	require.True(t, respOnRecvPacket.Success(), "OnRecvPacket should not fail due to insufficient module funds")

	// Check that balances remained UNCHANGED (no partial transactions)
	receiverNativeBalanceAfter := bankKeeper.GetBalance(ctx, receiverAddr, constants.BondDenom)
	require.Equal(t, sdkmath.ZeroInt(), receiverNativeBalanceAfter.Amount, "Receiver should still have no native tokens")

	receiverIBCBalanceAfter := bankKeeper.GetBalance(ctx, receiverAddr, fixture.ibcDenom)
	require.Equal(t, amountInt, receiverIBCBalanceAfter.Amount, "Receiver should still have original IBC tokens")

	moduleNativeBalanceAfter := bankKeeper.GetBalance(ctx, moduleAddr, constants.BondDenom)
	require.Equal(t, insufficientAmount, moduleNativeBalanceAfter.Amount, "Module should still have insufficient native tokens")

	moduleIBCBalanceAfter := bankKeeper.GetBalance(ctx, moduleAddr, fixture.ibcDenom)
	require.Equal(t, sdkmath.ZeroInt(), moduleIBCBalanceAfter.Amount, "Module should still have no IBC tokens")

	// Check that stats remained unchanged
	finalTotalTransferredIn := k.GetTotalTransferredIn(ctx)
	require.Equal(t, sdkmath.ZeroInt(), finalTotalTransferredIn, "TotalTransferredIn should remain zero on failure")

	finalTotalTransferredOut := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), finalTotalTransferredOut, "TotalTransferredOut should remain zero on failure")

	// Check that the ERROR event was emitted
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// Find the error event
	var foundErrorEvent bool
	for _, event := range events {
		if event.Type == types.EventTypeTokenWrapperError {
			foundErrorEvent = true
			// Check that error message mentions insufficient balance
			for _, attr := range event.Attributes {
				if attr.Key == "error" {
					require.Contains(t, attr.Value, "does not have enough balance",
						"Error should mention insufficient balance")
				}
			}
			break
		}
	}

	require.True(t, foundErrorEvent, "Expected EventTypeTokenWrapperError event to be emitted")

	// Verify NO success event was emitted
	var foundSuccessEvent bool
	for _, event := range events {
		if event.Type == types.EventTypeTokenWrapperOnRecvPacket {
			foundSuccessEvent = true
			break
		}
	}

	require.False(t, foundSuccessEvent, "Should NOT emit success event when transaction fails")
}

func TestOnRecvPacket(t *testing.T) {

	// Setup gomock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

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

	// Test cases
	tests := []struct {
		name        string
		enabled     bool
		packet      channeltypes.Packet
		data        transfertypes.FungibleTokenPacketData
		dataBz      []byte // For cases where we need custom packet data
		setupMocks  func(ctx sdk.Context, packet channeltypes.Packet, keeperMock *testutil.MockTokenwrapperKeeper, transferKeeperMock *testutil.MockTransferKeeper, bankKeeperMock *testutil.MockBankKeeper, appMock *testutil.MockCallbacksCompatibleModule, channelKeeperMock *testutil.MockChannelKeeper, connectionKeeperMock *testutil.MockConnectionKeeper)
		expectedAck ibcexported.Acknowledgement
		expectError bool
	}{
		{
			name:    "Happy path - successful token wrapping",
			enabled: true,
			packet: channeltypes.Packet{
				Sequence:           1,
				SourcePort:         counterpartyPort,
				SourceChannel:      counterpartyChannel,
				DestinationPort:    nativePort,
				DestinationChannel: nativeChannel,
				Data:               nil, // Set in setupMocks
			},
			data: transfertypes.FungibleTokenPacketData{
				Denom:    moduleDenom,
				Amount:   "1000000000000000000", // 1 token with 18 decimals
				Sender:   sender,
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, packet channeltypes.Packet, keeperMock *testutil.MockTokenwrapperKeeper, transferKeeperMock *testutil.MockTransferKeeper, bankKeeperMock *testutil.MockBankKeeper, appMock *testutil.MockCallbacksCompatibleModule, channelKeeperMock *testutil.MockChannelKeeper, connectionKeeperMock *testutil.MockConnectionKeeper) {

				// Check: Channel is active and Open
				channelKeeperMock.EXPECT().GetChannel(ctx, nativePort, nativeChannel).Return(channeltypes.Channel{
					State:          channeltypes.OPEN,
					ConnectionHops: []string{nativeConnectionId},
					Counterparty: channeltypes.Counterparty{
						PortId:    counterpartyPort,
						ChannelId: counterpartyChannel,
					},
				}, true).Times(3)

				// Connection information matches the ClientId
				connectionKeeperMock.EXPECT().GetConnection(ctx, nativeConnectionId).Return(connectiontypes.ConnectionEnd{
					ClientId: nativeClientId,
					Counterparty: connectiontypes.Counterparty{
						ClientId: counterpartyClientId,
					},
				}, true).Times(1)

				//
				keeperMock.EXPECT().IsEnabled(ctx).Return(true).AnyTimes()
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort).AnyTimes()
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel).AnyTimes()
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom).AnyTimes()
				keeperMock.EXPECT().GetNativeClientId(ctx).Return(nativeClientId).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyClientId(ctx).Return(counterpartyClientId).AnyTimes()
				keeperMock.EXPECT().AddToTotalTransferredIn(ctx, sdkmath.NewInt(1000000)).Return()
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				keeperMock.EXPECT().GetIBCRecvDenom(ctx, moduleDenom).Return(ibcDenom)
				keeperMock.EXPECT().ScaleDownTokenPrecision(ctx, sdkmath.NewInt(1000000000000000000)).Return(sdkmath.NewInt(1000000), nil)
				keeperMock.EXPECT().CheckBalances(ctx, sdk.MustAccAddressFromBech32(receiver), sdkmath.NewInt(1000000000000000000), ibcDenom, sdkmath.NewInt(1000000)).Return(nil)
				keeperMock.EXPECT().LockIBCTokens(ctx, sdk.MustAccAddressFromBech32(receiver), sdkmath.NewInt(1000000000000000000), ibcDenom).Return(sdk.NewCoins(sdk.NewCoin(ibcDenom, sdkmath.NewInt(1000000000000000000))), nil)
				keeperMock.EXPECT().UnlockNativeTokens(
					ctx,
					sdk.MustAccAddressFromBech32(receiver),
					sdkmath.NewInt(1000000),
					sdk.NewCoins(sdk.NewCoin(ibcDenom, sdkmath.NewInt(1000000000000000000)))).
					Return(sdk.NewCoins(sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(1000000))), nil)
				appMock.EXPECT().OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).Return(channeltypes.NewResultAcknowledgement([]byte("success")))
			},
			expectedAck: channeltypes.NewResultAcknowledgement([]byte("success")),
			expectError: false,
		},
		{
			name:    "Error - channel not found",
			enabled: true,
			packet: channeltypes.Packet{
				Sequence:           1,
				SourcePort:         counterpartyPort,
				SourceChannel:      counterpartyChannel,
				DestinationPort:    nativePort,
				DestinationChannel: nativeChannel,
				Data:               nil, // Set in setupMocks
			},
			data: transfertypes.FungibleTokenPacketData{
				Denom:    moduleDenom,
				Amount:   "1000000000000000000",
				Sender:   sender,
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, packet channeltypes.Packet, keeperMock *testutil.MockTokenwrapperKeeper, transferKeeperMock *testutil.MockTransferKeeper, bankKeeperMock *testutil.MockBankKeeper, appMock *testutil.MockCallbacksCompatibleModule, channelKeeperMock *testutil.MockChannelKeeper, connectionKeeperMock *testutil.MockConnectionKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, nativePort, nativeChannel).Return(channeltypes.Channel{}, false)
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
			},
			expectedAck: channeltypes.NewErrorAcknowledgement(types.ErrChannelNotFound.Wrapf("%s/%s", counterpartyPort, counterpartyChannel)),
			expectError: false,
		},
		{
			name:    "Error - channel not open",
			enabled: true,
			packet: channeltypes.Packet{
				Sequence:           1,
				SourcePort:         counterpartyPort,
				SourceChannel:      counterpartyChannel,
				DestinationPort:    nativePort,
				DestinationChannel: nativeChannel,
				Data:               nil, // Set in setupMocks
			},
			data: transfertypes.FungibleTokenPacketData{
				Denom:    moduleDenom,
				Amount:   "1000000000000000000",
				Sender:   sender,
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, packet channeltypes.Packet, keeperMock *testutil.MockTokenwrapperKeeper, transferKeeperMock *testutil.MockTransferKeeper, bankKeeperMock *testutil.MockBankKeeper, appMock *testutil.MockCallbacksCompatibleModule, channelKeeperMock *testutil.MockChannelKeeper, connectionKeeperMock *testutil.MockConnectionKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, nativePort, nativeChannel).Return(channeltypes.Channel{
					State: channeltypes.CLOSED,
				}, true)
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
			},
			expectedAck: channeltypes.NewErrorAcknowledgement(types.ErrChannelNotOpen.Wrapf("%s/%s", counterpartyPort, counterpartyChannel)),
			expectError: false,
		},
		{
			name:    "Error - invalid packet data",
			enabled: true,
			packet: channeltypes.Packet{
				Sequence:           1,
				SourcePort:         counterpartyPort,
				SourceChannel:      counterpartyChannel,
				DestinationPort:    nativePort,
				DestinationChannel: nativeChannel,
				Data:               []byte("invalid json"),
			},
			data: transfertypes.FungibleTokenPacketData{},
			setupMocks: func(ctx sdk.Context, packet channeltypes.Packet, keeperMock *testutil.MockTokenwrapperKeeper, transferKeeperMock *testutil.MockTransferKeeper, bankKeeperMock *testutil.MockBankKeeper, appMock *testutil.MockCallbacksCompatibleModule, channelKeeperMock *testutil.MockChannelKeeper, connectionKeeperMock *testutil.MockConnectionKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, nativePort, nativeChannel).Return(channeltypes.Channel{
					State:          channeltypes.OPEN,
					ConnectionHops: []string{nativeConnectionId},
				}, true).Times(1)
				connectionKeeperMock.EXPECT().GetConnection(ctx, nativeConnectionId).Return(connectiontypes.ConnectionEnd{
					ClientId: nativeClientId,
				}, true).Times(0)
				keeperMock.EXPECT().IsEnabled(ctx).Return(true).AnyTimes()
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort).AnyTimes()
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel).AnyTimes()
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom).AnyTimes()
				keeperMock.EXPECT().GetNativeClientId(ctx).Return(nativeClientId).AnyTimes()
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				appMock.EXPECT().OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).Return(channeltypes.NewErrorAcknowledgement(fmt.Errorf("cannot unmarshal ICS-20 transfer packet data: invalid character 'i' looking for beginning of value"))).Times(0)
			},
			expectedAck: channeltypes.NewErrorAcknowledgement(fmt.Errorf("cannot unmarshal ICS-20 transfer packet data: invalid character 'i' looking for beginning of value")),
			expectError: false,
		},
		{
			name:    "Error - malformed packet data",
			enabled: true,
			packet: channeltypes.Packet{
				Sequence:           1,
				SourcePort:         counterpartyPort,
				SourceChannel:      counterpartyChannel,
				DestinationPort:    nativePort,
				DestinationChannel: nativeChannel,
				Data:               []byte(`{"denom":"test","amount":"100","sender":"invalid","receiver":"invalid"}`), // Missing required fields
			},
			data: transfertypes.FungibleTokenPacketData{},
			setupMocks: func(ctx sdk.Context, packet channeltypes.Packet, keeperMock *testutil.MockTokenwrapperKeeper, transferKeeperMock *testutil.MockTransferKeeper, bankKeeperMock *testutil.MockBankKeeper, appMock *testutil.MockCallbacksCompatibleModule, channelKeeperMock *testutil.MockChannelKeeper, connectionKeeperMock *testutil.MockConnectionKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, nativePort, nativeChannel).Return(channeltypes.Channel{
					State:          channeltypes.OPEN,
					ConnectionHops: []string{nativeConnectionId},
				}, true).Times(1)
				connectionKeeperMock.EXPECT().GetConnection(ctx, nativeConnectionId).Return(connectiontypes.ConnectionEnd{
					ClientId: nativeClientId,
				}, true).Times(0)
				keeperMock.EXPECT().IsEnabled(ctx).Return(true).AnyTimes()
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort).AnyTimes()
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel).AnyTimes()
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom).AnyTimes()
				keeperMock.EXPECT().GetNativeClientId(ctx).Return(nativeClientId).AnyTimes()
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				appMock.EXPECT().OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).Return(channeltypes.NewErrorAcknowledgement(fmt.Errorf("cannot unmarshal ICS-20 transfer packet data: missing required fields"))).Times(0)
			},
			expectedAck: channeltypes.NewErrorAcknowledgement(fmt.Errorf("cannot unmarshal ICS-20 transfer packet data: missing required fields")),
			expectError: false,
		},
		{
			name:    "Error - invalid receiver address",
			enabled: true,
			packet: channeltypes.Packet{
				Sequence:           1,
				SourcePort:         counterpartyPort,
				SourceChannel:      counterpartyChannel,
				DestinationPort:    nativePort,
				DestinationChannel: nativeChannel,
				Data:               nil,
			},
			data: transfertypes.FungibleTokenPacketData{
				Denom:    moduleDenom,
				Amount:   "1000000000000000000",
				Sender:   sender,
				Receiver: "invalid-address",
			},
			setupMocks: func(ctx sdk.Context, packet channeltypes.Packet, keeperMock *testutil.MockTokenwrapperKeeper, transferKeeperMock *testutil.MockTransferKeeper, bankKeeperMock *testutil.MockBankKeeper, appMock *testutil.MockCallbacksCompatibleModule, channelKeeperMock *testutil.MockChannelKeeper, connectionKeeperMock *testutil.MockConnectionKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, nativePort, nativeChannel).Return(channeltypes.Channel{
					State:          channeltypes.OPEN,
					ConnectionHops: []string{nativeConnectionId},
				}, true).Times(1)
				connectionKeeperMock.EXPECT().GetConnection(ctx, nativeConnectionId).Return(connectiontypes.ConnectionEnd{
					ClientId: nativeClientId,
				}, true).Times(0)
				keeperMock.EXPECT().IsEnabled(ctx).Return(true).AnyTimes()
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort).AnyTimes()
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel).AnyTimes()
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom).AnyTimes()
				keeperMock.EXPECT().GetNativeClientId(ctx).Return(nativeClientId).AnyTimes()
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				appMock.EXPECT().OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).Return(channeltypes.NewResultAcknowledgement([]byte("success"))).Times(0)
			},
			expectedAck: channeltypes.NewErrorAcknowledgement(fmt.Errorf("invalid address")),
			expectError: false,
		},
		{
			name:    "Error - invalid amount",
			enabled: true,
			packet: channeltypes.Packet{
				Sequence:           1,
				SourcePort:         counterpartyPort,
				SourceChannel:      counterpartyChannel,
				DestinationPort:    nativePort,
				DestinationChannel: nativeChannel,
				Data:               nil,
			},
			data: transfertypes.FungibleTokenPacketData{
				Denom:    moduleDenom,
				Amount:   "invalid-amount",
				Sender:   sender,
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, packet channeltypes.Packet, keeperMock *testutil.MockTokenwrapperKeeper, transferKeeperMock *testutil.MockTransferKeeper, bankKeeperMock *testutil.MockBankKeeper, appMock *testutil.MockCallbacksCompatibleModule, channelKeeperMock *testutil.MockChannelKeeper, connectionKeeperMock *testutil.MockConnectionKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, nativePort, nativeChannel).Return(channeltypes.Channel{
					State:          channeltypes.OPEN,
					ConnectionHops: []string{nativeConnectionId},
				}, true).Times(1)
				connectionKeeperMock.EXPECT().GetConnection(ctx, nativeConnectionId).Return(connectiontypes.ConnectionEnd{
					ClientId: nativeClientId,
				}, true).Times(0)
				keeperMock.EXPECT().IsEnabled(ctx).Return(true).AnyTimes()
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort).AnyTimes()
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel).AnyTimes()
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom).AnyTimes()
				keeperMock.EXPECT().GetNativeClientId(ctx).Return(nativeClientId).AnyTimes()
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				appMock.EXPECT().OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).Return(channeltypes.NewResultAcknowledgement([]byte("success"))).Times(0)
			},
			expectedAck: channeltypes.NewErrorAcknowledgement(fmt.Errorf("invalid amount: invalid-amount")),
			expectError: false,
		},
		{
			name:    "Error - zero amount",
			enabled: true,
			packet: channeltypes.Packet{
				Sequence:           1,
				SourcePort:         counterpartyPort,
				SourceChannel:      counterpartyChannel,
				DestinationPort:    nativePort,
				DestinationChannel: nativeChannel,
				Data:               nil,
			},
			data: transfertypes.FungibleTokenPacketData{
				Denom:    moduleDenom,
				Amount:   "0",
				Sender:   sender,
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, packet channeltypes.Packet, keeperMock *testutil.MockTokenwrapperKeeper, transferKeeperMock *testutil.MockTransferKeeper, bankKeeperMock *testutil.MockBankKeeper, appMock *testutil.MockCallbacksCompatibleModule, channelKeeperMock *testutil.MockChannelKeeper, connectionKeeperMock *testutil.MockConnectionKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, nativePort, nativeChannel).Return(channeltypes.Channel{
					State:          channeltypes.OPEN,
					ConnectionHops: []string{nativeConnectionId},
				}, true).Times(1)
				connectionKeeperMock.EXPECT().GetConnection(ctx, nativeConnectionId).Return(connectiontypes.ConnectionEnd{
					ClientId: nativeClientId,
				}, true).Times(0)
				keeperMock.EXPECT().IsEnabled(ctx).Return(true).AnyTimes()
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort).AnyTimes()
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel).AnyTimes()
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom).AnyTimes()
				keeperMock.EXPECT().GetNativeClientId(ctx).Return(nativeClientId).AnyTimes()
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				keeperMock.EXPECT().GetIBCRecvDenom(ctx, moduleDenom).Return(ibcDenom).Times(0)
				keeperMock.EXPECT().ScaleDownTokenPrecision(ctx, sdkmath.ZeroInt()).Return(sdkmath.Int{}, fmt.Errorf("converted amount is zero or negative: %s", sdkmath.ZeroInt().String())).Times(0)
				keeperMock.EXPECT().CheckBalances(ctx, sdk.MustAccAddressFromBech32(receiver), sdkmath.ZeroInt(), ibcDenom, sdkmath.ZeroInt()).Return(nil).Times(0)
				keeperMock.EXPECT().LockIBCTokens(ctx, sdk.MustAccAddressFromBech32(receiver), sdkmath.ZeroInt(), ibcDenom).Return(sdk.NewCoins(sdk.NewCoin(ibcDenom, sdkmath.ZeroInt())), nil).Times(0)
				keeperMock.EXPECT().UnlockNativeTokens(
					ctx,
					sdk.MustAccAddressFromBech32(receiver),
					sdkmath.ZeroInt(),
					sdk.NewCoins(sdk.NewCoin(ibcDenom, sdkmath.ZeroInt()))).
					Return(sdk.NewCoins(sdk.NewCoin(constants.BondDenom, sdkmath.ZeroInt())), nil).Times(0)
				appMock.EXPECT().OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).Return(channeltypes.NewResultAcknowledgement([]byte("success"))).Times(0)
			},
			expectedAck: channeltypes.NewResultAcknowledgement([]byte("success")),
			expectError: false,
		},
		{
			name:    "No wrapping - non-source chain",
			enabled: true,
			packet: channeltypes.Packet{
				Sequence:           1,
				SourcePort:         counterpartyPort,
				SourceChannel:      counterpartyChannel,
				DestinationPort:    nativePort,
				DestinationChannel: nativeChannel,
				Data:               nil,
			},
			data: transfertypes.FungibleTokenPacketData{
				Denom:    "ibc/" + moduleDenom, // Prefixed to indicate non-source chain
				Amount:   "1000000000000000000",
				Sender:   sender,
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, packet channeltypes.Packet, keeperMock *testutil.MockTokenwrapperKeeper, transferKeeperMock *testutil.MockTransferKeeper, bankKeeperMock *testutil.MockBankKeeper, appMock *testutil.MockCallbacksCompatibleModule, channelKeeperMock *testutil.MockChannelKeeper, connectionKeeperMock *testutil.MockConnectionKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, nativePort, nativeChannel).Return(channeltypes.Channel{
					State:          channeltypes.OPEN,
					ConnectionHops: []string{nativeConnectionId},
					Counterparty: channeltypes.Counterparty{
						PortId:    counterpartyPort,
						ChannelId: counterpartyChannel,
					},
				}, true).Times(3)
				connectionKeeperMock.EXPECT().GetConnection(ctx, nativeConnectionId).Return(connectiontypes.ConnectionEnd{
					ClientId: nativeClientId,
					Counterparty: connectiontypes.Counterparty{
						ClientId: counterpartyClientId,
					},
				}, true).Times(1)
				keeperMock.EXPECT().IsEnabled(ctx).Return(true).AnyTimes()
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort).AnyTimes()
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel).AnyTimes()
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom).AnyTimes()
				keeperMock.EXPECT().GetNativeClientId(ctx).Return(nativeClientId).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyClientId(ctx).Return(counterpartyClientId).AnyTimes()
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				appMock.EXPECT().OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).Return(channeltypes.NewResultAcknowledgement([]byte("success")))
			},
			expectedAck: channeltypes.NewResultAcknowledgement([]byte("success")),
			expectError: false,
		},
		{
			name:    "No wrapping - empty IBC settings",
			enabled: true,
			packet: channeltypes.Packet{
				Sequence:           1,
				SourcePort:         counterpartyPort,
				SourceChannel:      counterpartyChannel,
				DestinationPort:    nativePort,
				DestinationChannel: nativeChannel,
				Data:               nil,
			},
			data: transfertypes.FungibleTokenPacketData{
				Denom:    moduleDenom,
				Amount:   "1000000000000000000",
				Sender:   sender,
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, packet channeltypes.Packet, keeperMock *testutil.MockTokenwrapperKeeper, transferKeeperMock *testutil.MockTransferKeeper, bankKeeperMock *testutil.MockBankKeeper, appMock *testutil.MockCallbacksCompatibleModule, channelKeeperMock *testutil.MockChannelKeeper, connectionKeeperMock *testutil.MockConnectionKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, nativePort, nativeChannel).Return(channeltypes.Channel{
					State:          channeltypes.OPEN,
					ConnectionHops: []string{nativeConnectionId},
					Counterparty: channeltypes.Counterparty{
						PortId:    counterpartyPort,
						ChannelId: counterpartyChannel,
					},
				}, true).Times(3)
				connectionKeeperMock.EXPECT().GetConnection(ctx, nativeConnectionId).Return(connectiontypes.ConnectionEnd{
					ClientId: nativeClientId,
					Counterparty: connectiontypes.Counterparty{
						ClientId: counterpartyClientId,
					},
				}, true).Times(1)
				keeperMock.EXPECT().IsEnabled(ctx).Return(true).AnyTimes()
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort).AnyTimes()
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel).AnyTimes()
				keeperMock.EXPECT().GetDenom(ctx).Return("").AnyTimes()
				keeperMock.EXPECT().GetNativeClientId(ctx).Return(nativeClientId).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyClientId(ctx).Return(counterpartyClientId).AnyTimes()
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				appMock.EXPECT().OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).Return(channeltypes.NewResultAcknowledgement([]byte("success")))
			},
			expectedAck: channeltypes.NewResultAcknowledgement([]byte("success")),
			expectError: false,
		},
		{
			name:    "No wrapping - mismatched denom",
			enabled: false,
			packet: channeltypes.Packet{
				Sequence:           1,
				SourcePort:         counterpartyPort,
				SourceChannel:      counterpartyChannel,
				DestinationPort:    nativePort,
				DestinationChannel: nativeChannel,
				Data:               nil,
			},
			data: transfertypes.FungibleTokenPacketData{
				Denom:    "other-denom",
				Amount:   "1000000000000000000",
				Sender:   sender,
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, packet channeltypes.Packet, keeperMock *testutil.MockTokenwrapperKeeper, transferKeeperMock *testutil.MockTransferKeeper, bankKeeperMock *testutil.MockBankKeeper, appMock *testutil.MockCallbacksCompatibleModule, channelKeeperMock *testutil.MockChannelKeeper, connectionKeeperMock *testutil.MockConnectionKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, nativePort, nativeChannel).Return(channeltypes.Channel{
					State:          channeltypes.OPEN,
					ConnectionHops: []string{nativeConnectionId},
					Counterparty: channeltypes.Counterparty{
						PortId:    counterpartyPort,
						ChannelId: counterpartyChannel,
					},
				}, true).Times(3)
				connectionKeeperMock.EXPECT().GetConnection(ctx, nativeConnectionId).Return(connectiontypes.ConnectionEnd{
					ClientId: nativeClientId,
					Counterparty: connectiontypes.Counterparty{
						ClientId: counterpartyClientId,
					},
				}, true).Times(1)
				keeperMock.EXPECT().IsEnabled(ctx).Return(true).AnyTimes()
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort).AnyTimes()
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel).AnyTimes()
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom).AnyTimes()
				keeperMock.EXPECT().GetNativeClientId(ctx).Return(nativeClientId).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyClientId(ctx).Return(counterpartyClientId).AnyTimes()
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				appMock.EXPECT().OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).Return(channeltypes.NewResultAcknowledgement([]byte("success")))
			},
			expectedAck: channeltypes.NewResultAcknowledgement([]byte("success")),
			expectError: false,
		},
		{
			name:    "Error - LockTokens failure",
			enabled: true,
			packet: channeltypes.Packet{
				Sequence:           1,
				SourcePort:         counterpartyPort,
				SourceChannel:      counterpartyChannel,
				DestinationPort:    nativePort,
				DestinationChannel: nativeChannel,
				Data:               nil,
			},
			data: transfertypes.FungibleTokenPacketData{
				Denom:    moduleDenom,
				Amount:   "1000000000000000000",
				Sender:   sender,
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, packet channeltypes.Packet, keeperMock *testutil.MockTokenwrapperKeeper, transferKeeperMock *testutil.MockTransferKeeper, bankKeeperMock *testutil.MockBankKeeper, appMock *testutil.MockCallbacksCompatibleModule, channelKeeperMock *testutil.MockChannelKeeper, connectionKeeperMock *testutil.MockConnectionKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, nativePort, nativeChannel).Return(channeltypes.Channel{
					State:          channeltypes.OPEN,
					ConnectionHops: []string{nativeConnectionId},
					Counterparty: channeltypes.Counterparty{
						PortId:    counterpartyPort,
						ChannelId: counterpartyChannel,
					},
				}, true).Times(3)
				connectionKeeperMock.EXPECT().GetConnection(ctx, nativeConnectionId).Return(connectiontypes.ConnectionEnd{
					ClientId: nativeClientId,
					Counterparty: connectiontypes.Counterparty{
						ClientId: counterpartyClientId,
					},
				}, true).Times(1)
				keeperMock.EXPECT().IsEnabled(ctx).Return(true).AnyTimes()
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort).AnyTimes()
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel).AnyTimes()
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom).AnyTimes()
				keeperMock.EXPECT().GetNativeClientId(ctx).Return(nativeClientId).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyClientId(ctx).Return(counterpartyClientId).AnyTimes()
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				keeperMock.EXPECT().GetIBCRecvDenom(ctx, moduleDenom).Return(ibcDenom)
				keeperMock.EXPECT().ScaleDownTokenPrecision(ctx, sdkmath.NewInt(1000000000000000000)).Return(sdkmath.NewInt(1000000), nil)
				keeperMock.EXPECT().CheckBalances(ctx, sdk.MustAccAddressFromBech32(receiver), sdkmath.NewInt(1000000000000000000), ibcDenom, sdkmath.NewInt(1000000)).Return(nil)
				keeperMock.EXPECT().LockIBCTokens(ctx, sdk.MustAccAddressFromBech32(receiver), sdkmath.NewInt(1000000000000000000), ibcDenom).Return(sdk.NewCoins(sdk.NewCoin(ibcDenom, sdkmath.NewInt(1000000000000000000))), fmt.Errorf("failed to lock tokens:"))
				keeperMock.EXPECT().UnlockNativeTokens(
					ctx,
					sdk.MustAccAddressFromBech32(receiver),
					sdkmath.NewInt(1000000),
					sdk.NewCoins(sdk.NewCoin(ibcDenom, sdkmath.NewInt(1000000000000000000)))).
					Return(sdk.NewCoins(sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(1000000))), nil).Times(0)
				appMock.EXPECT().OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).Return(channeltypes.NewResultAcknowledgement([]byte("success")))
			},
			expectedAck: channeltypes.NewResultAcknowledgement([]byte("success")),
			expectError: false,
		},
		{
			name:    "Error - UnlockTokens failure",
			enabled: true,
			packet: channeltypes.Packet{
				Sequence:           1,
				SourcePort:         counterpartyPort,
				SourceChannel:      counterpartyChannel,
				DestinationPort:    nativePort,
				DestinationChannel: nativeChannel,
				Data:               nil,
			},
			data: transfertypes.FungibleTokenPacketData{
				Denom:    moduleDenom,
				Amount:   "1000000000000000000",
				Sender:   sender,
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, packet channeltypes.Packet, keeperMock *testutil.MockTokenwrapperKeeper, transferKeeperMock *testutil.MockTransferKeeper, bankKeeperMock *testutil.MockBankKeeper, appMock *testutil.MockCallbacksCompatibleModule, channelKeeperMock *testutil.MockChannelKeeper, connectionKeeperMock *testutil.MockConnectionKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, nativePort, nativeChannel).Return(channeltypes.Channel{
					State:          channeltypes.OPEN,
					ConnectionHops: []string{nativeConnectionId},
					Counterparty: channeltypes.Counterparty{
						PortId:    counterpartyPort,
						ChannelId: counterpartyChannel,
					},
				}, true).Times(3)
				connectionKeeperMock.EXPECT().GetConnection(ctx, nativeConnectionId).Return(connectiontypes.ConnectionEnd{
					ClientId: nativeClientId,
					Counterparty: connectiontypes.Counterparty{
						ClientId: counterpartyClientId,
					},
				}, true).Times(1)
				keeperMock.EXPECT().IsEnabled(ctx).Return(true).AnyTimes()
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort).AnyTimes()
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel).AnyTimes()
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom).AnyTimes()
				keeperMock.EXPECT().GetNativeClientId(ctx).Return(nativeClientId).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyClientId(ctx).Return(counterpartyClientId).AnyTimes()
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				keeperMock.EXPECT().GetIBCRecvDenom(ctx, moduleDenom).Return(ibcDenom)
				keeperMock.EXPECT().ScaleDownTokenPrecision(ctx, sdkmath.NewInt(1000000000000000000)).Return(sdkmath.NewInt(1000000), nil)
				keeperMock.EXPECT().CheckBalances(ctx, sdk.MustAccAddressFromBech32(receiver), sdkmath.NewInt(1000000000000000000), ibcDenom, sdkmath.NewInt(1000000)).Return(nil)
				keeperMock.EXPECT().LockIBCTokens(ctx, sdk.MustAccAddressFromBech32(receiver), sdkmath.NewInt(1000000000000000000), ibcDenom).Return(sdk.NewCoins(sdk.NewCoin(ibcDenom, sdkmath.NewInt(1000000000000000000))), nil)
				keeperMock.EXPECT().UnlockNativeTokens(
					ctx,
					sdk.MustAccAddressFromBech32(receiver),
					sdkmath.NewInt(1000000),
					sdk.NewCoins(sdk.NewCoin(ibcDenom, sdkmath.NewInt(1000000000000000000)))).
					Return(sdk.Coins{}, fmt.Errorf("failed to unlock tokens:"))
				appMock.EXPECT().OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).Return(channeltypes.NewResultAcknowledgement([]byte("success")))
			},
			expectedAck: channeltypes.NewResultAcknowledgement([]byte("success")),
			expectError: false,
		},
		{
			name:    "Error - app callback failure",
			enabled: true,
			packet: channeltypes.Packet{
				Sequence:           1,
				SourcePort:         counterpartyPort,
				SourceChannel:      counterpartyChannel,
				DestinationPort:    nativePort,
				DestinationChannel: nativeChannel,
				Data:               nil,
			},
			data: transfertypes.FungibleTokenPacketData{
				Denom:    moduleDenom,
				Amount:   "1000000000000000000",
				Sender:   sender,
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, packet channeltypes.Packet, keeperMock *testutil.MockTokenwrapperKeeper, transferKeeperMock *testutil.MockTransferKeeper, bankKeeperMock *testutil.MockBankKeeper, appMock *testutil.MockCallbacksCompatibleModule, channelKeeperMock *testutil.MockChannelKeeper, connectionKeeperMock *testutil.MockConnectionKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, nativePort, nativeChannel).Return(channeltypes.Channel{
					State:          channeltypes.OPEN,
					ConnectionHops: []string{nativeConnectionId},
					Counterparty: channeltypes.Counterparty{
						PortId:    counterpartyPort,
						ChannelId: counterpartyChannel,
					},
				}, true).Times(1)
				connectionKeeperMock.EXPECT().GetConnection(ctx, nativeConnectionId).Return(connectiontypes.ConnectionEnd{
					ClientId: nativeClientId,
					Counterparty: connectiontypes.Counterparty{
						ClientId: counterpartyClientId,
					},
				}, true).Times(0)
				keeperMock.EXPECT().IsEnabled(ctx).Return(true).AnyTimes()
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort).AnyTimes()
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel).AnyTimes()
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom).AnyTimes()
				keeperMock.EXPECT().GetNativeClientId(ctx).Return(nativeClientId).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyClientId(ctx).Return(counterpartyClientId).AnyTimes()
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				appMock.EXPECT().OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).Return(channeltypes.NewErrorAcknowledgement(fmt.Errorf("app callback failed")))
			},
			expectedAck: channeltypes.NewErrorAcknowledgement(fmt.Errorf("app callback failed")),
			expectError: false,
		},
		{
			name:    "Success - large amount",
			enabled: true,
			packet: channeltypes.Packet{
				Sequence:           1,
				SourcePort:         counterpartyPort,
				SourceChannel:      counterpartyChannel,
				DestinationPort:    nativePort,
				DestinationChannel: nativeChannel,
				Data:               nil,
			},
			data: transfertypes.FungibleTokenPacketData{
				Denom:    moduleDenom,
				Amount:   "1000000000000000000000000000000", // Very large amount
				Sender:   sender,
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, packet channeltypes.Packet, keeperMock *testutil.MockTokenwrapperKeeper, transferKeeperMock *testutil.MockTransferKeeper, bankKeeperMock *testutil.MockBankKeeper, appMock *testutil.MockCallbacksCompatibleModule, channelKeeperMock *testutil.MockChannelKeeper, connectionKeeperMock *testutil.MockConnectionKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, nativePort, nativeChannel).Return(channeltypes.Channel{
					State:          channeltypes.OPEN,
					ConnectionHops: []string{nativeConnectionId},
					Counterparty: channeltypes.Counterparty{
						PortId:    counterpartyPort,
						ChannelId: counterpartyChannel,
					},
				}, true).Times(3)
				connectionKeeperMock.EXPECT().GetConnection(ctx, nativeConnectionId).Return(connectiontypes.ConnectionEnd{
					ClientId: nativeClientId,
					Counterparty: connectiontypes.Counterparty{
						ClientId: counterpartyClientId,
					},
				}, true).Times(1)
				keeperMock.EXPECT().IsEnabled(ctx).Return(true).AnyTimes()
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort).AnyTimes()
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel).AnyTimes()
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom).AnyTimes()
				keeperMock.EXPECT().GetNativeClientId(ctx).Return(nativeClientId).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyClientId(ctx).Return(counterpartyClientId).AnyTimes()
				amount, _ := sdkmath.NewIntFromString("1000000000000000000000000000000")
				keeperMock.EXPECT().AddToTotalTransferredIn(ctx, sdkmath.NewInt(1000000000000000000)).Return()
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				keeperMock.EXPECT().GetIBCRecvDenom(ctx, moduleDenom).Return(ibcDenom)
				keeperMock.EXPECT().ScaleDownTokenPrecision(ctx, amount).Return(sdkmath.NewInt(1000000000000000000), nil)
				keeperMock.EXPECT().CheckBalances(ctx, sdk.MustAccAddressFromBech32(receiver), amount, ibcDenom, sdkmath.NewInt(1000000000000000000)).Return(nil)
				keeperMock.EXPECT().LockIBCTokens(ctx, sdk.MustAccAddressFromBech32(receiver), amount, ibcDenom).Return(sdk.NewCoins(sdk.NewCoin(ibcDenom, amount)), nil)
				keeperMock.EXPECT().UnlockNativeTokens(
					ctx,
					sdk.MustAccAddressFromBech32(receiver),
					sdkmath.NewInt(1000000000000000000),
					sdk.NewCoins(sdk.NewCoin(ibcDenom, amount))).
					Return(sdk.NewCoins(sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(1000000000000000000))), nil)
				appMock.EXPECT().OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).Return(channeltypes.NewResultAcknowledgement([]byte("success")))
			},
			expectedAck: channeltypes.NewResultAcknowledgement([]byte("success")),
			expectError: false,
		},
		{
			name:    "No wrapping - empty denom",
			enabled: true,
			packet: channeltypes.Packet{
				Sequence:           1,
				SourcePort:         counterpartyPort,
				SourceChannel:      counterpartyChannel,
				DestinationPort:    nativePort,
				DestinationChannel: nativeChannel,
				Data:               nil,
			},
			data: transfertypes.FungibleTokenPacketData{
				Denom:    "",
				Amount:   "1000000000000000000",
				Sender:   sender,
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, packet channeltypes.Packet, keeperMock *testutil.MockTokenwrapperKeeper, transferKeeperMock *testutil.MockTransferKeeper, bankKeeperMock *testutil.MockBankKeeper, appMock *testutil.MockCallbacksCompatibleModule, channelKeeperMock *testutil.MockChannelKeeper, connectionKeeperMock *testutil.MockConnectionKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, nativePort, nativeChannel).Return(channeltypes.Channel{
					State:          channeltypes.OPEN,
					ConnectionHops: []string{nativeConnectionId},
					Counterparty: channeltypes.Counterparty{
						PortId:    counterpartyPort,
						ChannelId: counterpartyChannel,
					},
				}, true).Times(3)
				connectionKeeperMock.EXPECT().GetConnection(ctx, nativeConnectionId).Return(connectiontypes.ConnectionEnd{
					ClientId: nativeClientId,
					Counterparty: connectiontypes.Counterparty{
						ClientId: counterpartyClientId,
					},
				}, true).Times(1)
				keeperMock.EXPECT().IsEnabled(ctx).Return(true).AnyTimes()
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort).AnyTimes()
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel).AnyTimes()
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom).AnyTimes()
				keeperMock.EXPECT().GetNativeClientId(ctx).Return(nativeClientId).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyClientId(ctx).Return(counterpartyClientId).AnyTimes()
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				appMock.EXPECT().OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).Return(channeltypes.NewResultAcknowledgement([]byte("success")))
			},
			expectedAck: channeltypes.NewResultAcknowledgement([]byte("success")),
			expectError: false,
		},
		{
			name:    "Wrong connections order - should fail",
			enabled: true,
			packet: channeltypes.Packet{
				Sequence:           1,
				SourcePort:         nativePort,
				SourceChannel:      nativeChannel,
				DestinationPort:    counterpartyPort,
				DestinationChannel: counterpartyChannel,
				Data:               nil, // Set in setupMocks
			},
			data: transfertypes.FungibleTokenPacketData{
				Denom:    moduleDenom,
				Amount:   "1000000000000000000", // 1 token with 18 decimals
				Sender:   sender,
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, packet channeltypes.Packet, keeperMock *testutil.MockTokenwrapperKeeper, transferKeeperMock *testutil.MockTransferKeeper, bankKeeperMock *testutil.MockBankKeeper, appMock *testutil.MockCallbacksCompatibleModule, channelKeeperMock *testutil.MockChannelKeeper, connectionKeeperMock *testutil.MockConnectionKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, packet.GetDestPort(), packet.GetDestChannel()).Return(channeltypes.Channel{
					State:          channeltypes.OPEN,
					ConnectionHops: []string{counterpartyConnectionId},
					Counterparty: channeltypes.Counterparty{
						PortId:    counterpartyPort,
						ChannelId: counterpartyChannel,
					},
				}, true).Times(3)
				connectionKeeperMock.EXPECT().GetConnection(ctx, counterpartyConnectionId).Return(connectiontypes.ConnectionEnd{
					ClientId: counterpartyClientId,
					Counterparty: connectiontypes.Counterparty{
						ClientId: counterpartyClientId,
					},
				}, true).Times(1)
				keeperMock.EXPECT().IsEnabled(ctx).Return(true).AnyTimes()
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort).AnyTimes()
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel).AnyTimes()
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom).AnyTimes()
				keeperMock.EXPECT().GetNativeClientId(ctx).Return(nativeClientId).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyClientId(ctx).Return(counterpartyClientId).AnyTimes()
				keeperMock.EXPECT().GetDecimalConversionFactor(ctx).Times(0)
				keeperMock.EXPECT().LockTokens(ctx, sdk.MustAccAddressFromBech32(receiver), sdk.NewCoins(sdk.NewCoin(ibcDenom, sdkmath.NewInt(1000000000000000000)))).Times(0)
				keeperMock.EXPECT().UnlockTokens(ctx, sdk.MustAccAddressFromBech32(receiver), sdk.NewCoins(sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(1000000)))).Times(0)
				keeperMock.EXPECT().AddToTotalTransferredIn(ctx, sdkmath.NewInt(1000000)).Times(0)
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				appMock.EXPECT().OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).Return(channeltypes.NewResultAcknowledgement([]byte("success"))).Times(1)
			},
			expectedAck: channeltypes.NewResultAcknowledgement([]byte("success")),
			expectError: false,
		},
		{
			name:    "Error - native client ID not set",
			enabled: true,
			packet: channeltypes.Packet{
				Sequence:           1,
				SourcePort:         counterpartyPort,
				SourceChannel:      counterpartyChannel,
				DestinationPort:    nativePort,
				DestinationChannel: nativeChannel,
				Data:               nil,
			},
			data: transfertypes.FungibleTokenPacketData{
				Denom:    moduleDenom,
				Amount:   "1000000000000000000",
				Sender:   sender,
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, packet channeltypes.Packet, keeperMock *testutil.MockTokenwrapperKeeper, transferKeeperMock *testutil.MockTransferKeeper, bankKeeperMock *testutil.MockBankKeeper, appMock *testutil.MockCallbacksCompatibleModule, channelKeeperMock *testutil.MockChannelKeeper, connectionKeeperMock *testutil.MockConnectionKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, nativePort, nativeChannel).Return(channeltypes.Channel{
					State:          channeltypes.OPEN,
					ConnectionHops: []string{nativeConnectionId},
					Counterparty: channeltypes.Counterparty{
						PortId:    counterpartyPort,
						ChannelId: counterpartyChannel,
					},
				}, true).Times(2)
				connectionKeeperMock.EXPECT().GetConnection(ctx, nativeConnectionId).Return(connectiontypes.ConnectionEnd{
					ClientId: nativeClientId,
					Counterparty: connectiontypes.Counterparty{
						ClientId: counterpartyClientId,
					},
				}, true).Times(0)
				keeperMock.EXPECT().IsEnabled(ctx).Return(true).AnyTimes()
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort).AnyTimes()
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel).AnyTimes()
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom).AnyTimes()
				keeperMock.EXPECT().GetNativeClientId(ctx).Return("").AnyTimes()
				keeperMock.EXPECT().GetCounterpartyClientId(ctx).Return(counterpartyClientId).AnyTimes()
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()

				// App callback should not be called
				appMock.EXPECT().OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).Return(channeltypes.NewResultAcknowledgement([]byte("success"))).Times(1)
			},
			expectedAck: channeltypes.NewResultAcknowledgement([]byte("success")),
			expectError: false,
		},
		{
			name:    "Error - no connection hops found",
			enabled: true,
			packet: channeltypes.Packet{
				Sequence:           1,
				SourcePort:         counterpartyPort,
				SourceChannel:      counterpartyChannel,
				DestinationPort:    nativePort,
				DestinationChannel: nativeChannel,
				Data:               nil, // Set in setupMocks
			},
			data: transfertypes.FungibleTokenPacketData{
				Denom:    moduleDenom,
				Amount:   "1000000000000000000",
				Sender:   sender,
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, packet channeltypes.Packet, keeperMock *testutil.MockTokenwrapperKeeper, transferKeeperMock *testutil.MockTransferKeeper, bankKeeperMock *testutil.MockBankKeeper, appMock *testutil.MockCallbacksCompatibleModule, channelKeeperMock *testutil.MockChannelKeeper, connectionKeeperMock *testutil.MockConnectionKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, nativePort, nativeChannel).Return(channeltypes.Channel{
					State:          channeltypes.OPEN,
					ConnectionHops: []string{},
				}, true).Times(2)
				connectionKeeperMock.EXPECT().GetConnection(ctx, nativeConnectionId).Return(connectiontypes.ConnectionEnd{
					ClientId: nativeClientId,
				}, true).Times(0)
				connectionKeeperMock.EXPECT().GetConnection(ctx, counterpartyConnectionId).Times(0)
				keeperMock.EXPECT().IsEnabled(ctx).Return(true).AnyTimes()
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort).AnyTimes()
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort).AnyTimes()
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel).AnyTimes()
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom).AnyTimes()
				keeperMock.EXPECT().GetNativeClientId(ctx).Return(nativeClientId).AnyTimes()
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()

				// App callback should not be called
				appMock.EXPECT().OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).Return(channeltypes.NewResultAcknowledgement([]byte("success"))).Times(1)
			},
			expectedAck: channeltypes.NewResultAcknowledgement([]byte("success")),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.enabled {
				fmt.Println("test \"", tt.name, "\" disabled, skipping…")
				return
			}

			// Create fresh mocks for each test case
			keeperMock := testutil.NewMockTokenwrapperKeeper(ctrl)
			transferKeeperMock := testutil.NewMockTransferKeeper(ctrl)
			bankKeeperMock := testutil.NewMockBankKeeper(ctrl)
			appMock := testutil.NewMockCallbacksCompatibleModule(ctrl)
			channelKeeperMock := testutil.NewMockChannelKeeper(ctrl)
			connectionKeeperMock := testutil.NewMockConnectionKeeper(ctrl)

			// Create fresh IBCModule instance for each test case
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

			// Marshal packet data
			var dataBz []byte
			var err error
			if len(tt.dataBz) > 0 {
				dataBz = tt.dataBz
			} else {
				dataBz, err = transfertypes.ModuleCdc.MarshalJSON(&tt.data)
				require.NoError(t, err)
			}
			tt.packet.Data = dataBz

			// Setup mocks
			tt.setupMocks(ctx, tt.packet, keeperMock, transferKeeperMock, bankKeeperMock, appMock, channelKeeperMock, connectionKeeperMock)

			// Call OnRecvPacket
			ack := ibcModule.OnRecvPacket(ctx, "channel-version", tt.packet, sdk.AccAddress{})

			// Verify results
			if tt.expectError {
				require.Nil(t, ack, "Expected nil acknowledgement due to error")
			} else {
				require.NotNil(t, ack, "Expected non-nil acknowledgement")
				if ack.Success() {
					require.Equal(t, tt.expectedAck.Success(), ack.Success(), "Acknowledgement success mismatch")
				} else {
					// For error acknowledgements, compare the error message approximately
					if errAck, ok := tt.expectedAck.(channeltypes.Acknowledgement); ok {
						expectedErr := errAck.GetError()
						actualErr := ack.Acknowledgement()
						// Check if the actual error contains the expected error message
						require.Contains(t, string(actualErr), expectedErr, "Acknowledgement error message mismatch")
					} else {
						require.Equal(t, tt.expectedAck, ack, "Acknowledgement mismatch")
					}
				}
			}
		})
	}
}

func TestOnRecvPacket_InsufficientAccountBalance(t *testing.T) {
	// This test checks the scenario where the account does not have sufficient IBC tokens

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fixture := getRecvPacketPositiveFixture()

	// mocks
	transferKeeperMock := testutil.NewMockTransferKeeper(ctrl)
	bankKeeperMock := testutil.NewMockBankKeeper(ctrl)
	appMock := testutil.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := testutil.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := testutil.NewMockConnectionKeeper(ctrl)

	// keeper setup
	k, ctx := keepertest.TokenwrapperKeeper(t, bankKeeperMock)
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	ibcModule := tokenwrapper.NewIBCModule(
		k, transferKeeperMock, bankKeeperMock, appMock, channelKeeperMock, connectionKeeperMock,
	)

	// packet
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
		SourcePort:         fixture.counterpartyPort,
		SourceChannel:      fixture.counterpartyChannel,
		DestinationPort:    fixture.nativePort,
		DestinationChannel: fixture.nativeChannel,
		Data:               dataBz,
	}

	// channel & connection OK
	validChan := channeltypes.Channel{
		State:          channeltypes.OPEN,
		ConnectionHops: []string{fixture.nativeConnectionId},
		Counterparty: channeltypes.Counterparty{
			PortId:    fixture.counterpartyPort,
			ChannelId: fixture.counterpartyChannel,
		},
	}

	channelKeeperMock.
		EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(validChan, true).
		Times(3)

	connectionKeeperMock.
		EXPECT().
		GetConnection(ctx, fixture.nativeConnectionId).
		Return(connectiontypes.ConnectionEnd{
			ClientId: fixture.nativeClientId,
			Counterparty: connectiontypes.Counterparty{
				ClientId: fixture.counterpartyClientId,
			},
		}, true).
		Times(1)

	// middleware callback must be called so we reach balance check
	appMock.
		EXPECT().
		OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).
		Return(channeltypes.NewResultAcknowledgement([]byte("success")))

	// account has no IBC tokens
	amountInt, _ := sdkmath.NewIntFromString(fixture.amount)
	bankKeeperMock.
		EXPECT().
		HasBalance(ctx, sdk.MustAccAddressFromBech32(fixture.receiver), sdk.NewCoin(fixture.ibcDenom, amountInt)).
		Return(false)

	ack := ibcModule.OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{})
	require.True(t, ack.Success())

	events := ctx.EventManager().Events()
	require.Len(t, events, 1)
	require.Equal(t, types.EventTypeTokenWrapperError, events[0].Type)
	require.Equal(t, events[0].Attributes[0].Value,
		fmt.Sprintf("account %s does not have enough balance of %s",
			fixture.receiver,
			fixture.amount+fixture.ibcDenom,
		))
}

func TestOnRecvPacket_InsufficientModuleBalance(t *testing.T) {
	// Scenario: module has no native-token balance to pay out

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fixture := getRecvPacketPositiveFixture()

	// mocks
	transferKeeperMock := testutil.NewMockTransferKeeper(ctrl)
	bankKeeperMock := testutil.NewMockBankKeeper(ctrl)
	appMock := testutil.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := testutil.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := testutil.NewMockConnectionKeeper(ctrl)

	// keeper setup
	k, ctx := keepertest.TokenwrapperKeeper(t, bankKeeperMock)
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	ibcModule := tokenwrapper.NewIBCModule(
		k, transferKeeperMock, bankKeeperMock, appMock, channelKeeperMock, connectionKeeperMock,
	)

	// packet setup
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
		SourcePort:         fixture.counterpartyPort,
		SourceChannel:      fixture.counterpartyChannel,
		DestinationPort:    fixture.nativePort,
		DestinationChannel: fixture.nativeChannel,
		Data:               dataBz,
	}

	// channel & connection OK
	validChan := channeltypes.Channel{
		State:          channeltypes.OPEN,
		ConnectionHops: []string{fixture.nativeConnectionId},
		Counterparty: channeltypes.Counterparty{
			PortId:    fixture.counterpartyPort,
			ChannelId: fixture.counterpartyChannel,
		},
	}
	channelKeeperMock.
		EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(validChan, true).
		Times(3)

	connectionKeeperMock.
		EXPECT().
		GetConnection(ctx, fixture.nativeConnectionId).
		Return(connectiontypes.ConnectionEnd{
			ClientId: fixture.nativeClientId,
			Counterparty: connectiontypes.Counterparty{
				ClientId: fixture.counterpartyClientId,
			},
		}, true).
		Times(1)

	// middleware callback
	appMock.
		EXPECT().
		OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).
		Return(channeltypes.NewResultAcknowledgement([]byte("success")))

	// account has IBC tokens
	amountInt, _ := sdkmath.NewIntFromString(fixture.amount)
	bankKeeperMock.
		EXPECT().
		HasBalance(ctx, sdk.MustAccAddressFromBech32(fixture.receiver), sdk.NewCoin(fixture.ibcDenom, amountInt)).
		Return(true)

	// module lacks bond-denom
	converted := amountInt.Quo(k.GetDecimalConversionFactor(ctx))
	bankKeeperMock.
		EXPECT().
		HasBalance(ctx, authtypes.NewModuleAddress(types.ModuleName), sdk.NewCoin(constants.BondDenom, converted)).
		Return(false)

	ack := ibcModule.OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{})
	require.True(t, ack.Success())

	events := ctx.EventManager().Events()
	require.Len(t, events, 1)
	require.Equal(t, types.EventTypeTokenWrapperError, events[0].Type)
	require.Equal(t, events[0].Attributes[0].Value,
		fmt.Sprintf("module does not have enough balance of %s%s",
			converted.String(),
			constants.BondDenom,
		),
	)
}

func TestOnRecvPacket_ConnectionNotFound(t *testing.T) {
	// Scenario: IBC connection lookup fails

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fixture := getRecvPacketPositiveFixture()

	// mocks
	transferKeeperMock := testutil.NewMockTransferKeeper(ctrl)
	bankKeeperMock := testutil.NewMockBankKeeper(ctrl)
	appMock := testutil.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := testutil.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := testutil.NewMockConnectionKeeper(ctrl)

	// keeper setup
	k, ctx := keepertest.TokenwrapperKeeper(t, bankKeeperMock)
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	ibcModule := tokenwrapper.NewIBCModule(
		k, transferKeeperMock, bankKeeperMock, appMock, channelKeeperMock, connectionKeeperMock,
	)

	// packet setup
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
		SourcePort:         fixture.counterpartyPort,
		SourceChannel:      fixture.counterpartyChannel,
		DestinationPort:    fixture.nativePort,
		DestinationChannel: fixture.nativeChannel,
		Data:               dataBz,
	}

	// channel OK
	validChan := channeltypes.Channel{
		State:          channeltypes.OPEN,
		ConnectionHops: []string{fixture.nativeConnectionId},
		Counterparty: channeltypes.Counterparty{
			PortId:    fixture.counterpartyPort,
			ChannelId: fixture.counterpartyChannel,
		},
	}
	channelKeeperMock.
		EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(validChan, true).
		Times(3)

	// connection lookup fails
	connectionKeeperMock.
		EXPECT().
		GetConnection(ctx, fixture.nativeConnectionId).
		Return(connectiontypes.ConnectionEnd{
			ClientId: fixture.nativeClientId,
			Counterparty: connectiontypes.Counterparty{
				ClientId: fixture.counterpartyClientId,
			},
		}, false).
		Times(1)

	appMock.EXPECT().
		OnRecvPacket(ctx, "channel-version", gomock.Any(), sdk.AccAddress{}).
		Return(channeltypes.NewResultAcknowledgement([]byte("success"))).
		Times(1)

	ack := ibcModule.OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{})
	require.True(t, ack.Success())

	events := ctx.EventManager().Events()
	require.Len(t, events, 1)
	require.Equal(t, types.EventTypeTokenWrapperError, events[0].Type)
	require.Contains(t, events[0].Attributes[0].Value, "client state validation failed")
}

func TestOnRecvPacket_RealKeepers_Negative_InvalidDataFields(t *testing.T) {
	testCases := []struct {
		name           string
		denom          string
		amount         string
		sender         string
		receiver       string
		expectedErrMsg string
		description    string
	}{
		// ================================
		// AMOUNT FIELD TESTS (Errors Only)
		// ================================
		{
			name:           "Amount_Field_Missing_Should_Error",
			denom:          "axl",
			sender:         "axelar15yk64u7zc9g9k2yr2wmzeva5qgwxps6yzue8jl",
			receiver:       "zig1fcf3p6exuhq53myh0k3wweym5ffv00gcnuum07",
			expectedErrMsg: "invalid amount",
			description:    "Missing amount field should cause parsing error",
		},
		{
			name:           "Amount_Field_Empty_Should_Error",
			denom:          "axl",
			amount:         "",
			sender:         "axelar15yk64u7zc9g9k2yr2wmzeva5qgwxps6yzue8jl",
			receiver:       "zig1fcf3p6exuhq53myh0k3wweym5ffv00gcnuum07",
			expectedErrMsg: "invalid amount",
			description:    "Empty amount should cause parsing error",
		},
		{
			name:           "Amount_Field_Whitespace_Only_Should_Error",
			denom:          "axl",
			amount:         "   ",
			sender:         "axelar15yk64u7zc9g9k2yr2wmzeva5qgwxps6yzue8jl",
			receiver:       "zig1fcf3p6exuhq53myh0k3wweym5ffv00gcnuum07",
			expectedErrMsg: "invalid amount",
			description:    "Whitespace-only amount should cause parsing error",
		},
		{
			name:           "Amount_Field_Zero_Should_Error",
			denom:          "axl",
			amount:         "0",
			sender:         "axelar15yk64u7zc9g9k2yr2wmzeva5qgwxps6yzue8jl",
			receiver:       "zig1fcf3p6exuhq53myh0k3wweym5ffv00gcnuum07",
			expectedErrMsg: "amount is zero or negative: 0",
			description:    "Zero amount should cause conversion validation error",
		},
		{
			name:           "Amount_Field_Negative_Should_Error",
			denom:          "axl",
			amount:         "-1000000000000000000",
			sender:         "axelar15yk64u7zc9g9k2yr2wmzeva5qgwxps6yzue8jl",
			receiver:       "zig1fcf3p6exuhq53myh0k3wweym5ffv00gcnuum07",
			expectedErrMsg: "amount is zero or negative: -1000000",
			description:    "Negative amount should cause parsing error",
		},
		{
			name:           "Amount_Field_Scientific_Notation_Should_Error",
			denom:          "axl",
			amount:         "1e18",
			sender:         "axelar15yk64u7zc9g9k2yr2wmzeva5qgwxps6yzue8jl",
			receiver:       "zig1fcf3p6exuhq53myh0k3wweym5ffv00gcnuum07",
			expectedErrMsg: "invalid amount",
			description:    "Scientific notation amount should cause parsing error",
		},
		{
			name:           "Amount_Field_With_Decimal_Point_Should_Error",
			denom:          "axl",
			amount:         "1000.5",
			sender:         "axelar15yk64u7zc9g9k2yr2wmzeva5qgwxps6yzue8jl",
			receiver:       "zig1fcf3p6exuhq53myh0k3wweym5ffv00gcnuum07",
			expectedErrMsg: "invalid amount",
			description:    "Amount with decimal point should cause parsing error",
		},
		{
			name:           "Amount_Field_With_Leading_Spaces_Should_Error",
			denom:          "axl",
			amount:         " 1000000000000000000",
			sender:         "axelar15yk64u7zc9g9k2yr2wmzeva5qgwxps6yzue8jl",
			receiver:       "zig1fcf3p6exuhq53myh0k3wweym5ffv00gcnuum07",
			expectedErrMsg: "invalid amount",
			description:    "Amount with leading spaces should cause parsing error",
		},
		{
			name:           "Amount_Field_With_Trailing_Spaces_Should_Error",
			denom:          "axl",
			amount:         "1000000000000000000 ",
			sender:         "axelar15yk64u7zc9g9k2yr2wmzeva5qgwxps6yzue8jl",
			receiver:       "zig1fcf3p6exuhq53myh0k3wweym5ffv00gcnuum07",
			expectedErrMsg: "invalid amount",
			description:    "Amount with trailing spaces should cause parsing error",
		},
		{
			name:           "Amount_Field_With_Special_Characters_Should_Error",
			denom:          "axl",
			amount:         "1000#000",
			sender:         "axelar15yk64u7zc9g9k2yr2wmzeva5qgwxps6yzue8jl",
			receiver:       "zig1fcf3p6exuhq53myh0k3wweym5ffv00gcnuum07",
			expectedErrMsg: "invalid amount",
			description:    "Amount with special characters should cause parsing error",
		},
		{
			name:           "Amount_Field_With_Unicode_Should_Error",
			denom:          "axl",
			amount:         "100０", // Contains Unicode zero
			sender:         "axelar15yk64u7zc9g9k2yr2wmzeva5qgwxps6yzue8jl",
			receiver:       "zig1fcf3p6exuhq53myh0k3wweym5ffv00gcnuum07",
			expectedErrMsg: "invalid amount",
			description:    "Amount with Unicode characters should cause parsing error",
		},

		// ================================
		// SENDER FIELD TESTS (Errors Only)
		// ================================
		{
			name:           "Sender_Field_Missing_Should_Error",
			denom:          "axl",
			amount:         "1000000000000000000",
			sender:         "", // Missing/empty sender
			receiver:       "zig1fcf3p6exuhq53myh0k3wweym5ffv00gcnuum07",
			expectedErrMsg: "sender address is empty",
			description:    "Missing sender field should cause validation error",
		},
		{
			name:           "Sender_Field_Empty_Should_Error",
			denom:          "axl",
			amount:         "1000000000000000000",
			sender:         "",
			receiver:       "zig1fcf3p6exuhq53myh0k3wweym5ffv00gcnuum07",
			expectedErrMsg: "sender address is empty",
			description:    "Empty sender should cause validation error",
		},

		// ================================
		// RECEIVER FIELD TESTS (Errors Only)
		// ================================
		{
			name:           "Receiver_Field_Missing_Should_Error",
			denom:          "axl",
			amount:         "1000000000000000000",
			sender:         "axelar15yk64u7zc9g9k2yr2wmzeva5qgwxps6yzue8jl",
			expectedErrMsg: "empty address string is not allowed",
			description:    "Missing receiver field should cause address parsing error",
		},
		{
			name:           "Receiver_Field_Empty_Should_Error",
			denom:          "axl",
			amount:         "1000000000000000000",
			sender:         "axelar15yk64u7zc9g9k2yr2wmzeva5qgwxps6yzue8jl",
			receiver:       "",
			expectedErrMsg: "empty address string is not allowed",
			description:    "Empty receiver should cause address parsing error",
		},
		{
			name:           "Receiver_Field_Whitespace_Only_Should_Error",
			denom:          "axl",
			amount:         "1000000000000000000",
			sender:         "axelar15yk64u7zc9g9k2yr2wmzeva5qgwxps6yzue8jl",
			receiver:       "   ",
			expectedErrMsg: "empty address string is not allowed",
			description:    "Whitespace-only receiver should fail bech32 validation",
		},
		{
			name:           "Receiver_Field_Invalid_Address_Format_Should_Error",
			denom:          "axl",
			amount:         "1000000000000000000",
			sender:         "axelar15yk64u7zc9g9k2yr2wmzeva5qgwxps6yzue8jl",
			receiver:       "invalid-address",
			expectedErrMsg: "decoding bech32 failed: invalid separator index -1",
			description:    "Invalid receiver address should fail bech32 validation",
		},
		{
			name:           "Receiver_Field_Wrong_Prefix_Should_Error",
			denom:          "axl",
			amount:         "1000000000000000000",
			sender:         "axelar15yk64u7zc9g9k2yr2wmzeva5qgwxps6yzue8jl",
			receiver:       "cosmos1fcf3p6exuhq53myh0k3wweym5ffv00gcnuum07",
			expectedErrMsg: "decoding bech32 failed: invalid checksum (expected jkefdk got nuum07)",
			description:    "Receiver with wrong prefix should fail validation",
		},
		{
			name:           "Receiver_Field_Invalid_Checksum_Should_Error",
			denom:          "axl",
			amount:         "1000000000000000000",
			sender:         "axelar15yk64u7zc9g9k2yr2wmzeva5qgwxps6yzue8jl",
			receiver:       "zig1fcf3p6exuhq53myh0k3wweym5ffv00gcnuum00",
			expectedErrMsg: "decoding bech32 failed: invalid checksum (expected nuum07 got nuum00)",
			description:    "Receiver with invalid checksum should fail bech32 validation",
		},
		{
			name:           "Receiver_Field_Too_Short_Should_Error",
			denom:          "axl",
			amount:         "1000000000000000000",
			sender:         "axelar15yk64u7zc9g9k2yr2wmzeva5qgwxps6yzue8jl",
			receiver:       "zig1abc",
			expectedErrMsg: "invalid bech32",
			description:    "Receiver too short should fail bech32 validation",
		},
		{
			name:           "Receiver_Field_With_Special_Characters_Should_Error",
			denom:          "axl",
			amount:         "1000000000000000000",
			sender:         "axelar15yk64u7zc9g9k2yr2wmzeva5qgwxps6yzue8jl",
			receiver:       "zig1fcf3p6exuhq53myh0k3wweym5ffv00gcnu#um07",
			expectedErrMsg: "decoding bech32 failed: invalid character not part of charset: 35",
			description:    "Receiver with special characters should fail bech32 validation",
		},
		{
			name:           "Receiver_Field_With_Leading_Spaces_Should_Error",
			denom:          "axl",
			amount:         "1000000000000000000",
			sender:         "axelar15yk64u7zc9g9k2yr2wmzeva5qgwxps6yzue8jl",
			receiver:       " zig1fcf3p6exuhq53myh0k3wweym5ffv00gcnuum07",
			expectedErrMsg: "decoding bech32 failed: invalid character in string: ' '",
			description:    "Receiver with leading spaces should fail bech32 validation",
		},
		{
			name:           "Receiver_Field_With_Trailing_Spaces_Should_Error",
			denom:          "axl",
			amount:         "1000000000000000000",
			sender:         "axelar15yk64u7zc9g9k2yr2wmzeva5qgwxps6yzue8jl",
			receiver:       "zig1fcf3p6exuhq53myh0k3wweym5ffv00gcnuum07 ",
			expectedErrMsg: "decoding bech32 failed: invalid character in string: ' '",
			description:    "Receiver with trailing spaces should fail bech32 validation",
		},
		{
			name:           "Receiver_Field_All_Zeros_Should_Error",
			denom:          "axl",
			amount:         "1000000000000000000",
			sender:         "axelar15yk64u7zc9g9k2yr2wmzeva5qgwxps6yzue8jl",
			receiver:       "zig1000000000000000000000000000000000000000",
			expectedErrMsg: "decoding bech32 failed: invalid checksum (expected srpg7j got 000000)",
			description:    "Receiver with all zeros should fail bech32 validation",
		},
		{
			name:           "Receiver_Field_Extremely_Long_Should_Error",
			denom:          "axl",
			amount:         "1000000000000000000",
			sender:         "axelar15yk64u7zc9g9k2yr2wmzeva5qgwxps6yzue8jl",
			receiver:       "zig1" + string(make([]byte, 200)), // Very long string
			expectedErrMsg: "decoding bech32 failed: invalid character in string: ",
			description:    "Extremely long receiver should fail bech32 validation",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			fixture := getRecvPacketPositiveFixture()
			k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

			// Configure keeper
			k.SetEnabled(ctx, true)
			k.SetNativePort(ctx, fixture.nativePort)
			k.SetNativeChannel(ctx, fixture.nativeChannel)
			k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
			k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
			k.SetDenom(ctx, fixture.moduleDenom) // "axl"
			k.SetNativeClientId(ctx, fixture.nativeClientId)
			k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
			_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

			// Create IBC mocks
			transferKeeperMock := testutil.NewMockTransferKeeper(ctrl)
			appMock := testutil.NewMockCallbacksCompatibleModule(ctrl)
			channelKeeperMock := testutil.NewMockChannelKeeper(ctrl)
			connectionKeeperMock := testutil.NewMockConnectionKeeper(ctrl)

			ibcModule := tokenwrapper.NewIBCModule(
				k, transferKeeperMock, nil, appMock, channelKeeperMock, connectionKeeperMock,
			)

			// Setup balances (even though processing will fail)
			receiverAddr := sdk.MustAccAddressFromBech32(fixture.receiver)
			moduleAddr := authtypes.NewModuleAddress(types.ModuleName)

			amountInt, _ := sdkmath.NewIntFromString(fixture.amount)
			conversionFactor := k.GetDecimalConversionFactor(ctx)
			convertedAmount := amountInt.Quo(conversionFactor)

			// Give receiver IBC tokens
			ibcCoins := sdk.NewCoins(sdk.NewCoin(fixture.ibcDenom, amountInt))
			require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, ibcCoins))
			require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, receiverAddr, ibcCoins))

			// Give module native tokens
			nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, convertedAmount))
			require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, nativeCoins))
			require.NoError(t, bankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, nativeCoins))

			// Record initial state
			initialReceiverIBC := bankKeeper.GetBalance(ctx, receiverAddr, fixture.ibcDenom)
			initialReceiverNative := bankKeeper.GetBalance(ctx, receiverAddr, constants.BondDenom)
			initialModuleIBC := bankKeeper.GetBalance(ctx, moduleAddr, fixture.ibcDenom)
			initialModuleNative := bankKeeper.GetBalance(ctx, moduleAddr, constants.BondDenom)
			initialTotalTransferredIn := k.GetTotalTransferredIn(ctx)
			initialTotalTransferredOut := k.GetTotalTransferredOut(ctx)

			// Setup IBC mocks (validation succeeds, errors happen during processing)
			channelKeeperMock.EXPECT().
				GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
				Return(channeltypes.Channel{
					State:          channeltypes.OPEN,
					ConnectionHops: []string{fixture.nativeConnectionId},
					Counterparty: channeltypes.Counterparty{
						PortId:    fixture.counterpartyPort,
						ChannelId: fixture.counterpartyChannel,
					},
				}, true).Times(1)

			connectionKeeperMock.EXPECT().
				GetConnection(ctx, fixture.nativeConnectionId).
				Return(connectiontypes.ConnectionEnd{ClientId: fixture.nativeClientId}, true).Times(0)

			appMock.EXPECT().
				OnRecvPacket(ctx, "channel-version", gomock.Any(), sdk.AccAddress{}).
				Return(channeltypes.NewResultAcknowledgement([]byte("success"))).Times(0)

			// Create and execute packet
			data := transfertypes.FungibleTokenPacketData{
				Denom:    tc.denom,
				Amount:   tc.amount,
				Sender:   tc.sender,
				Receiver: tc.receiver,
			}
			dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
			require.NoError(t, err)

			packet := channeltypes.Packet{
				Sequence:           1,
				SourcePort:         fixture.counterpartyPort,
				SourceChannel:      fixture.counterpartyChannel,
				DestinationPort:    fixture.nativePort,
				DestinationChannel: fixture.nativeChannel,
				Data:               dataBz,
			}

			// Execute
			ack := ibcModule.OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{})

			// Verify ERROR occurred
			require.False(t, ack.Success(), "Packet should fail: %s", tc.description)

			// Verify error event was emitted
			events := ctx.EventManager().Events()
			var errorEvent *sdk.Event
			for i := range events {
				if events[i].Type == types.EventTypeTokenWrapperError {
					errorEvent = &events[i]
					break
				}
			}
			require.NotNil(t, errorEvent, "Should emit error event")

			if tc.expectedErrMsg != "" {
				require.Contains(t, errorEvent.Attributes[0].Value, tc.expectedErrMsg,
					"Error message should contain expected text")
			}

			// Verify ALL balances remained unchanged (error occurred, no processing)
			finalReceiverIBC := bankKeeper.GetBalance(ctx, receiverAddr, fixture.ibcDenom)
			finalReceiverNative := bankKeeper.GetBalance(ctx, receiverAddr, constants.BondDenom)
			finalModuleIBC := bankKeeper.GetBalance(ctx, moduleAddr, fixture.ibcDenom)
			finalModuleNative := bankKeeper.GetBalance(ctx, moduleAddr, constants.BondDenom)
			finalTotalTransferredIn := k.GetTotalTransferredIn(ctx)
			finalTotalTransferredOut := k.GetTotalTransferredOut(ctx)

			require.Equal(t, initialReceiverIBC, finalReceiverIBC,
				"Receiver IBC balance should be unchanged on error")
			require.Equal(t, initialReceiverNative, finalReceiverNative,
				"Receiver native balance should be unchanged on error")
			require.Equal(t, initialModuleIBC, finalModuleIBC,
				"Module IBC balance should be unchanged on error")
			require.Equal(t, initialModuleNative, finalModuleNative,
				"Module native balance should be unchanged on error")
			require.Equal(t, initialTotalTransferredIn, finalTotalTransferredIn,
				"TotalTransferredIn should be unchanged on error")
			require.Equal(t, initialTotalTransferredOut, finalTotalTransferredOut,
				"TotalTransferredOut should be unchanged on error")

		})
	}
}

func TestOnRecvPacket_UnmarshalError_InvalidJSON(t *testing.T) {
	// Test case: Packet data cannot be unmarshalled due to invalid JSON format

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keeperMock := testutil.NewMockTokenwrapperKeeper(ctrl)
	transferKeeperMock := testutil.NewMockTransferKeeper(ctrl)
	bankKeeperMock := testutil.NewMockBankKeeper(ctrl)
	appMock := testutil.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := testutil.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := testutil.NewMockConnectionKeeper(ctrl)

	ibcModule := tokenwrapper.NewIBCModule(
		keeperMock,
		transferKeeperMock,
		bankKeeperMock,
		appMock,
		channelKeeperMock,
		connectionKeeperMock,
	)

	eventManager := sdk.NewEventManager()
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger()).WithEventManager(eventManager)

	packet := channeltypes.Packet{
		Data:               []byte("invalid json"),
		DestinationPort:    "transfer",
		DestinationChannel: "channel-12",
	}

	channelKeeperMock.EXPECT().GetChannel(ctx, "transfer", "channel-12").Return(channeltypes.Channel{State: channeltypes.OPEN}, true).Times(1)
	keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()

	ack := ibcModule.OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{})

	require.NotNil(t, ack)
	require.False(t, ack.Success())
	errAck, ok := ack.(channeltypes.Acknowledgement)
	require.True(t, ok)
	require.Contains(t, string(errAck.Acknowledgement()), "ABCI code: 1: error handling packet: see events for details")

	events := eventManager.Events()
	require.Len(t, events, 1)
	require.Equal(t, types.EventTypeTokenWrapperError, events[0].Type)
	require.Contains(t, events[0].Attributes[0].Value, "failed to unmarshal packet data: invalid character")
}

func TestOnRecvPacket_UnmarshalError_EmptyData(t *testing.T) {
	// Test case: Packet data is empty
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keeperMock := testutil.NewMockTokenwrapperKeeper(ctrl)
	transferKeeperMock := testutil.NewMockTransferKeeper(ctrl)
	bankKeeperMock := testutil.NewMockBankKeeper(ctrl)
	appMock := testutil.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := testutil.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := testutil.NewMockConnectionKeeper(ctrl)

	ibcModule := tokenwrapper.NewIBCModule(
		keeperMock,
		transferKeeperMock,
		bankKeeperMock,
		appMock,
		channelKeeperMock,
		connectionKeeperMock,
	)

	eventManager := sdk.NewEventManager()
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger()).WithEventManager(eventManager)

	packet := channeltypes.Packet{
		Data:               []byte{},
		DestinationPort:    "transfer",
		DestinationChannel: "channel-12",
	}

	channelKeeperMock.EXPECT().GetChannel(ctx, "transfer", "channel-12").Return(channeltypes.Channel{State: channeltypes.OPEN}, true).Times(1)
	keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()

	ack := ibcModule.OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{})

	require.NotNil(t, ack)
	require.False(t, ack.Success())
	errAck, ok := ack.(channeltypes.Acknowledgement)
	require.True(t, ok)
	require.Contains(t, string(errAck.Acknowledgement()), "ABCI code: 1: error handling packet: see events for details")

	events := eventManager.Events()
	require.Len(t, events, 1)
	require.Equal(t, types.EventTypeTokenWrapperError, events[0].Type)
	require.Contains(t, events[0].Attributes[0].Value, "failed to unmarshal packet data: EOF")
}

func TestOnRecvPacket_UnmarshalError_NonJSONData(t *testing.T) {
	// Test case: Packet data is not valid JSON (e.g., plain text)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keeperMock := testutil.NewMockTokenwrapperKeeper(ctrl)
	transferKeeperMock := testutil.NewMockTransferKeeper(ctrl)
	bankKeeperMock := testutil.NewMockBankKeeper(ctrl)
	appMock := testutil.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := testutil.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := testutil.NewMockConnectionKeeper(ctrl)

	ibcModule := tokenwrapper.NewIBCModule(
		keeperMock,
		transferKeeperMock,
		bankKeeperMock,
		appMock,
		channelKeeperMock,
		connectionKeeperMock,
	)

	eventManager := sdk.NewEventManager()
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger()).WithEventManager(eventManager)

	packet := channeltypes.Packet{
		Data:               []byte("not valid json"),
		DestinationPort:    "transfer",
		DestinationChannel: "channel-12",
	}

	channelKeeperMock.EXPECT().GetChannel(ctx, "transfer", "channel-12").Return(channeltypes.Channel{State: channeltypes.OPEN}, true).Times(1)
	keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()

	ack := ibcModule.OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{})

	require.NotNil(t, ack)
	require.False(t, ack.Success())
	errAck, ok := ack.(channeltypes.Acknowledgement)
	require.True(t, ok)
	require.Contains(t, string(errAck.Acknowledgement()), "ABCI code: 1: error handling packet: see events for details")

	events := eventManager.Events()
	require.Len(t, events, 1)
	require.Equal(t, types.EventTypeTokenWrapperError, events[0].Type)
	require.Contains(t, events[0].Attributes[0].Value, "failed to unmarshal packet data: invalid character")
}

func TestOnRecvPacket_RealKeepers_PacketDenomNotModuleDenom(t *testing.T) {
	// Test case: Packet denom does not match the module denom

	// set up positive fixture
	fixture := getRecvPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := testutil.NewMockTransferKeeper(ctrl)
	appMock := testutil.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := testutil.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := testutil.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k, transferKeeperMock, nil, appMock, channelKeeperMock, connectionKeeperMock,
	)

	// Create the data and the packet with prefixed denom
	prefixedDenom := fmt.Sprintf("%s/%s/%s", fixture.counterpartyPort, fixture.counterpartyChannel, fixture.moduleDenom)
	data := transfertypes.FungibleTokenPacketData{
		Denom:    prefixedDenom,
		Amount:   fixture.amount,
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         fixture.counterpartyPort,
		SourceChannel:      fixture.counterpartyChannel,
		DestinationPort:    fixture.nativePort,
		DestinationChannel: fixture.nativeChannel,
		Data:               dataBz,
	}

	// Set up channel and connection mocks
	channelKeeperMock.
		EXPECT().
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
		Times(3)

	connectionKeeperMock.EXPECT().
		GetConnection(ctx, fixture.nativeConnectionId).
		Return(connectiontypes.ConnectionEnd{
			ClientId: fixture.nativeClientId,
			Counterparty: connectiontypes.Counterparty{
				ClientId: fixture.counterpartyClientId,
			},
		}, true).
		Times(1)

	appMock.
		EXPECT().
		OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).
		Return(channeltypes.NewResultAcknowledgement([]byte("success")))

	// Set up initial balances
	receiverAddr := sdk.MustAccAddressFromBech32(fixture.receiver)
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)

	// Mint IBC tokens to receiver
	amountInt, _ := sdkmath.NewIntFromString(fixture.amount)
	ibcCoins := sdk.NewCoins(sdk.NewCoin(fixture.ibcDenom, amountInt))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, ibcCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, receiverAddr, ibcCoins))

	// Mint native tokens to module
	conversionFactor := k.GetDecimalConversionFactor(ctx)
	convertedAmount := amountInt.Quo(conversionFactor)
	nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, convertedAmount))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, nativeCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, nativeCoins))

	// Check initial balances
	receiverNativeBalanceBefore := bankKeeper.GetBalance(ctx, receiverAddr, constants.BondDenom)
	require.Equal(t, sdkmath.ZeroInt(), receiverNativeBalanceBefore.Amount)

	receiverIBCBalanceBefore := bankKeeper.GetBalance(ctx, receiverAddr, fixture.ibcDenom)
	require.Equal(t, amountInt, receiverIBCBalanceBefore.Amount)

	moduleNativeBalanceBefore := bankKeeper.GetBalance(ctx, moduleAddr, constants.BondDenom)
	require.Equal(t, convertedAmount, moduleNativeBalanceBefore.Amount)

	moduleIBCBalanceBefore := bankKeeper.GetBalance(ctx, moduleAddr, fixture.ibcDenom)
	require.Equal(t, sdkmath.ZeroInt(), moduleIBCBalanceBefore.Amount)

	// Execute OnRecvPacket
	respOnRecvPacket := ibcModule.OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{})

	// Validate that no error occurred
	require.NotNil(t, respOnRecvPacket)
	require.True(t, respOnRecvPacket.Success())

	// Check final balances unchanged
	receiverNativeBalanceAfter := bankKeeper.GetBalance(ctx, receiverAddr, constants.BondDenom)
	require.Equal(t, sdkmath.ZeroInt(), receiverNativeBalanceAfter.Amount)

	receiverIBCBalanceAfter := bankKeeper.GetBalance(ctx, receiverAddr, fixture.ibcDenom)
	require.Equal(t, amountInt, receiverIBCBalanceAfter.Amount)

	moduleNativeBalanceAfter := bankKeeper.GetBalance(ctx, moduleAddr, constants.BondDenom)
	require.Equal(t, convertedAmount, moduleNativeBalanceAfter.Amount)

	moduleIBCBalanceAfter := bankKeeper.GetBalance(ctx, moduleAddr, fixture.ibcDenom)
	require.Equal(t, sdkmath.ZeroInt(), moduleIBCBalanceAfter.Amount)

	// Check TotalTransferredIn
	totalTransferredIn := k.GetTotalTransferredIn(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredIn)

	// Check TotalTransferredOut
	totalTransferredOut := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOut)

	// Check the info event
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// Find the info event
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventTypeTokenWrapperInfo {
			foundEvent = true
			// Check event attributes
			require.Equal(t, "info", event.Attributes[0].Key)
			require.Equal(t, "packet denom is not the module denom, skipping wrapping: transfer/channel-24/axl", event.Attributes[0].Value)
			break
		}
	}

	require.True(t, foundEvent, "Expected EventTypeTokenWrapperInfo event to be emitted")
}

func TestOnRecvPacket_RealKeepers_Positive_SenderNotSourceChain(t *testing.T) {
	// Test case: OnRecvPacket with real keepers to verify skipping wrapping when sender chain is not the source chain (denom has prefix but module denom is set to prefixed)

	// set up positive fixture
	fixture := getRecvPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set module denom to prefixed to trigger the condition
	prefixedDenom := fmt.Sprintf("%s/%s/%s", fixture.counterpartyPort, fixture.counterpartyChannel, fixture.moduleDenom)
	k.SetDenom(ctx, prefixedDenom)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := testutil.NewMockTransferKeeper(ctrl)
	appMock := testutil.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := testutil.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := testutil.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k, transferKeeperMock, nil, appMock, channelKeeperMock, connectionKeeperMock,
	)

	// Create the data and the packet with prefixed denom
	data := transfertypes.FungibleTokenPacketData{
		Denom:    prefixedDenom,
		Amount:   fixture.amount,
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         fixture.counterpartyPort,
		SourceChannel:      fixture.counterpartyChannel,
		DestinationPort:    fixture.nativePort,
		DestinationChannel: fixture.nativeChannel,
		Data:               dataBz,
	}

	// Set up channel and connection mocks
	channelKeeperMock.
		EXPECT().
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
		Times(3)

	connectionKeeperMock.EXPECT().
		GetConnection(ctx, fixture.nativeConnectionId).
		Return(connectiontypes.ConnectionEnd{
			ClientId: fixture.nativeClientId,
			Counterparty: connectiontypes.Counterparty{
				ClientId: fixture.counterpartyClientId,
			},
		}, true).
		Times(1)

	appMock.
		EXPECT().
		OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).
		Return(channeltypes.NewResultAcknowledgement([]byte("success")))

	// Set up initial balances
	receiverAddr := sdk.MustAccAddressFromBech32(fixture.receiver)
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)

	// Compute ibc denom for prefixed
	prefixedForHash := fmt.Sprintf("%s/%s/%s", fixture.nativePort, fixture.nativeChannel, prefixedDenom)
	hashBytes := sha256.Sum256([]byte(prefixedForHash))
	hexString := hex.EncodeToString(hashBytes[:])
	ibcHash := "ibc/" + strings.ToUpper(hexString)
	amountInt, _ := sdkmath.NewIntFromString(fixture.amount)

	// Mint IBC tokens to receiver
	ibcCoins := sdk.NewCoins(sdk.NewCoin(ibcHash, amountInt))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, ibcCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, receiverAddr, ibcCoins))

	// Mint native tokens to module
	conversionFactor := k.GetDecimalConversionFactor(ctx)
	convertedAmount := amountInt.Quo(conversionFactor)
	nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, convertedAmount))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, nativeCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, nativeCoins))

	// Check initial balances
	receiverNativeBalanceBefore := bankKeeper.GetBalance(ctx, receiverAddr, constants.BondDenom)
	require.Equal(t, sdkmath.ZeroInt(), receiverNativeBalanceBefore.Amount)

	receiverIBCBalanceBefore := bankKeeper.GetBalance(ctx, receiverAddr, ibcHash)
	require.Equal(t, amountInt, receiverIBCBalanceBefore.Amount)

	moduleNativeBalanceBefore := bankKeeper.GetBalance(ctx, moduleAddr, constants.BondDenom)
	require.Equal(t, convertedAmount, moduleNativeBalanceBefore.Amount)

	moduleIBCBalanceBefore := bankKeeper.GetBalance(ctx, moduleAddr, ibcHash)
	require.Equal(t, sdkmath.ZeroInt(), moduleIBCBalanceBefore.Amount)

	// Execute OnRecvPacket
	respOnRecvPacket := ibcModule.OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{})

	// Validate that no error occurred
	require.NotNil(t, respOnRecvPacket)
	require.True(t, respOnRecvPacket.Success())

	// Check final balances unchanged
	receiverNativeBalanceAfter := bankKeeper.GetBalance(ctx, receiverAddr, constants.BondDenom)
	require.Equal(t, sdkmath.ZeroInt(), receiverNativeBalanceAfter.Amount)

	receiverIBCBalanceAfter := bankKeeper.GetBalance(ctx, receiverAddr, ibcHash)
	require.Equal(t, amountInt, receiverIBCBalanceAfter.Amount)

	moduleNativeBalanceAfter := bankKeeper.GetBalance(ctx, moduleAddr, constants.BondDenom)
	require.Equal(t, convertedAmount, moduleNativeBalanceAfter.Amount)

	moduleIBCBalanceAfter := bankKeeper.GetBalance(ctx, moduleAddr, ibcHash)
	require.Equal(t, sdkmath.ZeroInt(), moduleIBCBalanceAfter.Amount)

	// Check TotalTransferredIn
	totalTransferredIn := k.GetTotalTransferredIn(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredIn)

	// Check TotalTransferredOut
	totalTransferredOut := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredOut)

	// Check the info event
	events := ctx.EventManager().Events()
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventTypeTokenWrapperInfo {
			foundEvent = true
			require.Contains(t, event.Attributes[0].Value, "sender chain is not the source chain, skipping wrapping")
			break
		}
	}
	require.True(t, foundEvent, "Expected EventTypeTokenWrapperInfo event for sender not source chain")
}

func TestOnRecvPacket_EmitsTokenConversionErrorEvent(t *testing.T) {
	// Test case: Emits error event when token conversion fails due to scaling down to zero

	// Set up positive fixture
	fixture := getRecvPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper with a large decimal difference to cause scaling issues
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.counterpartyPort)
	k.SetCounterpartyChannel(ctx, fixture.counterpartyChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, 18) // Large decimal difference to cause scaling to zero

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := testutil.NewMockTransferKeeper(ctrl)
	appMock := testutil.NewMockCallbacksCompatibleModule(ctrl)
	channelKeeperMock := testutil.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := testutil.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ibcModule := tokenwrapper.NewIBCModule(
		k, transferKeeperMock, nil, appMock, channelKeeperMock, connectionKeeperMock,
	)

	// Create a packet with a small amount that will round to zero when scaled down
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
		SourcePort:         fixture.counterpartyPort,
		SourceChannel:      fixture.counterpartyChannel,
		DestinationPort:    fixture.nativePort,
		DestinationChannel: fixture.nativeChannel,
		Data:               dataBz,
	}

	// Set up channel and connection mocks
	channelKeeperMock.
		EXPECT().
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
		Times(3)

	connectionKeeperMock.
		EXPECT().
		GetConnection(ctx, fixture.nativeConnectionId).
		Return(
			connectiontypes.ConnectionEnd{
				ClientId: fixture.nativeClientId,
				Counterparty: connectiontypes.Counterparty{
					ClientId: fixture.counterpartyClientId,
				},
			}, true).
		Times(1)

	// Set up app mock to succeed
	appMock.
		EXPECT().
		OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{}).
		Return(channeltypes.NewResultAcknowledgement([]byte("success")))

	// Set up initial balances
	receiverAddr := sdk.MustAccAddressFromBech32(fixture.receiver)
	// Mint IBC tokens to receiver
	amountInt, _ := sdkmath.NewIntFromString(smallAmount)
	recvDenom := k.GetIBCRecvDenom(ctx, fixture.moduleDenom)
	ibcCoins := sdk.NewCoins(sdk.NewCoin(recvDenom, amountInt))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, ibcCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, receiverAddr, ibcCoins))

	// Check initial TotalTransferredIn
	totalTransferredInBefore := k.GetTotalTransferredIn(ctx)
	require.Equal(t, sdkmath.ZeroInt(), totalTransferredInBefore)

	// Execute OnRecvPacket - should fail at token conversion
	respOnRecvPacket := ibcModule.OnRecvPacket(ctx, "channel-version", packet, sdk.AccAddress{})

	// Validate that no error occurred in the IBC stack (ack is still successful)
	require.NotNil(t, respOnRecvPacket)
	require.True(t, respOnRecvPacket.Success(), "OnRecvPacket should return success even when token conversion fails")

	// EVENT CHECK
	// ----------------------------------------------------------------
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// find the token wrapper error event for token conversion failure
	var foundConversionErrorEvent bool
	for _, event := range events {
		if event.Type == types.EventTypeTokenWrapperError {
			// check event attributes - the EVENT contains the ORIGINAL error from ScaleDownTokenPrecision
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyError {
					if strings.Contains(string(attr.Value), "converted amount is zero or negative") {
						foundConversionErrorEvent = true
						t.Logf("Found token conversion error event: %s", string(attr.Value))
						break
					}
				}
			}
			if foundConversionErrorEvent {
				break
			}
		}
	}
	require.True(t, foundConversionErrorEvent, "Expected EventTypeTokenWrapperError event to be emitted for token conversion failure")

	// Check final TotalTransferredIn (no change - conversion failed)
	totalTransferredInAfter := k.GetTotalTransferredIn(ctx)
	require.Equal(t, totalTransferredInBefore, totalTransferredInAfter, "TotalTransferredIn should not change when token conversion fails")

	// Check that IBC token balance remained unchanged (no wrapping occurred)
	receiverIBCBalanceAfter := bankKeeper.GetBalance(ctx, receiverAddr, recvDenom)
	require.Equal(t, amountInt, receiverIBCBalanceAfter.Amount, "Receiver should still have IBC tokens since conversion failed")

	// Check that receiver has no native tokens (conversion failed before unlocking)
	receiverNativeBalanceAfter := bankKeeper.GetBalance(ctx, receiverAddr, constants.BondDenom)
	require.True(t, receiverNativeBalanceAfter.IsZero(), "Receiver should not have received native tokens since conversion failed")

	t.Log("SUCCESS: Verified that OnRecvPacket emits error event when token conversion fails")
}
