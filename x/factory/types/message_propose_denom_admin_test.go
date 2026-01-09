package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	errorPacks "zigchain/testutil/data"
	types "zigchain/testutil/helpers"
	"zigchain/testutil/sample"
)

var msgProposeDenomAdminSample = MsgProposeDenomAdmin{
	Signer:        sample.AccAddress(),
	Denom:         "coin" + FactoryDenomDelimiterChar + sample.AccAddress() + FactoryDenomDelimiterChar + "abc",
	BankAdmin:     sample.AccAddress(),
	MetadataAdmin: sample.AccAddress(),
}

// Positive test cases

func TestMsgProposeDenomAdmin_NewMsgProposeDenomAdmin_Positive(t *testing.T) {
	// Test case: propose denom admin with valid input data

	signer := sample.AccAddress()
	denom := "coin" + FactoryDenomDelimiterChar + sample.AccAddress() + FactoryDenomDelimiterChar + "abc"
	bankAdmin := sample.AccAddress()
	metadataAdmin := sample.AccAddress()

	// create a new MsgProposeDenomAdmin instance
	msg := NewMsgProposeDenomAdmin(signer, denom, bankAdmin, metadataAdmin)

	// check if the message is created correctly
	require.Equal(t, signer, msg.Signer)
	require.Equal(t, denom, msg.Denom)
	require.Equal(t, bankAdmin, msg.BankAdmin)
	require.Equal(t, metadataAdmin, msg.MetadataAdmin)
}

func TestMsgProposeDenomAdmin_Positive(t *testing.T) {
	// Test case: validate basic properties of MsgProposeDenomAdmin

	// make a copy of the sample message
	msg := msgProposeDenomAdminSample

	// validate the message
	err := msg.ValidateBasic()

	// assert that the error is nil
	require.NoError(t, err)
}

// Negative test cases

func TestMsgProposeDenomAdmin_InvalidSignerAddress(t *testing.T) {
	// Test case: invalid signer addresses

	// make a copy of the sample message
	msg := msgProposeDenomAdminSample

	// set an invalid creator address and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Signer,
		&errorPacks.InvalidSignerAddress,
		sdkerrors.ErrInvalidAddress,
	)
}

func TestMsgProposeDenomAdmin_InvalidDenom(t *testing.T) {
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
				"(Creator address: 'cosmos193fxruxcm8y32njzt23c86en7hd8tajma79tt3' (decoding bech32 failed: " +
				"invalid checksum (expected u5qefe got a79tt3)): invalid address): " +
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

	// make a copy of the sample message
	msg := msgProposeDenomAdminSample

	// iterate through each test case
	for _, test := range tests {
		t.Run("msgProposeDenomAdminSample.denom "+test.name, func(t *testing.T) {
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

func TestMsgProposeDenomAdmin_InvalidBankAdminAddress(t *testing.T) {
	// Test case: invalid bank admin addresses

	// make a copy of the sample message
	msg := msgProposeDenomAdminSample

	// auto generate invalid address errors with the field name "Recipient" build in them
	errorPack := errorPacks.InvalidAdminAddressErrors("Bank admin")
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.BankAdmin,
		&errorPack,
		ErrInvalidBankAdminAddress,
	)
}

func TestMsgProposeDenomAdmin_InvalidMetaAdminAddress(t *testing.T) {
	// Test case: invalid bank admin addresses

	// make a copy of the sample message
	msg := msgProposeDenomAdminSample

	// auto generate invalid address errors with the field name "Recipient" build in them
	errorPack := errorPacks.InvalidMetadataAdminAddressErrors("Metadata admin")
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.MetadataAdmin,
		&errorPack,
		ErrInvalidMetadataAdminAddress,
	)
}
