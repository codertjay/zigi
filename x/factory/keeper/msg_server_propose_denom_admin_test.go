package keeper_test

import (
	"fmt"
	"testing"

	cosmosmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	keepertest "zigchain/testutil/keeper"
	"zigchain/testutil/sample"
	"zigchain/x/factory/keeper"
	"zigchain/x/factory/testutil"
	"zigchain/x/factory/types"
)

// Positive test cases

func TestMsgServer_ProposeDenomAdmin_Positive(t *testing.T) {
	// Test case: propose a new denom bank admin and metadata admin
	// with valid input data

	// SETUP - CREATE DENOM
	// ------------------------------

	// create denom first
	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// create a sample signer address
	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create a sample sub denom name
	// the full name will be in format "coin.{creator}.{subDenom}"
	subDenom := "abc"
	mintingCap := cosmosmath.NewUint(1000000)
	uri := "ipfs://uri"
	uriHash := "sha256hash"
	fee := sdk.NewCoin("uzig", cosmosmath.NewInt(1000))

	fullDenom := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + subDenom

	denomMetaData := banktypes.Metadata{
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

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// code will first check if subDenom is matching one of the code tokens (e.g. "uzig")
	bankKeeper.
		EXPECT().
		HasSupply(gomock.Any(), subDenom).
		Return(false).
		Times(1)

	// code will check if the signer has the required balances to pay the fee
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, fee).
		Return(true).
		Times(1)

	// code will deduct the fee from the signer's account
	// SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error
	// it will return nil if the operation is successful - no error
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(fee)).
		Return(nil).
		Times(1)

	// check if denom already has metadata
	// GetDenomMetaData(context.Context, string) (banktypes.Metadata, bool)
	bankKeeper.
		EXPECT().
		GetDenomMetaData(gomock.Any(), fullDenom).
		Return(banktypes.Metadata{}, false).
		Times(1)

	// SetDenomMetaData(context.Context, banktypes.Metadata)
	bankKeeper.
		EXPECT().
		SetDenomMetaData(gomock.Any(), denomMetaData).
		Return().
		Times(1)

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, nil)

	// create a message server
	server := keeper.NewMsgServerImpl(k)

	// create a message to create a new denom
	msg := &types.MsgCreateDenom{
		Creator:             creator,
		SubDenom:            subDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}

	// create a new denom
	resp, err := server.CreateDenom(ctx, msg)

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct,
	// and we will set the expected response,
	// so we can compare it
	expectedResp := &types.MsgCreateDenomResponse{
		BankAdmin:           creator,
		MetadataAdmin:       creator,
		Denom:               "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + subDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
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
		Denom:         fullDenom,
		BankAdmin:     newBankAdmin,
		MetadataAdmin: newMetaAdmin,
		// since we are bypassing the grpc mechanism, we need to set the signer that is automatically set by the grpc
		Signer: creator,
	}

	// propose the denom auth
	proposeResp, err := server.ProposeDenomAdmin(ctx, proposeMsg)

	// CHECKS
	// ------------------------------------------------------------

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct,
	// and we will set the expected response,
	// so we can compare it
	expectedProposeResp := &types.MsgProposeDenomAdminResponse{
		Denom:         fullDenom,
		BankAdmin:     newBankAdmin,
		MetadataAdmin: newMetaAdmin,
	}

	// make sure that response is not nil
	require.NotNil(t, proposeResp)
	// compare the expected response with the actual response
	require.Equal(t, expectedProposeResp, proposeResp)

	// Ensure that in the token the admin is still the creator
	denomAuth, isFound := k.GetDenomAuth(ctx, fullDenom)
	require.True(t, isFound)
	require.Equal(t, creator, denomAuth.BankAdmin)
	require.Equal(t, creator, denomAuth.MetadataAdmin)

	// Ensure that the proposed DenomAuth is stored correctly
	proposedDenomAuth, isFound := k.GetProposedDenomAuth(ctx, fullDenom)
	require.True(t, isFound)
	require.Equal(t, fullDenom, proposedDenomAuth.Denom)
	require.Equal(t, newBankAdmin, proposedDenomAuth.BankAdmin)
	require.Equal(t, newMetaAdmin, proposedDenomAuth.MetadataAdmin)

	// EVENT CHECKS
	// ------------------------------------------------------------

	// check that the event was emitted
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// find the pauser address added event
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventDenomAuthProposed {
			foundEvent = true
			// check event attributes
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyModule {
					require.Equal(t, "factory", string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeySigner {
					require.Equal(t, signer.String(), string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyDenom {
					require.Equal(t, fullDenom, string(attr.Value))
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
	require.True(t, foundEvent, "Expected EventDenomAuthProposed event to be emitted")
}

func TestMsgServer_ProposeDenomAdmin_Positive2_OnlyBankAdmin(t *testing.T) {
	// Test case: propose a new denom bank admin and metadata admin is empty
	// with valid input data

	// SETUP - CREATE DENOM
	// ------------------------------

	// create denom first
	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// create a sample signer address
	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create a sample sub denom name
	// the full name will be in format "coin.{creator}.{subDenom}"
	subDenom := "abc"
	mintingCap := cosmosmath.NewUint(1000000)
	uri := "ipfs://uri"
	uriHash := "sha256hash"
	fee := sdk.NewCoin("uzig", cosmosmath.NewInt(1000))

	fullDenom := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + subDenom

	denomMetaData := banktypes.Metadata{
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

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// code will first check if subDenom is matching one of the code tokens (e.g. "uzig")
	bankKeeper.
		EXPECT().
		HasSupply(gomock.Any(), subDenom).
		Return(false).
		Times(1)

	// code will check if the signer has the required balances to pay the fee
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, fee).
		Return(true).
		Times(1)

	// code will deduct the fee from the signer's account
	// SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error
	// it will return nil if the operation is successful - no error
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(fee)).
		Return(nil).
		Times(1)

	// check if denom already has metadata
	// GetDenomMetaData(context.Context, string) (banktypes.Metadata, bool)
	bankKeeper.
		EXPECT().
		GetDenomMetaData(gomock.Any(), fullDenom).
		Return(banktypes.Metadata{}, false).
		Times(1)

	// SetDenomMetaData(context.Context, banktypes.Metadata)
	bankKeeper.
		EXPECT().
		SetDenomMetaData(gomock.Any(), denomMetaData).
		Return().
		Times(1)

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, nil)

	// create a message server
	server := keeper.NewMsgServerImpl(k)

	// create a message to create a new denom
	msg := &types.MsgCreateDenom{
		Creator:             creator,
		SubDenom:            subDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}

	// create a new denom
	resp, err := server.CreateDenom(ctx, msg)

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct,
	// and we will set the expected response,
	// so we can compare it
	expectedResp := &types.MsgCreateDenomResponse{
		BankAdmin:           creator,
		MetadataAdmin:       creator,
		Denom:               "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + subDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}
	// make sure that response is not nil
	require.NotNil(t, resp)
	// compare the expected response with the actual response
	require.Equal(t, expectedResp, resp)

	// PROPOSE NEW DENOM BANK ADMIN AND METADATA ADMIN
	// ------------------------------------------------------------

	newBankAdmin := sample.AccAddress()

	// create a message to propose a new denom bank admin and metadata admin
	proposeMsg := &types.MsgProposeDenomAdmin{
		Denom:         fullDenom,
		BankAdmin:     newBankAdmin,
		MetadataAdmin: "",
		// since we are bypassing the grpc mechanism, we need to set the signer that is automatically set by the grpc
		Signer: creator,
	}

	// propose the denom auth
	proposeResp, err := server.ProposeDenomAdmin(ctx, proposeMsg)

	// CHECKS
	// ------------------------------------------------------------

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct,
	// and we will set the expected response,
	// so we can compare it
	expectedProposeResp := &types.MsgProposeDenomAdminResponse{
		Denom:         fullDenom,
		BankAdmin:     newBankAdmin,
		MetadataAdmin: "",
	}

	// make sure that response is not nil
	require.NotNil(t, proposeResp)
	// compare the expected response with the actual response
	require.Equal(t, expectedProposeResp, proposeResp)

	// Ensure that in the token the admin is still the creator
	denomAuth, isFound := k.GetDenomAuth(ctx, fullDenom)
	require.True(t, isFound)
	require.Equal(t, creator, denomAuth.BankAdmin)
	require.Equal(t, creator, denomAuth.MetadataAdmin)

	// Ensure that the proposed DenomAuth is stored correctly
	proposedDenomAuth, isFound := k.GetProposedDenomAuth(ctx, fullDenom)
	require.True(t, isFound)
	require.Equal(t, fullDenom, proposedDenomAuth.Denom)
	require.Equal(t, newBankAdmin, proposedDenomAuth.BankAdmin)
	require.Equal(t, "", proposedDenomAuth.MetadataAdmin)

	// EVENT CHECKS
	// ------------------------------------------------------------

	// check that the event was emitted
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// find the pauser address added event
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventDenomAuthProposed {
			foundEvent = true
			// check event attributes
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyModule {
					require.Equal(t, "factory", string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeySigner {
					require.Equal(t, signer.String(), string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyDenom {
					require.Equal(t, fullDenom, string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyBankAdmin {
					require.Equal(t, newBankAdmin, string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyMetadataAdmin {
					require.Equal(t, "", string(attr.Value))
				}
			}
			break
		}
	}
	require.True(t, foundEvent, "Expected EventDenomAuthProposed event to be emitted")
}

func TestMsgServer_ProposeDenomAdmin_Positive3_DoubleProposal(t *testing.T) {
	// Test case: propose a new denom bank admin and metadata admin twice
	// with valid input data so the last one should be the one that is stored

	// SETUP - CREATE DENOM
	// ------------------------------

	// create denom first
	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// create a sample signer address
	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create a sample sub denom name
	// the full name will be in format "coin.{creator}.{subDenom}"
	subDenom := "abc"
	maxSupply := cosmosmath.NewUint(1000000)
	uri := "ipfs://uri"
	uriHash := "sha256hash"
	fee := sdk.NewCoin("uzig", cosmosmath.NewInt(1000))

	fullDenom := "coin." + creator + "." + subDenom

	denomMetaData := banktypes.Metadata{
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

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// code will first check if subDenom is matching one of the code tokens (e.g. "uzig")
	bankKeeper.
		EXPECT().
		HasSupply(gomock.Any(), subDenom).
		Return(false).
		Times(1)

	// code will check if the signer has the required balances to pay the fee
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, fee).
		Return(true).
		Times(1)

	// code will deduct the fee from the signer's account
	// SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error
	// it will return nil if the operation is successful - no error
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(fee)).
		Return(nil).
		Times(1)

	// check if denom already has metadata
	// GetDenomMetaData(context.Context, string) (banktypes.Metadata, bool)
	bankKeeper.
		EXPECT().
		GetDenomMetaData(gomock.Any(), fullDenom).
		Return(banktypes.Metadata{}, false).
		Times(1)

	// SetDenomMetaData(context.Context, banktypes.Metadata)
	bankKeeper.
		EXPECT().
		SetDenomMetaData(gomock.Any(), denomMetaData).
		Return().
		Times(1)

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, nil)

	// create a message server
	server := keeper.NewMsgServerImpl(k)

	// create a message to create a new denom
	msg := &types.MsgCreateDenom{
		Creator:             creator,
		SubDenom:            subDenom,
		MintingCap:          maxSupply,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}

	// create a new denom
	resp, err := server.CreateDenom(ctx, msg)

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct,
	// and we will set the expected response,
	// so we can compare it
	expectedResp := &types.MsgCreateDenomResponse{
		BankAdmin:           creator,
		MetadataAdmin:       creator,
		Denom:               "coin." + creator + "." + subDenom,
		MintingCap:          maxSupply,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
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
		Denom:         fullDenom,
		BankAdmin:     newBankAdmin,
		MetadataAdmin: newMetaAdmin,
		// since we are bypassing the grpc mechanism, we need to set the signer that is automatically set by the grpc
		Signer: creator,
	}

	// propose the denom auth
	proposeResp, err := server.ProposeDenomAdmin(ctx, proposeMsg)

	// CHECKS
	// ------------------------------------------------------------

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct,
	// and we will set the expected response,
	// so we can compare it
	expectedProposeResp := &types.MsgProposeDenomAdminResponse{
		Denom:         fullDenom,
		BankAdmin:     newBankAdmin,
		MetadataAdmin: newMetaAdmin,
	}

	// make sure that response is not nil
	require.NotNil(t, proposeResp)
	// compare the expected response with the actual response
	require.Equal(t, expectedProposeResp, proposeResp)

	// Ensure that in the token the admin is still the creator
	denomAuth, isFound := k.GetDenomAuth(ctx, fullDenom)
	require.True(t, isFound)
	require.Equal(t, creator, denomAuth.BankAdmin)
	require.Equal(t, creator, denomAuth.MetadataAdmin)

	// Ensure that the proposed DenomAuth is stored correctly
	proposedDenomAuth, isFound := k.GetProposedDenomAuth(ctx, fullDenom)
	require.True(t, isFound)
	require.Equal(t, fullDenom, proposedDenomAuth.Denom)
	require.Equal(t, newBankAdmin, proposedDenomAuth.BankAdmin)
	require.Equal(t, newMetaAdmin, proposedDenomAuth.MetadataAdmin)

	// PROPOSE A NEW DENOM BANK ADMIN AND METADATA ADMIN BEFORE THE PREVIOUS ONE IS APPROVED
	// ------------------------------------------------------------

	newBankAdmin2 := sample.AccAddress()
	newMetaAdmin2 := sample.AccAddress()

	// create a message to propose a new denom bank admin and metadata admin
	proposeMsg2 := &types.MsgProposeDenomAdmin{
		Denom:         fullDenom,
		BankAdmin:     newBankAdmin2,
		MetadataAdmin: newMetaAdmin2,
		// since we are bypassing the grpc mechanism, we need to set the signer that is automatically set by the grpc
		Signer: creator,
	}

	// propose the denom auth
	proposeResp2, err := server.ProposeDenomAdmin(ctx, proposeMsg2)

	// CHECKS
	// ------------------------------------------------------------

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct,
	// and we will set the expected response,
	// so we can compare it
	expectedProposeResp2 := &types.MsgProposeDenomAdminResponse{
		Denom:         fullDenom,
		BankAdmin:     newBankAdmin2,
		MetadataAdmin: newMetaAdmin2,
	}

	// make sure that response is not nil
	require.NotNil(t, proposeResp2)
	// compare the expected response with the actual response
	require.Equal(t, expectedProposeResp2, proposeResp2)

	// Ensure that in the token the admin is still the creator
	denomAuth, isFound = k.GetDenomAuth(ctx, fullDenom)
	require.True(t, isFound)
	require.Equal(t, creator, denomAuth.BankAdmin)
	require.Equal(t, creator, denomAuth.MetadataAdmin)

	// Ensure that the proposed DenomAuth is stored correctly
	proposedDenomAuth, isFound = k.GetProposedDenomAuth(ctx, fullDenom)
	require.True(t, isFound)
	require.Equal(t, fullDenom, proposedDenomAuth.Denom)
	require.Equal(t, newBankAdmin2, proposedDenomAuth.BankAdmin)
	require.Equal(t, newMetaAdmin2, proposedDenomAuth.MetadataAdmin)

	// EVENT CHECKS
	// ------------------------------------------------------------

	// check that the event was emitted
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// check that the last event is the EventDenomAuthProposed we are looking for
	lastEvent := events[len(events)-1]
	require.Equal(t, types.EventDenomAuthProposed, lastEvent.Type)

	for _, attr := range lastEvent.Attributes {
		if string(attr.Key) == types.AttributeKeyModule {
			require.Equal(t, "factory", string(attr.Value))
		}
		if string(attr.Key) == types.AttributeKeySigner {
			require.Equal(t, signer.String(), string(attr.Value))
		}
		if string(attr.Key) == types.AttributeKeyDenom {
			require.Equal(t, fullDenom, string(attr.Value))
		}
		if string(attr.Key) == types.AttributeKeyBankAdmin {
			require.Equal(t, newBankAdmin2, string(attr.Value))
		}
		if string(attr.Key) == types.AttributeKeyMetadataAdmin {
			require.Equal(t, newMetaAdmin2, string(attr.Value))
		}
	}
}

// Negative test cases

func TestMsgServer_ProposeDenomAdmin_DenomAuthNotFound(t *testing.T) {
	// Test case: try to propose denom auth if denom auth not found

	// SETUP
	// ------------------------------

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// create a sample sub denom name - the full name will be in format "coin.{creator}.{subDenom}"
	subDenom := "abc"

	fullDenom := "coin." + creator + "." + subDenom

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	// create a message server
	server := keeper.NewMsgServerImpl(k)

	// PROPOSE DENOM AUTH
	// ------------------------------

	newAdmin := sample.AccAddress()

	// create a message to propose the denom auth
	msg := &types.MsgProposeDenomAdmin{
		Denom:         fullDenom,
		BankAdmin:     newAdmin,
		MetadataAdmin: newAdmin,
	}

	// propose the denom auth
	_, err := server.ProposeDenomAdmin(ctx, msg)

	// check if the error is correct
	require.ErrorIs(t, err, types.ErrDenomAuthNotFound)

	// check if the error message is correct
	require.Equal(
		t,
		fmt.Sprintf("Denom: (%s): denom auth not found", fullDenom),
		err.Error(),
	)

	// Ensure that the denom auth is still the same
	_, isFound := k.GetDenomAuth(ctx, fullDenom)
	require.False(t, isFound)

	// Ensure that the proposed DenomAuth is still empty as it has not been set up yet ever
	_, isFound = k.GetProposedDenomAuth(ctx, fullDenom)
	require.False(t, isFound)

	// EVENT CHECKS
	// ------------------------------------------------------------

	// check that the event was not emitted
	events := ctx.EventManager().Events()

	// Ensure that the event EventDenomAuthClaimed has not occurred
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventDenomAuthProposed {
			foundEvent = true
		}
	}
	require.False(t, foundEvent, "Expected EventDenomAuthProposed event NOT to be emitted")
}

func TestMsgServer_ProposeDenomAdmin_Unauthorized(t *testing.T) {
	// Test case: try to propose denom auth if unauthorized

	// SETUP
	// ------------------------------

	// create denom first
	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// create a sample signer address
	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create a sample sub denom name - the full name will be in format "coin.{creator}.{subDenom}"
	subDenom := "abc"
	maxSupply := cosmosmath.NewUint(1000000)
	uri := "ipfs://uri"
	uriHash := "sha256hash"
	fee := sdk.NewCoin("uzig", cosmosmath.NewInt(1000))

	fullDenom := "coin." + creator + "." + subDenom

	denomMetaData := banktypes.Metadata{
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

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// code will first check if subDenom is matching one of the code tokens (e.g. "uzig")
	bankKeeper.
		EXPECT().
		HasSupply(gomock.Any(), subDenom).
		Return(false).
		Times(1)

	// code will check if the signer has the required balances to pay the fee
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, fee).
		Return(true).
		Times(1)

	// code will deduct the fee from the signer's account
	// SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error
	// it will return nil if the operation is successful - no error
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(fee)).
		Return(nil).
		Times(1)

	// check if denom already has metadata
	// GetDenomMetaData(context.Context, string) (banktypes.Metadata, bool)
	bankKeeper.
		EXPECT().
		GetDenomMetaData(gomock.Any(), fullDenom).
		Return(banktypes.Metadata{}, false).
		Times(1)

	// SetDenomMetaData(context.Context, banktypes.Metadata)
	bankKeeper.
		EXPECT().
		SetDenomMetaData(gomock.Any(), denomMetaData).
		Return().
		Times(1)

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, nil)

	// create a message server
	server := keeper.NewMsgServerImpl(k)

	// create a message to create a new denom
	msg := &types.MsgCreateDenom{
		Creator:             creator,
		SubDenom:            subDenom,
		MintingCap:          maxSupply,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}

	// create a new denom
	resp, err := server.CreateDenom(ctx, msg)

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct,
	// and we will set the expected response,
	// so we can compare it
	expectedResp := &types.MsgCreateDenomResponse{
		BankAdmin:           creator,
		MetadataAdmin:       creator,
		Denom:               "coin." + creator + "." + subDenom,
		MintingCap:          maxSupply,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}
	// make sure that response is not nil
	require.NotNil(t, resp)
	// compare the expected response with the actual response
	require.Equal(t, expectedResp, resp)

	// PROPOSE DENOM AUTH
	// ------------------------------

	newBankAdmin := sample.AccAddress()
	newMetaAdmin := sample.AccAddress()

	// create a message to propose the denom auth
	proposeMsg := &types.MsgProposeDenomAdmin{
		Denom:         fullDenom,
		BankAdmin:     newBankAdmin,
		MetadataAdmin: newMetaAdmin,
	}

	// propose the denom auth
	_, err = server.ProposeDenomAdmin(ctx, proposeMsg)

	// CHECKS
	// ------------------------------------------------------------

	// check if the error is correct
	require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)

	// check if the error message is correct
	require.Equal(
		t,
		fmt.Sprintf(
			"Auth: incorrect admin for denom (%s) only the current bank admin (%s) can perform this action (attempted with: ): unauthorized",
			fullDenom,
			creator,
		),
		err.Error(),
	)

	// Ensure that the denom auth is still the same
	denomAuth, isFound := k.GetDenomAuth(ctx, fullDenom)
	require.True(t, isFound)
	require.Equal(t, creator, denomAuth.BankAdmin)
	require.Equal(t, creator, denomAuth.MetadataAdmin)

	// Ensure that the proposed DenomAuth is still empty as it has not been set up yet ever
	_, isFound = k.GetProposedDenomAuth(ctx, fullDenom)
	require.False(t, isFound)

	// EVENT CHECKS
	// ------------------------------------------------------------

	// check that the event was not emitted
	events := ctx.EventManager().Events()

	// Ensure that the event EventDenomAuthClaimed has not occurred
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventDenomAuthProposed {
			foundEvent = true
		}
	}
	require.False(t, foundEvent, "Expected EventDenomAuthProposed event NOT to be emitted")
}

func TestMsgServer_ProposeDenomAdmin_DuplicateBankAdmin(t *testing.T) {
	// Test case: try to propose the same bank admin address

	// create denom first
	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// create a sample signer address
	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create a sample sub denom name - the full name will be in format "coin.{creator}.{subDenom}"
	subDenom := "abc"
	mintingCap := cosmosmath.NewUint(1000000)
	uri := "ipfs://uri"
	uriHash := "sha256hash"
	fee := sdk.NewCoin("uzig", cosmosmath.NewInt(1000))

	fullDenom := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + subDenom

	denomMetaData := banktypes.Metadata{
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

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// code will first check if subDenom is matching one of the code tokens (e.g. "uzig")
	bankKeeper.
		EXPECT().
		HasSupply(gomock.Any(), subDenom).
		Return(false).
		Times(1)

	// code will check if the signer has the required balances to pay the fee
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, fee).
		Return(true).
		Times(1)

	// code will deduct the fee from the signer's account
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(fee)).
		Return(nil).
		Times(1)

	// check if denom already has metadata
	bankKeeper.
		EXPECT().
		GetDenomMetaData(gomock.Any(), fullDenom).
		Return(banktypes.Metadata{}, false).
		Times(1)

	// SetDenomMetaData
	bankKeeper.
		EXPECT().
		SetDenomMetaData(gomock.Any(), denomMetaData).
		Return().
		Times(1)

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, nil)

	// create a message server
	server := keeper.NewMsgServerImpl(k)

	// create a message to create a new denom
	msg := &types.MsgCreateDenom{
		Creator:             creator,
		SubDenom:            subDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}

	// create a new denom
	resp, err := server.CreateDenom(ctx, msg)

	// check if the response is correct
	require.NoError(t, err)

	// verify the response
	expectedResp := &types.MsgCreateDenomResponse{
		BankAdmin:           creator,
		MetadataAdmin:       creator,
		Denom:               "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + subDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}
	require.NotNil(t, resp)
	require.Equal(t, expectedResp, resp)

	// create a message to propose the same bank admin
	proposeMsg := &types.MsgProposeDenomAdmin{
		Denom:         fullDenom,
		BankAdmin:     creator, // Same as current bank admin
		MetadataAdmin: sample.AccAddress(),
		Signer:        creator,
	}

	// propose the denom auth
	_, err = server.ProposeDenomAdmin(ctx, proposeMsg)

	// CHECKS
	// ------------------------------------------------------------

	// check if the error is correct
	require.ErrorIs(t, err, types.ErrDuplicateBankAdminProposal)

	// check if the error message is correct
	require.Equal(
		t,
		fmt.Sprintf("current=%s, proposed=%s: cannot propose the same bank admin address", creator, creator),
		err.Error(),
	)

	// Ensure that the denom auth is still the same
	denomAuth, isFound := k.GetDenomAuth(ctx, fullDenom)
	require.True(t, isFound)
	require.Equal(t, creator, denomAuth.BankAdmin)
	require.Equal(t, creator, denomAuth.MetadataAdmin)

	// Ensure that the proposed DenomAuth is still empty as it has not been set up yet ever
	_, isFound = k.GetProposedDenomAuth(ctx, fullDenom)
	require.False(t, isFound)

	// EVENT CHECKS
	// ------------------------------------------------------------

	// check that the event was not emitted
	events := ctx.EventManager().Events()

	// Ensure that the event EventDenomAuthClaimed has not occurred
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventDenomAuthProposed {
			foundEvent = true
		}
	}
	require.False(t, foundEvent, "Expected EventDenomAuthProposed event NOT to be emitted")
}

func TestMsgServer_ProposeDenomAdmin_InvalidSigner(t *testing.T) {
	// Test case: try to propose denom auth if the signer is not the creator of the denom

	// SETUP
	// ------------------------------

	// create denom first
	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// create a sample signer address
	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create a sample sub denom name - the full name will be in format "coin.{creator}.{subDenom}"
	subDenom := "abc"
	maxSupply := cosmosmath.NewUint(1000000)
	uri := "ipfs://uri"
	uriHash := "sha256hash"
	fee := sdk.NewCoin("uzig", cosmosmath.NewInt(1000))

	fullDenom := "coin." + creator + "." + subDenom

	denomMetaData := banktypes.Metadata{
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

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// code will first check if subDenom is matching one of the code tokens (e.g. "uzig")
	bankKeeper.
		EXPECT().
		HasSupply(gomock.Any(), subDenom).
		Return(false).
		Times(1)

	// code will check if the signer has the required balances to pay the fee
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, fee).
		Return(true).
		Times(1)

	// code will deduct the fee from the signer's account
	// SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error
	// it will return nil if the operation is successful - no error
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(fee)).
		Return(nil).
		Times(1)

	// check if denom already has metadata
	// GetDenomMetaData(context.Context, string) (banktypes.Metadata, bool)
	bankKeeper.
		EXPECT().
		GetDenomMetaData(gomock.Any(), fullDenom).
		Return(banktypes.Metadata{}, false).
		Times(1)

	// SetDenomMetaData(context.Context, banktypes.Metadata)
	bankKeeper.
		EXPECT().
		SetDenomMetaData(gomock.Any(), denomMetaData).
		Return().
		Times(1)

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, nil)

	// create a message server
	server := keeper.NewMsgServerImpl(k)

	// create a message to create a new denom
	msg := &types.MsgCreateDenom{
		Creator:             creator,
		SubDenom:            subDenom,
		MintingCap:          maxSupply,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}

	// create a new denom
	resp, err := server.CreateDenom(ctx, msg)

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct,
	// and we will set the expected response,
	// so we can compare it
	expectedResp := &types.MsgCreateDenomResponse{
		BankAdmin:           creator,
		MetadataAdmin:       creator,
		Denom:               "coin." + creator + "." + subDenom,
		MintingCap:          maxSupply,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}
	// make sure that response is not nil
	require.NotNil(t, resp)
	// compare the expected response with the actual response
	require.Equal(t, expectedResp, resp)

	// PROPOSE DENOM AUTH
	// ------------------------------

	newBankAdmin := sample.AccAddress()
	newMetaAdmin := sample.AccAddress()
	invalidSigner := sample.AccAddress() // using a different address than the creator

	// create a message to propose the denom auth
	proposeMsg := &types.MsgProposeDenomAdmin{
		Denom:         fullDenom,
		BankAdmin:     newBankAdmin,
		MetadataAdmin: newMetaAdmin,
		Signer:        invalidSigner,
	}

	// propose the denom auth
	_, err = server.ProposeDenomAdmin(ctx, proposeMsg)

	// CHECKS
	// ------------------------------------------------------------

	// check if the error is correct
	require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)

	// check if the error message is correct
	require.Equal(
		t,
		fmt.Sprintf(
			"Auth: incorrect admin for denom (%s) only the current bank admin (%s) can perform this action (attempted with: %s): unauthorized",
			fullDenom,
			creator,
			invalidSigner,
		),
		err.Error(),
	)

	// Ensure that the denom auth is still the same
	denomAuth, isFound := k.GetDenomAuth(ctx, fullDenom)
	require.True(t, isFound)
	require.Equal(t, creator, denomAuth.BankAdmin)
	require.Equal(t, creator, denomAuth.MetadataAdmin)

	// Ensure that the proposed DenomAuth is still empty as it has not been set up yet ever
	_, isFound = k.GetProposedDenomAuth(ctx, fullDenom)
	require.False(t, isFound)

	// EVENT CHECKS
	// ------------------------------------------------------------

	// check that the event was not emitted
	events := ctx.EventManager().Events()

	// Ensure that the event EventDenomAuthClaimed has not occurred
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventDenomAuthProposed {
			foundEvent = true
		}
	}
	require.False(t, foundEvent, "Expected EventDenomAuthProposed event NOT to be emitted")
}

