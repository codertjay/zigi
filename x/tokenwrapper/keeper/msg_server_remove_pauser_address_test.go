package keeper_test

import (
	"testing"

	"zigchain/app"
	"zigchain/testutil/sample"
	"zigchain/x/tokenwrapper/keeper"
	"zigchain/x/tokenwrapper/types"

	"github.com/stretchr/testify/require"
)

// Positive test cases

func TestMsgRemovePauserAddress_ValidRemoveWithOperator(t *testing.T) {
	// Test case: remove pauser address successfully with operator

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
	// generate a random address for the pauser to remove
	pauserToRemove := sample.AccAddress()
	// simulate on-chain state by setting the operator address
	// to match the test signer
	k.SetOperatorAddress(ctx, operatorSigner)
	// pre-add the pauser address that will be removed
	k.AddPauserAddress(ctx, pauserToRemove)

	// verify pauser was added initially
	isPauserBefore := k.IsPauserAddress(ctx, pauserToRemove)
	require.True(t, isPauserBefore)

	// create a message to remove pauser address
	// with the correct signer - Operator address
	msg := &types.MsgRemovePauserAddress{
		Signer: operatorSigner,
		Pauser: pauserToRemove,
	}

	// call the RemovePauserAddress method
	resp, err := ms.RemovePauserAddress(ctx, msg)
	// verify that the handler executed without any errors
	require.NoError(t, err)
	// confirm that a valid response was returned
	require.NotNil(t, resp)
	require.Equal(t, resp, &types.MsgRemovePauserAddressResponse{
		Signer:          operatorSigner,
		PauserAddresses: k.GetPauserAddresses(ctx),
	})

	// check that the pauser address was removed
	isPauserAfter := k.IsPauserAddress(ctx, pauserToRemove)
	require.False(t, isPauserAfter)

	// check that the event was emitted
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// find the pauser address removed event
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventTypePauserAddressRemoved {
			foundEvent = true
			// check event attributes
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyPauserAddress {
					require.Equal(t, pauserToRemove, string(attr.Value))
				}
			}
			break
		}
	}
	require.True(t, foundEvent, "Expected PauserAddressRemoved event to be emitted")
}

func TestMsgRemovePauserAddress_RemoveMultiplePausers(t *testing.T) {
	// Test case: remove multiple pauser addresses successfully

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
	// pre-add both pauser addresses
	k.AddPauserAddress(ctx, pauser1)
	k.AddPauserAddress(ctx, pauser2)

	// verify pausers were added initially
	require.True(t, k.IsPauserAddress(ctx, pauser1))
	require.True(t, k.IsPauserAddress(ctx, pauser2))

	// create the first message to remove pauser address
	msg1 := &types.MsgRemovePauserAddress{
		Signer: operatorSigner,
		Pauser: pauser1,
	}

	// call the RemovePauserAddress method for first pauser
	resp1, err1 := ms.RemovePauserAddress(ctx, msg1)
	require.NoError(t, err1)
	require.NotNil(t, resp1)

	// create the second message to remove pauser address
	msg2 := &types.MsgRemovePauserAddress{
		Signer: operatorSigner,
		Pauser: pauser2,
	}

	// call the RemovePauserAddress method for second pauser
	resp2, err2 := ms.RemovePauserAddress(ctx, msg2)
	require.NoError(t, err2)
	require.NotNil(t, resp2)

	// check that both pauser addresses were removed
	isPauser1 := k.IsPauserAddress(ctx, pauser1)
	require.False(t, isPauser1)
	isPauser2 := k.IsPauserAddress(ctx, pauser2)
	require.False(t, isPauser2)

	// check that events were emitted for both pausers
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// count pauser address removed events
	pauserRemovedEvents := 0
	pauserAddresses := make(map[string]bool)
	for _, event := range events {
		if event.Type == types.EventTypePauserAddressRemoved {
			pauserRemovedEvents++
			// collect pauser addresses from events
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyPauserAddress {
					pauserAddresses[string(attr.Value)] = true
				}
			}
		}
	}
	require.Equal(t, 2, pauserRemovedEvents, "Expected 2 PauserAddressRemoved events to be emitted")
	require.True(t, pauserAddresses[pauser1], "Expected pauser1 address in events")
	require.True(t, pauserAddresses[pauser2], "Expected pauser2 address in events")
}

