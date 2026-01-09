package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/zutils/validators"
)

var _ sdk.Msg = &MsgFundModuleWallet{}

func NewMsgEnableTokenWrapper(signer string) *MsgEnableTokenWrapper {
	return &MsgEnableTokenWrapper{
		Signer: signer,
	}
}

func (msg *MsgEnableTokenWrapper) ValidateBasic() error {

	if err := validators.SignerCheck(msg.Signer); err != nil {
		return err
	}

	return nil
}
