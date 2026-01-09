package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/zutils/constants"
	"zigchain/zutils/validators"
)

var _ sdk.Msg = &MsgUpdateDenomURI{}

func NewMsgUpdateDenomURI(signer string, denom string, uri string, urihash string) *MsgUpdateDenomURI {
	return &MsgUpdateDenomURI{
		Signer:  signer,
		Denom:   denom,
		URI:     uri,
		URIHash: urihash,
	}
}

func (msg *MsgUpdateDenomURI) ValidateBasic() error {
	if err := validators.SignerCheck(msg.Signer); err != nil {
		return err
	}

	_, _, err := DeconstructDenom(msg.Denom)
	if err != nil {
		return errorsmod.Wrapf(
			ErrInvalidDenom,
			"(%s) : %s",
			msg.Denom,
			err,
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
