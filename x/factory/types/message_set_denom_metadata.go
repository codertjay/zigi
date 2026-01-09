package types

import (
	"zigchain/zutils/constants"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"zigchain/zutils/validators"
)

var _ sdk.Msg = &MsgSetDenomMetadata{}

func NewMsgSetDenomMetadata(signer string, metadata banktypes.Metadata) *MsgSetDenomMetadata {
	return &MsgSetDenomMetadata{
		Signer:   signer,
		Metadata: metadata,
	}
}

// ValidateBasic performs basic validation checks on the MsgSetDenomMetadata message.
// It checks the signer's validity, the metadata description length, and validates the metadata itself.
//
// The function performs the following checks:
// 1. Validates the signer using the SignerCheck function.
// 2. If a description is provided, ensures its length is between 3 and 255 characters.
// 3. Validates the metadata using the Validate method from the bank module.
//
// Parameters:
//   - msg: The MsgSetDenomMetadata message to be validated.
//
// Returns:
//   - error: An error if any validation check fails, or nil if all checks pass.
func (msg MsgSetDenomMetadata) ValidateBasic() error {
	if err := validators.SignerCheck(msg.Signer); err != nil {
		return err
	}

	// Metadata is validated by bank module below,
	// we will only limit/validate the description length here
	if msg.Metadata.Description != "" {
		if !validators.StringLengthInRange(msg.Metadata.Description, 3, 255) {
			return errorsmod.Wrap(
				ErrInvalidMetadata,
				"Description length must be between 3 and 255",
			)
		}
	}

	if msg.Metadata.URI != "" {
		if !validators.StringLengthInRange(msg.Metadata.URI, 15, constants.MaxURILength) {
			return errorsmod.Wrapf(
				ErrInvalidMetadata,
				"URI: %s length must be between 15 and %d characters",
				msg.Metadata.URI,
				constants.MaxURILength,
			)
		}
		if !validators.IsURI(msg.Metadata.URI) {
			return errorsmod.Wrap(
				ErrInvalidMetadata,
				"URI must be a valid URL",
			)
		}
		if msg.Metadata.URIHash != "" {
			// Quick check for format
			if !validators.IsSHA256Hash(msg.Metadata.URIHash) {
				return errorsmod.Wrapf(
					ErrInvalidMetadata,
					"URIHash: %s (len of %d) for URL: %s must be a valid SHA256 hash string of 64 alpha numeric characters",
					msg.Metadata.URIHash,
					len(msg.Metadata.URIHash),
					msg.Metadata.URI,
				)
			}
		}
	}

	// Bank module validates the metadata
	if err := msg.Metadata.Validate(); err != nil {
		return errorsmod.Wrap(ErrInvalidMetadata, err.Error())
	}

	return nil
}
