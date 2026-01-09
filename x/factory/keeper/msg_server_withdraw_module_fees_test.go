package keeper_test

import (
	"fmt"
	"testing"

	cosmosmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	keepertest "zigchain/testutil/keeper"
	"zigchain/testutil/sample"
	"zigchain/x/factory/keeper"
	"zigchain/x/factory/testutil"
	"zigchain/x/factory/types"
)

// Positive test cases

func TestMsgServer_WithdrawModuleFees_Positive(t *testing.T) {
	// Test case: withdraw module fees successfully

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish()

	// create mock bank keeper and account keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	// create a new keeper and context
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, accountKeeper)
	server := keeper.NewMsgServerImpl(k)

	// update params
	params := types.DefaultParams()
	beneficiary := sample.AccAddress()
	params.Beneficiary = beneficiary
	require.NoError(t, k.SetParams(ctx, params))

	// create a mock module account
	moduleAccount := authtypes.NewEmptyModuleAccount(types.ModuleName)

	// mock balance of the module account
	moduleBalance := sdk.NewCoins(sdk.NewCoin("abc", cosmosmath.NewInt(1000)))

	accountKeeper.
		EXPECT().
		GetModuleAccount(gomock.Any(), types.ModuleName).
		Return(moduleAccount).
		Times(1)

	bankKeeper.
		EXPECT().
		GetAllBalances(gomock.Any(), moduleAccount.GetAddress()).
		Return(moduleBalance).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, sdk.MustAccAddressFromBech32(beneficiary), moduleBalance).
		Return(nil).
		Times(1)

	// create a message to withdraw module fees
	msg := &types.MsgWithdrawModuleFees{
		Signer: beneficiary,
	}

	// withdraw module fees
	resp, err := server.WithdrawModuleFees(ctx, msg)
	require.NoError(t, err)

	// check if the response is as expected
	expectedResp := &types.MsgWithdrawModuleFeesResponse{
		Signer:   beneficiary,
		Receiver: beneficiary,
		Amounts:  moduleBalance,
	}

	require.NotNil(t, resp)
	require.Equal(t, expectedResp, resp)
}

// Negative test cases

func TestMsgServer_WithdrawModuleFees_BeneficiaryNotSet(t *testing.T) {
	// Test case: try to withdraw module fees if the beneficiary is not set

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, accountKeeper)
	srv := keeper.NewMsgServerImpl(k)

	msg := &types.MsgWithdrawModuleFees{
		Signer: sample.AccAddress(),
	}

	_, err := srv.WithdrawModuleFees(ctx, msg)

	require.Error(t, err)
	require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)
	require.Equal(t, err.Error(), "beneficiary address is not set, cannot withdraw: unauthorized")
}

func TestMsgServer_WithdrawModuleFees_SignerNotBeneficiary(t *testing.T) {
	// Test case: try to withdraw module fees if the signer is not the beneficiary

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, accountKeeper)
	srv := keeper.NewMsgServerImpl(k)

	beneficiary := sample.AccAddress()
	nonBeneficiary := sample.AccAddress()

	// update params
	params := types.DefaultParams()
	params.Beneficiary = beneficiary
	require.NoError(t, k.SetParams(ctx, params))

	msg := &types.MsgWithdrawModuleFees{
		Signer: nonBeneficiary,
	}

	_, err := srv.WithdrawModuleFees(ctx, msg)

	require.Error(t, err)
	require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)
	require.Equal(
		t,
		fmt.Sprintf(
			"signer: %s is not the beneficiary: %s, only the beneficiary can withdraw: unauthorized",
			nonBeneficiary,
			beneficiary,
		),
		err.Error(),
	)
}

