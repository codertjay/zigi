package bindings

import (
	cosmosmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type ZMsg struct {
	// Contracts can create denoms, namespaced under the contract's address.
	// A contract may create any number of independent sub-denoms.
	CreateDenom *CreateDenom `json:"create_denom,omitempty"`

	// Contracts can set the metadata of a denom.
	SetDenomMetadata *SetDenomMetadata `json:"set_denom_metadata,omitempty"`

	// Contracts can update the admins address: bank, and metadata of a denom.
	ProposeDenomAdmin *ProposeDenomAdmin `json:"propose_denom_admin,omitempty"`

	// Contracts can claim the denom admin role.
	ClaimDenomAdmin *ClaimDenomAdmin `json:"claim_denom_admin,omitempty"`

	// Contracts can update the metadata admin of a denom.
	UpdateDenomMetadataAuth *UpdateDenomMetadataAuth `json:"update_denom_metadata_auth,omitempty"`

	// Contracts can update the URI of a denom and its sha256 hash.
	UpdateDenomURI *UpdateDenomURI `json:"update_denom_uri,omitempty"`

	// Contracts can update the minting cap and options to lock minting cap changes on a denom.
	UpdateDenomMintingCap *UpdateDenomMintingCap `json:"update_denom_minting_cap,omitempty"`

	// Contracts can mint native tokens for an existing factory denom
	// that they are the admin of.
	MintAndSendTokens *MintAndSendTokens `json:"mint_and_send_tokens,omitempty"`

	// Contracts can burn tokens from the signer's account.
	BurnTokens *BurnTokens `json:"burn_tokens,omitempty"`

	// CreatePool creates a new pool.
	CreatePool *CreatePool `json:"create_pool,omitempty"`

	// AddLiquidity adds liquidity to a pool and sends the pool tokens to the signer.
	AddLiquidity *AddLiquidity `json:"add_liquidity,omitempty"`

	// RemoveLiquidity removes liquidity from a pool and sends the base and quote tokens to the signer.
	RemoveLiquidity *RemoveLiquidity `json:"remove_liquidity,omitempty"`

	// SwapExactIn executes a swap between two tokens in a pool given incoming amount.
	SwapExactIn *SwapExactIn `json:"swap_exact_in,omitempty"`

	// SwapExactOut executes a swap between two tokens in a pool given outgoing amount.
	SwapExactOut *SwapExactOut `json:"swap_exact_out,omitempty"`
}

// CreateDenom creates a new factory denom:
type CreateDenom struct {
	Denom               string          `json:"denom"`
	MintingCap          cosmosmath.Uint `json:"minting_cap"`
	CanChangeMintingCap bool            `json:"can_change_minting_cap"`
	URI                 string          `json:"uri"`
	URIHash             string          `json:"uri_hash"`
}

// SetDenomMetadata sets the metadata of a denom.
type SetDenomMetadata struct {
	Metadata banktypes.Metadata `json:"metadata"`
}

// ProposeDenomAdmin proposes the admins address: bank, and metadata of a denom.
type ProposeDenomAdmin struct {
	Denom         string `json:"denom"`
	BankAdmin     string `json:"bank_admin"`
	MetadataAdmin string `json:"metadata_admin"`
}

// ClaimDenomAdmin claims the denom admin role.
type ClaimDenomAdmin struct {
	Denom string `json:"denom"`
}

// UpdateDenomMetadataAuth updates the metadata admin of a denom.
type UpdateDenomMetadataAuth struct {
	Denom         string `json:"denom"`
	MetadataAdmin string `json:"metadata_admin"`
}

// UpdateDenomURI updates the URI of a denom and its sha256 hash.
type UpdateDenomURI struct {
	Denom   string `json:"denom"`
	URI     string `json:"uri"`
	URIHash string `json:"uri_hash"`
}

// UpdateDenomMintingCap updates the minting cap and options to lock minting cap changes on a denom.
type UpdateDenomMintingCap struct {
	Denom               string          `json:"denom"`
	MintingCap          cosmosmath.Uint `json:"minting_cap"`
	CanChangeMintingCap bool            `json:"can_change_minting_cap"`
}

// MintAndSendTokens mints and sends tokens to an address.
type MintAndSendTokens struct {
	Token     sdk.Coin `json:"token"`
	Recipient string   `json:"recipient"`
}

// BurnTokens burns tokens from the signer's account.
type BurnTokens struct {
	Token sdk.Coin `json:"token"`
}

// DEX STUFF

// CreatePool creates a new pool.
type CreatePool struct {
	Base  sdk.Coin `json:"base"`
	Quote sdk.Coin `json:"quote"`
	// receiver is optional, if not provided, the signer is the receiver
	Receiver string `json:"receiver"`
}

// AddLiquidity adds liquidity to a pool and sends the pool tokens to the signer.
type AddLiquidity struct {
	PoolID string   `json:"pool_id"`
	Base   sdk.Coin `json:"base"`
	Quote  sdk.Coin `json:"quote"`
	// receiver is optional, if not provided, the signer is the receiver
	Receiver string `json:"receiver"`
}

// RemoveLiquidity removes liquidity from a pool and sends the base and quote tokens to the signer.
type RemoveLiquidity struct {
	LPToken sdk.Coin `json:"lp_token"`
	// receiver is optional, if not provided, the signer is the receiver
	Receiver string `json:"receiver"`
}

// SwapExactIn executes a swap between two tokens in a pool.
type SwapExactIn struct {
	Incoming sdk.Coin `json:"incoming"`
	PoolID   string   `json:"pool_id"`
	// receiver is optional, if not provided, the signer is the receiver
	Receiver string `json:"receiver"`
	// outgoing_min is optional
	OutgoingMin *sdk.Coin `json:"outgoing_min"`
}

// SwapExactOut executes a swap between two tokens in a pool.
type SwapExactOut struct {
	Outgoing sdk.Coin `json:"outgoing"`
	PoolID   string   `json:"pool_id"`
	// receiver is optional, if not provided, the signer is the receiver
	Receiver string `json:"receiver"`
	// incoming_max is optional
	IncomingMax *sdk.Coin `json:"incoming_max"`
}
