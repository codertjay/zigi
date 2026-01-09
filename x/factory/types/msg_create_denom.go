package types

import (
	cosmosmath "cosmossdk.io/math"

	"zigchain/zutils/constants"
	"zigchain/zutils/validators"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Validate Interface for the message
var _ sdk.Msg = &MsgCreateDenom{}

func NewMsgCreateDenom(
	creator string,
	subDenom string,
	mintingCap cosmosmath.Uint,
	canChangeMintingCap bool,
	uri string,
	uriHash string,

) *MsgCreateDenom {
	return &MsgCreateDenom{
		Creator:             creator,
		SubDenom:            subDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: canChangeMintingCap,
		URI:                 uri,
		URIHash:             uriHash,
	}
}

func (msg *MsgCreateDenom) ValidateBasic() error {

	// check if the creator address is valid
	if err := validators.SignerCheck(msg.Creator); err != nil {
		return err
	}

	// check if the denom is valid
	if err := validators.CheckSubDenomString(msg.SubDenom); err != nil {
		return err
	}

	// builds full denom and checks if sizes are within limits
	_, err := GetTokenDenom(msg.Creator, msg.SubDenom)
	if err != nil {
		return err
	}

	// MintingCap must be greater than zero
	if !msg.MintingCap.GT(cosmosmath.ZeroUint()) {
		return errorsmod.Wrapf(
			ErrInvalidMintingCap,
			"Minting Cap %s must be greater than 0",
			msg.MintingCap,
		)
	}

	if msg.URI != "" {
		if !validators.StringLengthInRange(msg.URI, 15, constants.MaxURILength) {
			return errorsmod.Wrapf(
				ErrInvalidMetadata,
				"URI: %s length must be between 15 and %d characters",
				msg.URI,
				constants.MaxURILength,
			)
		}
		if !validators.IsURI(msg.URI) {
			return errorsmod.Wrap(
				ErrInvalidMetadata,
				"URI must be a valid URL",
			)
		}
		if msg.URIHash != "" {
			// Quick check for format
			if !validators.IsSHA256Hash(msg.URIHash) {
				return errorsmod.Wrapf(
					ErrInvalidMetadata,
					"URIHash: %s (len of %d) for URL: %s must be a valid SHA256 hash string of 64 alpha numeric characters",
					msg.URIHash,
					len(msg.URIHash),
					msg.URI,
				)
			}
		}
	}
	return nil
}
