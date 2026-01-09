package keeper

import (
	"context"

	"zigchain/x/dex/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s queryServer) GetPoolsMeta(goCtx context.Context, req *types.QueryGetPoolsMetaRequest) (*types.QueryGetPoolsMetaResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	val, found := s.k.GetPoolsMeta(ctx)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetPoolsMetaResponse{PoolsMeta: val}, nil
}
