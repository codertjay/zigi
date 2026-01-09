package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/zutils/validators"
)

var _ sdk.Msg = &MsgSwapExactIn{}

func NewMsgSwapExactIn(signer string, incoming sdk.Coin, poolId string, receiver string, outgoingMin *sdk.Coin) *MsgSwapExactIn {
	return &MsgSwapExactIn{
		Signer:   signer,
		Incoming: incoming,
		PoolId:   poolId,
		Receiver: receiver,
		// Can be nil so we dereference
		OutgoingMin: outgoingMin,
	}
}

func (msg *MsgSwapExactIn) ValidateBasic() error {

	if err := validators.SignerCheck(msg.Signer); err != nil {
		return err
	}

	if err := validators.CoinCheck(msg.Incoming, false); err != nil {
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
	if msg.OutgoingMin != nil {
		if err := validators.CoinCheck(*msg.OutgoingMin, false); err != nil {
			return err
		}
	}

	return nil
}
