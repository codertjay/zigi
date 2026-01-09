package keeper

import (
	"context"

	"zigchain/x/factory/types"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) DenomsByAdmin(ctx context.Context, req *types.QueryDenomByAdminRequest) (*types.QueryDenomByAdminResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// Apply pagination to the denoms list
	var paginatedDenoms []string

	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.AdminDenomAuthListKey(req.Admin))

	pageRes, err := query.Paginate(store, req.Pagination, func(key []byte, value []byte) error {
		paginatedDenoms = append(paginatedDenoms, string(key))
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryDenomByAdminResponse{
		Denoms:     paginatedDenoms,
		Pagination: pageRes,
	}, nil
}
