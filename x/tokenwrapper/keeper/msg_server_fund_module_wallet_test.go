package keeper_test

import (
	"fmt"
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

func TestMsgFundModuleWallet_Valid(t *testing.T) {
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
	initialCoins := sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(1000)))

	// mint coins to the mint module
	require.NoError(t, testApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initialCoins))
	// send coins from mint module to the test account
	require.NoError(t, testApp.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, signer, initialCoins))

	// set the operator address
	k.SetOperatorAddress(ctx, signer.String())

	// create a message to fund the module wallet
	amount := sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(100)))
	msg := &types.MsgFundModuleWallet{
		Signer: signer.String(),
		Amount: amount,
	}

	// call the FundModuleWallet method
	resp, err := ms.FundModuleWallet(ctx, msg)
	// check if the response is not nil and no error occurred
	require.NoError(t, err)
	require.NotNil(t, resp)

	// verify response fields
	require.Equal(t, msg.Signer, resp.Signer)
	require.Equal(t, msg.Amount, resp.Amount)

	// verify module account balance
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	moduleBalance := testApp.BankKeeper.GetBalance(ctx, moduleAddr, "uzig")
	require.Equal(t, sdkmath.NewInt(100), moduleBalance.Amount)

	// verify sender's balance was reduced
	signerBalance := testApp.BankKeeper.GetBalance(ctx, signer, "uzig")
	require.Equal(t, sdkmath.NewInt(900), signerBalance.Amount)

	// CHECK EVENTS
	// ---------------------------------

	// check that the event was emitted
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// find the wallet module funded event
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventTypeModuleWalletFunded {
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
					require.Equal(t, "100uzig", string(attr.Value))
				}
			}
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyBalances {
					require.Equal(t, "100uzig", string(attr.Value))
				}
			}
			break
		}
	}
	require.True(t, foundEvent, "Expected EventTypeTokenWrapperDisabled event to be emitted")
}

func TestMsgFundModuleWallet_ZeroAmount(t *testing.T) {
	// Test case: fund module wallet with zero amount

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

	// create a message to fund the module wallet - zero amount
	msg := &types.MsgFundModuleWallet{
		Signer: signer.String(),
		Amount: sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(0))),
	}

	// call the FundModuleWallet method
	resp, err := ms.FundModuleWallet(ctx, msg)
	// check if the response is not nil and no error occurred
	require.NoError(t, err)
	require.NotNil(t, resp)

	// verify response fields
	require.Equal(t, msg.Signer, resp.Signer)
	require.Equal(t, msg.Amount, resp.Amount)

	// verify module account balance is zero
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	moduleBalance := testApp.BankKeeper.GetBalance(ctx, moduleAddr, "uzig")
	require.Equal(t, sdkmath.NewInt(0), moduleBalance.Amount)

	// verify sender's balance has not changed
	signerBalance := testApp.BankKeeper.GetBalance(ctx, signer, "uzig")
	require.Equal(t, sdkmath.NewInt(1000), signerBalance.Amount)
}

// Negative test cases

func TestMsgFundModuleWallet_EmptySignerAddress(t *testing.T) {
	// Test case: try to fund module wallet if the signer is empty

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

	// create a message to fund the module wallet - empty signer
	msg := &types.MsgFundModuleWallet{
		Signer: "",
		Amount: sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(100))),
	}

	// call the FundModuleWallet method
	resp, err := ms.FundModuleWallet(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)
	// check if the error message contains the expected string
	require.Contains(t, err.Error(), "empty address string is not allowed")
}

func TestMsgFundModuleWallet_SignerAddressTooShort(t *testing.T) {
	// Test case: try to fund module wallet if the signer address is too short

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

	// create a message to fund the module wallet - too short signer address
	msg := &types.MsgFundModuleWallet{
		Signer: "a",
		Amount: sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(100))),
	}

	// call the FundModuleWallet method
	resp, err := ms.FundModuleWallet(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)
	// check if the error message contains the expected string
	require.Contains(t, err.Error(), "decoding bech32 failed: invalid bech32 string length 1")
}

