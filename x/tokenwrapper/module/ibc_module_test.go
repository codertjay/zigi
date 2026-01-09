package tokenwrapper_test

import (
	"fmt"
	"testing"

	tokenwrapper "zigchain/x/tokenwrapper/module"
	mocks "zigchain/x/tokenwrapper/testutil"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestOnChanOpenInit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	// Test cases
	tests := []struct {
		name            string
		order           channeltypes.Order
		connectionHops  []string
		portID          string
		channelID       string
		counterparty    channeltypes.Counterparty
		version         string
		expectedVersion string
		expectedError   error
		setupMocks      func()
	}{
		{
			name:            "Happy path",
			order:           channeltypes.ORDERED,
			connectionHops:  []string{"connection-1"},
			portID:          "transfer",
			channelID:       "channel-1",
			counterparty:    channeltypes.Counterparty{PortId: "transfer", ChannelId: "channel-2"},
			version:         "ics20-1",
			expectedVersion: "ics20-1",
			expectedError:   nil,
			setupMocks: func() {
				appMock.EXPECT().OnChanOpenInit(
					gomock.Any(),
					channeltypes.ORDERED,
					[]string{"connection-1"},
					"transfer",
					"channel-1",
					channeltypes.Counterparty{PortId: "transfer", ChannelId: "channel-2"},
					"ics20-1",
				).Return("ics20-1", nil)
			},
		},
		{
			name:            "Error from app callback",
			order:           channeltypes.ORDERED,
			connectionHops:  []string{"connection-1"},
			portID:          "transfer",
			channelID:       "channel-1",
			counterparty:    channeltypes.Counterparty{PortId: "transfer", ChannelId: "channel-2"},
			version:         "ics20-1",
			expectedVersion: "",
			expectedError:   fmt.Errorf("app callback error"),
			setupMocks: func() {
				appMock.EXPECT().OnChanOpenInit(
					gomock.Any(),
					channeltypes.ORDERED,
					[]string{"connection-1"},
					"transfer",
					"channel-1",
					channeltypes.Counterparty{PortId: "transfer", ChannelId: "channel-2"},
					"ics20-1",
				).Return("", fmt.Errorf("app callback error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			version, err := ibcModule.OnChanOpenInit(
				sdk.Context{},
				tt.order,
				tt.connectionHops,
				tt.portID,
				tt.channelID,
				tt.counterparty,
				tt.version,
			)
			if tt.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedVersion, version)
			}
		})
	}
}

func TestOnChanOpenTry(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	// Test cases
	tests := []struct {
		name                string
		order               channeltypes.Order
		connectionHops      []string
		portID              string
		channelID           string
		counterparty        channeltypes.Counterparty
		counterpartyVersion string
		expectedVersion     string
		expectedError       error
		setupMocks          func()
	}{
		{
			name:                "Happy path",
			order:               channeltypes.ORDERED,
			connectionHops:      []string{"connection-1"},
			portID:              "transfer",
			channelID:           "channel-1",
			counterparty:        channeltypes.Counterparty{PortId: "transfer", ChannelId: "channel-2"},
			counterpartyVersion: "ics20-1",
			expectedVersion:     "ics20-1",
			expectedError:       nil,
			setupMocks: func() {
				appMock.EXPECT().OnChanOpenTry(
					gomock.Any(),
					channeltypes.ORDERED,
					[]string{"connection-1"},
					"transfer",
					"channel-1",
					channeltypes.Counterparty{PortId: "transfer", ChannelId: "channel-2"},
					"ics20-1",
				).Return("ics20-1", nil)
			},
		},
		{
			name:                "Error from app callback",
			order:               channeltypes.ORDERED,
			connectionHops:      []string{"connection-1"},
			portID:              "transfer",
			channelID:           "channel-1",
			counterparty:        channeltypes.Counterparty{PortId: "transfer", ChannelId: "channel-2"},
			counterpartyVersion: "ics20-1",
			expectedVersion:     "",
			expectedError:       fmt.Errorf("app callback error"),
			setupMocks: func() {
				appMock.EXPECT().OnChanOpenTry(
					gomock.Any(),
					channeltypes.ORDERED,
					[]string{"connection-1"},
					"transfer",
					"channel-1",
					channeltypes.Counterparty{PortId: "transfer", ChannelId: "channel-2"},
					"ics20-1",
				).Return("", fmt.Errorf("app callback error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			version, err := ibcModule.OnChanOpenTry(
				sdk.Context{},
				tt.order,
				tt.connectionHops,
				tt.portID,
				tt.channelID,
				tt.counterparty,
				tt.counterpartyVersion,
			)
			if tt.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedVersion, version)
			}
		})
	}
}

