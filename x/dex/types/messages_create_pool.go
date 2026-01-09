package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"zigchain/zutils/validators"
)

var _ sdk.Msg = &MsgCreatePool{}

func NewMsgCreatePool(
	creator string,
	base sdk.Coin,
	quote sdk.Coin,
	receiver string,

) *MsgCreatePool {
	return &MsgCreatePool{
		Creator:  creator,
		Base:     base,
		Quote:    quote,
		Receiver: receiver,
	}
}

func (msg *MsgCreatePool) ValidateBasic() error {

	if err := validators.SignerCheck(msg.Creator); err != nil {
		return err
	}

	// we need some coins to create a pool
	if err := validators.CoinCheck(msg.Base, false); err != nil {
		return err
	}

	// on both sides
	if err := validators.CoinCheck(msg.Quote, false); err != nil {
		return err
	}

	// Ensure base and quote denoms are not the same
	if msg.Base.Denom == msg.Quote.Denom {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidCoins,
			"Base and quote denom must be different: both are %s",
			msg.Base.Denom,
		)
	}

	// Mostly used by rust cosmwasm contracts that act on behalf of the user
	// Receiver, in that case user can coin directly, so there is no need to perform additional pull tx
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