func TestMsgRemovePauserAddress_RemoveNonExistentPauser(t *testing.T) {
	// Test case: remove pauser address that doesn't exist

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
	// generate a random address for the non-existent pauser
	nonExistentPauser := sample.AccAddress()
	// simulate on-chain state by setting the operator address
	k.SetOperatorAddress(ctx, operatorSigner)
	// intentionally NOT adding the pauser address

	// verify pauser doesn't exist initially
	isPauserBefore := k.IsPauserAddress(ctx, nonExistentPauser)
	require.False(t, isPauserBefore)

	// create a message to remove the non-existent pauser address
	msg := &types.MsgRemovePauserAddress{
		Signer: operatorSigner,
		Pauser: nonExistentPauser,
	}

	// call the RemovePauserAddress method
	resp, err := ms.RemovePauserAddress(ctx, msg)
	// verify that the handler executed without any errors
	require.NoError(t, err)
	// confirm that a valid response was returned
	require.NotNil(t, resp)

	// check that the pauser address is still not a pauser
	isPauserAfter := k.IsPauserAddress(ctx, nonExistentPauser)
	require.False(t, isPauserAfter)

	// check that the event was emitted (even for non-existent pauser)
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// find the pauser address removed event
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventTypePauserAddressRemoved {
			foundEvent = true
			// check event attributes
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyPauserAddress {
					require.Equal(t, nonExistentPauser, string(attr.Value))
				}
			}
			break
		}
	}
	require.True(t, foundEvent, "Expected PauserAddressRemoved event to be emitted")
}

// Negative test cases

