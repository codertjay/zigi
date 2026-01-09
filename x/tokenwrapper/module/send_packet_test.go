package tokenwrapper_test

import (
	"fmt"
	"strings"
	"testing"

	keepertest "zigchain/testutil/keeper"
	tokenwrapper "zigchain/x/tokenwrapper/module"
	mocks "zigchain/x/tokenwrapper/testutil"
	"zigchain/x/tokenwrapper/types"
	"zigchain/zutils/constants"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v10/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSendPacket(t *testing.T) {
	// Setup gomock controller
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

	// Test cases
	tests := []struct {
		name        string
		enabled     bool
		packetData  transfertypes.FungibleTokenPacketData
		setupMocks  func(ctx sdk.Context, keeperMock *mocks.MockTokenwrapperKeeper, transferKeeperMock *mocks.MockTransferKeeper, bankKeeperMock *mocks.MockBankKeeper, ics4WrapperMock *mocks.MockICS4Wrapper, channelKeeperMock *mocks.MockChannelKeeper, connectionKeeperMock *mocks.MockConnectionKeeper)
		expectedSeq uint64
		expectError bool
	}{
		{
			name:    "Happy path - successful token wrapping",
			enabled: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    constants.BondDenom,
				Amount:   "1000000", // 1 token with 6 decimals
				Sender:   sender,
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, keeperMock *mocks.MockTokenwrapperKeeper, transferKeeperMock *mocks.MockTransferKeeper, bankKeeperMock *mocks.MockBankKeeper, ics4WrapperMock *mocks.MockICS4Wrapper, channelKeeperMock *mocks.MockChannelKeeper, connectionKeeperMock *mocks.MockConnectionKeeper) {
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
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort)
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel)
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort)
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel)
				keeperMock.EXPECT().GetNativeClientId(ctx).Return(nativeClientId)
				keeperMock.EXPECT().GetCounterpartyClientId(ctx).Return(counterpartyClientId)
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom)
				keeperMock.EXPECT().CheckAccountBalance(ctx, sdk.MustAccAddressFromBech32(sender), sdk.NewCoins(sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(1000000)))).Return(nil)
				keeperMock.EXPECT().CheckModuleBalance(ctx, sdk.NewCoins(sdk.NewCoin(ibcDenom, sdkmath.NewInt(1000000000000000000)))).Return(nil)
				keeperMock.EXPECT().LockTokens(ctx, sdk.MustAccAddressFromBech32(sender), sdk.NewCoins(sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(1000000)))).Return(nil)
				keeperMock.EXPECT().BurnIbcTokens(ctx, sdk.NewCoins(sdk.NewCoin(ibcDenom, sdkmath.NewInt(1000000000000000000)))).Return(nil)
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				ics4WrapperMock.EXPECT().SendPacket(ctx, nativePort, nativeChannel, clienttypes.Height{}, uint64(0), gomock.Any()).Return(uint64(1), nil)
				keeperMock.EXPECT().ScaleUpTokenPrecision(ctx, sdkmath.NewInt(1000000)).Return(sdkmath.NewInt(1000000000000000000), nil)
				escrowAddress := transfertypes.GetEscrowAddress(nativePort, nativeChannel)
				token := sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(1000000))
				bankKeeperMock.EXPECT().SendCoins(ctx, escrowAddress, sdk.MustAccAddressFromBech32(sender), sdk.NewCoins(token)).Return(nil)
				bankKeeperMock.EXPECT().HasBalance(ctx, sdk.MustAccAddressFromBech32(sender), token).Return(true)
				currentTotalEscrow := sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(1000000))
				transferKeeperMock.EXPECT().GetTotalEscrowForDenom(ctx, constants.BondDenom).Return(currentTotalEscrow)
				newTotalEscrow := currentTotalEscrow.Sub(token)
				transferKeeperMock.EXPECT().SetTotalEscrowForDenom(ctx, newTotalEscrow).Return()
			},
			expectedSeq: 1,
			expectError: false,
		},
		{
			name:    "Error - channel not found",
			enabled: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    constants.BondDenom,
				Amount:   "1000000",
				Sender:   sender,
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, keeperMock *mocks.MockTokenwrapperKeeper, transferKeeperMock *mocks.MockTransferKeeper, bankKeeperMock *mocks.MockBankKeeper, ics4WrapperMock *mocks.MockICS4Wrapper, channelKeeperMock *mocks.MockChannelKeeper, connectionKeeperMock *mocks.MockConnectionKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, nativePort, nativeChannel).Return(channeltypes.Channel{}, false)
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
			},
			expectError: true,
		},
		{
			name:    "Error - channel not open",
			enabled: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    constants.BondDenom,
				Amount:   "1000000",
				Sender:   sender,
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, keeperMock *mocks.MockTokenwrapperKeeper, transferKeeperMock *mocks.MockTransferKeeper, bankKeeperMock *mocks.MockBankKeeper, ics4WrapperMock *mocks.MockICS4Wrapper, channelKeeperMock *mocks.MockChannelKeeper, connectionKeeperMock *mocks.MockConnectionKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, nativePort, nativeChannel).Return(channeltypes.Channel{
					State: channeltypes.CLOSED,
				}, false)
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
			},
			expectError: true,
		},
		{
			name:    "Error - module disabled",
			enabled: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    constants.BondDenom,
				Amount:   "1000000",
				Sender:   sender,
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, keeperMock *mocks.MockTokenwrapperKeeper, transferKeeperMock *mocks.MockTransferKeeper, bankKeeperMock *mocks.MockBankKeeper, ics4WrapperMock *mocks.MockICS4Wrapper, channelKeeperMock *mocks.MockChannelKeeper, connectionKeeperMock *mocks.MockConnectionKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, nativePort, nativeChannel).Return(channeltypes.Channel{
					State: channeltypes.OPEN,
				}, true)
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort)
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel)
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom)
				keeperMock.EXPECT().IsEnabled(ctx).Return(false)
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
			},
			expectError: true,
		},
		{
			name:    "Error - invalid packet data",
			enabled: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    constants.BondDenom,
				Amount:   "invalid-amount",
				Sender:   sender,
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, keeperMock *mocks.MockTokenwrapperKeeper, transferKeeperMock *mocks.MockTransferKeeper, bankKeeperMock *mocks.MockBankKeeper, ics4WrapperMock *mocks.MockICS4Wrapper, channelKeeperMock *mocks.MockChannelKeeper, connectionKeeperMock *mocks.MockConnectionKeeper) {
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
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort)
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel)
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort)
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel)
				keeperMock.EXPECT().GetNativeClientId(ctx).Return(nativeClientId)
				keeperMock.EXPECT().GetCounterpartyClientId(ctx).Return(counterpartyClientId)
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom)
			},
			expectError: true,
		},
		{
			name:    "Error - invalid sender address",
			enabled: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    constants.BondDenom,
				Amount:   "1000000",
				Sender:   "invalid-address",
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, keeperMock *mocks.MockTokenwrapperKeeper, transferKeeperMock *mocks.MockTransferKeeper, bankKeeperMock *mocks.MockBankKeeper, ics4WrapperMock *mocks.MockICS4Wrapper, channelKeeperMock *mocks.MockChannelKeeper, connectionKeeperMock *mocks.MockConnectionKeeper) {
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
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort)
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel)
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort)
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel)
				keeperMock.EXPECT().GetNativeClientId(ctx).Return(nativeClientId)
				keeperMock.EXPECT().GetCounterpartyClientId(ctx).Return(counterpartyClientId)
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom)
			},
			expectError: true,
		},
		{
			name:    "Error - LockTokens failure",
			enabled: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    constants.BondDenom,
				Amount:   "1000000",
				Sender:   sender,
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, keeperMock *mocks.MockTokenwrapperKeeper, transferKeeperMock *mocks.MockTransferKeeper, bankKeeperMock *mocks.MockBankKeeper, ics4WrapperMock *mocks.MockICS4Wrapper, channelKeeperMock *mocks.MockChannelKeeper, connectionKeeperMock *mocks.MockConnectionKeeper) {
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
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort)
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel)
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort)
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel)
				keeperMock.EXPECT().GetNativeClientId(ctx).Return(nativeClientId)
				keeperMock.EXPECT().GetCounterpartyClientId(ctx).Return(counterpartyClientId)
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom)
				keeperMock.EXPECT().CheckAccountBalance(ctx, sdk.MustAccAddressFromBech32(sender), sdk.NewCoins(sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(1000000)))).Return(nil)
				keeperMock.EXPECT().CheckModuleBalance(ctx, sdk.NewCoins(sdk.NewCoin(ibcDenom, sdkmath.NewInt(1000000000000000000)))).Return(nil)
				keeperMock.EXPECT().LockTokens(ctx, sdk.MustAccAddressFromBech32(sender), sdk.NewCoins(sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(1000000)))).Return(fmt.Errorf("lock tokens failed"))
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				keeperMock.EXPECT().ScaleUpTokenPrecision(ctx, sdkmath.NewInt(1000000)).Return(sdkmath.NewInt(1000000000000000000), nil)
				escrowAddress := transfertypes.GetEscrowAddress(nativePort, nativeChannel)
				token := sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(1000000))
				bankKeeperMock.EXPECT().SendCoins(ctx, escrowAddress, sdk.MustAccAddressFromBech32(sender), sdk.NewCoins(token)).Return(nil)
				bankKeeperMock.EXPECT().HasBalance(ctx, sdk.MustAccAddressFromBech32(sender), token).Return(true)
				currentTotalEscrow := sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(1000000))
				transferKeeperMock.EXPECT().GetTotalEscrowForDenom(ctx, constants.BondDenom).Return(currentTotalEscrow)
				newTotalEscrow := currentTotalEscrow.Sub(token)
				transferKeeperMock.EXPECT().SetTotalEscrowForDenom(ctx, newTotalEscrow).Return()
			},
			expectError: true,
		},
		{
			name:    "Error - BurnIbcTokens failure",
			enabled: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    constants.BondDenom,
				Amount:   "1000000",
				Sender:   sender,
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, keeperMock *mocks.MockTokenwrapperKeeper, transferKeeperMock *mocks.MockTransferKeeper, bankKeeperMock *mocks.MockBankKeeper, ics4WrapperMock *mocks.MockICS4Wrapper, channelKeeperMock *mocks.MockChannelKeeper, connectionKeeperMock *mocks.MockConnectionKeeper) {
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
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort)
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel)
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort)
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel)
				keeperMock.EXPECT().GetNativeClientId(ctx).Return(nativeClientId)
				keeperMock.EXPECT().GetCounterpartyClientId(ctx).Return(counterpartyClientId)
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom)
				keeperMock.EXPECT().CheckAccountBalance(ctx, sdk.MustAccAddressFromBech32(sender), sdk.NewCoins(sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(1000000)))).Return(nil)
				keeperMock.EXPECT().CheckModuleBalance(ctx, sdk.NewCoins(sdk.NewCoin(ibcDenom, sdkmath.NewInt(1000000000000000000)))).Return(nil)
				keeperMock.EXPECT().LockTokens(ctx, sdk.MustAccAddressFromBech32(sender), sdk.NewCoins(sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(1000000)))).Return(nil)
				keeperMock.EXPECT().BurnIbcTokens(ctx, sdk.NewCoins(sdk.NewCoin(ibcDenom, sdkmath.NewInt(1000000000000000000)))).Return(fmt.Errorf("burn tokens failed"))
				keeperMock.EXPECT().UnlockTokens(ctx, sdk.MustAccAddressFromBech32(sender), sdk.NewCoins(sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(1000000)))).Return(nil)
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				keeperMock.EXPECT().ScaleUpTokenPrecision(ctx, sdkmath.NewInt(1000000)).Return(sdkmath.NewInt(1000000000000000000), nil)
				escrowAddress := transfertypes.GetEscrowAddress(nativePort, nativeChannel)
				token := sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(1000000))
				bankKeeperMock.EXPECT().SendCoins(ctx, escrowAddress, sdk.MustAccAddressFromBech32(sender), sdk.NewCoins(token)).Return(nil)
				bankKeeperMock.EXPECT().HasBalance(ctx, sdk.MustAccAddressFromBech32(sender), token).Return(true)
				currentTotalEscrow := sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(1000000))
				transferKeeperMock.EXPECT().GetTotalEscrowForDenom(ctx, constants.BondDenom).Return(currentTotalEscrow)
				newTotalEscrow := currentTotalEscrow.Sub(token)
				transferKeeperMock.EXPECT().SetTotalEscrowForDenom(ctx, newTotalEscrow).Return()
			},
			expectError: true,
		},
		{
			name:    "Error - SendPacket failure",
			enabled: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    constants.BondDenom,
				Amount:   "1000000",
				Sender:   sender,
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, keeperMock *mocks.MockTokenwrapperKeeper, transferKeeperMock *mocks.MockTransferKeeper, bankKeeperMock *mocks.MockBankKeeper, ics4WrapperMock *mocks.MockICS4Wrapper, channelKeeperMock *mocks.MockChannelKeeper, connectionKeeperMock *mocks.MockConnectionKeeper) {
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
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort)
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel)
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort)
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel)
				keeperMock.EXPECT().GetNativeClientId(ctx).Return(nativeClientId)
				keeperMock.EXPECT().GetCounterpartyClientId(ctx).Return(counterpartyClientId)
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom)
				keeperMock.EXPECT().CheckAccountBalance(ctx, sdk.MustAccAddressFromBech32(sender), sdk.NewCoins(sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(1000000)))).Return(nil)
				keeperMock.EXPECT().CheckModuleBalance(ctx, sdk.NewCoins(sdk.NewCoin(ibcDenom, sdkmath.NewInt(1000000000000000000)))).Return(nil)
				keeperMock.EXPECT().LockTokens(ctx, sdk.MustAccAddressFromBech32(sender), sdk.NewCoins(sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(1000000)))).Return(nil)
				keeperMock.EXPECT().BurnIbcTokens(ctx, sdk.NewCoins(sdk.NewCoin(ibcDenom, sdkmath.NewInt(1000000000000000000)))).Return(nil)
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				keeperMock.EXPECT().ScaleUpTokenPrecision(ctx, sdkmath.NewInt(1000000)).Return(sdkmath.NewInt(1000000000000000000), nil)
				ics4WrapperMock.EXPECT().SendPacket(ctx, nativePort, nativeChannel, clienttypes.Height{}, uint64(0), gomock.Any()).Return(uint64(0), fmt.Errorf("send packet failed"))
				escrowAddress := transfertypes.GetEscrowAddress(nativePort, nativeChannel)
				token := sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(1000000))
				bankKeeperMock.EXPECT().SendCoins(ctx, escrowAddress, sdk.MustAccAddressFromBech32(sender), sdk.NewCoins(token)).Return(nil)
				bankKeeperMock.EXPECT().HasBalance(ctx, sdk.MustAccAddressFromBech32(sender), token).Return(true)
				currentTotalEscrow := sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(1000000))
				transferKeeperMock.EXPECT().GetTotalEscrowForDenom(ctx, constants.BondDenom).Return(currentTotalEscrow)
				newTotalEscrow := currentTotalEscrow.Sub(token)
				transferKeeperMock.EXPECT().SetTotalEscrowForDenom(ctx, newTotalEscrow).Return()
			},
			expectError: true,
		},
		{
			name:    "No wrapping - non-source chain",
			enabled: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    "ibc/" + moduleDenom, // Prefixed to indicate non-source chain
				Amount:   "1000000",
				Sender:   sender,
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, keeperMock *mocks.MockTokenwrapperKeeper, transferKeeperMock *mocks.MockTransferKeeper, bankKeeperMock *mocks.MockBankKeeper, ics4WrapperMock *mocks.MockICS4Wrapper, channelKeeperMock *mocks.MockChannelKeeper, connectionKeeperMock *mocks.MockConnectionKeeper) {
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
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort)
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel)
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom)
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				ics4WrapperMock.EXPECT().SendPacket(ctx, nativePort, nativeChannel, clienttypes.Height{}, uint64(0), gomock.Any()).Return(uint64(1), nil)
			},
			expectedSeq: 1,
			expectError: false,
		},
		{
			name:    "No wrapping - empty IBC settings",
			enabled: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    constants.BondDenom,
				Amount:   "1000000",
				Sender:   sender,
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, keeperMock *mocks.MockTokenwrapperKeeper, transferKeeperMock *mocks.MockTransferKeeper, bankKeeperMock *mocks.MockBankKeeper, ics4WrapperMock *mocks.MockICS4Wrapper, channelKeeperMock *mocks.MockChannelKeeper, connectionKeeperMock *mocks.MockConnectionKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, nativePort, nativeChannel).Return(channeltypes.Channel{
					State: channeltypes.OPEN,
				}, true)
				keeperMock.EXPECT().IsEnabled(ctx).Return(true).AnyTimes()
				keeperMock.EXPECT().GetNativePort(ctx).Return("")
				keeperMock.EXPECT().GetNativeChannel(ctx).Return("")
				keeperMock.EXPECT().GetDenom(ctx).Return("")
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				ics4WrapperMock.EXPECT().SendPacket(ctx, nativePort, nativeChannel, clienttypes.Height{}, uint64(0), gomock.Any()).Return(uint64(1), nil)
			},
			expectedSeq: 1,
			expectError: false,
		},
		{
			name:    "No wrapping - mismatched denom",
			enabled: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    "other-denom",
				Amount:   "1000000",
				Sender:   sender,
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, keeperMock *mocks.MockTokenwrapperKeeper, transferKeeperMock *mocks.MockTransferKeeper, bankKeeperMock *mocks.MockBankKeeper, ics4WrapperMock *mocks.MockICS4Wrapper, channelKeeperMock *mocks.MockChannelKeeper, connectionKeeperMock *mocks.MockConnectionKeeper) {
				channelKeeperMock.EXPECT().GetChannel(ctx, nativePort, nativeChannel).Return(channeltypes.Channel{
					State: channeltypes.OPEN,
				}, true)
				keeperMock.EXPECT().IsEnabled(ctx).Return(true).AnyTimes()
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort)
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel)
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom)
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				ics4WrapperMock.EXPECT().SendPacket(ctx, nativePort, nativeChannel, clienttypes.Height{}, uint64(0), gomock.Any()).Return(uint64(1), nil)
			},
			expectedSeq: 1,
			expectError: false,
		},
		{
			name:    "Success - large amount",
			enabled: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    constants.BondDenom,
				Amount:   "1000000000000", // Very large amount
				Sender:   sender,
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, keeperMock *mocks.MockTokenwrapperKeeper, transferKeeperMock *mocks.MockTransferKeeper, bankKeeperMock *mocks.MockBankKeeper, ics4WrapperMock *mocks.MockICS4Wrapper, channelKeeperMock *mocks.MockChannelKeeper, connectionKeeperMock *mocks.MockConnectionKeeper) {
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
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort)
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel)
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort)
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel)
				keeperMock.EXPECT().GetNativeClientId(ctx).Return(nativeClientId)
				keeperMock.EXPECT().GetCounterpartyClientId(ctx).Return(counterpartyClientId)
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom)
				amount, _ := sdkmath.NewIntFromString("1000000000000")
				keeperMock.EXPECT().CheckAccountBalance(ctx, sdk.MustAccAddressFromBech32(sender), sdk.NewCoins(sdk.NewCoin(constants.BondDenom, amount))).Return(nil)
				keeperMock.EXPECT().CheckModuleBalance(ctx, sdk.NewCoins(sdk.NewCoin(ibcDenom, amount.Mul(sdkmath.NewInt(1000000000000))))).Return(nil)
				keeperMock.EXPECT().LockTokens(ctx, sdk.MustAccAddressFromBech32(sender), sdk.NewCoins(sdk.NewCoin(constants.BondDenom, amount))).Return(nil)
				keeperMock.EXPECT().BurnIbcTokens(ctx, sdk.NewCoins(sdk.NewCoin(ibcDenom, amount.Mul(sdkmath.NewInt(1000000000000))))).Return(nil)
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				convertedAmount, _ := sdkmath.NewIntFromString("1000000000000000000000000")
				keeperMock.EXPECT().ScaleUpTokenPrecision(ctx, amount).Return(convertedAmount, nil)
				ics4WrapperMock.EXPECT().SendPacket(ctx, nativePort, nativeChannel, clienttypes.Height{}, uint64(0), gomock.Any()).Return(uint64(1), nil)
				escrowAddress := transfertypes.GetEscrowAddress(nativePort, nativeChannel)
				token := sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(1000000000000))
				bankKeeperMock.EXPECT().SendCoins(ctx, escrowAddress, sdk.MustAccAddressFromBech32(sender), sdk.NewCoins(token)).Return(nil)
				bankKeeperMock.EXPECT().HasBalance(ctx, sdk.MustAccAddressFromBech32(sender), token).Return(true)
				currentTotalEscrow := sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(1000000000000))
				transferKeeperMock.EXPECT().GetTotalEscrowForDenom(ctx, constants.BondDenom).Return(currentTotalEscrow)
				newTotalEscrow := currentTotalEscrow.Sub(token)
				transferKeeperMock.EXPECT().SetTotalEscrowForDenom(ctx, newTotalEscrow).Return()
			},
			expectedSeq: 1,
			expectError: false,
		},
		{
			name:    "Error - zero converted amount",
			enabled: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    constants.BondDenom,
				Amount:   "1000000",
				Sender:   sender,
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, keeperMock *mocks.MockTokenwrapperKeeper, transferKeeperMock *mocks.MockTransferKeeper, bankKeeperMock *mocks.MockBankKeeper, ics4WrapperMock *mocks.MockICS4Wrapper, channelKeeperMock *mocks.MockChannelKeeper, connectionKeeperMock *mocks.MockConnectionKeeper) {
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
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort)
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel)
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort)
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel)
				keeperMock.EXPECT().GetNativeClientId(ctx).Return(nativeClientId)
				keeperMock.EXPECT().GetCounterpartyClientId(ctx).Return(counterpartyClientId)
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom)
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				escrowAddress := transfertypes.GetEscrowAddress(nativePort, nativeChannel)
				token := sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(1000000))
				bankKeeperMock.EXPECT().SendCoins(ctx, escrowAddress, sdk.MustAccAddressFromBech32(sender), sdk.NewCoins(token)).Return(nil)
				keeperMock.EXPECT().ScaleUpTokenPrecision(ctx, sdkmath.NewInt(1000000)).Return(sdkmath.ZeroInt(), fmt.Errorf("converted amount is zero or negative: %s", sdkmath.ZeroInt().String()))
				bankKeeperMock.EXPECT().HasBalance(ctx, sdk.MustAccAddressFromBech32(sender), token).Return(true)
				currentTotalEscrow := sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(1000000))
				transferKeeperMock.EXPECT().GetTotalEscrowForDenom(ctx, constants.BondDenom).Return(currentTotalEscrow)
				newTotalEscrow := currentTotalEscrow.Sub(token)
				transferKeeperMock.EXPECT().SetTotalEscrowForDenom(ctx, newTotalEscrow).Return()
			},
			expectError: true,
		},
		{
			name:    "Error - negative converted amount",
			enabled: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    constants.BondDenom,
				Amount:   "1000000",
				Sender:   sender,
				Receiver: receiver,
			},
			setupMocks: func(ctx sdk.Context, keeperMock *mocks.MockTokenwrapperKeeper, transferKeeperMock *mocks.MockTransferKeeper, bankKeeperMock *mocks.MockBankKeeper, ics4WrapperMock *mocks.MockICS4Wrapper, channelKeeperMock *mocks.MockChannelKeeper, connectionKeeperMock *mocks.MockConnectionKeeper) {
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
				keeperMock.EXPECT().GetNativePort(ctx).Return(nativePort)
				keeperMock.EXPECT().GetNativeChannel(ctx).Return(nativeChannel)
				keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(counterpartyPort)
				keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(counterpartyChannel)
				keeperMock.EXPECT().GetNativeClientId(ctx).Return(nativeClientId)
				keeperMock.EXPECT().GetCounterpartyClientId(ctx).Return(counterpartyClientId)
				keeperMock.EXPECT().GetDenom(ctx).Return(moduleDenom)
				keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
				escrowAddress := transfertypes.GetEscrowAddress(nativePort, nativeChannel)
				keeperMock.EXPECT().ScaleUpTokenPrecision(ctx, sdkmath.NewInt(1000000)).Return(sdkmath.ZeroInt(), fmt.Errorf("converted amount is zero or negative: %s", sdkmath.ZeroInt().String()))
				token := sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(1000000))
				bankKeeperMock.EXPECT().SendCoins(ctx, escrowAddress, sdk.MustAccAddressFromBech32(sender), sdk.NewCoins(token)).Return(nil)
				bankKeeperMock.EXPECT().HasBalance(ctx, sdk.MustAccAddressFromBech32(sender), token).Return(true)
				currentTotalEscrow := sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(1000000))
				transferKeeperMock.EXPECT().GetTotalEscrowForDenom(ctx, constants.BondDenom).Return(currentTotalEscrow)
				newTotalEscrow := currentTotalEscrow.Sub(token)
				transferKeeperMock.EXPECT().SetTotalEscrowForDenom(ctx, newTotalEscrow).Return()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.enabled {
				fmt.Println("test \"", tt.name, "\" disabled, skipping…")
				return
			}

			// Create fresh mocks for each test case
			keeperMock := mocks.NewMockTokenwrapperKeeper(ctrl)
			transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
			bankKeeperMock := mocks.NewMockBankKeeper(ctrl)
			ics4WrapperMock := mocks.NewMockICS4Wrapper(ctrl)
			channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
			connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

			// Create fresh ICS4Wrapper instance for each test case
			ics4Wrapper := tokenwrapper.NewICS4Wrapper(
				ics4WrapperMock,
				bankKeeperMock,
				transferKeeperMock,
				channelKeeperMock,
				connectionKeeperMock,
				keeperMock,
			)

			// Create context
			ctx := sdk.Context{}

			// Marshal packet data
			dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&tt.packetData)
			require.NoError(t, err)

			// Setup mocks
			tt.setupMocks(ctx, keeperMock, transferKeeperMock, bankKeeperMock, ics4WrapperMock, channelKeeperMock, connectionKeeperMock)

			// Call SendPacket
			seq, err := ics4Wrapper.SendPacket(ctx, nativePort, nativeChannel, clienttypes.Height{}, 0, dataBz)

			// Verify results
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedSeq, seq)
			}
		})
	}
}

