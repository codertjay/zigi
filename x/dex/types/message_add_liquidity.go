package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"zigchain/zutils/validators"
)

var _ sdk.Msg = &MsgAddLiquidity{}

func NewMsgAddLiquidity(
	creator string,
	poolId string,
	base sdk.Coin,
	quote sdk.Coin,
	receiver string,
) *MsgAddLiquidity {
	return &MsgAddLiquidity{
		Creator:  creator,
		PoolId:   poolId,
		Base:     base,
		Quote:    quote,
		Receiver: receiver,
	}
}

func (msg *MsgAddLiquidity) ValidateBasic() error {

	if err := validators.SignerCheck(msg.Creator); err != nil {
		return err
	}

	if err := validators.CheckPoolId(msg.PoolId); err != nil {
		return err
	}

	if err := validators.CoinCheck(msg.Base, true); err != nil {
		return err
	}

	if err := validators.CoinCheck(msg.Quote, true); err != nil {
		return err
	}

	// if both base and quote are not positive, fail. we need at least one positive coin incoming
	if !msg.Base.IsPositive() && !msg.Quote.IsPositive() {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidCoins,
			"Both base and quote coins are not positive: base: %s, quote: %s",
			msg.Base.String(),
			msg.Quote.String(),
		)
	}

	// empty for return coins to signer
	if msg.Receiver != "" {
		// switch to validators.AddressCheck on the next release
		if err := validators.SignerCheck(msg.Receiver); err != nil {
			return errorsmod.Wrapf(
				ErrorInvalidReceiver,
				"Invalid receiver address: %s",
				msg.Receiver,
			)
		}
	}

	// Ensure base and quote denoms are not the same
	if msg.Base.Denom == msg.Quote.Denom {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidCoins,
			"Base and quote denom must be different: both are %s",
			msg.Base.Denom,
		)
	}
	return nil
}
