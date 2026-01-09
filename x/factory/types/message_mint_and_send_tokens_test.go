package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	errorPacks "zigchain/testutil/data"
	types "zigchain/testutil/helpers"
	"zigchain/testutil/sample"

	"github.com/stretchr/testify/require"
)

var MsgMintAndSendTokensSample = MsgMintAndSendTokens{
	Signer:    sample.AccAddress(),
	Token:     sample.Coin("abc", 100),
	Recipient: sample.AccAddress(),
}

// Positive test cases

func TestMsgMintAndSendTokens_NewMsgMintAndSendTokens_Positive(t *testing.T) {
	// Test case: mint and sent tokens with valid input data

	signer := sample.AccAddress()
	token := sample.Coin("abc", 100)
	recipient := sample.AccAddress()

	// create a new MsgMintAndSendTokens instance
	msg := NewMsgMintAndSendTokens(signer, token, recipient)

	// validate the fields
	require.NotNil(t, msg, "expected the message to be non-nil")
	require.Equal(t, signer, msg.Signer, "expected the signer to match the input signer")
	require.Equal(t, token, msg.Token, "expected the token to match the input token")
	require.Equal(t, recipient, msg.Recipient, "expected the recipient to match the input recipient")
}

func TestMsgMintAndSendTokens_ValidateBasic_Positive(t *testing.T) {
	// Test case: validate basic properties of MsgMintAndSendTokens

	// make a copy of sample message
	msg := MsgMintAndSendTokensSample

	// validate the message
	err := msg.ValidateBasic()

	// assert that the error is nil
	require.NoError(t, err)
}

// Negative test cases

func TestMsgMintAndSendTokens_ValidateBasic_InvalidSignerAddress(t *testing.T) {
	// Test case: invalid creator address

	// make a copy of sample message
	msg := MsgMintAndSendTokensSample

	// set invalid creator address and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Signer,
		&errorPacks.InvalidSignerAddress,
		sdkerrors.ErrInvalidAddress,
	)
}

func TestMsgMintAndSendTokens_ValidateBasic_InvalidRecipientAddress(t *testing.T) {
	// Test case: invalid recipient address

	// make a copy of sample message
	msg := MsgMintAndSendTokensSample

	// auto generate invalid address errors with the field name "Recipient" build in them
	errorPack := errorPacks.InvalidAddressErrors("Recipient")
	// set invalid recipient address and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Recipient,
		&errorPack,
		sdkerrors.ErrInvalidAddress,
	)
}

func TestMsgMintAndSendTokens_ValidateBasic_InvalidDenomZeroAmountNotOK(t *testing.T) {
	// Test case: invalid token

	// make a copy of sample message
	msg := MsgMintAndSendTokensSample

	// set invalid token and check for errors
	types.CheckMessageCoinFieldErrors(
		t,
		&msg,
		&msg.Token,
		&errorPacks.InvalidDenomZeroAmountNotOK,
		sdkerrors.ErrInvalidCoins,
	)
}
