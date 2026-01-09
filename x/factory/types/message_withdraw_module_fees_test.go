package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	errorPacks "zigchain/testutil/data"
	types "zigchain/testutil/helpers"
	"zigchain/testutil/sample"
)

var MsgWithdrawModuleFeesSample = MsgWithdrawModuleFees{
	Signer:   sample.AccAddress(),
	Receiver: sample.AccAddress(),
}

// Positive test case

func TestMsgWithdrawModuleFees_NewMsgWithdrawModuleFees_Positive(t *testing.T) {
	// Test case: withdraw module fees with valid input data

	signer := sample.AccAddress()
	receiver := ""

	// create a new MsgWithdrawModuleFees instance
	msg := NewMsgWithdrawModuleFees(signer, receiver)

	// validate fields
	require.NotNil(t, msg, "expected the message to be non-nil")
	require.Equal(t, signer, msg.Signer, "expected the creator to match the input creator")
	require.Equal(t, receiver, msg.Receiver, "expected the receiver to match the input receiver")
}

func TestMsgWithdrawModuleFees_ValidateBasic_Positive(t *testing.T) {
	// Test case: validate basic properties of MsgWithdrawModuleFees

	// make a copy of sample message
	msg := MsgWithdrawModuleFeesSample

	// validate message
	err := msg.ValidateBasic()

	// make sure there is NO error
	require.NoError(t, err)
}

// Negative test cases

func TestMsgWithdrawModuleFees_ValidateBasic_InvalidSigner(t *testing.T) {
	// Test case: invalid signer address

	// make a copy of sample message
	msg := MsgWithdrawModuleFeesSample

	// set an invalid signer address and check for the error
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Signer,
		&errorPacks.InvalidSignerAddress,
		sdkerrors.ErrInvalidAddress,
	)
}

func TestMsgWithdrawModuleFees_ValidateBasic_InvalidReceiver(t *testing.T) {
	// Test case: invalid receiver address

	// make a copy of sample message
	msg := MsgWithdrawModuleFeesSample

	// set an invalid receiver address and check for the error
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Receiver,
		&errorPacks.InvalidReceiverAddress,
		sdkerrors.ErrInvalidAddress,
	)
}
