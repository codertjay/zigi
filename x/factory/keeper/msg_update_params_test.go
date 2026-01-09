package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	keepertest "zigchain/testutil/keeper"
	"zigchain/x/factory/keeper"
	"zigchain/x/factory/testutil"
	"zigchain/x/factory/types"
)

// Positive test cases

func TestMsgUpdateParams_Positive(t *testing.T) {
	// Test case: update params with valid authority and params

	// create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish() // ensure all expectations are met

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// create mock account keeper
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, accountKeeper)
	ms := keeper.NewMsgServerImpl(k)

	params := types.DefaultParams()

	require.NoError(t, k.SetParams(ctx, params))

	// default params
	testCases := []struct {
		name      string
		input     *types.MsgUpdateParams
		expErr    bool
		expErrMsg string
	}{
		{
			name: "send enabled param",
			input: &types.MsgUpdateParams{
				Authority: k.GetAuthority(),
				Params:    types.Params{},
			},
			expErr: false,
		},
		{
			name: "all good",
			input: &types.MsgUpdateParams{
				Authority: k.GetAuthority(),
				Params:    params,
			},
			expErr: false,
		},
		{
			name: "custom create fee denom",
			input: &types.MsgUpdateParams{
				Authority: k.GetAuthority(),
				Params: types.Params{
					CreateFeeDenom: "coin.zig1wz7n45yh4cptr27yf7g59pfhc28z6jcyax85ng.pandacebbdcc",
				},
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			k.SetDenom(ctx, types.Denom{
				Denom: tc.input.Params.CreateFeeDenom,
			})

			_, err := ms.UpdateParams(ctx, tc.input)
			require.NoError(t, err)
		})
	}
}

// Negative test cases

func TestMsgUpdateParams_InvalidAuthority(t *testing.T) {
	// Test case: try to update params with invalid authority

	k, ms, ctx := setupMsgServer(t)
	params := types.DefaultParams()
	require.NoError(t, k.SetParams(ctx, params))

	// invalid authority
	testCases := []struct {
		name      string
		authority string
		expErrMsg string
	}{
		{
			name:      "completely invalid authority",
			authority: "invalid",
			expErrMsg: "invalid authority",
		},
		{
			name:      "empty authority",
			authority: "",
			expErrMsg: "invalid authority",
		},
		{
			name:      "wrong zigchain address format",
			authority: "zig1wrongaddress",
			expErrMsg: "invalid authority",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ms.UpdateParams(ctx, &types.MsgUpdateParams{
				Authority: tc.authority,
				Params:    params,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expErrMsg)
		})
	}
}

func TestMsgUpdateParams_InvalidCreateFeeDenom(t *testing.T) {
	// Test case: try to update params with non-existent create fee denom

	// create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish() // ensure all expectations are met

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// create mock account keeper
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, accountKeeper)
	ms := keeper.NewMsgServerImpl(k)

	params := types.DefaultParams()

	require.NoError(t, k.SetParams(ctx, params))

	testCases := []struct {
		name           string
		createFeeDenom string
		hasSupply      bool
		expErrMsg      string
	}{
		{
			name:           "non-existent denom",
			createFeeDenom: "coin.zig1wz7n45yh4cptr27yf7g59pfhc28z6jcyax85ng.pandacebbdcc",
			hasSupply:      false,
			expErrMsg:      "does not exist in the factory module and is not a native denom with supply",
		},
		{
			name:           "empty denom",
			createFeeDenom: "",
			hasSupply:      false,
			expErrMsg:      "does not exist in the factory module and is not a native denom with supply",
		},
		{
			name:           "invalid denom format",
			createFeeDenom: "invalid-denom-123",
			hasSupply:      false,
			expErrMsg:      "does not exist in the factory module and is not a native denom with supply",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bankKeeper.EXPECT().HasSupply(gomock.Any(), tc.createFeeDenom).Return(tc.hasSupply)

			_, err := ms.UpdateParams(ctx, &types.MsgUpdateParams{
				Authority: k.GetAuthority(),
				Params: types.Params{
					CreateFeeDenom: tc.createFeeDenom,
				},
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expErrMsg)
		})
	}
}

func TestMsgUpdateParams_NativeDenomWithSupply(t *testing.T) {
	// Test case: update params with native denom that has supply

	// create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish() // ensure all expectations are met

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// create mock account keeper
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, accountKeeper)
	ms := keeper.NewMsgServerImpl(k)

	params := types.DefaultParams()

	require.NoError(t, k.SetParams(ctx, params))

	testCases := []struct {
		name           string
		createFeeDenom string
		hasSupply      bool
		expErr         bool
	}{
		{
			name:           "native denom with supply",
			createFeeDenom: "uzig",
			hasSupply:      true,
			expErr:         false,
		},
		{
			name:           "another native denom with supply",
			createFeeDenom: "uatom",
			hasSupply:      true,
			expErr:         false,
		},
		{
			name:           "native denom without supply",
			createFeeDenom: "uzig",
			hasSupply:      false,
			expErr:         true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Expect HasSupply to be called when the denom is not found in factory module
			bankKeeper.EXPECT().HasSupply(gomock.Any(), tc.createFeeDenom).Return(tc.hasSupply)

			_, err := ms.UpdateParams(ctx, &types.MsgUpdateParams{
				Authority: k.GetAuthority(),
				Params: types.Params{
					CreateFeeDenom: tc.createFeeDenom,
				},
			})

			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), "does not exist in the factory module and is not a native denom with supply")
			} else {
				require.NoError(t, err)
			}
		})
	}
}