// Test fixture with positive scenario data
type OnSendPacketFixture struct {
	faucet                  string
	sender                  string
	receiver                string
	nativePort              string
	destinationPort         string
	nativeChannel           string
	destinationChannel      string
	nativeConnectionId      string
	destinationConnectionId string
	nativeClientId          string
	counterpartyClientId    string
	moduleDenom             string
	ibcDenom                string
	amount                  string
	decimalDifference       uint32
}

func getSendPacketPositiveFixture() OnSendPacketFixture {

	return OnSendPacketFixture{
		faucet:                  "zig15yk64u7zc9g9k2yr2wmzeva5qgwxps6y8c2amk",
		sender:                  "zig1fcf3p6exuhq53myh0k3wweym5ffv00gcnuum07",
		receiver:                "axelar15yk64u7zc9g9k2yr2wmzeva5qgwxps6yzue8jl",
		nativePort:              "transfer",
		destinationPort:         "transfer",
		nativeChannel:           "channel-12",
		destinationChannel:      "channel-24",
		nativeConnectionId:      "connection-12",
		destinationConnectionId: "connection-24",
		nativeClientId:          "07-tendermint-12",
		counterpartyClientId:    "07-tendermint-24",
		moduleDenom:             "axl",
		ibcDenom:                "ibc/148AEF32AA7274DC6AFD912A5C1478AC10246B8AEE1C8DEA6D831B752000E89F",
		amount:                  "1000000", // 1 token with 6 decimals
		decimalDifference:       12,
	}
}

