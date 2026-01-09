package keeper

import (
	"context"

	"zigchain/x/tokenwrapper/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) ModuleInfo(goCtx context.Context, req *types.QueryModuleInfoRequest) (*types.QueryModuleInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	moduleAddr, balances := k.GetModuleWalletBalances(ctx)
	operatorAddress := k.GetOperatorAddress(ctx)
	tokenWrapperEnabled := k.IsEnabled(ctx)

	return &types.QueryModuleInfoResponse{
		ModuleAddress:           moduleAddr,
		Balances:                balances,
		OperatorAddress:         operatorAddress,
		ProposedOperatorAddress: k.GetProposedOperatorAddress(ctx),
		PauserAddresses:         k.GetPauserAddresses(ctx),
		TokenWrapperEnabled:     tokenWrapperEnabled,
		NativeClientId:          k.GetNativeClientId(ctx),
		CounterpartyClientId:    k.GetCounterpartyClientId(ctx),
		NativePort:              k.GetNativePort(ctx),
		CounterpartyPort:        k.GetCounterpartyPort(ctx),
		NativeChannel:           k.GetNativeChannel(ctx),
		CounterpartyChannel:     k.GetCounterpartyChannel(ctx),
		Denom:                   k.GetDenom(ctx),
		DecimalDifference:       k.GetDecimalDifference(ctx),
	}, nil
}
