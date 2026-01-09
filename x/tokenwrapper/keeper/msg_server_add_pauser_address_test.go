package keeper_test

import (
	"testing"

	_ "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"zigchain/app"
	"zigchain/testutil/sample"
	"zigchain/x/tokenwrapper/keeper"
	"zigchain/x/tokenwrapper/types"
)

// Positive test cases

func TestMsgAddPauserAddress_ValidAddWithOperator(t *testing.T) {
	// Test case: add pauser address successfully with operator

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate a random address for the operator signer
	operatorSigner := sample.AccAddress()
	// generate a random address for the new pauser
	newPauser := sample.AccAddress()
	// simulate on-chain state by setting the operator address
	// to match the test signer
	k.SetOperatorAddress(ctx, operatorSigner)

	// create a message to add pauser address
	// with the correct signer - Operator address
	msg := &types.MsgAddPauserAddress{
		Signer:    operatorSigner,
		NewPauser: newPauser,
	}

	// call the AddPauserAddress method
	resp, err := ms.AddPauserAddress(ctx, msg)
	// verify that the handler executed without any errors
	require.NoError(t, err)
	// confirm that a valid response was returned
	require.NotNil(t, resp)
	require.Equal(t, resp, &types.MsgAddPauserAddressResponse{
		Signer:          operatorSigner,
		PauserAddresses: k.GetPauserAddresses(ctx),
	})

	// check that the new pauser address was added
	isPauser := k.IsPauserAddress(ctx, newPauser)
	require.True(t, isPauser)

	// check that the event was emitted
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// find the pauser address added event
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventTypePauserAddressAdded {
			foundEvent = true
			// check event attributes
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyPauserAddress {
					require.Equal(t, newPauser, string(attr.Value))
				}
			}
			break
		}
	}
	require.True(t, foundEvent, "Expected PauserAddressAdded event to be emitted")
}

func TestMsgAddPauserAddress_AddMultiplePausers(t *testing.T) {
	// Test case: add multiple pauser addresses successfully

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate a random address for the operator signer
	operatorSigner := sample.AccAddress()
	// generate random addresses for multiple pausers
	pauser1 := sample.AccAddress()
	pauser2 := sample.AccAddress()
	// simulate on-chain state by setting the operator address
	k.SetOperatorAddress(ctx, operatorSigner)

	// create the first message to add pauser address
	msg1 := &types.MsgAddPauserAddress{
		Signer:    operatorSigner,
		NewPauser: pauser1,
	}

	// call the AddPauserAddress method for first pauser
	resp1, err1 := ms.AddPauserAddress(ctx, msg1)
	require.NoError(t, err1)
	require.NotNil(t, resp1)

	// create the second message to add pauser address
	msg2 := &types.MsgAddPauserAddress{
		Signer:    operatorSigner,
		NewPauser: pauser2,
	}

	// call the AddPauserAddress method for second pauser
	resp2, err2 := ms.AddPauserAddress(ctx, msg2)
	require.NoError(t, err2)
	require.NotNil(t, resp2)

	// check that both pauser addresses were added
	isPauser1 := k.IsPauserAddress(ctx, pauser1)
	require.True(t, isPauser1)
	isPauser2 := k.IsPauserAddress(ctx, pauser2)
	require.True(t, isPauser2)

	// check that events were emitted for both pausers
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// count pauser address added events
	pauserAddedEvents := 0
	pauserAddresses := make(map[string]bool)
	for _, event := range events {
		if event.Type == types.EventTypePauserAddressAdded {
			pauserAddedEvents++
			// collect pauser addresses from events
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyPauserAddress {
					pauserAddresses[string(attr.Value)] = true
				}
			}
		}
	}
	require.Equal(t, 2, pauserAddedEvents, "Expected 2 PauserAddressAdded events to be emitted")
	require.True(t, pauserAddresses[pauser1], "Expected pauser1 address in events")
	require.True(t, pauserAddresses[pauser2], "Expected pauser2 address in events")
}

func TestMsgAddPauserAddress_AddExistingPauser(t *testing.T) {
	// Test case: add pauser address that already exists - no effect

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate a random address for the operator signer
	operatorSigner := sample.AccAddress()
	// generate a random address for the pauser
	pauser := sample.AccAddress()
	// simulate on-chain state by setting the operator address
	k.SetOperatorAddress(ctx, operatorSigner)
	// pre-add the pauser address
	k.AddPauserAddress(ctx, pauser)

	// create a message to add the same pauser address again
	msg := &types.MsgAddPauserAddress{
		Signer:    operatorSigner,
		NewPauser: pauser,
	}

	// call the AddPauserAddress method
	resp, err := ms.AddPauserAddress(ctx, msg)
	// verify that the handler executed without any errors
	require.NoError(t, err)
	// confirm that a valid response was returned
	require.NotNil(t, resp)

	// check that the pauser address is still a pauser
	isPauser := k.IsPauserAddress(ctx, pauser)
	require.True(t, isPauser)

	// check that the event was emitted (even for existing pauser)
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// find the pauser address added event
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventTypePauserAddressAdded {
			foundEvent = true
			// check event attributes
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyPauserAddress {
					require.Equal(t, pauser, string(attr.Value))
				}
			}
			break
		}
	}
	require.True(t, foundEvent, "Expected PauserAddressAdded event to be emitted")
}