func TestMsgServer_WithdrawModuleFees_ModuleAccountNotFound(t *testing.T) {
	// Test case: try to withdraw module fees if the module account is not found

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, accountKeeper)
	srv := keeper.NewMsgServerImpl(k)

	// update params
	params := types.DefaultParams()
	beneficiary := sample.AccAddress()
	params.Beneficiary = beneficiary
	require.NoError(t, k.SetParams(ctx, params))

	accountKeeper.
		EXPECT().
		GetModuleAccount(gomock.Any(), types.ModuleName).
		Return(nil).
		Times(1)

	msg := &types.MsgWithdrawModuleFees{
		Signer: beneficiary,
	}

	_, err := srv.WithdrawModuleFees(ctx, msg)

	require.Error(t, err)
	require.ErrorIs(t, err, sdkerrors.ErrInvalidAddress)
	require.Equal(t, err.Error(), "module account not found: invalid address")
}

func TestMsgServer_WithdrawModuleFees_InsufficientFunds(t *testing.T) {
	// Test case: try to withdraw module fees when the module account has no funds

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, accountKeeper)
	srv := keeper.NewMsgServerImpl(k)

	// update params
	params := types.DefaultParams()
	beneficiary := sample.AccAddress()
	params.Beneficiary = beneficiary
	require.NoError(t, k.SetParams(ctx, params))

	// create a mock module account
	moduleAccount := authtypes.NewEmptyModuleAccount(types.ModuleName)

	accountKeeper.
		EXPECT().
		GetModuleAccount(gomock.Any(), types.ModuleName).
		Return(moduleAccount).
		Times(1)

	// simulate an empty balance (insufficient funds)
	bankKeeper.
		EXPECT().
		GetAllBalances(gomock.Any(), moduleAccount.GetAddress()).
		Return(sdk.NewCoins()).
		Times(1)

	// create a withdrawal message
	msg := &types.MsgWithdrawModuleFees{
		Signer: beneficiary,
	}

	// attempt withdrawal
	_, err := srv.WithdrawModuleFees(ctx, msg)

	// validate expected error
	require.Error(t, err)
	require.ErrorIs(t, err, sdkerrors.ErrInsufficientFunds)
	require.Equal(t, "module factory account has no balance to withdraw: insufficient funds", err.Error())
}

func TestMsgServer_WithdrawModuleFees_InvalidSignerAddress(t *testing.T) {
	// Test case: try to withdraw module fees with an invalid signer address

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, accountKeeper)
	srv := keeper.NewMsgServerImpl(k)

	// update params
	params := types.DefaultParams()
	beneficiary := sample.AccAddress()
	params.Beneficiary = beneficiary
	require.NoError(t, k.SetParams(ctx, params))

	msg := &types.MsgWithdrawModuleFees{
		Signer: "invalid_address",
	}

	_, err := srv.WithdrawModuleFees(ctx, msg)

	require.Error(t, err)
	require.Equal(
		t,
		fmt.Sprintf(
			"signer: invalid_address is not the beneficiary: %s, only the beneficiary can withdraw: unauthorized",
			beneficiary,
		),
		err.Error(),
	)
}

func TestMsgServer_WithdrawModuleFees_InvalidReceiverAddress(t *testing.T) {
	// Test case: try to withdraw module fees to an invalid receiver address

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, accountKeeper)
	srv := keeper.NewMsgServerImpl(k)

	// update params
	params := types.DefaultParams()
	beneficiary := sample.AccAddress()
	params.Beneficiary = beneficiary
	require.NoError(t, k.SetParams(ctx, params))

	// create a mock module account
	moduleAccount := authtypes.NewEmptyModuleAccount(types.ModuleName)

	// mock balance of the module account
	moduleBalance := sdk.NewCoins(sdk.NewCoin("abc", cosmosmath.NewInt(1000)))

	accountKeeper.
		EXPECT().
		GetModuleAccount(gomock.Any(), types.ModuleName).
		Return(moduleAccount).
		Times(1)

	bankKeeper.
		EXPECT().
		GetAllBalances(gomock.Any(), moduleAccount.GetAddress()).
		Return(moduleBalance).
		Times(1)

	// create a withdrawal message with an invalid receiver address
	invReceiver := "invalid_address"
	msg := &types.MsgWithdrawModuleFees{
		Signer:   beneficiary,
		Receiver: invReceiver,
	}

	// attempt withdrawal
	_, err := srv.WithdrawModuleFees(ctx, msg)

	// validate expected error
	require.Error(t, err)
	require.ErrorIs(t, err, sdkerrors.ErrInvalidAddress)
	require.Equal(
		t,
		fmt.Sprintf(
			"invalid receiver address: %s (decoding bech32 failed: invalid separator index -1): invalid address",
			invReceiver,
		),
		err.Error(),
	)
}

