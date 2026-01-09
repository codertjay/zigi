package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/zutils/validators"
)

var _ sdk.Msg = &MsgRemovePauserAddress{}

func NewMsgRemovePauserAddress(signer string, pauser string) *MsgRemovePauserAddress {
	return &MsgRemovePauserAddress{
		Signer: signer,
		Pauser: pauser,
	}
}

func (msg *MsgRemovePauserAddress) ValidateBasic() error {
	if err := validators.SignerCheck(msg.Signer); err != nil {
		return err
	}

	if err := validators.AddressCheck("PAUSER", msg.Pauser); err != nil {
		return err
	}

	return nil
}
