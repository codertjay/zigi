package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"zigchain/app"
	"zigchain/testutil/sample"
	"zigchain/x/tokenwrapper/keeper"
	"zigchain/x/tokenwrapper/types"
)

// Positive tests cases

func TestMsgUpdateIbcSettings_Valid(t *testing.T) {
	// Test case: update IBC settings with valid parameters

	// initialize the test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate and set valid operator
	signer := sample.AccAddress()
	k.SetOperatorAddress(ctx, signer)

	// build a valid message
	msg := &types.MsgUpdateIbcSettings{
		Signer:               signer,
		NativeClientId:       "client-123",
		CounterpartyClientId: "client-456",
		NativePort:           "transfer",
		CounterpartyPort:     "transfer_01",
		NativeChannel:        "channel-0",
		CounterpartyChannel:  "channel-1",
		Denom:                "uzig",
		DecimalDifference:    18,
	}

	// call the UpdateIbcSettings method
	resp, err := ms.UpdateIbcSettings(ctx, msg)
	// check if the response is not nil and no error occurred
	require.NoError(t, err)
	require.NotNil(t, resp)

	// verify state
	require.Equal(t, msg.NativeClientId, k.GetNativeClientId(ctx))
	require.Equal(t, msg.CounterpartyClientId, k.GetCounterpartyClientId(ctx))
	require.Equal(t, msg.NativePort, k.GetNativePort(ctx))
	require.Equal(t, msg.CounterpartyPort, k.GetCounterpartyPort(ctx))
	require.Equal(t, msg.NativeChannel, k.GetNativeChannel(ctx))
	require.Equal(t, msg.CounterpartyChannel, k.GetCounterpartyChannel(ctx))
	require.Equal(t, msg.Denom, k.GetDenom(ctx))
	require.Equal(t, msg.DecimalDifference, k.GetDecimalDifference(ctx))

	// CHECK EVENTS
	// ---------------------------------

	// check that the event was emitted
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// find the pauser address added event
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventTypeIbcSettingsUpdated {
			foundEvent = true
			// check event attributes
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeySigner {
					require.Equal(t, signer, string(attr.Value))
				}
			}
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyNativeClientId {
					require.Equal(t, "client-123", string(attr.Value))
				}
			}
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyCounterpartyClientId {
					require.Equal(t, "client-456", string(attr.Value))
				}
			}
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyNativePort {
					require.Equal(t, "transfer", string(attr.Value))
				}
			}
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyCounterpartyPort {
					require.Equal(t, "transfer_01", string(attr.Value))
				}
			}
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyNativeChannel {
					require.Equal(t, "channel-0", string(attr.Value))
				}
			}
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyCounterpartyChannel {
					require.Equal(t, "channel-1", string(attr.Value))
				}
			}
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyDenom {
					require.Equal(t, "uzig", string(attr.Value))
				}
			}
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyDecimalDifference {
					require.Equal(t, "18", string(attr.Value))
				}
			}
			break
		}
	}
	require.True(t, foundEvent, "Expected EventTypeIbcSettingsUpdated event to be emitted")

	// build a valid update message
	updateMsg := &types.MsgUpdateIbcSettings{
		Signer:               signer,
		NativeClientId:       "client-456",
		CounterpartyClientId: "client-789",
		NativePort:           "transfer_01",
		CounterpartyPort:     "transfer_02",
		NativeChannel:        "channel-1",
		CounterpartyChannel:  "channel-2",
		Denom:                "uzig",
		DecimalDifference:    8,
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

	// check the response
	require.Equal(t, updateResp, &types.MsgUpdateIbcSettingsResponse{
		Signer:               signer,
		NativeClientId:       updateMsg.NativeClientId,
		CounterpartyClientId: updateMsg.CounterpartyClientId,
		NativePort:           updateMsg.NativePort,
		CounterpartyPort:     updateMsg.CounterpartyPort,
		NativeChannel:        updateMsg.NativeChannel,
		CounterpartyChannel:  updateMsg.CounterpartyChannel,
		Denom:                updateMsg.Denom,
		DecimalDifference:    updateMsg.DecimalDifference,
	})
}

