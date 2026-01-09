package dex

import (
	"cosmossdk.io/math"
	// nosem: math-random-used
	"math/rand" // checked: we use math/rand to generate random numbers predictably
	// so we can reproduce the same results in tests
	ztests "zigchain/zutils/tests"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"zigchain/testutil/sample"
	dexsimulation "zigchain/x/dex/simulation"
	"zigchain/x/dex/types"
	factorytypes "zigchain/x/factory/types"
)

// avoid unused import issue
var (
	_ = dexsimulation.FindAccount
	_ = rand.Rand{}
	_ = sample.AccAddress
	_ = sdk.AccAddress{}
	_ = simulation.MsgEntryKind
)

// #nosec G101 -- These are simulation operation weight constants, not credentials
// False positive, this is not a password or a secret
const (
	opWeightMsgCreatePool          = "op_weight_msg_pool"
	defaultWeightMsgCreatePool int = 10

	opWeightMsgSwap          = "op_weight_msg_swap"
	defaultWeightMsgSwap int = 100

	opWeightMsgAddLiquidity          = "op_weight_msg_add_liquidity"
	defaultWeightMsgAddLiquidity int = 30

	opWeightMsgRemoveLiquidity          = "op_weight_msg_remove_liquidity"
	defaultWeightMsgRemoveLiquidity int = 20

	// this line is used by starport scaffolding # simapp/module/const
)

// GenerateGenesisState creates a randomized GenState of the module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {

	// extract addresses from the simulation state accounts
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}

	// get length of addresses, so we can randomly pick from them
	addressLen := len(accs)

	// Pick random addresses from simState.Accounts
	address1 := accs[simState.Rand.Intn(addressLen)]
	address2 := accs[simState.Rand.Intn(addressLen)]
	address3 := accs[simState.Rand.Intn(addressLen)]
	address4 := accs[simState.Rand.Intn(addressLen)]

	// generate factory format denoms
	factoryDenom1 := "coin" + factorytypes.FactoryDenomDelimiterChar + address1 + factorytypes.FactoryDenomDelimiterChar + ztests.RandomSubDenom(simState.Rand, 42)
	factoryDenom2 := "coin" + factorytypes.FactoryDenomDelimiterChar + address2 + factorytypes.FactoryDenomDelimiterChar + ztests.RandomSubDenom(simState.Rand, 42)
	factoryDenom3 := "coin" + factorytypes.FactoryDenomDelimiterChar + address3 + factorytypes.FactoryDenomDelimiterChar + ztests.RandomSubDenom(simState.Rand, 42)
	factoryDenom4 := "coin" + factorytypes.FactoryDenomDelimiterChar + address4 + factorytypes.FactoryDenomDelimiterChar + ztests.RandomSubDenom(simState.Rand, 42)

	// generate random pool denoms and amounts
	pool1Denom1 := sdk.NewCoin(factoryDenom1, math.NewInt(simState.Rand.Int63n(1_000_000_000)))
	pool1Denom2 := sdk.NewCoin(factoryDenom2, math.NewInt(simState.Rand.Int63n(1_000_000_000)))
	// create pool denoms, we use NewCoins to ensure sorting
	pool1Denoms := sdk.NewCoins(pool1Denom1, pool1Denom2)

	pool2Denom1 := sdk.NewCoin(factoryDenom2, math.NewInt(simState.Rand.Int63n(1_000_000_000)))
	pool2Denom2 := sdk.NewCoin(factoryDenom3, math.NewInt(simState.Rand.Int63n(1_000_000_000)))
	pool2Denoms := sdk.NewCoins(pool2Denom1, pool2Denom2)

	pool3Denom1 := sdk.NewCoin(factoryDenom3, math.NewInt(simState.Rand.Int63n(1_000_000_000)))
	pool3Denom2 := sdk.NewCoin(factoryDenom4, math.NewInt(simState.Rand.Int63n(1_000_000_000)))
	pool3Denoms := sdk.NewCoins(pool3Denom1, pool3Denom2)

	pool4Denom1 := sdk.NewCoin(factoryDenom1, math.NewInt(simState.Rand.Int63n(1_000_000_000)))
	pool4Denom2 := sdk.NewCoin("uzig", math.NewInt(simState.Rand.Int63n(1_000_000_000)))
	pool4Denoms := sdk.NewCoins(pool4Denom1, pool4Denom2)

	// Create one pool with Huge numbers - max int64 (positive)
	pool5Denom1 := sdk.NewCoin(factoryDenom2, math.NewInt(simState.Rand.Int63()))
	pool5Denom2 := sdk.NewCoin("uzig", math.NewInt(simState.Rand.Int63()))
	pool5Denoms := sdk.NewCoins(pool5Denom1, pool5Denom2)

	// We want to randomize the genesis state as much as possible.
	dexGenesis := types.GenesisState{
		Params: types.Params{
			// Random number between 0 and 50%, because the scaling factor is 100_000
			// cast to uint32 is safe, because the random number is always positive
			// #nosec G115
			NewPoolFeePct: uint32(simState.Rand.Int31n(50_000)),
			// CreationFee is a random number between 0 and 1_000_000 of the native token
			// cast to uint32 is safe, because the random number is always positive
			// #nosec G115
			CreationFee: uint32(simState.Rand.Int31n(1_000_000)),
			// Beneficiary
			Beneficiary: "",
		},
		PoolList: []types.Pool{
			{
				PoolId: "zp1",
				LpToken: sdk.Coin{
					Denom:  "zp1",
					Amount: pool1Denom1.Amount.Add(pool1Denom2.Amount),
				},
				Creator: address1,
				// Random number between 0 and 50%, fee represents percentage where 100_000 is 100%
				// int32 cast to uint32 is safe,
				// because the random number is always positive
				// #nosec G115
				Fee:     uint32(simState.Rand.Int31n(50_000)),
				Formula: "constant_product",
				Coins:   pool1Denoms,
			},
			{
				PoolId: "zp2",
				LpToken: sdk.Coin{
					Denom:  "zp1",
					Amount: pool2Denom1.Amount.Add(pool2Denom2.Amount),
				},
				Creator: address2,
				// Random number between 0 and 50%
				// #nosec G115
				Fee:     uint32(simState.Rand.Int31n(50_000)),
				Formula: "constant_product",
				Coins:   pool2Denoms,
			},
			{
				PoolId: "zp3",
				LpToken: sdk.Coin{
					Denom:  "zp1",
					Amount: pool3Denom1.Amount.Add(pool3Denom2.Amount),
				},
				Creator: address3,
				// Random number between 0 and 50%
				// #nosec G115
				Fee:     uint32(simState.Rand.Int31n(50_000)),
				Formula: "constant_product",
				Coins:   pool3Denoms,
			},
			{
				PoolId: "zp4",
				LpToken: sdk.Coin{
					Denom:  "zp1",
					Amount: pool4Denom1.Amount.Add(pool4Denom2.Amount),
				},
				Creator: address4,
				// Random number between 0 and 50%
				// #nosec G115
				Fee:     uint32(simState.Rand.Int31n(50_000)),
				Formula: "constant_product",
				Coins:   pool4Denoms,
			},
			{
				PoolId: "zp5",
				LpToken: sdk.Coin{
					Denom:  "zp1",
					Amount: pool5Denom1.Amount.Add(pool5Denom2.Amount),
				},
				Creator: address1,
				// Random number between 0 and 50%
				// int32 cast to uint32 is safe,
				// because the random number is always positive
				// #nosec G115
				Fee:     uint32(simState.Rand.Int31n(50_000)),
				Formula: "constant_product",
				Coins:   pool5Denoms,
			},
		},
		PoolsMeta: &types.PoolsMeta{
			// this is the next pool id, pools counter -1
			NextPoolId: 6,
		},
		PoolUidsList: []types.PoolUids{
			{
				// Unique UID is created from the pool denoms, since they are always sorted this is deterministic,
				// and prevent duplicates when creating pools and user flips the order of the coins
				PoolUid: pool1Denoms[0].Denom + "/" + pool1Denoms[1].Denom,
				PoolId:  "zp1",
			},
			{
				PoolUid: pool2Denoms[0].Denom + "/" + pool2Denoms[1].Denom,
				PoolId:  "zp2",
			},
			{
				PoolUid: pool3Denoms[0].Denom + "/" + pool3Denoms[1].Denom,
				PoolId:  "zp3",
			},
			{
				PoolUid: pool4Denoms[0].Denom + "/" + pool4Denoms[1].Denom,
				PoolId:  "zp4",
			},
			{
				PoolUid: pool5Denoms[0].Denom + "/" + pool5Denoms[1].Denom,
				PoolId:  "zp5",
			},
		},
		// this line is used by starport scaffolding # simapp/module/genesisState
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&dexGenesis)
}

