package types

import (
	"fmt"
	"strings"
	"testing"

	errorPacks "zigchain/testutil/data"
	types "zigchain/testutil/helpers"
	"zigchain/testutil/sample"
	"zigchain/zutils/validators"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

var MsgUpdateIbcSettingsSample = MsgUpdateIbcSettings{
	Signer:               sample.AccAddress(),
	NativeClientId:       "07-tendermint-0",
	CounterpartyClientId: "07-tendermint-0",
	NativePort:           "transfer",
	CounterpartyPort:     "transfer",
	NativeChannel:        "channel-0",
	CounterpartyChannel:  "channel-0",
	Denom:                "uzig",
	DecimalDifference:    uint32(12),
}

// Positive test cases

func TestMsgUpdateIbcSettings_NewMsgUpdateIbcSettings_Positive(t *testing.T) {
	// Test case: Valid input data

	signer := sample.AccAddress()
	nativeClientId := "07-tendermint-0"
	counterpartyClientId := "07-tendermint-0"
	nativePort := "transfer"
	counterpartyPort := "transfer"
	nativeChannel := "channel-0"
	counterpartyChannel := "channel-0"
	denom := "uzig"
	decimalDifference := uint32(12)

	// create a new MsgUpdateIbcSettings instance
	msg := NewMsgUpdateIbcSettings(signer, nativeClientId, counterpartyClientId, nativePort, counterpartyPort, nativeChannel, counterpartyChannel, denom, decimalDifference)

	// validate the fields
	require.NotNil(t, msg, "expected the message to be non-nil")
	require.Equal(t, signer, msg.Signer, "expected the signer to match the input signer")
	require.Equal(t, nativeClientId, msg.NativeClientId, "expected the native client ID to match the input native client ID")
	require.Equal(t, counterpartyClientId, msg.CounterpartyClientId, "expected the counterparty client ID to match the input counterparty client ID")
	require.Equal(t, nativePort, msg.NativePort, "expected the native port to match the input native port")
	require.Equal(t, counterpartyPort, msg.CounterpartyPort, "expected the counterparty port to match the input counterparty port")
	require.Equal(t, nativeChannel, msg.NativeChannel, "expected the native channel to match the input native channel")
	require.Equal(t, counterpartyChannel, msg.CounterpartyChannel, "expected the counterparty channel to match the input counterparty channel")
	require.Equal(t, denom, msg.Denom, "expected the denom to match the input denom")
	require.Equal(t, decimalDifference, msg.DecimalDifference, "expected the decimal difference to match the input decimal difference")
}

func TestMsgUpdateIbcSettings_ValidateBasic_ValidClientID(t *testing.T) {
	// Test case: Valid ClientID

	tests := []struct {
		name        string
		clientID    string
		shouldPass  bool
		description string
	}{
		{
			name:        "invalid simple client id",
			clientID:    "client123",
			shouldPass:  false,
			description: "simple alphanumeric client ID",
		},
		{
			name:        "invalid complex client id",
			clientID:    "client.id_1+2-3#[4]<5>",
			shouldPass:  false,
			description: "client ID with invalid special characters",
		},
		{
			name:        "invalid minimum length",
			clientID:    "abc",
			shouldPass:  false,
			description: "client ID at minimum length (3 chars)",
		},
		{
			name:        "invalid maximum length",
			clientID:    "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefgh",
			shouldPass:  false,
			description: "client ID at maximum length (64 chars)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := MsgUpdateIbcSettingsSample
			msg.NativeClientId = tt.clientID

			err := validators.ValidateClientId(msg.NativeClientId)
			if tt.shouldPass {
				require.NoError(t, err, "expected valid ClientID '%s' to pass validation: %s", tt.clientID, tt.description)
			} else {
				require.Error(t, err, "expected invalid ClientID '%s' to fail validation: %s", tt.clientID, tt.description)
			}
		})
	}
}

func TestValidatePort_Valid(t *testing.T) {
	validPorts := []string{
		"port1",
		"port.name_with+valid#chars[0]<x>",
		"ab",                           // min length
		"p" + strings.Repeat("x", 127), // 128 chars
	}

	for _, port := range validPorts {
		t.Run("valid: "+port, func(t *testing.T) {
			err := validators.ValidatePort(port)
			require.NoError(t, err, "expected valid port '%s' to pass validation", port)
		})
	}
}

