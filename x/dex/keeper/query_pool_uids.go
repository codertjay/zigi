package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/x/dex/types"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s queryServer) ListPoolUids(ctx context.Context, req *types.QueryAllPoolUidsRequest) (*types.QueryAllPoolUidsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var poolUidss []types.PoolUids

	store := runtime.KVStoreAdapter(s.k.storeService.OpenKVStore(ctx))
	poolUidsStore := prefix.NewStore(store, types.KeyPrefix(types.PoolUidsKeyPrefix))

	pageRes, err := query.Paginate(poolUidsStore, req.Pagination, func(key []byte, value []byte) error {
		var poolUids types.PoolUids
		if err := s.k.cdc.Unmarshal(value, &poolUids); err != nil {
			return err
		}

		poolUidss = append(poolUidss, poolUids)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllPoolUidsResponse{PoolUids: poolUidss, Pagination: pageRes}, nil
}

func (s queryServer) GetPoolUid(ctx context.Context, req *types.QueryGetPoolUidRequest) (*types.QueryGetPoolUidResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// Sort them so it works both ways - cannot use zero as zero will be removed
	coins := sdk.NewCoins(sdk.NewCoin(req.Base, math.NewInt(1)), sdk.NewCoin(req.Quote, math.NewInt(1)))

	// Sanity check
	if len(coins) != 2 {
		return nil,
			sdkerrors.Wrapf(
				types.ErrPoolNotFound,
				"invalid coins: %s",
				coins.String(),
			)

	}
	poolUidString := coins[0].Denom + types.PoolUidSeparator + coins[1].Denom

	poolUid, found := s.k.GetPoolUids(
		ctx,
		poolUidString,
	)

	if !found {
		return nil,
			sdkerrors.Wrapf(
				types.ErrPoolNotFound,
				"Pool %s with base: %s and quote: %s not found",
				poolUidString,
				req.Base,
				req.Quote,
			)
	}

	return &types.QueryGetPoolUidResponse{PoolUids: poolUid}, nil
}
