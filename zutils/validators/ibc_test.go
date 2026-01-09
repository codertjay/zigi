package validators_test

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"zigchain/zutils/constants"
	"zigchain/zutils/validators"

	"fmt"
	"regexp"
)

func TestValidatePort(t *testing.T) {
	tests := []struct {
		name    string
		port    interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid port",
			port:    "transfer",
			wantErr: false,
		},
		{
			name:    "empty port",
			port:    "",
			wantErr: true,
			errMsg:  "port cannot be empty",
		},
		{
			name:    "port too short",
			port:    "a",
			wantErr: true,
			errMsg:  "port length must be between 2 and 128 characters",
		},
		{
			name:    "port too long",
			port:    makeString(129),
			wantErr: true,
			errMsg:  "port length must be between 2 and 128 characters",
		},
		{
			name:    "port with invalid chars",
			port:    "transfer/port",
			wantErr: true,
			errMsg:  "port contains invalid characters",
		},
		{
			name:    "non-string input",
			port:    123,
			wantErr: true,
			errMsg:  "invalid parameter type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validators.ValidatePort(tt.port)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestValidateChannel(t *testing.T) {
	tests := []struct {
		name    string
		channel interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid channel",
			channel: "channel-0",
			wantErr: false,
		},
		{
			name:    "empty channel",
			channel: "",
			wantErr: true,
			errMsg:  "channel cannot be empty",
		},
		{
			name:    "channel too short",
			channel: "ch",
			wantErr: true,
			errMsg:  "invalid channel identifier: invalid channel ID format",
		},
		{
			name:    "channel too long",
			channel: makeString(65),
			wantErr: true,
			errMsg:  "invalid channel identifier: invalid channel ID format",
		},
		{
			name:    "channel with invalid chars",
			channel: "channel/0",
			wantErr: true,
			errMsg:  "invalid channel identifier: invalid channel ID format",
		},
		{
			name:    "non-string input",
			channel: 123,
			wantErr: true,
			errMsg:  "invalid parameter type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validators.ValidateChannel(tt.channel)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
		})
	}
}

// ValidateDenom is the default validation function for Coin.Denom.
func MockSDKValidateDenom(denom string) error {
	var (
		// Denominations can be 3 ~ 128 characters long and support letters, followed by either
		// a letter, a number or a separator ('/', ':', '.', '_' or '-').
		reDnmString = `[a-zA-Z][a-zA-Z0-9/:._-]{2,127}`
		reDnm       *regexp.Regexp
	)

	reDnm = regexp.MustCompile(fmt.Sprintf(`^%s$`, reDnmString))

	if !reDnm.MatchString(denom) {
		return fmt.Errorf("invalid denom: %s", denom)
	}
	return nil
}

func TestSDKValidateDenom(t *testing.T) {
	tests := []struct {
		name    string
		denom   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid denom with dash",
			denom:   "unit-zig",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MockSDKValidateDenom(tt.denom)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
				return
			}
			err = sdk.ValidateDenom(tt.denom)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestValidateDenom(t *testing.T) {
	tests := []struct {
		name    string
		denom   interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid denom",
			denom:   "uzig",
			wantErr: false,
		},
		{
			name:    "valid denom with dash",
			denom:   "unit-zig",
			wantErr: false,
		},
		{
			name:    "empty denom",
			denom:   "",
			wantErr: true,
			errMsg:  "denom cannot be empty",
		},
		{
			name:    "denom with invalid chars",
			denom:   "uzig/token",
			wantErr: true,
			errMsg:  "denom contains invalid characters",
		},
		{
			name:    "non-string input",
			denom:   123,
			wantErr: true,
			errMsg:  "invalid parameter type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validators.ValidateDenom(tt.denom)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestIsValidIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid identifier with alphanumeric",
			input:    "abc123",
			expected: true,
		},
		{
			name:     "valid identifier with special chars",
			input:    "abc.123_+-#[]<>",
			expected: true,
		},
		{
			name:     "invalid identifier with slash",
			input:    "abc/123",
			expected: false,
		},
		{
			name:     "invalid identifier with space",
			input:    "abc 123",
			expected: false,
		},
		{
			name:     "invalid identifier with special char",
			input:    "abc@123",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validators.IsValidIdentifier(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateDecimalDifference(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid decimal difference - zero",
			input:   uint32(0),
			wantErr: false,
		},
		{
			name:    "valid decimal difference - middle value",
			input:   uint32(9),
			wantErr: false,
		},
		{
			name:    "valid decimal difference - maximum allowed",
			input:   uint32(18),
			wantErr: false,
		},
		{
			name:    "invalid decimal difference - exceeds maximum",
			input:   uint32(19),
			wantErr: true,
			errMsg:  "decimal difference cannot be greater than 18",
		},
		{
			name:    "invalid decimal difference - negative value",
			input:   int32(-1),
			wantErr: true,
			errMsg:  "invalid parameter type",
		},
		{
			name:    "invalid input type - string",
			input:   "5",
			wantErr: true,
			errMsg:  "invalid parameter type",
		},
		{
			name:    "invalid input type - float",
			input:   5.5,
			wantErr: true,
			errMsg:  "invalid parameter type",
		},
		{
			name:    "invalid input type - nil",
			input:   nil,
			wantErr: true,
			errMsg:  "invalid parameter type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validators.ValidateDecimalDifference(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
		})
	}
}

// Helper function to create a string of specified length
func makeString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}

