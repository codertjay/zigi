package types_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/assert"

	errorPacks "zigchain/testutil/data"
	"zigchain/x/factory/types"
)

// Positive test cases

func TestGetTokenDenom_ValidSubdenom(t *testing.T) {
	// Test cases for GetTokenDenom function with valid inputs for subdenom
	// The function should return the correct denom and no error

	defaultCreator := "zig1vm3v4yrd3rrwkf3fe8qxutaz27098t76270qc5"

	tests := []struct {
		name     string
		creator  string
		subdenom string
		expected string
	}{
		{
			name:     "normal",
			creator:  defaultCreator,
			subdenom: "bitcoin",
			expected: "coin." + defaultCreator + ".bitcoin",
		},
		{
			name:     "subdenom min length",
			creator:  defaultCreator,
			subdenom: "bit",
			expected: "coin." + defaultCreator + ".bit",
		},
		{
			name:     "subdenom max length",
			creator:  defaultCreator,
			subdenom: strings.Repeat("a", 44),
			expected: "coin." + defaultCreator + "." + strings.Repeat("a", 44),
		},
		{
			name:     "subdenom with numbers",
			creator:  defaultCreator,
			subdenom: "bit0123456789",
			expected: "coin." + defaultCreator + ".bit0123456789",
		},
		{
			name:     "subdenom lowercase letters",
			creator:  defaultCreator,
			subdenom: "abcdefghijklmnopqrstuvwxyz",
			expected: "coin." + defaultCreator + ".abcdefghijklmnopqrstuvwxyz",
		},
		{
			name:     "creator max length",
			creator:  strings.Repeat("a", 75),
			subdenom: "bitcoin",
			expected: "coin." + strings.Repeat("a", 75) + ".bitcoin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := types.GetTokenDenom(tt.creator, tt.subdenom)
			require.NoError(t, err)
			require.Equal(t, tt.expected, got)
		})
	}
}

func TestGetTokenDenom_ValidCreator(t *testing.T) {
	// Test cases for GetTokenDenom function with valid inputs for creator
	// The function should return the correct denom and no error

	tests := []struct {
		name     string
		creator  string
		subdenom string
		expected string
	}{
		{
			name:     "creator max length",
			creator:  strings.Repeat("a", 75),
			subdenom: "bitcoin",
			expected: "coin." + strings.Repeat("a", 75) + ".bitcoin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := types.GetTokenDenom(tt.creator, tt.subdenom)
			require.NoError(t, err)
			require.Equal(t, tt.expected, got)
		})
	}
}

func TestDeconstructDenom_Positive(t *testing.T) {
	// Test case for DeconstructDenom function with valid inputs

	defaultCreator := "zig1vm3v4yrd3rrwkf3fe8qxutaz27098t76270qc5"

	tests := []struct {
		name       string
		denom      string
		expCreator string
		expSub     string
	}{
		{
			name:       "normal",
			denom:      "coin" + types.FactoryDenomDelimiterChar + "zig1vm3v4yrd3rrwkf3fe8qxutaz27098t76270qc5" + types.FactoryDenomDelimiterChar + "bitcoin",
			expCreator: defaultCreator,
			expSub:     "bitcoin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCreator, gotSub, err := types.DeconstructDenom(tt.denom)
			require.NoError(t, err)
			require.Equal(t, tt.expCreator, gotCreator)
			require.Equal(t, tt.expSub, gotSub)
		})
	}
}

// Negative test cases

func TestGetTokenDenom_InvalidSubDenom(t *testing.T) {
	// Test case: invalid subdenom
	// The function should return an error

	// Get the invalid Subdenom from the error pack
	var invalidSubDenom = &errorPacks.InvalidSubDenomString

	var creator = "zig1vm3v4yrd3rrwkf3fe8qxutaz27098t76270qc5"

	// Loop through the invalid SubDenoms
	for _, tc := range *invalidSubDenom {
		t.Run(tc.TestName, func(t *testing.T) {

			value, err := types.GetTokenDenom(creator, tc.FieldValue)

			require.Error(t, err)

			// Assert that the value returned is an empty string
			assert.Equal(t, "", value)

			// Assert that the error message is equal to the expected error message
			// append to the Error Text ": invalid coins
			var errorText = tc.ErrorText + ": invalid coins"

			assert.Equal(t, errorText, err.Error())

		})
	}
}

func TestGetTokenDenom_InvalidCreator(t *testing.T) {
	// Test case: invalid subdenom
	// The function should return an error

	tests := []struct {
		name     string
		creator  string
		subdenom string
		errStr   string
	}{
		{
			name:     "creator too long",
			creator:  strings.Repeat("a", 76),
			subdenom: "bitcoin",
			errStr:   "creator too long, max length is 75 bytes",
		},
		{
			name:     "creator contains /",
			creator:  "zig1vm3v4yrd3rrwkf3fe/qxutaz27098t76270qc5",
			subdenom: "bitcoin",
			errStr:   "Signer account is not valid",
		},
		{
			name:     "creator contains .",
			creator:  "zig1vm3v4yrd3rrwkf3fe.qxutaz27098t76270qc5",
			subdenom: "bitcoin",
			errStr:   "Signer account is not valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := types.GetTokenDenom(tt.creator, tt.subdenom)
			require.Error(t, err)
			require.Equal(t, "", got)
			require.EqualError(t, err, tt.errStr)
		})
	}
}

