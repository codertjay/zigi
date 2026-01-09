package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	errorPacks "zigchain/testutil/data"
	types "zigchain/testutil/helpers"
	"zigchain/testutil/sample"
)

var MsgRemoveLiquiditySample = MsgRemoveLiquidity{
	Creator:  sample.AccAddress(),
	Lptoken:  sample.Coin("lptoken", 10),
	Receiver: "",
}

// Positive test cases

func TestMsgRemoveLiquidity_NewMsgRemoveLiquidity_Positive(t *testing.T) {
	// Test case: remove liquidity with valid input data

	creator := sample.AccAddress()
	lptoken := sample.Coin("lptoken", 10)
	receiver := ""

	// create NewMsgRemoveLiquidity instance
	msg := NewMsgRemoveLiquidity(creator, lptoken, receiver)

	// validate fields
	require.NotNil(t, msg, "expected the message to be non-nil")
	require.Equal(t, creator, msg.Creator, "expected the creator to match the input creator")
	require.Equal(t, lptoken, msg.Lptoken, "expected the lptoken to match the input lptoken")
	require.Equal(t, receiver, msg.Receiver, "expected the receiver to match the input receiver")
}

func TestMsgRemoveLiquidity_ValidateBasic_Positive(t *testing.T) {
	// Test case: validate basic properties of MsgRemoveLiquidity

	// make a copy of sample message
	msg := MsgRemoveLiquiditySample

	// validate message
	err := msg.ValidateBasic()

	// make sure there is NO error
	require.NoError(t, err)
}

func TestMsgRemoveLiquidity_ValidateBasic_WithReceiver(t *testing.T) {
	// Test case: validate basic properties of MsgRemoveLiquidity with a receiver address

	// make a copy of sample message
	msg := MsgRemoveLiquiditySample

	// set a receiver address
	msg.Receiver = sample.AccAddress()

	// validate message
	err := msg.ValidateBasic()

	// make sure there is NO error
	require.NoError(t, err)
}

// Negative test cases

func TestMsgRemoveLiquidity_ValidateBasic_InvalidSigner(t *testing.T) {
	// Test case: invalid signer address

	// make a copy of sample message
	msg := MsgRemoveLiquiditySample

	// set invalid signer address and check for error
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Creator,
		&errorPacks.InvalidSignerAddress,
		sdkerrors.ErrInvalidAddress,
	)
}

func TestMsgRemoveLiquidity_ValidateBasic_InvalidLptoken(t *testing.T) {
	// Test case: invalid lptoken

	// make a copy of sample message
	msg := MsgRemoveLiquiditySample

	// set invalid lptoken and check for error
	types.CheckMessageCoinFieldErrors(
		t,
		&msg,
		&msg.Lptoken,
		&errorPacks.InvalidDenomZeroAmountNotOK,
		sdkerrors.ErrInvalidCoins,
	)
}

func TestMsgRemoveLiquidity_ValidateBasic_InvalidReceiver(t *testing.T) {
	// Test case: invalid receiver address

	// make a copy of sample message
	msg := MsgRemoveLiquiditySample

	// set an invalid receiver address and check for the error
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Receiver,
		&errorPacks.InvalidReceiverAddress,
		ErrorInvalidReceiver,
	)
}
