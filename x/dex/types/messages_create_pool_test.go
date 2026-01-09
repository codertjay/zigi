package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	errorPacks "zigchain/testutil/data"
	types "zigchain/testutil/helpers"
	"zigchain/testutil/sample"

	"github.com/stretchr/testify/require"
)

var MsgCreatePoolSample = MsgCreatePool{
	Creator:  sample.AccAddress(),
	Base:     sample.Coin("abc", 100),
	Quote:    sample.Coin("usdt", 100),
	Receiver: "",
}

// Positive test case

func TestMsgCreatePool_NewMsgCreatePool_Positive(t *testing.T) {
	// Test case: create a new pool with valid input data

	creator := sample.AccAddress()
	base := sample.Coin("abc", 100)
	quote := sample.Coin("usdt", 100)
	receiver := ""

	// create a new MsgCreatePool instance
	msg := NewMsgCreatePool(creator, base, quote, receiver)

	// validate fields
	require.NotNil(t, msg, "expected the message to be non-nil")
	require.Equal(t, creator, msg.Creator, "expected the creator to match the input creator")
	require.Equal(t, base, msg.Base, "expected the base coin to match the input base coin")
	require.Equal(t, quote, msg.Quote, "expected the quote coin to match the input quote coin")
	require.Equal(t, receiver, msg.Receiver, "expected the receiver to match the input receiver")
}

func TestMsgCreatePool_ValidateBasic_Positive(t *testing.T) {
	// Test case: validate basic properties of MsgCreatePool

	// make a copy of sample message
	msg := MsgCreatePoolSample

	// validate message
	err := msg.ValidateBasic()

	// make sure there is NO error
	require.NoError(t, err)
}

func TestMsgCreatePool_ValidateBasic_WithReceiver(t *testing.T) {
	// Test case: validate basic properties of MsgCreatePool with receiver address

	// make a copy of sample message
	msg := MsgCreatePoolSample

	// set a receiver address
	msg.Receiver = sample.AccAddress()

	// validate message
	err := msg.ValidateBasic()

	// make sure there is NO error
	require.NoError(t, err)
}

// Negative test cases

func TestMsgCreatePool_ValidateBasic_InvalidCreator(t *testing.T) {
	// Test case: invalid creator address

	// make a copy of sample message
	msg := MsgCreatePoolSample

	// set an invalid creator address and check for the error
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Creator,
		&errorPacks.InvalidSignerAddress,
		sdkerrors.ErrInvalidAddress,
	)

}

func TestMsgCreatePool_ValidateBasic_InvalidBase(t *testing.T) {
	// Test case: invalid base coin

	// make a copy of sample message
	msg := MsgCreatePoolSample

	// set an invalid base coin and check for the error
	types.CheckMessageCoinFieldErrors(
		t,
		&msg,
		&msg.Base,
		&errorPacks.InvalidDenomZeroAmountNotOK,
		sdkerrors.ErrInvalidCoins,
	)

}

func TestMsgCreatePool_ValidateBasic_InvalidQuote(t *testing.T) {
	// Test case: invalid quote coin

	// make a copy of sample message
	msg := MsgCreatePoolSample

	// set an invalid quote coin and check for the error
	types.CheckMessageCoinFieldErrors(
		t,
		&msg,
		&msg.Quote,
		&errorPacks.InvalidDenomZeroAmountNotOK,
		sdkerrors.ErrInvalidCoins,
	)
}

func TestMsgCreatePool_ValidateBasic_InvalidReceiver(t *testing.T) {
	// Test case: invalid receiver address

	// make a copy of sample message
	msg := MsgCreatePoolSample

	// set an invalid receiver address and check for the error
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Receiver,
		&errorPacks.InvalidReceiverAddress,
		ErrorInvalidReceiver,
	)
}

func TestMsgCreatePool_ValidateBasic_SameDenom(t *testing.T) {
	// Test case: base and quote coins have the same denom

	msg := MsgCreatePoolSample
	msg.Base = sample.Coin("abc", 100)
	msg.Quote = sample.Coin("abc", 200)

	err := msg.ValidateBasic()

	require.Error(t, err)
	require.ErrorContains(t, err, "Base and quote denom must be different")
}

func TestIsValidFormula(t *testing.T) {
	// Test cases for IsValidFormula function

	require.True(t, IsValidFormula("constant_product"), "constant_product should be valid")
	require.False(t, IsValidFormula("invalid_formula"), "invalid_formula should not be valid")
}
