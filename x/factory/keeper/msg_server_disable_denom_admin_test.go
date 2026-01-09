package keeper_test

import (
	"fmt"
	"testing"

	"zigchain/testutil/sample"
	"zigchain/x/factory/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

// Positive test cases

func TestMsgServer_DisableDenomAdmin_Positive(t *testing.T) {
	// Test case: disable denom admin with valid input data

	// SETUP - CREATE DENOM
	// ------------------------------

	// create the Test Fixture first
	f := setupTestFixture(t)

	// call this when the test is done, or exits prematurely
	defer f.ctrl.Finish() // assert that all expectations are met

	// create a new denom with bank keeper mocks
	resp := setupDenomWithBankMocks(f)

	// we will also need to check if the response is correct,
	// and we will set the expected response,
	// so we can compare it
	expectedResp := &types.MsgCreateDenomResponse{
		BankAdmin:           f.creator,
		MetadataAdmin:       f.creator,
		Denom:               "coin." + f.creator + "." + f.subDenom,
		MintingCap:          f.maxSupply,
		CanChangeMintingCap: true,
		URI:                 f.uri,
		URIHash:             f.uriHash,
	}
	// make sure that response is not nil
	require.NotNil(t, resp)
	// compare the expected response with the actual response
	require.Equal(t, expectedResp, resp)

	// DISABLE DENOM ADMIN
	// ------------------------------

	// create a message to disable the denom admin
	msgDisable := &types.MsgDisableDenomAdmin{
		Denom:  f.fullDenom,
		Signer: f.creator,
	}

	// disable the denom admin
	respDisable, err := f.server.DisableDenomAdmin(f.ctx, msgDisable)

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct,
	// and we will set the expected response,
	// so we can compare it
	expectedRespDisable := &types.MsgDisableDenomAdminResponse{
		Denom: "coin." + f.creator + "." + f.subDenom,
	}
	// make sure that response is not nil
	require.NotNil(t, respDisable)
	// compare the expected response with the actual response
	require.Equal(t, expectedRespDisable, respDisable)

	// get the denom auth
	denomAuth, isFound := f.k.GetDenomAuth(f.ctx, f.fullDenom)
	require.True(t, isFound)

	// check if the denom auth is correct
	require.Equal(t, "", denomAuth.BankAdmin)
	require.Equal(t, f.creator, denomAuth.MetadataAdmin)

	// Check if there is any pending Admin proposal for the denom is empty or not existing
	pendingAuth, found := f.k.GetProposedDenomAuth(f.ctx, f.fullDenom)
	require.False(t, found, "Expected no pending denom auth proposal for the denom")
	require.Equal(t, types.DenomAuth{
		Denom:         "",
		BankAdmin:     "",
		MetadataAdmin: "",
	},
		pendingAuth,
		"Expected pending denom auth proposal to be empty after disabling the denom admin",
	)

	// EVENT CHECKS
	// ------------------------------------------------------------

	// check that the event was emitted
	events := f.ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// find the pauser address added event
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventDenomAuthDisabled {
			foundEvent = true
			// check event attributes
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyModule {
					require.Equal(t, "factory", string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeySigner {
					require.Equal(t, f.creator, string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyDenom {
					require.Equal(t, f.fullDenom, string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyBankAdmin {
					require.Equal(t, f.creator, string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyMetadataAdmin {
					require.Equal(t, f.creator, string(attr.Value))
				}
			}
			break
		}
	}
	require.True(t, foundEvent, "Expected EventDenomAuthClaimed event to be emitted")
}

// Negative test cases

func TestMsgServer_DisableDenomAdmin_DenomNotExist(t *testing.T) {
	// Test case: disable denom admin for a non-existing denom

	// SETUP
	// ------------------------------

	// create the Test Fixture first
	f := setupTestFixture(t)

	// call this when the test is done, or exits prematurely
	defer f.ctrl.Finish() // assert that all expectations are met

	// DISABLE DENOM ADMIN
	// ------------------------------

	// create a message to disable the denom admin
	msgDisable := &types.MsgDisableDenomAdmin{
		Denom:  f.fullDenom,
		Signer: f.creator,
	}

	// disable the denom auth
	_, err := f.server.DisableDenomAdmin(f.ctx, msgDisable)

	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrDenomAuthNotFound)

	// check if the error message is correct
	require.Equal(
		t,
		fmt.Sprintf("Denom: (%s): denom auth not found", f.fullDenom),
		err.Error(),
	)

	// EVENT CHECKS
	// ------------------------------------------------------------

	// check that the event was not emitted
	events := f.ctx.EventManager().Events()

	// Ensure that the event EventDenomAuthClaimed has not occurred
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventDenomAuthDisabled {
			foundEvent = true
		}
	}
	require.False(t, foundEvent)
}

func TestMsgServer_DisableDenomAdmin_DenomLocked(t *testing.T) {
	// Test case: try to disable denom auth if denom is locked

	// SETUP - CREATE DENOM
	// ------------------------------

	// create the Test Fixture first
	f := setupTestFixture(t)

	// call this when the test is done, or exits prematurely
	defer f.ctrl.Finish() // assert that all expectations are met

	// create a new denom with bank keeper mocks
	resp := setupDenomWithBankMocks(f)

	// we will also need to check if the response is correct,
	// and we will set the expected response,
	// so we can compare it
	expectedResp := &types.MsgCreateDenomResponse{
		BankAdmin:           f.creator,
		MetadataAdmin:       f.creator,
		Denom:               "coin." + f.creator + "." + f.subDenom,
		MintingCap:          f.maxSupply,
		CanChangeMintingCap: true,
		URI:                 f.uri,
		URIHash:             f.uriHash,
	}
	// make sure that response is not nil
	require.NotNil(t, resp)
	// compare the expected response with the actual response
	require.Equal(t, expectedResp, resp)

	// DISABLE DENOM ADMIN
	// ------------------------------

	// create a message to disable a bank admin to empty string
	disableMsg := &types.MsgDisableDenomAdmin{
		Denom:  f.fullDenom,
		Signer: f.creator,
	}

	// disable the denom auth
	_, err := f.server.DisableDenomAdmin(f.ctx, disableMsg)

	// require Error to be nil
	require.Nil(t, err)

	// create a message to disable the denom auth
	disableMsg2 := &types.MsgDisableDenomAdmin{
		Denom:  f.fullDenom,
		Signer: f.creator,
	}

	// disable the denom auth
	_, err = f.server.DisableDenomAdmin(f.ctx, disableMsg2)

	require.Error(t, err)

	// check if the error message is correct
	require.Equal(
		t,
		fmt.Sprintf("Auth: bank admin disabled (%s): denom changes are permanently disabled", f.fullDenom),
		err.Error(),
	)

	// EVENT CHECKS
	// ------------------------------------------------------------

	// check that the event was not emitted
	events := f.ctx.EventManager().Events()

	// Ensure that the event EventDenomAuthClaimed has not occurred
	var foundEvent int
	for _, event := range events {
		if event.Type == types.EventDenomAuthDisabled {
			foundEvent += 1
		}
	}
	require.Equal(t, 1, foundEvent, "Expected EventDenomAuthDisabled event to be emitted only once")
}

func TestMsgServer_DisableDenomAdmin_NotAdmin(t *testing.T) {
	// Test case: try to disable denom

	// SETUP - CREATE DENOM
	// ------------------------------

	// create the Test Fixture first
	f := setupTestFixture(t)

	// call this when the test is done, or exits prematurely
	defer f.ctrl.Finish() // assert that all expectations are met

	// create a new denom with bank keeper mocks
	resp := setupDenomWithBankMocks(f)

	// we will also need to check if the response is correct,
	// and we will set the expected response,
	// so we can compare it
	expectedResp := &types.MsgCreateDenomResponse{
		BankAdmin:           f.creator,
		MetadataAdmin:       f.creator,
		Denom:               "coin." + f.creator + "." + f.subDenom,
		MintingCap:          f.maxSupply,
		CanChangeMintingCap: true,
		URI:                 f.uri,
		URIHash:             f.uriHash,
	}
	// make sure that response is not nil
	require.NotNil(t, resp)
	// compare the expected response with the actual response
	require.Equal(t, expectedResp, resp)

	// DISABLE DENOM ADMIN
	// ------------------------------

	newAddress := sample.AccAddress()

	// create a message to disable a bank admin to empty string
	disableMsg := &types.MsgDisableDenomAdmin{
		Denom:  f.fullDenom,
		Signer: newAddress,
	}

	// disable the denom auth
	_, err := f.server.DisableDenomAdmin(f.ctx, disableMsg)

	require.Error(t, err)
	require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)

	// check if the error message is correct
	require.Equal(
		t,
		fmt.Sprintf("Auth: incorrect admin for denom (%s) only the current bank admin (%s) can perform this action (attempted with: %s): unauthorized",
			f.fullDenom,
			f.creator,
			newAddress,
		),
		err.Error(),
	)

	// EVENT CHECKS
	// ------------------------------------------------------------

	// check that the event was not emitted
	events := f.ctx.EventManager().Events()

	// Ensure that the event EventDenomAuthClaimed has not occurred
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventDenomAuthDisabled {
			foundEvent = true
		}
	}
	require.False(t, foundEvent, "Expected EventDenomAuthDisabled event to not be emitted")
}

func TestMsgServer_DisableDenomAdmin_InvalidDenomFormat(t *testing.T) {
	// Test case: attempt to disable denom admin with an invalid denom format

	f := setupTestFixture(t)
	defer f.ctrl.Finish()

	// Generate a valid creator address
	creator := sample.AccAddress()

	// Create a denom auth with a valid denom manually to ensure it exists
	validDenom := "coin." + creator + ".subdenom"
	denomAuth := types.DenomAuth{
		Denom:         validDenom,
		BankAdmin:     creator,
		MetadataAdmin: creator,
	}
	f.k.SetDenomAuth(f.ctx, denomAuth)

	// Create a message with an invalid denom format
	invalidDenom := "invalid-denom-format"
	msgDisable := &types.MsgDisableDenomAdmin{
		Denom:  invalidDenom,
		Signer: creator,
	}

	// Attempt to disable the denom admin
	resp, err := f.server.DisableDenomAdmin(f.ctx, msgDisable)

	// Check that an error is returned and the response is nil
	require.Error(t, err, "Expected an error when attempting to disable denom admin with invalid denom format")
	require.Contains(t, err.Error(), "Denom: (invalid-denom-format): denom auth not found", "Expected error to indicate invalid denom format")
	require.Nil(t, resp, "Expected nil response when attempting to disable denom admin with invalid denom format")

	// EVENT CHECKS
	// ------------------------------
	events := f.ctx.EventManager().Events()
	for _, event := range events {
		require.NotEqual(t, types.EventDenomAuthDisabled, event.Type, "Expected no EventDenomAuthDisabled event to be emitted on failure")
	}
}

func TestMsgServer_DisableDenomAdmin_AlreadyDisabled(t *testing.T) {
	// Test case: attempt to disable denom admin when it is already disabled

	// SETUP - CREATE DENOM
	// ------------------------------
	f := setupTestFixture(t)
	defer f.ctrl.Finish() // assert that all expectations are met

	// create a new denom with bank keeper mocks
	resp := setupDenomWithBankMocks(f)

	// verify the response
	expectedResp := &types.MsgCreateDenomResponse{
		BankAdmin:           f.creator,
		MetadataAdmin:       f.creator,
		Denom:               "coin." + f.creator + "." + f.subDenom,
		MintingCap:          f.maxSupply,
		CanChangeMintingCap: true,
		URI:                 f.uri,
		URIHash:             f.uriHash,
	}
	require.NotNil(t, resp)
	require.Equal(t, expectedResp, resp)

	// DISABLE DENOM ADMIN (FIRST TIME)
	// ------------------------------
	msgDisable := &types.MsgDisableDenomAdmin{
		Denom:  f.fullDenom,
		Signer: f.creator,
	}
	respDisable, err := f.server.DisableDenomAdmin(f.ctx, msgDisable)
	require.NoError(t, err)
	require.NotNil(t, respDisable)
	require.Equal(t, &types.MsgDisableDenomAdminResponse{Denom: f.fullDenom}, respDisable)

	// Verify that the denom admin is disabled
	denomAuth, isFound := f.k.GetDenomAuth(f.ctx, f.fullDenom)
	require.True(t, isFound)
	require.Equal(t, "", denomAuth.BankAdmin, "Expected bank admin to be empty after first disable")

	// ATTEMPT TO DISABLE AGAIN
	// ------------------------------
	msgDisableAgain := &types.MsgDisableDenomAdmin{
		Denom:  f.fullDenom,
		Signer: f.creator,
	}
	respDisableAgain, err := f.server.DisableDenomAdmin(f.ctx, msgDisableAgain)

	// Check that an error is returned and the response is nil
	require.Error(t, err, "Expected an error when attempting to disable an already disabled denom admin")
	require.Equal(t,
		fmt.Sprintf("Auth: bank admin disabled (%s): denom changes are permanently disabled", f.fullDenom),
		err.Error(),
		"Expected error message to indicate denom changes are permanently disabled")
	require.Nil(t, respDisableAgain, "Expected nil response when attempting to disable an already disabled denom admin")

	// EVENT CHECKS
	// ------------------------------
	events := f.ctx.EventManager().Events()
	foundEventCount := 0
	for _, event := range events {
		if event.Type == types.EventDenomAuthDisabled {
			foundEventCount++
		}
	}
	require.Equal(t, 1, foundEventCount, "Expected EventDenomAuthDisabled event to be emitted only once (from first disable)")
}