// RegisterStoreDecoder registers a decoder.
func (am AppModule) RegisterStoreDecoder(_ simtypes.StoreDecoderRegistry) {}

// WeightedOperations returns all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)

	var weightMsgCreatePool int
	simState.AppParams.GetOrGenerate(opWeightMsgCreatePool, &weightMsgCreatePool, nil,
		func(_ *rand.Rand) {
			weightMsgCreatePool = defaultWeightMsgCreatePool
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCreatePool,
		dexsimulation.SimulateMsgCreatePool(simState.TxConfig, am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgSwap int
	simState.AppParams.GetOrGenerate(opWeightMsgSwap, &weightMsgSwap, nil,
		func(_ *rand.Rand) {
			weightMsgSwap = defaultWeightMsgSwap
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgSwap,
		dexsimulation.SimulateMsgSwap(simState.TxConfig, am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgAddLiquidity int
	simState.AppParams.GetOrGenerate(opWeightMsgAddLiquidity, &weightMsgAddLiquidity, nil,
		func(_ *rand.Rand) {
			weightMsgAddLiquidity = defaultWeightMsgAddLiquidity
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgAddLiquidity,
		dexsimulation.SimulateMsgAddLiquidity(simState.TxConfig, am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgRemoveLiquidity int
	simState.AppParams.GetOrGenerate(opWeightMsgRemoveLiquidity, &weightMsgRemoveLiquidity, nil,
		func(_ *rand.Rand) {
			weightMsgRemoveLiquidity = defaultWeightMsgRemoveLiquidity
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgRemoveLiquidity,
		dexsimulation.SimulateMsgRemoveLiquidity(simState.TxConfig, am.accountKeeper, am.bankKeeper, am.keeper),
	))
	// this line is used by starport scaffolding # simapp/module/operation

	return operations
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (am AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{
		simulation.NewWeightedProposalMsg(
			opWeightMsgCreatePool,
			defaultWeightMsgCreatePool,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				dexsimulation.SimulateMsgCreatePool(simState.TxConfig, am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgSwap,
			defaultWeightMsgSwap,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				dexsimulation.SimulateMsgSwap(simState.TxConfig, am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgAddLiquidity,
			defaultWeightMsgAddLiquidity,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				dexsimulation.SimulateMsgAddLiquidity(simState.TxConfig, am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgRemoveLiquidity,
			defaultWeightMsgRemoveLiquidity,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				dexsimulation.SimulateMsgRemoveLiquidity(simState.TxConfig, am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		// this line is used by starport scaffolding # simapp/module/OpMsg
	}
}
