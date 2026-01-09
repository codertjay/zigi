package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/zutils/validators"
)

var _ sdk.Msg = &MsgUpdateDenomMetadataAuth{}

func NewMsgUpdateDenomMetadataAuth(signer string, denom string, metadataAdmin string) *MsgUpdateDenomMetadataAuth {
	return &MsgUpdateDenomMetadataAuth{
		Signer:        signer,
		Denom:         denom,
		MetadataAdmin: metadataAdmin,
	}
}

func (msg *MsgUpdateDenomMetadataAuth) ValidateBasic() error {

	// check if the signer address is valid
	if err := validators.SignerCheck(msg.Signer); err != nil {
		return err
	}
	// check if the denom is valid
	_, _, err := DeconstructDenom(msg.Denom)
	if err != nil {
		return errorsmod.Wrapf(
			ErrInvalidDenom,
			"(%s) : %s",
			msg.Denom,
			err,
		)
	}
	// check if the metadata admin address is valid
	if err := validators.AddressCheck("MetadataAdmin", msg.MetadataAdmin); err != nil {
		return err
	}
	return nil
}
