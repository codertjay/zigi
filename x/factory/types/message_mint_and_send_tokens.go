package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/zutils/validators"
)

var _ sdk.Msg = &MsgMintAndSendTokens{}

func NewMsgMintAndSendTokens(signer string, token sdk.Coin, recipient string) *MsgMintAndSendTokens {
	return &MsgMintAndSendTokens{
		Signer:    signer,    // address of the owner
		Token:     token,     // denom name
		Recipient: recipient, // address of the recipient
	}
}

func (msg *MsgMintAndSendTokens) ValidateBasic() error {

	if err := validators.SignerCheck(msg.Signer); err != nil {
		return err
	}

	if err := validators.CheckCoinDenom(msg.Token); err != nil {
		return err
	}

	if err := validators.CheckCoinAmount(msg.Token, false); err != nil {
		return err
	}

	if err := validators.AddressCheck("Recipient", msg.Recipient); err != nil {
		return err

	}

	return nil
}
