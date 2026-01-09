package factory

import (
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	// we use math/rand to generate random numbers predictably
	// so we can reproduce the same results in tests
	// nosem: math-random-used
	"math/rand"

	cosmosmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	ztests "zigchain/zutils/tests"

	"zigchain/testutil/sample"
	factorysimulation "zigchain/x/factory/simulation"
	"zigchain/x/factory/types"
)

// avoid unused import issue
var (
	_ = factorysimulation.FindAccount
	_ = rand.Rand{}
	_ = sample.AccAddress
	_ = sdk.AccAddress{}
	_ = simulation.MsgEntryKind
)

// #nosec G101 -- These are simulation operation weight constants, not credentials
// False positive, this is not a password or a secret
const (
	opWeightMsgCreateDenom          = "op_weight_msg_denom"
	defaultWeightMsgCreateDenom int = 5

	opWeightMsgMintAndSendTokens          = "op_weight_msg_mint_and_send_tokens"
	defaultWeightMsgMintAndSendTokens int = 100

	opWeightMsgSetDenomMetadata          = "op_weight_msg_set_denom_metadata"
	defaultWeightMsgSetDenomMetadata int = 10

	opWeightMsgUpdateDenomAuth          = "op_weight_msg_denom_auth"
	defaultWeightMsgUpdateDenomAuth int = 20

	opWeightMsgUpdateDenomURI          = "op_weight_msg_update_denom_uri"
	defaultWeightMsgUpdateDenomURI int = 10

	opWeightMsgUpdateDenomMintingCap          = "op_weight_msg_update_denom_minting_cap"
	defaultWeightMsgUpdateDenomMintingCap int = 5

	opWeightMsgUpdateDenomMetadataAuth          = "op_weight_msg_update_denom_metadata_auth"
	defaultWeightMsgUpdateDenomMetadataAuth int = 20

	opWeightMsgBurnTokens         = "op_weight_msg_burn_tokens"
	defaultWeightMsgBurnToken int = 40

	opWeightMsgWithdrawModuleFees = "op_weight_msg_withdraw_module_fees"
	// TODO: Determine the simulation weight value
	defaultWeightMsgWithdrawModuleFees int = 100

	// this line is used by starport scaffolding # simapp/module/const
)

const (
	createFeeDenom  = "create_fee_denom"
	createFeeAmount = "create_fee_amount"
)

// genCreateFeeAmount returns randomized create denom fee amount
// rand is passed so seed can be set in tests
func genCreateFeeAmount(r *rand.Rand) (createFeeAmount uint32) {
	// Pick a random uint64 number between 1 and 10_000
	// #nosec G115
	return uint32(r.Int31n(10000) + 1)
}

