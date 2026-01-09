package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"zigchain/x/dex/types"
	"zigchain/zutils/validators"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s queryServer) SwapIn(goCtx context.Context, req *types.QuerySwapInRequest) (*types.QuerySwapInResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := validators.CheckPoolId(req.PoolId); err != nil {
		return nil, err
	}

	coinIn, err := sdk.ParseCoinNormalized(req.CoinIn)

	if err != nil {
		return nil,
			errorsmod.Wrapf(
				err,
				"Invalid token format: failed to parse incoming token: %s",
				req.CoinIn,
			)
	}

	if err = validators.CheckCoinDenom(coinIn); err != nil {
		return nil,
			errorsmod.Wrapf(
				err,
				"Invalid token format: failed to validate incoming token: %s",
				req.CoinIn,
			)
	}

	pool, found := s.k.GetPool(ctx, req.PoolId)
	if !found {
		return nil,
			errorsmod.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"Liquidity pool (%s) can not be found",
				req.PoolId,
			)
	}

	coinOut, feeCoin, err := CalculateSwapAmount(&pool, coinIn)

	if err != nil {
		return nil, err
	}

	return &types.QuerySwapInResponse{
		CoinOut: coinOut,
		Fee:     feeCoin,
	}, nil
}