func TestValidateChannel(t *testing.T) {
	// Test case: Valid Channel

	tests := []struct {
		name        string
		channelID   string
		shouldPass  bool
		description string
	}{
		{
			name:        "valid standard format",
			channelID:   "channel-01",
			shouldPass:  true,
			description: "standard channel format with number",
		},
		{
			name:        "invalid with special chars",
			channelID:   "channel.name_with+invalid#chars[0]<x>",
			shouldPass:  false,
			description: "channel with invalid special characters",
		},
		{
			name:        "invalid common format",
			channelID:   "chan.port1",
			shouldPass:  false,
			description: "common channel format with invalid port",
		},
		{
			name:        "invalid max length",
			channelID:   "cha_" + strings.Repeat("x", 60),
			shouldPass:  false,
			description: "channel at maximum length (64 chars)",
		},
		{
			name:        "invalid min length",
			channelID:   "abcdefgh",
			shouldPass:  false,
			description: "channel at minimum length (8 chars)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validators.ValidateChannel(tt.channelID)
			if tt.shouldPass {
				require.NoError(t, err, "expected valid channel '%s' to pass validation: %s", tt.channelID, tt.description)
			} else {
				require.Error(t, err, "expected invalid channel '%s' to fail validation: %s", tt.channelID, tt.description)
			}
		})
	}
}

func TestValidateDecimalDifference_Valid(t *testing.T) {
	// Test case: Valid DecimalDifference

	validValues := []uint32{
		0,
		1,
		18, // upper bound
	}

	for _, val := range validValues {
		t.Run(fmt.Sprintf("valid: %d", val), func(t *testing.T) {
			err := validators.ValidateDecimalDifference(val)
			require.NoError(t, err)
		})
	}
}

func TestMsgUpdateIbcSettings_ValidateBasic_Positive(t *testing.T) {
	// Test case: Validate basic properties of MsgUpdateIbcSettings

	// make a copy of sample message
	msg := MsgUpdateIbcSettingsSample

	// print the message
	fmt.Println("MsgUpdateIbcSettingsSample:", msg)

	// validate the message
	err := msg.ValidateBasic()

	// assert that there are no errors
	require.NoError(t, err)
}

// Negative test cases

func TestMsgUpdateIbcSettings_ValidateBasic_InvalidSignerAddress(t *testing.T) {
	// Test case: invalid Signer address

	// make a copy of sample message
	msg := MsgUpdateIbcSettingsSample

	// set invalid signer address and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Signer,
		&errorPacks.InvalidSignerAddress,
		sdkerrors.ErrInvalidAddress,
	)
}

func TestMsgUpdateIbcSettings_ValidateBasic_InvalidNativeClientId(t *testing.T) {
	// Test case: invalid NativeClientId values

	for _, tc := range errorPacks.InvalidClientID {
		t.Run(tc.TestName, func(t *testing.T) {
			msg := MsgUpdateIbcSettingsSample
			msg.NativeClientId = tc.FieldValue

			err := msg.ValidateBasic()
			require.Error(t, err)
			require.EqualError(t, err, tc.ErrorText)
		})
	}
}

func TestMsgUpdateIbcSettings_ValidateClientId_InvalidClientID(t *testing.T) {
	// Test case: invalid ClientID

	for _, tc := range errorPacks.InvalidClientID {
		t.Run(tc.TestName, func(t *testing.T) {
			msg := MsgUpdateIbcSettingsSample
			msg.NativeClientId = tc.FieldValue

			err := validators.ValidateClientId(msg.NativeClientId)
			require.Error(t, err)
			require.EqualError(t, err, tc.ErrorText)
		})
	}
}

func TestMsgUpdateIbcSettings_ValidateClientId_InvalidClientIDTypes(t *testing.T) {
	// Test case: invalid ClientID types

	for _, tc := range errorPacks.InvalidClientIDTypes {
		t.Run(tc.TestName, func(t *testing.T) {
			err := validators.ValidateClientId(tc.Input)
			require.Error(t, err)
			require.EqualError(t, err, tc.ErrorText)
		})
	}
}

