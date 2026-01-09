package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"zigchain/app"
	"zigchain/testutil/sample"
	"zigchain/x/tokenwrapper/keeper"
	"zigchain/x/tokenwrapper/types"
)

// Positive test cases

func TestMsgDisableTokenWrapper_ValidDisableWithOperator(t *testing.T) {
	// Test case: disable token wrapper successfully

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate a random address for the signer
	signer := sample.AccAddress()
	// simulate on-chain state by setting the operator address
	// to match the test signer
	k.SetOperatorAddress(ctx, signer)
	// set the token wrapper to enabled state
	k.SetEnabled(ctx, true)

	// create a message to disable the token wrapper
	// with the correct signer - Operator address
	msg := &types.MsgDisableTokenWrapper{
		Signer: signer,
	}

	// call the DisableTokenWrapper method
	resp, err := ms.DisableTokenWrapper(ctx, msg)
	// verify that the handler executed without any errors
	require.NoError(t, err)
	// confirm that a valid response was returned
	require.NotNil(t, resp)
	require.Equal(t, resp, &types.MsgDisableTokenWrapperResponse{
		Signer:  signer,
		Enabled: false,
	})

	// check that the token wrapper is now disabled
	enabled := k.IsEnabled(ctx)
	require.False(t, enabled)

	// CHECK EVENTS
	// ---------------------------------

	// check that the event was emitted
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// find the pauser address added event
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventTypeTokenWrapperDisabled {
			foundEvent = true
			// check event attributes
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeySigner {
					require.Equal(t, signer, string(attr.Value))
				}
			}
			break
		}
	}
	require.True(t, foundEvent, "Expected EventTypeTokenWrapperDisabled event to be emitted")
}

func TestMsgDisableTokenWrapper_ValidDisableWithPauser(t *testing.T) {
	// Test case: disable token wrapper successfully with a pauser

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate a random address for the signer
	signer := sample.AccAddress()

	// simulate on-chain state by setting the signer as a pauser
	k.AddPauserAddress(ctx, signer)
	// set the token wrapper to enabled state
	k.SetEnabled(ctx, true)

	// create a message to disable the token wrapper
	// with the correct signer - Operator address
	msg := &types.MsgDisableTokenWrapper{
		Signer: signer,
	}

	// call the DisableTokenWrapper method
	resp, err := ms.DisableTokenWrapper(ctx, msg)
	// verify that the handler executed without any errors
	require.NoError(t, err)
	// confirm that a valid response was returned
	require.NotNil(t, resp)

	// check that the token wrapper is now disabled
	enabled := k.IsEnabled(ctx)
	require.False(t, enabled)
}

func TestMsgDisableTokenWrapper_AlreadyDisabled(t *testing.T) {
	// Test case: disable token wrapper if already disabled

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate a random address for the signer
	signer := sample.AccAddress()
	// simulate on-chain state by setting the operator address
	// to match the test signer
	k.SetOperatorAddress(ctx, signer)
	// set the token wrapper to disabled state
	k.SetEnabled(ctx, false)

	// create a message to disable the token wrapper
	msg := &types.MsgDisableTokenWrapper{
		Signer: signer,
	}

	// call the DisableTokenWrapper method
	resp, err := ms.DisableTokenWrapper(ctx, msg)
	// verify that the handler executed without any errors
	require.NoError(t, err)
	require.NotNil(t, resp)

	// check that the token wrapper is still disabled
	enabled := k.IsEnabled(ctx)
	require.False(t, enabled)
}

// Negative test cases

