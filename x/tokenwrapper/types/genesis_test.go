package types_test

import (
	"testing"

	"zigchain/testutil/sample"
	"zigchain/x/tokenwrapper/types"

	"cosmossdk.io/math"

	"github.com/stretchr/testify/require"
)

// Positive Tests

func TestGenesisState_Validate_DefaultIsValid(t *testing.T) {
	// Test case: valid default genesis state
	genState := types.DefaultGenesis()
	err := genState.Validate()
	require.NoError(t, err)
}

func TestGenesisState_Validate_ValidGenesisState(t *testing.T) {
	// Test case: valid genesis state with all fields set
	genState := &types.GenesisState{
		Params:                  types.NewParams(),
		TotalTransferredIn:      math.ZeroInt(),
		TotalTransferredOut:     math.ZeroInt(),
		OperatorAddress:         sample.AccAddress(),
		ProposedOperatorAddress: sample.AccAddress(),
		PauserAddresses:         []string{},
		Enabled:                 true,
		NativeClientId:          "07-tendermint-0",
		CounterpartyClientId:    "07-tendermint-1",
		NativePort:              "transfer",
		CounterpartyPort:        "transfer",
		NativeChannel:           "channel-0",
		CounterpartyChannel:     "channel-1",
		Denom:                   "uzig",
		DecimalDifference:       0,
	}
	err := genState.Validate()
	require.NoError(t, err)
}

func TestGenesisState_Validate_ValidGenesisStateWithPauserAddresses(t *testing.T) {
	// Test case: valid genesis state with pauser addresses
	genState := &types.GenesisState{
		Params:                  types.NewParams(),
		TotalTransferredIn:      math.ZeroInt(),
		TotalTransferredOut:     math.ZeroInt(),
		OperatorAddress:         sample.AccAddress(),
		ProposedOperatorAddress: sample.AccAddress(),
		PauserAddresses:         []string{sample.AccAddress()},
		Enabled:                 true,
		NativeClientId:          "07-tendermint-0",
		CounterpartyClientId:    "07-tendermint-1",
		NativePort:              "transfer",
		CounterpartyPort:        "transfer",
		NativeChannel:           "channel-0",
		CounterpartyChannel:     "channel-1",
		Denom:                   "uzig",
		DecimalDifference:       0,
	}
	err := genState.Validate()
	require.NoError(t, err)
}

func TestGenesisState_Validate_ValidNilPauserAddresses(t *testing.T) {
	// Test case: valid genesis state with nil pauser addresses
	genState := &types.GenesisState{
		Params:                  types.NewParams(),
		TotalTransferredIn:      math.ZeroInt(),
		TotalTransferredOut:     math.ZeroInt(),
		OperatorAddress:         sample.AccAddress(),
		ProposedOperatorAddress: sample.AccAddress(),
		PauserAddresses:         nil,
		Enabled:                 true,
		NativeClientId:          "07-tendermint-0",
		CounterpartyClientId:    "07-tendermint-1",
		NativePort:              "transfer",
		CounterpartyPort:        "transfer",
		NativeChannel:           "channel-0",
		CounterpartyChannel:     "channel-1",
		Denom:                   "uzig",
		DecimalDifference:       0,
	}
	err := genState.Validate()
	require.NoError(t, err)
}

func TestGenesisState_Validate_ValidMaxDecimalDifference(t *testing.T) {
	// Test case: valid genesis state with maximum decimal difference
	genState := &types.GenesisState{
		Params:                  types.NewParams(),
		TotalTransferredIn:      math.ZeroInt(),
		TotalTransferredOut:     math.ZeroInt(),
		OperatorAddress:         sample.AccAddress(),
		ProposedOperatorAddress: sample.AccAddress(),
		PauserAddresses:         []string{},
		Enabled:                 true,
		NativeClientId:          "07-tendermint-0",
		CounterpartyClientId:    "07-tendermint-1",
		NativePort:              "transfer",
		CounterpartyPort:        "transfer",
		NativeChannel:           "channel-0",
		CounterpartyChannel:     "channel-1",
		Denom:                   "uzig",
		DecimalDifference:       18,
	}
	err := genState.Validate()
	require.NoError(t, err)
}

