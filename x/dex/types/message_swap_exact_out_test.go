package types

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	errorPacks "zigchain/testutil/data"
	types "zigchain/testutil/helpers"
	"zigchain/testutil/sample"
	"zigchain/zutils/constants"

	"github.com/stretchr/testify/require"
)

var MsgSwapExactOutSample = MsgSwapExactOut{
	Signer:   sample.AccAddress(),
	Outgoing: sample.Coin("usdt", 100),
	PoolId:   constants.PoolPrefix + "1",
	Receiver: sample.AccAddress(),
	IncomingMax: &sdk.Coin{ // <- add a pointer to Coin
		Denom:  "abc",
		Amount: math.NewInt(50),
	},
}

// Positive test cases

func TestMsgSwapExactOut_NewMsgSwapExactOut_Positive(t *testing.T) {
	// Test case: swap exact out with valid input data

	creator := sample.AccAddress()
	outgoing := sample.Coin("usdt", 100)
	poolId := constants.PoolPrefix + "1"
	receiver := sample.AccAddress()
	incomingMax := &sdk.Coin{
		Denom:  "abc",
		Amount: math.NewInt(50),
	}

	// create NewMsgSwapExactOut instance
	msg := NewMsgSwapExactOut(creator, outgoing, poolId, receiver, incomingMax)

	// validate fields
	require.NotNil(t, msg, "expected the message to be non-nil")
	require.Equal(t, creator, msg.Signer, "expected the creator to match the input creator")
	require.Equal(t, outgoing, msg.Outgoing, "expected the incoming coin to match the input incoming coin")
	require.Equal(t, poolId, msg.PoolId, "expected the pool id to match the input pool id")
	require.Equal(t, receiver, msg.Receiver, "expected the receiver to match the input receiver")
	require.Equal(t, incomingMax, msg.IncomingMax, "expected the outgoing min to match the input outgoing min")
}

func TestMsgSwapExactOut_ValidateBasic_Positive(t *testing.T) {
	// Test case: validate basic properties of MsgSwapExactOut

	// make a copy of a sample message
	msg := MsgSwapExactOutSample

	// validate message
	err := msg.ValidateBasic()

	// make sure there is NO error
	require.NoError(t, err)
}

// Negative test cases

func TestMsgSwapExactOut_ValidateBasic_InvalidSigner(t *testing.T) {
	// Test case: invalid signer address

	// make a copy of a sample message
	msg := MsgSwapExactOutSample

	// set an invalid signer address and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Signer,
		&errorPacks.InvalidSignerAddress,
		sdkerrors.ErrInvalidAddress,
	)
}

func TestMsgSwapExactOut_ValidateBasic_InvalidOutgoing(t *testing.T) {
	// Test case: invalid outgoing coin

	// make a copy of a sample message
	msg := MsgSwapExactOutSample

	// set invalid incoming coin and check for errors
	types.CheckMessageCoinFieldErrors(
		t,
		&msg,
		&msg.Outgoing,
		&errorPacks.InvalidDenomZeroAmountNotOK,
		sdkerrors.ErrInvalidCoins,
	)
}

func TestMsgSwapExactOut_ValidateBasic_InvalidPoolId(t *testing.T) {
	// Test case: invalid pool id

	// make a copy of a sample message
	msg := MsgSwapExactOutSample

	// set invalid pool id and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.PoolId,
		&errorPacks.InvalidPoolId,
		sdkerrors.ErrInvalidCoins,
	)
}

func TestMsgSwapExactOut_ValidateBasic_InvalidReceiver(t *testing.T) {
	// Test case: invalid receiver address

	// make a copy of a sample message
	msg := MsgSwapExactOutSample

	// set an invalid receiver address and check for the error
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Receiver,
		&errorPacks.InvalidReceiverAddress,
		ErrorInvalidReceiver,
	)
}

func TestMsgSwapExactOut_ValidateBasic_InvalidIncomingMax(t *testing.T) {
	// Test case: invalid incoming max

	// make a copy of a sample message
	msg := MsgSwapExactOutSample

	// set invalid incoming max and check for errors
	types.CheckMessageCoinFieldErrors(
		t,
		&msg,
		msg.IncomingMax,
		&errorPacks.InvalidDenomZeroAmountNotOK,
		sdkerrors.ErrInvalidCoins,
	)
}
