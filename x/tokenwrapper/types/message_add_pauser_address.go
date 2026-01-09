package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/zutils/validators"
)

var _ sdk.Msg = &MsgAddPauserAddress{}

func NewMsgAddPauserAddress(signer string, newPauser string) *MsgAddPauserAddress {
	return &MsgAddPauserAddress{
		Signer:    signer,
		NewPauser: newPauser,
	}
}

func (msg *MsgAddPauserAddress) ValidateBasic() error {
	if err := validators.SignerCheck(msg.Signer); err != nil {
		return err
	}

	if err := validators.AddressCheck("new_pauser", msg.NewPauser); err != nil {
		return err
	}

	return nil
}
