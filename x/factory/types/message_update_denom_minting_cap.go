package types

import (
	errorsmod "cosmossdk.io/errors"
	cosmosmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/zutils/validators"
)

var _ sdk.Msg = &MsgUpdateDenomMintingCap{}

func NewMsgUpdateDenomMintingCap(signer string, denom string, mintingCap cosmosmath.Uint, canChangeMintingCap bool) *MsgUpdateDenomMintingCap {
	return &MsgUpdateDenomMintingCap{
		Signer:              signer,
		Denom:               denom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: canChangeMintingCap,
	}
}

func (msg *MsgUpdateDenomMintingCap) ValidateBasic() error {
	if err := validators.SignerCheck(msg.Signer); err != nil {
		return err
	}

	_, _, err := DeconstructDenom(msg.Denom)
	if err != nil {
		return errorsmod.Wrapf(
			ErrInvalidDenom,
			"(%s) : %s",
			msg.Denom,
			err,
		)
	}

	if !msg.MintingCap.GT(cosmosmath.ZeroUint()) {
		return errorsmod.Wrapf(
			ErrInvalidMintingCap,
			"Minting Cap %s must be greater than 0",
			msg.MintingCap,
		)
	}

	return nil
}
