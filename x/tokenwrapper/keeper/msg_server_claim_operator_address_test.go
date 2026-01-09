package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "zigchain/testutil/keeper"
	"zigchain/x/tokenwrapper/keeper"
	"zigchain/x/tokenwrapper/types"
)

// Positive test cases

func TestProposeAndClaimOperatorAddress(t *testing.T) {
	// Test case: current operator proposes a new operator and the proposed operator claims it

	k, ctx := keepertest.TokenwrapperKeeper(t, nil)
	msgServer := keeper.NewMsgServerImpl(k)

	// Setup test addresses
	currentOperator := sdk.AccAddress([]byte("current_operator"))
	proposedOperator := sdk.AccAddress([]byte("proposed_operator"))
	unauthorized := sdk.AccAddress([]byte("unauthorized"))

	// Set initial operator
	k.SetOperatorAddress(ctx, currentOperator.String())

	// Test 1: Current operator proposes new operator
	proposeMsg := &types.MsgProposeOperatorAddress{
		Signer:      currentOperator.String(),
		NewOperator: proposedOperator.String(),
	}
	_, err := msgServer.ProposeOperatorAddress(ctx, proposeMsg)
	require.NoError(t, err)

	// Verify the proposed operator is set
	proposed := k.GetProposedOperatorAddress(ctx)
	require.Equal(t, proposedOperator.String(), proposed)

	// Test 2: Unauthorized address cannot propose
	unauthorizedProposeMsg := &types.MsgProposeOperatorAddress{
		Signer:      unauthorized.String(),
		NewOperator: proposedOperator.String(),
	}
	_, err = msgServer.ProposeOperatorAddress(ctx, unauthorizedProposeMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "only the current operator can propose a new operator address")

	// Test 3: Proposed operator claims the role
	claimMsg := &types.MsgClaimOperatorAddress{
		Signer: proposedOperator.String(),
	}
	resp, err := msgServer.ClaimOperatorAddress(ctx, claimMsg)
	require.NoError(t, err)

	require.Equal(t, resp, &types.MsgClaimOperatorAddressResponse{
		Signer:          proposedOperator.String(),
		OperatorAddress: proposedOperator.String(),
	})

	// Verify operator is updated and the proposed operator is cleared
	operator := k.GetOperatorAddress(ctx)
	require.Equal(t, proposedOperator.String(), operator)
	proposed = k.GetProposedOperatorAddress(ctx)
	require.Empty(t, proposed)

	// CHECK EVENTS
	// ----------------------------

	// check that the event was emitted
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// find the pauser address added event
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventTypeOperatorAddressClaimed {
			foundEvent = true
			// check event attributes
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyOldOperator {
					require.Equal(t, currentOperator.String(), string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyNewOperator {
					require.Equal(t, proposedOperator.String(), string(attr.Value))
				}
			}
			break
		}
	}
	require.True(t, foundEvent, "Expected EventTypeOperatorAddressClaimed event to be emitted")

	// Test 4: Cannot claim if no operator is proposed
	_, err = msgServer.ClaimOperatorAddress(ctx, claimMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no operator address has been proposed")

	// Test 5: Unauthorized address cannot claim
	unauthorizedClaimMsg := &types.MsgClaimOperatorAddress{
		Signer: unauthorized.String(),
	}
	_, err = msgServer.ClaimOperatorAddress(ctx, unauthorizedClaimMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no operator address has been proposed")

	// Test 6: Invalid address format
	invalidProposeMsg := &types.MsgProposeOperatorAddress{
		Signer:      currentOperator.String(),
		NewOperator: "invalid_address",
	}
	_, err = msgServer.ProposeOperatorAddress(ctx, invalidProposeMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "only the current operator can propose a new operator address")
}

func TestClaimOperatorAddress_MultipleSequentialProposals(t *testing.T) {
	// Test case: multiple sequential proposals and claims

	k, ctx := keepertest.TokenwrapperKeeper(t, nil)
	msgServer := keeper.NewMsgServerImpl(k)

	// Setup test addresses
	currentOperator := sdk.AccAddress([]byte("current_operator"))
	proposedOperator1 := sdk.AccAddress([]byte("proposed_operator1"))
	proposedOperator2 := sdk.AccAddress([]byte("proposed_operator2"))

	// Set initial operator
	k.SetOperatorAddress(ctx, currentOperator.String())

	// First proposal
	proposeMsg1 := &types.MsgProposeOperatorAddress{
		Signer:      currentOperator.String(),
		NewOperator: proposedOperator1.String(),
	}
	_, err := msgServer.ProposeOperatorAddress(ctx, proposeMsg1)
	require.NoError(t, err)

	// First claim
	claimMsg1 := &types.MsgClaimOperatorAddress{
		Signer: proposedOperator1.String(),
	}
	resp1, err := msgServer.ClaimOperatorAddress(ctx, claimMsg1)
	require.NoError(t, err)
	require.Equal(t, &types.MsgClaimOperatorAddressResponse{
		Signer:          proposedOperator1.String(),
		OperatorAddress: proposedOperator1.String(),
	}, resp1)

	// Verify first claim
	operator := k.GetOperatorAddress(ctx)
	require.Equal(t, proposedOperator1.String(), operator)
	proposed := k.GetProposedOperatorAddress(ctx)
	require.Empty(t, proposed)

	// Second proposal by new operator
	proposeMsg2 := &types.MsgProposeOperatorAddress{
		Signer:      proposedOperator1.String(),
		NewOperator: proposedOperator2.String(),
	}
	_, err = msgServer.ProposeOperatorAddress(ctx, proposeMsg2)
	require.NoError(t, err)

	// Second claim
	claimMsg2 := &types.MsgClaimOperatorAddress{
		Signer: proposedOperator2.String(),
	}
	resp2, err := msgServer.ClaimOperatorAddress(ctx, claimMsg2)
	require.NoError(t, err)
	require.Equal(t, &types.MsgClaimOperatorAddressResponse{
		Signer:          proposedOperator2.String(),
		OperatorAddress: proposedOperator2.String(),
	}, resp2)

	// Verify the second claim
	operator = k.GetOperatorAddress(ctx)
	require.Equal(t, proposedOperator2.String(), operator)
	proposed = k.GetProposedOperatorAddress(ctx)
	require.Empty(t, proposed)

	// Verify events for second claim
	events := ctx.EventManager().Events()
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventTypeOperatorAddressClaimed {
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyOldOperator && string(attr.Value) == proposedOperator1.String() {
					foundEvent = true
				}
			}
		}
	}
	require.True(t, foundEvent, "Expected EventTypeOperatorAddressClaimed event for second claim")
}

