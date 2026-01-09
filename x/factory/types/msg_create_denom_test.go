package types

import (
	"strings"
	"testing"

	cosmosmath "cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	errorPacks "zigchain/testutil/data"
	types "zigchain/testutil/helpers"
	"zigchain/testutil/sample"
	"zigchain/zutils/validators"

	"github.com/stretchr/testify/require"
)

var msgCreateDenomSample = MsgCreateDenom{
	// use sample.AccAddress() to get a valid address
	Creator:             sample.AccAddress(),
	SubDenom:            "test",
	MintingCap:          cosmosmath.NewUint(10000),
	CanChangeMintingCap: true,
	URI:                 "https://example.com",
	URIHash:             validators.SHA256HashOfURL("https://example.com"),
}

// Positive test cases

func TestMsgCreateDenom_NewMsgCreateDenom_Positive(t *testing.T) {
	// Test case: create denom with valid input data

	creator := sample.AccAddress()
	subDenom := "testsubdenom"
	mintingCap := cosmosmath.NewUint(1000000)
	uri := "https://example.com/metadata"
	uriHash := "abc123hash"

	// create a new MsgCreateDenom instance
	msg := NewMsgCreateDenom(creator, subDenom, mintingCap, true, uri, uriHash)

	// validate the fields
	require.NotNil(t, msg, "expected the message to be non-nil")
	require.Equal(t, creator, msg.Creator, "expected the creator to match the input creator")
	require.Equal(t, subDenom, msg.SubDenom, "expected the subDenom to match the input subDenom")
	require.Equal(t, mintingCap, msg.MintingCap, "expected the mintingCap to match the input mintingCap")
	require.True(t, msg.CanChangeMintingCap, "expected CanChangeMintingCap to be true")
	require.Equal(t, uri, msg.URI, "expected the URI to match the input URI")
	require.Equal(t, uriHash, msg.URIHash, "expected the URIHash to match the input URIHash")
}

func TestMsgCreateDenom_ValidateBasic_Positive(t *testing.T) {
	// Test case: validate basic properties of MsgCreateDenom

	// make a copy of sample message
	msg := msgCreateDenomSample

	// validate the message
	err := msg.ValidateBasic()

	// assert that the error is nil
	require.NoError(t, err)
}

// Negative test cases

func TestMsgCreateDenom_ValidateBasic_InvalidCreatorAddress(t *testing.T) {
	// Test case: invalid creator addresses

	// make a copy of sample message
	msg := msgCreateDenomSample

	// set the creator to an invalid address and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Creator,
		&errorPacks.InvalidSignerAddress,
		sdkerrors.ErrInvalidAddress,
	)
}

func TestMsgCreateDenom_ValidateBasic_InvalidSubDenom(t *testing.T) {
	// Test case: invalid denom

	// make a copy of sample message
	msg := msgCreateDenomSample

	// set invalid denom and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.SubDenom,
		&errorPacks.InvalidSubDenomString,
		sdkerrors.ErrInvalidCoins,
	)
}

func TestMsgCreateDenom_ValidateBasic_InvalidMintingCap(t *testing.T) {
	// Test case: invalid MintingCap

	// create a slice of subtests (define test cases with different invalid MintingCap's)
	tests := []struct {
		// name/description of the test case
		name string
		// MintingCap that will be created on MsgCreateDenom
		mintingCap cosmosmath.Uint
		// expected error message for the test case
		errorMessage string
	}{
		{
			name:         "is zero",
			mintingCap:   cosmosmath.NewUint(0),
			errorMessage: "Minting Cap 0 must be greater than 0: Minting Cap is not valid",
		},
		{
			name:         "negative number",
			mintingCap:   cosmosmath.Uint(cosmosmath.NewInt(-10)),
			errorMessage: "Minting Cap -10 must be greater than 0: Minting Cap is not valid",
		},
	}

	// make a copy of sample message
	msg := msgCreateDenomSample

	// iterate through each test case
	for _, test := range tests {
		t.Run("MsgCreateDenom.mintingCap "+test.name, func(t *testing.T) {
			// set the MintingCap to the test case MintingCap
			msg.MintingCap = test.mintingCap
			// validate the message
			err := msg.ValidateBasic()
			// check if there was an error
			require.Error(t, err)
			// assert that the error is of type ErrInvalidMintingCap
			require.ErrorIs(t, err, ErrInvalidMintingCap)
			// assert that the error message matches the expected error message
			require.EqualError(t, err, test.errorMessage)
		})
	}
}

