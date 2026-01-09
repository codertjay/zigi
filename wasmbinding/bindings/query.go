package bindings

import (
	cosmosmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ZQuery is a query message option we can query with contracts.
type ZQuery struct {
	// returns the denom as used by `BankMsg::Send`.
	Denom *Denom `json:"denom,omitempty"`

	// returns pool information.
	Pool *Pool `json:"pool,omitempty"`

	// returns swap information based on the current pool state and incoming token.
	SwapIn *SwapIn `json:"swap_in,omitempty"`
}

// Denom is a query message option to get the full denom info based on denom name.
type Denom struct {
	Denom string `json:"denom"`
}

// DenomResponse is the response to the Denom query.
type DenomResponse struct {
	Denom               string          `json:"denom"`
	Minted              cosmosmath.Uint `json:"minted"`
	MintingCap          cosmosmath.Uint `json:"minting_cap"`
	CanChangeMintingCap bool            `json:"can_change_minting_cap"`
	Creator             string          `json:"creator"`
	BankAdmin           string          `json:"bank_admin"`
	MetadataAdmin       string          `json:"metadata_admin"`
}

// Pool is a query message option to get the full pool info based on pool ID.
type Pool struct {
	PoolID string `json:"pool_id"`
}

// PoolResponse is the response to the GetPool query.
type PoolResponse struct {
	PoolID  string     `json:"pool_id"`
	LPToken sdk.Coin   `json:"lp_token"`
	Creator string     `json:"creator"`
	Fee     uint32     `json:"fee"`
	Formula string     `json:"formula"`
	Coins   []sdk.Coin `json:"coins"`
}

// SwapIn is a query message option to get the swap in info based on pool ID and incoming token.
type SwapIn struct {
	PoolID string   `json:"pool_id"`
	CoinIn sdk.Coin `json:"coin_in"`
}

// SwapInResponse is the response to the SwapIn query.
type SwapInResponse struct {
	CoinOut sdk.Coin `json:"coin_out"`
	Fee     sdk.Coin `json:"fee"`
}

// SwapOut is a query message option to get the swap out info based on pool ID and outgoing token.
type SwapOut struct {
	PoolID string   `json:"pool_id"`
	CoinIn sdk.Coin `json:"coin_out"`
}

// SwapOutResponse is the response to the SwapOut query.
type SwapOutResponse struct {
	CoinIn sdk.Coin `json:"coin_in"`
	Fee    sdk.Coin `json:"fee"`
}
