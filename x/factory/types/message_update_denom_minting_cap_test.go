package types

import (
	"testing"

	cosmosmath "cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	errorPacks "zigchain/testutil/data"
	types "zigchain/testutil/helpers"
	"zigchain/testutil/sample"

	"github.com/stretchr/testify/require"
)

var SampleMsgUpdateDenomMintingCap = MsgUpdateDenomMintingCap{
	Signer:              sample.AccAddress(),
	Denom:               "coin" + FactoryDenomDelimiterChar + sample.AccAddress() + FactoryDenomDelimiterChar + "abc",
	MintingCap:          cosmosmath.NewUint(100),
	CanChangeMintingCap: true,
}

// Positive test cases

func TestMsgUpdateDenomMintingCap_NewMsgUpdateDenomMintingCap_Positive(t *testing.T) {
	// Test case: update denom max supply with valid input data

	signer := sample.AccAddress()
	denom := "coin" + FactoryDenomDelimiterChar + sample.AccAddress() + FactoryDenomDelimiterChar + "abc"
	mintingCap := cosmosmath.NewUint(100)

	// create a new MsgUpdateDenomMint instance
	msg := NewMsgUpdateDenomMintingCap(signer, denom, mintingCap, true)

	// validate the fields
	require.NotNil(t, msg, "expected the message to be non-nil")
	require.Equal(t, signer, msg.Signer, "expected the signer to match the input signer")
	require.Equal(t, denom, msg.Denom, "expected the denom to match the input denom")
	require.Equal(t, mintingCap, msg.MintingCap, "expected the mintingCap to match the input mintingCap")
	require.True(t, msg.CanChangeMintingCap, "expected CanChangeMintingCap to be true")
}

func TestMsgUpdateDenomMintingCap_Positive(t *testing.T) {
	// Test case: validate basic properties of MsgUpdateDenomMintingCap

	err := SampleMsgUpdateDenomMintingCap.ValidateBasic()
	require.NoError(t, err)
}

// Negative test cases

func TestMsgUpdateDenomMintingCap_InvalidSignerAddress(t *testing.T) {
	// Test case: set an invalid admin address and check for errors

	msg := SampleMsgUpdateDenomMintingCap

	// set the owner address to an invalid address and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Signer,
		&errorPacks.InvalidSignerAddress,
		sdkerrors.ErrInvalidAddress,
	)
}

func TestMsgUpdateDenomMintingCap_InvalidMintingCap(t *testing.T) {
	// Test case: set MintingCap to 0 and check for errors

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
	msg := SampleMsgUpdateDenomMintingCap

	// iterate through each test case
	for _, test := range tests {
		t.Run("SampleMsgUpdateDenomMintingCap.mintingCap "+test.name, func(t *testing.T) {
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

func TestMsgUpdateDenomMintingCap_InvalidDenom(t *testing.T) {
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
			errorMessage: "(coin.cosmos193fxruxcm8y32njzt23c86en7hd8tajma79tt3.abc) : Invalid creator address " +
				"(Creator address: 'cosmos193fxruxcm8y32njzt23c86en7hd8tajma79tt3' " +
				"(decoding bech32 failed: invalid checksum (expected u5qefe got a79tt3)): invalid address): " +
				"Factory Denom name is not valid: Factory Denom name is not valid",
		},
		{
			name:  "not enough parts of denom",
			denom: "coin" + FactoryDenomDelimiterChar + "zig1lqhrr0kczjjl2e9jukzfjygp2jvtq3qvnjurh6",
			errorMessage: "(coin.zig1lqhrr0kczjjl2e9jukzfjygp2jvtq3qvnjurh6) : " +
				"not enough parts of denom coin.zig1lqhrr0kczjjl2e9jukzfjygp2jvtq3qvnjurh6: " +
				"Factory Denom name is not valid: Factory Denom name is not valid",
		},
		{
			name:  "too many parts of denom",
			denom: "coin" + FactoryDenomDelimiterChar + "zig193fxruxcm8y32njzt23c86en7hd8tajma79tt3" + FactoryDenomDelimiterChar + "text" + FactoryDenomDelimiterChar + "abc",
			errorMessage: "(coin.zig193fxruxcm8y32njzt23c86en7hd8tajma79tt3.text.abc) : too many parts of denom " +
				"coin" + FactoryDenomDelimiterChar + "zig193fxruxcm8y32njzt23c86en7hd8tajma79tt3" + FactoryDenomDelimiterChar + "text" + FactoryDenomDelimiterChar + "abc: " +
				"Factory Denom name is not valid: Factory Denom name is not valid",
		},
		{
			name:  "not starting with coin",
			denom: "blah" + FactoryDenomDelimiterChar + "zig193fxruxcm8y32njzt23c86en7hd8tajma79tt3" + FactoryDenomDelimiterChar + "abc",
			errorMessage: "(blah" + FactoryDenomDelimiterChar + "zig193fxruxcm8y32njzt23c86en7hd8tajma79tt3" + FactoryDenomDelimiterChar + "abc) : denom prefix is incorrect. Is: blah.  " +
				"Should be: coin: Factory Denom name is not valid: Factory Denom name is not valid",
		},
	}

	// make a copy of sample message
	msg := SampleMsgUpdateDenomMintingCap

	// iterate through each test case
	for _, test := range tests {
		t.Run("SampleMsgUpdateDenomMintingCap.denom "+test.name, func(t *testing.T) {
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
