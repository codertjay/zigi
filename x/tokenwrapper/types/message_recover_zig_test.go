package types

import (
	"testing"

	errorPacks "zigchain/testutil/data"
	types "zigchain/testutil/helpers"
	"zigchain/testutil/sample"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

var SampleMsgRecoverZig = MsgRecoverZig{
	Signer:  sample.AccAddress(),
	Address: sample.AccAddress(),
}

// Positive test cases

func TestMsgRecoverZig_NewMsgRecoverZig_Positive(t *testing.T) {
	// Test case: create new MsgRecoverZig with valid input data

	signer := sample.AccAddress()
	newPauser := sample.AccAddress()

	// create a new MsgRecoverZig instance
	msg := NewMsgRecoverZig(signer, newPauser)

	// validate the fields
	require.NotNil(t, msg, "expected the message to be non-nil")
	require.Equal(t, signer, msg.Signer, "expected the signer to match the input signer")
	require.Equal(t, newPauser, msg.Address, "expected the address to match the input address")
}

func TestMsgRecoverZig_ValidateBasic_Positive(t *testing.T) {
	// Test case: validate basic properties of MsgRecoverZig

	err := SampleMsgRecoverZig.ValidateBasic()
	require.NoError(t, err)
}

func TestMsgRecoverZig_ValidateBasic_SameSignerAndNewPauser(t *testing.T) {
	// Test case: set signer and address to the same address (edge case)

	sameAddress := sample.AccAddress()
	msg := MsgRecoverZig{
		Signer:  sameAddress,
		Address: sameAddress,
	}

	// This should still be valid as there's no business logic preventing it
	err := msg.ValidateBasic()
	require.NoError(t, err)
}

// Negative test cases

func TestMsgRecoverZig_NewMsgRecoverZig_InvalidSignerAddress(t *testing.T) {
	// Test case: set an invalid signer address and check for errors

	msg := SampleMsgRecoverZig

	// set the signer address to an invalid address and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Signer,
		&errorPacks.InvalidSignerAddress,
		sdkerrors.ErrInvalidAddress,
	)
}

func TestMsgRecoverZig_NewMsgRecoverZig_InvalidAddress(t *testing.T) {
	// Test case: set an invalid address and check for errors

	msg := SampleMsgRecoverZig

	errorPack := errorPacks.InvalidAddressErrors("address")

	// set the address to an invalid address and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Address,
		&errorPack,
		sdkerrors.ErrInvalidAddress,
	)
}
