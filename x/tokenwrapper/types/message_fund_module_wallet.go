package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/zutils/validators"
)

var _ sdk.Msg = &MsgFundModuleWallet{}

func NewMsgFundModuleWallet(signer string, amount sdk.Coins) *MsgFundModuleWallet {
	return &MsgFundModuleWallet{
		Signer: signer,
		Amount: amount,
	}
}

func (msg *MsgFundModuleWallet) ValidateBasic() error {

	if err := validators.SignerCheck(msg.Signer); err != nil {
		return err
	}

	for _, coin := range msg.Amount {
		if err := validators.CoinCheck(coin, false); err != nil {
			return err
		}
	}

	return nil
}
