package validators

import (
	"strings"

	"zigchain/zutils/constants"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// SignerCheck checks if the signer address is valid
//
// It performs the following checks:
// 1. Ensures the address is not empty.
// 2. Verifies that the address is a valid Bech32 address.
// 3. Checks if the address has the correct prefix as defined in constants.AddressPrefix.
//
// Parameters:
//   - field: A string representing the name of the field being validated. This is used in error messages.
//   - address: The address string to be validated.
//
// Returns:
//   - error: nil if the address is valid, otherwise returns an error with a description of the validation failure.

func SignerCheck(signer string) error {

	if signer == "" {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidAddress,
			"SIGNER ADDRESS: '%s' (empty address string is not allowed)", signer,
		)
	}

	_, err := sdk.AccAddressFromBech32(signer)
	if err != nil {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidAddress,
			"SIGNER ADDRESS: '%s' (%s)", signer, err,
		)
	}

	if !strings.HasPrefix(signer, constants.AddressPrefix) {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidAddress,
			"SIGNER ADDRESS: invalid prefix: expected '%s', got '%s'", constants.AddressPrefix, signer[:len(constants.AddressPrefix)],
		)
	}

	return nil
}

// AddressCheck validates the given address string.
//
// It performs the following checks:
// 1. Ensures the address is not empty.
// 2. Verifies that the address is a valid Bech32 address.
// 3. Checks if the address has the correct prefix as defined in constants.AddressPrefix.
//
// Parameters:
//   - field: A string representing the name of the field being validated. This is used in error messages.
//   - address: The address string to be validated.
//
// Returns:
//   - error: nil if the address is valid, otherwise returns an error with a description of the validation failure.
func AddressCheck(field string, address string) error {

	if address == "" {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidAddress,
			"%s address: cannot be empty",
			field,
		)
	}

	_, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidAddress,
			"%s address: '%s' (%s)",
			field,
			address,
			err,
		)
	}

	if !strings.HasPrefix(address, constants.AddressPrefix) {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidAddress,
			"%s address: '%s' has invalid prefix: expected '%s', got '%s'",
			field,
			address,
			constants.AddressPrefix,
			address[:len(constants.AddressPrefix)],
		)
	}

	return nil
}
