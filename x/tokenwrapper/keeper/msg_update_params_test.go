package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"zigchain/app"
	"zigchain/x/tokenwrapper/keeper"
	"zigchain/x/tokenwrapper/types"
)

// Positive tests cases

func TestMsgUpdateParams_Valid(t *testing.T) {
	// Test case: valid authority updates params successfully

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// instantiate a new msgServer
	ms := keeper.NewMsgServerImpl(k)
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// set valid authority
	authority := testApp.TokenwrapperKeeper.GetAuthority()
	// set the authority in the context
	newParams := types.DefaultParams()

	// create a message to update params
	msg := &types.MsgUpdateParams{
		Authority: authority,
		Params:    newParams,
	}

	// call the UpdateParams method
	resp, err := ms.UpdateParams(ctx, msg)
	// check if the response is not nil and no error occurred
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, resp, &types.MsgUpdateParamsResponse{
		Authority: authority,
		Params:    newParams,
	})

	// optionally: confirm the params are updated
	currentParams := k.GetParams(ctx)
	require.Equal(t, newParams, currentParams)
}

func TestMsgUpdateParams_Positive(t *testing.T) {
	// Test case: update params successfully

	k, ms, ctx := setupMsgServer(t)
	params := types.DefaultParams()
	require.NoError(t, k.SetParams(ctx, params))
	wctx := sdk.UnwrapSDKContext(ctx)

	// create a message to update params
	msg := &types.MsgUpdateParams{
		Authority: k.GetAuthority(),
		Params:    params,
	}

	// call the UpdateParams method
	_, err := ms.UpdateParams(wctx, msg)
	// check that no error occurred
	require.NoError(t, err)
}

func TestMsgUpdateParams_EmptyParams(t *testing.T) {
	// Test case: update params with empty params

	k, ms, ctx := setupMsgServer(t)
	require.NoError(t, k.SetParams(ctx, types.DefaultParams()))
	wctx := sdk.UnwrapSDKContext(ctx)

	// create a message to update params with empty params
	msg := &types.MsgUpdateParams{
		Authority: k.GetAuthority(),
		Params:    types.Params{},
	}

	// call the UpdateParams method
	_, err := ms.UpdateParams(wctx, msg)
	// check that no error occurred
	require.NoError(t, err)
}

// Negative test cases

func TestMsgUpdateParams_InvalidAuthority(t *testing.T) {
	// Test case: invalid authority updates params fails

	k, ms, ctx := setupMsgServer(t)
	require.NoError(t, k.SetParams(ctx, types.DefaultParams()))
	wctx := sdk.UnwrapSDKContext(ctx)

	// create a message to update params with invalid authority
	msg := &types.MsgUpdateParams{
		Authority: "invalid",
		Params:    types.DefaultParams(),
	}

	// call the UpdateParams method
	_, err := ms.UpdateParams(wctx, msg)
	// check that an error occurred
	require.Error(t, err)
	// check that the error message contains "invalid authority"
	require.Contains(t, err.Error(), "invalid authority")
}