func TestMsgRemovePauserAddress_SignerNotOperator(t *testing.T) {
	// Test case: try to remove pauser address if signer is not the operator

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
	pauserToRemove := sample.AccAddress()
	// set the operator address to a different address than the signer
	k.SetOperatorAddress(ctx, operatorAddress)
	// pre-add the pauser address
	k.AddPauserAddress(ctx, pauserToRemove)

	// create a message to remove pauser address with unauthorized signer
	msg := &types.MsgRemovePauserAddress{
		Signer: unauthorizedSigner,
		Pauser: pauserToRemove,
	}

	// call the RemovePauserAddress method
	resp, err := ms.RemovePauserAddress(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check error message
	require.Contains(t, err.Error(), "only the operator can remove pauser addresses")

	// check that the pauser address was not removed
	isPauser := k.IsPauserAddress(ctx, pauserToRemove)
	require.True(t, isPauser)

	// check that no pauser address removed event was emitted
	events := ctx.EventManager().Events()
	for _, event := range events {
		require.NotEqual(t, types.EventTypePauserAddressRemoved, event.Type, "PauserAddressRemoved event should not be emitted on failure")
	}
}

func TestMsgRemovePauserAddress_EmptyOperator(t *testing.T) {
	// Test case: try to remove pauser address when operator is not set

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
	pauserToRemove := sample.AccAddress()
	// intentionally NOT setting operator address
	// pre-add the pauser address
	k.AddPauserAddress(ctx, pauserToRemove)

	// create a message to remove pauser address
	msg := &types.MsgRemovePauserAddress{
		Signer: signer,
		Pauser: pauserToRemove,
	}

	// call the RemovePauserAddress method
	resp, err := ms.RemovePauserAddress(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check error message
	require.Contains(t, err.Error(), "only the operator can remove pauser addresses")

	// check that the pauser address was not removed
	isPauser := k.IsPauserAddress(ctx, pauserToRemove)
	require.True(t, isPauser)

	// check that no pauser address removed event was emitted
	events := ctx.EventManager().Events()
	for _, event := range events {
		require.NotEqual(t, types.EventTypePauserAddressRemoved, event.Type, "PauserAddressRemoved event should not be emitted on failure")
	}
}

func TestMsgRemovePauserAddress_InvalidPauserAddress(t *testing.T) {
	// Test case: try to remove invalid pauser address

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
	msg := &types.MsgRemovePauserAddress{
		Signer: operatorSigner,
		Pauser: "invalid_address",
	}

	// call the RemovePauserAddress method
	resp, err := ms.RemovePauserAddress(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check error message contains address validation error
	require.Contains(t, err.Error(), "invalid separator index -1")

	// check that no pauser address removed event was emitted
	events := ctx.EventManager().Events()
	for _, event := range events {
		require.NotEqual(t, types.EventTypePauserAddressRemoved, event.Type, "PauserAddressRemoved event should not be emitted on failure")
	}
}

func TestMsgRemovePauserAddress_EmptyPauserAddress(t *testing.T) {
	// Test case: try to remove empty pauser address

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
	msg := &types.MsgRemovePauserAddress{
		Signer: operatorSigner,
		Pauser: "",
	}

	// call the RemovePauserAddress method
	resp, err := ms.RemovePauserAddress(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check that error is related to address validation
	require.Error(t, err)

	// check that no pauser address removed event was emitted
	events := ctx.EventManager().Events()
	for _, event := range events {
		require.NotEqual(t, types.EventTypePauserAddressRemoved, event.Type, "PauserAddressRemoved event should not be emitted on failure")
	}
}

func TestMsgRemovePauserAddress_EmptySignerAddress(t *testing.T) {
	// Test case: try to remove pauser address with empty signer

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
	pauserToRemove := sample.AccAddress()
	// set the operator address
	k.SetOperatorAddress(ctx, operatorAddress)
	// pre-add the pauser address
	k.AddPauserAddress(ctx, pauserToRemove)

	// create a message with empty signer address
	msg := &types.MsgRemovePauserAddress{
		Signer: "",
		Pauser: pauserToRemove,
	}

	// call the RemovePauserAddress method
	resp, err := ms.RemovePauserAddress(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check error message
	require.Contains(t, err.Error(), "only the operator can remove pauser addresses")

	// check that the pauser address was not removed
	isPauser := k.IsPauserAddress(ctx, pauserToRemove)
	require.True(t, isPauser)

	// check that no pauser address removed event was emitted
	events := ctx.EventManager().Events()
	for _, event := range events {
		require.NotEqual(t, types.EventTypePauserAddressRemoved, event.Type, "PauserAddressRemoved event should not be emitted on failure")
	}
}

func TestMsgRemovePauserAddress_InvalidSignerAddress(t *testing.T) {
	// Test case: try to remove pauser address with invalid signer address format

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
	pauserToRemove := sample.AccAddress()
	// set the operator address
	k.SetOperatorAddress(ctx, operatorAddress)
	// pre-add the pauser address
	k.AddPauserAddress(ctx, pauserToRemove)

	// create a message with invalid signer address
	msg := &types.MsgRemovePauserAddress{
		Signer: "zig15yk64u7zc9g9k2yr2wmzeva5qgwxps6y8c2amk",
		Pauser: pauserToRemove,
	}

	// call the RemovePauserAddress method
	resp, err := ms.RemovePauserAddress(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check error message
	require.Contains(t, err.Error(), "only the operator can remove pauser addresses")

	// check that the pauser address was not removed
	isPauser := k.IsPauserAddress(ctx, pauserToRemove)
	require.True(t, isPauser)

	// check that no pauser address removed event was emitted
	events := ctx.EventManager().Events()
	for _, event := range events {
		require.NotEqual(t, types.EventTypePauserAddressRemoved, event.Type, "PauserAddressRemoved event should not be emitted on failure")
	}
}
