package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/zutils/validators"
)

var _ sdk.Msg = &MsgRemoveLiquidity{}

func NewMsgRemoveLiquidity(
	creator string,
	lptoken sdk.Coin,
	receiver string,
) *MsgRemoveLiquidity {
	return &MsgRemoveLiquidity{
		Creator:  creator,
		Lptoken:  lptoken,
		Receiver: receiver,
	}
}

func (msg *MsgRemoveLiquidity) ValidateBasic() error {

	if err := validators.SignerCheck(msg.Creator); err != nil {
		return err
	}

	if err := validators.CoinCheck(msg.Lptoken, false); err != nil {
		return err
	}

	if msg.Receiver != "" {
		if err := validators.SignerCheck(msg.Receiver); err != nil {
			return errorsmod.Wrapf(
				ErrorInvalidReceiver,
				"Invalid receiver address: %s",
				msg.Receiver,
			)
		}
	}
	return nil
}
