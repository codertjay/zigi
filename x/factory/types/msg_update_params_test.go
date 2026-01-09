package types

import (
	"testing"

	"zigchain/zutils/constants"

	"github.com/stretchr/testify/require"

	"zigchain/testutil/sample"
)

var msgUpdateParamsSample = MsgUpdateParams{
	// use sample.AccAddress() to get a valid address
	Authority: sample.AccAddress(),
	Params: Params{
		CreateFeeDenom:  constants.BondDenom,
		CreateFeeAmount: uint32(3000),
	},
}

// Positive test cases

func TestMsgUpdateParams_ValidateBasic_Positive(t *testing.T) {
	// Test case: update params with valid input data

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
		Params: Params{
			CreateFeeDenom:  constants.BondDenom,
			CreateFeeAmount: uint32(3000),
		},
	}

	// validate the message
	err := msg.ValidateBasic()

	// assert that the error is not nil
	require.Error(t, err)

	// assert that the error message matches the expected error message
	require.Contains(t, err.Error(), "invalid authority address")
}

func TestMsgUpdateParams_ValidateBasic_InvalidCreateFeeDenom(t *testing.T) {
	// Test case: invalid CreateFeeDenom

	// make a copy of sample message
	msg := msgUpdateParamsSample
	msg.Params.CreateFeeDenom = "invalid_denom"

	// validate the message
	err := msg.ValidateBasic()

	// assert that the error is not nil
	require.Error(t, err)

	// assert that the error message matches the expected error message
	require.Contains(t, err.Error(), "invalid create fee denom parameter")
}
