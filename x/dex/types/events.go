package types

const (
	EventLiquidityAdded   = "liquidity_added"
	EventLiquidityRemoved = "liquidity_removed"
	EventTokenSwapped     = "token_swapped"
	EventPoolCreated      = "pool_created"

	AttributeValueCategory    = ModuleName
	AttributeKeyPoolId        = "pool_id"
	AttributeKeySwapFee       = "swap_fee"
	AttributeKeyTokensIn      = "token_in"
	AttributeKeyTokensOut     = "token_out"
	AttributeKeyLPTokenIn     = "lp_token_in"
	AttributeKeyLPTokenOut    = "lp_token_out"
	AttributeKeyReturnedCoins = "returned_coins"
	AttributeKeyPoolState     = "pool_snapshot"
	AttributeKeyReceiver      = "receiver"
	AttributeKeyPoolAddress   = "pool_address"
)