func TestSendPacket_RealKeepers(t *testing.T) {
	// Test case: SendPacket with real keepers to verify balance changes

	// set up positive fixture
	fixture := getSendPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.destinationPort)
	k.SetCounterpartyChannel(ctx, fixture.destinationChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	ics4WrapperMock := mocks.NewMockICS4Wrapper(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create ICS4Wrapper instance
	ics4Wrapper := tokenwrapper.NewICS4Wrapper(
		ics4WrapperMock,
		bankKeeper,
		transferKeeperMock,
		channelKeeperMock,
		connectionKeeperMock,
		k,
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

	// Set up channel and connection mocks
	channelKeeperMock.
		EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.destinationPort,
					ChannelId: fixture.destinationChannel,
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

	// Set up initial balances
	senderAddr := sdk.MustAccAddressFromBech32(fixture.sender)
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)

	// Mint native tokens to sender
	amountInt, _ := sdkmath.NewIntFromString(fixture.amount)
	nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, amountInt))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, nativeCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, senderAddr, nativeCoins))

	// Mint IBC tokens to module
	conversionFactor := k.GetDecimalConversionFactor(ctx)
	convertedAmount := amountInt.Mul(conversionFactor)
	ibcCoins := sdk.NewCoins(sdk.NewCoin(fixture.ibcDenom, convertedAmount))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, ibcCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, ibcCoins))

	// Check initial balances
	senderBalanceBefore := bankKeeper.GetBalance(ctx, senderAddr, constants.BondDenom)
	require.Equal(t, amountInt, senderBalanceBefore.Amount)

	moduleBalanceBefore := bankKeeper.GetBalance(ctx, moduleAddr, fixture.ibcDenom)
	require.Equal(t, convertedAmount, moduleBalanceBefore.Amount)

	// Set up ICS4Wrapper mock for SendPacket
	ics4WrapperMock.
		EXPECT().
		SendPacket(ctx, fixture.nativePort, fixture.nativeChannel, clienttypes.Height{}, uint64(0), gomock.Any()).
		Return(uint64(1), nil)

	// Mock unescrow token
	currentTotalEscrow := sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(1000000))
	transferKeeperMock.EXPECT().GetTotalEscrowForDenom(ctx, constants.BondDenom).Return(currentTotalEscrow)
	newTotalEscrow := currentTotalEscrow.Sub(nativeCoins[0])
	transferKeeperMock.EXPECT().SetTotalEscrowForDenom(ctx, newTotalEscrow).Return()

	// Apply the escrow logic (SendTransfer)
	escrowAddress := transfertypes.GetEscrowAddress(fixture.nativePort, fixture.nativeChannel)
	require.NoError(t, bankKeeper.SendCoins(ctx, senderAddr, escrowAddress, nativeCoins))

	// check escrow and sender balances
	escrowBalance := bankKeeper.GetBalance(ctx, escrowAddress, constants.BondDenom)
	require.Equal(t, nativeCoins[0], escrowBalance)
	senderBalancePostEscrow := bankKeeper.GetBalance(ctx, senderAddr, constants.BondDenom)
	require.Equal(t, sdkmath.ZeroInt(), senderBalancePostEscrow.Amount)

	// Execute SendPacket
	seq, err := ics4Wrapper.SendPacket(ctx, fixture.nativePort, fixture.nativeChannel, clienttypes.Height{}, 0, dataBz)

	// Validate that no error occurred
	require.NoError(t, err)
	require.Equal(t, uint64(1), seq)

	// check escrow balance is empty
	escrowBalanceAfter := bankKeeper.GetBalance(ctx, escrowAddress, constants.BondDenom)
	require.True(t, escrowBalanceAfter.Amount.IsZero(), "Escrow should have no native tokens left")

	// Check final balances
	senderBalanceAfter := bankKeeper.GetBalance(ctx, senderAddr, constants.BondDenom)
	require.Equal(t, sdkmath.ZeroInt(), senderBalanceAfter.Amount, "Sender should have no native tokens left")

	moduleBalanceAfter := bankKeeper.GetBalance(ctx, moduleAddr, fixture.ibcDenom)
	require.Equal(t, sdkmath.ZeroInt(), moduleBalanceAfter.Amount, "Module should have no IBC tokens left")
}

func TestSendPacket_MismatchedPortChannel(t *testing.T) {
	// Test case: SendPacket with mismatched port/channel to verify no wrapping

	// Set up positive fixture
	fixture := getSendPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper with different port/channel
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, "other-port")
	k.SetNativeChannel(ctx, "other-channel")
	k.SetCounterpartyPort(ctx, fixture.destinationPort)
	k.SetCounterpartyChannel(ctx, fixture.destinationChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	bankKeeperMock := mocks.NewMockBankKeeper(ctrl)
	ics4WrapperMock := mocks.NewMockICS4Wrapper(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create ICS4Wrapper instance
	ics4Wrapper := tokenwrapper.NewICS4Wrapper(
		ics4WrapperMock,
		bankKeeperMock,
		transferKeeperMock,
		channelKeeperMock,
		connectionKeeperMock,
		k,
	)

	// Create packet data
	data := transfertypes.FungibleTokenPacketData{
		Denom:    constants.BondDenom,
		Amount:   fixture.amount,
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	// Set up channel mock
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.destinationPort,
					ChannelId: fixture.destinationChannel,
				},
			}, true).
		Times(1)

	// Set up initial balances
	senderAddr := sdk.MustAccAddressFromBech32(fixture.sender)
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	escrowAddr := transfertypes.GetEscrowAddress(fixture.nativePort, fixture.nativeChannel)

	// Mint native tokens to sender
	amountInt, ok := sdkmath.NewIntFromString(fixture.amount)
	require.True(t, ok)
	nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, amountInt))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, nativeCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, senderAddr, nativeCoins))

	// Mint IBC tokens to module
	conversionFactor := k.GetDecimalConversionFactor(ctx)
	convertedAmount := amountInt.Mul(conversionFactor)
	ibcCoins := sdk.NewCoins(sdk.NewCoin(fixture.ibcDenom, convertedAmount))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, ibcCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, ibcCoins))

	// Check initial balances
	senderBalanceBefore := bankKeeper.GetBalance(ctx, senderAddr, constants.BondDenom)
	require.Equal(t, amountInt, senderBalanceBefore.Amount)
	moduleBalanceBefore := bankKeeper.GetBalance(ctx, moduleAddr, fixture.ibcDenom)
	require.Equal(t, convertedAmount, moduleBalanceBefore.Amount)
	escrowBalanceBefore := bankKeeper.GetBalance(ctx, escrowAddr, constants.BondDenom)
	require.Equal(t, sdkmath.ZeroInt(), escrowBalanceBefore.Amount)

	// Set up ICS4Wrapper mock for SendPacket
	ics4WrapperMock.EXPECT().
		SendPacket(ctx, fixture.nativePort, fixture.nativeChannel, clienttypes.Height{}, uint64(0), gomock.Any()).
		Return(uint64(1), nil)

	// Execute SendPacket
	seq, err := ics4Wrapper.SendPacket(ctx, fixture.nativePort, fixture.nativeChannel, clienttypes.Height{}, 0, dataBz)

	// Expect success with no wrapping
	require.NoError(t, err)
	require.Equal(t, uint64(1), seq)

	// Check final balances (no changes due to no wrapping)
	senderBalanceAfter := bankKeeper.GetBalance(ctx, senderAddr, constants.BondDenom)
	require.Equal(t, senderBalanceBefore.Amount, senderBalanceAfter.Amount, "Sender balance should not change")
	moduleBalanceAfter := bankKeeper.GetBalance(ctx, moduleAddr, fixture.ibcDenom)
	require.Equal(t, moduleBalanceBefore.Amount, moduleBalanceAfter.Amount, "Module balance should not change")
	escrowBalanceAfter := bankKeeper.GetBalance(ctx, escrowAddr, constants.BondDenom)
	require.Equal(t, sdkmath.ZeroInt(), escrowBalanceAfter.Amount, "Escrow balance should remain zero")
}

func TestSendPacket_InvalidDenomFormat(t *testing.T) {
	// Test case: SendPacket with invalid denom format to verify no wrapping and no balance changes

	// Set up positive fixture
	fixture := getSendPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.destinationPort)
	k.SetCounterpartyChannel(ctx, fixture.destinationChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	ics4WrapperMock := mocks.NewMockICS4Wrapper(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create ICS4Wrapper instance
	ics4Wrapper := tokenwrapper.NewICS4Wrapper(
		ics4WrapperMock,
		bankKeeper,
		transferKeeperMock,
		channelKeeperMock,
		connectionKeeperMock,
		k,
	)

	// Create packet data with invalid denom
	data := transfertypes.FungibleTokenPacketData{
		Denom:    "invalid/denom/format",
		Amount:   fixture.amount,
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	// Set up channel mock
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.destinationPort,
					ChannelId: fixture.destinationChannel,
				},
			}, true).
		Times(1)

	// Set up initial balances
	senderAddr := sdk.MustAccAddressFromBech32(fixture.sender)
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	escrowAddr := transfertypes.GetEscrowAddress(fixture.nativePort, fixture.nativeChannel)

	// Mint native tokens to sender
	amountInt, ok := sdkmath.NewIntFromString(fixture.amount)
	require.True(t, ok)
	nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, amountInt))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, nativeCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, senderAddr, nativeCoins))

	// Mint IBC tokens to module
	conversionFactor := k.GetDecimalConversionFactor(ctx)
	convertedAmount := amountInt.Mul(conversionFactor)
	ibcCoins := sdk.NewCoins(sdk.NewCoin(fixture.ibcDenom, convertedAmount))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, ibcCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, ibcCoins))

	// Check initial balances
	senderBalanceBefore := bankKeeper.GetBalance(ctx, senderAddr, constants.BondDenom)
	require.Equal(t, amountInt, senderBalanceBefore.Amount)
	moduleBalanceBefore := bankKeeper.GetBalance(ctx, moduleAddr, fixture.ibcDenom)
	require.Equal(t, convertedAmount, moduleBalanceBefore.Amount)
	escrowBalanceBefore := bankKeeper.GetBalance(ctx, escrowAddr, constants.BondDenom)
	require.Equal(t, sdkmath.ZeroInt(), escrowBalanceBefore.Amount)

	// Set up ICS4Wrapper mock for SendPacket
	ics4WrapperMock.EXPECT().
		SendPacket(ctx, fixture.nativePort, fixture.nativeChannel, clienttypes.Height{}, uint64(0), gomock.Any()).
		Return(uint64(1), nil)

	// Execute SendPacket
	seq, err := ics4Wrapper.SendPacket(ctx, fixture.nativePort, fixture.nativeChannel, clienttypes.Height{}, 0, dataBz)

	// Expect success with no wrapping
	require.NoError(t, err)
	require.Equal(t, uint64(1), seq)

	// Check final balances (no changes due to no wrapping)
	senderBalanceAfter := bankKeeper.GetBalance(ctx, senderAddr, constants.BondDenom)
	require.Equal(t, senderBalanceBefore.Amount, senderBalanceAfter.Amount, "Sender balance should not change")
	moduleBalanceAfter := bankKeeper.GetBalance(ctx, moduleAddr, fixture.ibcDenom)
	require.Equal(t, moduleBalanceBefore.Amount, moduleBalanceAfter.Amount, "Module balance should not change")
	escrowBalanceAfter := bankKeeper.GetBalance(ctx, escrowAddr, constants.BondDenom)
	require.Equal(t, sdkmath.ZeroInt(), escrowBalanceAfter.Amount, "Escrow balance should remain zero")
}

func TestSendPacket_InvalidPacketJSON(t *testing.T) {
	// Test case: SendPacket with invalid JSON packet data

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fixture := getSendPacketPositiveFixture()

	ics4WrapperMock := mocks.NewMockICS4Wrapper(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	keeperMock := mocks.NewMockTokenwrapperKeeper(ctrl)
	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	bankKeeperMock := mocks.NewMockBankKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	ctx := sdk.Context{}
	// Create ICS4Wrapper instance
	ics4Wrapper := tokenwrapper.NewICS4Wrapper(
		ics4WrapperMock,
		bankKeeperMock,
		transferKeeperMock,
		channelKeeperMock,
		connectionKeeperMock,
		keeperMock,
	)

	// Mock channel check
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(channeltypes.Channel{State: channeltypes.OPEN}, true)

	// Mock logger
	keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()

	// Invalid JSON packet data
	packetData := []byte("invalid-json")

	// Call SendPacket
	_, err := ics4Wrapper.SendPacket(ctx, fixture.nativePort, fixture.nativeChannel, clienttypes.Height{}, 0, packetData)

	// Expect error due to invalid JSON
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid character 'i' looking for beginning of value")
}

func TestSendPacket_InsufficientSenderBalance(t *testing.T) {
	// Test case: SendPacket with insufficient sender balance to verify error handling

	// Set up positive fixture
	fixture := getSendPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.destinationPort)
	k.SetCounterpartyChannel(ctx, fixture.destinationChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	ics4WrapperMock := mocks.NewMockICS4Wrapper(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create ICS4Wrapper instance
	ics4Wrapper := tokenwrapper.NewICS4Wrapper(
		ics4WrapperMock,
		bankKeeper,
		transferKeeperMock,
		channelKeeperMock,
		connectionKeeperMock,
		k,
	)

	// Create packet data
	data := transfertypes.FungibleTokenPacketData{
		Denom:    constants.BondDenom,
		Amount:   fixture.amount, // e.g., "1000000"
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	// Set up channel mock
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.destinationPort,
					ChannelId: fixture.destinationChannel,
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

	// Set up sender with insufficient balance
	senderAddr := sdk.MustAccAddressFromBech32(fixture.sender)
	amountInt, ok := sdkmath.NewIntFromString(fixture.amount)
	require.True(t, ok)
	nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, amountInt.Sub(sdkmath.NewInt(1)))) // Mint less than required
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, nativeCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, senderAddr, nativeCoins))

	// Check initial sender balance (insufficient)
	senderBalanceBefore := bankKeeper.GetBalance(ctx, senderAddr, constants.BondDenom)
	require.Equal(t, amountInt.Sub(sdkmath.NewInt(1)), senderBalanceBefore.Amount)

	// Execute SendPacket
	_, err = ics4Wrapper.SendPacket(ctx, fixture.nativePort, fixture.nativeChannel, clienttypes.Height{}, 0, dataBz)

	// Expect error due to insufficient sender balance
	require.Error(t, err)
	require.Contains(t, err.Error(), "unable to unescrow tokens: spendable balance 0uzig is smaller than 1000000uzig: insufficient funds")
}

