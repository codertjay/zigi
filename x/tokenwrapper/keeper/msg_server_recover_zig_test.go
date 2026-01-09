package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"

	"zigchain/app"
	"zigchain/testutil/sample"
	"zigchain/x/tokenwrapper/keeper"
	"zigchain/x/tokenwrapper/types"
	"zigchain/zutils/constants"
)

// Positive test cases

func TestMsgRecoverZig_Valid(t *testing.T) {
	// Test case: fund module wallet with valid parameters

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate a test account with some coins
	creator := sample.AccAddress()
	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// set the operator address
	k.SetDenom(ctx, "unit-zig")
	_ = k.SetDecimalDifference(ctx, 12)
	k.SetEnabled(ctx, true)

	// get the recv IBC denom
	ibcDenom := k.GetIBCRecvDenom(ctx, k.GetDenom(ctx))

	ibcAmount, _ := sdkmath.NewIntFromString("1000000000000000000")
	nativeAmount, _ := sdkmath.NewIntFromString("1000000")

	ibcCoins := sdk.NewCoins(sdk.NewCoin(ibcDenom, ibcAmount))
	nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, nativeAmount))

	coins := sdk.NewCoins(ibcCoins[0], nativeCoins[0])

	// mint coins to the mint module
	require.NoError(t, testApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins))

	// send coins from mint module to the test account
	require.NoError(t, testApp.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, signer, ibcCoins))

	// send coins from mint to tokenwrapper module
	require.NoError(t, testApp.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, nativeCoins))

	// verify module account pre-recovery
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	moduleBalance := testApp.BankKeeper.GetBalance(ctx, moduleAddr, ibcDenom)
	require.Equal(t, sdkmath.ZeroInt(), moduleBalance.Amount)
	moduleBalance = testApp.BankKeeper.GetBalance(ctx, moduleAddr, constants.BondDenom)
	require.Equal(t, nativeAmount, moduleBalance.Amount)

	// verify sender account balance pre-recovery
	signerBalance := testApp.BankKeeper.GetBalance(ctx, signer, ibcDenom)
	require.Equal(t, ibcAmount, signerBalance.Amount)
	signerBalance = testApp.BankKeeper.GetBalance(ctx, signer, constants.BondDenom)
	require.Equal(t, sdkmath.ZeroInt(), signerBalance.Amount)

	// create a message to recover ZIG
	msg := &types.MsgRecoverZig{
		Signer:  signer.String(),
		Address: signer.String(),
	}

	// call the FundModuleWallet method
	resp, err := ms.RecoverZig(ctx, msg)
	// check if the response is not nil and no error occurred
	require.NoError(t, err)
	require.NotNil(t, resp)

	// verify response fields
	require.Equal(t, msg.Signer, resp.ReceivingAddress)
	require.Equal(t, msg.Address, resp.ReceivingAddress)
	require.Equal(t, ibcCoins[0], resp.LockedIbcAmount)
	require.Equal(t, nativeCoins[0], resp.UnlockedNativeAmount)

	// verify module account balance post-recovery
	moduleBalance = testApp.BankKeeper.GetBalance(ctx, moduleAddr, ibcDenom)
	require.Equal(t, ibcAmount, moduleBalance.Amount)
	moduleBalance = testApp.BankKeeper.GetBalance(ctx, moduleAddr, constants.BondDenom)
	require.Equal(t, sdkmath.ZeroInt(), moduleBalance.Amount)

	// verify sender's balance post-recovery
	signerBalance = testApp.BankKeeper.GetBalance(ctx, signer, ibcDenom)
	require.Equal(t, sdkmath.ZeroInt(), signerBalance.Amount)
	signerBalance = testApp.BankKeeper.GetBalance(ctx, signer, constants.BondDenom)
	require.Equal(t, nativeAmount, signerBalance.Amount)

	// CHECK EVENTS
	// ---------------------------------

	// check that the event was emitted
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// find the wallet module funded event
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventTypeAddressZigRecovered {
			foundEvent = true
			// check event attributes
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeySigner {
					require.Equal(t, signer.String(), string(attr.Value))
				}
			}
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyAddress {
					require.Equal(t, signer.String(), string(attr.Value))
				}
			}
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyLockedIbcAmount {
					require.Equal(t, ibcCoins[0].String(), string(attr.Value))
				}
			}
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyUnlockedNativeAmount {
					require.Equal(t, nativeCoins[0].String(), string(attr.Value))
				}
			}
			break
		}
	}
	require.True(t, foundEvent, "Expected EventTypeAddressZigRecovered event to be emitted")
}

// Negative test cases

