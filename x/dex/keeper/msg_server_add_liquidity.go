package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/ibc-go/v10/modules/core/errors"

	"zigchain/x/dex/events"
	"zigchain/x/dex/types"
)

func (k msgServer) AddLiquidity(
	goCtx context.Context,
	msg *types.MsgAddLiquidity,

) (*types.MsgAddLiquidityResponse, error) {

	ctx := sdk.UnwrapSDKContext(goCtx)

	// extract sender address
	signer, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidAddress,
			"Invalid address: %s",
			msg.Creator,
		)
	}

	// Update liquidity pool balances
	pool, found := k.GetPool(ctx, msg.PoolId)
	if !found {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"Liquidity pool (%s) can not be found",
			msg.PoolId,
		)
	}

	// this will sort the coins in the same way pool does,
	// so we can check first against 1st and snd against 2nd
	sortedCoins := sdk.NewCoins(msg.Base, msg.Quote)

	if len(sortedCoins) != 2 {
		return nil, errorsmod.Wrap(
			errors.ErrInvalidRequest,
			"invalid coins pair, must send two coins, at positive (non zero) value",
		)
	}

	if sortedCoins[0].Denom != pool.Coins[0].Denom || sortedCoins[1].Denom != pool.Coins[1].Denom {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidCoins,
			"invalid coins pair: %s/%s, does not match pool coins: %s/%s",
			sortedCoins[0].Denom,
			sortedCoins[1].Denom,
			pool.Coins[0].Denom,
			pool.Coins[1].Denom,
		)

	}

	if !k.bankKeeper.HasBalance(ctx, signer, msg.Base) {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInsufficientFunds,
			"Wallet has insufficient funds %s to add liquidity",
			msg.Base.String(),
		)
	}

	if !k.bankKeeper.HasBalance(ctx, signer, msg.Quote) {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInsufficientFunds,
			"Wallet has insufficient funds %s to add liquidity",
			msg.Quote.String(),
		)
	}

	// Get params
	params := k.GetParams(ctx)

	// Issue liquidity pool shares to user
	shares, actualBase, actualQuote, returnedCoins, err := CalculateLiquidityShares(pool, params, sortedCoins[0], sortedCoins[1])
	if err != nil {
		return nil, errorsmod.Wrap(
			err,
			"Failed to calculate liquidity shares",
		)
	}

	actualCoins := sdk.NewCoins(actualBase, actualQuote)

	if err := k.SendFromAddressToPool(ctx, signer, msg.PoolId, sortedCoins); err != nil {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInsufficientFunds,
			"AddLiquidity: Failed to send coins (%s)",
			actualCoins.String(),
		)
	}

	// Mint liquidity tokens we will send to the signer
	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(shares))
	if err != nil {
		return nil, errorsmod.Wrap(
			err,
			"Failed to mint liquidity shares",
		)
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

	// Send the minted liquidity tokens to the signer
	err = k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		receiver,
		sdk.NewCoins(shares),
	)
	if err != nil {
		return nil, errorsmod.Wrapf(
			err,
			"Failed to send liquidity shares: %s, from pool ID: %s to receiver: %s",
			sdk.NewCoins(shares),
			msg.PoolId,
			receiver.String(),
		)
	}

	if returnedCoins.IsAllPositive() {
		if err := k.SendFromPoolToAddress(ctx, msg.PoolId, receiver, returnedCoins); err != nil {
			return nil, errorsmod.Wrapf(
				sdkerrors.ErrInsufficientFunds,
				"AddLiquidity: Failed to send back 'returned coins' (%s) to receiver: %s",
				returnedCoins.String(),
				receiver.String(),
			)
		}
	}

	// Update pool balances
	pool.LpToken = pool.LpToken.AddAmount(shares.Amount)

	pool.Coins[0] = pool.Coins[0].AddAmount(actualBase.Amount)
	pool.Coins[1] = pool.Coins[1].AddAmount(actualQuote.Amount)

	k.SetPool(ctx, pool)

	events.EmitAddLiquidityEvent(ctx, signer, &pool, &actualCoins, &shares, &returnedCoins, receiver)

	// all good
	return &types.MsgAddLiquidityResponse{
		Lptoken:       shares,
		ReturnedCoins: returnedCoins,
		ActualBase:    actualBase,
		ActualQuote:   actualQuote,
	}, nil
}