func TestMsgUpdateIbcSettings_ValidateBasic_InvalidNativePort(t *testing.T) {
	// Test case: invalid NativePort values

	for _, tc := range errorPacks.InvalidPort {
		t.Run(tc.TestName, func(t *testing.T) {
			msg := MsgUpdateIbcSettingsSample
			msg.NativePort = tc.FieldValue

			err := msg.ValidateBasic()
			require.Error(t, err)
			require.EqualError(t, err, tc.ErrorText)
		})
	}
}

func TestMsgUpdateIbcSettings_ValidatePort_InvalidNativePort(t *testing.T) {
	// Test case: invalid NativePort values

	for _, tc := range errorPacks.InvalidPort {
		t.Run(tc.TestName, func(t *testing.T) {
			msg := MsgUpdateIbcSettingsSample
			msg.NativePort = tc.FieldValue

			err := validators.ValidatePort(msg.NativePort)
			require.Error(t, err)
			require.EqualError(t, err, tc.ErrorText)
		})
	}
}

func TestMsgUpdateIbcSettings_ValidatePort_InvalidNativePortTypes(t *testing.T) {
	// Test case: invalid NativePort types

	for _, tc := range errorPacks.InvalidPortTypes {
		t.Run(tc.TestName, func(t *testing.T) {
			err := validators.ValidatePort(tc.Input)
			require.Error(t, err)
			require.EqualError(t, err, tc.ErrorText)
		})
	}
}

func TestMsgUpdateIbcSettings_ValidateBasic_InvalidNativeChannel(t *testing.T) {
	// Test case: invalid NativeChannel values

	for _, tc := range errorPacks.InvalidChannel {
		t.Run(tc.TestName, func(t *testing.T) {
			msg := MsgUpdateIbcSettingsSample
			msg.NativeChannel = tc.FieldValue

			err := msg.ValidateBasic()
			require.Error(t, err)
			require.EqualError(t, err, tc.ErrorText)
		})
	}
}

func TestMsgUpdateIbcSettings_ValidateChannel_InvalidNativeChannel(t *testing.T) {
	// Test case: invalid NativeChannel values

	for _, tc := range errorPacks.InvalidChannel {
		t.Run(tc.TestName, func(t *testing.T) {
			msg := MsgUpdateIbcSettingsSample
			msg.NativeChannel = tc.FieldValue

			err := validators.ValidateChannel(msg.NativeChannel)
			require.Error(t, err)
			require.EqualError(t, err, tc.ErrorText)
		})
	}
}

func TestMsgUpdateIbcSettings_ValidateChannel_InvalidNativeChannelTypes(t *testing.T) {
	// Test case: invalid NativeChannel types

	for _, tc := range errorPacks.InvalidChannelTypes {
		t.Run(tc.TestName, func(t *testing.T) {
			err := validators.ValidateChannel(tc.Input)
			require.Error(t, err)
			require.EqualError(t, err, tc.ErrorText)
		})
	}
}
func TestMsgUpdateIbcSettings_ValidateBasic_InvalidCounterpartyClientId(t *testing.T) {
	// Test case: invalid CounterpartyClientId values

	for _, tc := range errorPacks.InvalidClientID {
		t.Run(tc.TestName, func(t *testing.T) {
			msg := MsgUpdateIbcSettingsSample
			msg.CounterpartyClientId = tc.FieldValue

			err := msg.ValidateBasic()
			require.Error(t, err)
			require.EqualError(t, err, tc.ErrorText)
		})
	}
}

func TestMsgUpdateIbcSettings_ValidateBasic_InvalidCounterpartyPort(t *testing.T) {
	// Test case: invalid CounterpartyPort values

	for _, tc := range errorPacks.InvalidPort {
		t.Run(tc.TestName, func(t *testing.T) {
			msg := MsgUpdateIbcSettingsSample
			msg.CounterpartyPort = tc.FieldValue

			err := msg.ValidateBasic()
			require.Error(t, err)
			require.EqualError(t, err, tc.ErrorText)
		})
	}
}

func TestMsgUpdateIbcSettings_ValidateClientId_InvalidCounterpartyClientId(t *testing.T) {
	// Test case: invalid CounterpartyClientId values

	for _, tc := range errorPacks.InvalidClientID {
		t.Run(tc.TestName, func(t *testing.T) {
			msg := MsgUpdateIbcSettingsSample
			msg.CounterpartyClientId = tc.FieldValue

			err := validators.ValidateClientId(msg.CounterpartyClientId)
			require.Error(t, err)
			require.EqualError(t, err, tc.ErrorText)
		})
	}
}