// Negative Tests

func TestGenesisState_Validate_InvalidNegativeTotalTransferredIn(t *testing.T) {
	// Test case: invalid genesis state with negative TotalTransferredIn
	genState := &types.GenesisState{
		Params:                  types.NewParams(),
		TotalTransferredIn:      math.NewInt(-1),
		TotalTransferredOut:     math.ZeroInt(),
		OperatorAddress:         sample.AccAddress(),
		ProposedOperatorAddress: sample.AccAddress(),
		PauserAddresses:         []string{},
		Enabled:                 true,
		NativeClientId:          "07-tendermint-0",
		CounterpartyClientId:    "07-tendermint-1",
		NativePort:              "transfer",
		CounterpartyPort:        "transfer",
		NativeChannel:           "channel-0",
		CounterpartyChannel:     "channel-1",
		Denom:                   "uzig",
		DecimalDifference:       0,
	}
	err := genState.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "total transferred in cannot be negative")
}

func TestGenesisState_Validate_InvalidNegativeTotalTransferredOut(t *testing.T) {
	// Test case: invalid genesis state with negative TotalTransferredOut
	genState := &types.GenesisState{
		Params:                  types.NewParams(),
		TotalTransferredIn:      math.ZeroInt(),
		TotalTransferredOut:     math.NewInt(-1),
		OperatorAddress:         sample.AccAddress(),
		ProposedOperatorAddress: sample.AccAddress(),
		PauserAddresses:         []string{},
		Enabled:                 true,
		NativeClientId:          "07-tendermint-0",
		CounterpartyClientId:    "07-tendermint-1",
		NativePort:              "transfer",
		CounterpartyPort:        "transfer",
		NativeChannel:           "channel-0",
		CounterpartyChannel:     "channel-1",
		Denom:                   "uzig",
		DecimalDifference:       0,
	}
	err := genState.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "total transferred out cannot be negative")
}

func TestGenesisState_Validate_InvalidOperatorAddress(t *testing.T) {
	// Test case: invalid genesis state with invalid OperatorAddress
	genState := &types.GenesisState{
		Params:                  types.NewParams(),
		TotalTransferredIn:      math.ZeroInt(),
		TotalTransferredOut:     math.ZeroInt(),
		OperatorAddress:         "invalid-address",
		ProposedOperatorAddress: sample.AccAddress(),
		PauserAddresses:         []string{},
		Enabled:                 true,
		NativeClientId:          "07-tendermint-0",
		CounterpartyClientId:    "07-tendermint-1",
		NativePort:              "transfer",
		CounterpartyPort:        "transfer",
		NativeChannel:           "channel-0",
		CounterpartyChannel:     "channel-1",
		Denom:                   "uzig",
		DecimalDifference:       0,
	}
	err := genState.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "operator address: 'invalid-address' (decoding bech32 failed: invalid separator index -1)")
}

func TestGenesisState_Validate_InvalidEmptyOperatorAddress(t *testing.T) {
	// Test case: invalid genesis state with empty OperatorAddress
	genState := &types.GenesisState{
		Params:                  types.NewParams(),
		TotalTransferredIn:      math.ZeroInt(),
		TotalTransferredOut:     math.ZeroInt(),
		OperatorAddress:         "",
		ProposedOperatorAddress: sample.AccAddress(),
		PauserAddresses:         []string{},
		Enabled:                 true,
		NativeClientId:          "07-tendermint-0",
		CounterpartyClientId:    "07-tendermint-1",
		NativePort:              "transfer",
		CounterpartyPort:        "transfer",
		NativeChannel:           "channel-0",
		CounterpartyChannel:     "channel-1",
		Denom:                   "uzig",
		DecimalDifference:       0,
	}
	err := genState.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "operator address: cannot be empty: invalid address")
}