func TestDeconstructDenom_InvalidDenom(t *testing.T) {
	// Test cases for DeconstructDenom function with invalid denom

	defaultCreator := "zig1vm3v4yrd3rrwkf3fe8qxutaz27098t76270qc5"

	tests := []struct {
		name   string
		denom  string
		errStr string
	}{
		{
			name:   "empty denom",
			denom:  "",
			errStr: "invalid denom: ",
		},
		{
			name:   "no delimiters",
			denom:  "coin" + defaultCreator + "bitcoin",
			errStr: "not enough parts of denom coin" + defaultCreator + "bitcoin: Factory Denom name is not valid",
		},
		{
			name:   "too many parts",
			denom:  "coin." + defaultCreator + ".bitcoin.1",
			errStr: "too many parts of denom coin." + defaultCreator + ".bitcoin.1: Factory Denom name is not valid",
		},
		{
			name:   "too few parts",
			denom:  "coin." + defaultCreator,
			errStr: "not enough parts of denom coin." + defaultCreator + ": Factory Denom name is not valid",
		},
		{
			name:   "wrong prefix",
			denom:  "ibc." + defaultCreator + ".bitcoin",
			errStr: "denom prefix is incorrect. Is: ibc.  Should be: coin: Factory Denom name is not valid",
		},
		{
			name:   "empty subdenom",
			denom:  "coin." + defaultCreator + ".",
			errStr: "subdenom is empty: Factory Denom name is not valid",
		},
		{
			name:   "invalid subdenom char",
			denom:  "coin." + defaultCreator + ".bit_coin",
			errStr: "invalid denom: coin.zig1vm3v4yrd3rrwkf3fe8qxutaz27098t76270qc5.bit_coin",
		},
		{
			name:   "denom includes space",
			denom:  "coinzig1vm3v4yrd3rrwkf 3fe8qxutaz27098t76270qc5bitcoin",
			errStr: "invalid denom: coinzig1vm3v4yrd3rrwkf 3fe8qxutaz27098t76270qc5bitcoin",
		},
		{
			name:   "denom includes not allowed characters :_",
			denom:  "coin" + types.FactoryDenomDelimiterChar + "zig1vm3v4yrd3rrwkf3fe8qxutaz27098t76270qc5" + types.FactoryDenomDelimiterChar + "bitcoin:_",
			errStr: "invalid denom: coin.zig1vm3v4yrd3rrwkf3fe8qxutaz27098t76270qc5.bitcoin:_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCreator, gotSub, err := types.DeconstructDenom(tt.denom)
			require.Error(t, err)
			require.Equal(t, "", gotCreator)
			require.Equal(t, "", gotSub)
			require.EqualError(t, err, tt.errStr)
		})
	}
}

func TestDeconstructDenom_InvalidCreator(t *testing.T) {
	// Test cases with invalid Creator address

	tests := []struct {
		name   string
		denom  string
		errStr string
	}{
		{
			name:   "empty creator",
			denom:  "coin..bitcoin",
			errStr: "Invalid creator address (Creator address: cannot be empty: invalid address): Factory Denom name is not valid",
		},
		{
			name:   "invalid creator address",
			denom:  "coin.invalidaddr.bitcoin",
			errStr: "Invalid creator address (Creator address: 'invalidaddr' (decoding bech32 failed: invalid separator index -1): invalid address): Factory Denom name is not valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCreator, gotSub, err := types.DeconstructDenom(tt.denom)
			require.Error(t, err)
			require.Equal(t, "", gotCreator)
			require.Equal(t, "", gotSub)
			require.EqualError(t, err, tt.errStr)
		})
	}
}

// this function is nto designed to validate
func TestDeconstructDenom_InvalidDenom_InvalidSubDenom(t *testing.T) {
	// Test case: invalid subdenom
	// The function should return an error

	// Get the invalid Subdenom from the error pack
	var invalidSubDenom = &errorPacks.InvalidSubDenomString

	// Loop through the invalid SubDenoms
	for _, tc := range *invalidSubDenom {
		t.Run(tc.TestName, func(t *testing.T) {

			denom := "coin" + types.FactoryDenomDelimiterChar + "zig1vm3v4yrd3rrwkf3fe8qxutaz27098t76270qc5" + types.FactoryDenomDelimiterChar + tc.FieldValue
			creator, subdenom, err := types.DeconstructDenom(denom)

			require.Error(t, err)

			// Assert that the creator and subdenom returned is an empty string
			assert.Equal(t, "", creator)
			assert.Equal(t, "", subdenom)
		})
	}
}