func TestMsgRecoverZig_ZeroAmount(t *testing.T) {
	// Test case: recover ZIG with zero amount

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate a test account with some coins
	creator := sample.AccAddress()
	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// set the operator address
	k.SetDenom(ctx, "unit-zig")
	_ = k.SetDecimalDifference(ctx, 12)
	k.SetEnabled(ctx, true)

	// get the recv IBC denom
	ibcDenom := k.GetIBCRecvDenom(ctx, k.GetDenom(ctx))

	ibcAmount, _ := sdkmath.NewIntFromString("0")

	ibcCoins := sdk.NewCoins(sdk.NewCoin(ibcDenom, ibcAmount))

	// mint coins to the mint module
	require.NoError(t, testApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, ibcCoins))
	// send coins from mint module to the test account
	require.NoError(t, testApp.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, signer, ibcCoins))

	// create a message to fund the module wallet - zero amount
	msg := &types.MsgRecoverZig{
		Signer:  signer.String(),
		Address: signer.String(),
	}

	// call the RecoverZig method
	resp, err := ms.RecoverZig(ctx, msg)
	// check if the response is not nil and no error occurred
	require.EqualError(t, err, "no IBC vouchers available in address: no ibc vouchers available in address")
	require.Nil(t, resp)
}

func TestMsgRecoverZig_NegativeAmount(t *testing.T) {
	// Test case: recover ZIG with negative amount

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// set the operator address
	k.SetDenom(ctx, "unit-zig")
	_ = k.SetDecimalDifference(ctx, 12)
	k.SetEnabled(ctx, true)

	// get the recv IBC denom
	ibcDenom := k.GetIBCRecvDenom(ctx, k.GetDenom(ctx))

	ibcAmount, _ := sdkmath.NewIntFromString("-1000000000000000000")

	// this should panic because the amount is negative
	require.Panics(t, func() {
		_ = sdk.NewCoins(sdk.NewCoin(ibcDenom, ibcAmount))
	}, "expected panic because the amount is negative")
}

func TestMsgRecoverZig_InvalidSigner(t *testing.T) {
	// Test case: recover ZIG with invalid signer address

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// set the operator address
	k.SetDenom(ctx, "unit-zig")
	_ = k.SetDecimalDifference(ctx, 12)
	k.SetEnabled(ctx, true)

	// create a message with an invalid Bech32 signer address
	invalidSigner := "invalid_bech32_address"
	msg := &types.MsgRecoverZig{
		Signer:  invalidSigner,
		Address: sample.AccAddress(), // Valid address for the other field
	}

	// call the RecoverZig method
	resp, err := ms.RecoverZig(ctx, msg)

	// check that an error is returned and the response is nil
	require.Error(t, err, "expected an error due to invalid signer address")
	require.Contains(t, err.Error(), "decoding bech32 failed", "expected error to indicate invalid Bech32 address")
	require.Nil(t, resp, "expected nil response due to invalid signer address")
}

func TestMsgRecoverZig_InvalidAddress(t *testing.T) {
	// Test case: recover ZIG with invalid address

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// set the operator address
	k.SetDenom(ctx, "unit-zig")
	_ = k.SetDecimalDifference(ctx, 12)
	k.SetEnabled(ctx, true)

	// generate a valid test account for the signer
	signer := sample.AccAddress()

	// create a message with an invalid Bech32 address
	invalidAddress := "invalid_bech32_address"
	msg := &types.MsgRecoverZig{
		Signer:  signer,
		Address: invalidAddress,
	}

	// call the RecoverZig method
	resp, err := ms.RecoverZig(ctx, msg)

	// check that an error is returned and the response is nil
	require.Error(t, err, "expected an error due to invalid address")
	require.Contains(t, err.Error(), "decoding bech32 failed", "expected error to indicate invalid Bech32 address")
	require.Nil(t, resp, "expected nil response due to invalid address")
}

func TestMsgRecoverZig_ModuleDisabled(t *testing.T) {
	// Test case: recover ZIG when module is disabled

	// initialize test blockchain app
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate a test account with some coins
	signer := sample.AccAddress()

	// set module parameters, but keep module disabled
	k.SetDenom(ctx, "unit-zig")
	_ = k.SetDecimalDifference(ctx, 12)
	k.SetEnabled(ctx, false) // Module is disabled

	// create a message to recover ZIG
	msg := &types.MsgRecoverZig{
		Signer:  signer,
		Address: signer,
	}

	// call the RecoverZig method
	resp, err := ms.RecoverZig(ctx, msg)

	// check that an error is returned and the response is nil
	require.Error(t, err, "expected an error due to disabled module")
	require.Contains(t, err.Error(), "module disabled: module functionality is not enabled", "expected error to indicate module is disabled")
	require.Nil(t, resp, "expected nil response due to disabled module")
}

