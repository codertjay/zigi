package keeper

import (
	"fmt"
	"zigchain/x/factory/types"

	cosmosmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterInvariants registers all module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(
		types.ModuleName,
		"supply-more-than-minting-cap",
		SupplyMoreThenMintingCapInvariant(k),
	)
	ir.RegisterRoute(
		types.ModuleName,
		"factory-supply-gte-bank-supply",
		FactorySupplyGTEBankSupplyInvariant(k),
	)
}

// SupplyMoreThenMintingCapInvariant checks that total supply is less than or equal to minting cap
func SupplyMoreThenMintingCapInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {

		// Fetch all denominations directly, without using the query interface, to make it more efficient
		all := k.GetAllDenom(ctx)

		// Iterate over all denoms indexes
		for idx := range all {

			// get denom using index in the slice (array)
			denom := all[idx]

			// check if total supply is less than or equal to minting cap
			if denom.Minted.GT(denom.MintingCap) {
				// return error and true for invariant error
				// returning false will stop checking further denominations
				// if a denomination's total supply exceeds minting cap, we already have broken system
				return fmt.Sprintf("total supply is greater than minting cap for denom %s", denom.Denom), true
			}

		}
		// return nothing for error and false for invariant error
		return "", false
	}
}

// FactorySupplyGTEBankSupplyInvariant checks
// that the total supply of each denom in factory is equal to the total supply of each denom in bank
func FactorySupplyGTEBankSupplyInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {

		// Fetch all denominations directly, without using the query interface, to make it more efficient
		all := k.GetAllDenom(ctx)

		// Iterate over all denoms indexes
		for idx := range all {

			// get denom using index in the slice (array)
			denom := all[idx]

			// get total supply from bank keeper
			bankSupply := k.bankKeeper.GetSupply(ctx, denom.Denom)

			// check if total supply is greater or equal to bank supply
			// in case of IBC transfer coins are burned and re-minted on another chain,
			// in that case total supply will be more than bank supply
			if !denom.Minted.GTE(cosmosmath.Uint(bankSupply.Amount)) {
				return fmt.Sprintf("total supply is not equal to bank supply for denom %s", denom.Denom), true
			}
		}

		return "", false
	}
}
