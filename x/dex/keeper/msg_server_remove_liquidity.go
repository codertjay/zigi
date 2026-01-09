package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"zigchain/x/dex/events"
	"zigchain/x/dex/types"
)

func (k msgServer) RemoveLiquidity(goCtx context.Context, msg *types.MsgRemoveLiquidity) (*types.MsgRemoveLiquidityResponse, error) {

	// Retrieve and update liquidity pool balances

	ctx := sdk.UnwrapSDKContext(goCtx)

	signer, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidAddress,
			"Invalid address: %s",
			msg.Creator,
		)
	}

	// for readability
	poolIDString := msg.Lptoken.Denom

	pool, found := k.GetPool(ctx, poolIDString)
	if !found {
		return nil,
			errorsmod.Wrapf(types.ErrPoolNotFound,
				"pool %s not found",
				poolIDString,
			)
	}

	// Sanity check of pool structure
	if pool.LpToken.Denom != msg.Lptoken.Denom {
		return nil,
			errorsmod.Wrapf(types.ErrPoolNotFound,
				"pool lp token: %s different then incoming lp token: %s",
				pool.LpToken.Denom,
				msg.Lptoken.Denom,
			)
	}

	// Check if the user has enough liquidity pool tokens to remove
	if !k.bankKeeper.HasBalance(ctx, signer, msg.Lptoken) {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInsufficientFunds,
			"Wallet has insufficient funds %s, can not remove liquidity",
			msg.Lptoken.String(),
		)
	}

	if err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, signer, types.ModuleName, sdk.NewCoins(msg.Lptoken)); err != nil {
		return nil, errorsmod.Wrapf(
			err,
			"RemoveLiquidity: Failed to send %s coins, from signer %s to module %s",
			msg.Lptoken.String(),
			signer,
			types.ModuleName,
		)
	}

	// Calculate the amounts of base and quote tokens to return to the user
	coinsOut, err := CoinsToRemove(pool, msg.Lptoken)
	if err != nil {
		return nil, err
	}

	var receiver sdk.AccAddress
	if msg.Receiver != "" {
		receiver, err = sdk.AccAddressFromBech32(msg.Receiver)
		if err != nil {
			return nil, err
		}
	} else {
		receiver = signer
	}

	if err = k.SendFromPoolToAddress(
		ctx,
		poolIDString,
		receiver,
		coinsOut,
	); err != nil {
		return nil, errorsmod.Wrapf(
			err,
			"RemoveLiquidity: Failed to send coins %s and %s",
			coinsOut[0].String(),
			coinsOut[1].String(),
		)
	}

	// Burn liquidity pool shares and update user balances
	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(msg.Lptoken))
	if err != nil {
		return nil, errorsmod.Wrapf(
			types.ErrInsufficientBalance,
			"BurnCoins: Could not burn liquidity pool token: %s",
			msg.Lptoken.String(),
		)
	}

	pool.Coins[0].Amount = pool.Coins[0].Amount.Sub(coinsOut[0].Amount)
	pool.Coins[1].Amount = pool.Coins[1].Amount.Sub(coinsOut[1].Amount)
	pool.LpToken.Amount = pool.LpToken.Amount.Sub(msg.Lptoken.Amount)

	k.SetPool(ctx, pool)

	events.EmitRemoveLiquidityEvent(ctx, signer, &pool, &msg.Lptoken, &coinsOut, receiver)

	return &types.MsgRemoveLiquidityResponse{
		Base:  coinsOut[0],
		Quote: coinsOut[1],
	}, nil

}

// CoinsToRemove calculates the amount of base and quote tokens to return based on the removed shares.
func CoinsToRemove(pool types.Pool, lptoken sdk.Coin) (coins sdk.Coins, err error) {
	// sanity check
	if pool.LpToken.Amount.IsZero() {
		return nil, errorsmod.Wrapf(
			types.ErrInsufficientBalance,
			"Insufficient balance: %s pool lp token: %s",
			lptoken.String(),
			pool.LpToken.String(),
		)
	}

	if lptoken.Amount.GT(pool.LpToken.Amount) {
		return nil, errorsmod.Wrapf(
			types.ErrInsufficientBalance,
			"Insufficient balance: %s lptoken: %s",
			lptoken.String(),
			pool.LpToken.String(),
		)
	}

	// Calculate the amount of base and quote tokens to remove based on the adjusted share ratio
	denom1 := pool.Coins[0].Amount.Mul(lptoken.Amount).Quo(pool.LpToken.Amount)
	denom2 := pool.Coins[1].Amount.Mul(lptoken.Amount).Quo(pool.LpToken.Amount)

	// Calculate the amount of base and quote tokens to remove based on the adjusted share ratio
	coins = sdk.NewCoins(
		sdk.NewCoin(pool.Coins[0].Denom, denom1),
		sdk.NewCoin(pool.Coins[1].Denom, denom2),
	)

	return coins, nil
}
