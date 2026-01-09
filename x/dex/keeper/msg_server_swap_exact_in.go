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

func (k msgServer) SwapExactIn(goCtx context.Context, msg *types.MsgSwapExactIn) (*types.MsgSwapExactInResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !msg.Incoming.IsPositive() {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"Incoming amount must be positive (%s%s)",
			msg.Incoming.Amount.String(),
			msg.Incoming.Denom,
		)
	}

	// Update user balances
	sender, err := sdk.AccAddressFromBech32(msg.Signer)
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
		receiver = sender
	}

	// Check if the sender has enough balances on incoming denom to satisfy the incoming amount
	if !k.bankKeeper.HasBalance(ctx, sender, msg.Incoming) {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInsufficientFunds,
			"Sender (%s) does not have enough balance for incoming amount (%s)",
			sender.String(),
			msg.Incoming.String(),
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

	switch msg.Incoming.Denom {
	case pool.Coins[0].Denom:
		fromCoin, toCoin = 0, 1
	case pool.Coins[1].Denom:
		fromCoin, toCoin = 1, 0
	default:
		// Invalid incoming coin
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"Invalid incoming coin (%s), this pool only supports base (%s) and quote (%s) tokens",
			msg.Incoming.Denom,
			pool.Coins[0].Denom,
			pool.Coins[1].Denom,
		)
	}

	outCoin, fee, err := CalculateSwapAmount(&pool, msg.Incoming)
	if err != nil {
		return nil, err
	}

	// Check if the swap amount is less than the minimum outgoing amount
	if msg.OutgoingMin != nil {

		// validates that the outgoing minimum amount is a positive number
		if !msg.OutgoingMin.IsPositive() {
			return nil, errorsmod.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"Outgoing minimum amount must be positive (%s)",
				msg.OutgoingMin.String(),
			)
		}
		// validates that the denom is the correct one
		if msg.OutgoingMin.Denom != outCoin.Denom {
			return nil, errorsmod.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"Outgoing minimum denom (%s) must be the same as outgoing denom (%s)",
				msg.OutgoingMin.Denom,
				outCoin.Denom,
			)
		}

		// validates that the amount above the minimum
		if outCoin.IsLT(*msg.OutgoingMin) {
			return nil, errorsmod.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"SwapExactIn amount (%s) is less than minimum outgoing amount (%s)",
				outCoin.String(),
				msg.OutgoingMin.String(),
			)
		}
	}

	// Update liquidity pool balances
	pool.Coins[fromCoin] = pool.Coins[fromCoin].AddAmount(msg.Incoming.Amount)
	pool.Coins[toCoin] = pool.Coins[toCoin].SubAmount(outCoin.Amount)
	k.SetPool(ctx, pool)

	// Deduct base tokens from sender
	err = k.SendFromAddressToPool(
		ctx,
		sender,
		poolIDString,
		sdk.NewCoins(msg.Incoming))
	if err != nil {

		return nil,
			errorsmod.Wrapf(
				err,
				"SwapExactIn: Failed to send %s coins",
				msg.Incoming.String(),
			)
	}

	// Credit quote tokens to sender
	err = k.SendFromPoolToAddress(
		ctx,
		poolIDString,
		receiver,
		sdk.NewCoins(outCoin))
	if err != nil {
		return nil,
			errorsmod.Wrapf(
				err,
				"SwapExactIn: Failed to send coins %s",
				outCoin.String(),
			)
	}

	events.EmitSwapEvent(ctx, sender, receiver, &pool, &msg.Incoming, &outCoin, &fee)

	// Return response
	return &types.MsgSwapExactInResponse{
		PoolId:      pool.PoolId,
		Incoming:    msg.Incoming,
		Outgoing:    outCoin,
		Fee:         fee,
		Receiver:    receiver.String(),
		OutgoingMin: msg.OutgoingMin,
	}, nil
}

func CalculateSwapAmount(pool *types.Pool, incoming sdk.Coin) (out sdk.Coin, feeCoin sdk.Coin, err error) {

	var fromCoin, toCoin int32

	switch incoming.Denom {
	case pool.Coins[0].Denom:
		fromCoin, toCoin = 0, 1
	case pool.Coins[1].Denom:
		fromCoin, toCoin = 1, 0
	default:
		// Invalid incoming coin
		return sdk.Coin{}, sdk.Coin{}, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"Invalid incoming coin (%s), this pool only supports base (%s) and quote (%s) tokens",
			incoming.Denom,
			pool.Coins[0].Denom,
			pool.Coins[1].Denom,
		)
	}

	// Constant product formula: x * y = k
	K := pool.Coins[fromCoin].Amount.Mul(pool.Coins[toCoin].Amount)

	// create math int from pool setting, so we can do safe math
	feeRatePerHundredThousand := math.NewInt(int64(pool.Fee))

	// create math int from pool setting, so we can do safe math
	scalingFactor := math.NewInt(constants.PoolFeeScalingFactor)

	// calculate fee
	fee := incoming.Amount.Mul(feeRatePerHundredThousand).Quo(scalingFactor)

	// Enforce minimum fee
	if fee.LT(math.NewInt(types.MinSwapFee)) {
		if incoming.Amount.LT(math.NewInt(types.MinSwapFee)) {
			return sdk.Coin{}, sdk.Coin{}, errorsmod.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"Swap amount too low to cover minimum fee (%d)", types.MinSwapFee,
			)
		}
		fee = math.NewInt(types.MinSwapFee)
	}

	// what we are left with after fee
	IncomingBaseAmountAfterFee := incoming.Amount.Sub(fee)

	// calculate new quote token balance for constant formula k = x * y
	newBaseTokenBalance := pool.Coins[fromCoin].Amount.Add(IncomingBaseAmountAfterFee)

	// sanity check
	if !newBaseTokenBalance.IsPositive() {
		return sdk.Coin{}, sdk.Coin{},
			errorsmod.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"Invalid new %s token balance: %s, has to be positive.",
				pool.Coins[fromCoin].Denom,
				newBaseTokenBalance.String(),
			)
	}

	// calculate a new quote token balance - this is adjusted for fee
	newQuoteTokenBalance := K.Quo(newBaseTokenBalance)

	// calculate swap amount - how much are we sending to signer
	swapAmount := pool.Coins[toCoin].Amount.Sub(newQuoteTokenBalance)

	// create swap coin - Quote side of the swap in this case
	out = sdk.NewCoin(pool.Coins[toCoin].Denom, swapAmount)

	// sanity check
	if out.IsZero() || out.IsNegative() {
		return sdk.Coin{}, sdk.Coin{},
			errorsmod.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"Invalid swap amount: %s balance: %s, has to be positive.",
				out.Denom,
				out.Amount.String(),
			)
	}

	feeCoin = sdk.NewCoin(pool.Coins[fromCoin].Denom, fee)

	return out, feeCoin, nil
}
