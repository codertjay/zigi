package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/zutils/validators"
)

var _ sdk.Msg = &MsgProposeDenomAdmin{}

func NewMsgProposeDenomAdmin(
	signer string,
	denom string,
	bankAdmin string,
	metadataAdmin string,
) *MsgProposeDenomAdmin {
	return &MsgProposeDenomAdmin{
		Signer:        signer,
		Denom:         denom,
		BankAdmin:     bankAdmin,
		MetadataAdmin: metadataAdmin,
	}
}

func (msg *MsgProposeDenomAdmin) ValidateBasic() error {
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

	if err = validators.AddressCheck("BankAdmin", msg.BankAdmin); err != nil {
		return errorsmod.Wrapf(
			ErrInvalidBankAdminAddress,
			"Bank admin address (%s)",
			msg.BankAdmin)
	}

	// Validate only if not empty, empty is allowed and will permanently lock the denom
	if msg.MetadataAdmin != "" {
		if err = validators.AddressCheck("MetadataAdmin", msg.MetadataAdmin); err != nil {
			return errorsmod.Wrapf(
				ErrInvalidMetadataAdminAddress,
				"Metadata admin address (%s)",
				msg.MetadataAdmin)
		}
	}

	return nil
}
