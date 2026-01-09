package tokenwrapper_test

import (
	"fmt"
	"testing"

	tokenwrapper "zigchain/x/tokenwrapper/module"
	mocks "zigchain/x/tokenwrapper/testutil"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetAppVersion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	keeperMock := mocks.NewMockTokenwrapperKeeper(ctrl)
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
		keeperMock,
	)

	// Test cases
	tests := []struct {
		name            string
		portID          string
		channelID       string
		expectedVersion string
		expectedFound   bool
		setupMocks      func()
	}{
		{
			name:            "Happy path",
			portID:          "transfer",
			channelID:       "channel-1",
			expectedVersion: "ics20-1",
			expectedFound:   true,
			setupMocks: func() {
				ics4WrapperMock.EXPECT().GetAppVersion(
					gomock.Any(),
					"transfer",
					"channel-1",
				).Return("ics20-1", true)
			},
		},
		{
			name:            "Version not found",
			portID:          "transfer",
			channelID:       "channel-1",
			expectedVersion: "",
			expectedFound:   false,
			setupMocks: func() {
				ics4WrapperMock.EXPECT().GetAppVersion(
					gomock.Any(),
					"transfer",
					"channel-1",
				).Return("", false)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			version, found := ics4Wrapper.GetAppVersion(
				sdk.Context{},
				tt.portID,
				tt.channelID,
			)
			require.Equal(t, tt.expectedVersion, version)
			require.Equal(t, tt.expectedFound, found)
		})
	}
}

func TestWriteAcknowledgement(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	keeperMock := mocks.NewMockTokenwrapperKeeper(ctrl)
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
		keeperMock,
	)

	// Test cases
	tests := []struct {
		name          string
		packet        ibcexported.PacketI
		ack           ibcexported.Acknowledgement
		expectedError error
		setupMocks    func()
	}{
		{
			name:          "Happy path",
			packet:        channeltypes.Packet{},
			ack:           channeltypes.NewResultAcknowledgement([]byte("success")),
			expectedError: nil,
			setupMocks: func() {
				ics4WrapperMock.EXPECT().WriteAcknowledgement(
					gomock.Any(),
					channeltypes.Packet{},
					channeltypes.NewResultAcknowledgement([]byte("success")),
				).Return(nil)
			},
		},
		{
			name:          "Error from ics4Wrapper",
			packet:        channeltypes.Packet{},
			ack:           channeltypes.NewResultAcknowledgement([]byte("success")),
			expectedError: fmt.Errorf("write acknowledgement error"),
			setupMocks: func() {
				ics4WrapperMock.EXPECT().WriteAcknowledgement(
					gomock.Any(),
					channeltypes.Packet{},
					channeltypes.NewResultAcknowledgement([]byte("success")),
				).Return(fmt.Errorf("write acknowledgement error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			err := ics4Wrapper.WriteAcknowledgement(
				sdk.Context{},
				tt.packet,
				tt.ack,
			)
			if tt.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestWriteAcknowledgement_InvalidAck(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	keeperMock := mocks.NewMockTokenwrapperKeeper(ctrl)
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
		keeperMock,
	)

	// Test cases
	tests := []struct {
		name          string
		ack           ibcexported.Acknowledgement
		expectedError error
		setupMocks    func()
	}{
		{
			name:          "Nil acknowledgement",
			ack:           nil,
			expectedError: fmt.Errorf("nil acknowledgement"),
			setupMocks: func() {
				ics4WrapperMock.EXPECT().WriteAcknowledgement(
					gomock.Any(),
					channeltypes.Packet{},
					nil,
				).Return(fmt.Errorf("nil acknowledgement")).Times(1)
			},
		},
		{
			name:          "Empty acknowledgement",
			ack:           channeltypes.NewResultAcknowledgement([]byte{}),
			expectedError: fmt.Errorf("empty acknowledgement"),
			setupMocks: func() {
				ics4WrapperMock.EXPECT().WriteAcknowledgement(
					gomock.Any(),
					channeltypes.Packet{},
					channeltypes.NewResultAcknowledgement([]byte{}),
				).Return(fmt.Errorf("empty acknowledgement")).Times(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			err := ics4Wrapper.WriteAcknowledgement(
				sdk.Context{},
				channeltypes.Packet{},
				tt.ack,
			)
			if tt.expectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError.Error())
			}
		})
	}
}
