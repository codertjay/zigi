package types

import (
	"testing"

	errorPacks "zigchain/testutil/data"
	types "zigchain/testutil/helpers"
	"zigchain/testutil/sample"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

var MsgEnableTokenWrapperSample = MsgEnableTokenWrapper{
	Signer: sample.AccAddress(),
}

// Positive test cases

func TestMsgEnableTokenWrapper_NewMsgEnableTokenWrapper_Positive(t *testing.T) {
	// Test case: Valid input data

	signer := sample.AccAddress()

	// create a new MsgEnableTokenWrapper instance
	msg := NewMsgEnableTokenWrapper(signer)

	// validate the fields
	require.NotNil(t, msg, "expected the message to be non-nil")
	require.Equal(t, signer, msg.Signer, "expected the signer to match the input signer")
}

func TestMsgEnableTokenWrapper_ValidateBasic_Positive(t *testing.T) {
	// Test case: Validate basic properties of MsgEnableTokenWrapper

	// make a copy of sample message
	msg := MsgEnableTokenWrapperSample

	// validate the message
	err := msg.ValidateBasic()

	// assert that there are no errors
	require.NoError(t, err)
}

// Negative test cases

func TestMsgEnableTokenWrapper_ValidateBasic_InvalidSignerAddress(t *testing.T) {
	// Test case: invalid Signer address

	// make a copy of sample message
	msg := MsgEnableTokenWrapperSample

	// set invalid signer address and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Signer,
		&errorPacks.InvalidSignerAddress,
		sdkerrors.ErrInvalidAddress,
	)
}