func TestMsgRecoverZig_UnsetDenom(t *testing.T) {
	// Test case: recover ZIG with unset denom

	// initialize test blockchain app
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate a test account with some coins
	signer := sample.AccAddress()

	// set module parameters, but do not set denom (leave it empty)
	_ = k.SetDecimalDifference(ctx, 12)
	k.SetEnabled(ctx, true)

	// create a message to recover ZIG
	msg := &types.MsgRecoverZig{
		Signer:  signer,
		Address: signer,
	}

	// call the RecoverZig method
	resp, err := ms.RecoverZig(ctx, msg)

	// check that an error is returned and the response is nil
	require.Error(t, err, "expected an error due to unset denom")
	require.Contains(t, err.Error(), "no IBC vouchers available in address", "expected error to indicate unset denom")
	require.Nil(t, resp, "expected nil response due to unset denom")
}

func TestMsgRecoverZig_InsufficientNativeTokens(t *testing.T) {
	// Test case: recover ZIG with insufficient native tokens in module

	// initialize test blockchain app
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate a test account with some coins
	signer := sdk.MustAccAddressFromBech32(sample.AccAddress())

	// set module parameters
	k.SetDenom(ctx, "unit-zig")
	_ = k.SetDecimalDifference(ctx, 12)
	k.SetEnabled(ctx, true)

	// get the recv IBC denom
	ibcDenom := k.GetIBCRecvDenom(ctx, k.GetDenom(ctx))

	// mint IBC coins to the test account
	ibcAmount, _ := sdkmath.NewIntFromString("1000000000000000000")
	ibcCoins := sdk.NewCoins(sdk.NewCoin(ibcDenom, ibcAmount))
	require.NoError(t, testApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, ibcCoins))
	require.NoError(t, testApp.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, signer, ibcCoins))

	// do not send native coins to the module (simulate insufficient native tokens)
	// module has zero native tokens

	// create a message to recover ZIG
	msg := &types.MsgRecoverZig{
		Signer:  signer.String(),
		Address: signer.String(),
	}

	// call the RecoverZig method
	resp, err := ms.RecoverZig(ctx, msg)

	// check that an error is returned and the response is nil
	require.Error(t, err, "expected an error due to insufficient native tokens in module")
	require.Contains(t, err.Error(), "module does not have enough balance of 1000000uzig", "expected error to indicate insufficient native tokens")
	require.Nil(t, resp, "expected nil response due to insufficient native tokens")
}

func TestMsgRecoverZig_OperatorAddress(t *testing.T) {
	// Test case: recover ZIG using operator address (should fail)

	// initialize test blockchain app
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate a test account for the operator
	operator := sample.AccAddress()
	operatorAddr := sdk.MustAccAddressFromBech32(operator)

	// set the operator address
	k.SetOperatorAddress(ctx, operator)
	k.SetDenom(ctx, "unit-zig")
	_ = k.SetDecimalDifference(ctx, 12)
	k.SetEnabled(ctx, true)

	// get the recv IBC denom
	ibcDenom := k.GetIBCRecvDenom(ctx, k.GetDenom(ctx))

	ibcAmount, _ := sdkmath.NewIntFromString("1000000000000000000")
	nativeAmount, _ := sdkmath.NewIntFromString("1000000")

	ibcCoins := sdk.NewCoins(sdk.NewCoin(ibcDenom, ibcAmount))
	nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, nativeAmount))

	coins := sdk.NewCoins(ibcCoins[0], nativeCoins[0])

	// mint coins to the mint module
	require.NoError(t, testApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins))

	// send coins from mint module to the operator account
	require.NoError(t, testApp.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, operatorAddr, ibcCoins))

	// send coins from mint to tokenwrapper module
	require.NoError(t, testApp.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, nativeCoins))

	// create a message to recover ZIG using the operator address
	msg := &types.MsgRecoverZig{
		Signer:  operator,
		Address: operator, // Using operator address should fail
	}

	// call the RecoverZig method
	resp, err := ms.RecoverZig(ctx, msg)

	// check that an error is returned and the response is nil
	require.Error(t, err, "expected an error due to using operator address")
	require.Contains(t, err.Error(), "recovery not allowed on operator address", "expected error to indicate recovery not allowed on operator address")
	require.Nil(t, resp, "expected nil response due to using operator address")

	// CHECK EVENTS
	// ---------------------------------

	// check that the error event was emitted
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// find the error event
	var foundErrorEvent bool
	for _, event := range events {
		if event.Type == types.EventTypeTokenWrapperError {
			foundErrorEvent = true
			// check event attributes
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyError {
					require.Contains(t, string(attr.Value), "recovery not allowed on operator address", "expected error message in event")
				}
			}
			break
		}
	}
	require.True(t, foundErrorEvent, "Expected EventTypeTokenWrapperError event to be emitted")
}