// validateTokenRatio checks if the deposit ratio matches the pool ratio within a tolerance
func validateTokenRatio(pool types.Pool, params types.Params, base sdk.Coin, quote sdk.Coin) error {
	// Skip validation if pool is empty (first deposit)
	if pool.LpToken.Amount.IsZero() {
		return nil
	}
	// Get max slippage from params
	maxSlippage := types.ConvertFromBasisPointsToDecimal(params.MaxSlippage) // Convert basis points to decimal
	// Calculate current pool ratio
	poolRatio := pool.Coins[0].Amount.ToLegacyDec().Quo(pool.Coins[1].Amount.ToLegacyDec())
	// Calculate deposit ratio
	depositRatio := base.Amount.ToLegacyDec().Quo(quote.Amount.ToLegacyDec())
	// Calculate the difference between ratios
	ratioDiff := poolRatio.Sub(depositRatio).Abs()
	// Calculate the percentage difference
	percentageDiff := ratioDiff.Quo(poolRatio)
	// Check if the percentage difference is greater than the max slippage
	if percentageDiff.GT(maxSlippage) {
		return errorsmod.Wrapf(
			errors.ErrInvalidRequest,
			"deposit ratio (%s) does not match pool ratio (%s) within allowed slippage of %s",
			depositRatio.String(),
			poolRatio.String(),
			maxSlippage.String(),
		)
	}
	return nil
}

// CalculateLiquidityShares calculates the number of shares to mint based on the added liquidity, similar to Uniswap's model.
func CalculateLiquidityShares(pool types.Pool, params types.Params, base sdk.Coin, quote sdk.Coin) (sdk.Coin, sdk.Coin, sdk.Coin, sdk.Coins, error) {
	// Validate the pool structure to ensure it's usable
	if len(pool.Coins) != 2 {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, sdk.Coins{}, errorsmod.Wrap(errors.ErrInvalidRequest, "pool must have exactly two coins")
	}

	// Validate token ratios
	if err := validateTokenRatio(pool, params, base, quote); err != nil {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, sdk.Coins{}, err
	}

	// When the pool is empty, calculate initial LP tokens based on the product of incoming amounts.
	if pool.LpToken.Amount.IsZero() {
		if !base.Amount.IsPositive() || !quote.Amount.IsPositive() {
			return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, sdk.Coins{}, errorsmod.Wrap(errors.ErrInvalidRequest, "when pool is empty both amounts must be positive")
		}

		// Compute the product of the base and quote amounts
		product := base.Amount.Mul(quote.Amount)

		// Compute the integer square root deterministically
		lpAmount := integerSqrt(product)

		// lpAmount as sdk.Coin
		lpMinted := sdk.NewCoin(pool.LpToken.Denom, lpAmount)

		// Return the resulting LP token as a sdk.Coin
		return lpMinted, base, quote, sdk.NewCoins(), nil
	}

	// Extract the existing amounts of each token in the pool
	poolBase := pool.Coins[0].Amount
	poolQuote := pool.Coins[1].Amount

	// Ensure the pool amounts are valid
	if !poolBase.IsPositive() || !poolQuote.IsPositive() {
		return sdk.Coin{}, sdk.Coin{}, sdk.Coin{}, sdk.Coins{}, errorsmod.Wrap(errors.ErrInvalidRequest, "pool amounts must be positive")
	}

	// Calculate LP tokens using the standard formula
	lpMintedFromBase := base.Amount.Mul(pool.LpToken.Amount).Quo(poolBase)
	lpMintedFromQuote := quote.Amount.Mul(pool.LpToken.Amount).Quo(poolQuote)

	// Calculate current pool ratio
	poolRatio := poolBase.ToLegacyDec().Quo(poolQuote.ToLegacyDec())

	var lpMinted, actualBase, actualQuote sdk.Coin

	if lpMintedFromQuote.LT(lpMintedFromBase) {
		lpMinted = sdk.NewCoin(pool.LpToken.Denom, lpMintedFromQuote)
		actualQuote = quote
		actualBase = sdk.NewCoin(base.Denom, actualQuote.Amount.ToLegacyDec().Mul(poolRatio).TruncateInt())
	} else {
		// if lp minted from base is less than lp minted from quote, use lp minted from base
		lpMinted = sdk.NewCoin(pool.LpToken.Denom, lpMintedFromBase)
		actualBase = base
		actualQuote = sdk.NewCoin(quote.Denom, actualBase.Amount.ToLegacyDec().Quo(poolRatio).TruncateInt())
	}

	deltaBase := base.Sub(actualBase)
	deltaQuote := quote.Sub(actualQuote)

	return lpMinted, actualBase, actualQuote, sdk.NewCoins(deltaBase, deltaQuote), nil
}
