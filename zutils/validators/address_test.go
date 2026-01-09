package validators_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"zigchain/testutil/sample"
	"zigchain/zutils/constants"
	"zigchain/zutils/validators"
)

// Positive test cases

func TestSignerCheckValid(t *testing.T) {
	// Test case: valid signer

	validSigner := sample.AccAddress()

	err := validators.SignerCheck(validSigner)
	require.NoError(t, err, "expected no error for valid signer")
}

func TestAddressCheckValid(t *testing.T) {
	// Test case: valid address

	validAddress := sample.AccAddress()
	field := "testField"

	err := validators.AddressCheck(field, validAddress)
	require.NoError(t, err, "expected no error for valid address")
}

// Negative test cases

func TestSignerCheckEmpty(t *testing.T) {
	// Test case: empty signer address

	emptySigner := ""

	err := validators.SignerCheck(emptySigner)
	require.Error(t, err, "expected an error for empty signer address")
	require.ErrorIs(t, err, sdkerrors.ErrInvalidAddress)
	require.Equal(
		t,
		fmt.Sprintf(
			"SIGNER ADDRESS: '%s' (empty address string is not allowed): invalid address",
			emptySigner,
		),
		err.Error(),
	)
}

func TestSignerCheckInvalid(t *testing.T) {
	// Test case: invalid signer

	invalidSigner := "invalidAddress123"

	err := validators.SignerCheck(invalidSigner)
	require.Error(t, err, "expected an error for invalid signer")
	require.ErrorIs(t, err, sdkerrors.ErrInvalidAddress)
	require.Equal(
		t,
		fmt.Sprintf(
			"SIGNER ADDRESS: '%s' (decoding bech32 failed: string not all lowercase or all uppercase): invalid address",
			invalidSigner,
		),
		err.Error(),
	)
}

func TestAddressCheckInvalid(t *testing.T) {
	// Test case: invalid address

	invalidAddress := "invalidAddress123"
	field := "testField"

	err := validators.AddressCheck(field, invalidAddress)
	require.Error(t, err, "expected an error for invalid address")
	require.ErrorIs(t, err, sdkerrors.ErrInvalidAddress)
	require.Equal(
		t,
		fmt.Sprintf(
			"%s address: '%s' (decoding bech32 failed: string not all lowercase or all uppercase): invalid address",
			field,
			invalidAddress,
		),
		err.Error(),
	)
}

func TestSignerCheckInvalidPrefix(t *testing.T) {
	// Test case: signer address with invalid prefix

	// Generate a valid Bech32 address with a different prefix
	accAddr := sdk.AccAddress([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14})
	invalidPrefixAddr, err := sdk.Bech32ifyAddressBytes("wro", accAddr)
	require.NoError(t, err, "expected no error when creating Bech32 address with wrong prefix")

	err = validators.SignerCheck(invalidPrefixAddr)
	require.Error(t, err, "expected an error for signer address with invalid prefix")
	require.ErrorIs(t, err, sdkerrors.ErrInvalidAddress)
	require.Equal(
		t,
		fmt.Sprintf(
			"SIGNER ADDRESS: '%s' (invalid Bech32 prefix; expected %s, got %s): invalid address",
			invalidPrefixAddr,
			constants.AddressPrefix,
			invalidPrefixAddr[:len(constants.AddressPrefix)],
		),
		err.Error(),
	)
}

func TestAddressCheckEmpty(t *testing.T) {
	// Test case: empty address

	field := "testField"
	emptyAddress := ""

	err := validators.AddressCheck(field, emptyAddress)
	require.Error(t, err, "expected an error for empty address")
	require.ErrorIs(t, err, sdkerrors.ErrInvalidAddress)
	require.Equal(
		t,
		fmt.Sprintf(
			"%s address: cannot be empty: invalid address",
			field,
		),
		err.Error(),
	)
}

func TestAddressCheckInvalidPrefix(t *testing.T) {
	// Test case: address with invalid prefix

	field := "testField"

	// Generate a valid Bech32 address with a different prefix
	accAddr := sdk.AccAddress([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14})
	invalidPrefixAddr, err := sdk.Bech32ifyAddressBytes("wro", accAddr)
	require.NoError(t, err, "expected no error when creating Bech32 address with wrong prefix")

	err = validators.AddressCheck(field, invalidPrefixAddr)
	require.Error(t, err, "expected an error for address with invalid prefix")
	require.ErrorIs(t, err, sdkerrors.ErrInvalidAddress)
	require.Equal(
		t,
		fmt.Sprintf(
			"%s address: '%s' (invalid Bech32 prefix; expected %s, got %s): invalid address",
			field,
			invalidPrefixAddr,
			constants.AddressPrefix,
			invalidPrefixAddr[:len(constants.AddressPrefix)],
		),
		err.Error(),
	)
}