func TestMsgFundModuleWallet_SignerAddressTooLong(t *testing.T) {
	// Test case: try to fund module wallet if the signer address is too long

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

	// create a message to fund the module wallet - too long signer address
	msg := &types.MsgFundModuleWallet{
		// MaxAddrLen = 255
		Signer: "zig1" + strings.Repeat("a", 100),
		Amount: sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(100))),
	}

	// call the FundModuleWallet method
	resp, err := ms.FundModuleWallet(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)
	// check if the error message contains the expected string
	require.Contains(t, err.Error(), "decoding bech32 failed: invalid checksum")
}

func TestMsgFundModuleWallet_SignerAddressBadChars(t *testing.T) {
	// Test case: try to fund module wallet if the signer address contains bad characters

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

	// create a message to fund the module wallet - bad characters in signer address
	msg := &types.MsgFundModuleWallet{
		// MaxAddrLen = 255
		Signer: "zig1/\\.%&?32njzt23c86en7hd8tajma79tt3",
		Amount: sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(100))),
	}

	// call the FundModuleWallet method
	resp, err := ms.FundModuleWallet(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)
	// check if the error message contains the expected string
	require.Contains(t, err.Error(), "decoding bech32 failed: invalid character not part of charset")
}

func TestMsgFundModuleWallet_SignerNotCurrentOperator(t *testing.T) {
	// Test case: try to fund module wallet if the signer is not the current operator

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

	// create a message to fund the module wallet
	msg := &types.MsgFundModuleWallet{
		Signer: sample.AccAddress(), // generate address different from the operator address
		Amount: sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(100))),
	}

	// call the FundModuleWallet method
	resp, err := ms.FundModuleWallet(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)
	// check if the error message contains the expected string
	require.Contains(t, err.Error(), "only the current operator can fund the module wallet")
}

func TestMsgFundModuleWallet_InsufficientBalance(t *testing.T) {
	// Test case: try to fund module wallet if the sender has insufficient balance

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
	initialCoins := sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(500)))

	// mint coins to the mint module
	require.NoError(t, testApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initialCoins))
	// send coins from mint module to the test account
	require.NoError(t, testApp.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, signer, initialCoins))

	// set the operator address
	k.SetOperatorAddress(ctx, signer.String())

	// create a message to fund the module wallet
	msg := &types.MsgFundModuleWallet{
		Signer: signer.String(),
		Amount: sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(1000))),
	}

	// call the FundModuleWallet method
	resp, err := ms.FundModuleWallet(ctx, msg)
	require.Error(t, err)
	require.Nil(t, resp)

	// check if the error message is correct
	require.Equal(t,
		fmt.Sprintf(
			"account %s does not have enough balance of 1000uzig",
			signer,
		),
		err.Error(),
	)

	// verify module account balance
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	moduleBalance := testApp.BankKeeper.GetBalance(ctx, moduleAddr, "uzig")
	require.Equal(t, sdkmath.NewInt(0), moduleBalance.Amount)

	// verify sender's balance has not changed
	signerBalance := testApp.BankKeeper.GetBalance(ctx, signer, "uzig")
	require.Equal(t, sdkmath.NewInt(500), signerBalance.Amount)
}

func TestMsgFundModuleWallet_DenomNotFunded(t *testing.T) {
	// Test case: try to fund module wallet if the denom is not funded

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

	// create a message to fund the module wallet
	msg := &types.MsgFundModuleWallet{
		Signer: signer.String(),
		Amount: sdk.NewCoins(sdk.NewCoin("abc", sdkmath.NewInt(1000))),
	}

	// call the FundModuleWallet method
	resp, err := ms.FundModuleWallet(ctx, msg)
	require.Error(t, err)
	require.Nil(t, resp)

	// check if the error message is correct
	require.Equal(t,
		fmt.Sprintf(
			"account %s does not have enough balance of 1000abc",
			signer,
		),
		err.Error(),
	)
}
