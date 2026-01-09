package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	errorPacks "zigchain/testutil/data"
	types "zigchain/testutil/helpers"
	"zigchain/testutil/sample"

	"github.com/stretchr/testify/require"
)

var SampleMsgRemovePauserAddress = MsgRemovePauserAddress{
	Signer: sample.AccAddress(),
	Pauser: sample.AccAddress(),
}

// Positive test cases

func TestMsgRemovePauserAddress_NewMsgRemovePauserAddress_Positive(t *testing.T) {
	// Test case: create new MsgRemovePauserAddress with valid input data

	signer := sample.AccAddress()
	pauser := sample.AccAddress()

	// create a new MsgRemovePauserAddress instance
	msg := NewMsgRemovePauserAddress(signer, pauser)

	// validate the fields
	require.NotNil(t, msg, "expected the message to be non-nil")
	require.Equal(t, signer, msg.Signer, "expected the signer to match the input signer")
	require.Equal(t, pauser, msg.Pauser, "expected the pauser to match the input pauser")
}

func TestMsgRemovePauserAddress_ValidateBasic_Positive(t *testing.T) {
	// Test case: validate basic properties of MsgRemovePauserAddress

	err := SampleMsgRemovePauserAddress.ValidateBasic()
	require.NoError(t, err)
}

func TestMsgRemovePauserAddress_ValidateBasic_SameSignerAndPauser(t *testing.T) {
	// Test case: set signer and new pauser to the same address (edge case)

	sameAddress := sample.AccAddress()
	msg := MsgRemovePauserAddress{
		Signer: sameAddress,
		Pauser: sameAddress,
	}

	// This should still be valid as there's no business logic preventing it
	err := msg.ValidateBasic()
	require.NoError(t, err)
}

// Negative test cases

func TestMsgRemovePauserAddress_NewMsgBurnTokens_InvalidSignerAddress(t *testing.T) {
	// Test case: set an invalid signer address and check for errors

	msg := SampleMsgRemovePauserAddress

	// set the signer address to an invalid address and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Signer,
		&errorPacks.InvalidSignerAddress,
		sdkerrors.ErrInvalidAddress,
	)
}

func TestMsgRemovePauserAddress_ValidateBasic_InvalidPauserAddress(t *testing.T) {
	// Test case: set an invalid pauser address and check for errors

	msg := SampleMsgRemovePauserAddress

	// auto generate invalid address errors with the field name "PAUSER" build in them
	errorPack := errorPacks.InvalidAddressErrors("PAUSER")

	// set the pauser address to an invalid address and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Pauser,
		&errorPack,
		sdkerrors.ErrInvalidAddress,
	)
}