func TestMsgUpdateIbcSettings_ValidWithHyphen(t *testing.T) {
	// Test case: update IBC settings with valid parameters

	// initialize the test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate and set valid operator
	signer := sample.AccAddress()
	k.SetOperatorAddress(ctx, signer)

	// build a valid message
	msg := &types.MsgUpdateIbcSettings{
		Signer:               signer,
		NativeClientId:       "client-123",
		CounterpartyClientId: "client-456",
		NativePort:           "transfer",
		CounterpartyPort:     "transfer_01",
		NativeChannel:        "channel-0",
		CounterpartyChannel:  "channel-1",
		Denom:                "unit-zig",
		DecimalDifference:    18,
	}

	// call the UpdateIbcSettings method
	resp, err := ms.UpdateIbcSettings(ctx, msg)
	// check if the response is not nil and no error occurred
	require.NoError(t, err)
	require.NotNil(t, resp)

	// verify state
	require.Equal(t, msg.NativeClientId, k.GetNativeClientId(ctx))
	require.Equal(t, msg.CounterpartyClientId, k.GetCounterpartyClientId(ctx))
	require.Equal(t, msg.NativePort, k.GetNativePort(ctx))
	require.Equal(t, msg.CounterpartyPort, k.GetCounterpartyPort(ctx))
	require.Equal(t, msg.NativeChannel, k.GetNativeChannel(ctx))
	require.Equal(t, msg.CounterpartyChannel, k.GetCounterpartyChannel(ctx))
	require.Equal(t, msg.Denom, k.GetDenom(ctx))
	require.Equal(t, msg.DecimalDifference, k.GetDecimalDifference(ctx))

	// CHECK EVENTS
	// ---------------------------------

	// check that the event was emitted
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// find the pauser address added event
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventTypeIbcSettingsUpdated {
			foundEvent = true
			// check event attributes
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeySigner {
					require.Equal(t, signer, string(attr.Value))
				}
			}
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyNativeClientId {
					require.Equal(t, "client-123", string(attr.Value))
				}
			}
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyCounterpartyClientId {
					require.Equal(t, "client-456", string(attr.Value))
				}
			}
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyNativePort {
					require.Equal(t, "transfer", string(attr.Value))
				}
			}
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyCounterpartyPort {
					require.Equal(t, "transfer_01", string(attr.Value))
				}
			}
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyNativeChannel {
					require.Equal(t, "channel-0", string(attr.Value))
				}
			}
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyCounterpartyChannel {
					require.Equal(t, "channel-1", string(attr.Value))
				}
			}
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyDenom {
					require.Equal(t, "unit-zig", string(attr.Value))
				}
			}
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyDecimalDifference {
					require.Equal(t, "18", string(attr.Value))
				}
			}
			break
		}
	}
	require.True(t, foundEvent, "Expected EventTypeIbcSettingsUpdated event to be emitted")

	// build a valid update message
	updateMsg := &types.MsgUpdateIbcSettings{
		Signer:               signer,
		NativeClientId:       "client-456",
		CounterpartyClientId: "client-789",
		NativePort:           "transfer_01",
		CounterpartyPort:     "transfer_02",
		NativeChannel:        "channel-1",
		CounterpartyChannel:  "channel-2",
		Denom:                "unit-zig",
		DecimalDifference:    8,
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

	// check the response
	require.Equal(t, updateResp, &types.MsgUpdateIbcSettingsResponse{
		Signer:               signer,
		NativeClientId:       updateMsg.NativeClientId,
		CounterpartyClientId: updateMsg.CounterpartyClientId,
		NativePort:           updateMsg.NativePort,
		CounterpartyPort:     updateMsg.CounterpartyPort,
		NativeChannel:        updateMsg.NativeChannel,
		CounterpartyChannel:  updateMsg.CounterpartyChannel,
		Denom:                updateMsg.Denom,
		DecimalDifference:    updateMsg.DecimalDifference,
	})
}

// Negative test cases