// Negative test cases

func TestClaimOperatorAddress_NoProposal(t *testing.T) {
	// Test case: trying to claim operator address when no proposal exists

	k, ctx := keepertest.TokenwrapperKeeper(t, nil)
	msgServer := keeper.NewMsgServerImpl(k)

	// Test claiming when no operator is proposed
	msg := &types.MsgClaimOperatorAddress{
		Signer: sdk.AccAddress([]byte("any_address")).String(),
	}
	_, err := msgServer.ClaimOperatorAddress(ctx, msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no operator address has been proposed")
}

func TestClaimOperatorAddress_UnauthorizedSigner(t *testing.T) {
	// Test case: trying to claim operator address with an unauthorized signer

	k, ctx := keepertest.TokenwrapperKeeper(t, nil)
	msgServer := keeper.NewMsgServerImpl(k)

	// Setup test addresses
	currentOperator := sdk.AccAddress([]byte("current_operator"))
	proposedOperator := sdk.AccAddress([]byte("proposed_operator"))
	unauthorizedSigner := sdk.AccAddress([]byte("unauthorized_signer"))

	// Set initial operator
	k.SetOperatorAddress(ctx, currentOperator.String())

	// Propose a new operator
	proposeMsg := &types.MsgProposeOperatorAddress{
		Signer:      currentOperator.String(),
		NewOperator: proposedOperator.String(),
	}
	_, err := msgServer.ProposeOperatorAddress(ctx, proposeMsg)
	require.NoError(t, err)

	// Verify the proposed operator is set
	proposed := k.GetProposedOperatorAddress(ctx)
	require.Equal(t, proposedOperator.String(), proposed)

	// Attempt to claim with an unauthorized signer
	claimMsg := &types.MsgClaimOperatorAddress{
		Signer: unauthorizedSigner.String(),
	}
	_, err = msgServer.ClaimOperatorAddress(ctx, claimMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "only the proposed operator can claim the role")

	// Verify that the operator and proposed operator addresses remain unchanged
	operator := k.GetOperatorAddress(ctx)
	require.Equal(t, currentOperator.String(), operator)
	proposed = k.GetProposedOperatorAddress(ctx)
	require.Equal(t, proposedOperator.String(), proposed)
}

func TestClaimOperatorAddress_InvalidSignerFormat(t *testing.T) {
	// Test case: trying to claim operator address with an invalid signer format

	k, ctx := keepertest.TokenwrapperKeeper(t, nil)
	msgServer := keeper.NewMsgServerImpl(k)

	// Setup test addresses
	currentOperator := sdk.AccAddress([]byte("current_operator"))
	proposedOperator := sdk.AccAddress([]byte("proposed_operator"))

	// Set initial operator and propose a new operator
	k.SetOperatorAddress(ctx, currentOperator.String())
	proposeMsg := &types.MsgProposeOperatorAddress{
		Signer:      currentOperator.String(),
		NewOperator: proposedOperator.String(),
	}
	_, err := msgServer.ProposeOperatorAddress(ctx, proposeMsg)
	require.NoError(t, err)

	// Test claim with invalid signer format
	claimMsg := &types.MsgClaimOperatorAddress{
		Signer: "invalid_address_format",
	}
	_, err = msgServer.ClaimOperatorAddress(ctx, claimMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "only the proposed operator can claim the role")
}

func TestClaimOperatorAddress_AfterAlreadyClaimed(t *testing.T) {
	// Test case: trying to claim operator address after it has already been claimed

	k, ctx := keepertest.TokenwrapperKeeper(t, nil)
	msgServer := keeper.NewMsgServerImpl(k)

	// Setup test addresses
	currentOperator := sdk.AccAddress([]byte("current_operator"))
	proposedOperator := sdk.AccAddress([]byte("proposed_operator"))

	// Set initial operator and propose a new operator
	k.SetOperatorAddress(ctx, currentOperator.String())
	proposeMsg := &types.MsgProposeOperatorAddress{
		Signer:      currentOperator.String(),
		NewOperator: proposedOperator.String(),
	}
	_, err := msgServer.ProposeOperatorAddress(ctx, proposeMsg)
	require.NoError(t, err)

	// First claim (successful)
	claimMsg := &types.MsgClaimOperatorAddress{
		Signer: proposedOperator.String(),
	}
	_, err = msgServer.ClaimOperatorAddress(ctx, claimMsg)
	require.NoError(t, err)

	// Attempt to claim again with the same address
	_, err = msgServer.ClaimOperatorAddress(ctx, claimMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no operator address has been proposed")
}
