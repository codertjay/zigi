package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"zigchain/app"
	keepertest "zigchain/testutil/keeper"
	"zigchain/testutil/sample"
	"zigchain/x/tokenwrapper/keeper"
	"zigchain/x/tokenwrapper/types"
)

// Positive test cases

func TestMsgProposeOperatorAddress_Valid(t *testing.T) {
	// Test case: valid operator proposal

	// initialize a test blockchain app
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

	// generate a new random address for the new operator
	newOperator := sample.AccAddress()

	// create a message to propose an operator address
	msg := &types.MsgProposeOperatorAddress{
		Signer:      signer,
		NewOperator: newOperator,
	}

	// call the ProposeOperatorAddress method
	resp, err := ms.ProposeOperatorAddress(ctx, msg)
	// verify that the handler executed without any errors
	require.NoError(t, err)
	// confirm that a valid response was returned
	require.NotNil(t, resp)
	require.Equal(t, resp, &types.MsgProposeOperatorAddressResponse{
		Signer:                  signer,
		ProposedOperatorAddress: newOperator,
	})

	// verify the proposed operator address was set
	require.Equal(t, newOperator, k.GetProposedOperatorAddress(ctx))

	// CHECK EVENTS
	// ---------------------------------

	// check that the event was emitted
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// find the pauser address added event
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventTypeOperatorAddressProposed {
			foundEvent = true
			// check event attributes
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyOldOperator {
					require.Equal(t, signer, string(attr.Value))
				}
			}
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyNewOperator {
					require.Equal(t, newOperator, string(attr.Value))
				}
			}
			break
		}
	}
	require.True(t, foundEvent, "Expected EventTypeOperatorAddressProposed event to be emitted")
}

// Negative test cases

