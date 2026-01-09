package keeper_test

import (
	"strings"
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
)

// Positive test cases

func TestMsgWithdrawFromModuleWallet_Valid(t *testing.T) {
	// Test case: withdraw from module wallet with valid parameters

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
	initialCoins := sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(1000)))

	// mint coins to the mint module
	require.NoError(t, testApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initialCoins))
	// send coins from mint module to the test account
	require.NoError(t, testApp.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, signer, initialCoins))

	// set the operator address
	k.SetOperatorAddress(ctx, signer.String())

	// fund module wallet first

	// create a message to fund the module wallet
	fundMsg := &types.MsgFundModuleWallet{
		Signer: signer.String(),
		Amount: sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(100))),
	}

	// call the FundModuleWallet method
	fundResp, err := ms.FundModuleWallet(ctx, fundMsg)
	// check if the response is not nil and no error occurred
	require.NoError(t, err)
	require.NotNil(t, fundResp)

	// verify response fields
	require.Equal(t, fundMsg.Signer, fundResp.Signer)
	require.Equal(t, fundMsg.Amount, fundResp.Amount)

	// verify module account balance
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	moduleBalance := testApp.BankKeeper.GetBalance(ctx, moduleAddr, "uzig")
	require.Equal(t, sdkmath.NewInt(100), moduleBalance.Amount)

	// verify sender's balance was reduced
	signerBalance := testApp.BankKeeper.GetBalance(ctx, signer, "uzig")
	require.Equal(t, sdkmath.NewInt(900), signerBalance.Amount)

	// create a message to withdraw from module wallet
	withdrawMsg := &types.MsgWithdrawFromModuleWallet{
		Signer: signer.String(),
		Amount: sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(50))),
	}

	// call the WithdrawFromModuleWallet method
	withdrawResp, err := ms.WithdrawFromModuleWallet(ctx, withdrawMsg)
	// check if the response is not nil and no error occurred
	require.NoError(t, err)
	require.NotNil(t, withdrawResp)

	// verify response fields
	require.Equal(t, withdrawMsg.Signer, withdrawResp.Signer)
	require.Equal(t, withdrawMsg.Amount, withdrawResp.Amount)

	// verify module account balance is reduced
	newModuleBalance := testApp.BankKeeper.GetBalance(ctx, moduleAddr, "uzig")
	require.Equal(t, sdkmath.NewInt(50), newModuleBalance.Amount)

	// verify sender's balance is increased
	newSignerBalance := testApp.BankKeeper.GetBalance(ctx, signer, "uzig")
	require.Equal(t, sdkmath.NewInt(950), newSignerBalance.Amount)

	// CHECK EVENTS
	// ---------------------------------

	// check that the event was emitted
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// find the pauser address added event
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventTypeModuleWalletWithdrawn {
			foundEvent = true
			// check event attributes
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeySigner {
					require.Equal(t, signer.String(), string(attr.Value))
				}
			}
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyModuleAddress {
					require.Equal(t, moduleAddr.String(), string(attr.Value))
				}
			}
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyAmount {
					require.Equal(t, "50uzig", string(attr.Value))
				}
			}
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyBalances {
					require.Equal(t, "50uzig", string(attr.Value))
				}
			}
			break
		}
	}
	require.True(t, foundEvent, "Expected EventTypeTokenWrapperDisabled event to be emitted")
}

func TestMsgWithdrawFromModuleWallet_ZeroAmount(t *testing.T) {
	// Test case: withdraw from module wallet with zero amount

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
	initialCoins := sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(1000)))

	// mint coins to the mint module
	require.NoError(t, testApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initialCoins))
	// send coins from mint module to the test account
	require.NoError(t, testApp.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, signer, initialCoins))

	// set the operator address
	k.SetOperatorAddress(ctx, signer.String())

	// fund module wallet first

	// create a message to fund the module wallet
	fundMsg := &types.MsgFundModuleWallet{
		Signer: signer.String(),
		Amount: sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(100))),
	}

	// call the FundModuleWallet method
	fundResp, err := ms.FundModuleWallet(ctx, fundMsg)
	// check if the response is not nil and no error occurred
	require.NoError(t, err)
	require.NotNil(t, fundResp)

	// verify response fields
	require.Equal(t, fundMsg.Signer, fundResp.Signer)
	require.Equal(t, fundMsg.Amount, fundResp.Amount)

	// verify module account balance
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	moduleBalance := testApp.BankKeeper.GetBalance(ctx, moduleAddr, "uzig")
	require.Equal(t, sdkmath.NewInt(100), moduleBalance.Amount)

	// verify sender's balance was reduced
	signerBalance := testApp.BankKeeper.GetBalance(ctx, signer, "uzig")
	require.Equal(t, sdkmath.NewInt(900), signerBalance.Amount)

	// create a message to withdraw from module wallet
	withdrawMsg := &types.MsgWithdrawFromModuleWallet{
		Signer: signer.String(),
		Amount: sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(0))),
	}

	// call the WithdrawFromModuleWallet method
	withdrawResp, err := ms.WithdrawFromModuleWallet(ctx, withdrawMsg)
	// check if the response is not nil and no error occurred
	require.NoError(t, err)
	require.NotNil(t, withdrawResp)

	// verify response fields
	require.Equal(t, withdrawMsg.Signer, withdrawResp.Signer)
	require.Equal(t, withdrawMsg.Amount, withdrawResp.Amount)

	// verify module account balance has not changed
	newModuleBalance := testApp.BankKeeper.GetBalance(ctx, moduleAddr, "uzig")
	require.Equal(t, sdkmath.NewInt(100), newModuleBalance.Amount)

	// verify sender's balance has not changed
	newSignerBalance := testApp.BankKeeper.GetBalance(ctx, signer, "uzig")
	require.Equal(t, sdkmath.NewInt(900), newSignerBalance.Amount)
}

