package events

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/x/dex/types"
)

func EmitPoolCreateEvent(
	ctx sdk.Context,
	sender sdk.AccAddress,
	pool *types.Pool,
	receiver sdk.AccAddress,
) {
	ctx.EventManager().EmitEvents(sdk.Events{
		newPoolEvent(sender, pool, receiver),
	})
}

func newPoolEvent(sender sdk.AccAddress, pool *types.Pool, receiver sdk.AccAddress) sdk.Event {

	coinsIn := sdk.NewCoins((pool.Coins)[0], (pool.Coins)[1])

	return sdk.NewEvent(
		types.EventPoolCreated,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		sdk.NewAttribute(sdk.AttributeKeySender, sender.String()),
		sdk.NewAttribute(types.AttributeKeyPoolId, pool.PoolId),
		sdk.NewAttribute(types.AttributeKeyTokensIn, coinsIn.String()),
		sdk.NewAttribute(types.AttributeKeyLPTokenOut, pool.LpToken.String()),
		sdk.NewAttribute(types.AttributeKeyReceiver, receiver.String()),
		sdk.NewAttribute(types.AttributeKeyPoolAddress, pool.Address),
	)
}

func EmitSwapEvent(ctx sdk.Context, sender sdk.AccAddress, receiver sdk.AccAddress, pool *types.Pool, input *sdk.Coin, output *sdk.Coin, fee *sdk.Coin) {
	ctx.EventManager().EmitEvents(sdk.Events{
		newSwapEvent(sender, receiver, pool, input, output, fee),
	})
}

func newSwapEvent(sender sdk.AccAddress, receiver sdk.AccAddress, pool *types.Pool, input *sdk.Coin, output *sdk.Coin, fee *sdk.Coin) sdk.Event {

	poolCoinsAfter := sdk.NewCoins((pool.Coins)[0], (pool.Coins)[1], pool.LpToken)

	return sdk.NewEvent(
		types.EventTokenSwapped,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		sdk.NewAttribute(sdk.AttributeKeySender, sender.String()),
		sdk.NewAttribute(types.AttributeKeyReceiver, receiver.String()),
		sdk.NewAttribute(types.AttributeKeyPoolId, pool.PoolId),
		sdk.NewAttribute(types.AttributeKeyTokensIn, input.String()),
		sdk.NewAttribute(types.AttributeKeyTokensOut, output.String()),
		sdk.NewAttribute(types.AttributeKeySwapFee, fee.String()),
		sdk.NewAttribute(types.AttributeKeyPoolState, poolCoinsAfter.String()),
	)
}

func EmitSwapExactOutEvent(
	ctx sdk.Context,
	sender sdk.AccAddress,
	receiver sdk.AccAddress,
	pool *types.Pool,
	input *sdk.Coin,
	output *sdk.Coin,
	fee *sdk.Coin,
) {
	ctx.EventManager().EmitEvents(sdk.Events{
		newSwapExactOutEvent(sender, receiver, pool, input, output, fee),
	})
}

func newSwapExactOutEvent(sender sdk.AccAddress, receiver sdk.AccAddress, pool *types.Pool, input *sdk.Coin, output *sdk.Coin, fee *sdk.Coin) sdk.Event {

	poolCoinsAfter := sdk.NewCoins((pool.Coins)[0], (pool.Coins)[1], pool.LpToken)

	return sdk.NewEvent(
		types.EventTokenSwapped,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		sdk.NewAttribute(sdk.AttributeKeySender, sender.String()),
		sdk.NewAttribute(types.AttributeKeyReceiver, receiver.String()),
		sdk.NewAttribute(types.AttributeKeyPoolId, pool.PoolId),
		sdk.NewAttribute(types.AttributeKeyTokensIn, input.String()),
		sdk.NewAttribute(types.AttributeKeyTokensOut, output.String()),
		sdk.NewAttribute(types.AttributeKeySwapFee, fee.String()),
		sdk.NewAttribute(types.AttributeKeyPoolState, poolCoinsAfter.String()),
	)
}

func EmitAddLiquidityEvent(
	ctx sdk.Context,
	sender sdk.AccAddress,
	pool *types.Pool,
	coinsIn *sdk.Coins,
	lpTokenOut *sdk.Coin,
	returnedCoins *sdk.Coins,
	receiver sdk.AccAddress,
) {
	ctx.EventManager().EmitEvents(sdk.Events{
		newAddLiquidityEvent(sender, pool, coinsIn, lpTokenOut, returnedCoins, receiver),
	})
}

func newAddLiquidityEvent(
	sender sdk.AccAddress,
	pool *types.Pool,
	coinsIn *sdk.Coins,
	lpTokenOut *sdk.Coin,
	returnedCoins *sdk.Coins,
	receiver sdk.AccAddress,
) sdk.Event {

	poolCoinsAfter := sdk.NewCoins((pool.Coins)[0], (pool.Coins)[1], pool.LpToken)

	return sdk.NewEvent(
		types.EventLiquidityAdded,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		sdk.NewAttribute(sdk.AttributeKeySender, sender.String()),
		sdk.NewAttribute(types.AttributeKeyPoolId, pool.PoolId),
		sdk.NewAttribute(types.AttributeKeyTokensIn, coinsIn.String()),
		sdk.NewAttribute(types.AttributeKeyLPTokenOut, lpTokenOut.String()),
		sdk.NewAttribute(types.AttributeKeyReturnedCoins, returnedCoins.String()),
		sdk.NewAttribute(types.AttributeKeyPoolState, poolCoinsAfter.String()),
		sdk.NewAttribute(types.AttributeKeyReceiver, receiver.String()),
	)
}

func EmitRemoveLiquidityEvent(
	ctx sdk.Context,
	sender sdk.AccAddress,
	pool *types.Pool,
	lpCoinIn *sdk.Coin,
	coinsOut *sdk.Coins,
	receiver sdk.AccAddress,
) {
	ctx.EventManager().EmitEvents(sdk.Events{
		newRemoveLiquidityEvent(sender, pool, lpCoinIn, coinsOut, receiver),
	})
}

func newRemoveLiquidityEvent(
	sender sdk.AccAddress,
	pool *types.Pool,
	lpCoinIn *sdk.Coin,
	coinsOut *sdk.Coins,
	receiver sdk.AccAddress,
) sdk.Event {

	poolCoinsAfter := sdk.NewCoins((pool.Coins)[0], (pool.Coins)[1], pool.LpToken)

	return sdk.NewEvent(
		types.EventLiquidityRemoved,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		sdk.NewAttribute(sdk.AttributeKeySender, sender.String()),
		sdk.NewAttribute(types.AttributeKeyPoolId, pool.PoolId),
		sdk.NewAttribute(types.AttributeKeyLPTokenIn, lpCoinIn.String()),
		sdk.NewAttribute(types.AttributeKeyTokensOut, coinsOut.String()),
		sdk.NewAttribute(types.AttributeKeyPoolState, poolCoinsAfter.String()),
		sdk.NewAttribute(types.AttributeKeyReceiver, receiver.String()),
	)
}
