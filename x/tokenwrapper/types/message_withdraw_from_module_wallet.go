package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/zutils/validators"
)

var _ sdk.Msg = &MsgFundModuleWallet{}

func NewMsgWithdrawFromModuleWallet(signer string, amount sdk.Coins) *MsgWithdrawFromModuleWallet {
	return &MsgWithdrawFromModuleWallet{
		Signer: signer,
		Amount: amount,
	}
}

func (msg *MsgWithdrawFromModuleWallet) ValidateBasic() error {

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
