package types

import (
	"strings"
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	errorPacks "zigchain/testutil/data"
	types "zigchain/testutil/helpers"
	"zigchain/testutil/sample"
)

var msgClaimDenomAdminSample = MsgClaimDenomAdmin{
	Signer: sample.AccAddress(),
	Denom:  "coin." + sample.AccAddress() + ".abc",
}

// Positive test cases

func TestMsgClaimDenomAdmin_NewMsgClaimDenomAdmin(t *testing.T) {
	// Test case: claim denom admin with valid input data

	signer := sample.AccAddress()
	denom := "coin." + sample.AccAddress() + ".abc"

	// create a new NewMsgClaimDenomAdmin instance
	msg := NewMsgClaimDenomAdmin(signer, denom)

	// check if the message is created correctly
	require.Equal(t, signer, msg.Signer)
	require.Equal(t, denom, msg.Denom)
}

func TestMsgClaimDenomAdmin_ValidateBasic_Positive(t *testing.T) {
	// Test case: validate basic properties of MsgProposeDenomAdmin

	// make a copy of the sample message
	msg := msgClaimDenomAdminSample

	// validate the message
	err := msg.ValidateBasic()

	// assert that the error is nil
	require.NoError(t, err)
}

// Negative test cases

func TestMsgClaimDenomAdmin_NewMsgClaimDenomAdmin_InvalidSignerAddress(t *testing.T) {
	// Test case: invalid signer addresses

	// make a copy of the sample message
	msg := msgClaimDenomAdminSample

	// set an invalid creator address and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Signer,
		&errorPacks.InvalidSignerAddress,
		sdkerrors.ErrInvalidAddress,
	)
}

func TestMsgClaimDenomAdmin_ValidateBasic_InvalidDenom(t *testing.T) {
	// Test case: set an invalid denom and check for errors

	validCreator := sample.AccAddress()

	// Test cases for invalid denoms
	tests := []struct {
		name         string
		denom        string
		errorMessage string
	}{
		{
			name:         "empty denom",
			denom:        "",
			errorMessage: "() : invalid denom: : Factory Denom name is not valid",
		},
		{
			name:         "invalid denom characters",
			denom:        "coin." + validCreator + ".bit#coin",
			errorMessage: "(coin." + validCreator + ".bit#coin) : invalid denom: coin." + validCreator + ".bit#coin: Factory Denom name is not valid",
		},
		{
			name:         "denom with spaces",
			denom:        "coin." + validCreator + ".bit coin",
			errorMessage: "(coin." + validCreator + ".bit coin) : invalid denom: coin." + validCreator + ".bit coin: Factory Denom name is not valid",
		},
		{
			name:         "incorrect prefix",
			denom:        "ibc." + validCreator + ".abc",
			errorMessage: "(ibc." + validCreator + ".abc) : denom prefix is incorrect. Is: ibc.  Should be: coin: Factory Denom name is not valid: Factory Denom name is not valid",
		},
		{
			name:         "not enough parts",
			denom:        "coin." + validCreator,
			errorMessage: "(coin." + validCreator + ") : not enough parts of denom coin." + validCreator + ": Factory Denom name is not valid: Factory Denom name is not valid",
		},
		{
			name:         "too many parts",
			denom:        "coin." + validCreator + ".abc.extra",
			errorMessage: "(coin." + validCreator + ".abc.extra) : too many parts of denom coin." + validCreator + ".abc.extra: Factory Denom name is not valid: Factory Denom name is not valid",
		},
		{
			name:         "invalid creator address",
			denom:        "coin.invalidaddr.abc",
			errorMessage: "(coin.invalidaddr.abc) : Invalid creator address (Creator address: 'invalidaddr' (decoding bech32 failed: invalid separator index -1): invalid address): Factory Denom name is not valid: Factory Denom name is not valid",
		},
		{
			name:         "empty creator",
			denom:        "coin..abc",
			errorMessage: "(coin..abc) : Invalid creator address (Creator address: cannot be empty: invalid address): Factory Denom name is not valid: Factory Denom name is not valid",
		},
		{
			name:         "empty subdenom",
			denom:        "coin." + validCreator + ".",
			errorMessage: "(coin." + validCreator + ".) : subdenom is empty: Factory Denom name is not valid: Factory Denom name is not valid",
		},
		{
			name:         "subdenom too short",
			denom:        "coin." + validCreator + ".ab",
			errorMessage: "(coin." + validCreator + ".ab) : invalid coin: 'ab' denom name is too short, minimum 3 characters e.g. uzig: invalid coins: Factory Denom name is not valid",
		},
		{
			name:         "subdenom too long",
			denom:        "coin." + validCreator + "." + strings.Repeat("a", 45),
			errorMessage: "(coin." + validCreator + "." + strings.Repeat("a", 45) + ") : invalid coin: 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa' denom name is too long (45), maximum 44 characters e.g. uzig: invalid coins: Factory Denom name is not valid",
		},
		{
			name:         "subdenom invalid characters",
			denom:        "coin." + validCreator + ".bit@coin",
			errorMessage: "(coin." + validCreator + ".bit@coin) : invalid denom: coin." + validCreator + ".bit@coin: Factory Denom name is not valid",
		},
		{
			name:         "denom too long",
			denom:        "coin." + validCreator + "." + strings.Repeat("a", 128-len("coin.")-len(validCreator)),
			errorMessage: "(coin." + validCreator + "." + strings.Repeat("a", 128-len("coin.")-len(validCreator)) + ") : invalid coin: 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa' denom name is too long (81), maximum 44 characters e.g. uzig: invalid coins: Factory Denom name is not valid",
		},
	}

	// Make a copy of the sample message
	msg := msgClaimDenomAdminSample

	// Iterate through each test case
	for _, tt := range tests {
		t.Run("msgClaimDenomAdmin.denom "+tt.name, func(t *testing.T) {
			// Set the denom
			msg.Denom = tt.denom
			// Validate the message
			err := msg.ValidateBasic()
			// Check if there was an error
			require.Error(t, err)
			// Assert that the error is of type ErrInvalidDenom
			require.ErrorIs(t, err, ErrInvalidDenom)
			// Assert that the error message matches the expected error message
			require.EqualError(t, err, tt.errorMessage)
		})
	}
}
