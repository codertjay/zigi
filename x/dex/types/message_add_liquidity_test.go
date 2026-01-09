package types

import (
	"testing"

	"cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	errorPacks "zigchain/testutil/data"
	types "zigchain/testutil/helpers"
	"zigchain/testutil/sample"
	"zigchain/zutils/constants"

	"github.com/stretchr/testify/require"
)

var MsgAddLiquiditySample = MsgAddLiquidity{
	Creator:  sample.AccAddress(),
	PoolId:   constants.PoolPrefix + "1",
	Base:     sample.Coin("abc", 10),
	Quote:    sample.Coin("usdt", 10),
	Receiver: sample.AccAddress(),
}

// Positive test cases

func TestMsgAddLiquidity_NewMsgAddLiquidity_Positive(t *testing.T) {
	// Test case: add liquidity with valid input data

	creator := sample.AccAddress()
	poolId := constants.PoolPrefix + "1"
	base := sample.Coin("abc", 10)
	quote := sample.Coin("usdt", 10)
	receiver := sample.AccAddress()

	// create NewMsgAddLiquidity instance
	msg := NewMsgAddLiquidity(creator, poolId, base, quote, receiver)

	// validate fields
	require.NotNil(t, msg, "expected the message to be non-nil")
	require.Equal(t, creator, msg.Creator, "expected the creator to match the input creator")
	require.Equal(t, poolId, msg.PoolId, "expected the pool id to match the input pool id")
	require.Equal(t, base, msg.Base, "expected the base coin to match the input base coin")
	require.Equal(t, quote, msg.Quote, "expected the quote coin to match the input quote coin")
	require.Equal(t, receiver, msg.Receiver, "expected the receiver to match the input receiver")
}

func TestMsgAddLiquidity_ValidateBasic_Positive(t *testing.T) {
	// Test case: validate basic properties of MsgAddLiquidity

	// make a copy of sample message
	msg := MsgAddLiquiditySample

	// validate message
	err := msg.ValidateBasic()

	// make sure there is NO error
	require.NoError(t, err)
}

// Negative test cases

func TestMsgAddLiquidity_ValidateBasic_InvalidSigner(t *testing.T) {
	// Test case: invalid signer address

	// make a copy of sample message
	msg := MsgAddLiquiditySample

	// set an invalid signer address and check for the error
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Creator,
		&errorPacks.InvalidSignerAddress,
		sdkerrors.ErrInvalidAddress,
	)
}

func TestMsgAddLiquidity_ValidateBasic_InvalidPoolId(t *testing.T) {
	// Test case: invalid pool id

	// make a copy of sample message
	msg := MsgAddLiquiditySample

	// set an invalid pool id and check for the error
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.PoolId,
		&errorPacks.InvalidPoolId,
		sdkerrors.ErrInvalidCoins,
	)
}

func TestMsgAddLiquidity_ValidateBasic_InvalidBase(t *testing.T) {
	// Test case: invalid base

	// make a copy of sample message
	msg := MsgAddLiquiditySample

	// set an invalid base and check for the error
	types.CheckMessageCoinFieldErrors(
		t,
		&msg,
		&msg.Base,
		&errorPacks.InvalidDenomZeroAmountOK,
		sdkerrors.ErrInvalidCoins,
	)
}

func TestMsgAddLiquidity_ValidateBasic_InvalidQuote(t *testing.T) {
	// Test case: invalid quote

	// make a copy of sample message
	msg := MsgAddLiquiditySample

	// set an invalid quote and check for the error
	types.CheckMessageCoinFieldErrors(
		t,
		&msg,
		&msg.Quote,
		&errorPacks.InvalidDenomZeroAmountOK,
		sdkerrors.ErrInvalidCoins,
	)
}

func TestMsgAddLiquidity_ValidateBasic_BothBaseAndQuoteZero(t *testing.T) {
	// Test case: try to add liquidity if both base and quote are zero

	// make a copy of sample message
	msg := MsgAddLiquiditySample

	// set both base and quote to zero
	msg.Base.Amount = math.NewInt(0)
	msg.Quote.Amount = math.NewInt(0)

	// validate message
	err := msg.ValidateBasic()

	// check for the error
	require.Error(t, err)
	require.ErrorIs(t, err, sdkerrors.ErrInvalidCoins)
	require.EqualError(t, err, "Both base and quote coins are not positive: base: 0abc, quote: 0usdt: invalid coins")
}

func TestMsgAddLiquidity_ValidateBasic_InvalidReceiver(t *testing.T) {
	// Test case: invalid receiver address

	// make a copy of sample message
	msg := MsgAddLiquiditySample

	// set an invalid receiver address and check for the error
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Receiver,
		&errorPacks.InvalidReceiverAddress,
		ErrorInvalidReceiver,
	)
}

func TestMsgAddLiquidity_ValidateBasic_SameDenom(t *testing.T) {
	// Test case: base and quote have the same denom

	msg := MsgAddLiquiditySample
	msg.Base = sample.Coin("abc", 10)
	msg.Quote = sample.Coin("abc", 20)

	err := msg.ValidateBasic()

	require.Error(t, err)
	require.ErrorContains(t, err, "Base and quote denom must be different")
}
