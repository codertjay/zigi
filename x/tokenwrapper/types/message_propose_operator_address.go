package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/zutils/validators"
)

var _ sdk.Msg = &MsgProposeOperatorAddress{}

func NewMsgProposeOperatorAddress(signer string, newOperator string) *MsgProposeOperatorAddress {
	return &MsgProposeOperatorAddress{
		Signer:      signer,
		NewOperator: newOperator,
	}
}

func (msg *MsgProposeOperatorAddress) ValidateBasic() error {
	if err := validators.SignerCheck(msg.Signer); err != nil {
		return err
	}

	if err := validators.SignerCheck(msg.NewOperator); err != nil {
		return err
	}

	return nil
}
