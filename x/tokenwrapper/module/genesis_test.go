package tokenwrapper_test

import (
	"testing"

	keepertest "zigchain/testutil/keeper"
	"zigchain/testutil/nullify"
	tokenwrapper "zigchain/x/tokenwrapper/module"
	"zigchain/x/tokenwrapper/types"

	sdkmath "cosmossdk.io/math"

	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params:              types.DefaultParams(),
		TotalTransferredIn:  sdkmath.NewInt(1000),
		TotalTransferredOut: sdkmath.NewInt(2000),
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.TokenwrapperKeeper(t, nil)
	tokenwrapper.InitGenesis(ctx, k, genesisState)
	got := tokenwrapper.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