// Negative test cases

func TestMsgWithdrawFromModuleWallet_EmptySignerAddress(t *testing.T) {
	// Test case: try to withdraw from module wallet if the signer is empty

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
	k.SetOperatorAddress(ctx, sample.AccAddress())

	// create a message to withdraw from module wallet - empty signer
	msg := &types.MsgWithdrawFromModuleWallet{
		Signer: "",
		Amount: sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(100))),
	}

	// call the WithdrawFromModuleWallet method
	resp, err := ms.WithdrawFromModuleWallet(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)
	// check if the error message contains the expected string
	require.Contains(t, err.Error(), "empty address string is not allowed")
}

func TestMsgWithdrawFromModuleWallet_SignerAddressTooShort(t *testing.T) {
	// Test case: try to withdraw from module wallet if the signer address is too short

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
	k.SetOperatorAddress(ctx, sample.AccAddress())

	// create a message to withdraw from module wallet - too short signer address
	msg := &types.MsgWithdrawFromModuleWallet{
		Signer: "a",
		Amount: sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(100))),
	}

	// call the WithdrawFromModuleWallet method
	resp, err := ms.WithdrawFromModuleWallet(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)
	// check if the error message contains the expected string
	require.Contains(t, err.Error(), "decoding bech32 failed: invalid bech32 string length 1")
}

func TestMsgWithdrawFromModuleWallet_SignerAddressTooLong(t *testing.T) {
	// Test case: try to withdraw from module wallet if the signer address is too long

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
	k.SetOperatorAddress(ctx, sample.AccAddress())

	// create a message to withdraw from module wallet - too long signer address
	msg := &types.MsgWithdrawFromModuleWallet{
		// MaxAddrLen = 255
		Signer: "zig1" + strings.Repeat("a", 100),
		Amount: sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(100))),
	}

	// call the WithdrawFromModuleWallet method
	resp, err := ms.WithdrawFromModuleWallet(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)
	// check if the error message contains the expected string
	require.Contains(t, err.Error(), "decoding bech32 failed: invalid checksum")
}

func TestMsgWithdrawFromModuleWallet_SignerAddressBadChars(t *testing.T) {
	// Test case: try to withdraw from module wallet if the signer address contains bad characters

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
	k.SetOperatorAddress(ctx, sample.AccAddress())

	// create a message to withdraw from module wallet - bad characters in signer address
	msg := &types.MsgWithdrawFromModuleWallet{
		// MaxAddrLen = 255
		Signer: "zig1/\\.%&?32njzt23c86en7hd8tajma79tt3",
		Amount: sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(100))),
	}

	// call the WithdrawFromModuleWallet method
	resp, err := ms.WithdrawFromModuleWallet(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)
	// check if the error message contains the expected string
	require.Contains(t, err.Error(), "decoding bech32 failed: invalid character not part of charset")
}

func TestMsgWithdrawFromModuleWallet_SignerNotCurrentOperator(t *testing.T) {
	// Test case: try to withdraw from module wallet if the signer is not the current operator

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
	k.SetOperatorAddress(ctx, sample.AccAddress())

	// create a message to withdraw from module wallet
	msg := &types.MsgWithdrawFromModuleWallet{
		Signer: sample.AccAddress(), // generate address different from the operator address
		Amount: sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(100))),
	}

	// call the WithdrawFromModuleWallet method
	resp, err := ms.WithdrawFromModuleWallet(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)
	// check if the error message contains the expected string
	require.Contains(t, err.Error(), "only the current operator can withdraw from the module wallet")
}

func TestMsgWithdrawFromModuleWallet_InsufficientBalance(t *testing.T) {
	// Test case: try to withdraw from module wallet if the sender has insufficient balance

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
	k.SetOperatorAddress(ctx, signer.String())

	// create a message to withdraw from module wallet
	msg := &types.MsgWithdrawFromModuleWallet{
		Signer: signer.String(),
		Amount: sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(1000))),
	}

	// call the WithdrawFromModuleWallet method
	resp, err := ms.WithdrawFromModuleWallet(ctx, msg)
	require.Error(t, err)
	require.Nil(t, resp)

	// check if the error message is correct
	require.Equal(t, "module does not have enough balance of 1000uzig", err.Error())

	// verify module account balance
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	moduleBalance := testApp.BankKeeper.GetBalance(ctx, moduleAddr, "uzig")
	require.Equal(t, sdkmath.NewInt(0), moduleBalance.Amount)
}

func TestMsgWithdrawFromModuleWallet_DenomNotFunded(t *testing.T) {
	// Test case: try to withdraw from module wallet if the denom is not funded

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
	k.SetOperatorAddress(ctx, signer.String())

	// create a message to withdraw from module wallet
	msg := &types.MsgWithdrawFromModuleWallet{
		Signer: signer.String(),
		Amount: sdk.NewCoins(sdk.NewCoin("abc", sdkmath.NewInt(1000))),
	}

	// call the WithdrawFromModuleWallet method
	resp, err := ms.WithdrawFromModuleWallet(ctx, msg)
	require.Error(t, err)
	require.Nil(t, resp)

	// check if the error message is correct
	require.Equal(t, "module does not have enough balance of 1000abc", err.Error())
}