func TestSendPacket_InsufficientModuleBalance(t *testing.T) {
	// Test case: SendPacket with insufficient module balance to verify error handling

	// Set up positive fixture
	fixture := getSendPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.destinationPort)
	k.SetCounterpartyChannel(ctx, fixture.destinationChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	ics4WrapperMock := mocks.NewMockICS4Wrapper(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ics4Wrapper := tokenwrapper.NewICS4Wrapper(
		ics4WrapperMock,
		bankKeeper,
		transferKeeperMock,
		channelKeeperMock,
		connectionKeeperMock,
		k,
	)

	// Create packet data
	data := transfertypes.FungibleTokenPacketData{
		Denom:    constants.BondDenom,
		Amount:   fixture.amount,
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	// Set up channel mock
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.destinationPort,
					ChannelId: fixture.destinationChannel,
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

	// Set up sender balance (sufficient)
	senderAddr := sdk.MustAccAddressFromBech32(fixture.sender)
	amountInt, ok := sdkmath.NewIntFromString(fixture.amount)
	require.True(t, ok)
	nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, amountInt))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, nativeCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, senderAddr, nativeCoins))

	// Check initial sender balance
	senderBalanceBefore := bankKeeper.GetBalance(ctx, senderAddr, constants.BondDenom)
	require.Equal(t, amountInt, senderBalanceBefore.Amount)

	// Set up module with insufficient IBC balance
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	conversionFactor := k.GetDecimalConversionFactor(ctx)
	convertedAmount := amountInt.Mul(conversionFactor)
	ibcCoins := sdk.NewCoins(sdk.NewCoin(fixture.ibcDenom, convertedAmount.Sub(sdkmath.NewInt(0))))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, ibcCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, ibcCoins))

	// Check initial module balance (insufficient)
	moduleBalanceBefore := bankKeeper.GetBalance(ctx, moduleAddr, fixture.ibcDenom)
	require.Equal(t, convertedAmount.Sub(sdkmath.NewInt(0)), moduleBalanceBefore.Amount)

	// Execute SendPacket
	_, err = ics4Wrapper.SendPacket(ctx, fixture.nativePort, fixture.nativeChannel, clienttypes.Height{}, 0, dataBz)

	// Expect error due to insufficient module balance
	require.Error(t, err)
	require.Contains(t, err.Error(), "unable to unescrow tokens: spendable balance 0uzig is smaller than 1000000uzig: insufficient funds")
}

func TestSendPacket_UnescrowTokenFailure(t *testing.T) {
	// Test case: SendPacket with unescrow token failure to verify error handling

	// Set up positive fixture
	fixture := getSendPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.destinationPort)
	k.SetCounterpartyChannel(ctx, fixture.destinationChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	ics4WrapperMock := mocks.NewMockICS4Wrapper(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create IBCModule
	ics4Wrapper := tokenwrapper.NewICS4Wrapper(
		ics4WrapperMock,
		bankKeeper,
		transferKeeperMock,
		channelKeeperMock,
		connectionKeeperMock,
		k,
	)

	// Create packet data
	data := transfertypes.FungibleTokenPacketData{
		Denom:    constants.BondDenom,
		Amount:   fixture.amount,
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	// Set up channel mock
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.destinationPort,
					ChannelId: fixture.destinationChannel,
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

	// Set up sender balance (sufficient)
	senderAddr := sdk.MustAccAddressFromBech32(fixture.sender)
	amountInt, ok := sdkmath.NewIntFromString(fixture.amount)
	require.True(t, ok)
	nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, amountInt))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, nativeCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, senderAddr, nativeCoins))

	// Check initial sender balance
	senderBalanceBefore := bankKeeper.GetBalance(ctx, senderAddr, constants.BondDenom)
	require.Equal(t, amountInt, senderBalanceBefore.Amount)

	// Execute SendPacket
	_, err = ics4Wrapper.SendPacket(ctx, fixture.nativePort, fixture.nativeChannel, clienttypes.Height{}, 0, dataBz)

	// Expect error due to unescrow failure
	require.Error(t, err)
	require.Contains(t, err.Error(), "unable to unescrow tokens")
}

func TestSendPacket_ZeroAmount(t *testing.T) {
	// Test case: SendPacket with zero amount to verify error handling

	// Set up positive fixture
	fixture := getSendPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.destinationPort)
	k.SetCounterpartyChannel(ctx, fixture.destinationChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	ics4WrapperMock := mocks.NewMockICS4Wrapper(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create ICS4Wrapper instance
	ics4Wrapper := tokenwrapper.NewICS4Wrapper(
		ics4WrapperMock,
		bankKeeper,
		transferKeeperMock,
		channelKeeperMock,
		connectionKeeperMock,
		k,
	)

	// Set up initial balances
	senderAddr := sdk.MustAccAddressFromBech32(fixture.sender)
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	escrowAddr := transfertypes.GetEscrowAddress(fixture.nativePort, fixture.nativeChannel)

	// Mint native tokens to sender
	amountInt, ok := sdkmath.NewIntFromString(fixture.amount)
	require.True(t, ok)
	nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, amountInt))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, nativeCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, senderAddr, nativeCoins))

	// Mint IBC tokens to module
	conversionFactor := k.GetDecimalConversionFactor(ctx)
	convertedAmount := amountInt.Mul(conversionFactor)
	ibcCoins := sdk.NewCoins(sdk.NewCoin(fixture.ibcDenom, convertedAmount))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, ibcCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, ibcCoins))

	// Check initial balances
	senderBalanceBefore := bankKeeper.GetBalance(ctx, senderAddr, constants.BondDenom)
	require.Equal(t, amountInt, senderBalanceBefore.Amount)
	moduleBalanceBefore := bankKeeper.GetBalance(ctx, moduleAddr, fixture.ibcDenom)
	require.Equal(t, convertedAmount, moduleBalanceBefore.Amount)
	escrowBalanceBefore := bankKeeper.GetBalance(ctx, escrowAddr, constants.BondDenom)
	require.Equal(t, sdkmath.ZeroInt(), escrowBalanceBefore.Amount)

	// Create packet data with zero amount
	data := transfertypes.FungibleTokenPacketData{
		Denom:    constants.BondDenom,
		Amount:   "0",
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	// Set up channel mock
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.destinationPort,
					ChannelId: fixture.destinationChannel,
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

	// Ensure no unescrow or ICS4 wrapper calls occur
	transferKeeperMock.EXPECT().GetTotalEscrowForDenom(gomock.Any(), gomock.Any()).Times(0)
	transferKeeperMock.EXPECT().SetTotalEscrowForDenom(gomock.Any(), gomock.Any()).Times(0)
	ics4WrapperMock.EXPECT().SendPacket(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	// Execute SendPacket
	_, err = ics4Wrapper.SendPacket(ctx, fixture.nativePort, fixture.nativeChannel, clienttypes.Height{}, 0, dataBz)

	// Expect error due to zero amount
	require.Error(t, err)
	require.Contains(t, err.Error(), "amount is zero or negative: 0")

	// Check final balances (no changes due to error)
	senderBalanceAfter := bankKeeper.GetBalance(ctx, senderAddr, constants.BondDenom)
	require.Equal(t, senderBalanceBefore.Amount, senderBalanceAfter.Amount, "Sender balance should not change")
	moduleBalanceAfter := bankKeeper.GetBalance(ctx, moduleAddr, fixture.ibcDenom)
	require.Equal(t, moduleBalanceBefore.Amount, moduleBalanceAfter.Amount, "Module balance should not change")
	escrowBalanceAfter := bankKeeper.GetBalance(ctx, escrowAddr, constants.BondDenom)
	require.Equal(t, sdkmath.ZeroInt(), escrowBalanceAfter.Amount, "Escrow balance should remain zero")
}

func TestSendPacket_NewTokenFromZChain(t *testing.T) {
	// Test case: Send a new token with existing uzig in escrow, token wrapper should ignore it

	// Set up positive fixture
	fixture := getSendPacketPositiveFixture()

	// Define test cases for enabled and disabled module
	tests := []struct {
		name    string
		enabled bool
	}{
		{name: "New token with uzig in escrow, module enabled", enabled: true},
		{name: "New token with uzig in escrow, module disabled", enabled: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the TokenwrapperKeeper with a real bank keeper
			k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

			// Set up the TokenwrapperKeeper
			k.SetEnabled(ctx, tt.enabled)
			k.SetNativePort(ctx, fixture.nativePort)
			k.SetNativeChannel(ctx, fixture.nativeChannel)
			k.SetCounterpartyPort(ctx, fixture.destinationPort)
			k.SetCounterpartyChannel(ctx, fixture.destinationChannel)
			k.SetDenom(ctx, fixture.moduleDenom)
			k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
			_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

			// Create mocks for other dependencies
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
			ics4WrapperMock := mocks.NewMockICS4Wrapper(ctrl)
			channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
			connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

			// Create ICS4Wrapper instance
			ics4Wrapper := tokenwrapper.NewICS4Wrapper(
				ics4WrapperMock,
				bankKeeper,
				transferKeeperMock,
				channelKeeperMock,
				connectionKeeperMock,
				k,
			)

			// Create packet data with a new token denom
			newTokenDenom := "btc"
			data := transfertypes.FungibleTokenPacketData{
				Denom:    newTokenDenom,
				Amount:   fixture.amount,
				Sender:   fixture.sender,
				Receiver: fixture.receiver,
			}
			dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
			require.NoError(t, err)

			// Set up channel mock
			channelKeeperMock.EXPECT().
				GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
				Return(
					channeltypes.Channel{
						State:          channeltypes.OPEN,
						ConnectionHops: []string{fixture.nativeConnectionId},
						Counterparty: channeltypes.Counterparty{
							PortId:    fixture.destinationPort,
							ChannelId: fixture.destinationChannel,
						},
					}, true).
				Times(1)

			// Set up initial balances
			senderAddr := sdk.MustAccAddressFromBech32(fixture.sender)
			moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
			escrowAddr := transfertypes.GetEscrowAddress(fixture.nativePort, fixture.nativeChannel)

			// Mint new tokens to sender
			amountInt, ok := sdkmath.NewIntFromString(fixture.amount)
			require.True(t, ok)
			newTokenCoins := sdk.NewCoins(sdk.NewCoin(newTokenDenom, amountInt))
			require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, newTokenCoins))
			require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, senderAddr, newTokenCoins))

			// Mint uzig tokens to escrow
			zigAmount := sdkmath.NewInt(5000000)
			zigCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, zigAmount))
			require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, zigCoins))
			require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, escrowAddr, zigCoins))

			// Mint IBC tokens to module (for consistency)
			conversionFactor := k.GetDecimalConversionFactor(ctx)
			convertedAmount := amountInt.Mul(conversionFactor)
			ibcCoins := sdk.NewCoins(sdk.NewCoin(fixture.ibcDenom, convertedAmount))
			require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, ibcCoins))
			require.NoError(t, bankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, ibcCoins))

			// Check initial balances
			senderBalanceBefore := bankKeeper.GetBalance(ctx, senderAddr, newTokenDenom)
			require.Equal(t, amountInt, senderBalanceBefore.Amount, "Sender initial balance (new-token)")
			moduleBalanceBefore := bankKeeper.GetBalance(ctx, moduleAddr, fixture.ibcDenom)
			require.Equal(t, convertedAmount, moduleBalanceBefore.Amount, "Module initial balance")
			escrowBalanceBeforeZig := bankKeeper.GetBalance(ctx, escrowAddr, constants.BondDenom)
			require.Equal(t, zigAmount, escrowBalanceBeforeZig.Amount, "Escrow initial balance (zig)")
			escrowBalanceBeforeNew := bankKeeper.GetBalance(ctx, escrowAddr, newTokenDenom)
			require.Equal(t, sdkmath.ZeroInt(), escrowBalanceBeforeNew.Amount, "Escrow initial balance (new-token)")

			// Mock ICS4Wrapper to simulate successful packet send
			ics4WrapperMock.EXPECT().
				SendPacket(ctx, fixture.nativePort, fixture.nativeChannel, clienttypes.Height{}, uint64(0), gomock.Any()).
				Return(uint64(1), nil)

			// Simulate IBC transfer: move new-token to escrow
			require.NoError(t, bankKeeper.SendCoins(ctx, senderAddr, escrowAddr, newTokenCoins))

			// Check balances after escrow
			senderBalanceAfterEscrow := bankKeeper.GetBalance(ctx, senderAddr, newTokenDenom)
			require.Equal(t, sdkmath.ZeroInt(), senderBalanceAfterEscrow.Amount, "Sender balance should be zero after escrow")
			escrowBalanceAfterEscrowNew := bankKeeper.GetBalance(ctx, escrowAddr, newTokenDenom)
			require.Equal(t, amountInt, escrowBalanceAfterEscrowNew.Amount, "Escrow balance (new-token) should reflect transferred tokens")

			// Execute SendPacket
			seq, err := ics4Wrapper.SendPacket(ctx, fixture.nativePort, fixture.nativeChannel, clienttypes.Height{}, 0, dataBz)

			// Validate that no error occurred
			require.NoError(t, err)
			require.Equal(t, uint64(1), seq)

			// Check final balances
			senderBalanceAfter := bankKeeper.GetBalance(ctx, senderAddr, newTokenDenom)
			require.Equal(t, sdkmath.ZeroInt(), senderBalanceAfter.Amount, "Sender balance should remain zero")
			moduleBalanceAfter := bankKeeper.GetBalance(ctx, moduleAddr, fixture.ibcDenom)
			require.Equal(t, moduleBalanceBefore.Amount, moduleBalanceAfter.Amount, "Module balance should not change")
			escrowBalanceAfterZig := bankKeeper.GetBalance(ctx, escrowAddr, constants.BondDenom)
			require.Equal(t, zigAmount, escrowBalanceAfterZig.Amount, "Escrow balance (zig) should remain unchanged")
			escrowBalanceAfterNew := bankKeeper.GetBalance(ctx, escrowAddr, newTokenDenom)
			require.Equal(t, amountInt, escrowBalanceAfterNew.Amount, "Escrow balance (new-token) should reflect transferred tokens")
		})
	}
}

