package keeper

import (
	"context"

	"zigchain/x/tokenwrapper/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) TotalTransfers(goCtx context.Context, req *types.QueryTotalTransfersRequest) (*types.QueryTotalTransfersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	totalTransferredIn := k.GetTotalTransferredIn(ctx)
	totalTransferredOut := k.GetTotalTransferredOut(ctx)

	return &types.QueryTotalTransfersResponse{
		TotalTransferredIn:  totalTransferredIn,
		TotalTransferredOut: totalTransferredOut,
	}, nil
}