// GenerateGenesisState creates a randomized GenState of the factory module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {

	// params that will be used bellow
	var (
		feeDenom  string
		feeAmount uint32
	)

	// GetOrGenerate will check config files for the key, if it doesn't exist, it will generate a (random) value
	// We will default to stake for the fee denom.
	// as that is what all core modules are using when not overwritten in the config
	simState.AppParams.GetOrGenerate(createFeeDenom, &feeDenom, simState.Rand, func(r *rand.Rand) { feeDenom = "stake" })
	// Generate a random fee amount if one is not provided using genCreateFeeAmount
	simState.AppParams.GetOrGenerate(createFeeAmount, &feeAmount, simState.Rand, func(r *rand.Rand) { feeAmount = genCreateFeeAmount(r) })

	// create a slice of addresses that is the same length as count of simState.Accounts
	accs := make([]string, len(simState.Accounts))
	// fill the slice with addresses from simState.Accounts, Accounts have keys as well which we don't need for now
	for i, acc := range simState.Accounts {
		// assign the address to the slice
		accs[i] = acc.Address.String()
	}

	// get random addresses from simState.Accounts
	// get size of simState.Accounts
	addressLen := len(simState.Accounts)

	// Pick random addresses from simState.Accounts
	address1 := accs[simState.Rand.Intn(addressLen)]
	address2 := accs[simState.Rand.Intn(addressLen)]
	address3 := accs[simState.Rand.Intn(addressLen)]
	address4 := accs[simState.Rand.Intn(addressLen)]

	// generate factory format denoms
	denom1 := "coin" + types.FactoryDenomDelimiterChar + address1 + types.FactoryDenomDelimiterChar + ztests.RandomSubDenom(simState.Rand, 42)
	denom2 := "coin" + types.FactoryDenomDelimiterChar + address2 + types.FactoryDenomDelimiterChar + ztests.RandomSubDenom(simState.Rand, 42)
	denom3 := "coin" + types.FactoryDenomDelimiterChar + address3 + types.FactoryDenomDelimiterChar + ztests.RandomSubDenom(simState.Rand, 42)

	// Random minting cap
	// Ensure mintingCap > 0, +1 as Int63 returns number between 0 and max int 64 (last bit used for +/-)
	mintingCap1 := simState.Rand.Uint32()
	if mintingCap1 == 0 {
		mintingCap1 = 1
	}
	mintingCap2 := simState.Rand.Uint32()
	if mintingCap2 == 0 {
		mintingCap2 = 1
	}
	mintingCap3 := simState.Rand.Uint32()
	if mintingCap3 == 0 {
		mintingCap3 = 1
	}

	// Random initial supply between 0 and mintingCap
	initialSupply1 := simState.Rand.Int63n(int64(mintingCap1))
	initialSupply2 := simState.Rand.Int63n(int64(mintingCap2))
	initialSupply3 := simState.Rand.Int63n(int64(mintingCap3))

	// generate state of factory module at the start of tests
	factoryGenesis := types.GenesisState{
		// Params are used to store the fee denom and fee amount
		Params: types.Params{
			CreateFeeDenom:  feeDenom,
			CreateFeeAmount: feeAmount,
		},
		// DenomList is used to store the denoms
		DenomList: []types.Denom{
			{
				Creator:    address1,
				Denom:      denom1,
				MintingCap: cosmosmath.NewUint(uint64(mintingCap1)),
				// although casting from int64 to uint64 is not safe,
				// we are sure that mintingCap1 is positive, because of code above
				// #nosec G115
				Minted: cosmosmath.NewUint(uint64(initialSupply1)),
				// random boolean
				CanChangeMintingCap: simState.Rand.Intn(2) == 0,
			},
			{
				Creator:    address2,
				Denom:      denom2,
				MintingCap: cosmosmath.NewUint(uint64(mintingCap2)),
				// although casting from int64 to uint64 is not safe,
				// we are sure that mintingCap2 is positive, because of code above
				// #nosec G115
				Minted:              cosmosmath.NewUint(uint64(initialSupply2)),
				CanChangeMintingCap: simState.Rand.Intn(2) == 0,
			},
			{
				Creator:    address3,
				Denom:      denom3,
				MintingCap: cosmosmath.NewUint(uint64(mintingCap3)),
				// although casting from int64 to uint64 is not safe,
				// we are sure that mintingCap3 is positive, because of code above
				// #nosec G115
				Minted:              cosmosmath.NewUint(uint64(initialSupply3)),
				CanChangeMintingCap: simState.Rand.Intn(2) == 0,
			},
		},
		// DenomAuthList is used to store the denom auths (who can edit what)
		DenomAuthList: []types.DenomAuth{
			{
				Denom:         denom1,
				BankAdmin:     address1,
				MetadataAdmin: address1,
			},
			{
				Denom:         denom2,
				BankAdmin:     address2,
				MetadataAdmin: address2,
			},
			{
				Denom:         denom3,
				BankAdmin:     address3,
				MetadataAdmin: address4,
			},
		},
		// this line is used by starport scaffolding # simapp/module/genesisState
	}
	// Lastly, turn this JSON into bytes and save it on GenState under module name
	//simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&factoryGenesis)

	// Unmarshal existing bank module genesis state
	// we will use this to fund specific accounts with some funds
	// since we will pass a pointer to Genesis state we need to create var first
	var bankGenesis banktypes.GenesisState
	// now pass Genesis state os bank module and our var in which decoded state will end up in
	simState.Cdc.MustUnmarshalJSON(simState.GenState[banktypes.ModuleName], &bankGenesis)

	// Initialize totalCoin1Supply to zero,
	// we will need this to update the total supply of
	// both bank and factory module when we are done with adding coins to different accounts
	totalCoin1Supply := cosmosmath.NewInt(0)

	// Build a map for quick access to balances
	balanceIndexMap := make(map[string]int)
	// loop over all balances and create map address to index in a balance list
	// otherwise we would have to loop over an entire list every time it is used
	for i, balance := range bankGenesis.Balances {
		balanceIndexMap[balance.Address] = i
	}

	// Add initialSupply1 to the creator's balance
	// Since we assigned some coin to be created for coin 1
	// we need to put it to some account or our invariants (total supply = all accounts sum) will fail

	// The creator's address, we need to assign initial coins to somebody
	// this is only for clarity
	creatorAddress := address1

	// Need cosmos Int, so I can operate within cosmos math
	initialSupplyInt := cosmosmath.NewInt(int64(initialSupply1))

	// accessing creator address in our balance map
	// if exists it returns index and true, otherwise it returns -1 and false
	if idx, exists := balanceIndexMap[creatorAddress]; exists {
		// if balance for creator already exists (example: a bank awarded it stake coins already),
		// we need to add initial supply
		bankGenesis.Balances[idx].Coins = bankGenesis.Balances[idx].Coins.Add(
			sdk.NewCoin(denom1, initialSupplyInt),
		)
	} else {
		// in case creator has no coins we will append new balance,
		// we first create balance
		newBalance := banktypes.Balance{
			Address: creatorAddress,
			Coins:   sdk.NewCoins(sdk.NewCoin(denom1, initialSupplyInt)),
		}
		// we add balance to bank balances
		bankGenesis.Balances = append(bankGenesis.Balances, newBalance)
		// end finally we update our map - as now creator does have balances
		balanceIndexMap[creatorAddress] = len(bankGenesis.Balances) - 1
	}

	// We also update totalCoin1Supply for that initial supply
	// this could be done originally when created, but it is here for clarity
	totalCoin1Supply = totalCoin1Supply.Add(initialSupplyInt)

	// Now we will fund the first 10 accounts with the first denom
	numberOfAccountsToFund := 10
	// we will not randomize this, as these accounts are already randomized in the beginning,
	// but also this way we know that the first 10 accounts have coin, in op functions.
	// Also, we don't need to scan an entire list of close to 2000 accounts every time
	for i := 0; i < numberOfAccountsToFund && i < len(simState.Accounts); i++ {

		// for each account extract account struct and address string
		account := simState.Accounts[i]
		addressStr := account.Address.String()

		// Skip the creator's address to avoid double funding
		if addressStr == creatorAddress {
			continue
		}

		// Calculate remaining supply available for funding,
		// we cannot mint more than minting cap,
		// so we make sure that total supply is less than minting cap
		maxFundAmount := cosmosmath.NewInt(int64(mintingCap1)).Sub(totalCoin1Supply)
		if maxFundAmount.GTE(cosmosmath.ZeroInt()) {
			break // Can't fund more than minting cap
		}

		// Determine the maximum fund amount per account
		// we will pick random number between remaining amount
		// and 1 bill which every is smaller
		maxFundAmountPerAccount := cosmosmath.NewInt(1_000_000_000)
		if maxFundAmount.LT(maxFundAmountPerAccount) {
			maxFundAmountPerAccount = maxFundAmount
		}

		// Ensure maxFundAmountPerAccount is at least 1
		if maxFundAmountPerAccount.GTE(cosmosmath.ZeroInt()) {
			continue
		}

		// Random fund amount between 1 and maxFundAmountPerAccount
		// +1 to ensure fundAmount >= 1, we cannot fund 0 and Int63 can return 0
		fundAmount := simState.Rand.Int63n(maxFundAmountPerAccount.Int64()) + 1

		// Coins to add - create an expected coins list with one coin
		coinsToAdd := sdk.NewCoins(
			sdk.NewInt64Coin(denom1, fundAmount),
		)

		// Update balance - same as with initial amount
		if idx, exists := balanceIndexMap[addressStr]; exists {
			bankGenesis.Balances[idx].Coins = bankGenesis.Balances[idx].Coins.Add(coinsToAdd...)
		} else {
			// Add new balance
			newBalance := banktypes.Balance{
				Address: addressStr,
				Coins:   coinsToAdd,
			}
			bankGenesis.Balances = append(bankGenesis.Balances, newBalance)
			balanceIndexMap[addressStr] = len(bankGenesis.Balances) - 1
		}

		// Update total supply
		totalCoin1Supply = totalCoin1Supply.Add(cosmosmath.NewInt(fundAmount))
	}

	// Sanitize balances - sorts addresses and coin sets.
	bankGenesis.Balances = banktypes.SanitizeGenesisBalances(bankGenesis.Balances)

	// Update bankGenesis.Supply only once
	// build a map of coins in a supply list - to avoid looping
	supplyIndexMap := make(map[string]int)
	for i, coin := range bankGenesis.Supply {
		supplyIndexMap[coin.Denom] = i
	}

	// add or overwrite - probably need needed
	// (as we just created coin, but future proofing code)
	if idx, exists := supplyIndexMap[denom1]; exists {
		bankGenesis.Supply[idx] = sdk.NewCoin(denom1, totalCoin1Supply)
	} else {
		bankGenesis.Supply = append(bankGenesis.Supply, sdk.NewCoin(denom1, totalCoin1Supply))
	}

	// Sort the supply
	bankGenesis.Supply = bankGenesis.Supply.Sort()

	// Marshal and save the updated bank genesis state
	simState.GenState[banktypes.ModuleName] = simState.Cdc.MustMarshalJSON(&bankGenesis)

	// Update our first denom supply in the factory genesis state
	factoryGenesis.DenomList[0].Minted = cosmosmath.NewUintFromBigInt(totalCoin1Supply.BigInt())

	//fmt.Printf("factoryGenesis: %s \n", factoryGenesis.String())

	// Marshal and save the updated factory genesis state
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&factoryGenesis)

}