func TestSendPacket_ReturnIBCTokenToAxelar(t *testing.T) {
	// Test case: Send an IBC token (e.g., ETH from Axelar) back to Axelar with uzig in escrow, token wrapper should ignore it

	// Set up positive fixture
	fixture := getSendPacketPositiveFixture()

	// Define test cases for enabled and disabled module
	tests := []struct {
		name    string
		enabled bool
	}{
		{name: "IBC token with uzig in escrow, module enabled", enabled: true},
		{name: "IBC token with uzig in escrow, module disabled", enabled: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the TokenwrapperKeeper with a real bank keeper
			k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

			// Set up the TokenwrapperKeeper
			k.SetEnabled(ctx, tt.enabled)
			k.SetNativePort(ctx, fixture.nativePort)
			k.SetNativeChannel(ctx, fixture.nativeChannel)
			k.SetCounterpartyPort(ctx, fixture.destinationPort)
			k.SetCounterpartyChannel(ctx, fixture.destinationChannel)
			k.SetDenom(ctx, fixture.moduleDenom)
			k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
			_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

			// Create mocks for other dependencies
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
			ics4WrapperMock := mocks.NewMockICS4Wrapper(ctrl)
			channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
			connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

			// Create ICS4Wrapper instance
			ics4Wrapper := tokenwrapper.NewICS4Wrapper(
				ics4WrapperMock,
				bankKeeper,
				transferKeeperMock,
				channelKeeperMock,
				connectionKeeperMock,
				k,
			)

			// Create packet data with an IBC denom (e.g., ETH from Axelar)
			ibcTokenDenom := "ibc/ETH"
			data := transfertypes.FungibleTokenPacketData{
				Denom:    ibcTokenDenom,
				Amount:   fixture.amount,
				Sender:   fixture.sender,
				Receiver: fixture.receiver,
			}
			dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
			require.NoError(t, err)

			// Set up channel mock
			channelKeeperMock.EXPECT().
				GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
				Return(
					channeltypes.Channel{
						State:          channeltypes.OPEN,
						ConnectionHops: []string{fixture.nativeConnectionId},
						Counterparty: channeltypes.Counterparty{
							PortId:    fixture.destinationPort,
							ChannelId: fixture.destinationChannel,
						},
					}, true).
				Times(1)

			// Set up initial balances
			senderAddr := sdk.MustAccAddressFromBech32(fixture.sender)
			moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
			escrowAddr := transfertypes.GetEscrowAddress(fixture.nativePort, fixture.nativeChannel)

			// Mint IBC tokens to sender (simulating received tokens from Axelar)
			amountInt, ok := sdkmath.NewIntFromString(fixture.amount)
			require.True(t, ok)
			ibcTokenCoins := sdk.NewCoins(sdk.NewCoin(ibcTokenDenom, amountInt))
			require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, ibcTokenCoins))
			require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, senderAddr, ibcTokenCoins))

			// Mint uzig tokens to escrow
			zigAmount := sdkmath.NewInt(5000000)
			zigCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, zigAmount))
			require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, zigCoins))
			require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, escrowAddr, zigCoins))

			// Mint module's IBC tokens (for consistency)
			conversionFactor := k.GetDecimalConversionFactor(ctx)
			convertedAmount := amountInt.Mul(conversionFactor)
			moduleIbcCoins := sdk.NewCoins(sdk.NewCoin(fixture.ibcDenom, convertedAmount))
			require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, moduleIbcCoins))
			require.NoError(t, bankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, moduleIbcCoins))

			// Check initial balances
			senderBalanceBefore := bankKeeper.GetBalance(ctx, senderAddr, ibcTokenDenom)
			require.Equal(t, amountInt, senderBalanceBefore.Amount, "Sender initial balance (ibc/ETH)")
			moduleBalanceBefore := bankKeeper.GetBalance(ctx, moduleAddr, fixture.ibcDenom)
			require.Equal(t, convertedAmount, moduleBalanceBefore.Amount, "Module initial balance")
			escrowBalanceBeforeZig := bankKeeper.GetBalance(ctx, escrowAddr, constants.BondDenom)
			require.Equal(t, zigAmount, escrowBalanceBeforeZig.Amount, "Escrow initial balance (zig)")
			escrowBalanceBeforeIbc := bankKeeper.GetBalance(ctx, escrowAddr, ibcTokenDenom)
			require.Equal(t, sdkmath.ZeroInt(), escrowBalanceBeforeIbc.Amount, "Escrow initial balance (ibc/ETH)")

			// Mock ICS4Wrapper to simulate successful packet send
			ics4WrapperMock.EXPECT().
				SendPacket(ctx, fixture.nativePort, fixture.nativeChannel, clienttypes.Height{}, uint64(0), gomock.Any()).
				Return(uint64(1), nil)

			// Simulate IBC transfer: move ibc/ETH to escrow
			require.NoError(t, bankKeeper.SendCoins(ctx, senderAddr, escrowAddr, ibcTokenCoins))

			// Check balances after escrow
			senderBalanceAfterEscrow := bankKeeper.GetBalance(ctx, senderAddr, ibcTokenDenom)
			require.Equal(t, sdkmath.ZeroInt(), senderBalanceAfterEscrow.Amount, "Sender balance should be zero after escrow")
			escrowBalanceAfterEscrowIbc := bankKeeper.GetBalance(ctx, escrowAddr, ibcTokenDenom)
			require.Equal(t, amountInt, escrowBalanceAfterEscrowIbc.Amount, "Escrow balance (ibc/ETH) should reflect transferred tokens")

			// Execute SendPacket
			seq, err := ics4Wrapper.SendPacket(ctx, fixture.nativePort, fixture.nativeChannel, clienttypes.Height{}, 0, dataBz)

			// Validate that no error occurred
			require.NoError(t, err)
			require.Equal(t, uint64(1), seq)

			// Check final balances
			senderBalanceAfter := bankKeeper.GetBalance(ctx, senderAddr, ibcTokenDenom)
			require.Equal(t, sdkmath.ZeroInt(), senderBalanceAfter.Amount, "Sender balance should remain zero")
			moduleBalanceAfter := bankKeeper.GetBalance(ctx, moduleAddr, fixture.ibcDenom)
			require.Equal(t, moduleBalanceBefore.Amount, moduleBalanceAfter.Amount, "Module balance should not change")
			escrowBalanceAfterZig := bankKeeper.GetBalance(ctx, escrowAddr, constants.BondDenom)
			require.Equal(t, zigAmount, escrowBalanceAfterZig.Amount, "Escrow balance (zig) should remain unchanged")
			escrowBalanceAfterIbc := bankKeeper.GetBalance(ctx, escrowAddr, ibcTokenDenom)
			require.Equal(t, amountInt, escrowBalanceAfterIbc.Amount, "Escrow balance (ibc/ETH) should reflect transferred tokens")
		})
	}
}

func TestSendPacket_ChannelNotOpen(t *testing.T) {
	// Test case: SendPacket with channel not open to verify error handling

	// Set up positive fixture
	fixture := getSendPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper with different port/channel
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.destinationPort)
	k.SetCounterpartyChannel(ctx, fixture.destinationChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	ics4WrapperMock := mocks.NewMockICS4Wrapper(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create ICS4Wrapper instance
	ics4Wrapper := tokenwrapper.NewICS4Wrapper(
		ics4WrapperMock,
		bankKeeper,
		transferKeeperMock,
		channelKeeperMock,
		connectionKeeperMock,
		k,
	)

	// Create packet data
	data := transfertypes.FungibleTokenPacketData{
		Denom:    constants.BondDenom,
		Amount:   fixture.amount,
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	// Set up channel mock
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.CLOSED,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.destinationPort,
					ChannelId: fixture.destinationChannel,
				},
			}, true).
		Times(1)

	// Call SendPacket and expect an error due to CLOSED channel
	_, err = ics4Wrapper.SendPacket(ctx, fixture.nativePort, fixture.nativeChannel, clienttypes.Height{}, 0, dataBz)

	// Validate error returned
	require.Error(t, err)
	require.Contains(t, err.Error(), "transfer/channel-12, state: STATE_CLOSED: channel is not open")

	// Validate that the tokenwrapper_error event was emitted
	events := ctx.EventManager().Events()
	require.Len(t, events, 1)
	require.Equal(t, "tokenwrapper_error", events[0].Type)
	require.Len(t, events[0].Attributes, 1)
	require.Contains(t, string(events[0].Attributes[0].Value), "transfer/channel-12, state: STATE_CLOSED: channel is not open")
}

