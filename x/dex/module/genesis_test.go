package dex_test

import (
	"testing"

	keepertest "zigchain/testutil/keeper"
	"zigchain/testutil/nullify"
	dex "zigchain/x/dex/module"
	"zigchain/x/dex/types"

	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		PoolList: []types.Pool{
			{
				PoolId: "0",
			},
			{
				PoolId: "1",
			},
		},
		PoolsMeta: &types.PoolsMeta{
			NextPoolId: 2,
		},
		PoolUidsList: []types.PoolUids{
			{
				PoolUid: "0",
			},
			{
				PoolUid: "1",
			},
		},
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	dex.InitGenesis(ctx, k, genesisState)
	got := dex.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.ElementsMatch(t, genesisState.PoolList, got.PoolList)
	require.Equal(t, genesisState.PoolsMeta, got.PoolsMeta)
	require.ElementsMatch(t, genesisState.PoolUidsList, got.PoolUidsList)
	// this line is used by starport scaffolding # genesis/test/assert
}
