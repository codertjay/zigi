package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	errorPacks "zigchain/testutil/data"
	types "zigchain/testutil/helpers"
	"zigchain/testutil/sample"
)

var msgClaimOperatorAddressSample = MsgClaimOperatorAddress{
	Signer: sample.AccAddress(),
}

// Positive test cases

func TestMsgClaimOperatorAddress_NewMsgClaimOperatorAddress_Positive(t *testing.T) {
	// Test case: claim denom admin with valid input data

	signer := sample.AccAddress()

	// create a new MsgProposeDenomAdmin instance
	msg := NewMsgClaimOperatorAddress(signer)

	// check if the message is created correctly
	require.Equal(t, signer, msg.Signer)
}

func TestMsgClaimOperatorAddress_ValidateBasic_Positive(t *testing.T) {
	// Test case: validate basic properties of MsgClaimOperatorAddress

	// make a copy of the sample message
	msg := msgClaimOperatorAddressSample

	// validate the message
	err := msg.ValidateBasic()

	// assert that the error is nil
	require.NoError(t, err)
}

// Negative test cases

func TestMsgClaimOperatorAddress_ValidateBasic_InvalidSignerAddress(t *testing.T) {
	// Test case: invalid signer addresses

	// make a copy of the sample message
	msg := msgClaimOperatorAddressSample

	// set an invalid creator address and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Signer,
		&errorPacks.InvalidSignerAddress,
		sdkerrors.ErrInvalidAddress,
	)
}