func TestMsgServer_ProposeDenomAdmin_InvalidSignerMetaAdmin(t *testing.T) {
	// Test case: try to propose denom auth if the signer is not the creator of the denom,
	// but the signer is the metadata admin

	// SETUP
	// ------------------------------

	// create denom first
	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// create a sample signer address
	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create a sample sub denom name - the full name will be in format "coin.{creator}.{subDenom}"
	subDenom := "abc"
	maxSupply := cosmosmath.NewUint(1000000)
	uri := "ipfs://uri"
	uriHash := "sha256hash"
	fee := sdk.NewCoin("uzig", cosmosmath.NewInt(1000))

	fullDenom := "coin." + creator + "." + subDenom

	denomMetaData := banktypes.Metadata{
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

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// code will first check if subDenom is matching one of the code tokens (e.g. "uzig")
	bankKeeper.
		EXPECT().
		HasSupply(gomock.Any(), subDenom).
		Return(false).
		Times(1)

	// code will check if the signer has the required balances to pay the fee
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, fee).
		Return(true).
		Times(1)

	// code will deduct the fee from the signer's account
	// SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error
	// it will return nil if the operation is successful - no error
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(fee)).
		Return(nil).
		Times(1)

	// check if denom already has metadata
	// GetDenomMetaData(context.Context, string) (banktypes.Metadata, bool)
	bankKeeper.
		EXPECT().
		GetDenomMetaData(gomock.Any(), fullDenom).
		Return(banktypes.Metadata{}, false).
		Times(1)

	// SetDenomMetaData(context.Context, banktypes.Metadata)
	bankKeeper.
		EXPECT().
		SetDenomMetaData(gomock.Any(), denomMetaData).
		Return().
		Times(1)

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, nil)

	// create a message server
	server := keeper.NewMsgServerImpl(k)

	// create a message to create a new denom
	msg := &types.MsgCreateDenom{
		Creator:             creator,
		SubDenom:            subDenom,
		MintingCap:          maxSupply,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}

	// create a new denom
	resp, err := server.CreateDenom(ctx, msg)

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct,
	// and we will set the expected response,
	// so we can compare it
	expectedResp := &types.MsgCreateDenomResponse{
		BankAdmin:           creator,
		MetadataAdmin:       creator,
		Denom:               "coin." + creator + "." + subDenom,
		MintingCap:          maxSupply,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}
	// make sure that response is not nil
	require.NotNil(t, resp)
	// compare the expected response with the actual response
	require.Equal(t, expectedResp, resp)

	// Update the metadata admin to a different address
	denomAuth, _ := k.GetDenomAuth(ctx, fullDenom)

	// Set the metadata admin to a different address
	metaAdmin := sample.AccAddress()
	denomAuth.MetadataAdmin = metaAdmin
	// Update the denom auth in the keeper
	k.SetDenomAuth(ctx, denomAuth)

	// Ensure that the denom auth is updated correctly
	denomAuth, isFound := k.GetDenomAuth(ctx, fullDenom)
	require.True(t, isFound)
	require.Equal(t, creator, denomAuth.BankAdmin)
	require.Equal(t, metaAdmin, denomAuth.MetadataAdmin)

	// PROPOSE DENOM AUTH
	// ------------------------------

	newBankAdmin := sample.AccAddress()
	newMetaAdmin := sample.AccAddress()

	// create a message to propose the denom auth
	proposeMsg := &types.MsgProposeDenomAdmin{
		Denom:         fullDenom,
		BankAdmin:     newBankAdmin,
		MetadataAdmin: newMetaAdmin,
		Signer:        metaAdmin,
	}

	// propose the denom auth
	_, err = server.ProposeDenomAdmin(ctx, proposeMsg)

	// CHECKS
	// ------------------------------------------------------------

	// check if the error is correct
	require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)

	// check if the error message is correct
	require.Equal(
		t,
		fmt.Sprintf(
			"Auth: incorrect admin for denom (%s) only the current bank admin (%s) can perform this action (attempted with: %s): unauthorized",
			fullDenom,
			creator,
			metaAdmin,
		),
		err.Error(),
	)

	// Ensure that the denom auth is still the same
	denomAuth, isFound = k.GetDenomAuth(ctx, fullDenom)
	require.True(t, isFound)
	require.Equal(t, creator, denomAuth.BankAdmin)
	require.Equal(t, metaAdmin, denomAuth.MetadataAdmin)

	// Ensure that the proposed DenomAuth is still empty as it has not been set up yet ever
	_, isFound = k.GetProposedDenomAuth(ctx, fullDenom)
	require.False(t, isFound)

	// EVENT CHECKS
	// ------------------------------------------------------------

	// check that the event was not emitted
	events := ctx.EventManager().Events()

	// Ensure that the event EventDenomAuthClaimed has not occurred
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventDenomAuthProposed {
			foundEvent = true
		}
	}
	require.False(t, foundEvent, "Expected EventDenomAuthProposed event NOT to be emitted")
}