func TestMsgUpdateIbcSettings_ValidateBasic_InvalidCounterpartyChannel(t *testing.T) {
	// Test case: invalid CounterpartyChannel values

	for _, tc := range errorPacks.InvalidChannel {
		t.Run(tc.TestName, func(t *testing.T) {
			msg := MsgUpdateIbcSettingsSample
			msg.CounterpartyChannel = tc.FieldValue

			err := msg.ValidateBasic()
			require.Error(t, err)
			require.EqualError(t, err, tc.ErrorText)
		})
	}
}

func TestMsgUpdateIbcSettings_ValidateChannel_InvalidCounterpartyChannel(t *testing.T) {
	// Test case: invalid CounterpartyChannel values

	for _, tc := range errorPacks.InvalidChannel {
		t.Run(tc.TestName, func(t *testing.T) {
			msg := MsgUpdateIbcSettingsSample
			msg.CounterpartyChannel = tc.FieldValue

			err := validators.ValidateChannel(msg.CounterpartyChannel)
			require.Error(t, err)
			require.EqualError(t, err, tc.ErrorText)
		})
	}
}

func TestMsgUpdateIbcSettings_ValidateBasic_InvalidDenom(t *testing.T) {
	// Test case: invalid Denom values

	for _, tc := range errorPacks.InvalidDenom {
		t.Run(tc.TestName, func(t *testing.T) {
			msg := MsgUpdateIbcSettingsSample
			msg.Denom = tc.FieldValue

			err := msg.ValidateBasic()
			require.Error(t, err)
			require.EqualError(t, err, tc.ErrorText)
		})
	}
}

func TestMsgUpdateIbcSettings_ValidateDenom_InvalidDenom(t *testing.T) {
	// Test case: invalid Denom values

	for _, tc := range errorPacks.InvalidDenom {
		t.Run(tc.TestName, func(t *testing.T) {
			msg := MsgUpdateIbcSettingsSample
			msg.Denom = tc.FieldValue

			err := validators.ValidateDenom(msg.Denom)
			require.Error(t, err)
			require.EqualError(t, err, tc.ErrorText)
		})
	}
}
func TestMsgUpdateIbcSettings_ValidateDenom_InvalidDenomTypes(t *testing.T) {
	// Test case: invalid Denom types

	for _, tc := range errorPacks.InvalidDenomTypes {
		t.Run(tc.TestName, func(t *testing.T) {
			err := validators.ValidateDenom(tc.Input)
			require.Error(t, err)
			require.EqualError(t, err, tc.ErrorText)
		})
	}
}

func TestValidateDecimalDifference_Invalid(t *testing.T) {
	// Test case: invalid DecimalDifference values

	for _, tc := range errorPacks.InvalidDecimalDifference {
		t.Run(tc.TestName, func(t *testing.T) {
			msg := MsgUpdateIbcSettingsSample
			msg.DecimalDifference = tc.FieldValue

			err := validators.ValidateDecimalDifference(msg.DecimalDifference)
			require.Error(t, err)
			require.EqualError(t, err, tc.ErrorText)
		})
	}
}

func TestValidateDecimalDifference_InvalidTypes(t *testing.T) {
	// Test case: invalid DecimalDifference types

	for _, tc := range errorPacks.InvalidDecimalDifferenceTypes {
		t.Run(tc.TestName, func(t *testing.T) {
			err := validators.ValidateDecimalDifference(tc.Input)
			require.Error(t, err)
			require.EqualError(t, err, tc.ErrorText)
		})
	}
}

func TestMsgUpdateIbcSettings_ValidateBasic_InvalidDecimalDifference(t *testing.T) {
	// Test case: invalid DecimalDifference values

	for _, tc := range errorPacks.InvalidDecimalDifference {
		t.Run(tc.TestName, func(t *testing.T) {
			msg := MsgUpdateIbcSettingsSample
			msg.DecimalDifference = tc.FieldValue

			err := msg.ValidateBasic()
			require.Error(t, err)
			require.EqualError(t, err, tc.ErrorText)
		})
	}
}