func TestGenesisState_Validate_InvalidProposedOperatorAddress(t *testing.T) {
	// Test case: invalid genesis state with invalid ProposedOperatorAddress
	genState := &types.GenesisState{
		Params:                  types.NewParams(),
		TotalTransferredIn:      math.ZeroInt(),
		TotalTransferredOut:     math.ZeroInt(),
		OperatorAddress:         sample.AccAddress(),
		ProposedOperatorAddress: "invalid-address",
		PauserAddresses:         []string{},
		Enabled:                 true,
		NativeClientId:          "07-tendermint-0",
		CounterpartyClientId:    "07-tendermint-1",
		NativePort:              "transfer",
		CounterpartyPort:        "transfer",
		NativeChannel:           "channel-0",
		CounterpartyChannel:     "channel-1",
		Denom:                   "uzig",
		DecimalDifference:       0,
	}
	err := genState.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "proposed_operator address: 'invalid-address' (decoding bech32 failed: invalid separator index -1)")
}

func TestGenesisState_Validate_InvalidEmptyProposedOperatorAddress(t *testing.T) {
	// Test case: invalid genesis state with empty ProposedOperatorAddress
	genState := &types.GenesisState{
		Params:                  types.NewParams(),
		TotalTransferredIn:      math.ZeroInt(),
		TotalTransferredOut:     math.ZeroInt(),
		OperatorAddress:         sample.AccAddress(),
		ProposedOperatorAddress: "",
		PauserAddresses:         []string{},
		Enabled:                 true,
		NativeClientId:          "07-tendermint-0",
		CounterpartyClientId:    "07-tendermint-1",
		NativePort:              "transfer",
		CounterpartyPort:        "transfer",
		NativeChannel:           "channel-0",
		CounterpartyChannel:     "channel-1",
		Denom:                   "uzig",
		DecimalDifference:       0,
	}
	err := genState.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "proposed_operator address: cannot be empty: invalid address")
}

func TestGenesisState_Validate_InvalidPauserAddress(t *testing.T) {
	// Test case: invalid genesis state with invalid PauserAddresses
	genState := &types.GenesisState{
		Params:                  types.NewParams(),
		TotalTransferredIn:      math.ZeroInt(),
		TotalTransferredOut:     math.ZeroInt(),
		OperatorAddress:         sample.AccAddress(),
		ProposedOperatorAddress: sample.AccAddress(),
		PauserAddresses:         []string{"invalid-address"},
		Enabled:                 true,
		NativeClientId:          "07-tendermint-0",
		CounterpartyClientId:    "07-tendermint-1",
		NativePort:              "transfer",
		CounterpartyPort:        "transfer",
		NativeChannel:           "channel-0",
		CounterpartyChannel:     "channel-1",
		Denom:                   "uzig",
		DecimalDifference:       0,
	}
	err := genState.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "pauser[0] address: 'invalid-address' (decoding bech32 failed: invalid separator index -1)")
}

func TestGenesisState_Validate_InvalidPauserAddressesWithEmptyString(t *testing.T) {
	// Test case: invalid genesis state with PauserAddresses containing an empty string
	genState := &types.GenesisState{
		Params:                  types.NewParams(),
		TotalTransferredIn:      math.ZeroInt(),
		TotalTransferredOut:     math.ZeroInt(),
		OperatorAddress:         sample.AccAddress(),
		ProposedOperatorAddress: sample.AccAddress(),
		PauserAddresses:         []string{""},
		Enabled:                 true,
		NativeClientId:          "07-tendermint-0",
		CounterpartyClientId:    "07-tendermint-1",
		NativePort:              "transfer",
		CounterpartyPort:        "transfer",
		NativeChannel:           "channel-0",
		CounterpartyChannel:     "channel-1",
		Denom:                   "uzig",
		DecimalDifference:       0,
	}
	err := genState.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "pauser[0] address: cannot be empty: invalid address")
}

