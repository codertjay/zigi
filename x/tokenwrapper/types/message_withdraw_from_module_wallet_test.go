package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	errorPacks "zigchain/testutil/data"
	types "zigchain/testutil/helpers"
	"zigchain/testutil/sample"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

var MsgWithdrawFromModuleWalletSample = MsgWithdrawFromModuleWallet{
	Signer: sample.AccAddress(),
	Amount: sdk.NewCoins(sample.Coin("uzig", 1000)),
}

// Positive test cases

func TestMsgWithdrawFromModuleWallet_NewMsgWithdrawFromModuleWallet_Positive(t *testing.T) {
	// Test case: Valid input data

	signer := sample.AccAddress()
	amount := sdk.NewCoins(sample.Coin("uzig", 1000))

	// create a new MsgWithdrawFromModuleWallet instance
	msg := NewMsgWithdrawFromModuleWallet(signer, amount)

	// validate the fields
	require.NotNil(t, msg, "expected the message to be non-nil")
	require.Equal(t, signer, msg.Signer, "expected the signer to match the input signer")
	require.Equal(t, []sdk.Coin(amount), msg.Amount, "expected the amount to match the input amount")
}

func TestMsgWithdrawFromModuleWallet_ValidateBasic_Positive(t *testing.T) {
	// Test case: Validate basic properties of MsgWithdrawFromModuleWallet

	// make a copy of sample message
	msg := MsgWithdrawFromModuleWalletSample

	// validate the message
	err := msg.ValidateBasic()

	// assert that there are no errors
	require.NoError(t, err)
}

// Negative test cases

func TestMsgWithdrawFromModuleWallet_ValidateBasic_InvalidSignerAddress(t *testing.T) {
	// Test case: invalid Signer address

	// make a copy of sample message
	msg := MsgWithdrawFromModuleWalletSample

	// set invalid signer address and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Signer,
		&errorPacks.InvalidSignerAddress,
		sdkerrors.ErrInvalidAddress,
	)
}

func TestMsgWithdrawFromModuleWallet_InvalidAmount(t *testing.T) {
	// Test case: invalid amount

	// make a copy of sample message
	msg := MsgWithdrawFromModuleWalletSample

	// make a local copy of the amount
	amount := msg.Amount

	// set invalid token and check for errors
	types.CheckMessageCoinFieldErrors(
		t,
		&msg,
		&amount[0],
		&errorPacks.InvalidDenomZeroAmountNotOK,
		sdkerrors.ErrInvalidCoins,
	)
}
