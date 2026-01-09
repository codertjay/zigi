package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	"zigchain/x/factory/events"
	"zigchain/x/factory/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) UpdateDenomMintingCap(goCtx context.Context, msg *types.MsgUpdateDenomMintingCap) (*types.MsgUpdateDenomMintingCapResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the value exists
	currentDenom, isFound := k.GetDenom(
		ctx,
		msg.Denom,
	)
	if !isFound {
		return nil, errorsmod.Wrapf(
			types.ErrDenomDoesNotExist,
			"Denom: %s",
			msg.Denom,
		)
	}

	if !currentDenom.CanChangeMintingCap {
		return nil, errorsmod.Wrapf(
			types.ErrCannotChangeMintingCap,
			"Denom: %s locked",
			msg.Denom,
		)
	}

	if currentDenom.Minted.GT(msg.MintingCap) {
		return nil, errorsmod.Wrapf(
			types.ErrInvalidMintingCap,
			"Denom: %s, current supply: %s bigger then new minting cap: %s, can't decrease minting cap lower then current supply",
			msg.Denom,
			currentDenom.Minted,
			msg.MintingCap,
		)
	}

	if err := k.Auth(ctx, msg.Denom, "bank", msg.Signer); err != nil {
		return nil, err
	}

	var newDenom = types.Denom{
		Creator:             currentDenom.Creator,
		Denom:               msg.Denom,
		MintingCap:          msg.MintingCap,
		Minted:              currentDenom.Minted,
		CanChangeMintingCap: msg.CanChangeMintingCap,
	}

	k.SetDenom(
		ctx,
		newDenom,
	)

	events.EmitMintingCapChanged(ctx, msg.Signer, &newDenom)

	return &types.MsgUpdateDenomMintingCapResponse{
		Denom:               newDenom.Denom,
		MintingCap:          newDenom.MintingCap,
		CanChangeMintingCap: newDenom.CanChangeMintingCap,
	}, nil
}
