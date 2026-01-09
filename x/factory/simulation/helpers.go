package simulation

import (
	"fmt"
	"math/big"

	// we use math/rand to generate random numbers predictably
	// so we can reproduce the same results in tests
	// nosem: math-random-used
	"math/rand"
	d "zigchain/zutils/debug"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"zigchain/x/factory/keeper"
	"zigchain/x/factory/types"
)

// FindAccount find a specific address from an account list
func FindAccount(accs []simtypes.Account, address string) (simtypes.Account, bool) {
	creator, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		panic(err)
	}
	return simtypes.FindAccount(accs, creator)
}

// RandomExistingDenom get one of the existing denoms
func RandomExistingDenom(r *rand.Rand, ctx sdk.Context, fk keeper.Keeper) (types.Denom, error) {
	// Create a message to get all denoms
	allDenomMsg := &types.QueryAllDenomRequest{}
	// Get all denoms from keeper
	allDenoms, err := fk.DenomAll(ctx, allDenomMsg)
	// Check if there is an error
	if err != nil {
		fk.Logger().Error(fmt.Sprintf("Could not get all denoms: %s", err))
		return types.Denom{}, err
	}
	// Get random denom from allDenoms
	randomDenom := allDenoms.Denom[r.Intn(len(allDenoms.Denom))]

	var denom = types.Denom{
		Creator:             randomDenom.Creator,
		Denom:               randomDenom.Denom,
		MintingCap:          randomDenom.MintingCap,
		Minted:              randomDenom.TotalMinted,
		CanChangeMintingCap: randomDenom.CanChangeMintingCap,
	}

	return denom, nil
}

// RandUIntBetween generates a random UInt between min and max (inclusive).
//func RandUIntBetween(r *rand.Rand, min math.Uint, max math.Uint) math.Uint {
//	// Calculate the range: diff = max - min + 1
//	diff := max.Sub(min).Add(math.NewUint(1))
//
//	for {
//		// Determine the number of bits needed
//		nBits := diff.BigInt().BitLen()
//		bytesNeeded := (nBits + 7) / 8
//		b := make([]byte, bytesNeeded)
//
//		// Read random bytes
//		_, err := r.Read(b)
//		if err != nil {
//			panic(err) // Handle the error appropriately
//		}
//
//		// Create a big.Int from the bytes
//		n := new(big.Int).SetBytes(b)
//
//		// Clear excess bits to ensure n < diff
//		excessBits := uint(bytesNeeded*8 - nBits)
//		if excessBits > 0 {
//			n.Rsh(n, excessBits)
//		}
//
//		// Check if n < diff
//		if n.Cmp(diff.BigInt()) < 0 {
//			// result = n + min
//			result := math.NewUintFromBigInt(n).Add(min)
//			return result
//		}
//		// Else, try again
//	}
//}

func RandUIntBetween(r *rand.Rand, min math.Uint, max math.Uint) math.Uint {
	// Calculate the range: diff = max - min + 1
	diff := max.Sub(min).Add(math.NewUint(1))

	for {
		// Determine the number of bits needed
		nBits := diff.BigInt().BitLen()
		bytesNeeded := (nBits + 7) / 8
		b := make([]byte, bytesNeeded)

		// Read random bytes
		_, err := r.Read(b)
		if err != nil {
			panic(err) // Handle the error appropriately
		}

		// Create a big.Int from the bytes
		n := new(big.Int).SetBytes(b)

		// Safely calculate the number of excess bits
		totalBits := bytesNeeded * 8
		if totalBits < nBits {
			panic("unexpected: totalBits is less than nBits") // Optional safety check
		}
		excessBits := totalBits - nBits // Safe subtraction of integers

		// Clear excess bits to ensure n < diff
		if excessBits > 0 {
			n.Rsh(n, uint(excessBits)) // Explicit uint conversion here is safe
		}

		// Check if n < diff
		if n.Cmp(diff.BigInt()) < 0 {
			// result = n + min
			result := math.NewUintFromBigInt(n).Add(min)
			return result
		}
		// Else, try again
	}
}

// GetBankerFromDenom Get banker for the denom
func GetBankerFromDenom(ctx sdk.Context, fk keeper.Keeper, accs []simtypes.Account, denom string) (simtypes.Account, error) {
	// Get denom auth struct from factory keeper
	denomAuth, found := fk.GetDenomAuth(ctx, denom)

	// Check for missing denom auth
	if !found {
		fk.Logger().Error(fmt.Sprintf("Could not get denom auth: %s", denom))
		return simtypes.Account{}, fmt.Errorf("could not get denom auth: %s", denom)
	}

	// Make sure bank admin is not empty, empty means no more minting allowed
	// This is a valid state, but we cannot execute minting in this case
	// Research how to mark this as ok in the simulation stats
	if denomAuth.BankAdmin == "" {

		fk.Logger().Info(fmt.Sprintf("Bank admin for Denom: %s is empty", denom))
		return simtypes.Account{}, nil
	}

	banker, found := FindAccount(accs, denomAuth.BankAdmin)

	if !found {
		fk.Logger().Error(fmt.Sprintf("Could not get banker from accounts: %s", denomAuth.BankAdmin))
		return simtypes.Account{}, fmt.Errorf("could not get banker  from accounts: %s", denomAuth.BankAdmin)
	}

	return banker, nil

}

func log(ctx sdk.Context, fk keeper.Keeper, msg string) {
	fk.Logger().Error(d.L(ctx, msg))
}
