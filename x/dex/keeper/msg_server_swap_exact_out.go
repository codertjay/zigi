package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"zigchain/x/dex/events"
	"zigchain/x/dex/types"
	"zigchain/zutils/constants"
)

// MinSwapFee is defined in msg_server_swap_exact_in.go and should be imported from there if needed

func (k msgServer) SwapExactOut(goCtx context.Context, msg *types.MsgSwapExactOut) (*types.MsgSwapExactOutResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !msg.Outgoing.IsPositive() {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"Outgoing amount must be positive (%s%s)",
			msg.Outgoing.Amount.String(),
			msg.Outgoing.Denom,
		)
	}

	poolIDString := msg.PoolId
	pool, found := k.GetPool(ctx, poolIDString)
	if !found {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"Liquidity pool (%s) can not be found",
			poolIDString,
		)
	}

	// sanity check
	for _, coin := range pool.Coins {
		if !coin.IsPositive() {
			return nil,
				errorsmod.Wrapf(
					sdkerrors.ErrInvalidRequest,
					"Invalid liquidity pool: %s token: %s balance: %s, has to be positive.",
					pool.PoolId,
					coin.Denom,
					coin.Amount.String(),
				)
		}
	}

	var fromCoin, toCoin int32

	switch msg.Outgoing.Denom {
	case pool.Coins[0].Denom:
		fromCoin, toCoin = 1, 0
	case pool.Coins[1].Denom:
		fromCoin, toCoin = 0, 1
	default:
		// Invalid incoming coin
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"Invalid outgoing coin (%s), this pool only supports base (%s) and quote (%s) tokens",
			msg.Outgoing.Denom,
			pool.Coins[0].Denom,
			pool.Coins[1].Denom,
		)
	}

	inCoin, fee, err := CalculateSwapExactOutAmount(&pool, msg.Outgoing)
	if err != nil {
		return nil, err
	}

	// Check if the swap amount is less than the maximum incoming amount
	if msg.IncomingMax != nil {

		// validates that the incoming maximum amount is a positive number
		if !msg.IncomingMax.IsPositive() {
			return nil, errorsmod.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"Incoming maximum amount must be positive (%s)",
				msg.IncomingMax.String(),
			)
		}
		// validates that the denom is the correct one
		if msg.IncomingMax.Denom != inCoin.Denom {
			return nil, errorsmod.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"Incoming maximum denom (%s) must be the same as incoming denom (%s)",
				msg.IncomingMax.Denom,
				inCoin.Denom,
			)
		}

		// validates that the amount below the maximum
		if inCoin.Amount.GT(msg.IncomingMax.Amount) {
			return nil, errorsmod.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"Required amount (%s) is more than incoming maximum amount (%s)",
				inCoin.String(),
				msg.IncomingMax.String(),
			)
		}
	}

	// Update liquidity pool balances
	pool.Coins[fromCoin] = pool.Coins[fromCoin].AddAmount(inCoin.Amount)
	pool.Coins[toCoin] = pool.Coins[toCoin].SubAmount(msg.Outgoing.Amount)
	k.SetPool(ctx, pool)

	// Update user balances
	sender, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return nil, err
	}

	// Deduct base tokens from sender
	err = k.SendFromAddressToPool(
		ctx,
		sender,
		poolIDString,
		sdk.NewCoins(inCoin))
	if err != nil {
		return nil,
			errorsmod.Wrapf(
				err,
				"SwapExactOut: Failed to send %s coins",
				inCoin.String(),
			)
	}

	var receiver sdk.AccAddress
	if msg.Receiver != "" {
		receiver, err = sdk.AccAddressFromBech32(msg.Receiver)
		if err != nil {
			return nil, err
		}
	} else {
		receiver = sender
	}

	// Credit quote tokens to sender
	err = k.SendFromPoolToAddress(
		ctx,
		poolIDString,
		receiver,
		sdk.NewCoins(msg.Outgoing))
	if err != nil {
		return nil,
			errorsmod.Wrapf(
				err,
				"SwapExactOut: Failed to send coins %s",
				msg.Outgoing.String(),
			)
	}

	events.EmitSwapExactOutEvent(ctx, sender, receiver, &pool, &inCoin, &msg.Outgoing, &fee)

	// Return response
	return &types.MsgSwapExactOutResponse{
		PoolId:      pool.PoolId,
		Incoming:    inCoin,
		Outgoing:    msg.Outgoing,
		Fee:         fee,
		Receiver:    receiver.String(),
		IncomingMax: msg.IncomingMax,
	}, nil
}

