package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"zigchain/zutils/validators"
)

var _ sdk.Msg = &MsgWithdrawModuleFees{}

func NewMsgWithdrawModuleFees(signer string, receiver string) *MsgWithdrawModuleFees {
	return &MsgWithdrawModuleFees{
		Signer:   signer,
		Receiver: receiver,
	}
}

func (msg *MsgWithdrawModuleFees) ValidateBasic() error {

	if err := validators.SignerCheck(msg.Signer); err != nil {
		return err
	}

	if msg.Receiver != "" {
		if err := validators.AddressCheck("Receiver", msg.Receiver); err != nil {
			return errorsmod.Wrapf(
				sdkerrors.ErrInvalidAddress,
				"Invalid receiver address: %s",
				msg.Receiver,
			)
		}
	}
	return nil
}