func TestMsgUpdateIbcSettings_SignerNotCurrentOperator(t *testing.T) {
	// Test case: try to update IBC settings if the signer is not the current operator
	// msg.Signer != currentOperator

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate random signer address
	signer := sample.AccAddress()
	// set the operator address to a different address
	k.SetOperatorAddress(ctx, sample.AccAddress())

	// build a message with the signer not being the operator
	msg := &types.MsgUpdateIbcSettings{
		Signer:               signer,
		NativeClientId:       "client-123",
		CounterpartyClientId: "client-456",
		NativePort:           "transfer",
		CounterpartyPort:     "transfer_01",
		NativeChannel:        "channel-0",
		CounterpartyChannel:  "channel-1",
		Denom:                "uzig",
		DecimalDifference:    18,
	}

	// call the UpdateIbcSettings method
	resp, err := ms.UpdateIbcSettings(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check the error msg
	require.Equal(t, err.Error(), "only the current operator can update the IBC settings")
}

func TestMsgUpdateIbcSettings_EmptySignerAddress(t *testing.T) {
	// Test case: try to update IBC settings if the signer address is empty

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate random signer address
	signer := sample.AccAddress()
	// set the operator address to match the signer
	k.SetOperatorAddress(ctx, signer)

	// build a message with the empty signer address
	msg := &types.MsgUpdateIbcSettings{
		Signer:               "",
		NativeClientId:       "client-123",
		CounterpartyClientId: "client-456",
		NativePort:           "transfer",
		CounterpartyPort:     "transfer_01",
		NativeChannel:        "channel-0",
		CounterpartyChannel:  "channel-1",
		Denom:                "uzig",
		DecimalDifference:    18,
	}

	// call the UpdateIbcSettings method
	resp, err := ms.UpdateIbcSettings(ctx, msg)
	// check if the error is returned
	require.Error(t, err)
	// check if the response is nil
	require.Nil(t, resp)

	// check the error msg
	require.Equal(t, err.Error(), "only the current operator can update the IBC settings")
}

func TestMsgUpdateIbcSettings_InvalidDecimalDifference(t *testing.T) {
	// Test case: try to update IBC settings with an invalid DecimalDifference

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ms := keeper.NewMsgServerImpl(k)
	ctx := testApp.BaseApp.NewContext(initChain)

	// Set valid operator
	signer := sample.AccAddress()
	k.SetOperatorAddress(ctx, signer)

	// Build message with invalid DecimalDifference
	msg := &types.MsgUpdateIbcSettings{
		Signer:               signer,
		NativeClientId:       "client-123",
		CounterpartyClientId: "client-456",
		NativePort:           "transfer",
		CounterpartyPort:     "transfer_01",
		NativeChannel:        "channel-0",
		CounterpartyChannel:  "channel-1",
		Denom:                "uzig",
		// exceeding the maximum allowed value
		DecimalDifference: 19,
	}

	// Call UpdateIbcSettings
	resp, err := ms.UpdateIbcSettings(ctx, msg)
	require.Error(t, err)
	require.Nil(t, resp)

	// Verify error
	require.Contains(t, err.Error(), "decimal difference must be between 0 and 18, got 19")
}

func TestMsgUpdateIbcSettings_InvalidSignerFormat(t *testing.T) {
	// Test case: try to update IBC settings with a malformed signer address

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ms := keeper.NewMsgServerImpl(k)
	ctx := testApp.BaseApp.NewContext(initChain)

	// Set valid operator
	signer := sample.AccAddress()
	k.SetOperatorAddress(ctx, signer)

	// Build message with invalid signer format (non-Bech32)
	msg := &types.MsgUpdateIbcSettings{
		Signer:               "invalid_address_format",
		NativeClientId:       "client-123",
		CounterpartyClientId: "client-456",
		NativePort:           "transfer",
		CounterpartyPort:     "transfer_01",
		NativeChannel:        "channel-0",
		CounterpartyChannel:  "channel-1",
		Denom:                "uzig",
		DecimalDifference:    18,
	}

	// Call UpdateIbcSettings
	resp, err := ms.UpdateIbcSettings(ctx, msg)
	require.Error(t, err)
	require.Nil(t, resp)

	// Verify error
	require.Equal(t, "only the current operator can update the IBC settings", err.Error())
}

func TestMsgUpdateIbcSettings_NoOperatorSet(t *testing.T) {
	// Test case: try to update IBC settings when no operator is set

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ms := keeper.NewMsgServerImpl(k)
	ctx := testApp.BaseApp.NewContext(initChain)

	// Do not set an operator (leave it empty)

	// Build message with a valid signer
	signer := sample.AccAddress()
	msg := &types.MsgUpdateIbcSettings{
		Signer:               signer,
		NativeClientId:       "client-123",
		CounterpartyClientId: "client-456",
		NativePort:           "transfer",
		CounterpartyPort:     "transfer_01",
		NativeChannel:        "channel-0",
		CounterpartyChannel:  "channel-1",
		Denom:                "uzig",
		DecimalDifference:    18,
	}

	// Call UpdateIbcSettings
	resp, err := ms.UpdateIbcSettings(ctx, msg)
	require.Error(t, err)
	require.Nil(t, resp)

	// Verify error
	require.Equal(t, "only the current operator can update the IBC settings", err.Error())
}