// Negative test cases

func TestMsgAddPauserAddress_SignerNotOperator(t *testing.T) {
	// Test case: try to add pauser address if signer is not the operator

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate random addresses
	operatorAddress := sample.AccAddress()
	unauthorizedSigner := sample.AccAddress()
	newPauser := sample.AccAddress()
	// set the operator address to a different address than the signer
	k.SetOperatorAddress(ctx, operatorAddress)

	// create a message to add pauser address with unauthorized signer
	msg := &types.MsgAddPauserAddress{
		Signer:    unauthorizedSigner,
		NewPauser: newPauser,
	}

	// call the AddPauserAddress method
	resp, err := ms.AddPauserAddress(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check the error message
	require.Contains(t, err.Error(), "only the operator can add pauser addresses")

	// check that the pauser address was not added
	isPauser := k.IsPauserAddress(ctx, newPauser)
	require.False(t, isPauser)

	// check that no pauser address added event was emitted
	events := ctx.EventManager().Events()
	for _, event := range events {
		require.NotEqual(t, types.EventTypePauserAddressAdded, event.Type, "PauserAddressAdded event should not be emitted on failure")
	}
}

func TestMsgAddPauserAddress_EmptyOperator(t *testing.T) {
	// Test case: try to add pauser address when the operator is not set

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate random addresses
	signer := sample.AccAddress()
	newPauser := sample.AccAddress()
	// intentionally NOT setting operator address

	// create a message to add pauser address
	msg := &types.MsgAddPauserAddress{
		Signer:    signer,
		NewPauser: newPauser,
	}

	// call the AddPauserAddress method
	resp, err := ms.AddPauserAddress(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check the error message
	require.Contains(t, err.Error(), "only the operator can add pauser addresses")

	// check that the pauser address was not added
	isPauser := k.IsPauserAddress(ctx, newPauser)
	require.False(t, isPauser)
}

func TestMsgAddPauserAddress_InvalidPauserAddress(t *testing.T) {
	// Test case: try to add invalid pauser address

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate a random address for the operator signer
	operatorSigner := sample.AccAddress()
	// simulate on-chain state by setting the operator address
	k.SetOperatorAddress(ctx, operatorSigner)

	// create a message with invalid pauser address
	msg := &types.MsgAddPauserAddress{
		Signer:    operatorSigner,
		NewPauser: "invalidaddress",
	}

	// call the AddPauserAddress method
	resp, err := ms.AddPauserAddress(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check error message contains address validation error
	require.Contains(t, err.Error(), "invalid separator index")

	// check that no pauser address added event was emitted
	events := ctx.EventManager().Events()
	for _, event := range events {
		require.NotEqual(t, types.EventTypePauserAddressAdded, event.Type, "PauserAddressAdded event should not be emitted on failure")
	}
}

func TestMsgAddPauserAddress_EmptyPauserAddress(t *testing.T) {
	// Test case: try to add empty pauser address

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate a random address for the operator signer
	operatorSigner := sample.AccAddress()
	// simulate on-chain state by setting the operator address
	k.SetOperatorAddress(ctx, operatorSigner)

	// create a message with empty pauser address
	msg := &types.MsgAddPauserAddress{
		Signer:    operatorSigner,
		NewPauser: "",
	}

	// call the AddPauserAddress method
	resp, err := ms.AddPauserAddress(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check that error is related to address validation
	require.Error(t, err)

	// check that no pauser address added event was emitted
	events := ctx.EventManager().Events()
	for _, event := range events {
		require.NotEqual(t, types.EventTypePauserAddressAdded, event.Type, "PauserAddressAdded event should not be emitted on failure")
	}
}

func TestMsgAddPauserAddress_EmptySignerAddress(t *testing.T) {
	// Test case: try to add pauser address with empty signer

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate a random address for the operator
	operatorAddress := sample.AccAddress()
	newPauser := sample.AccAddress()
	// set the operator address
	k.SetOperatorAddress(ctx, operatorAddress)

	// create a message with empty signer address
	msg := &types.MsgAddPauserAddress{
		Signer:    "",
		NewPauser: newPauser,
	}

	// call the AddPauserAddress method
	resp, err := ms.AddPauserAddress(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check error message
	require.Contains(t, err.Error(), "only the operator can add pauser addresses")

	// check that the pauser address was not added
	isPauser := k.IsPauserAddress(ctx, newPauser)
	require.False(t, isPauser)
}

func TestMsgAddPauserAddress_InvalidSignerAddress(t *testing.T) {
	// Test case: try to add pauser address with invalid signer address format

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate a random address for the operator
	operatorAddress := sample.AccAddress()
	newPauser := sample.AccAddress()
	// set the operator address
	k.SetOperatorAddress(ctx, operatorAddress)

	// create a message with invalid signer address
	msg := &types.MsgAddPauserAddress{
		Signer:    "zig15yk64u7zc9g9k2yr2wmzeva5qgwxps6y8c2amk",
		NewPauser: newPauser,
	}

	// call the AddPauserAddress method
	resp, err := ms.AddPauserAddress(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check error message
	require.Contains(t, err.Error(), "only the operator can add pauser addresses")

	// check that the pauser address was not added
	isPauser := k.IsPauserAddress(ctx, newPauser)
	require.False(t, isPauser)
}
