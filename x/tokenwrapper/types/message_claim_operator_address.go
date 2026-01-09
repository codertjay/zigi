package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/zutils/validators"
)

var _ sdk.Msg = &MsgClaimOperatorAddress{}

func NewMsgClaimOperatorAddress(signer string) *MsgClaimOperatorAddress {
	return &MsgClaimOperatorAddress{
		Signer: signer,
	}
}

func (msg *MsgClaimOperatorAddress) ValidateBasic() error {
	if err := validators.SignerCheck(msg.Signer); err != nil {
		return err
	}

	return nil
}
