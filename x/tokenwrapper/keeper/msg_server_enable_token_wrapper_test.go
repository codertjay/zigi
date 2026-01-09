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

func TestMsgEnableTokenWrapper_ValidEnable(t *testing.T) {
	// Test case: enable token wrapper successfully

	// initialize the test blockchain app
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

	// create a message to enable the token wrapper
	// with the correct signer
	msg := &types.MsgEnableTokenWrapper{
		Signer: signer,
	}

	// call the EnableTokenWrapper method
	resp, err := ms.EnableTokenWrapper(ctx, msg)
	// verify that the handler executed without any errors
	require.NoError(t, err)
	// confirm that a valid response was returned
	require.NotNil(t, resp)
	require.Equal(t, resp, &types.MsgEnableTokenWrapperResponse{
		Signer:  signer,
		Enabled: true,
	})

	// check that the token wrapper is now enabled
	enabled := k.IsEnabled(ctx)
	require.True(t, enabled)

	// CHECK EVENTS
	// ---------------------------------

	// check that the event was emitted
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// find the pauser address added event
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventTypeTokenWrapperEnabled {
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

func TestMsgEnableTokenWrapper_AlreadyEnabled(t *testing.T) {
	// Test case: enable token wrapper if already enabled

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

	// create a message to enable the token wrapper
	msg := &types.MsgEnableTokenWrapper{
		Signer: signer,
	}

	// call the EnableTokenWrapper method
	resp, err := ms.EnableTokenWrapper(ctx, msg)
	// verify that the handler executed without any errors
	require.NoError(t, err)
	// confirm that a valid response was returned
	require.NotNil(t, resp)

	// check that the token wrapper is still enabled
	enabled := k.IsEnabled(ctx)
	require.True(t, enabled)
}

// Negative test cases

func TestMsgEnableTokenWrapper_SignerNotCurrentOperator(t *testing.T) {
	// Test case: try to enable token wrapper if Signer != currentOperator

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

	// create a message to enable the token wrapper
	msg := &types.MsgEnableTokenWrapper{
		Signer: sample.AccAddress(),
	}

	// call the EnableTokenWrapper method
	resp, err := ms.EnableTokenWrapper(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check error msg
	require.Equal(t, err.Error(), "only the operator can enable the token wrapper: unauthorized")

	// check that the token wrapper is not enabled
	enabled := k.IsEnabled(ctx)
	require.False(t, enabled)
}

func TestMsgEnableTokenWrapper_SignerNotCurrentOperator2(t *testing.T) {
	// Test case: try to enable token wrapper if Signer != currentOperator

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
	// set operator address
	k.SetOperatorAddress(ctx, sample.AccAddress())
	// set the token wrapper to disabled state
	k.SetEnabled(ctx, false)

	// create a message to enable the token wrapper
	msg := &types.MsgEnableTokenWrapper{
		Signer: signer,
	}

	// call the EnableTokenWrapper method
	resp, err := ms.EnableTokenWrapper(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check error msg
	require.Equal(t, err.Error(), "only the operator can enable the token wrapper: unauthorized")

	// check that the token wrapper is not enabled
	enabled := k.IsEnabled(ctx)
	require.False(t, enabled)
}

func TestMsgEnableTokenWrapper_SignerIsPauser(t *testing.T) {
	// Test case: try to enable token wrapper if Signer != currentOperator and it is a pauser

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
	pauser := sample.AccAddress()
	// simulate on-chain state by setting the operator address
	// to match the test signer
	k.SetOperatorAddress(ctx, signer)
	// set the token wrapper to disabled state
	k.SetEnabled(ctx, false)

	// add the pauser address
	k.AddPauserAddress(ctx, pauser)

	// create a message to enable the token wrapper
	msg := &types.MsgEnableTokenWrapper{
		Signer: pauser,
	}

	// call the EnableTokenWrapper method
	resp, err := ms.EnableTokenWrapper(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check error msg
	require.Equal(t, err.Error(), "only the operator can enable the token wrapper: unauthorized")

	// check that the token wrapper is not enabled
	enabled := k.IsEnabled(ctx)
	require.False(t, enabled)
}

func TestMsgEnableTokenWrapper_InvalidOperator(t *testing.T) {
	// Test case: try to enable token wrapper if operator address is invalid

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
	// set operator address to match signer address
	k.SetOperatorAddress(ctx, signer)
	// set the token wrapper to disabled state
	k.SetEnabled(ctx, false)

	// create a message to enable the token wrapper
	msg := &types.MsgEnableTokenWrapper{
		Signer: "invalid_operator",
	}

	// call the EnableTokenWrapper method
	resp, err := ms.EnableTokenWrapper(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check error msg
	require.Equal(t, err.Error(), "only the operator can enable the token wrapper: unauthorized")

	// check that the token wrapper is not enabled
	enabled := k.IsEnabled(ctx)
	require.False(t, enabled)
}

func TestMsgEnableTokenWrapper_EmptySigner(t *testing.T) {
	// Test case: try to enable token wrapper if signer address is empty

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
	// set operator address to match signer address
	k.SetOperatorAddress(ctx, signer)
	// set the token wrapper to disabled state
	k.SetEnabled(ctx, false)

	// create a message to enable the token wrapper
	msg := &types.MsgEnableTokenWrapper{
		Signer: "",
	}

	// call the EnableTokenWrapper method
	resp, err := ms.EnableTokenWrapper(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check error msg
	require.Equal(t, err.Error(), "only the operator can enable the token wrapper: unauthorized")

	// check that the token wrapper is not enabled
	enabled := k.IsEnabled(ctx)
	require.False(t, enabled)
}

func TestMsgEnableTokenWrapper_InvalidSignerAddress(t *testing.T) {
	// Test case: try to enable token wrapper if signer address is invalid

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
	// set the token wrapper to disabled state
	k.SetEnabled(ctx, false)

	// create a message to enable the token wrapper
	msg := &types.MsgEnableTokenWrapper{
		Signer: "invalid_address",
	}

	// call the EnableTokenWrapper method
	resp, err := ms.EnableTokenWrapper(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check error msg
	require.Equal(t, err.Error(), "only the operator can enable the token wrapper: unauthorized")

	// check that the token wrapper is not enabled
	enabled := k.IsEnabled(ctx)
	require.False(t, enabled)
}

func TestMsgEnableTokenWrapper_OperatorNotSet(t *testing.T) {
	// Test case: try to enable token wrapper if operator address not set

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
	k.SetEnabled(ctx, false)

	// create a message to enable the token wrapper
	msg := &types.MsgEnableTokenWrapper{
		Signer: signer,
	}

	// call the EnableTokenWrapper method
	resp, err := ms.EnableTokenWrapper(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check error msg
	require.Equal(t, err.Error(), "only the operator can enable the token wrapper: unauthorized")

	// check that the token wrapper is not enabled
	enabled := k.IsEnabled(ctx)
	require.False(t, enabled)
}
