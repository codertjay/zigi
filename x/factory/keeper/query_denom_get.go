package keeper

import (
	"context"

	cosmosmath "cosmossdk.io/math"

	"zigchain/x/factory/types"
	"zigchain/zutils/validators"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) Denom(ctx context.Context, req *types.QueryGetDenomRequest) (*types.QueryDenomResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// Validate the original input first
	if err := validators.CheckDenomString(req.Denom); err != nil {
		return nil,
			status.Errorf(
				codes.InvalidArgument,
				"invalid denom (%s): %s",
				req.Denom,
				err,
			)
	}

	denom, found := k.GetDenom(
		ctx,
		req.Denom,
	)
	if !found {
		return nil,
			status.Errorf(
				codes.NotFound,
				"denom (%s) not found",
				req.Denom,
			)
	}

	denomAuth, found := k.GetDenomAuth(
		ctx,
		req.Denom,
	)
	if !found {
		return nil,
			status.Errorf(
				codes.NotFound,
				"denomAuth (%s) not found",
				req.Denom,
			)
	}

	if denom.Denom != denomAuth.Denom {
		return nil,
			status.Errorf(
				codes.NotFound,
				"denom (%s) not found",
				req.Denom,
			)
	}

	totalBurned, maxSupply, totalSupply := k.calculateDenomStats(ctx, denom)

	return &types.QueryDenomResponse{
		Denom:               denom.Denom,
		TotalMinted:         denom.Minted,
		TotalSupply:         cosmosmath.Uint(totalSupply.Amount),
		TotalBurned:         totalBurned,
		MaxSupply:           maxSupply,
		MintingCap:          denom.MintingCap,
		CanChangeMintingCap: denom.CanChangeMintingCap,
		Creator:             denom.Creator,
		BankAdmin:           denomAuth.BankAdmin,
		MetadataAdmin:       denomAuth.MetadataAdmin,
	}, nil
}
