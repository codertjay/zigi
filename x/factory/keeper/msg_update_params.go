package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/x/factory/types"
)

func (k msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if k.GetAuthority() != req.Authority {
		return nil, errorsmod.Wrapf(types.ErrInvalidSignerAccount, "invalid authority; expected %s, got %s", k.GetAuthority(), req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the CreateFeeDenom exists in the factory module or is a native denom with supply
	if _, found := k.GetDenom(ctx, req.Params.CreateFeeDenom); !found {
		// If not found in factory module, check if it's a native denom with supply
		if !k.bankKeeper.HasSupply(ctx, req.Params.CreateFeeDenom) {
			return nil, errorsmod.Wrapf(
				types.ErrorInvalidParamValue,
				"Denom for create fee denom %s does not exist in the factory module and is not a native denom with supply",
				req.Params.CreateFeeDenom,
			)
		}
	}

	if err := k.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
