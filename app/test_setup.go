package app

import (
	"encoding/json"
	"testing"
	"time"

	"zigchain/zutils/constants"

	tokenwrappertypes "zigchain/x/tokenwrapper/types"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	cmttypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	// sec-verified - used for test only
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	// sec-verified - used for test only
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

func InitTestApp(initChain bool, t *testing.T) *App {
	app := InitiateNewApp(t)
	if initChain {
		genesisState, valSet, _, _ := GenesisStateWithValSet(app)
		stateBytes, err := json.MarshalIndent(genesisState, "", " ")
		if err != nil {
			panic(err)
		}

		_, err = app.InitChain(&abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: simtestutil.DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		})
		if err != nil {
			panic(err)
		}

		// commit genesis changes
		_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height:             app.LastBlockHeight() + 1,
			Hash:               app.LastCommitID().Hash,
			NextValidatorsHash: valSet.Hash(),
		})
		if err != nil {
			panic(err)
		}
		_, err = app.BeginBlocker(app.BaseApp.NewContext(initChain))
		if err != nil {
			panic(err)
		}
	}
	return app
}

func InitiateNewApp(t *testing.T) *App {
	db := dbm.NewMemDB()
	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = t.TempDir()
	appOptions[server.FlagInvCheckPeriod] = simcli.FlagPeriodValue

	app, err := New(
		log.NewNopLogger(),
		db,
		nil,
		true,
		appOptions,
	)
	require.NoError(t, err)

	return app
}

func GenesisStateWithValSet(app *App) (GenesisState, *cmttypes.ValidatorSet, sdk.AccAddress, sdk.ValAddress) {
	privVal := mock.NewPV()
	pubKey, _ := privVal.GetPubKey()
	validator := cmttypes.NewValidator(pubKey, 1)
	valSet := cmttypes.NewValidatorSet([]*cmttypes.Validator{validator})

	// Generate genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	senderPrivKey.PubKey().Address()
	acc := authtypes.NewBaseAccountWithAddress(senderPrivKey.PubKey().Address().Bytes())
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(constants.BondDenom, sdkmath.NewInt(100000000000000))),
	}

	balances := []banktypes.Balance{balance}
	genesisState := NewDefaultGenesisState(app, app.AppCodec())
	genAccs := []authtypes.GenesisAccount{acc}
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(authGenesis)

	// Add default genesis state for tokenwrapper module
	tokenwrapperGenesis := tokenwrappertypes.DefaultGenesis()
	genesisState[tokenwrappertypes.ModuleName] = app.AppCodec().MustMarshalJSON(tokenwrapperGenesis)

	validators := make([]stakingtypes.Validator, 0, len(valSet.Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(valSet.Validators))

	bondAmt := sdk.DefaultPowerReduction
	initValPowers := []abci.ValidatorUpdate{}

	for _, val := range valSet.Validators {
		pk, _ := cryptocodec.FromCmtPubKeyInterface(val.PubKey)
		pkAny, _ := codectypes.NewAnyWithValue(pk)
		validator := stakingtypes.Validator{
			OperatorAddress:   sdk.ValAddress(val.Address).String(),
			ConsensusPubkey:   pkAny,
			Jailed:            false,
			Status:            stakingtypes.Bonded,
			Tokens:            bondAmt,
			DelegatorShares:   sdkmath.LegacyOneDec(),
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(sdkmath.LegacyNewDecWithPrec(5, 2), sdkmath.LegacyNewDecWithPrec(10, 2), sdkmath.LegacyNewDecWithPrec(10, 2)),
			MinSelfDelegation: sdkmath.OneInt(),
		}
		validators = append(validators, validator)
		delegations = append(delegations, stakingtypes.NewDelegation(genAccs[0].GetAddress().String(), sdk.ValAddress(val.Address).String(), sdkmath.LegacyOneDec()))

		// Add initial validator powers so consumer InitGenesis runs correctly
		pub, _ := val.ToProto()
		initValPowers = append(initValPowers, abci.ValidatorUpdate{
			Power:  val.VotingPower,
			PubKey: pub.PubKey,
		})
	}

	// Set validators and delegations
	params := stakingtypes.DefaultParams()
	params.BondDenom = constants.BondDenom

	// Set staking module genesis state
	stakingGenesis := stakingtypes.NewGenesisState(params, validators, delegations)
	genesisState[stakingtypes.ModuleName] = app.AppCodec().MustMarshalJSON(stakingGenesis)

	// Initialize mint module genesis state
	mintGenesis := minttypes.DefaultGenesisState()
	mintGenesis.Minter.Inflation = sdkmath.LegacyNewDecWithPrec(13, 2) // 13% starting inflation
	mintGenesis.Params.MintDenom = constants.BondDenom
	genesisState[minttypes.ModuleName] = app.AppCodec().MustMarshalJSON(mintGenesis)

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		// Add genesis acc tokens to total supply
		totalSupply = totalSupply.Add(b.Coins...)
	}

	for range delegations {
		// Add delegated tokens to total supply
		totalSupply = totalSupply.Add(sdk.NewCoin(constants.BondDenom, bondAmt))
	}

	// Add bonded amount to bonded pool module account
	balances = append(balances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   sdk.Coins{sdk.NewCoin(constants.BondDenom, bondAmt)},
	})

	// Update total supply
	bankGenesis := banktypes.NewGenesisState(banktypes.DefaultGenesisState().Params, balances, totalSupply, []banktypes.Metadata{}, []banktypes.SendEnabled{})
	genesisState[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)

	_, err := cmttypes.PB2TM.ValidatorUpdates(initValPowers)
	if err != nil {
		panic("failed to get vals")
	}

	return genesisState, valSet, genAccs[0].GetAddress(), sdk.ValAddress(validator.Address)
}
