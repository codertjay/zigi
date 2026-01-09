package keeper

import (
	"context"
	"fmt"

	"zigchain/x/tokenwrapper/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) WithdrawFromModuleWallet(goCtx context.Context, msg *types.MsgWithdrawFromModuleWallet) (*types.MsgWithdrawFromModuleWalletResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get the signer address
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return nil, err
	}

	// Check if the signer is the current operator
	currentOperator := k.GetOperatorAddress(ctx)
	if signer.String() != currentOperator {
		return nil, fmt.Errorf("only the current operator can withdraw from the module wallet")
	}

	// Unlock tokens in the module wallet
	if err := k.UnlockTokens(ctx, signer, msg.Amount); err != nil {
		return nil, err
	}

	// Get the module account balance
	moduleAddr, balances := k.GetModuleWalletBalances(ctx)

	// Emit event for withdrawing from module wallet
	types.EmitModuleWalletWithdrawnEvent(ctx, msg.Signer, moduleAddr, msg.Amount, balances)

	return &types.MsgWithdrawFromModuleWalletResponse{
		Signer:        msg.Signer,
		Amount:        msg.Amount,
		Balances:      balances,
		ModuleAddress: moduleAddr,
		Receiver:      signer.String(),
	}, nil
}
