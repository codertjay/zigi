package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"zigchain/x/factory/types"
	"zigchain/zutils/validators"
)

func (k Keeper) ListDenomAuth(ctx context.Context, req *types.QueryAllDenomAuthRequest) (*types.QueryAllDenomAuthResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var denomAuths []types.DenomAuth

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	denomAuthStore := prefix.NewStore(store, types.KeyPrefix(types.DenomAuthKeyPrefix))

	pageRes, err := query.Paginate(denomAuthStore, req.Pagination, func(key []byte, value []byte) error {
		var denomAuth types.DenomAuth
		if err := k.cdc.Unmarshal(value, &denomAuth); err != nil {
			return err
		}

		denomAuths = append(denomAuths, denomAuth)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllDenomAuthResponse{DenomAuth: denomAuths, Pagination: pageRes}, nil
}

func (k Keeper) DenomAuth(ctx context.Context, req *types.QueryGetDenomAuthRequest) (*types.QueryDenomAuthResponse, error) {
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

	val, found := k.GetDenomAuth(
		ctx,
		req.Denom,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryDenomAuthResponse{DenomAuth: val}, nil
}