func TestMsgServer_ProposeDenomAdmin_DisabledDenom(t *testing.T) {
	// Test case: propose a new denom bank admin and metadata admin
	// when the denom auth is disabled

	// SETUP - CREATE DENOM
	// ------------------------------

	// create denom first
	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// create a sample signer address
	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create a sample sub denom name
	// the full name will be in format "coin.{creator}.{subDenom}"
	subDenom := "abc"
	maxSupply := cosmosmath.NewUint(1000000)
	uri := "ipfs://uri"
	uriHash := "sha256hash"
	fee := sdk.NewCoin("uzig", cosmosmath.NewInt(1000))

	fullDenom := "coin." + creator + "." + subDenom

	denomMetaData := banktypes.Metadata{
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

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// code will first check if subDenom is matching one of the code tokens (e.g. "uzig")
	bankKeeper.
		EXPECT().
		HasSupply(gomock.Any(), subDenom).
		Return(false).
		Times(1)

	// code will check if the signer has the required balances to pay the fee
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, fee).
		Return(true).
		Times(1)

	// code will deduct the fee from the signer's account
	// SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error
	// it will return nil if the operation is successful - no error
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(fee)).
		Return(nil).
		Times(1)

	// check if denom already has metadata
	// GetDenomMetaData(context.Context, string) (banktypes.Metadata, bool)
	bankKeeper.
		EXPECT().
		GetDenomMetaData(gomock.Any(), fullDenom).
		Return(banktypes.Metadata{}, false).
		Times(1)

	// SetDenomMetaData(context.Context, banktypes.Metadata)
	bankKeeper.
		EXPECT().
		SetDenomMetaData(gomock.Any(), denomMetaData).
		Return().
		Times(1)

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, nil)

	// create a message server
	server := keeper.NewMsgServerImpl(k)

	// create a message to create a new denom
	msg := &types.MsgCreateDenom{
		Creator:             creator,
		SubDenom:            subDenom,
		MintingCap:          maxSupply,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}

	// create a new denom
	resp, err := server.CreateDenom(ctx, msg)

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct,
	// and we will set the expected response,
	// so we can compare it
	expectedResp := &types.MsgCreateDenomResponse{
		BankAdmin:           creator,
		MetadataAdmin:       creator,
		Denom:               "coin." + creator + "." + subDenom,
		MintingCap:          maxSupply,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}
	// make sure that response is not nil
	require.NotNil(t, resp)
	// compare the expected response with the actual response
	require.Equal(t, expectedResp, resp)

	// DISABLE THE DENOM AUTH
	// ------------------------------------------------------------
	// create a message to disable the denom auth
	disableMsg := &types.MsgDisableDenomAdmin{
		Denom:  fullDenom,
		Signer: creator,
	}

	// disable the denom auth
	disableResp, err := server.DisableDenomAdmin(ctx, disableMsg)

	// check if the response is correct
	require.NoError(t, err)
	// we will also need to check if the response is correct,
	expectedDisableResp := &types.MsgDisableDenomAdminResponse{
		Denom: fullDenom,
	}

	// make sure that response is not nil
	require.NotNil(t, disableResp)
	// compare the expected response with the actual response
	require.Equal(t, expectedDisableResp, disableResp)

	// PROPOSE NEW DENOM BANK ADMIN AND METADATA ADMIN
	// ------------------------------------------------------------

	newAddress := sample.AccAddress()

	// create a message to propose a new denom bank admin and metadata admin
	proposeMsg := &types.MsgProposeDenomAdmin{
		Denom:         fullDenom,
		BankAdmin:     newAddress,
		MetadataAdmin: newAddress,
		// since we are bypassing the grpc mechanism, we need to set the signer that is automatically set by the grpc
		Signer: creator,
	}

	// propose the denom auth
	_, err = server.ProposeDenomAdmin(ctx, proposeMsg)

	// CHECKS
	// ------------------------------------------------------------

	// check if the error is correct
	require.ErrorIs(t, err, types.ErrDenomLocked)

	// check if the error message is correct
	require.Equal(
		t,
		fmt.Sprintf(
			"Auth: bank admin disabled (%s): denom changes are permanently disabled",
			fullDenom,
		),
		err.Error(),
	)

	// Ensure that the denom auth is still the same
	denomAuth, isFound := k.GetDenomAuth(ctx, fullDenom)
	require.True(t, isFound)
	require.Equal(t, "", denomAuth.BankAdmin)
	require.Equal(t, creator, denomAuth.MetadataAdmin)

	// Ensure that the proposed DenomAuth is still empty as it has not been set up yet ever
	_, isFound = k.GetProposedDenomAuth(ctx, fullDenom)
	require.False(t, isFound)

	// EVENT CHECKS
	// ------------------------------------------------------------

	// check that the event was not emitted
	events := ctx.EventManager().Events()

	// Ensure that the event EventDenomAuthClaimed has not occurred
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventDenomAuthProposed {
			foundEvent = true
		}
	}
	require.False(t, foundEvent, "Expected EventDenomAuthProposed event NOT to be emitted")
}

