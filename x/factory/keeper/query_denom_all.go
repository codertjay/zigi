package keeper

import (
	"context"

	cosmosmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/x/factory/types"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// calculateDenomStats calculates total burned and max supply for a denom
func (k Keeper) calculateDenomStats(ctx context.Context, denom types.Denom) (totalBurned, maxSupply cosmosmath.Uint, totalSupply sdk.Coin) {
	// Get total supply from a bank
	totalSupply = k.bankKeeper.GetSupply(ctx, denom.Denom)

	// Calculate total burned with safety check
	if denom.Minted.GT(cosmosmath.Uint(totalSupply.Amount)) {
		totalBurned = denom.Minted.Sub(cosmosmath.Uint(totalSupply.Amount))
	} else {
		totalBurned = cosmosmath.ZeroUint()
	}

	// Calculate max supply with safety check
	if denom.MintingCap.GT(totalBurned) {
		maxSupply = denom.MintingCap.Sub(totalBurned)
	} else {
		maxSupply = cosmosmath.ZeroUint()
	}

	return totalBurned, maxSupply, totalSupply
}

func (k Keeper) DenomAll(ctx context.Context, req *types.QueryAllDenomRequest) (*types.QueryAllDenomResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var denoms []types.Denom

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	denomStore := prefix.NewStore(store, types.KeyPrefix(types.DenomKeyPrefix))

	pageRes, err := query.Paginate(denomStore, req.Pagination, func(key []byte, value []byte) error {
		var denom types.Denom
		if err := k.cdc.Unmarshal(value, &denom); err != nil {
			return err
		}

		denoms = append(denoms, denom)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var denomsResponse []types.DenomResponse

	for _, denom := range denoms {
		totalBurned, maxSupply, totalSupply := k.calculateDenomStats(ctx, denom)

		denomsResponse = append(denomsResponse, types.DenomResponse{
			Creator:             denom.Creator,
			Denom:               denom.Denom,
			MintingCap:          denom.MintingCap,
			MaxSupply:           maxSupply,
			CanChangeMintingCap: denom.CanChangeMintingCap,
			TotalMinted:         denom.Minted,
			TotalSupply:         cosmosmath.Uint(totalSupply.Amount),
			TotalBurned:         totalBurned,
		})
	}

	return &types.QueryAllDenomResponse{Denom: denomsResponse, Pagination: pageRes}, nil
}