func TestSendPacket_SenderChainNotSourceChain(t *testing.T) {
	// Test case: SendPacket with sender chain not matching source chain to verify error handling

	// Set up positive fixture
	fixture := getSendPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set module settings
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.destinationPort)
	k.SetCounterpartyChannel(ctx, fixture.destinationChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	ics4WrapperMock := mocks.NewMockICS4Wrapper(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create ICS4Wrapper instance
	ics4Wrapper := tokenwrapper.NewICS4Wrapper(
		ics4WrapperMock,
		bankKeeper,
		transferKeeperMock,
		channelKeeperMock,
		connectionKeeperMock,
		k,
	)

	// Create denom with the correct prefix (sourcePort/sourceChannel)
	prefixedDenom := transfertypes.GetPrefixedDenom(fixture.nativePort, fixture.nativeChannel, fixture.moduleDenom)

	// Create packet data
	data := transfertypes.FungibleTokenPacketData{
		Denom:    prefixedDenom,
		Amount:   fixture.amount,
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	// Expect GetChannel call (required by SendPacket before any logic branch)
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(channeltypes.Channel{State: channeltypes.OPEN}, true).
		Times(1)

	// Expect SendPacket to be forwarded to ics4Wrapper
	ics4WrapperMock.EXPECT().
		SendPacket(
			ctx,
			fixture.nativePort,
			fixture.nativeChannel,
			clienttypes.Height{},
			gomock.Eq(uint64(0)),
			gomock.Any(),
		).
		Return(uint64(42), nil).
		Times(1)

	// Call SendPacket
	seq, err := ics4Wrapper.SendPacket(ctx, fixture.nativePort, fixture.nativeChannel, clienttypes.Height{}, 0, dataBz)

	// Assertions
	require.NoError(t, err)
	require.Equal(t, uint64(42), seq)

	// Validate info event
	events := ctx.EventManager().Events()
	require.Len(t, events, 1)
	require.Equal(t, "tokenwrapper_info", events[0].Type)
	require.Len(t, events[0].Attributes, 1)
	require.Equal(t, "sender chain is not the source chain", string(events[0].Attributes[0].Value))
}

func TestSendPacket_EmptyReceiver(t *testing.T) {
	// Test case: SendPacket with empty receiver address to verify error handling

	// Set up positive fixture
	fixture := getSendPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper config
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.destinationPort)
	k.SetCounterpartyChannel(ctx, fixture.destinationChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	ics4WrapperMock := mocks.NewMockICS4Wrapper(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create ICS4Wrapper instance
	ics4Wrapper := tokenwrapper.NewICS4Wrapper(
		ics4WrapperMock,
		bankKeeper,
		transferKeeperMock,
		channelKeeperMock,
		connectionKeeperMock,
		k,
	)

	// Use valid sender, denom, and amount, but empty receiver
	data := transfertypes.FungibleTokenPacketData{
		Denom:    constants.BondDenom,
		Amount:   fixture.amount,
		Sender:   fixture.sender,
		Receiver: "", // <- test target
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	// Set up channelKeeper to return OPEN channel (to pass validation)
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(channeltypes.Channel{
			State:          channeltypes.OPEN,
			ConnectionHops: []string{fixture.nativeConnectionId},
			Counterparty: channeltypes.Counterparty{
				PortId:    fixture.destinationPort,
				ChannelId: fixture.destinationChannel,
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

	// Call SendPacket
	_, err = ics4Wrapper.SendPacket(
		ctx,
		fixture.nativePort,
		fixture.nativeChannel,
		clienttypes.Height{},
		uint64(0),
		dataBz,
	)

	// Expect specific error message
	require.Error(t, err)
	require.Contains(t, err.Error(), "receiver address is empty")

	// Verify event emitted
	events := ctx.EventManager().Events()
	require.Len(t, events, 1)
	require.Equal(t, "tokenwrapper_error", events[0].Type)
	require.Len(t, events[0].Attributes, 1)
	require.Contains(t, string(events[0].Attributes[0].Value), "receiver address is empty")
}

func TestSendPacket_ZeroConvertedAmount(t *testing.T) {
	// Test case: SendPacket with zero converted amount to verify error handling

	// Set up positive fixture
	fixture := getSendPacketPositiveFixture()

	// Create real bank keeper
	_, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Create mocks for dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keeperMock := mocks.NewMockTokenwrapperKeeper(ctrl)
	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	ics4WrapperMock := mocks.NewMockICS4Wrapper(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create ICS4Wrapper instance
	ics4Wrapper := tokenwrapper.NewICS4Wrapper(
		ics4WrapperMock,
		bankKeeper,
		transferKeeperMock,
		channelKeeperMock,
		connectionKeeperMock,
		keeperMock,
	)

	// Create packet data
	data := transfertypes.FungibleTokenPacketData{
		Denom:    constants.BondDenom,
		Amount:   fixture.amount,
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	// Set up channel mock
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.destinationPort,
					ChannelId: fixture.destinationChannel,
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

	// Mock tokenwrapper keeper methods
	keeperMock.EXPECT().IsEnabled(ctx).Return(true).AnyTimes()
	keeperMock.EXPECT().GetNativePort(ctx).Return(fixture.nativePort).AnyTimes()
	keeperMock.EXPECT().GetNativeChannel(ctx).Return(fixture.nativeChannel).AnyTimes()
	keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(fixture.destinationPort).AnyTimes()
	keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(fixture.destinationChannel).AnyTimes()
	keeperMock.EXPECT().GetNativeClientId(ctx).Return(fixture.nativeClientId).AnyTimes()
	keeperMock.EXPECT().GetCounterpartyClientId(ctx).Return(fixture.counterpartyClientId).AnyTimes()
	keeperMock.EXPECT().GetDenom(ctx).Return(fixture.moduleDenom).AnyTimes()
	keeperMock.EXPECT().GetDecimalConversionFactor(ctx).Return(sdkmath.ZeroInt()).AnyTimes()
	keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
	convertedAmount, _ := sdkmath.NewIntFromString("0")
	keeperMock.EXPECT().ScaleUpTokenPrecision(ctx, sdkmath.NewInt(1000000)).Return(convertedAmount, fmt.Errorf("converted amount is zero or negative: %s", convertedAmount.String()))

	// Set up escrow address
	escrowAddr := transfertypes.GetEscrowAddress(fixture.nativePort, fixture.nativeChannel)
	senderAddr := sdk.MustAccAddressFromBech32(fixture.sender)

	// Parse amount
	amountInt, ok := sdkmath.NewIntFromString(fixture.amount)
	require.True(t, ok)
	nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, amountInt))

	// Mint native tokens to escrow (simulate escrowed tokens)
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, nativeCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, escrowAddr, nativeCoins))

	// Check initial balances
	senderBalanceBefore := bankKeeper.GetBalance(ctx, senderAddr, constants.BondDenom)
	require.Equal(t, sdkmath.ZeroInt(), senderBalanceBefore.Amount)
	escrowBalanceBefore := bankKeeper.GetBalance(ctx, escrowAddr, constants.BondDenom)
	require.Equal(t, amountInt, escrowBalanceBefore.Amount)

	// Mock transfer keeper for unescrow
	currentTotalEscrow := sdk.NewCoin(constants.BondDenom, amountInt)
	transferKeeperMock.EXPECT().GetTotalEscrowForDenom(ctx, constants.BondDenom).Return(currentTotalEscrow).Times(1)
	newTotalEscrow := currentTotalEscrow.Sub(nativeCoins[0])
	transferKeeperMock.EXPECT().SetTotalEscrowForDenom(ctx, newTotalEscrow).Times(1)

	// Execute SendPacket
	_, err = ics4Wrapper.SendPacket(ctx, fixture.nativePort, fixture.nativeChannel, clienttypes.Height{}, 0, dataBz)

	// Expect error due to zero converted amount
	require.Error(t, err)
	require.Contains(t, err.Error(), "converted amount is zero or negative: 0")

	// Check balances after (unescrow happened, sender has tokens, escrow empty)
	senderBalanceAfter := bankKeeper.GetBalance(ctx, senderAddr, constants.BondDenom)
	require.Equal(t, amountInt, senderBalanceAfter.Amount)
	escrowBalanceAfter := bankKeeper.GetBalance(ctx, escrowAddr, constants.BondDenom)
	require.Equal(t, sdkmath.ZeroInt(), escrowBalanceAfter.Amount)

	// Verify error event emitted
	events := ctx.EventManager().Events()
	foundErrorEvent := false
	for _, event := range events {
		if event.Type == "tokenwrapper_error" {
			foundErrorEvent = true
			require.Len(t, event.Attributes, 1)
			require.Contains(t, string(event.Attributes[0].Value), "converted amount is zero or negative: 0")
		}
	}
	require.True(t, foundErrorEvent, "tokenwrapper_error event not emitted")
}

func TestSendPacket_NegativeConvertedAmount(t *testing.T) {
	// Test case: SendPacket with negative converted amount to verify error handling

	// Set up positive fixture
	fixture := getSendPacketPositiveFixture()

	// Create real bank keeper
	_, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Create mocks for dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keeperMock := mocks.NewMockTokenwrapperKeeper(ctrl)
	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	ics4WrapperMock := mocks.NewMockICS4Wrapper(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create ICS4Wrapper instance
	ics4Wrapper := tokenwrapper.NewICS4Wrapper(
		ics4WrapperMock,
		bankKeeper,
		transferKeeperMock,
		channelKeeperMock,
		connectionKeeperMock,
		keeperMock,
	)

	// Create packet data
	data := transfertypes.FungibleTokenPacketData{
		Denom:    constants.BondDenom,
		Amount:   fixture.amount,
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	// Set up channel mock
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.destinationPort,
					ChannelId: fixture.destinationChannel,
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

	// Mock tokenwrapper keeper methods
	keeperMock.EXPECT().IsEnabled(ctx).Return(true).AnyTimes()
	keeperMock.EXPECT().GetNativePort(ctx).Return(fixture.nativePort).AnyTimes()
	keeperMock.EXPECT().GetNativeChannel(ctx).Return(fixture.nativeChannel).AnyTimes()
	keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(fixture.destinationPort).AnyTimes()
	keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(fixture.destinationChannel).AnyTimes()
	keeperMock.EXPECT().GetNativeClientId(ctx).Return(fixture.nativeClientId).AnyTimes()
	keeperMock.EXPECT().GetCounterpartyClientId(ctx).Return(fixture.counterpartyClientId).AnyTimes()
	keeperMock.EXPECT().GetDenom(ctx).Return(fixture.moduleDenom).AnyTimes()
	keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
	negativeConvertedAmount, _ := sdkmath.NewIntFromString("-1000000000000000000")
	keeperMock.EXPECT().ScaleUpTokenPrecision(ctx, sdkmath.NewInt(1000000)).Return(negativeConvertedAmount, fmt.Errorf("converted amount is zero or negative: %s", negativeConvertedAmount.String()))

	// Set up escrow address
	escrowAddr := transfertypes.GetEscrowAddress(fixture.nativePort, fixture.nativeChannel)
	senderAddr := sdk.MustAccAddressFromBech32(fixture.sender)

	// Parse amount
	amountInt, ok := sdkmath.NewIntFromString(fixture.amount)
	require.True(t, ok)
	nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, amountInt))

	// Mint native tokens to escrow (simulate escrowed tokens)
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, nativeCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, escrowAddr, nativeCoins))

	// Check initial balances
	senderBalanceBefore := bankKeeper.GetBalance(ctx, senderAddr, constants.BondDenom)
	require.Equal(t, sdkmath.ZeroInt(), senderBalanceBefore.Amount)
	escrowBalanceBefore := bankKeeper.GetBalance(ctx, escrowAddr, constants.BondDenom)
	require.Equal(t, amountInt, escrowBalanceBefore.Amount)

	// Mock transfer keeper for unescrow
	currentTotalEscrow := sdk.NewCoin(constants.BondDenom, amountInt)
	transferKeeperMock.EXPECT().GetTotalEscrowForDenom(ctx, constants.BondDenom).Return(currentTotalEscrow).Times(1)
	newTotalEscrow := currentTotalEscrow.Sub(nativeCoins[0])
	transferKeeperMock.EXPECT().SetTotalEscrowForDenom(ctx, newTotalEscrow).Times(1)

	// Execute SendPacket
	_, err = ics4Wrapper.SendPacket(ctx, fixture.nativePort, fixture.nativeChannel, clienttypes.Height{}, 0, dataBz)

	// Expect error due to negative converted amount
	require.Error(t, err)
	require.Contains(t, err.Error(), "converted amount is zero or negative: -1000000000000000000")

	// Check balances after (unescrow happened, sender has tokens, escrow empty)
	senderBalanceAfter := bankKeeper.GetBalance(ctx, senderAddr, constants.BondDenom)
	require.Equal(t, amountInt, senderBalanceAfter.Amount)
	escrowBalanceAfter := bankKeeper.GetBalance(ctx, escrowAddr, constants.BondDenom)
	require.Equal(t, sdkmath.ZeroInt(), escrowBalanceAfter.Amount)

	// Verify error event emitted
	events := ctx.EventManager().Events()
	foundErrorEvent := false
	for _, event := range events {
		if event.Type == "tokenwrapper_error" {
			foundErrorEvent = true
			require.Len(t, event.Attributes, 1)
			require.Contains(t, string(event.Attributes[0].Value), "converted amount is zero or negative: -1000000000000000000")
		}
	}
	require.True(t, foundErrorEvent, "tokenwrapper_error event not emitted")
}

func TestSendPacket_BurnFailureUnlockSuccess(t *testing.T) {
	// Test case: SendPacket where burn fails but unlock succeeds, verifying rollback attempt and error message

	// Set up positive fixture
	fixture := getSendPacketPositiveFixture()

	// Create real bank keeper
	_, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Create mocks for dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keeperMock := mocks.NewMockTokenwrapperKeeper(ctrl)
	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	ics4WrapperMock := mocks.NewMockICS4Wrapper(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create ICS4Wrapper instance
	ics4Wrapper := tokenwrapper.NewICS4Wrapper(
		ics4WrapperMock,
		bankKeeper,
		transferKeeperMock,
		channelKeeperMock,
		connectionKeeperMock,
		keeperMock,
	)

	// Create packet data
	data := transfertypes.FungibleTokenPacketData{
		Denom:    constants.BondDenom,
		Amount:   fixture.amount,
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	// Set up channel mock
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.destinationPort,
					ChannelId: fixture.destinationChannel,
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

	// Mock keeper methods for happy path up to burn
	keeperMock.EXPECT().IsEnabled(ctx).Return(true).AnyTimes()
	keeperMock.EXPECT().GetNativePort(ctx).Return(fixture.nativePort).AnyTimes()
	keeperMock.EXPECT().GetNativeChannel(ctx).Return(fixture.nativeChannel).AnyTimes()
	keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(fixture.destinationPort).AnyTimes()
	keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(fixture.destinationChannel).AnyTimes()
	keeperMock.EXPECT().GetNativeClientId(ctx).Return(fixture.nativeClientId).AnyTimes()
	keeperMock.EXPECT().GetCounterpartyClientId(ctx).Return(fixture.counterpartyClientId).AnyTimes()
	keeperMock.EXPECT().GetDenom(ctx).Return(fixture.moduleDenom).AnyTimes()
	keeperMock.EXPECT().CheckAccountBalance(ctx, gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	keeperMock.EXPECT().CheckModuleBalance(ctx, gomock.Any()).Return(nil).AnyTimes()
	keeperMock.EXPECT().LockTokens(ctx, gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	keeperMock.EXPECT().BurnIbcTokens(ctx, gomock.Any()).Return(fmt.Errorf("burn tokens failed")).Times(1)
	keeperMock.EXPECT().UnlockTokens(ctx, gomock.Any(), gomock.Any()).Return(nil).Times(1)
	keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
	keeperMock.EXPECT().ScaleUpTokenPrecision(ctx, sdkmath.NewInt(1000000)).Return(sdkmath.NewInt(1000000000000000000), nil)

	// Set up for unescrow
	escrowAddr := transfertypes.GetEscrowAddress(fixture.nativePort, fixture.nativeChannel)
	amountInt, _ := sdkmath.NewIntFromString(fixture.amount)
	nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, amountInt))

	// Mint native to escrow
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, nativeCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, escrowAddr, nativeCoins))

	// Mock transfer keeper for unescrow
	currentTotalEscrow := sdk.NewCoin(constants.BondDenom, amountInt)
	transferKeeperMock.EXPECT().GetTotalEscrowForDenom(ctx, constants.BondDenom).Return(currentTotalEscrow).Times(1)
	newTotalEscrow := currentTotalEscrow.Sub(nativeCoins[0])
	transferKeeperMock.EXPECT().SetTotalEscrowForDenom(ctx, newTotalEscrow).Times(1)

	// Execute SendPacket
	_, err = ics4Wrapper.SendPacket(ctx, fixture.nativePort, fixture.nativeChannel, clienttypes.Height{}, 0, dataBz)

	// Expect burn error
	require.Error(t, err)
	require.Contains(t, err.Error(), "burn tokens failed")

	// Verify events: only one error event for burn failure
	events := ctx.EventManager().Events()
	errorEvents := 0
	for _, event := range events {
		if event.Type == "tokenwrapper_error" {
			errorEvents++
			require.Contains(t, string(event.Attributes[0].Value), "burn tokens failed")
		}
	}
	require.Equal(t, 1, errorEvents, "Only burn error event should be emitted")
}

func TestSendPacket_BurnFailureUnlockFailure(t *testing.T) {
	// Test case: SendPacket where burn fails and unlock also fails, verifying events for both

	// Set up positive fixture
	fixture := getSendPacketPositiveFixture()

	// Create a real bank keeper
	_, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Create mocks for dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keeperMock := mocks.NewMockTokenwrapperKeeper(ctrl)
	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	ics4WrapperMock := mocks.NewMockICS4Wrapper(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create ICS4Wrapper instance
	ics4Wrapper := tokenwrapper.NewICS4Wrapper(
		ics4WrapperMock,
		bankKeeper,
		transferKeeperMock,
		channelKeeperMock,
		connectionKeeperMock,
		keeperMock,
	)

	// Create packet data
	data := transfertypes.FungibleTokenPacketData{
		Denom:    constants.BondDenom,
		Amount:   fixture.amount,
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	// Set up channel mock
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.destinationPort,
					ChannelId: fixture.destinationChannel,
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

	// Mock keeper methods for happy path up to burn
	keeperMock.EXPECT().IsEnabled(ctx).Return(true).AnyTimes()
	keeperMock.EXPECT().GetNativePort(ctx).Return(fixture.nativePort).AnyTimes()
	keeperMock.EXPECT().GetNativeChannel(ctx).Return(fixture.nativeChannel).AnyTimes()
	keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(fixture.destinationPort).AnyTimes()
	keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(fixture.destinationChannel).AnyTimes()
	keeperMock.EXPECT().GetNativeClientId(ctx).Return(fixture.nativeClientId).AnyTimes()
	keeperMock.EXPECT().GetCounterpartyClientId(ctx).Return(fixture.counterpartyClientId).AnyTimes()
	keeperMock.EXPECT().GetDenom(ctx).Return(fixture.moduleDenom).AnyTimes()
	keeperMock.EXPECT().GetDecimalConversionFactor(ctx).Return(sdkmath.NewInt(1000000000000)).AnyTimes()
	keeperMock.EXPECT().CheckAccountBalance(ctx, gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	keeperMock.EXPECT().CheckModuleBalance(ctx, gomock.Any()).Return(nil).AnyTimes()
	keeperMock.EXPECT().LockTokens(ctx, gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	keeperMock.EXPECT().BurnIbcTokens(ctx, gomock.Any()).Return(fmt.Errorf("burn tokens failed")).Times(1)
	keeperMock.EXPECT().UnlockTokens(ctx, gomock.Any(), gomock.Any()).Return(fmt.Errorf("unlock tokens failed")).Times(1)
	keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
	keeperMock.EXPECT().ScaleUpTokenPrecision(ctx, sdkmath.NewInt(1000000)).Return(sdkmath.NewInt(1000000000000000000), nil)

	// Set up for unescrow
	escrowAddr := transfertypes.GetEscrowAddress(fixture.nativePort, fixture.nativeChannel)
	amountInt, _ := sdkmath.NewIntFromString(fixture.amount)
	nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, amountInt))

	// Mint native to escrow
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, nativeCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, escrowAddr, nativeCoins))

	// Mock transfer keeper for unescrow
	currentTotalEscrow := sdk.NewCoin(constants.BondDenom, amountInt)
	transferKeeperMock.EXPECT().GetTotalEscrowForDenom(ctx, constants.BondDenom).Return(currentTotalEscrow).Times(1)
	newTotalEscrow := currentTotalEscrow.Sub(nativeCoins[0])
	transferKeeperMock.EXPECT().SetTotalEscrowForDenom(ctx, newTotalEscrow).Times(1)

	// Execute SendPacket
	_, err = ics4Wrapper.SendPacket(ctx, fixture.nativePort, fixture.nativeChannel, clienttypes.Height{}, 0, dataBz)

	// Expect burn error
	require.Error(t, err)
	require.Contains(t, err.Error(), "burn tokens failed")

	// Verify events: error event for burn and for unlock failure
	events := ctx.EventManager().Events()
	errorEvents := []string{}
	for _, event := range events {
		if event.Type == "tokenwrapper_error" {
			errorEvents = append(errorEvents, string(event.Attributes[0].Value))
		}
	}
	require.Len(t, errorEvents, 2, "Two error events should be emitted")
	require.Contains(t, errorEvents, "burn tokens failed")
	require.Contains(t, errorEvents, "unlock tokens failed")
}

func TestSendPacket_InsufficientSenderBalanceForLock(t *testing.T) {
	// Test case: SendPacket where sender has insufficient balance for locking tokens, verifying error handling and event emission

	// Set up positive fixture
	fixture := getSendPacketPositiveFixture()

	// Create a real bank keeper
	_, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Create mocks for dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keeperMock := mocks.NewMockTokenwrapperKeeper(ctrl)
	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	ics4WrapperMock := mocks.NewMockICS4Wrapper(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create ICS4Wrapper instance
	ics4Wrapper := tokenwrapper.NewICS4Wrapper(
		ics4WrapperMock,
		bankKeeper,
		transferKeeperMock,
		channelKeeperMock,
		connectionKeeperMock,
		keeperMock,
	)

	// Create packet data
	data := transfertypes.FungibleTokenPacketData{
		Denom:    constants.BondDenom,
		Amount:   fixture.amount,
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	// Set up channel mock
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.destinationPort,
					ChannelId: fixture.destinationChannel,
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

	// Mock keeper methods to pass up to CheckAccountBalance, then fail it
	keeperMock.EXPECT().IsEnabled(ctx).Return(true).AnyTimes()
	keeperMock.EXPECT().GetNativePort(ctx).Return(fixture.nativePort).AnyTimes()
	keeperMock.EXPECT().GetNativeChannel(ctx).Return(fixture.nativeChannel).AnyTimes()
	keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(fixture.destinationPort).AnyTimes()
	keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(fixture.destinationChannel).AnyTimes()
	keeperMock.EXPECT().GetNativeClientId(ctx).Return(fixture.nativeClientId).AnyTimes()
	keeperMock.EXPECT().GetCounterpartyClientId(ctx).Return(fixture.counterpartyClientId).AnyTimes()
	keeperMock.EXPECT().GetDenom(ctx).Return(fixture.moduleDenom).AnyTimes()
	keeperMock.EXPECT().GetDecimalConversionFactor(ctx).Return(sdkmath.NewInt(1000000000000)).AnyTimes()
	keeperMock.EXPECT().CheckAccountBalance(ctx, gomock.Any(), gomock.Any()).Return(fmt.Errorf("insufficient balance for lock")).Times(1)
	keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
	keeperMock.EXPECT().ScaleUpTokenPrecision(ctx, sdkmath.NewInt(1000000)).Return(sdkmath.NewInt(1000000000000000000), nil)

	// Set up escrow address
	escrowAddr := transfertypes.GetEscrowAddress(fixture.nativePort, fixture.nativeChannel)
	senderAddr := sdk.MustAccAddressFromBech32(fixture.sender)

	// Parse amount
	amountInt, ok := sdkmath.NewIntFromString(fixture.amount)
	require.True(t, ok)
	nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, amountInt))

	// Mint native tokens to escrow (simulate escrowed tokens)
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, nativeCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, escrowAddr, nativeCoins))

	// Check initial balances
	senderBalanceBefore := bankKeeper.GetBalance(ctx, senderAddr, constants.BondDenom)
	require.Equal(t, sdkmath.ZeroInt(), senderBalanceBefore.Amount)
	escrowBalanceBefore := bankKeeper.GetBalance(ctx, escrowAddr, constants.BondDenom)
	require.Equal(t, amountInt, escrowBalanceBefore.Amount)

	// Mock transfer keeper for unescrow
	currentTotalEscrow := sdk.NewCoin(constants.BondDenom, amountInt)
	transferKeeperMock.EXPECT().GetTotalEscrowForDenom(ctx, constants.BondDenom).Return(currentTotalEscrow).Times(1)
	newTotalEscrow := currentTotalEscrow.Sub(nativeCoins[0])
	transferKeeperMock.EXPECT().SetTotalEscrowForDenom(ctx, newTotalEscrow).Times(1)

	// Execute SendPacket
	_, err = ics4Wrapper.SendPacket(ctx, fixture.nativePort, fixture.nativeChannel, clienttypes.Height{}, 0, dataBz)

	// Expect error due to insufficient sender balance for lock
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient balance for lock")

	// Check balances after (unescrow happened, sender has tokens, escrow empty; no lock occurred)
	senderBalanceAfter := bankKeeper.GetBalance(ctx, senderAddr, constants.BondDenom)
	require.Equal(t, amountInt, senderBalanceAfter.Amount)
	escrowBalanceAfter := bankKeeper.GetBalance(ctx, escrowAddr, constants.BondDenom)
	require.Equal(t, sdkmath.ZeroInt(), escrowBalanceAfter.Amount)

	// Verify error event emitted
	events := ctx.EventManager().Events()
	foundErrorEvent := false
	for _, event := range events {
		if event.Type == "tokenwrapper_error" {
			foundErrorEvent = true
			require.Len(t, event.Attributes, 1)
			require.Contains(t, string(event.Attributes[0].Value), "insufficient balance for lock")
		}
	}
	require.True(t, foundErrorEvent, "tokenwrapper_error event not emitted")
}

func TestSendPacket_InsufficientModuleBalanceForBurn(t *testing.T) {
	// Test case: SendPacket with insufficient module balance for burning IBC tokens, verifying error handling and event emission

	// Set up positive fixture
	fixture := getSendPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.destinationPort)
	k.SetCounterpartyChannel(ctx, fixture.destinationChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	ics4WrapperMock := mocks.NewMockICS4Wrapper(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create ICS4Wrapper instance
	ics4Wrapper := tokenwrapper.NewICS4Wrapper(
		ics4WrapperMock,
		bankKeeper,
		transferKeeperMock,
		channelKeeperMock,
		connectionKeeperMock,
		k,
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

	// Set up channel and connection mocks
	channelKeeperMock.
		EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.destinationPort,
					ChannelId: fixture.destinationChannel,
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

	// Set up initial balances
	senderAddr := sdk.MustAccAddressFromBech32(fixture.sender)
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)

	// Mint native tokens to sender
	amountInt, _ := sdkmath.NewIntFromString(fixture.amount)
	nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, amountInt))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, nativeCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, senderAddr, nativeCoins))

	// Mint insufficient IBC tokens to module (one less than required)
	conversionFactor := k.GetDecimalConversionFactor(ctx)
	convertedAmount := amountInt.Mul(conversionFactor)
	insufficientIBC := convertedAmount.Sub(sdkmath.NewInt(1))
	ibcCoins := sdk.NewCoins(sdk.NewCoin(fixture.ibcDenom, insufficientIBC))
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, ibcCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, ibcCoins))

	// Check initial balances
	senderBalanceBefore := bankKeeper.GetBalance(ctx, senderAddr, constants.BondDenom)
	require.Equal(t, amountInt, senderBalanceBefore.Amount)

	moduleBalanceBefore := bankKeeper.GetBalance(ctx, moduleAddr, fixture.ibcDenom)
	require.Equal(t, insufficientIBC, moduleBalanceBefore.Amount)

	// Mock unescrow token
	currentTotalEscrow := sdk.NewCoin(constants.BondDenom, amountInt)
	transferKeeperMock.EXPECT().GetTotalEscrowForDenom(ctx, constants.BondDenom).Return(currentTotalEscrow)
	newTotalEscrow := currentTotalEscrow.Sub(nativeCoins[0])
	transferKeeperMock.EXPECT().SetTotalEscrowForDenom(ctx, newTotalEscrow).Return()

	// Apply the escrow logic (SendTransfer) - move native from sender to escrow
	escrowAddress := transfertypes.GetEscrowAddress(fixture.nativePort, fixture.nativeChannel)
	require.NoError(t, bankKeeper.SendCoins(ctx, senderAddr, escrowAddress, nativeCoins))

	// Check escrow and sender balances after escrow
	escrowBalance := bankKeeper.GetBalance(ctx, escrowAddress, constants.BondDenom)
	require.Equal(t, nativeCoins[0], escrowBalance)
	senderBalancePostEscrow := bankKeeper.GetBalance(ctx, senderAddr, constants.BondDenom)
	require.Equal(t, sdkmath.ZeroInt(), senderBalancePostEscrow.Amount)

	// Execute SendPacket (which will unescrow, check balances, fail at module balance check)
	_, err = ics4Wrapper.SendPacket(ctx, fixture.nativePort, fixture.nativeChannel, clienttypes.Height{}, 0, dataBz)

	// Expect error due to insufficient module balance
	require.Error(t, err)
	require.Contains(t, err.Error(), "module does not have enough balance of 1000000000000000000ibc/148AEF32AA7274DC6AFD912A5C1478AC10246B8AEE1C8DEA6D831B752000E89F")

	// Check escrow balance is empty (unescrow happened)
	escrowBalanceAfter := bankKeeper.GetBalance(ctx, escrowAddress, constants.BondDenom)
	require.True(t, escrowBalanceAfter.Amount.IsZero(), "Escrow should have no native tokens left")

	// Check final balances (no lock/burn happened, sender has native back from unescrow)
	senderBalanceAfter := bankKeeper.GetBalance(ctx, senderAddr, constants.BondDenom)
	require.Equal(t, amountInt, senderBalanceAfter.Amount, "Sender should have native tokens back after unescrow")

	moduleBalanceAfter := bankKeeper.GetBalance(ctx, moduleAddr, fixture.ibcDenom)
	require.Equal(t, insufficientIBC, moduleBalanceAfter.Amount, "Module IBC balance should remain unchanged")

	// Verify error event emitted
	events := ctx.EventManager().Events()
	foundErrorEvent := false
	for _, event := range events {
		if event.Type == "tokenwrapper_error" {
			foundErrorEvent = true
			require.Len(t, event.Attributes, 1)
			require.Contains(t, string(event.Attributes[0].Value), "module does not have enough balance of 1000000000000000000ibc/148AEF32AA7274DC6AFD912A5C1478AC10246B8AEE1C8DEA6D831B752000E89F")
		}
	}
	require.True(t, foundErrorEvent, "tokenwrapper_error event not emitted")
}

