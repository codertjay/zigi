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

func (s queryServer) SwapOut(goCtx context.Context, req *types.QuerySwapOutRequest) (*types.QuerySwapOutResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := validators.CheckPoolId(req.PoolId); err != nil {
		return nil, err
	}

	coinOut, err := sdk.ParseCoinNormalized(req.CoinOut)

	if err != nil {
		return nil,
			errorsmod.Wrapf(
				err,
				"Invalid token format: failed to parse outgoing token: %s",
				req.CoinOut,
			)
	}

	if err = validators.CheckCoinDenom(coinOut); err != nil {
		return nil,
			errorsmod.Wrapf(
				err,
				"Invalid token format: failed to validate outgoing token: %s",
				req.CoinOut,
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

	coinIn, feeCoin, err := CalculateSwapExactOutAmount(&pool, coinOut)

	if err != nil {
		return nil, err
	}

	return &types.QuerySwapOutResponse{
		CoinIn: coinIn,
		Fee:    feeCoin,
	}, nil
}