func TestGenesisState_Validate_InvalidPauserAddressesWithMultipleInvalids(t *testing.T) {
	// Test case: invalid genesis state with PauserAddresses containing multiple invalid addresses
	genState := &types.GenesisState{
		Params:                  types.NewParams(),
		TotalTransferredIn:      math.ZeroInt(),
		TotalTransferredOut:     math.ZeroInt(),
		OperatorAddress:         sample.AccAddress(),
		ProposedOperatorAddress: sample.AccAddress(),
		PauserAddresses:         []string{"invalid1", "invalid2"},
		Enabled:                 true,
		NativeClientId:          "07-tendermint-0",
		CounterpartyClientId:    "07-tendermint-1",
		NativePort:              "transfer",
		CounterpartyPort:        "transfer",
		NativeChannel:           "channel-0",
		CounterpartyChannel:     "channel-1",
		Denom:                   "uzig",
		DecimalDifference:       0,
	}
	err := genState.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "pauser[0] address: 'invalid1' (decoding bech32 failed: invalid separator index 7)")
}

func TestGenesisState_Validate_InvalidNativeClientID(t *testing.T) {
	// Test case: invalid genesis state with invalid NativeClientId
	genState := &types.GenesisState{
		Params:                  types.NewParams(),
		TotalTransferredIn:      math.ZeroInt(),
		TotalTransferredOut:     math.ZeroInt(),
		OperatorAddress:         sample.AccAddress(),
		ProposedOperatorAddress: sample.AccAddress(),
		PauserAddresses:         []string{},
		Enabled:                 true,
		NativeClientId:          "invalid-client",
		CounterpartyClientId:    "07-tendermint-1",
		NativePort:              "transfer",
		CounterpartyPort:        "transfer",
		NativeChannel:           "channel-0",
		CounterpartyChannel:     "channel-1",
		Denom:                   "uzig",
		DecimalDifference:       0,
	}
	err := genState.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "light client is invalid: invalid client ID format")
}

func TestGenesisState_Validate_InvalidEmptyNativeClientId(t *testing.T) {
	// Test case: invalid genesis state with empty NativeClientId
	genState := &types.GenesisState{
		Params:                  types.NewParams(),
		TotalTransferredIn:      math.ZeroInt(),
		TotalTransferredOut:     math.ZeroInt(),
		OperatorAddress:         sample.AccAddress(),
		ProposedOperatorAddress: sample.AccAddress(),
		PauserAddresses:         []string{},
		Enabled:                 true,
		NativeClientId:          "",
		CounterpartyClientId:    "07-tendermint-1",
		NativePort:              "transfer",
		CounterpartyPort:        "transfer",
		NativeChannel:           "channel-0",
		CounterpartyChannel:     "channel-1",
		Denom:                   "uzig",
		DecimalDifference:       0,
	}
	err := genState.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "light client is invalid: client ID cannot be empty")
}

func TestGenesisState_Validate_InvalidNativePort(t *testing.T) {
	// Test case: invalid genesis state with invalid NativePort
	genState := &types.GenesisState{
		Params:                  types.NewParams(),
		TotalTransferredIn:      math.ZeroInt(),
		TotalTransferredOut:     math.ZeroInt(),
		OperatorAddress:         sample.AccAddress(),
		ProposedOperatorAddress: sample.AccAddress(),
		PauserAddresses:         []string{},
		Enabled:                 true,
		NativeClientId:          "07-tendermint-0",
		CounterpartyClientId:    "07-tendermint-1",
		NativePort:              "invalid/port",
		CounterpartyPort:        "transfer",
		NativeChannel:           "channel-0",
		CounterpartyChannel:     "channel-1",
		Denom:                   "uzig",
		DecimalDifference:       0,
	}
	err := genState.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid port: port contains invalid characters. Only alphanumeric, ., _, +, -, #, [, ], <, > are allowed")
}