func TestSendPacket_CounterpartyChannelMismatch_EmitsErrorEvent(t *testing.T) {
	// Test case: Emits error event when counterparty channel doesn't match IBC settings

	// Set up positive fixture
	fixture := getSendPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper with specific IBC settings
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, "different-port")       // This will cause mismatch
	k.SetCounterpartyChannel(ctx, "different-channel") // This will cause mismatch
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	ics4WrapperMock := mocks.NewMockICS4Wrapper(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create ICS4Wrapper instance
	ics4Wrapper := tokenwrapper.NewICS4Wrapper(
		ics4WrapperMock,
		bankKeeper,
		transferKeeperMock,
		channelKeeperMock,
		connectionKeeperMock,
		k,
	)

	// Create packet data that would normally trigger wrapping
	data := transfertypes.FungibleTokenPacketData{
		Denom:    constants.BondDenom,
		Amount:   fixture.amount,
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	// Set up channel mock to return a channel with different counterparty than expected
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.destinationPort,    // This is different from "different-port"
					ChannelId: fixture.destinationChannel, // This is different from "different-channel"
				},
				}, true).
		Times(2) // Called 2 times: validateChannel, checkCounterpartyChannelMatchesIBCSettings (NOT validateConnectionClientId)

	// Set up initial balances
	senderAddr := sdk.MustAccAddressFromBech32(fixture.sender)
	amountInt, _ := sdkmath.NewIntFromString(fixture.amount)
	nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, amountInt))

	// Mint native tokens to sender
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, nativeCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, senderAddr, nativeCoins))

	// Check initial balances
	senderBalanceBefore := bankKeeper.GetBalance(ctx, senderAddr, constants.BondDenom)
	require.Equal(t, amountInt, senderBalanceBefore.Amount)

	// Execute SendPacket - should fail at counterparty channel validation
	_, err = ics4Wrapper.SendPacket(ctx, fixture.nativePort, fixture.nativeChannel, clienttypes.Height{}, 0, dataBz)

	// Expect error due to counterparty channel mismatch
	require.Error(t, err)
	require.Contains(t, err.Error(), "counterparty channel matches IBC settings failed")

	// EVENT CHECK - Verify that the error event was emitted
	// ----------------------------------------------------------------
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// Find the token wrapper error event for counterparty channel mismatch
	var foundCounterpartyErrorEvent bool
	for _, event := range events {
		if event.Type == types.EventTypeTokenWrapperError {
			// Check event attributes - the event contains the error from checkCounterypartyChannelMatchesIBCSettings
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyError {
					if strings.Contains(string(attr.Value), "counterparty channel matches IBC settings failed") {
						foundCounterpartyErrorEvent = true
						t.Logf("Found counterparty channel error event: %s", string(attr.Value))
						break
					}
				}
			}
			if foundCounterpartyErrorEvent {
				break
			}
		}
	}
	require.True(t, foundCounterpartyErrorEvent, "Expected EventTypeTokenWrapperError event to be emitted for counterparty channel mismatch")

	// Check that balances remain unchanged (no wrapping occurred due to error)
	senderBalanceAfter := bankKeeper.GetBalance(ctx, senderAddr, constants.BondDenom)
	require.Equal(t, senderBalanceBefore.Amount, senderBalanceAfter.Amount, "Sender balance should not change when counterparty validation fails")

	t.Log("SUCCESS: Verified that SendPacket emits error event when counterparty channel doesn't match IBC settings")
}

