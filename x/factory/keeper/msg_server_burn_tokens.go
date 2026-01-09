package keeper

import (
	"context"
	"fmt"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"zigchain/x/factory/events"
	"zigchain/x/factory/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) BurnTokens(
	goCtx context.Context,
	msg *types.MsgBurnTokens,
) (*types.MsgBurnTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return nil, err
	}

	_, _, notFactoryDenomErr := types.DeconstructDenom(msg.Token.Denom)

	if notFactoryDenomErr == nil {
		_, isFound := k.GetDenom(
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
	}

	// Check if the signer has enough balances to burn
	if !k.bankKeeper.HasBalance(ctx, signer, msg.Token) {

		bankerBalance := k.bankKeeper.GetBalance(ctx, signer, msg.Token.Denom)

		return nil, errorsmod.Wrapf(
			types.ErrInsufficientFunds,
			"Signer: %s does not have enough funds, burn amount: %s, existing balance: %s",
			msg.Signer,
			msg.Token,
			bankerBalance,
		)
	}

	// Moving tokens from signer to a module account so module can burn them in the next step
	if err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, signer, types.ModuleName, sdk.NewCoins(msg.Token)); err != nil {
		return nil, err
	}

	//k.Logger().Error(fmt.Sprintf("Burning Coins: %s", msg.Token))
	// // Burning the tokens
	if err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(msg.Token)); err != nil {
		return nil, err
	}

	// Emitting an event
	events.EmitDenomBurned(ctx, msg.Signer, msg.Token)

	return &types.MsgBurnTokensResponse{
		AmountBurned: msg.Token,
	}, nil
}