func TestValidateClientId(t *testing.T) {
	tests := []struct {
		name     string
		clientId interface{}
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid client ID",
			clientId: "07-tendermint-0",
			wantErr:  false,
		},
		{
			name:     "valid client ID with different type",
			clientId: "08-solo-123",
			wantErr:  false,
		},
		{
			name:     "empty client ID",
			clientId: "",
			wantErr:  true,
			errMsg:   "client ID cannot be empty",
		},
		{
			name:     "invalid client ID format - missing number",
			clientId: "07-tendermint",
			wantErr:  true,
			errMsg:   "invalid client ID format",
		},
		{
			name:     "invalid client ID format - invalid characters",
			clientId: "07-tendermint-@",
			wantErr:  true,
			errMsg:   "invalid client ID format",
		},
		{
			name:     "non-string input - integer",
			clientId: 123,
			wantErr:  true,
			errMsg:   "invalid parameter type",
		},
		{
			name:     "non-string input - nil",
			clientId: nil,
			wantErr:  true,
			errMsg:   "invalid parameter type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validators.ValidateClientId(tt.clientId)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestValidateDenom_CheckDenomString(t *testing.T) {
	tests := []struct {
		name    string
		denom   interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid denom - minimum length",
			denom:   "abc",
			wantErr: false,
		},
		{
			name:    "valid denom - maximum length",
			denom:   "abc" + strings.Repeat("d", constants.MaxDenomLength-3),
			wantErr: false,
		},
		{
			name:    "empty denom",
			denom:   "",
			wantErr: true,
			errMsg:  "denom cannot be empty",
		},
		{
			name:    "denom too short",
			denom:   "ab",
			wantErr: true,
			errMsg: fmt.Sprintf(
				"invalid coin: 'ab' denom name is too short, minimum %d characters e.g. 10uzig: invalid coins",
				constants.MinSubDenomLength,
			),
		},
		{
			name:    "denom too long",
			denom:   "abc" + strings.Repeat("d", constants.MaxDenomLength),
			wantErr: true,
			errMsg: fmt.Sprintf(
				"invalid coin: 'abc%s' denom name is too long (%d), maximum %d characters e.g. uzig: invalid coins",
				strings.Repeat("d", constants.MaxDenomLength),
				constants.MaxDenomLength+3,
				constants.MaxDenomLength,
			),
		},
		{
			name:    "denom with invalid characters",
			denom:   "abc#123",
			wantErr: true,
			errMsg:  "invalid coin: 'abc#123' only letters (a-z, A-Z), numbers (0-9), dots (.) and forward slashes (/) are allowed e.g. 10uzig: invalid coins",
		},
		{
			name:    "non-string input",
			denom:   123,
			wantErr: true,
			errMsg:  "invalid parameter type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validators.ValidateDenom(tt.denom)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
		})
	}
}
