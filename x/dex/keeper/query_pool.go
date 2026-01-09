package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"zigchain/x/dex/types"
)

func (s queryServer) ListPool(ctx context.Context, req *types.QueryAllPoolRequest) (*types.QueryAllPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var pools []types.Pool

	store := runtime.KVStoreAdapter(s.k.storeService.OpenKVStore(ctx))
	poolStore := prefix.NewStore(store, types.KeyPrefix(types.PoolKeyPrefix))

	pageRes, err := query.Paginate(poolStore, req.Pagination, func(key []byte, value []byte) error {
		var pool types.Pool
		if err := s.k.cdc.Unmarshal(value, &pool); err != nil {
			return err
		}

		pools = append(pools, pool)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllPoolResponse{Pool: pools, Pagination: pageRes}, nil
}

func (s queryServer) GetPool(ctx context.Context, req *types.QueryGetPoolRequest) (*types.QueryGetPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	val, found := s.k.GetPool(
		ctx,
		req.PoolId,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetPoolResponse{
		Pool: val,
	}, nil
}

func (s queryServer) GetPoolBalances(ctx context.Context, req *types.QueryGetPoolBalancesRequest) (*types.QueryGetPoolBalancesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	val, found := s.k.GetPool(
		ctx,
		req.PoolId,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	poolAddress := types.GetPoolAddress(req.PoolId)
	balances := s.k.bankKeeper.GetAllBalances(ctx, poolAddress)

	return &types.QueryGetPoolBalancesResponse{
		Pool:     val,
		Balances: balances,
	}, nil
}
