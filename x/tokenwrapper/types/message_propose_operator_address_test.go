package types

import (
	"testing"

	errorPacks "zigchain/testutil/data"
	types "zigchain/testutil/helpers"
	"zigchain/testutil/sample"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

var MsgProposeOperatorAddressSample = MsgProposeOperatorAddress{
	Signer:      sample.AccAddress(),
	NewOperator: sample.AccAddress(),
}

// Positive test cases

func TestMsgProposeOperatorAddress_NewMsgProposeOperatorAddress_Positive(t *testing.T) {
	// Test case: Valid input data

	signer := sample.AccAddress()
	newOperator := sample.AccAddress()

	// create a new MsgUpdateOperatorAddress instance
	msg := NewMsgProposeOperatorAddress(signer, newOperator)

	// validate the fields
	require.NotNil(t, msg, "expected the message to be non-nil")
	require.Equal(t, signer, msg.Signer, "expected the signer to match the input signer")
	require.Equal(t, newOperator, msg.NewOperator, "expected the new operator to match the input new operator")
}

func TestMsgProposeOperatorAddress_ValidateBasic_Positive(t *testing.T) {
	// Test case: Validate basic properties of MsgProposeOperatorAddress

	// make a copy of sample message
	msg := MsgProposeOperatorAddressSample

	// validate the message
	err := msg.ValidateBasic()

	// assert that there are no errors
	require.NoError(t, err)
}

// Negative test cases

func TestMsgProposeOperatorAddress_InvalidSignerAddress(t *testing.T) {
	// Test case: invalid Signer address

	// make a copy of sample message
	msg := MsgProposeOperatorAddressSample

	// set invalid signer address and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Signer,
		&errorPacks.InvalidSignerAddress,
		sdkerrors.ErrInvalidAddress,
	)
}

func TestMsgProposeOperatorAddress_InvalidNewOperatorAddress(t *testing.T) {
	// Test case: invalid NewOperator address

	// make a copy of sample message
	msg := MsgProposeOperatorAddressSample

	// set invalid signer address and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.NewOperator,
		&errorPacks.InvalidSignerAddress,
		sdkerrors.ErrInvalidAddress,
	)
}
