package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"zigchain/x/factory/events"
	"zigchain/x/factory/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) MintAndSendTokens(goCtx context.Context, msg *types.MsgMintAndSendTokens) (*types.MsgMintAndSendTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the value exists
	currentDenom, isFound := k.GetDenom(
		ctx,
		msg.Token.Denom,
	)
	if !isFound {
		return nil,
			errorsmod.Wrap(
				sdkerrors.ErrKeyNotFound,
				fmt.Sprintf("Token: %s does not exist", msg.Token.Denom),
			)
	}

	// Checks if signer has permission to mint tokens
	if err := k.Auth(ctx, msg.Token.Denom, "bank", msg.Signer); err != nil {
		return nil, err
	}

	// calculate new supply
	totalMinted := currentDenom.Minted.Add(math.Uint(msg.Token.Amount))
	if totalMinted.GT(currentDenom.MintingCap) {
		return nil,
			errorsmod.Wrap(
				sdkerrors.ErrInvalidRequest,
				fmt.Sprintf(
					"Minting %s would exceed Minting Cap of %s",
					msg.Token.String(),
					currentDenom.MintingCap.String(),
				),
			)
	}

	recipientAddress, err := sdk.AccAddressFromBech32(msg.Recipient)
	if err != nil {
		return nil, err
	}

	mintCoins := sdk.NewCoins(msg.Token)

	if err = k.bankKeeper.MintCoins(ctx, types.ModuleName, mintCoins); err != nil {
		return nil, err
	}

	if err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipientAddress, mintCoins); err != nil {
		return nil, err
	}

	var denom = types.Denom{
		Creator:             currentDenom.Creator,
		Denom:               currentDenom.Denom,
		MintingCap:          currentDenom.MintingCap,
		Minted:              currentDenom.Minted.Add(math.Uint(msg.Token.Amount)),
		CanChangeMintingCap: currentDenom.CanChangeMintingCap,
	}

	k.SetDenom(
		ctx,
		denom,
	)
	minted := sdk.NewCoin(denom.Denom, math.Int(denom.Minted))

	// Get total supply from the bank keeper
	totalSupply := k.bankKeeper.GetSupply(ctx, denom.Denom)

	events.EmitDenomMintedAndSent(
		ctx,
		&denom,
		msg.Token,
		msg.Recipient,
		minted,
		totalSupply,
	)

	return &types.MsgMintAndSendTokensResponse{
		TokenMinted: &msg.Token,
		Recipient:   msg.Recipient,
		TotalMinted: &minted,
		TotalSupply: &totalSupply,
	}, nil
}
