package keeper_test

import (
	"fmt"
	"testing"

	cosmosmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	keepertest "zigchain/testutil/keeper"
	"zigchain/testutil/sample"
	"zigchain/x/factory/keeper"
	"zigchain/x/factory/testutil"
	"zigchain/x/factory/types"
)

// Fixtures

// TestFixture holds common test data and provides setup utilities
type TestFixture struct {
	ctrl       *gomock.Controller
	k          keeper.Keeper
	ctx        sdk.Context
	server     types.MsgServer
	bankKeeper *testutil.MockBankKeeper
	creator    string
	signer     sdk.AccAddress
	subDenom   string
	fullDenom  string
	maxSupply  cosmosmath.Uint
	uri        string
	uriHash    string
	fee        sdk.Coin
	denomMeta  banktypes.Metadata
}

// setupTestFixture creates and initializes a test fixture with common test data
func setupTestFixture(t *testing.T) *TestFixture {
	ctrl := gomock.NewController(t)

	creator := sample.AccAddress()
	signer := sdk.MustAccAddressFromBech32(creator)
	subDenom := "abc"
	maxSupply := cosmosmath.NewUint(1000000)
	uri := "ipfs://uri"
	uriHash := "sha256hash"
	fee := sdk.NewCoin("uzig", cosmosmath.NewInt(1000))
	fullDenom := "coin." + creator + "." + subDenom

	denomMeta := banktypes.Metadata{
		Description: "",
		DenomUnits: []*banktypes.DenomUnit{{
			Denom:    fullDenom,
			Exponent: 0,
		}},
		Base:    fullDenom,
		Name:    fullDenom,
		Symbol:  subDenom,
		Display: fullDenom,
		URI:     uri,
		URIHash: uriHash,
	}

	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, nil)
	server := keeper.NewMsgServerImpl(k)

	return &TestFixture{
		ctrl:       ctrl,
		k:          k,
		ctx:        ctx,
		server:     server,
		bankKeeper: bankKeeper,
		creator:    creator,
		signer:     signer,
		subDenom:   subDenom,
		fullDenom:  fullDenom,
		maxSupply:  maxSupply,
		uri:        uri,
		uriHash:    uriHash,
		fee:        fee,
		denomMeta:  denomMeta,
	}
}

// setupDenomWithBankMocks sets up bank keeper expectations and creates a denom
func setupDenomWithBankMocks(f *TestFixture) *types.MsgCreateDenomResponse {
	// Setup bank keeper expectations
	f.bankKeeper.
		EXPECT().
		HasSupply(gomock.Any(), f.subDenom).
		Return(false).
		Times(1)

	f.bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), f.signer, f.fee).
		Return(true).
		Times(1)

	f.bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), f.signer, types.ModuleName, sdk.NewCoins(f.fee)).
		Return(nil).
		Times(1)

	f.bankKeeper.
		EXPECT().
		GetDenomMetaData(gomock.Any(), f.fullDenom).
		Return(banktypes.Metadata{}, false).
		Times(1)

	f.bankKeeper.
		EXPECT().
		SetDenomMetaData(gomock.Any(), f.denomMeta).
		Return().
		Times(1)

	// Create denom
	msg := &types.MsgCreateDenom{
		Creator:             f.creator,
		SubDenom:            f.subDenom,
		MintingCap:          f.maxSupply,
		CanChangeMintingCap: true,
		URI:                 f.uri,
		URIHash:             f.uriHash,
	}

	resp, err := f.server.CreateDenom(f.ctx, msg)
	if err != nil {
		panic(fmt.Sprintf("Failed to create denom in setup: %v", err))
	}

	return resp
}

// Positive test cases

