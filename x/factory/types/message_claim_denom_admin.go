package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/zutils/validators"
)

var _ sdk.Msg = &MsgClaimDenomAdmin{}

func NewMsgClaimDenomAdmin(
	signer string,
	denom string,
) *MsgClaimDenomAdmin {
	return &MsgClaimDenomAdmin{
		Signer: signer,
		Denom:  denom,
	}
}

func (msg *MsgClaimDenomAdmin) ValidateBasic() error {
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

	return nil
}
