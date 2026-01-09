package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"zigchain/testutil/sample"
)

var msgUpdateParamsSample = MsgUpdateParams{
	Authority: sample.AccAddress(),
	Params:    Params{},
}

// Positive test cases

func TestMsgUpdateParams_ValidateBasic_Positive(t *testing.T) {
	// Test case: valid MsgUpdateParams

	// make a copy of sample message
	msg := msgUpdateParamsSample

	// validate the message
	err := msg.ValidateBasic()

	// assert that the error is nil
	require.NoError(t, err)
}

// Negative test cases

func TestMsgUpdateParams_ValidateBasic_InvalidAuthority(t *testing.T) {
	// Test case: invalid authority address

	// make a copy of sample message
	msg := MsgUpdateParams{
		Authority: "invalid_address",
		Params:    Params{},
	}

	// validate the message
	err := msg.ValidateBasic()

	// assert that the error is not nil
	require.Error(t, err)

	// assert that the error message matches the expected error message
	require.Contains(t, err.Error(), "invalid authority address")
}