func TestMsgServer_WithdrawModuleFees_BankModuleFailure(t *testing.T) {
	// Test case: simulate a failure in the bank module when transferring funds

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, accountKeeper)
	srv := keeper.NewMsgServerImpl(k)

	// update params
	params := types.DefaultParams()
	beneficiary := sample.AccAddress()
	params.Beneficiary = beneficiary
	require.NoError(t, k.SetParams(ctx, params))

	// create a mock module account
	moduleAccount := authtypes.NewEmptyModuleAccount(types.ModuleName)

	// mock balance of the module account
	moduleBalance := sdk.NewCoins(sdk.NewCoin("abc", cosmosmath.NewInt(1000)))

	accountKeeper.
		EXPECT().
		GetModuleAccount(gomock.Any(), types.ModuleName).
		Return(moduleAccount).
		Times(1)

	bankKeeper.
		EXPECT().
		GetAllBalances(gomock.Any(), moduleAccount.GetAddress()).
		Return(moduleBalance).
		Times(1)

	// simulate a failure in transferring coins
	bankKeeper.
		EXPECT().
		SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, sdk.MustAccAddressFromBech32(beneficiary), moduleBalance).
		Return(sdkerrors.ErrInsufficientFunds).
		Times(1)

	// create a withdrawal message
	msg := &types.MsgWithdrawModuleFees{
		Signer: beneficiary,
	}

	// attempt withdrawal
	_, err := srv.WithdrawModuleFees(ctx, msg)

	// validate expected error
	require.Error(t, err)
	require.ErrorIs(t, err, sdkerrors.ErrInsufficientFunds)
	require.Equal(t, "failed to transfer funds: insufficient funds: insufficient funds", err.Error())
}

func TestMsgServer_WithdrawModuleFees_InvalidSignerAsReceiver(t *testing.T) {
	// Test case: try to withdraw module fees with invalid signer address when used as receiver

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, accountKeeper)
	srv := keeper.NewMsgServerImpl(k)

	// update params with invalid beneficiary address
	params := types.DefaultParams()
	invalidAddress := "invalid_address"
	params.Beneficiary = invalidAddress
	require.NoError(t, k.SetParams(ctx, params))

	// create a mock module account
	moduleAccount := authtypes.NewEmptyModuleAccount(types.ModuleName)

	// mock balance of the module account (non-zero to pass check)
	moduleBalance := sdk.NewCoins(sdk.NewCoin("abc", cosmosmath.NewInt(1000)))

	accountKeeper.
		EXPECT().
		GetModuleAccount(gomock.Any(), types.ModuleName).
		Return(moduleAccount).
		Times(1)

	bankKeeper.
		EXPECT().
		GetAllBalances(gomock.Any(), moduleAccount.GetAddress()).
		Return(moduleBalance).
		Times(1)

	// create a withdrawal message with invalid signer and empty receiver
	msg := &types.MsgWithdrawModuleFees{
		Signer:   invalidAddress,
		Receiver: "",
	}

	// attempt withdrawal
	_, err := srv.WithdrawModuleFees(ctx, msg)

	// validate expected error
	require.Error(t, err)
	require.ErrorIs(t, err, sdkerrors.ErrInvalidAddress)
	require.Equal(
		t,
		fmt.Sprintf(
			"invalid signer address: %s (decoding bech32 failed: invalid separator index -1): invalid address",
			invalidAddress,
		),
		err.Error(),
	)
}