func TestMsgDisableTokenWrapper_SignerNotCurrentOperator(t *testing.T) {
	// Test case: try to disable token wrapper if Signer != currentOperator

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate a random address for the signer
	signer := sample.AccAddress()
	// simulate on-chain state by setting the operator address
	// to match the test signer
	k.SetOperatorAddress(ctx, signer)
	// set the token wrapper to enabled state
	k.SetEnabled(ctx, true)

	// create a message to disable the token wrapper
	msg := &types.MsgDisableTokenWrapper{
		Signer: sample.AccAddress(),
	}

	// call the DisableTokenWrapper method
	resp, err := ms.DisableTokenWrapper(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check error msg
	require.Equal(t, err.Error(), "signer is neither a pauser nor the operator: unauthorized")

	// check that the token wrapper is still enabled
	enabled := k.IsEnabled(ctx)
	require.True(t, enabled)
}

func TestMsgDisableTokenWrapper_SignerNotCurrentOperator2(t *testing.T) {
	// Test case: try to disable token wrapper if Signer != currentOperator

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate a random address for the signer
	signer := sample.AccAddress()
	// set the operator address
	k.SetOperatorAddress(ctx, sample.AccAddress())
	// set the token wrapper to enabled state
	k.SetEnabled(ctx, true)

	// create a message to disable the token wrapper
	msg := &types.MsgDisableTokenWrapper{
		Signer: signer,
	}

	// call the DisableTokenWrapper method
	resp, err := ms.DisableTokenWrapper(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check error msg
	require.Equal(t, err.Error(), "signer is neither a pauser nor the operator: unauthorized")

	// check that the token wrapper is still enabled
	enabled := k.IsEnabled(ctx)
	require.True(t, enabled)
}

func TestMsgDisableTokenWrapper_InvalidOperator(t *testing.T) {
	// Test case: try to disable token wrapper if operator address is invalid

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate a random address for the signer
	signer := sample.AccAddress()
	// simulate on-chain state by setting the operator address
	// to match the test signer
	k.SetOperatorAddress(ctx, signer)
	// set the token wrapper to enabled state
	k.SetEnabled(ctx, true)

	// create a message to disable the token wrapper
	msg := &types.MsgDisableTokenWrapper{
		Signer: "zig15yk64u7zc9g9k2yr2wmzeva5qgwxps6y8c2amk",
	}

	// call the DisableTokenWrapper method
	resp, err := ms.DisableTokenWrapper(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check error msg
	require.Equal(t, err.Error(), "signer is neither a pauser nor the operator: unauthorized")

	// check that the token wrapper is still enabled
	enabled := k.IsEnabled(ctx)
	require.True(t, enabled)
}

func TestMsgDisableTokenWrapper_EmptySigner(t *testing.T) {
	// Test case: try to disable token wrapper if signer address is empty

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate a random address for the signer
	signer := sample.AccAddress()
	// simulate on-chain state by setting the operator address
	// to match the test signer
	k.SetOperatorAddress(ctx, signer)
	// set the token wrapper to enabled state
	k.SetEnabled(ctx, true)

	// create a message to disable the token wrapper
	msg := &types.MsgDisableTokenWrapper{
		Signer: "",
	}

	// call the DisableTokenWrapper method
	resp, err := ms.DisableTokenWrapper(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check error msg
	require.Equal(t, err.Error(), "empty address string is not allowed")

	// check that the token wrapper is still enabled
	enabled := k.IsEnabled(ctx)
	require.True(t, enabled)
}

func TestMsgDisableTokenWrapper_InvalidSignerAddress(t *testing.T) {
	// Test case: try to disable token wrapper if signer address is invalid

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// set operator address
	k.SetOperatorAddress(ctx, sample.AccAddress())
	// set the token wrapper to enabled state
	k.SetEnabled(ctx, true)

	// create a message to disable the token wrapper
	msg := &types.MsgDisableTokenWrapper{
		Signer: "zig15yk64u7zc9g9k2yr2wmzeva5qgwxps6y8c2amk",
	}

	// call the DisableTokenWrapper method
	resp, err := ms.DisableTokenWrapper(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check error msg
	require.Equal(t, err.Error(), "signer is neither a pauser nor the operator: unauthorized")

	// check that the token wrapper is still enabled
	enabled := k.IsEnabled(ctx)
	require.True(t, enabled)
}

func TestMsgDisableTokenWrapper_OperatorNotSet(t *testing.T) {
	// Test case: operator address not set

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate a random address for the signer
	signer := sample.AccAddress()
	// intentionally NOT calling k.SetOperatorAddress(ctx, ...)
	k.SetEnabled(ctx, true)

	// create a message to disable the token wrapper
	msg := &types.MsgDisableTokenWrapper{
		Signer: signer,
	}

	// call the DisableTokenWrapper method
	resp, err := ms.DisableTokenWrapper(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check error msg
	require.Equal(t, err.Error(), "signer is neither a pauser nor the operator: unauthorized")

	// check that the token wrapper is still enabled
	enabled := k.IsEnabled(ctx)
	require.True(t, enabled)
}
