package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/zutils/validators"
)

var _ sdk.Msg = &MsgSwapExactOut{}

func NewMsgSwapExactOut(signer string, outgoing sdk.Coin, poolId string, receiver string, incomingMax *sdk.Coin) *MsgSwapExactOut {
	return &MsgSwapExactOut{
		Signer:   signer,
		Outgoing: outgoing,
		PoolId:   poolId,
		Receiver: receiver,
		// Can be nil so we dereference
		IncomingMax: incomingMax,
	}
}

func (msg *MsgSwapExactOut) ValidateBasic() error {

	if err := validators.SignerCheck(msg.Signer); err != nil {
		return err
	}

	if err := validators.CoinCheck(msg.Outgoing, false); err != nil {
		return err
	}

	if err := validators.CheckPoolId(msg.PoolId); err != nil {
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
	// Validate if different from nil
	if msg.IncomingMax != nil {
		if err := validators.CoinCheck(*msg.IncomingMax, false); err != nil {
			return err
		}
	}

	return nil
}
