package factory_test

import (
	"testing"

	keepertest "zigchain/testutil/keeper"
	"zigchain/testutil/nullify"
	factory "zigchain/x/factory/module"
	"zigchain/x/factory/types"

	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		DenomList: []types.Denom{
			{
				Denom: "0",
			},
			{
				Denom: "1",
			},
		},
		DenomAuthList: []types.DenomAuth{
			{
				Denom: "0",
			},
			{
				Denom: "1",
			},
		},
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.FactoryKeeper(t, nil, nil)
	factory.InitGenesis(ctx, k, genesisState)
	got := factory.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.ElementsMatch(t, genesisState.DenomList, got.DenomList)
	require.ElementsMatch(t, genesisState.DenomAuthList, got.DenomAuthList)
	// this line is used by starport scaffolding # genesis/test/assert
}
