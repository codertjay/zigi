package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "zigchain/testutil/keeper"
	"zigchain/testutil/sample"
	"zigchain/x/tokenwrapper/keeper"
	"zigchain/x/tokenwrapper/types"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"zigchain/app"
)

// Positive test cases

func TestModuleWalletBalancesQuery_Valid(t *testing.T) {
	// Test case: querying the module wallet balances

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

	// Generate and add a pauser address
	pauser := sample.AccAddress()
	k.AddPauserAddress(ctx, pauser)

	// fund the module wallet first
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

	// update IBC settings
	// build a valid update message
	updateMsg := &types.MsgUpdateIbcSettings{
		Signer:               signer.String(),
		NativeClientId:       "client-01",
		CounterpartyClientId: "client-02",
		NativePort:           "transfer_01",
		CounterpartyPort:     "transfer_02",
		NativeChannel:        "channel-1",
		CounterpartyChannel:  "channel-2",
		Denom:                "uzig",
		DecimalDifference:    12,
	}

	// call the UpdateIbcSettings method
	updateResp, err := ms.UpdateIbcSettings(ctx, updateMsg)
	// check if the response is not nil and no error occurred
	require.NoError(t, err)
	require.NotNil(t, updateResp)

	// verify state
	require.Equal(t, updateMsg.NativeClientId, k.GetNativeClientId(ctx))
	require.Equal(t, updateMsg.CounterpartyClientId, k.GetCounterpartyClientId(ctx))
	require.Equal(t, updateMsg.NativePort, k.GetNativePort(ctx))
	require.Equal(t, updateMsg.CounterpartyPort, k.GetCounterpartyPort(ctx))
	require.Equal(t, updateMsg.NativeChannel, k.GetNativeChannel(ctx))
	require.Equal(t, updateMsg.CounterpartyChannel, k.GetCounterpartyChannel(ctx))
	require.Equal(t, updateMsg.Denom, k.GetDenom(ctx))
	require.Equal(t, updateMsg.DecimalDifference, k.GetDecimalDifference(ctx))

	// call the ModuleInfo method
	response, err := k.ModuleInfo(ctx, &types.QueryModuleInfoRequest{})
	require.NoError(t, err)

	require.Equal(t, &types.QueryModuleInfoResponse{
		ModuleAddress:           moduleAddr.String(),
		Balances:                sdk.NewCoins(moduleBalance),
		OperatorAddress:         signer.String(),
		ProposedOperatorAddress: sample.ZeroAccAddress(),
		TokenWrapperEnabled:     false,
		NativeClientId:          "client-01",
		CounterpartyClientId:    "client-02",
		NativePort:              "transfer_01",
		CounterpartyPort:        "transfer_02",
		NativeChannel:           "channel-1",
		CounterpartyChannel:     "channel-2",
		Denom:                   "uzig",
		DecimalDifference:       12,
		PauserAddresses:         []string{pauser},
	}, response)
}

func TestModuleWalletBalancesQuery_TokenWrapperEnabled(t *testing.T) {
	// Test case: querying the module wallet balances if TokenWrapper is enabled

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

	// verify module account balance
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	moduleBalance := testApp.BankKeeper.GetBalance(ctx, moduleAddr, "uzig")
	require.Equal(t, sdkmath.NewInt(0), moduleBalance.Amount)

	// verify sender's balance was reduced
	signerBalance := testApp.BankKeeper.GetBalance(ctx, signer, "uzig")
	require.Equal(t, sdkmath.NewInt(1000), signerBalance.Amount)

	// update IBC settings
	// build a valid update message
	updateMsg := &types.MsgUpdateIbcSettings{
		Signer:               signer.String(),
		NativeClientId:       "client-01",
		CounterpartyClientId: "client-02",
		NativePort:           "transfer_01",
		CounterpartyPort:     "transfer_02",
		NativeChannel:        "channel-1",
		CounterpartyChannel:  "channel-2",
		Denom:                "uzig",
		DecimalDifference:    12,
	}

	// call the UpdateIbcSettings method
	updateResp, err := ms.UpdateIbcSettings(ctx, updateMsg)
	// check if the response is not nil and no error occurred
	require.NoError(t, err)
	require.NotNil(t, updateResp)

	// verify state
	require.Equal(t, updateMsg.NativeClientId, k.GetNativeClientId(ctx))
	require.Equal(t, updateMsg.CounterpartyClientId, k.GetCounterpartyClientId(ctx))
	require.Equal(t, updateMsg.NativePort, k.GetNativePort(ctx))
	require.Equal(t, updateMsg.CounterpartyPort, k.GetCounterpartyPort(ctx))
	require.Equal(t, updateMsg.NativeChannel, k.GetNativeChannel(ctx))
	require.Equal(t, updateMsg.CounterpartyChannel, k.GetCounterpartyChannel(ctx))
	require.Equal(t, updateMsg.Denom, k.GetDenom(ctx))
	require.Equal(t, updateMsg.DecimalDifference, k.GetDecimalDifference(ctx))

	// create a message to enable the token wrapper
	// with the correct signer
	msg := &types.MsgEnableTokenWrapper{
		Signer: signer.String(),
	}

	// call the EnableTokenWrapper method
	resp, err := ms.EnableTokenWrapper(ctx, msg)
	// verify that the handler executed without any errors
	require.NoError(t, err)
	// confirm that a valid response was returned
	require.NotNil(t, resp)

	// check that the token wrapper is now enabled
	enabled := k.IsEnabled(ctx)
	require.True(t, enabled)

	// call the ModuleInfo method
	response, err := k.ModuleInfo(ctx, &types.QueryModuleInfoRequest{})
	require.NoError(t, err)

	require.Equal(t, &types.QueryModuleInfoResponse{
		ModuleAddress:           moduleAddr.String(),
		Balances:                sdk.NewCoins(moduleBalance),
		OperatorAddress:         signer.String(),
		ProposedOperatorAddress: sample.ZeroAccAddress(),
		PauserAddresses:         []string{},
		TokenWrapperEnabled:     true,
		NativeClientId:          "client-01",
		CounterpartyClientId:    "client-02",
		NativePort:              "transfer_01",
		CounterpartyPort:        "transfer_02",
		NativeChannel:           "channel-1",
		CounterpartyChannel:     "channel-2",
		Denom:                   "uzig",
		DecimalDifference:       12,
	}, response)
}

// Negative test cases

func TestModuleInfoQuery_InvalidRequest(t *testing.T) {
	// Test case: request is nil

	k, ctx := keepertest.TokenwrapperKeeper(t, nil)

	// calling ModuleInfo with nil request
	resp, err := k.ModuleInfo(ctx, nil)

	// assert error returned
	require.Error(t, err)
	require.Nil(t, resp)

	// assert gRPC status
	statusErr, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, statusErr.Code())
	require.Equal(t, "invalid request", statusErr.Message())
}
