package types

import (
	"testing"

	errorPacks "zigchain/testutil/data"
	types "zigchain/testutil/helpers"

	"zigchain/testutil/sample"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

var MsgBurnTokensSample = MsgBurnTokens{
	Signer: sample.AccAddress(),
	Token:  sample.Coin("abc", 100),
}

// Positive test cases

func TestMsgBurnTokens_NewMsgBurnTokens_Positive(t *testing.T) {
	// Test case: burn tokens with valid input data

	signer := sample.AccAddress()
	token := sample.Coin("abc", 100)

	// create a new MsgBurnTokens instance
	msg := NewMsgBurnTokens(signer, token)

	// validate the fields
	require.NotNil(t, msg, "expected the message to be non-nil")
	require.Equal(t, signer, msg.Signer, "expected the signer to match the input signer")
	require.Equal(t, token, msg.Token, "expected the token to match the input token")
}

func TestMsgBurnTokens_ValidateBasic_Positive(t *testing.T) {
	// Test case: validate basic properties of MsgBurnTokens

	// make a copy of sample message
	msg := MsgBurnTokensSample

	// validate the message
	err := msg.ValidateBasic()

	// assert that the error is nil
	require.NoError(t, err)
}

// Negative test cases

func TestMsgBurnTokens_ValidateBasic_InvalidSignerAddress(t *testing.T) {
	// Test case: invalid Signer (owner) address

	// make a copy of sample message
	msg := MsgBurnTokensSample

	// set invalid signer (owner) address and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,                             // msgToTest
		&msg.Signer,                      // fieldPtr
		&errorPacks.InvalidSignerAddress, // testsPtr
		sdkerrors.ErrInvalidAddress,
	)
}

func TestMsgBurnTokens_ValidateBasic_InvalidToken(t *testing.T) {
	// Test case: invalid token

	// make a copy of sample message
	msg := MsgBurnTokensSample

	// set invalid token and check for errors
	types.CheckMessageCoinFieldErrors(
		t,
		&msg,
		&msg.Token,
		&errorPacks.InvalidDenomZeroAmountNotOK,
		sdkerrors.ErrInvalidCoins,
	)
}