func CalculateSwapExactOutAmount(pool *types.Pool, outgoingQuote sdk.Coin) (incomingBaseCoin sdk.Coin, feeBaseCoin sdk.Coin, err error) {

	var baseCoin, quoteCoin int32

	switch outgoingQuote.Denom {
	case pool.Coins[0].Denom:
		baseCoin, quoteCoin = 1, 0
	case pool.Coins[1].Denom:
		baseCoin, quoteCoin = 0, 1
	default:
		// Invalid outgoingQuote coin
		return sdk.Coin{}, sdk.Coin{}, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"Invalid outgoing coin (%s), this pool only supports (%s) and (%s) tokens",
			outgoingQuote.Denom,
			pool.Coins[0].Denom,
			pool.Coins[1].Denom,
		)
	}

	// Constant product formula: x * y = k
	K := pool.Coins[baseCoin].Amount.Mul(pool.Coins[quoteCoin].Amount)

	// create math int from pool setting, so we can do safe math
	feeRatePerHundredThousand := math.NewInt(int64(pool.Fee))

	// create math int from pool setting, so we can do safe math
	scalingFactor := math.NewInt(constants.PoolFeeScalingFactor)

	// calculate fee
	// feeQuote := outgoingQuote.Amount.Mul(feeRatePerHundredThousand).Quo(scalingFactor)

	// sanity check
	if outgoingQuote.Amount.GTE(pool.Coins[quoteCoin].Amount) {
		return sdk.Coin{}, sdk.Coin{}, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"Outgoing amount (%s) must be less than pool's quote coin amount (%s)",
			outgoingQuote.Amount.String(),
			pool.Coins[quoteCoin].Amount.String(),
		)
	}

	// calculate new quote token balance for constant formula k = x * y
	newQuoteTokenBalance := pool.Coins[quoteCoin].Amount.Sub(outgoingQuote.Amount)

	// sanity check
	if !newQuoteTokenBalance.IsPositive() {
		return sdk.Coin{}, sdk.Coin{}, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"Invalid new quote token balance: %s, has to be positive.",
			newQuoteTokenBalance.String(),
		)
	}

	newBaseTokenBalance := K.Quo(newQuoteTokenBalance)

	incomingAfterFee := newBaseTokenBalance.Sub(pool.Coins[baseCoin].Amount)

	// sanity check
	if !incomingAfterFee.IsPositive() {
		return sdk.Coin{}, sdk.Coin{},
			errorsmod.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"Invalid new %s token balance: %s, has to be positive, after fee.",
				pool.Coins[baseCoin].Denom,
				incomingAfterFee.String(),
			)
	}

	// Calculation should be: incomingBase = incomingAfterFee / (1 - fee/scale)
	// But we use to avoid a problem with decimals: incomingBase = incomingAfterFee * scale / (scale - fee)
	denominator := scalingFactor.Sub(feeRatePerHundredThousand)
	incomingBase := incomingAfterFee.Mul(scalingFactor).Quo(denominator)

	// Enforce minimum fee
	feeAmount := incomingBase.Mul(feeRatePerHundredThousand).Quo(scalingFactor)
	if feeAmount.LT(math.NewInt(types.MinSwapFee)) {
		if incomingBase.LT(math.NewInt(types.MinSwapFee)) {
			return sdk.Coin{}, sdk.Coin{}, errorsmod.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"Swap amount too low to cover minimum fee (%d)", types.MinSwapFee,
			)
		}
		feeAmount = math.NewInt(types.MinSwapFee)
	}
	feeBaseCoin = sdk.NewCoin(pool.Coins[baseCoin].Denom, feeAmount)
	incomingBaseCoin = sdk.NewCoin(pool.Coins[baseCoin].Denom, incomingBase)

	// sanity check
	if incomingBaseCoin.IsZero() || incomingBaseCoin.IsNegative() {
		return sdk.Coin{}, sdk.Coin{},
			errorsmod.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"Invalid swap incoming amount: %s balance: %s, has to be positive.",
				incomingBaseCoin.Denom,
				incomingBaseCoin.Amount.String(),
			)
	}

	return incomingBaseCoin, feeBaseCoin, nil
}
