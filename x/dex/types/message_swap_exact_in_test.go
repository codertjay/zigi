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

var MsgSwapExactInSample = MsgSwapExactIn{
	Signer:   sample.AccAddress(),
	Incoming: sample.Coin("abc", 100),
	PoolId:   constants.PoolPrefix + "1",
	Receiver: sample.AccAddress(),
	OutgoingMin: &sdk.Coin{ // <- add a pointer to Coin
		Denom:  "usdt",
		Amount: math.NewInt(50),
	},
}

// Positive test cases

func TestMsgSwapExactIn_NewMsgSwap_Positive(t *testing.T) {
	// Test case: swap exact in coins with valid input data

	creator := sample.AccAddress()
	incoming := sample.Coin("abc", 100)
	poolId := constants.PoolPrefix + "1"
	receiver := sample.AccAddress()
	outgoingMin := &sdk.Coin{
		Denom:  "usdt",
		Amount: math.NewInt(50),
	}

	// create NewMsgSwap instance
	msg := NewMsgSwapExactIn(creator, incoming, poolId, receiver, outgoingMin)

	// validate fields
	require.NotNil(t, msg, "expected the message to be non-nil")
	require.Equal(t, creator, msg.Signer, "expected the creator to match the input creator")
	require.Equal(t, incoming, msg.Incoming, "expected the incoming coin to match the input incoming coin")
	require.Equal(t, poolId, msg.PoolId, "expected the pool id to match the input pool id")
	require.Equal(t, receiver, msg.Receiver, "expected the receiver to match the input receiver")
	require.Equal(t, outgoingMin, msg.OutgoingMin, "expected the outgoing min to match the input outgoing min")
}

func TestMsgSwapExactIn_ValidateBasic_Positive(t *testing.T) {
	// Test case: validate basic properties of MsgSwapExactIn

	// make a copy of sample message
	msg := MsgSwapExactInSample

	// validate message
	err := msg.ValidateBasic()

	// make sure there is NO error
	require.NoError(t, err)
}

// Negative test cases

func TestMsgSwapExactIn_ValidateBasic_InvalidSigner(t *testing.T) {
	// Test case: invalid signer address

	// make a copy of sample message
	msg := MsgSwapExactInSample

	// set invalid signer address and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Signer,
		&errorPacks.InvalidSignerAddress,
		sdkerrors.ErrInvalidAddress,
	)
}

func TestMsgSwapExactIn_ValidateBasic_InvalidIncoming(t *testing.T) {
	// Test case: invalid incoming coin

	// make a copy of sample message
	msg := MsgSwapExactInSample

	// set invalid incoming coin and check for errors
	types.CheckMessageCoinFieldErrors(
		t,
		&msg,
		&msg.Incoming,
		&errorPacks.InvalidDenomZeroAmountNotOK,
		sdkerrors.ErrInvalidCoins,
	)
}

func TestMsgSwapExactIn_ValidateBasic_InvalidPoolId(t *testing.T) {
	// Test case: invalid pool id

	// make a copy of sample message
	msg := MsgSwapExactInSample

	// set invalid pool id and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.PoolId,
		&errorPacks.InvalidPoolId,
		sdkerrors.ErrInvalidCoins,
	)
}

func TestMsgSwapExactIn_ValidateBasic_InvalidReceiver(t *testing.T) {
	// Test case: invalid receiver address

	// make a copy of sample message
	msg := MsgSwapExactInSample

	// set an invalid receiver address and check for the error
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Receiver,
		&errorPacks.InvalidReceiverAddress,
		ErrorInvalidReceiver,
	)
}

func TestMsgSwapExactIn_ValidateBasic_InvalidOutgoingMin(t *testing.T) {
	// Test case: invalid outgoing min

	// make a copy of sample message
	msg := MsgSwapExactInSample

	// set invalid outgoing min and check for errors
	types.CheckMessageCoinFieldErrors(
		t,
		&msg,
		msg.OutgoingMin,
		&errorPacks.InvalidDenomZeroAmountNotOK,
		sdkerrors.ErrInvalidCoins,
	)
}
