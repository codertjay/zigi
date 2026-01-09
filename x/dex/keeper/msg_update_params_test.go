package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"zigchain/x/dex/types"
)

// Positive test cases

func TestMsgUpdateParams_Valid(t *testing.T) {
	// Test case for updating parameters with valid values --> default params

	k, ms, ctx := setupMsgServer(t)
	params := types.DefaultParams()
	require.NoError(t, k.SetParams(ctx, params))
	wctx := sdk.UnwrapSDKContext(ctx)

	msg := &types.MsgUpdateParams{
		Authority: k.GetAuthority(),
		Params:    params,
	}

	_, err := ms.UpdateParams(wctx, msg)
	require.NoError(t, err)

	// verify state was updated
	stored := k.GetParams(ctx)
	require.Equal(t, params, stored)
}

func TestMsgUpdateParams_EmptyParamsWithValidAuthority(t *testing.T) {
	// Test case for updating parameters with empty params and valid authority

	k, ms, ctx := setupMsgServer(t)
	require.NoError(t, k.SetParams(ctx, types.DefaultParams()))
	wctx := sdk.UnwrapSDKContext(ctx)

	msg := &types.MsgUpdateParams{
		Authority: k.GetAuthority(),
		Params:    types.Params{}, // sending empty params
	}

	_, err := ms.UpdateParams(wctx, msg)
	require.NoError(t, err)
}

func TestMsgUpdateParams_ValidUpdate(t *testing.T) {
	// Test case for updating parameters with valid values

	k, ms, ctx := setupMsgServer(t)
	wctx := sdk.UnwrapSDKContext(ctx)

	customParams := types.Params{
		NewPoolFeePct: 600,
		CreationFee:   200_000_000,
	}

	msg := types.MsgUpdateParams{
		Authority: k.GetAuthority(),
		Params:    customParams,
	}

	_, err := ms.UpdateParams(wctx, &msg)

	require.NoError(t, err)

	// verify params were updated
	stored := k.GetParams(ctx)
	require.Equal(t, customParams, stored)
}

// Negative test cases

func TestMsgUpdateParams_InvalidAuthority(t *testing.T) {
	// Test case for updating parameters with invalid authority

	k, ms, ctx := setupMsgServer(t)
	params := types.DefaultParams()
	require.NoError(t, k.SetParams(ctx, params))
	wctx := sdk.UnwrapSDKContext(ctx)

	msg := &types.MsgUpdateParams{
		Authority: "invalid",
		Params:    params,
	}

	_, err := ms.UpdateParams(wctx, msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid authority")
}