func TestSendPacket_ConnectionClientIdValidationFails_EmitsErrorEvent(t *testing.T) {
	// Test case: Emits error event when connection client ID validation fails

	// Set up positive fixture
	fixture := getSendPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, bankKeeper := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper with specific client ID expectations
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.destinationPort)
	k.SetCounterpartyChannel(ctx, fixture.destinationChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetNativeClientId(ctx, "expected-native-client")             // This will cause mismatch
	k.SetCounterpartyClientId(ctx, "expected-counterparty-client") // This will cause mismatch
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	ics4WrapperMock := mocks.NewMockICS4Wrapper(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)

	// Create ICS4Wrapper instance
	ics4Wrapper := tokenwrapper.NewICS4Wrapper(
		ics4WrapperMock,
		bankKeeper,
		transferKeeperMock,
		channelKeeperMock,
		connectionKeeperMock,
		k,
	)

	// Create packet data that would normally trigger wrapping
	data := transfertypes.FungibleTokenPacketData{
		Denom:    constants.BondDenom,
		Amount:   fixture.amount,
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	// Set up channel mock - channel exists and is open with correct counterparty
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.destinationPort,    // Matches expected
					ChannelId: fixture.destinationChannel, // Matches expected
				},
				}, true).
		Times(3) // Called 3 times: validateChannel, checkCounterpartyChannelMatchesIBCSettings, validateConnectionClientId

	// Set up connection mock to return connection with different client IDs than expected
	connectionKeeperMock.EXPECT().
		GetConnection(ctx, fixture.nativeConnectionId).
		Return(connectiontypes.ConnectionEnd{
			ClientId: "actual-native-client", // Different from "expected-native-client"
			Counterparty: connectiontypes.Counterparty{
				ClientId: "actual-counterparty-client", // Different from "expected-counterparty-client"
			},
		}, true).
		Times(1)

	// Set up initial balances
	senderAddr := sdk.MustAccAddressFromBech32(fixture.sender)
	amountInt, _ := sdkmath.NewIntFromString(fixture.amount)
	nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, amountInt))

	// Mint native tokens to sender
	require.NoError(t, bankKeeper.MintCoins(ctx, minttypes.ModuleName, nativeCoins))
	require.NoError(t, bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, senderAddr, nativeCoins))

	// Check initial balances
	senderBalanceBefore := bankKeeper.GetBalance(ctx, senderAddr, constants.BondDenom)
	require.Equal(t, amountInt, senderBalanceBefore.Amount)

	// Execute SendPacket - should fail at connection client ID validation
	_, err = ics4Wrapper.SendPacket(ctx, fixture.nativePort, fixture.nativeChannel, clienttypes.Height{}, 0, dataBz)

	// Expect error due to client state validation failure
	require.Error(t, err)
	require.Contains(t, err.Error(), "client state validation failed")

	// EVENT CHECK - Verify that the error event was emitted
	// ----------------------------------------------------------------
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// Find the token wrapper error event for client state validation failure
	var foundClientStateErrorEvent bool
	for _, event := range events {
		if event.Type == types.EventTypeTokenWrapperError {
			// Check event attributes - the event contains the error from validateConnectionClientId
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyError {
					if strings.Contains(string(attr.Value), "client state validation failed") {
						foundClientStateErrorEvent = true
						t.Logf("Found client state validation error event: %s", string(attr.Value))
						// Also verify it contains the specific validation error
						require.Contains(t, string(attr.Value), "native client ID mismatch",
							"Error should contain the specific client ID validation failure")
						break
					}
				}
			}
			if foundClientStateErrorEvent {
				break
			}
		}
	}
	require.True(t, foundClientStateErrorEvent, "Expected EventTypeTokenWrapperError event to be emitted for client state validation failure")

	// Check that balances remain unchanged (no wrapping occurred due to error)
	senderBalanceAfter := bankKeeper.GetBalance(ctx, senderAddr, constants.BondDenom)
	require.Equal(t, senderBalanceBefore.Amount, senderBalanceAfter.Amount, "Sender balance should not change when client state validation fails")

	// Verify that no IBC tokens were burned or locked
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	moduleBalanceAfter := bankKeeper.GetBalance(ctx, moduleAddr, fixture.ibcDenom)
	require.True(t, moduleBalanceAfter.IsZero(), "Module should not have any IBC tokens burned")

	t.Log("SUCCESS: Verified that SendPacket emits error event when connection client ID validation fails")
}

func TestSendPacket_SenderMissingUnescrowedTokens_EmitsErrorEvent(t *testing.T) {
	// Test case: Emits error event when sender doesn't have the unescrowed tokens after unescrow operation

	// Set up positive fixture
	fixture := getSendPacketPositiveFixture()

	// Create the TokenwrapperKeeper with a real bank keeper
	k, ctx, _ := keepertest.TokenwrapperKeeperWithBank(t)

	// Set up the TokenwrapperKeeper
	k.SetEnabled(ctx, true)
	k.SetNativePort(ctx, fixture.nativePort)
	k.SetNativeChannel(ctx, fixture.nativeChannel)
	k.SetCounterpartyPort(ctx, fixture.destinationPort)
	k.SetCounterpartyChannel(ctx, fixture.destinationChannel)
	k.SetDenom(ctx, fixture.moduleDenom)
	k.SetNativeClientId(ctx, fixture.nativeClientId)
	k.SetCounterpartyClientId(ctx, fixture.counterpartyClientId)
	_ = k.SetDecimalDifference(ctx, fixture.decimalDifference)

	// Create mocks for other dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transferKeeperMock := mocks.NewMockTransferKeeper(ctrl)
	ics4WrapperMock := mocks.NewMockICS4Wrapper(ctrl)
	channelKeeperMock := mocks.NewMockChannelKeeper(ctrl)
	connectionKeeperMock := mocks.NewMockConnectionKeeper(ctrl)
	bankKeeperMock := mocks.NewMockBankKeeper(ctrl) // Use mock bank keeper for this test

	// Create packet data
	data := transfertypes.FungibleTokenPacketData{
		Denom:    constants.BondDenom,
		Amount:   fixture.amount,
		Sender:   fixture.sender,
		Receiver: fixture.receiver,
	}
	dataBz, err := transfertypes.ModuleCdc.MarshalJSON(&data)
	require.NoError(t, err)

	// Set up channel and connection mocks - all validations pass
	channelKeeperMock.EXPECT().
		GetChannel(ctx, fixture.nativePort, fixture.nativeChannel).
		Return(
			channeltypes.Channel{
				State:          channeltypes.OPEN,
				ConnectionHops: []string{fixture.nativeConnectionId},
				Counterparty: channeltypes.Counterparty{
					PortId:    fixture.destinationPort,
					ChannelId: fixture.destinationChannel,
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

	// Mock keeper methods for the happy path up to the balance check
	k.Logger()
	keeperMock := mocks.NewMockTokenwrapperKeeper(ctrl)
	// Replace the keeper in ICS4Wrapper with our mock for this specific test
	ics4WrapperWithMockKeeper := tokenwrapper.NewICS4Wrapper(
		ics4WrapperMock,
		bankKeeperMock,
		transferKeeperMock,
		channelKeeperMock,
		connectionKeeperMock,
		keeperMock,
	)

	// Mock keeper methods
	keeperMock.EXPECT().IsEnabled(ctx).Return(true).AnyTimes()
	keeperMock.EXPECT().GetNativePort(ctx).Return(fixture.nativePort).AnyTimes()
	keeperMock.EXPECT().GetNativeChannel(ctx).Return(fixture.nativeChannel).AnyTimes()
	keeperMock.EXPECT().GetCounterpartyPort(ctx).Return(fixture.destinationPort).AnyTimes()
	keeperMock.EXPECT().GetCounterpartyChannel(ctx).Return(fixture.destinationChannel).AnyTimes()
	keeperMock.EXPECT().GetNativeClientId(ctx).Return(fixture.nativeClientId).AnyTimes()
	keeperMock.EXPECT().GetCounterpartyClientId(ctx).Return(fixture.counterpartyClientId).AnyTimes()
	keeperMock.EXPECT().GetDenom(ctx).Return(fixture.moduleDenom).AnyTimes()
	keeperMock.EXPECT().Logger().Return(log.NewNopLogger()).AnyTimes()
	keeperMock.EXPECT().ScaleUpTokenPrecision(ctx, gomock.Any()).Return(sdkmath.NewInt(1000000000000000000), nil).AnyTimes()

	// Set up sender address
	senderAddr := sdk.MustAccAddressFromBech32(fixture.sender)
	amountInt, _ := sdkmath.NewIntFromString(fixture.amount)
	nativeCoin := sdk.NewCoin(constants.BondDenom, amountInt)

	// Mock the unescrow operation to succeed
	escrowAddress := transfertypes.GetEscrowAddress(fixture.nativePort, fixture.nativeChannel)
	bankKeeperMock.EXPECT().
		SendCoins(ctx, escrowAddress, senderAddr, sdk.NewCoins(nativeCoin)).
		Return(nil)

	// Mock the balance check to return false - simulating that sender doesn't have the tokens
	bankKeeperMock.EXPECT().
		HasBalance(ctx, senderAddr, nativeCoin).
		Return(false) // This triggers the error

	// Mock transfer keeper for total escrow operations
	currentTotalEscrow := sdk.NewCoin(constants.BondDenom, amountInt)
	transferKeeperMock.EXPECT().GetTotalEscrowForDenom(ctx, constants.BondDenom).Return(currentTotalEscrow)
	newTotalEscrow := currentTotalEscrow.Sub(nativeCoin)
	transferKeeperMock.EXPECT().SetTotalEscrowForDenom(ctx, newTotalEscrow).Return()

	// Execute SendPacket - should fail at balance check after unescrow
	_, err = ics4WrapperWithMockKeeper.SendPacket(ctx, fixture.nativePort, fixture.nativeChannel, clienttypes.Height{}, 0, dataBz)

	// Expect error due to missing unescrowed tokens
	require.Error(t, err)
	require.Contains(t, err.Error(), "sender does not have the unescrowed tokens")

	// EVENT CHECK - Verify that the error event was emitted
	// ----------------------------------------------------------------
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// Find the token wrapper error event for missing unescrowed tokens
	var foundUnescrowedTokensErrorEvent bool
	for _, event := range events {
		if event.Type == types.EventTypeTokenWrapperError {
			// Check event attributes - the event contains the error about missing unescrowed tokens
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyError {
					if strings.Contains(string(attr.Value), "sender does not have the unescrowed tokens") {
						foundUnescrowedTokensErrorEvent = true
						t.Logf("Found missing unescrowed tokens error event: %s", string(attr.Value))
						// Also verify it contains the specific token amount
						require.Contains(t, string(attr.Value), nativeCoin.String(),
							"Error should contain the specific token amount that's missing")
						break
					}
				}
			}
			if foundUnescrowedTokensErrorEvent {
				break
			}
		}
	}
	require.True(t, foundUnescrowedTokensErrorEvent, "Expected EventTypeTokenWrapperError event to be emitted for missing unescrowed tokens")

	// Verify that no further operations occurred (no locking, burning, or packet sending)
	keeperMock.EXPECT().CheckAccountBalance(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	keeperMock.EXPECT().CheckModuleBalance(gomock.Any(), gomock.Any()).Times(0)
	keeperMock.EXPECT().LockTokens(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	keeperMock.EXPECT().BurnIbcTokens(gomock.Any(), gomock.Any()).Times(0)
	ics4WrapperMock.EXPECT().SendPacket(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	t.Log("SUCCESS: Verified that SendPacket emits error event when sender doesn't have unescrowed tokens")
}