func TestGenesisState_Validate_InvalidEmptyNativePort(t *testing.T) {
	// Test case: invalid genesis state with empty NativePort
	genState := &types.GenesisState{
		Params:                  types.NewParams(),
		TotalTransferredIn:      math.ZeroInt(),
		TotalTransferredOut:     math.ZeroInt(),
		OperatorAddress:         sample.AccAddress(),
		ProposedOperatorAddress: sample.AccAddress(),
		PauserAddresses:         []string{},
		Enabled:                 true,
		NativeClientId:          "07-tendermint-0",
		CounterpartyClientId:    "07-tendermint-1",
		NativePort:              "",
		CounterpartyPort:        "transfer",
		NativeChannel:           "channel-0",
		CounterpartyChannel:     "channel-1",
		Denom:                   "uzig",
		DecimalDifference:       0,
	}
	err := genState.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid port: port cannot be empty")
}

func TestGenesisState_Validate_InvalidNativeChannel(t *testing.T) {
	// Test case: invalid genesis state with invalid NativeChannel
	genState := &types.GenesisState{
		Params:                  types.NewParams(),
		TotalTransferredIn:      math.ZeroInt(),
		TotalTransferredOut:     math.ZeroInt(),
		OperatorAddress:         sample.AccAddress(),
		ProposedOperatorAddress: sample.AccAddress(),
		PauserAddresses:         []string{},
		Enabled:                 true,
		NativeClientId:          "07-tendermint-0",
		CounterpartyClientId:    "07-tendermint-1",
		NativePort:              "transfer",
		CounterpartyPort:        "transfer",
		NativeChannel:           "invalid-channel",
		CounterpartyChannel:     "channel-1",
		Denom:                   "uzig",
		DecimalDifference:       0,
	}
	err := genState.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid channel identifier: invalid channel ID format")
}

func TestGenesisState_Validate_InvalidEmptyNativeChannel(t *testing.T) {
	// Test case: invalid genesis state with empty NativeChannel
	genState := &types.GenesisState{
		Params:                  types.NewParams(),
		TotalTransferredIn:      math.ZeroInt(),
		TotalTransferredOut:     math.ZeroInt(),
		OperatorAddress:         sample.AccAddress(),
		ProposedOperatorAddress: sample.AccAddress(),
		PauserAddresses:         []string{},
		Enabled:                 true,
		NativeClientId:          "07-tendermint-0",
		CounterpartyClientId:    "07-tendermint-1",
		NativePort:              "transfer",
		CounterpartyPort:        "transfer",
		NativeChannel:           "",
		CounterpartyChannel:     "channel-1",
		Denom:                   "uzig",
		DecimalDifference:       0,
	}
	err := genState.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid channel identifier: channel cannot be empty")
}

func TestGenesisState_Validate_InvalidDenom(t *testing.T) {
	// Test case: invalid genesis state with invalid Denom
	genState := &types.GenesisState{
		Params:                  types.NewParams(),
		TotalTransferredIn:      math.ZeroInt(),
		TotalTransferredOut:     math.ZeroInt(),
		OperatorAddress:         sample.AccAddress(),
		ProposedOperatorAddress: sample.AccAddress(),
		PauserAddresses:         []string{},
		Enabled:                 true,
		NativeClientId:          "07-tendermint-0",
		CounterpartyClientId:    "07-tendermint-1",
		NativePort:              "transfer",
		CounterpartyPort:        "transfer",
		NativeChannel:           "channel-0",
		CounterpartyChannel:     "channel-1",
		Denom:                   "invalid/denom",
		DecimalDifference:       0,
	}
	err := genState.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "denom contains invalid characters. Only alphanumeric, ., _, +, -, #, [, ], <, > are allowed")
}

func TestGenesisState_Validate_InvalidEmptyDenom(t *testing.T) {
	// Test case: invalid genesis state with empty Denom
	genState := &types.GenesisState{
		Params:                  types.NewParams(),
		TotalTransferredIn:      math.ZeroInt(),
		TotalTransferredOut:     math.ZeroInt(),
		OperatorAddress:         sample.AccAddress(),
		ProposedOperatorAddress: sample.AccAddress(),
		PauserAddresses:         []string{},
		Enabled:                 true,
		NativeClientId:          "07-tendermint-0",
		CounterpartyClientId:    "07-tendermint-1",
		NativePort:              "transfer",
		CounterpartyPort:        "transfer",
		NativeChannel:           "channel-0",
		CounterpartyChannel:     "channel-1",
		Denom:                   "",
		DecimalDifference:       0,
	}
	err := genState.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "denom cannot be empty")
}