func TestMsgProposeOperatorAddress_SignerNotCurrentOperator(t *testing.T) {
	// Test case: try to propose operator address if Signer != currentOperator

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

	// generate a new random address for the new operator
	newOperator := sample.AccAddress()

	// create a message to propose an operator address
	msg := &types.MsgProposeOperatorAddress{
		Signer:      sample.AccAddress(),
		NewOperator: newOperator,
	}

	// call the ProposeOperatorAddress method
	resp, err := ms.ProposeOperatorAddress(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check error msg
	require.Equal(t, err.Error(), "only the current operator can propose a new operator address")

	// check that the proposed operator address was not set
	currentOperator := k.GetOperatorAddress(ctx)
	require.Equal(t, currentOperator, signer, "Proposed operator address should not change")

	// Check that no event was emitted
	events := ctx.EventManager().Events()
	require.Equal(t, 0, len(events), "No events should be emitted when proposal fails")
}

func TestMsgProposeOperatorAddress_SignerNotCurrentOperator2(t *testing.T) {
	// Test case: try to propose operator address if Signer != currentOperator

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

	// set the operator address to random address
	randomOperator := sample.AccAddress()
	k.SetOperatorAddress(ctx, randomOperator)

	// generate a new random address for the new operator
	newOperator := sample.AccAddress()

	// create a message to propose an operator address
	msg := &types.MsgProposeOperatorAddress{
		Signer:      signer,
		NewOperator: newOperator,
	}

	// call the ProposeOperatorAddress method
	resp, err := ms.ProposeOperatorAddress(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check error msg
	require.Equal(t, err.Error(), "only the current operator can propose a new operator address")

	// check that the proposed operator address was not set
	currentOperator := k.GetOperatorAddress(ctx)
	require.Equal(t, currentOperator, randomOperator, "Proposed operator address should not change")

	// Check that no event was emitted
	events := ctx.EventManager().Events()
	require.Equal(t, 0, len(events), "No events should be emitted when proposal fails")
}

func TestMsgProposeOperatorAddress_EmptySigner(t *testing.T) {
	// Test case: try to propose operator address if the signer address is empty

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
	// set the operator address to random address
	k.SetOperatorAddress(ctx, signer)

	// generate a new random address for the new operator
	newOperator := sample.AccAddress()

	// create a message to propose an operator address
	msg := &types.MsgProposeOperatorAddress{
		Signer:      "",
		NewOperator: newOperator,
	}

	// call the ProposeOperatorAddress method
	resp, err := ms.ProposeOperatorAddress(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check error msg
	require.Equal(t, err.Error(), "only the current operator can propose a new operator address")

	// check that the proposed operator address was not set
	currentOperator := k.GetOperatorAddress(ctx)
	require.Equal(t, currentOperator, signer, "Proposed operator address should not change")

	// Check that no event was emitted
	events := ctx.EventManager().Events()
	require.Equal(t, 0, len(events), "No events should be emitted when proposal fails")
}

func TestMsgProposeOperatorAddress_OperatorNotSet(t *testing.T) {
	// Test case: try to propose the operator address if the operator address is not set

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
	// intentionally NOT calling k.SetOperatorAddress(ctx, ...)

	// generate a new random address for the new operator
	newOperator := sample.AccAddress()

	// create a message to propose an operator address
	msg := &types.MsgProposeOperatorAddress{
		Signer:      signer,
		NewOperator: newOperator,
	}

	// call the ProposeOperatorAddress method
	resp, err := ms.ProposeOperatorAddress(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check error msg
	require.Equal(t, err.Error(), "only the current operator can propose a new operator address")

	// check that the proposed operator address was not set
	currentOperator := k.GetOperatorAddress(ctx)
	t.Log("Current operator address:", currentOperator)
	require.Equal(t, currentOperator, sample.ZeroAccAddress(), "Proposed operator address should not change")

	// Check that no event was emitted
	events := ctx.EventManager().Events()
	require.Equal(t, 0, len(events), "No events should be emitted when proposal fails")
}

func TestMsgProposeOperatorAddress_EmptyOperatorAddress(t *testing.T) {
	// Test case: try to propose a new operator address if the operator address is empty

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
	// set the operator address to random address
	k.SetOperatorAddress(ctx, signer)

	// create a message to propose an operator address
	msg := &types.MsgProposeOperatorAddress{
		Signer:      signer,
		NewOperator: "",
	}

	// call the ProposeOperatorAddress method
	resp, err := ms.ProposeOperatorAddress(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check error msg
	require.Equal(t, err.Error(), "invalid new operator address: ")

	// check that the proposed operator address was not set
	currentOperator := k.GetOperatorAddress(ctx)
	require.Equal(t, currentOperator, signer, "Proposed operator address should not change")

	// Check that no event was emitted
	events := ctx.EventManager().Events()
	require.Equal(t, 0, len(events), "No events should be emitted when proposal fails")
}

func TestMsgProposeOperatorAddress_InvalidOperatorAddress(t *testing.T) {
	// Test case: try to propose a new operator address if the operator address is invalid

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
	// set the operator address to random address
	k.SetOperatorAddress(ctx, signer)

	// create a message to propose an operator address
	msg := &types.MsgProposeOperatorAddress{
		Signer:      signer,
		NewOperator: "bad_address",
	}

	// call the ProposeOperatorAddress method
	resp, err := ms.ProposeOperatorAddress(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check error msg
	require.Equal(t, err.Error(), "invalid new operator address: bad_address")

	// check that the proposed operator address was not set
	currentOperator := k.GetOperatorAddress(ctx)
	require.Equal(t, currentOperator, signer, "Proposed operator address should not change")

	// Check that no event was emitted
	events := ctx.EventManager().Events()
	require.Equal(t, 0, len(events), "No events should be emitted when proposal fails")
}

func TestMsgProposeOperatorAddress_SameOperatorAddress(t *testing.T) {
	// Test case: try to propose the same operator address

	// initialize test blockchain app
	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ms := keeper.NewMsgServerImpl(k)
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate a random address for the signer/operator
	operator := sample.AccAddress()
	k.SetOperatorAddress(ctx, operator)

	// create a message to propose the same operator address
	msg := &types.MsgProposeOperatorAddress{
		Signer:      operator,
		NewOperator: operator,
	}

	// call the ProposeOperatorAddress method
	resp, err := ms.ProposeOperatorAddress(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check error msg
	require.Equal(t, err.Error(), "cannot propose the same operator address")

	// check that the proposed operator address was not set
	currentOperator := k.GetOperatorAddress(ctx)
	require.Equal(t, currentOperator, operator, "Proposed operator address should not change")

	// Check that no event was emitted
	events := ctx.EventManager().Events()
	require.Equal(t, 0, len(events), "No events should be emitted when proposal fails")
}

func TestProposeOperatorAddress_Unauthorized(t *testing.T) {
	// Test case: an unauthorized address tries to propose a new operator

	k, ctx := keepertest.TokenwrapperKeeper(t, nil)
	msgServer := keeper.NewMsgServerImpl(k)

	currentOperator := sdk.AccAddress([]byte("current_operator"))
	unauthorized := sdk.AccAddress([]byte("unauthorized"))
	k.SetOperatorAddress(ctx, currentOperator.String())

	// Test proposing with unauthorized address
	msg := &types.MsgProposeOperatorAddress{
		Signer:      unauthorized.String(),
		NewOperator: sdk.AccAddress([]byte("new_operator")).String(),
	}
	_, err := msgServer.ProposeOperatorAddress(ctx, msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "only the current operator can propose a new operator address")
}
