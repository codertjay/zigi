package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/zutils/validators"
)

var _ sdk.Msg = &MsgBurnTokens{}

func NewMsgBurnTokens(signer string, token sdk.Coin) *MsgBurnTokens {
	return &MsgBurnTokens{
		Signer: signer,
		Token:  token,
	}
}

// ValidateBasic performs basic validation checks on the MsgBurnTokens message.
// It checks the validity of the signer, token denomination, and token amount.
//
// The function validates:
// - The signer's address format
// - The token's denomination
// - The token's amount (ensuring it's positive)
//
// Returns an error if any validation check fails, otherwise returns nil.
func (msg *MsgBurnTokens) ValidateBasic() error {
	if err := validators.SignerCheck(msg.Signer); err != nil {
		return err
	}

	if err := validators.CheckCoinDenom(msg.Token); err != nil {
		return err
	}

	if err := validators.CheckCoinAmount(msg.Token, false); err != nil {
		return err
	}

	return nil
}