func TestMsgCreateDenom_ValidateBasic_InvalidURI(t *testing.T) {
	// Test case: invalid URI

	// make a copy of sample message
	msg := msgCreateDenomSample

	// set invalid uri and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.URI,
		&errorPacks.InvalidDenomURI,
		ErrInvalidMetadata,
	)
}

func TestMsgCreateDenom_ValidateBasic_InvalidURIHash(t *testing.T) {
	// Test case: invalid URI hash

	// make a copy of sample message
	msg := msgCreateDenomSample

	// set invalid uri hash and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.URIHash,
		&errorPacks.InvalidDenomURIHash,
		ErrInvalidMetadata,
	)
}

func TestMsgCreateDenom_ValidateBasic_InvalidGetTokenDenom(t *testing.T) {
	// Test case: invalid inputs to GetTokenDenom (creator or subdenom)

	// Valid base message for constructing test cases
	validCreator := sample.AccAddress()
	validSubDenom := "test"

	// Test cases for invalid creator or subdenom that cause GetTokenDenom to fail
	tests := []struct {
		name         string
		creator      string
		subDenom     string
		errorMessage string
	}{
		// Invalid subdenom cases
		{
			name:         "empty subdenom",
			creator:      validCreator,
			subDenom:     "",
			errorMessage: "Invalid subdenom name: denom name is empty e.g. uzig: invalid coins",
		},
		{
			name:         "subdenom too short",
			creator:      validCreator,
			subDenom:     "ab",
			errorMessage: "invalid coin: 'ab' denom name is too short, minimum 3 characters e.g. uzig: invalid coins",
		},
		{
			name:         "subdenom too long",
			creator:      validCreator,
			subDenom:     strings.Repeat("a", 45),
			errorMessage: "invalid coin: 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa' denom name is too long (45), maximum 44 characters e.g. uzig: invalid coins",
		},
		{
			name:         "subdenom invalid characters",
			creator:      validCreator,
			subDenom:     "bit@coin",
			errorMessage: "invalid coin: 'bit@coin' only lowercase letters (a-z) and numbers (0-9) are allowed e.g. uzig123: invalid coins",
		},
		// Invalid creator cases
		{
			name:         "creator too long",
			creator:      strings.Repeat("a", 76),
			subDenom:     validSubDenom,
			errorMessage: "SIGNER ADDRESS: 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa' (decoding bech32 failed: invalid separator index -1): invalid address",
		},
		{
			name:         "creator too short",
			creator:      "zig123",
			subDenom:     validSubDenom,
			errorMessage: "SIGNER ADDRESS: 'zig123' (decoding bech32 failed: invalid bech32 string length 6): invalid address",
		},
		{
			name:         "creator contains dot",
			creator:      "zig1vm3v4yrd3rrwkf3fe.qxutaz27098t76270qc5",
			subDenom:     validSubDenom,
			errorMessage: "SIGNER ADDRESS: 'zig1vm3v4yrd3rrwkf3fe.qxutaz27098t76270qc5' (decoding bech32 failed: invalid character not part of charset: 46): invalid address",
		},
		// Invalid full denom case
		{
			name:         "full denom too long",
			creator:      validCreator,
			subDenom:     strings.Repeat("a", 128-len("coin.")-len(validCreator)),
			errorMessage: "invalid coin: 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa' denom name is too long (81), maximum 44 characters e.g. uzig: invalid coins",
		},
	}

	// make a copy of sample message
	msg := msgCreateDenomSample

	// Iterate through each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a message with the test case inputs
			msg.Creator = tt.creator
			msg.SubDenom = tt.subDenom

			// Validate the message
			err := msg.ValidateBasic()

			// Check if there was an error
			require.Error(t, err)
			// Assert that the error message matches the expected error message
			require.EqualError(t, err, tt.errorMessage)
		})
	}
}

func TestMsgCreateDenom_ValidateBasic_InvalidFullDenomTooLong(t *testing.T) {
	// Test case: full denom exceeds maximum length of 128 characters

	// Valid creator address
	creator := sample.AccAddress() // length ~42
	// subdenom crafted to exceed 128 characters in full denom when combined
	maxAllowedDenomLength := 128
	prefixLength := len("coin.") + len(creator) + len(".") // ~len("coin.creator.")
	excessLength := maxAllowedDenomLength - prefixLength + 1
	longSubdenom := strings.Repeat("a", excessLength)

	msg := NewMsgCreateDenom(
		creator,
		longSubdenom,
		cosmosmath.NewUint(10000),
		true,
		"",
		"",
	)

	err := msg.ValidateBasic()
	require.Error(t, err)
	require.Contains(t, err.Error(), "denom name is too long (81), maximum 44 characters")
}