func TestGenesisState_Validate_InvalidDecimalDifference(t *testing.T) {
	// Test case: invalid genesis state with invalid DecimalDifference
	genState := &types.GenesisState{
		Params:                  types.NewParams(),
		TotalTransferredIn:      math.ZeroInt(),
		TotalTransferredOut:     math.ZeroInt(),
		OperatorAddress:         sample.AccAddress(),
		ProposedOperatorAddress: sample.AccAddress(),
		PauserAddresses:         []string{},
		Enabled:                 true,
		NativeClientId:          "07-tendermint-0",
		CounterpartyClientId:    "07-tendermint-1",
		NativePort:              "transfer",
		CounterpartyPort:        "transfer",
		NativeChannel:           "channel-0",
		CounterpartyChannel:     "channel-1",
		Denom:                   "uzig",
		DecimalDifference:       256,
	}
	err := genState.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "decimal difference cannot be greater than 18")
}

func TestGenesisState_Validate_InvalidOverMaxDecimalDifference(t *testing.T) {
	// Test case: invalid genesis state with DecimalDifference over maximum allowed
	genState := &types.GenesisState{
		Params:                  types.NewParams(),
		TotalTransferredIn:      math.ZeroInt(),
		TotalTransferredOut:     math.ZeroInt(),
		OperatorAddress:         sample.AccAddress(),
		ProposedOperatorAddress: sample.AccAddress(),
		PauserAddresses:         []string{},
		Enabled:                 true,
		NativeClientId:          "07-tendermint-0",
		CounterpartyClientId:    "07-tendermint-1",
		NativePort:              "transfer",
		CounterpartyPort:        "transfer",
		NativeChannel:           "channel-0",
		CounterpartyChannel:     "channel-1",
		Denom:                   "uzig",
		DecimalDifference:       19,
	}
	err := genState.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "decimal difference cannot be greater than 18")
}

func TestGenesisState_Validate_InvalidCounterpartyClientID(t *testing.T) {
	// Test case: invalid genesis state with invalid CounterpartyClientId
	genState := &types.GenesisState{
		Params:                  types.NewParams(),
		TotalTransferredIn:      math.ZeroInt(),
		TotalTransferredOut:     math.ZeroInt(),
		OperatorAddress:         sample.AccAddress(),
		ProposedOperatorAddress: sample.AccAddress(),
		PauserAddresses:         []string{},
		Enabled:                 true,
		NativeClientId:          "07-tendermint-0",
		CounterpartyClientId:    "invalid-client",
		NativePort:              "transfer",
		CounterpartyPort:        "transfer",
		NativeChannel:           "channel-0",
		CounterpartyChannel:     "channel-1",
		Denom:                   "uzig",
		DecimalDifference:       0,
	}
	err := genState.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "light client is invalid: invalid client ID format")
}

func TestGenesisState_Validate_InvalidEmptyCounterpartyClientId(t *testing.T) {
	// Test case: invalid genesis state with empty CounterpartyClientId
	genState := &types.GenesisState{
		Params:                  types.NewParams(),
		TotalTransferredIn:      math.ZeroInt(),
		TotalTransferredOut:     math.ZeroInt(),
		OperatorAddress:         sample.AccAddress(),
		ProposedOperatorAddress: sample.AccAddress(),
		PauserAddresses:         []string{},
		Enabled:                 true,
		NativeClientId:          "07-tendermint-0",
		CounterpartyClientId:    "",
		NativePort:              "transfer",
		CounterpartyPort:        "transfer",
		NativeChannel:           "channel-0",
		CounterpartyChannel:     "channel-1",
		Denom:                   "uzig",
		DecimalDifference:       0,
	}
	err := genState.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "light client is invalid: client ID cannot be empty")
}

