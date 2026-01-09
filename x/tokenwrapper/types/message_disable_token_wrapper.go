package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/zutils/validators"
)

var _ sdk.Msg = &MsgFundModuleWallet{}

func NewMsgDisableTokenWrapper(signer string) *MsgDisableTokenWrapper {
	return &MsgDisableTokenWrapper{
		Signer: signer,
	}
}

func (msg *MsgDisableTokenWrapper) ValidateBasic() error {

	if err := validators.SignerCheck(msg.Signer); err != nil {
		return err
	}

	return nil
}