func TestOnChanOpenAck(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	// Test cases
	tests := []struct {
		name                  string
		portID                string
		channelID             string
		counterpartyChannelId string
		counterpartyVersion   string
		expectedError         error
		setupMocks            func()
	}{
		{
			name:                  "Happy path",
			portID:                "transfer",
			channelID:             "channel-1",
			counterpartyChannelId: "channel-2",
			counterpartyVersion:   "ics20-1",
			expectedError:         nil,
			setupMocks: func() {
				appMock.EXPECT().OnChanOpenAck(
					gomock.Any(),
					"transfer",
					"channel-1",
					"channel-2",
					"ics20-1",
				).Return(nil)
			},
		},
		{
			name:                  "Error from app callback",
			portID:                "transfer",
			channelID:             "channel-1",
			counterpartyChannelId: "channel-2",
			counterpartyVersion:   "ics20-1",
			expectedError:         fmt.Errorf("app callback error"),
			setupMocks: func() {
				appMock.EXPECT().OnChanOpenAck(
					gomock.Any(),
					"transfer",
					"channel-1",
					"channel-2",
					"ics20-1",
				).Return(fmt.Errorf("app callback error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			err := ibcModule.OnChanOpenAck(
				sdk.Context{},
				tt.portID,
				tt.channelID,
				tt.counterpartyChannelId,
				tt.counterpartyVersion,
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

func TestOnChanOpenConfirm(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	// Test cases
	tests := []struct {
		name          string
		portID        string
		channelID     string
		expectedError error
		setupMocks    func()
	}{
		{
			name:          "Happy path",
			portID:        "transfer",
			channelID:     "channel-1",
			expectedError: nil,
			setupMocks: func() {
				appMock.EXPECT().OnChanOpenConfirm(
					gomock.Any(),
					"transfer",
					"channel-1",
				).Return(nil)
			},
		},
		{
			name:          "Error from app callback",
			portID:        "transfer",
			channelID:     "channel-1",
			expectedError: fmt.Errorf("app callback error"),
			setupMocks: func() {
				appMock.EXPECT().OnChanOpenConfirm(
					gomock.Any(),
					"transfer",
					"channel-1",
				).Return(fmt.Errorf("app callback error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			err := ibcModule.OnChanOpenConfirm(
				sdk.Context{},
				tt.portID,
				tt.channelID,
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

func TestOnChanCloseInit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	// Test cases
	tests := []struct {
		name          string
		portID        string
		channelID     string
		expectedError error
		setupMocks    func()
	}{
		{
			name:          "Happy path",
			portID:        "transfer",
			channelID:     "channel-1",
			expectedError: nil,
			setupMocks: func() {
				appMock.EXPECT().OnChanCloseInit(
					gomock.Any(),
					"transfer",
					"channel-1",
				).Return(nil)
			},
		},
		{
			name:          "Error from app callback",
			portID:        "transfer",
			channelID:     "channel-1",
			expectedError: fmt.Errorf("app callback error"),
			setupMocks: func() {
				appMock.EXPECT().OnChanCloseInit(
					gomock.Any(),
					"transfer",
					"channel-1",
				).Return(fmt.Errorf("app callback error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			err := ibcModule.OnChanCloseInit(
				sdk.Context{},
				tt.portID,
				tt.channelID,
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

func TestOnChanCloseConfirm(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	// Test cases
	tests := []struct {
		name          string
		portID        string
		channelID     string
		expectedError error
		setupMocks    func()
	}{
		{
			name:          "Happy path",
			portID:        "transfer",
			channelID:     "channel-1",
			expectedError: nil,
			setupMocks: func() {
				appMock.EXPECT().OnChanCloseConfirm(
					gomock.Any(),
					"transfer",
					"channel-1",
				).Return(nil)
			},
		},
		{
			name:          "Error from app callback",
			portID:        "transfer",
			channelID:     "channel-1",
			expectedError: fmt.Errorf("app callback error"),
			setupMocks: func() {
				appMock.EXPECT().OnChanCloseConfirm(
					gomock.Any(),
					"transfer",
					"channel-1",
				).Return(fmt.Errorf("app callback error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			err := ibcModule.OnChanCloseConfirm(
				sdk.Context{},
				tt.portID,
				tt.channelID,
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

func TestGetDenomPrefixAndNewHopAndIBCDenom(t *testing.T) {
	tests := []struct {
		name           string
		nativePort     string
		nativeChannel  string
		denom          string
		expectedPrefix string
		expectedIBC    string
	}{
		{
			name:           "Valid IBC denom with transfer port",
			nativePort:     "transfer",
			nativeChannel:  "channel-0",
			denom:          "uzig",
			expectedPrefix: "transfer/channel-0",
			expectedIBC:    "ibc/B6695315AE86FC41615A2746E62851A22F430D929F2512E773470B6E7187F8C5",
		},
		{
			name:           "Valid IBC denom with custom port",
			nativePort:     "custom-port",
			nativeChannel:  "channel-1",
			denom:          "uzig",
			expectedPrefix: "custom-port/channel-1",
			expectedIBC:    "ibc/738811BF0CB3E487D3CD14EE2C59CCDA163299C63F56E855933CD0F196F5EE54",
		},
		{
			name:           "Empty denom",
			nativePort:     "transfer",
			nativeChannel:  "channel-0",
			denom:          "",
			expectedPrefix: "transfer/channel-0",
			expectedIBC:    "ibc/AC64FD65731C63C968A7BB1E711DB157BE85F7538F128FDA3363EDDD170FABF1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1) With GetDenomPrefix

			// Get the denom prefix
			//lint:ignore SA1019 -- this is to compare the deprecated function with the new one
			prefix := transfertypes.GetDenomPrefix(tt.nativePort, tt.nativeChannel)
			require.Equal(t, fmt.Sprintf("%s/", tt.expectedPrefix), prefix)

			// Create the prefixed denom
			prefixedDenom := prefix + tt.denom

			// Parse the IBC denom
			//lint:ignore SA1019 -- this is to compare the deprecated function with the new one
			ibcDenom := transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom()
			require.Equal(t, tt.expectedIBC, ibcDenom)

			// 2) With NewHop

			// Get the denom prefix
			prefix2 := transfertypes.NewHop(tt.nativePort, tt.nativeChannel)
			require.Equal(t, tt.expectedPrefix, prefix2.String())

			// Create the prefixed denom
			prefixedDenom = fmt.Sprintf("%s/%s", prefix2.String(), tt.denom)

			// Parse the IBC denom
			ibcDenom = transfertypes.ExtractDenomFromPath(prefixedDenom).IBCDenom()
			require.Equal(t, tt.expectedIBC, ibcDenom)
		})
	}
}

func TestSenderChainIsSourceAndHasPrefix(t *testing.T) {
	tests := []struct {
		name          string
		nativePort    string
		nativeChannel string
		denom         string
		expected      bool
	}{
		{
			name:          "Not source chain - simple denom",
			nativePort:    "transfer",
			nativeChannel: "channel-0",
			denom:         "transfer/channel-0/uzig",
			expected:      false,
		},
		{
			name:          "Not source chain - custom port",
			nativePort:    "custom-port",
			nativeChannel: "channel-1",
			denom:         "custom-port/channel-1/uzig",
			expected:      false,
		},
		{
			name:          "Not source chain - malformed IBC denom",
			nativePort:    "transfer",
			nativeChannel: "channel-0",
			denom:         "transfer/channel-1/uzig",
			expected:      true,
		},
		{
			name:          "Source chain - IBC prefixed denom",
			nativePort:    "transfer",
			nativeChannel: "channel-0",
			denom:         "ibc/B6695315AE86FC41615A2746E62851A22F430D929F2512E773470B6E7187F8C5",
			expected:      true,
		},
		{
			name:          "Source chain - different port",
			nativePort:    "customport-0",
			nativeChannel: "channel-0",
			denom:         "ibc/B6695315AE86FC41615A2746E62851A22F430D929F2512E773470B6E7187F8C5",
			expected:      true,
		},
		{
			name:          "Source chain - different channel",
			nativePort:    "transfer",
			nativeChannel: "channel-1",
			denom:         "ibc/B6695315AE86FC41615A2746E62851A22F430D929F2512E773470B6E7187F8C5",
			expected:      true,
		},
		{
			name:          "Not source chain - empty denom",
			nativePort:    "transfer",
			nativeChannel: "channel-0",
			denom:         "",
			expected:      true,
		},
		{
			name:          "Source chain - malformed IBC denom",
			nativePort:    "transfer",
			nativeChannel: "channel-0",
			denom:         "ibc/invalid/format",
			expected:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1) With SenderChainIsSource
			//lint:ignore SA1019 -- this is to compare the deprecated function with the new one
			result := transfertypes.SenderChainIsSource(tt.nativePort, tt.nativeChannel, tt.denom)
			require.Equal(t, tt.expected, result)

			// 2) With ReceiverChainIsSource
			denom := transfertypes.ExtractDenomFromPath(tt.denom)
			result = !denom.HasPrefix(tt.nativePort, tt.nativeChannel)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractDenomFromPath(t *testing.T) {
	tests := []struct {
		name          string
		fullPath      string
		expectedDenom string
	}{
		{
			name:          "Extract base denom from full path",
			fullPath:      "transfer/channel-0/waxlzig",
			expectedDenom: "waxlzig",
		},
		{
			name:          "Extract base denom from simple path",
			fullPath:      "uzig",
			expectedDenom: "uzig",
		},
		{
			name:          "Extract base denom from empty path",
			fullPath:      "",
			expectedDenom: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			denom := transfertypes.ExtractDenomFromPath(tt.fullPath)
			require.Equal(t, tt.expectedDenom, denom.GetBase())
		})
	}
}

func TestOnChanOpenInit_InvalidPortID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	// Test cases
	tests := []struct {
		name            string
		portID          string
		expectedVersion string
		expectedError   error
		setupMocks      func()
	}{
		{
			name:            "Empty port ID",
			portID:          "",
			expectedVersion: "",
			expectedError:   fmt.Errorf("invalid port ID"),
			setupMocks: func() {
				appMock.EXPECT().OnChanOpenInit(
					gomock.Any(),
					channeltypes.ORDERED,
					[]string{"connection-1"},
					"",
					"channel-1",
					channeltypes.Counterparty{PortId: "transfer", ChannelId: "channel-2"},
					"ics20-1",
				).Return("", fmt.Errorf("invalid port ID")).Times(1)
			},
		},
		{
			name:            "Invalid port ID format",
			portID:          "invalid/port",
			expectedVersion: "",
			expectedError:   fmt.Errorf("invalid port ID format"),
			setupMocks: func() {
				appMock.EXPECT().OnChanOpenInit(
					gomock.Any(),
					channeltypes.ORDERED,
					[]string{"connection-1"},
					"invalid/port",
					"channel-1",
					channeltypes.Counterparty{PortId: "transfer", ChannelId: "channel-2"},
					"ics20-1",
				).Return("", fmt.Errorf("invalid port ID format")).Times(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			version, err := ibcModule.OnChanOpenInit(
				sdk.Context{},
				channeltypes.ORDERED,
				[]string{"connection-1"},
				tt.portID,
				"channel-1",
				channeltypes.Counterparty{PortId: "transfer", ChannelId: "channel-2"},
				"ics20-1",
			)
			if tt.expectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError.Error())
				require.Equal(t, tt.expectedVersion, version)
			}
		})
	}
}

func TestOnChanOpenTry_InvalidChannelID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	// Test cases
	tests := []struct {
		name            string
		channelID       string
		expectedVersion string
		expectedError   error
		setupMocks      func()
	}{
		{
			name:            "Empty channel ID",
			channelID:       "",
			expectedVersion: "",
			expectedError:   fmt.Errorf("invalid channel ID"),
			setupMocks: func() {
				appMock.EXPECT().OnChanOpenTry(
					gomock.Any(),
					channeltypes.ORDERED,
					[]string{"connection-1"},
					"transfer",
					"",
					channeltypes.Counterparty{PortId: "transfer", ChannelId: "channel-2"},
					"ics20-1",
				).Return("", fmt.Errorf("invalid channel ID")).Times(1)
			},
		},
		{
			name:            "Invalid channel ID format",
			channelID:       "invalid/channel",
			expectedVersion: "",
			expectedError:   fmt.Errorf("invalid channel ID format"),
			setupMocks: func() {
				appMock.EXPECT().OnChanOpenTry(
					gomock.Any(),
					channeltypes.ORDERED,
					[]string{"connection-1"},
					"transfer",
					"invalid/channel",
					channeltypes.Counterparty{PortId: "transfer", ChannelId: "channel-2"},
					"ics20-1",
				).Return("", fmt.Errorf("invalid channel ID format")).Times(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			version, err := ibcModule.OnChanOpenTry(
				sdk.Context{},
				channeltypes.ORDERED,
				[]string{"connection-1"},
				"transfer",
				tt.channelID,
				channeltypes.Counterparty{PortId: "transfer", ChannelId: "channel-2"},
				"ics20-1",
			)
			if tt.expectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError.Error())
				require.Equal(t, tt.expectedVersion, version)
			}
		})
	}
}

func TestOnChanOpenAck_InvalidCounterpartyChannelID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	// Test cases
	tests := []struct {
		name                  string
		counterpartyChannelId string
		expectedError         error
		setupMocks            func()
	}{
		{
			name:                  "Empty counterparty channel ID",
			counterpartyChannelId: "",
			expectedError:         fmt.Errorf("invalid counterparty channel ID"),
			setupMocks: func() {
				appMock.EXPECT().OnChanOpenAck(
					gomock.Any(),
					"transfer",
					"channel-1",
					"",
					"ics20-1",
				).Return(fmt.Errorf("invalid counterparty channel ID")).Times(1)
			},
		},
		{
			name:                  "Invalid counterparty channel ID format",
			counterpartyChannelId: "invalid/channel",
			expectedError:         fmt.Errorf("invalid counterparty channel ID format"),
			setupMocks: func() {
				appMock.EXPECT().OnChanOpenAck(
					gomock.Any(),
					"transfer",
					"channel-1",
					"invalid/channel",
					"ics20-1",
				).Return(fmt.Errorf("invalid counterparty channel ID format")).Times(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			err := ibcModule.OnChanOpenAck(
				sdk.Context{},
				"transfer",
				"channel-1",
				tt.counterpartyChannelId,
				"ics20-1",
			)
			if tt.expectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError.Error())
			}
		})
	}
}

func TestOnChanOpenConfirm_InvalidPortID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	// Test cases
	tests := []struct {
		name          string
		portID        string
		expectedError error
		setupMocks    func()
	}{
		{
			name:          "Empty port ID",
			portID:        "",
			expectedError: fmt.Errorf("invalid port ID"),
			setupMocks: func() {
				appMock.EXPECT().OnChanOpenConfirm(
					gomock.Any(),
					"",
					"channel-1",
				).Return(fmt.Errorf("invalid port ID")).Times(1)
			},
		},
		{
			name:          "Invalid port ID format",
			portID:        "invalid/port",
			expectedError: fmt.Errorf("invalid port ID format"),
			setupMocks: func() {
				appMock.EXPECT().OnChanOpenConfirm(
					gomock.Any(),
					"invalid/port",
					"channel-1",
				).Return(fmt.Errorf("invalid port ID format")).Times(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			err := ibcModule.OnChanOpenConfirm(
				sdk.Context{},
				tt.portID,
				"channel-1",
			)
			if tt.expectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError.Error())
			}
		})
	}
}

func TestOnChanCloseInit_InvalidChannelID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	// Test cases
	tests := []struct {
		name          string
		channelID     string
		expectedError error
		setupMocks    func()
	}{
		{
			name:          "Empty channel ID",
			channelID:     "",
			expectedError: fmt.Errorf("invalid channel ID"),
			setupMocks: func() {
				appMock.EXPECT().OnChanCloseInit(
					gomock.Any(),
					"transfer",
					"",
				).Return(fmt.Errorf("invalid channel ID")).Times(1)
			},
		},
		{
			name:          "Invalid channel ID format",
			channelID:     "invalid/channel",
			expectedError: fmt.Errorf("invalid channel ID format"),
			setupMocks: func() {
				appMock.EXPECT().OnChanCloseInit(
					gomock.Any(),
					"transfer",
					"invalid/channel",
				).Return(fmt.Errorf("invalid channel ID format")).Times(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			err := ibcModule.OnChanCloseInit(
				sdk.Context{},
				"transfer",
				tt.channelID,
			)
			if tt.expectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError.Error())
			}
		})
	}
}

func TestOnChanCloseConfirm_InvalidPortID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	// Test cases
	tests := []struct {
		name          string
		portID        string
		expectedError error
		setupMocks    func()
	}{
		{
			name:          "Empty port ID",
			portID:        "",
			expectedError: fmt.Errorf("invalid port ID"),
			setupMocks: func() {
				appMock.EXPECT().OnChanCloseConfirm(
					gomock.Any(),
					"",
					"channel-1",
				).Return(fmt.Errorf("invalid port ID")).Times(1)
			},
		},
		{
			name:          "Invalid port ID format",
			portID:        "invalid/port",
			expectedError: fmt.Errorf("invalid port ID format"),
			setupMocks: func() {
				appMock.EXPECT().OnChanCloseConfirm(
					gomock.Any(),
					"invalid/port",
					"channel-1",
				).Return(fmt.Errorf("invalid port ID format")).Times(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			err := ibcModule.OnChanCloseConfirm(
				sdk.Context{},
				tt.portID,
				"channel-1",
			)
			if tt.expectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError.Error())
			}
		})
	}
}

func TestOnChanOpenInit_InvalidConnectionHops(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	// Test cases
	tests := []struct {
		name            string
		connectionHops  []string
		expectedVersion string
		expectedError   error
		setupMocks      func()
	}{
		{
			name:            "Nil connection hops",
			connectionHops:  nil,
			expectedVersion: "",
			expectedError:   fmt.Errorf("nil connection hops"),
			setupMocks: func() {
				appMock.EXPECT().OnChanOpenInit(
					gomock.Any(),
					channeltypes.ORDERED,
					nil,
					"transfer",
					"channel-1",
					channeltypes.Counterparty{PortId: "transfer", ChannelId: "channel-2"},
					"ics20-1",
				).Return("", fmt.Errorf("nil connection hops")).Times(1)
			},
		},
		{
			name:            "Empty connection hops",
			connectionHops:  []string{},
			expectedVersion: "",
			expectedError:   fmt.Errorf("empty connection hops"),
			setupMocks: func() {
				appMock.EXPECT().OnChanOpenInit(
					gomock.Any(),
					channeltypes.ORDERED,
					[]string{},
					"transfer",
					"channel-1",
					channeltypes.Counterparty{PortId: "transfer", ChannelId: "channel-2"},
					"ics20-1",
				).Return("", fmt.Errorf("empty connection hops")).Times(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			version, err := ibcModule.OnChanOpenInit(
				sdk.Context{},
				channeltypes.ORDERED,
				tt.connectionHops,
				"transfer",
				"channel-1",
				channeltypes.Counterparty{PortId: "transfer", ChannelId: "channel-2"},
				"ics20-1",
			)
			if tt.expectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError.Error())
				require.Equal(t, tt.expectedVersion, version)
			}
		})
	}
}

func TestOnChanOpenTry_InvalidCounterparty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	// Test cases
	tests := []struct {
		name            string
		counterparty    channeltypes.Counterparty
		expectedVersion string
		expectedError   error
		setupMocks      func()
	}{
		{
			name:            "Empty counterparty port ID",
			counterparty:    channeltypes.Counterparty{PortId: "", ChannelId: "channel-2"},
			expectedVersion: "",
			expectedError:   fmt.Errorf("invalid counterparty port ID"),
			setupMocks: func() {
				appMock.EXPECT().OnChanOpenTry(
					gomock.Any(),
					channeltypes.ORDERED,
					[]string{"connection-1"},
					"transfer",
					"channel-1",
					channeltypes.Counterparty{PortId: "", ChannelId: "channel-2"},
					"ics20-1",
				).Return("", fmt.Errorf("invalid counterparty port ID")).Times(1)
			},
		},
		{
			name:            "Empty counterparty channel ID",
			counterparty:    channeltypes.Counterparty{PortId: "transfer", ChannelId: ""},
			expectedVersion: "",
			expectedError:   fmt.Errorf("invalid counterparty channel ID"),
			setupMocks: func() {
				appMock.EXPECT().OnChanOpenTry(
					gomock.Any(),
					channeltypes.ORDERED,
					[]string{"connection-1"},
					"transfer",
					"channel-1",
					channeltypes.Counterparty{PortId: "transfer", ChannelId: ""},
					"ics20-1",
				).Return("", fmt.Errorf("invalid counterparty channel ID")).Times(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			version, err := ibcModule.OnChanOpenTry(
				sdk.Context{},
				channeltypes.ORDERED,
				[]string{"connection-1"},
				"transfer",
				"channel-1",
				tt.counterparty,
				"ics20-1",
			)
			if tt.expectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError.Error())
				require.Equal(t, tt.expectedVersion, version)
			}
		})
	}
}

func TestOnChanOpenInit_InvalidVersion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	// Test cases
	tests := []struct {
		name            string
		version         string
		expectedVersion string
		expectedError   error
		setupMocks      func()
	}{
		{
			name:            "Empty version",
			version:         "",
			expectedVersion: "",
			expectedError:   fmt.Errorf("invalid version"),
			setupMocks: func() {
				appMock.EXPECT().OnChanOpenInit(
					gomock.Any(),
					channeltypes.ORDERED,
					[]string{"connection-1"},
					"transfer",
					"channel-1",
					channeltypes.Counterparty{PortId: "transfer", ChannelId: "channel-2"},
					"",
				).Return("", fmt.Errorf("invalid version")).Times(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			version, err := ibcModule.OnChanOpenInit(
				sdk.Context{},
				channeltypes.ORDERED,
				[]string{"connection-1"},
				"transfer",
				"channel-1",
				channeltypes.Counterparty{PortId: "transfer", ChannelId: "channel-2"},
				tt.version,
			)
			if tt.expectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError.Error())
				require.Equal(t, tt.expectedVersion, version)
			}
		})
	}
}