func TestMsgServer_ProposeDenomAdmin_InvalidBankAdminAddress(t *testing.T) {
	// Test case: try to propose denom auth with invalid bank admin address

	// SETUP - CREATE DENOM
	// ------------------------------

	// create denom first
	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// create a sample signer address
	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create a sample sub denom name
	// the full name will be in format "coin.{creator}.{subDenom}"
	subDenom := "abc"
	mintingCap := cosmosmath.NewUint(1000000)
	uri := "ipfs://uri"
	uriHash := "sha256hash"
	fee := sdk.NewCoin("uzig", cosmosmath.NewInt(1000))

	fullDenom := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + subDenom

	denomMetaData := banktypes.Metadata{
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

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// code will first check if subDenom is matching one of the code tokens (e.g. "uzig")
	bankKeeper.
		EXPECT().
		HasSupply(gomock.Any(), subDenom).
		Return(false).
		Times(1)

	// code will check if the signer has the required balances to pay the fee
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, fee).
		Return(true).
		Times(1)

	// code will deduct the fee from the signer's account
	// SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error
	// it will return nil if the operation is successful - no error
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(fee)).
		Return(nil).
		Times(1)

	// check if denom already has metadata
	// GetDenomMetaData(context.Context, string) (banktypes.Metadata, bool)
	bankKeeper.
		EXPECT().
		GetDenomMetaData(gomock.Any(), fullDenom).
		Return(banktypes.Metadata{}, false).
		Times(1)

	// SetDenomMetaData(context.Context, banktypes.Metadata)
	bankKeeper.
		EXPECT().
		SetDenomMetaData(gomock.Any(), denomMetaData).
		Return().
		Times(1)

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, nil)

	// create a message server
	server := keeper.NewMsgServerImpl(k)

	// create a message to create a new denom
	msg := &types.MsgCreateDenom{
		Creator:             creator,
		SubDenom:            subDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}

	// create a new denom
	resp, err := server.CreateDenom(ctx, msg)

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct,
	// and we will set the expected response,
	// so we can compare it
	expectedResp := &types.MsgCreateDenomResponse{
		BankAdmin:           creator,
		MetadataAdmin:       creator,
		Denom:               "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + subDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}
	// make sure that response is not nil
	require.NotNil(t, resp)
	// compare the expected response with the actual response
	require.Equal(t, expectedResp, resp)

	// PROPOSE NEW DENOM BANK ADMIN AND METADATA ADMIN
	// ------------------------------------------------------------

	invalidBankAdmin := "invalid_address"
	newMetaAdmin := sample.AccAddress()

	// create a message to propose a new denom bank admin and metadata admin
	proposeMsg := &types.MsgProposeDenomAdmin{
		Denom:         fullDenom,
		BankAdmin:     invalidBankAdmin,
		MetadataAdmin: newMetaAdmin,
		// since we are bypassing the grpc mechanism, we need to set the signer that is automatically set by the grpc
		Signer: creator,
	}

	// propose the denom auth
	_, err = server.ProposeDenomAdmin(ctx, proposeMsg)

	// CHECKS
	// ------------------------------------------------------------

	// check if the error is correct
	require.Error(t, err)

	// check if the error message is correct
	require.Equal(
		t,
		fmt.Sprintf("invalid new admin address: %s", invalidBankAdmin),
		err.Error(),
	)

	// Ensure that in the token the admin is still the creator
	denomAuth, isFound := k.GetDenomAuth(ctx, fullDenom)
	require.True(t, isFound)
	require.Equal(t, creator, denomAuth.BankAdmin)
	require.Equal(t, creator, denomAuth.MetadataAdmin)

	// Ensure that the proposed DenomAuth is not stored
	_, isFound = k.GetProposedDenomAuth(ctx, fullDenom)
	require.False(t, isFound)

	// EVENT CHECKS
	// ------------------------------------------------------------

	// check that the event was not emitted
	events := ctx.EventManager().Events()

	// Ensure that the event EventDenomAuthProposed has not occurred
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventDenomAuthProposed {
			foundEvent = true
		}
	}
	require.False(t, foundEvent, "Expected EventDenomAuthProposed event NOT to be emitted")
}
