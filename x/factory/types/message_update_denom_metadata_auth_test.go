package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	errorPacks "zigchain/testutil/data"
	types "zigchain/testutil/helpers"
	"zigchain/testutil/sample"
)

var msgUpdateDenomMetadataAuthSample = MsgUpdateDenomMetadataAuth{
	Signer:        sample.AccAddress(),
	Denom:         "coin" + FactoryDenomDelimiterChar + sample.AccAddress() + FactoryDenomDelimiterChar + "abc",
	MetadataAdmin: sample.AccAddress(),
}

// Positive test cases

func TestMsgUpdateDenomMetadataAuth_NewMsgUpdateDenomMetadataAuth_Positive(t *testing.T) {
	// Test case: update denom metadata auth with valid input data

	signer := sample.AccAddress()
	denom := "test"
	metadataAdmin := sample.AccAddress()

	// create a new MsgUpdateDenomMetadataAuth instance
	msg := NewMsgUpdateDenomMetadataAuth(signer, denom, metadataAdmin)

	// check if the message is created correctly
	require.Equal(t, signer, msg.Signer)
	require.Equal(t, denom, msg.Denom)
	require.Equal(t, metadataAdmin, msg.MetadataAdmin)
}

func TestMsgUpdateDenomMetadataAuth_Positive(t *testing.T) {
	// Test case: validate basic properties of MsgUpdateDenomMetadataAuth

	// make a copy of sample message
	msg := msgUpdateDenomMetadataAuthSample

	// validate the message
	err := msg.ValidateBasic()

	// assert that the error is nil
	require.NoError(t, err)
}

// Negative test cases

func TestMsgUpdateDenomMetadataAuth_InvalidSignerAddress(t *testing.T) {
	// Test case: invalid signer addresses

	// make a copy of sample message
	msg := msgUpdateDenomMetadataAuthSample

	// set invalid creator address and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Signer,
		&errorPacks.InvalidSignerAddress,
		sdkerrors.ErrInvalidAddress,
	)
}

func TestNewMsgUpdateDenomMetadataAuth_ValidateBasic_InvalidMetadataAdminAddress(t *testing.T) {
	// Test case: invalid recipient address

	// make a copy of sample message
	msg := msgUpdateDenomMetadataAuthSample

	// auto generate invalid address errors with the field name "Recipient" build in them
	errorPack := errorPacks.InvalidAddressErrors("MetadataAdmin")
	// set invalid recipient address and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.MetadataAdmin,
		&errorPack,
		sdkerrors.ErrInvalidAddress,
	)
}

func TestNewMsgUpdateDenomMetadataAuth_ValidateBasic_InvalidDenom(t *testing.T) {
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
				"(Creator address: 'cosmos193fxruxcm8y32njzt23c86en7hd8tajma79tt3' (decoding bech32 failed: invalid checksum " +
				"(expected u5qefe got a79tt3)): invalid address): Factory Denom name is not valid: Factory Denom name is not valid",
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
			errorMessage: "(coin" + FactoryDenomDelimiterChar + "zig193fxruxcm8y32njzt23c86en7hd8tajma79tt3" + FactoryDenomDelimiterChar + "text" + FactoryDenomDelimiterChar + "abc) : " +
				"too many parts of denom coin" + FactoryDenomDelimiterChar + "zig193fxruxcm8y32njzt23c86en7hd8tajma79tt3" + FactoryDenomDelimiterChar + "text" + FactoryDenomDelimiterChar + "abc: " +
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
	msg := msgUpdateDenomMetadataAuthSample

	// iterate through each test case
	for _, test := range tests {
		t.Run("msgUpdateDenomMetadataAuthSample.denom "+test.name, func(t *testing.T) {
			// set the denom
			msg.Denom = test.denom
			// validate the message
			err := msg.ValidateBasic()
			// check if there was an error
			require.Error(t, err)
			// assert that the error message matches the expected error message
			require.EqualError(t, err, test.errorMessage)
		})
	}
}
