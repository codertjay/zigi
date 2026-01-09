package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	errorPacks "zigchain/testutil/data"
	types "zigchain/testutil/helpers"
	"zigchain/testutil/sample"

	"github.com/stretchr/testify/require"
)

var SampleMsgAddPauserAddress = MsgAddPauserAddress{
	Signer:    sample.AccAddress(),
	NewPauser: sample.AccAddress(),
}

// Positive test cases

func TestMsgAddPauserAddress_NewMsgAddPauserAddress_Positive(t *testing.T) {
	// Test case: create new MsgAddPauserAddress with valid input data

	signer := sample.AccAddress()
	newPauser := sample.AccAddress()

	// create a new MsgAddPauserAddress instance
	msg := NewMsgAddPauserAddress(signer, newPauser)

	// validate the fields
	require.NotNil(t, msg, "expected the message to be non-nil")
	require.Equal(t, signer, msg.Signer, "expected the signer to match the input signer")
	require.Equal(t, newPauser, msg.NewPauser, "expected the newPauser to match the input newPauser")
}

func TestMsgAddPauserAddress_ValidateBasic_Positive(t *testing.T) {
	// Test case: validate basic properties of MsgAddPauserAddress

	err := SampleMsgAddPauserAddress.ValidateBasic()
	require.NoError(t, err)
}

func TestMsgAddPauserAddress_ValidateBasic_SameSignerAndNewPauser(t *testing.T) {
	// Test case: set signer and new pauser to the same address (edge case)

	sameAddress := sample.AccAddress()
	msg := MsgAddPauserAddress{
		Signer:    sameAddress,
		NewPauser: sameAddress,
	}

	// This should still be valid as there's no business logic preventing it
	err := msg.ValidateBasic()
	require.NoError(t, err)
}

// Negative test cases

func TestMsgAddPauserAddress_NewMsgAddPauserAddress_InvalidSignerAddress(t *testing.T) {
	// Test case: set an invalid signer address and check for errors

	msg := SampleMsgAddPauserAddress

	// set the signer address to an invalid address and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Signer,
		&errorPacks.InvalidSignerAddress,
		sdkerrors.ErrInvalidAddress,
	)
}

func TestMsgAddPauserAddress_NewMsgAddPauserAddress_InvalidNewPauserAddress(t *testing.T) {
	// Test case: set an invalid new pauser address and check for errors

	msg := SampleMsgAddPauserAddress

	errorPack := errorPacks.InvalidAddressErrors("new_pauser")

	// set the new pauser address to an invalid address and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.NewPauser,
		&errorPack,
		sdkerrors.ErrInvalidAddress,
	)
}