func TestGenesisState_Validate_InvalidCounterpartyPort(t *testing.T) {
	// Test case: invalid genesis state with invalid CounterpartyPort
	genState := &types.GenesisState{
		Params:                  types.NewParams(),
		TotalTransferredIn:      math.ZeroInt(),
		TotalTransferredOut:     math.ZeroInt(),
		OperatorAddress:         sample.AccAddress(),
		ProposedOperatorAddress: sample.AccAddress(),
		PauserAddresses:         []string{},
		Enabled:                 true,
		NativeClientId:          "07-tendermint-0",
		CounterpartyClientId:    "07-tendermint-1",
		NativePort:              "transfer",
		CounterpartyPort:        "invalid/port",
		NativeChannel:           "channel-0",
		CounterpartyChannel:     "channel-1",
		Denom:                   "uzig",
		DecimalDifference:       0,
	}
	err := genState.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid port: port contains invalid characters. Only alphanumeric, ., _, +, -, #, [, ], <, > are allowed")
}

func TestGenesisState_Validate_InvalidEmptyCounterpartyPort(t *testing.T) {
	// Test case: invalid genesis state with empty CounterpartyPort
	genState := &types.GenesisState{
		Params:                  types.NewParams(),
		TotalTransferredIn:      math.ZeroInt(),
		TotalTransferredOut:     math.ZeroInt(),
		OperatorAddress:         sample.AccAddress(),
		ProposedOperatorAddress: sample.AccAddress(),
		PauserAddresses:         []string{},
		Enabled:                 true,
		NativeClientId:          "07-tendermint-0",
		CounterpartyClientId:    "07-tendermint-1",
		NativePort:              "transfer",
		CounterpartyPort:        "",
		NativeChannel:           "channel-0",
		CounterpartyChannel:     "channel-1",
		Denom:                   "uzig",
		DecimalDifference:       0,
	}
	err := genState.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid port: port cannot be empty")
}

func TestGenesisState_Validate_InvalidCounterpartyChannel(t *testing.T) {
	// Test case: invalid genesis state with invalid CounterpartyChannel
	genState := &types.GenesisState{
		Params:                  types.NewParams(),
		TotalTransferredIn:      math.ZeroInt(),
		TotalTransferredOut:     math.ZeroInt(),
		OperatorAddress:         sample.AccAddress(),
		ProposedOperatorAddress: sample.AccAddress(),
		PauserAddresses:         []string{},
		Enabled:                 true,
		NativeClientId:          "07-tendermint-0",
		CounterpartyClientId:    "07-tendermint-1",
		NativePort:              "transfer",
		CounterpartyPort:        "transfer",
		NativeChannel:           "channel-0",
		CounterpartyChannel:     "invalid-channel",
		Denom:                   "uzig",
		DecimalDifference:       0,
	}
	err := genState.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid channel identifier: invalid channel ID format")
}

func TestGenesisState_Validate_InvalidEmptyCounterpartyChannel(t *testing.T) {
	// Test case: invalid genesis state with empty CounterpartyChannel
	genState := &types.GenesisState{
		Params:                  types.NewParams(),
		TotalTransferredIn:      math.ZeroInt(),
		TotalTransferredOut:     math.ZeroInt(),
		OperatorAddress:         sample.AccAddress(),
		ProposedOperatorAddress: sample.AccAddress(),
		PauserAddresses:         []string{},
		Enabled:                 true,
		NativeClientId:          "07-tendermint-0",
		CounterpartyClientId:    "07-tendermint-1",
		NativePort:              "transfer",
		CounterpartyPort:        "transfer",
		NativeChannel:           "channel-0",
		CounterpartyChannel:     "",
		Denom:                   "uzig",
		DecimalDifference:       0,
	}
	err := genState.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid channel identifier: channel cannot be empty")
}