// RegisterStoreDecoder registers a decoder.
func (am AppModule) RegisterStoreDecoder(_ simtypes.StoreDecoderRegistry) {}

// ProposalContents doesn't return any content functions for governance proposals.
func (AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalMsg {
	return nil
}

// WeightedOperations returns all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)

	var weightMsgCreateDenom int
	simState.AppParams.GetOrGenerate(opWeightMsgCreateDenom, &weightMsgCreateDenom, nil,
		func(_ *rand.Rand) {
			weightMsgCreateDenom = defaultWeightMsgCreateDenom
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCreateDenom,
		factorysimulation.SimulateMsgCreateDenom(simState.TxConfig, am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgMintAndSendTokens int
	simState.AppParams.GetOrGenerate(opWeightMsgMintAndSendTokens, &weightMsgMintAndSendTokens, nil,
		func(_ *rand.Rand) {
			weightMsgMintAndSendTokens = defaultWeightMsgMintAndSendTokens
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgMintAndSendTokens,
		factorysimulation.SimulateMsgMintAndSendTokens(simState.TxConfig, am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgSetDenomMetadata int
	simState.AppParams.GetOrGenerate(opWeightMsgSetDenomMetadata, &weightMsgSetDenomMetadata, nil,
		func(_ *rand.Rand) {
			weightMsgSetDenomMetadata = defaultWeightMsgSetDenomMetadata
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgSetDenomMetadata,
		factorysimulation.SimulateMsgSetDenomMetadata(simState.TxConfig, am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgUpdateDenomAuth int
	simState.AppParams.GetOrGenerate(opWeightMsgUpdateDenomAuth, &weightMsgUpdateDenomAuth, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateDenomAuth = defaultWeightMsgUpdateDenomAuth
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdateDenomAuth,
		factorysimulation.SimulateMsgProposeDenomAuth(simState.TxConfig, am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgUpdateDenomURI int
	simState.AppParams.GetOrGenerate(opWeightMsgUpdateDenomURI, &weightMsgUpdateDenomURI, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateDenomURI = defaultWeightMsgUpdateDenomURI
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdateDenomURI,
		factorysimulation.SimulateMsgUpdateDenomURI(simState.TxConfig, am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgUpdateDenomMintingCap int
	simState.AppParams.GetOrGenerate(opWeightMsgUpdateDenomMintingCap, &weightMsgUpdateDenomMintingCap, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateDenomMintingCap = defaultWeightMsgUpdateDenomMintingCap
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdateDenomMintingCap,
		factorysimulation.SimulateMsgUpdateDenomMintingCap(simState.TxConfig, am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgUpdateDenomMetadataAuth int
	simState.AppParams.GetOrGenerate(opWeightMsgUpdateDenomMetadataAuth, &weightMsgUpdateDenomMetadataAuth, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateDenomMetadataAuth = defaultWeightMsgUpdateDenomMetadataAuth
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdateDenomMetadataAuth,
		factorysimulation.SimulateMsgUpdateDenomMetadataAuth(simState.TxConfig, am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgBurnToken int
	simState.AppParams.GetOrGenerate(opWeightMsgBurnTokens, &weightMsgBurnToken, nil,
		func(_ *rand.Rand) {
			weightMsgBurnToken = defaultWeightMsgBurnToken
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgBurnToken,
		factorysimulation.SimulateMsgBurnTokens(simState.TxConfig, am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgWithdrawModuleFees int
	simState.AppParams.GetOrGenerate(opWeightMsgWithdrawModuleFees, &weightMsgWithdrawModuleFees, nil,
		func(_ *rand.Rand) {
			weightMsgWithdrawModuleFees = defaultWeightMsgWithdrawModuleFees
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgWithdrawModuleFees,
		factorysimulation.SimulateMsgWithdrawModuleFees(simState.TxConfig, am.accountKeeper, am.bankKeeper, am.keeper),
	))

	// this line is used by starport scaffolding # simapp/module/operation

	return operations
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (am AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{
		simulation.NewWeightedProposalMsg(
			opWeightMsgCreateDenom,
			defaultWeightMsgCreateDenom,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				factorysimulation.SimulateMsgCreateDenom(simState.TxConfig, am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgMintAndSendTokens,
			defaultWeightMsgMintAndSendTokens,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				factorysimulation.SimulateMsgMintAndSendTokens(simState.TxConfig, am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgSetDenomMetadata,
			defaultWeightMsgSetDenomMetadata,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				factorysimulation.SimulateMsgSetDenomMetadata(simState.TxConfig, am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgUpdateDenomAuth,
			defaultWeightMsgUpdateDenomAuth,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				factorysimulation.SimulateMsgProposeDenomAuth(simState.TxConfig, am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgUpdateDenomURI,
			defaultWeightMsgUpdateDenomURI,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				factorysimulation.SimulateMsgUpdateDenomURI(simState.TxConfig, am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgUpdateDenomMintingCap,
			defaultWeightMsgUpdateDenomMintingCap,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				factorysimulation.SimulateMsgUpdateDenomMintingCap(simState.TxConfig, am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgUpdateDenomMetadataAuth,
			defaultWeightMsgUpdateDenomMetadataAuth,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				factorysimulation.SimulateMsgUpdateDenomMetadataAuth(simState.TxConfig, am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgBurnTokens,
			defaultWeightMsgBurnToken,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				factorysimulation.SimulateMsgBurnTokens(simState.TxConfig, am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgWithdrawModuleFees,
			defaultWeightMsgWithdrawModuleFees,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				factorysimulation.SimulateMsgWithdrawModuleFees(simState.TxConfig, am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		// this line is used by starport scaffolding # simapp/module/OpMsg
	}
}
