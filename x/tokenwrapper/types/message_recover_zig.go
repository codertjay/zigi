package types

import (
	"zigchain/zutils/validators"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgRecoverZig{}

func NewMsgRecoverZig(signer string, address string) *MsgRecoverZig {
	return &MsgRecoverZig{
		Signer:  signer,
		Address: address,
	}
}

func (msg *MsgRecoverZig) ValidateBasic() error {
	if err := validators.SignerCheck(msg.Signer); err != nil {
		return err
	}

	if err := validators.AddressCheck("address", msg.Address); err != nil {
		return err
	}

	return nil
}