func TestMsgServer_ClaimDenomAdmin_Positive(t *testing.T) {
	// Test case: claim a proposed change on denom admin

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

	// PROPOSE NEW DENOM BANK ADMIN AND METADATA ADMIN
	// ------------------------------------------------------------

	newBankAdmin := sample.AccAddress()
	newMetaAdmin := sample.AccAddress()

	// create a message to propose a new denom bank admin and metadata admin
	proposeMsg := &types.MsgProposeDenomAdmin{
		Denom:         f.fullDenom,
		BankAdmin:     newBankAdmin,
		MetadataAdmin: newMetaAdmin,
		// since we are bypassing the grpc mechanism, we need to set the signer that is automatically set by the grpc
		Signer: f.creator,
	}

	// propose the denom auth
	proposeResp, err := f.server.ProposeDenomAdmin(f.ctx, proposeMsg)

	// CHECKS
	// ------------------------------------------------------------

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct,
	// and we will set the expected response,
	// so we can compare it
	expectedProposeResp := &types.MsgProposeDenomAdminResponse{
		Denom:         f.fullDenom,
		BankAdmin:     newBankAdmin,
		MetadataAdmin: newMetaAdmin,
	}

	// make sure that response is not nil
	require.NotNil(t, proposeResp)
	// compare the expected response with the actual response
	require.Equal(t, expectedProposeResp, proposeResp)

	// Ensure that in the token the admin is still the creator
	denomAuth, isFound := f.k.GetDenomAuth(f.ctx, f.fullDenom)
	require.True(t, isFound)
	require.Equal(t, f.creator, denomAuth.BankAdmin)
	require.Equal(t, f.creator, denomAuth.MetadataAdmin)

	// Ensure that the proposed DenomAuth is stored correctly
	proposedDenomAuth, isFound := f.k.GetProposedDenomAuth(f.ctx, f.fullDenom)
	require.True(t, isFound)
	require.Equal(t, f.fullDenom, proposedDenomAuth.Denom)
	require.Equal(t, newBankAdmin, proposedDenomAuth.BankAdmin)
	require.Equal(t, newMetaAdmin, proposedDenomAuth.MetadataAdmin)

	// CLAIM THE PROPOSED DENOM AUTH
	// ------------------------------------------------------------

	// create a message to claim the proposed denom admin
	claimMsg := &types.MsgClaimDenomAdmin{
		Denom:  f.fullDenom,
		Signer: newBankAdmin, // the new bank admin is the one claiming
	}

	// claim the denom admin
	claimResp, err := f.server.ClaimDenomAdmin(f.ctx, claimMsg)
	// check if the response is correct
	require.NoError(t, err)
	// we will also need to check if the response is correct,
	expectedClaimResp := &types.MsgClaimDenomAdminResponse{
		Denom: f.fullDenom,
	}

	// make sure that response is not nil
	require.NotNil(t, claimResp)
	// compare the expected response with the actual response
	require.Equal(t, expectedClaimResp, claimResp)

	// Ensure that the denom auth is updated correctly
	denomAuth, isFound = f.k.GetDenomAuth(f.ctx, f.fullDenom)
	require.True(t, isFound)
	require.Equal(t, newBankAdmin, denomAuth.BankAdmin)
	require.Equal(t, newMetaAdmin, denomAuth.MetadataAdmin)

	// Ensure that the proposed DenomAuth is cleared
	proposedDenomAuth, isFound = f.k.GetProposedDenomAuth(f.ctx, f.fullDenom)
	require.False(t, isFound, "Expected proposed denom auth to be removed after claiming")

	// EVENT CHECKS
	// ------------------------------------------------------------

	// check that the event was emitted
	events := f.ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// find the pauser address added event
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventDenomAuthClaimed {
			foundEvent = true
			// check event attributes
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyModule {
					require.Equal(t, "factory", string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeySigner {
					require.Equal(t, newBankAdmin, string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyDenom {
					require.Equal(t, f.fullDenom, string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyBankAdmin {
					require.Equal(t, newBankAdmin, string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyMetadataAdmin {
					require.Equal(t, newMetaAdmin, string(attr.Value))
				}
			}
			break
		}
	}
	require.True(t, foundEvent, "Expected EventDenomAuthClaimed event to be emitted")
}

// Negative test cases

func TestMsgServer_ClaimDenomAdmin_OtherDenom(t *testing.T) {
	// Test case: claim a proposed change on denom admin for a different denom

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

	// PROPOSE NEW DENOM BANK ADMIN AND METADATA ADMIN
	// ------------------------------------------------------------

	newBankAdmin := sample.AccAddress()
	newMetaAdmin := sample.AccAddress()

	// create a message to propose a new denom bank admin and metadata admin
	proposeMsg := &types.MsgProposeDenomAdmin{
		Denom:         f.fullDenom,
		BankAdmin:     newBankAdmin,
		MetadataAdmin: newMetaAdmin,
		// since we are bypassing the grpc mechanism, we need to set the signer that is automatically set by the grpc
		Signer: f.creator,
	}

	// propose the denom auth
	proposeResp, err := f.server.ProposeDenomAdmin(f.ctx, proposeMsg)

	// CHECKS
	// ------------------------------------------------------------

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct,
	// and we will set the expected response,
	// so we can compare it
	expectedProposeResp := &types.MsgProposeDenomAdminResponse{
		Denom:         f.fullDenom,
		BankAdmin:     newBankAdmin,
		MetadataAdmin: newMetaAdmin,
	}

	// make sure that response is not nil
	require.NotNil(t, proposeResp)
	// compare the expected response with the actual response
	require.Equal(t, expectedProposeResp, proposeResp)

	// Ensure that in the token the admin is still the creator
	denomAuth, isFound := f.k.GetDenomAuth(f.ctx, f.fullDenom)
	require.True(t, isFound)
	require.Equal(t, f.creator, denomAuth.BankAdmin)
	require.Equal(t, f.creator, denomAuth.MetadataAdmin)

	// Ensure that the proposed DenomAuth is stored correctly
	proposedDenomAuth, isFound := f.k.GetProposedDenomAuth(f.ctx, f.fullDenom)
	require.True(t, isFound)
	require.Equal(t, f.fullDenom, proposedDenomAuth.Denom)
	require.Equal(t, newBankAdmin, proposedDenomAuth.BankAdmin)
	require.Equal(t, newMetaAdmin, proposedDenomAuth.MetadataAdmin)

	// CLAIM THE PROPOSED DENOM AUTH - DIFFERENT DENOM
	// ------------------------------------------------------------

	newDenom := "coin." + sample.AccAddress() + ".xyz"

	// create a message to claim the proposed denom admin
	claimMsg := &types.MsgClaimDenomAdmin{
		Denom:  newDenom,
		Signer: newBankAdmin, // the new bank admin is the one claiming
	}

	// claim the denom admin
	_, err = f.server.ClaimDenomAdmin(f.ctx, claimMsg)

	// CHECKS
	// ------------------------------------------------------------

	// check if the response is correct
	require.Error(t, err)

	require.ErrorIs(t, err, types.ErrDenomAuthNotFound)

	// check if the error message is correct
	require.Equal(
		t,
		fmt.Sprintf("Denom: (%s): denom auth not found", newDenom),
		err.Error(),
	)

	// Ensure that the denom auth is still the same for the original denom
	denomAuth, isFound = f.k.GetDenomAuth(f.ctx, f.fullDenom)
	require.True(t, isFound)
	require.Equal(t, f.creator, denomAuth.BankAdmin)
	require.Equal(t, f.creator, denomAuth.MetadataAdmin)

	// Ensure that the proposed DenomAuth is still the one proposed
	proposedDenomAuth, isFound = f.k.GetProposedDenomAuth(f.ctx, f.fullDenom)
	require.True(t, isFound)
	require.Equal(t, f.fullDenom, proposedDenomAuth.Denom)
	require.Equal(t, newBankAdmin, proposedDenomAuth.BankAdmin)
	require.Equal(t, newMetaAdmin, proposedDenomAuth.MetadataAdmin)

	// EVENT CHECKS
	// ------------------------------------------------------------

	// check that the event was not emitted
	events := f.ctx.EventManager().Events()

	// Ensure that the event EventDenomAuthClaimed has not occurred
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventDenomAuthClaimed {
			foundEvent = true
		}
	}
	require.False(t, foundEvent, "Expected EventDenomAuthClaimed event NOT to be emitted")
}

func TestMsgServer_ClaimDenomAdmin_DisabledDenom(t *testing.T) {
	// Test case: claim on a demon that it has been locked / disabled

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

	// PROPOSE NEW DENOM BANK ADMIN AND METADATA ADMIN
	// ------------------------------------------------------------

	newBankAdmin := sample.AccAddress()
	newMetaAdmin := sample.AccAddress()

	// create a message to propose a new denom bank admin and metadata admin
	proposeMsg := &types.MsgProposeDenomAdmin{
		Denom:         f.fullDenom,
		BankAdmin:     newBankAdmin,
		MetadataAdmin: newMetaAdmin,
		// since we are bypassing the grpc mechanism, we need to set the signer that is automatically set by the grpc
		Signer: f.creator,
	}

	// propose the denom auth
	proposeResp, err := f.server.ProposeDenomAdmin(f.ctx, proposeMsg)

	// CHECKS
	// ------------------------------------------------------------

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct,
	// and we will set the expected response,
	// so we can compare it
	expectedProposeResp := &types.MsgProposeDenomAdminResponse{
		Denom:         f.fullDenom,
		BankAdmin:     newBankAdmin,
		MetadataAdmin: newMetaAdmin,
	}

	// make sure that response is not nil
	require.NotNil(t, proposeResp)
	// compare the expected response with the actual response
	require.Equal(t, expectedProposeResp, proposeResp)

	// Ensure that in the token the admin is still the creator
	denomAuth, isFound := f.k.GetDenomAuth(f.ctx, f.fullDenom)
	require.True(t, isFound)
	require.Equal(t, f.creator, denomAuth.BankAdmin)
	require.Equal(t, f.creator, denomAuth.MetadataAdmin)

	// Ensure that the proposed DenomAuth is stored correctly
	proposedDenomAuth, isFound := f.k.GetProposedDenomAuth(f.ctx, f.fullDenom)
	require.True(t, isFound)
	require.Equal(t, f.fullDenom, proposedDenomAuth.Denom)
	require.Equal(t, newBankAdmin, proposedDenomAuth.BankAdmin)
	require.Equal(t, newMetaAdmin, proposedDenomAuth.MetadataAdmin)

	// DISABLE THE DENOM AUTH
	// ------------------------------------------------------------
	// create a message to disable the denom auth
	disableMsg := &types.MsgDisableDenomAdmin{
		Denom:  f.fullDenom,
		Signer: f.creator,
	}

	// disable the denom auth
	disableResp, err := f.server.DisableDenomAdmin(f.ctx, disableMsg)

	// check if the response is correct
	require.NoError(t, err)
	// we will also need to check if the response is correct,
	expectedDisableResp := &types.MsgDisableDenomAdminResponse{
		Denom: f.fullDenom,
	}

	// make sure that response is not nil
	require.NotNil(t, disableResp)
	// compare the expected response with the actual response
	require.Equal(t, expectedDisableResp, disableResp)

	// TRY TO CLAIM THE PROPOSED DENOM AUTH
	// ------------------------------------------------------------

	// create a message to claim the proposed denom admin
	claimMsg := &types.MsgClaimDenomAdmin{
		Denom:  f.fullDenom,
		Signer: newBankAdmin,
	}

	// claim the denom admin
	_, err = f.server.ClaimDenomAdmin(f.ctx, claimMsg)

	// CHECKS
	// ------------------------------------------------------------

	// check if the error is correct
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrDenomLocked)
	require.Equal(
		t,
		fmt.Sprintf("denom admin was permanently disabled for denom: %s: denom changes are permanently disabled", f.fullDenom),
		err.Error(),
	)

	// Ensure that the denom auth is still the same
	denomAuth, isFound = f.k.GetDenomAuth(f.ctx, f.fullDenom)
	require.True(t, isFound)
	require.Equal(t, "", denomAuth.BankAdmin)
	require.Equal(t, f.creator, denomAuth.MetadataAdmin)

	// Ensure that the proposed DenomAuth is still the one proposed
	proposedDenomAuth, isFound = f.k.GetProposedDenomAuth(f.ctx, f.fullDenom)
	require.True(t, isFound)
	require.Equal(t, f.fullDenom, proposedDenomAuth.Denom)
	require.Equal(t, newBankAdmin, proposedDenomAuth.BankAdmin)
	require.Equal(t, newMetaAdmin, proposedDenomAuth.MetadataAdmin)

	// EVENT CHECKS
	// ------------------------------------------------------------

	// check that the event was not emitted
	events := f.ctx.EventManager().Events()

	// Ensure that the event EventDenomAuthClaimed has not occurred
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventDenomAuthClaimed {
			foundEvent = true
		}
	}
	require.False(t, foundEvent, "Expected EventDenomAuthClaimed event NOT to be emitted")

}

func TestMsgServer_ClaimDenomAdmin_NoProposal(t *testing.T) {
	// Test case: claim a change on denom admin when there is no proposal

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

	// CLAIM THE PROPOSED DENOM AUTH WITHOUT PROPOSAL
	// ------------------------------------------------------------

	address := sample.AccAddress()

	// create a message to claim the proposed denom admin
	claimMsg := &types.MsgClaimDenomAdmin{
		Denom:  f.fullDenom,
		Signer: address,
	}

	// claim the denom admin
	_, err := f.server.ClaimDenomAdmin(f.ctx, claimMsg)

	// CHECKS
	// ------------------------------------------------------------

	// check if the response is correct
	require.Error(t, err)
	// check if the error is correct
	require.ErrorIs(t, err, types.ErrNoAdminProposal)

	// check if the error message is correct
	require.Equal(
		t,
		fmt.Sprintf(
			"no admin has been proposed for denom %s: no admin has been proposed for denom",
			f.fullDenom,
		),
		err.Error(),
	)

	// Ensure that the denom auth is still the same
	denomAuth, isFound := f.k.GetDenomAuth(f.ctx, f.fullDenom)
	require.True(t, isFound)
	require.Equal(t, f.creator, denomAuth.BankAdmin)
	require.Equal(t, f.creator, denomAuth.MetadataAdmin)

	// Ensure that the proposed DenomAuth doesn't exist
	_, isFound = f.k.GetProposedDenomAuth(f.ctx, f.fullDenom)
	require.False(t, isFound)

	// EVENT CHECKS
	// ------------------------------------------------------------

	// check that the event was not emitted
	events := f.ctx.EventManager().Events()

	// Ensure that the event EventDenomAuthClaimed has not occurred
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventDenomAuthClaimed {
			foundEvent = true
		}
	}
	require.False(t, foundEvent, "Expected EventDenomAuthClaimed event NOT to be emitted")
}

func TestMsgServer_ClaimDenomAdmin_ClaimerNotProposed(t *testing.T) {
	// Test case: claim a proposed change on denom admin with a claimer that is not the proposed admin

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

	// PROPOSE NEW DENOM BANK ADMIN AND METADATA ADMIN
	// ------------------------------------------------------------

	newBankAdmin := sample.AccAddress()
	newMetaAdmin := sample.AccAddress()

	// create a message to propose a new denom bank admin and metadata admin
	proposeMsg := &types.MsgProposeDenomAdmin{
		Denom:         f.fullDenom,
		BankAdmin:     newBankAdmin,
		MetadataAdmin: newMetaAdmin,
		// since we are bypassing the grpc mechanism, we need to set the signer that is automatically set by the grpc
		Signer: f.creator,
	}

	// propose the denom auth
	proposeResp, err := f.server.ProposeDenomAdmin(f.ctx, proposeMsg)

	// CHECKS
	// ------------------------------------------------------------

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct,
	// and we will set the expected response,
	// so we can compare it
	expectedProposeResp := &types.MsgProposeDenomAdminResponse{
		Denom:         f.fullDenom,
		BankAdmin:     newBankAdmin,
		MetadataAdmin: newMetaAdmin,
	}

	// make sure that response is not nil
	require.NotNil(t, proposeResp)
	// compare the expected response with the actual response
	require.Equal(t, expectedProposeResp, proposeResp)

	// Ensure that in the token the admin is still the creator
	denomAuth, isFound := f.k.GetDenomAuth(f.ctx, f.fullDenom)
	require.True(t, isFound)
	require.Equal(t, f.creator, denomAuth.BankAdmin)
	require.Equal(t, f.creator, denomAuth.MetadataAdmin)

	// Ensure that the proposed DenomAuth is stored correctly
	proposedDenomAuth, isFound := f.k.GetProposedDenomAuth(f.ctx, f.fullDenom)
	require.True(t, isFound)
	require.Equal(t, f.fullDenom, proposedDenomAuth.Denom)
	require.Equal(t, newBankAdmin, proposedDenomAuth.BankAdmin)
	require.Equal(t, newMetaAdmin, proposedDenomAuth.MetadataAdmin)

	// CLAIM THE PROPOSED DENOM AUTH
	// ------------------------------------------------------------

	otherAddress := sample.AccAddress()

	// create a message to claim the proposed denom admin
	claimMsg := &types.MsgClaimDenomAdmin{
		Denom:  f.fullDenom,
		Signer: otherAddress, // otherAddress is not the proposed bank admin
	}

	// claim the denom admin
	_, err = f.server.ClaimDenomAdmin(f.ctx, claimMsg)

	// CHECKS
	// ------------------------------------------------------------

	// check if the response is correct
	require.Error(t, err)

	require.ErrorIs(t, err, types.ErrUnauthorizedAdminClaim)

	// check if the error message is correct
	require.Equal(
		t,
		"only the proposed admin can claim the role: only the proposed admin can claim the role",
		err.Error(),
	)

	// Ensure that the denom auth is still the same
	denomAuth, isFound = f.k.GetDenomAuth(f.ctx, f.fullDenom)
	require.True(t, isFound)
	require.Equal(t, f.creator, denomAuth.BankAdmin)
	require.Equal(t, f.creator, denomAuth.MetadataAdmin)

	// Ensure that the proposed DenomAuth is still the one proposed
	proposedDenomAuth, isFound = f.k.GetProposedDenomAuth(f.ctx, f.fullDenom)
	require.True(t, isFound)
	require.Equal(t, f.fullDenom, proposedDenomAuth.Denom)
	require.Equal(t, newBankAdmin, proposedDenomAuth.BankAdmin)
	require.Equal(t, newMetaAdmin, proposedDenomAuth.MetadataAdmin)

	// EVENT CHECKS
	// ------------------------------------------------------------

	// check that the event was not emitted
	events := f.ctx.EventManager().Events()

	// Ensure that the event EventDenomAuthClaimed has not occurred
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventDenomAuthClaimed {
			foundEvent = true
		}
	}
	require.False(t, foundEvent, "Expected EventDenomAuthClaimed event NOT to be emitted")
}

func TestMsgServer_ClaimDenomAdmin_NewMetaAdminTryClaim(t *testing.T) {
	// Test case: claim a proposed change on denom admin with the newMetaAdmin

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

	// PROPOSE NEW DENOM BANK ADMIN AND METADATA ADMIN
	// ------------------------------------------------------------

	newBankAdmin := sample.AccAddress()
	newMetaAdmin := sample.AccAddress()

	// create a message to propose a new denom bank admin and metadata admin
	proposeMsg := &types.MsgProposeDenomAdmin{
		Denom:         f.fullDenom,
		BankAdmin:     newBankAdmin,
		MetadataAdmin: newMetaAdmin,
		// since we are bypassing the grpc mechanism, we need to set the signer that is automatically set by the grpc
		Signer: f.creator,
	}

	// propose the denom auth
	proposeResp, err := f.server.ProposeDenomAdmin(f.ctx, proposeMsg)

	// CHECKS
	// ------------------------------------------------------------

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct,
	// and we will set the expected response,
	// so we can compare it
	expectedProposeResp := &types.MsgProposeDenomAdminResponse{
		Denom:         f.fullDenom,
		BankAdmin:     newBankAdmin,
		MetadataAdmin: newMetaAdmin,
	}

	// make sure that response is not nil
	require.NotNil(t, proposeResp)
	// compare the expected response with the actual response
	require.Equal(t, expectedProposeResp, proposeResp)

	// Ensure that in the token the admin is still the creator
	denomAuth, isFound := f.k.GetDenomAuth(f.ctx, f.fullDenom)
	require.True(t, isFound)
	require.Equal(t, f.creator, denomAuth.BankAdmin)
	require.Equal(t, f.creator, denomAuth.MetadataAdmin)

	// Ensure that the proposed DenomAuth is stored correctly
	proposedDenomAuth, isFound := f.k.GetProposedDenomAuth(f.ctx, f.fullDenom)
	require.True(t, isFound)
	require.Equal(t, f.fullDenom, proposedDenomAuth.Denom)
	require.Equal(t, newBankAdmin, proposedDenomAuth.BankAdmin)
	require.Equal(t, newMetaAdmin, proposedDenomAuth.MetadataAdmin)

	// CLAIM THE PROPOSED DENOM AUTH WITH newMetaAdmin
	// ------------------------------------------------------------

	// create a message to claim the proposed denom admin
	claimMsg := &types.MsgClaimDenomAdmin{
		Denom:  f.fullDenom,
		Signer: newMetaAdmin, // newMetaAdmin is not the proposed bank admin
	}

	// claim the denom admin
	_, err = f.server.ClaimDenomAdmin(f.ctx, claimMsg)

	// CHECKS
	// ------------------------------------------------------------

	// check if the response is correct
	require.Error(t, err)

	require.ErrorIs(t, err, types.ErrUnauthorizedAdminClaim)

	// check if the error message is correct
	require.Equal(
		t,
		"only the proposed admin can claim the role: only the proposed admin can claim the role",
		err.Error(),
	)

	// Ensure that the denom auth is still the same
	denomAuth, isFound = f.k.GetDenomAuth(f.ctx, f.fullDenom)
	require.True(t, isFound)
	require.Equal(t, f.creator, denomAuth.BankAdmin)
	require.Equal(t, f.creator, denomAuth.MetadataAdmin)

	// Ensure that the proposed DenomAuth is still the one proposed
	proposedDenomAuth, isFound = f.k.GetProposedDenomAuth(f.ctx, f.fullDenom)
	require.True(t, isFound)
	require.Equal(t, f.fullDenom, proposedDenomAuth.Denom)
	require.Equal(t, newBankAdmin, proposedDenomAuth.BankAdmin)
	require.Equal(t, newMetaAdmin, proposedDenomAuth.MetadataAdmin)

	// EVENT CHECKS
	// ------------------------------------------------------------

	// check that the event was not emitted
	events := f.ctx.EventManager().Events()

	// Ensure that the event EventDenomAuthClaimed has not occurred
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventDenomAuthClaimed {
			foundEvent = true
		}
	}
	require.False(t, foundEvent, "Expected EventDenomAuthClaimed event NOT to be emitted")
}
