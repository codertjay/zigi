package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	errorPacks "zigchain/testutil/data"
	types "zigchain/testutil/helpers"
	"zigchain/testutil/sample"
	"zigchain/zutils/validators"

	"github.com/stretchr/testify/require"
)

var MsgUpdateDenomURISample = MsgUpdateDenomURI{
	Signer:  sample.AccAddress(),
	Denom:   "coin" + FactoryDenomDelimiterChar + sample.AccAddress() + FactoryDenomDelimiterChar + "abc",
	URI:     "https://example.com",
	URIHash: validators.SHA256HashOfURL("https://example.com"),
}

// Positive test cases

func TestMsgUpdateDenomURI_NewMsgUpdateDenomURI_Positive(t *testing.T) {
	// Test case: update denom URI with valid input data

	signer := sample.AccAddress()
	denom := "coin" + FactoryDenomDelimiterChar + sample.AccAddress() + FactoryDenomDelimiterChar + "abc"
	uri := "https://example.com"
	uriHash := validators.SHA256HashOfURL("https://example.com")

	// create a new MsgUpdateDenomURI instance
	msg := NewMsgUpdateDenomURI(signer, denom, uri, uriHash)

	// check if the message is created correctly
	require.Equal(t, signer, msg.Signer)
	require.Equal(t, denom, msg.Denom)
	require.Equal(t, uri, msg.URI)
	require.Equal(t, uriHash, msg.URIHash)
}

func TestMsgUpdateDenomURI_Positive(t *testing.T) {
	// Test case: validate basic properties of MsgUpdateDenomURI

	// make a copy of sample message
	msg := MsgUpdateDenomURISample

	// validate the message
	err := msg.ValidateBasic()

	// assert that the error is nil
	require.NoError(t, err)
}

// Negative test cases

func TestMsgUpdateDenomURI_InvalidSignerAddress(t *testing.T) {
	// Test case: invalid creator address

	// make a copy of sample message
	msg := MsgUpdateDenomURISample

	// set invalid creator address and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Signer,
		&errorPacks.InvalidSignerAddress,
		sdkerrors.ErrInvalidAddress,
	)
}

func TestMsgUpdateDenomURI_InvalidDenom(t *testing.T) {
	// Test case: set an invalid denom and check for errors

	// create a slice of subtests (define test cases with different invalid denoms)
	tests := []struct {
		// name/description of the test case
		name string
		// Denom that will be created on MsgCreateDenom
		denom string
		// expected error message for the test case
		errorMessage string
	}{
		{
			name:         "empty",
			denom:        "",
			errorMessage: "() : invalid denom: : Factory Denom name is not valid",
		},
		{
			name:  "not zig",
			denom: "coin" + FactoryDenomDelimiterChar + "cosmos193fxruxcm8y32njzt23c86en7hd8tajma79tt3" + FactoryDenomDelimiterChar + "abc",
			errorMessage: "(coin" + FactoryDenomDelimiterChar + "cosmos193fxruxcm8y32njzt23c86en7hd8tajma79tt3" + FactoryDenomDelimiterChar + "abc) : Invalid creator address " +
				"(Creator address: 'cosmos193fxruxcm8y32njzt23c86en7hd8tajma79tt3' " +
				"(decoding bech32 failed: invalid checksum (expected u5qefe got a79tt3)): invalid address): " +
				"Factory Denom name is not valid: Factory Denom name is not valid",
		},
		{
			name:  "not enough parts of denom",
			denom: "coin" + FactoryDenomDelimiterChar + "zig1lqhrr0kczjjl2e9jukzfjygp2jvtq3qvnjurh6",
			errorMessage: "(coin" + FactoryDenomDelimiterChar + "zig1lqhrr0kczjjl2e9jukzfjygp2jvtq3qvnjurh6) : " +
				"not enough parts of denom coin" + FactoryDenomDelimiterChar + "zig1lqhrr0kczjjl2e9jukzfjygp2jvtq3qvnjurh6: " +
				"Factory Denom name is not valid: Factory Denom name is not valid",
		},
		{
			name:  "too many parts of denom",
			denom: "coin" + FactoryDenomDelimiterChar + "zig193fxruxcm8y32njzt23c86en7hd8tajma79tt3" + FactoryDenomDelimiterChar + "text" + FactoryDenomDelimiterChar + "abc",
			errorMessage: "(coin" + FactoryDenomDelimiterChar + "zig193fxruxcm8y32njzt23c86en7hd8tajma79tt3" + FactoryDenomDelimiterChar + "text" + FactoryDenomDelimiterChar + "abc) : too many parts of denom " +
				"coin" + FactoryDenomDelimiterChar + "zig193fxruxcm8y32njzt23c86en7hd8tajma79tt3" + FactoryDenomDelimiterChar + "text" + FactoryDenomDelimiterChar + "abc: " +
				"Factory Denom name is not valid: Factory Denom name is not valid",
		},
		{
			name:  "not starting with coin",
			denom: "blah" + FactoryDenomDelimiterChar + "zig193fxruxcm8y32njzt23c86en7hd8tajma79tt3" + FactoryDenomDelimiterChar + "abc",
			errorMessage: "(blah" + FactoryDenomDelimiterChar + "zig193fxruxcm8y32njzt23c86en7hd8tajma79tt3" + FactoryDenomDelimiterChar + "abc) : denom prefix is incorrect. Is: blah" + FactoryDenomDelimiterChar + "  " +
				"Should be: coin: Factory Denom name is not valid: Factory Denom name is not valid",
		},
	}

	// make a copy of sample message
	msg := MsgUpdateDenomURISample

	// iterate through each test case
	for _, test := range tests {
		t.Run("MsgUpdateDenomURISample.denom "+test.name, func(t *testing.T) {
			// set the denom
			msg.Denom = test.denom
			// validate the message
			err := msg.ValidateBasic()
			// check if there was an error
			require.Error(t, err)
			// assert that the error is of type ErrInvalidDenom
			require.ErrorIs(t, err, ErrInvalidDenom)
			// assert that the error message matches the expected error message
			require.EqualError(t, err, test.errorMessage)
		})
	}
}

func TestMsgUpdateDenomURI_InvalidURI(t *testing.T) {
	// Test case: invalid URI

	// make a copy of sample message
	msg := MsgUpdateDenomURISample

	// set invalid URI and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.URI,
		&errorPacks.InvalidDenomURI,
		ErrInvalidMetadata,
	)
}

func TestMsgUpdateDenomURI_InvalidURIHash(t *testing.T) {
	// Test case: invalid URI hash

	// make a copy of sample message
	msg := MsgUpdateDenomURISample

	// set invalid URI hash and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.URIHash,
		&errorPacks.InvalidDenomURIHash,
		ErrInvalidMetadata,
	)
}
