package types_test

import (
	"fmt"
	"testing"

	cosmosmath "cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"zigchain/testutil/sample"
	"zigchain/x/factory/types"

	"github.com/stretchr/testify/require"
)

// Positive test cases

func TestGenesisState_Validate_Positive(t *testing.T) {
	// Test case: unit test function for validating a GenesisState

	creator := sample.AccAddress()
	fullDenomAbc := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + "abc"
	fullDenomUsdt := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + "usdt"

	// define a slice of test cases (table approach)
	// each test case consists of a description, a GenesisState, and whether it's expected to be valid
	tests := []struct {
		// desc is a human-readable description or label for the test case
		// helps to understand what each test case is checking or validating
		desc string
		// pointer to an instance of the types.GenesisState struct
		// it represents the specific state configuration that the test case is checking
		genState *types.GenesisState
		// indicates whether the genState in the test case is expected to be valid or not
		valid bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc: "valid genesis state",
			// create a new instance of the types.GenesisState
			// & symbol is used to take the address of the newly created struct, creating a pointer to the struct
			genState: &types.GenesisState{
				// initializing the DenomList field of the GenesisState struct
				DenomList: []types.Denom{
					{Creator: creator, Denom: fullDenomAbc, MintingCap: cosmosmath.NewUint(1000), Minted: cosmosmath.NewUint(0)},
					{Creator: creator, Denom: fullDenomUsdt, MintingCap: cosmosmath.NewUint(1000), Minted: cosmosmath.NewUint(0)},
				},
				DenomAuthList: []types.DenomAuth{
					{Denom: fullDenomAbc, BankAdmin: creator, MetadataAdmin: creator},
					{Denom: fullDenomUsdt, BankAdmin: creator, MetadataAdmin: creator},
				},
				Params: types.Params{
					CreateFeeDenom:  "uzig",
					CreateFeeAmount: 1000,
				},
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
		{
			desc: "valid genesis state with DenomList and DenomAuthList equal to Nil",
			// create a new instance of the types.GenesisState
			// & symbol is used to take the address of the newly created struct, creating a pointer to the struct
			genState: &types.GenesisState{
				// initializing the DenomList field of the GenesisState struct
				DenomList:     nil,
				DenomAuthList: nil,
				Params: types.Params{
					CreateFeeDenom:  "uzig",
					CreateFeeAmount: 1000,
				},
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
		{
			desc: "valid genesis state with empty DenomList and DenomAuthList",
			// create a new instance of the types.GenesisState
			// & symbol is used to take the address of the newly created struct, creating a pointer to the struct
			genState: &types.GenesisState{
				// initializing the DenomList field of the GenesisState struct
				DenomList:     []types.Denom{},
				DenomAuthList: []types.DenomAuth{},
				Params: types.Params{
					CreateFeeDenom:  "uzig",
					CreateFeeAmount: 1000,
				},
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
		{
			desc: "valid genesis state with missing DenomList and DenomAuthList",
			// create a new instance of the types.GenesisState
			// & symbol is used to take the address of the newly created struct, creating a pointer to the struct
			genState: &types.GenesisState{
				// initializing the DenomList field of the GenesisState struct
				Params: types.Params{
					CreateFeeDenom:  "uzig",
					CreateFeeAmount: 1000,
				},
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
	}
	// iterate over each test case
	for _, tc := range tests {
		// run a subtest with the description as the test name
		t.Run(tc.desc, func(t *testing.T) {
			// call the Validate method on the GenesisState
			err := tc.genState.Validate()
			// assert that no error occurred
			require.NoError(t, err)
		})
	}
}

func TestGenesisState_Validate_Empty(t *testing.T) {
	// Test case: check the Validate method of an empty GenesisState

	// create an empty GenesisState using the DefaultGenesis function
	emptyGenesis := types.DefaultGenesis()

	// modify the DenomList field to be nil for this specific test case
	emptyGenesis.DenomList = nil

	// call the Validate method on the GenesisState
	err := emptyGenesis.Validate()

	// assertion: the Validate method should not return an error for an empty GenesisState
	require.NoError(t, err)
}

func TestGenesisState_Validate_ValidDenomAuth(t *testing.T) {
	// Test case: check the Validate method of a GenesisState with a valid DenomAuth

	creator := sample.AccAddress()
	fullDenomAbc := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + "abc"

	// create a valid DenomAuth
	validDenomAuth := types.DenomAuth{
		Denom:         fullDenomAbc,
		BankAdmin:     creator,
		MetadataAdmin: creator,
	}

	// call the Validate method on the DenomAuth
	err := validDenomAuth.Validate()

	// assertion: the Validate method should not return an error for a valid DenomAuth
	require.NoError(t, err, "Valid DenomAuth validation should not return an error")
}

func TestDenom_Validate_Valid(t *testing.T) {
	// Test case: check the Validate method of a valid Denom

	creator := sample.AccAddress()
	fullDenomAbc := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + "abc"

	// create a valid Denom
	validDenom := types.Denom{
		Creator:    creator,
		Denom:      fullDenomAbc,
		Minted:     cosmosmath.NewUint(100),
		MintingCap: cosmosmath.NewUint(1000),
	}

	// call the Validate method on the Denom
	err := validDenom.Validate()

	// assertion: the Validate method should not return an error for a valid Denom
	require.NoError(t, err, "Valid Denom validation should not return an error")
}

func TestDefaultParamsValue(t *testing.T) {
	// Test case: check whether the DefaultParams value is correctly set in the DefaultGenesis function
	// it compares the Params field of a GenesisState created by DefaultGenesis with the expected DefaultParams value

	// create a new GenesisState using the DefaultGenesis function
	genState := types.DefaultGenesis()

	// assertion: the Params field of the GenesisState should be equal to the DefaultParams value
	require.Equal(t, types.DefaultParams(), genState.Params, "DefaultParams should be set in DefaultGenesis")
}

func TestDefaultParams_Validate(t *testing.T) {
	// Test case: check the Validate method of a valid Params

	// create a valid Params
	params := types.DefaultParams()

	// call the Validate method on the Params
	err := params.Validate()

	// assertion: the Validate method should not return an error for a valid Params
	require.NoError(t, err, "Valid Params validation should not return an error")
}

func TestGenesisState_Validate_NilDenomList(t *testing.T) {
	// Test case: check the Validate method of a GenesisState with a Nil DenomList

	genState := &types.GenesisState{
		DenomList:     nil,
		DenomAuthList: nil,
		Params:        types.DefaultParams(),
	}

	err := genState.Validate()
	require.NoError(t, err, "GenesisState with nil DenomList should not return an error")
}

// Negative test cases

func TestGenesisState_Validate_MissingDenomAuth(t *testing.T) {
	// Test case: check the Validate method of a GenesisState with a non-empty DenomList
	// it checks if the validation of a GenesisState with a non-empty DenomList
	// and default Params returns no error

	creator := sample.AccAddress()
	fullDenomAbc := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + "abc"
	fullDenomUsdt := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + "usdt"

	// create a valid GenesisState with a non-empty DenomList and default Params
	validGenesis := &types.GenesisState{
		DenomList: []types.Denom{
			{Creator: creator, Denom: fullDenomAbc, MintingCap: cosmosmath.NewUint(1000), Minted: cosmosmath.NewUint(0)},
			{Creator: creator, Denom: fullDenomUsdt, MintingCap: cosmosmath.NewUint(1000), Minted: cosmosmath.NewUint(0)},
		},
		Params: types.DefaultParams(),
	}

	// call the Validate method on the GenesisState
	err := validGenesis.Validate()

	// assertion: the Validate method should not return an error for a valid GenesisState
	require.Error(t, err)
	require.EqualError(t, err, fmt.Sprintf("missing DenomAuth for denom '%s'", fullDenomAbc))
}

func TestGenesisState_Validate_DuplicatedDenom(t *testing.T) {
	// Test case: try to validate a GenesisState with duplicated Denom

	creator := sample.AccAddress()
	fullDenomAbc := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + "abc"
	fullDenomUsdt := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + "usdt"

	// create a GenesisState with a duplicated Denom
	duplicatedDenom := &types.GenesisState{
		DenomList: []types.Denom{
			{Creator: creator, Denom: fullDenomAbc, MintingCap: cosmosmath.NewUint(1000), Minted: cosmosmath.NewUint(0)},
			{Creator: creator, Denom: fullDenomAbc, MintingCap: cosmosmath.NewUint(1000), Minted: cosmosmath.NewUint(0)},
			{Creator: creator, Denom: fullDenomAbc, MintingCap: cosmosmath.NewUint(1000), Minted: cosmosmath.NewUint(0)},
		},
		DenomAuthList: []types.DenomAuth{
			{Denom: fullDenomAbc, BankAdmin: creator, MetadataAdmin: creator},
			{Denom: fullDenomUsdt, BankAdmin: creator, MetadataAdmin: creator},
		},
		Params: types.DefaultParams(),
	}

	// call the Validate method on the GenesisState
	err := duplicatedDenom.Validate()

	// assertion: the Validate method should return an error for a GenesisState with duplicated Denom
	require.Error(t, err)

	// assert that the error message matches the expected error message
	require.EqualError(t, err, "duplicated index for denom")
}

func TestGenesisState_Validate_DuplicatedDenomAuth(t *testing.T) {
	// Test case: try to validate a GenesisState with duplicated DenomAuth

	creator := sample.AccAddress()
	fullDenomAbc := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + "abc"
	fullDenomUsdt := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + "usdt"

	// create a GenesisState with a duplicated DenomAuth
	duplicatedDenomAuth := &types.GenesisState{
		DenomList: []types.Denom{
			{Creator: creator, Denom: fullDenomAbc, MintingCap: cosmosmath.NewUint(1000), Minted: cosmosmath.NewUint(0)},
			{Creator: creator, Denom: fullDenomUsdt, MintingCap: cosmosmath.NewUint(1000), Minted: cosmosmath.NewUint(0)},
		},
		DenomAuthList: []types.DenomAuth{
			{Denom: fullDenomAbc, BankAdmin: creator, MetadataAdmin: creator},
			{Denom: fullDenomAbc, BankAdmin: creator, MetadataAdmin: creator},
		},
		Params: types.DefaultParams(),
	}

	// call the Validate method on the GenesisState
	err := duplicatedDenomAuth.Validate()

	// assertion: the Validate method should return an error for a GenesisState with duplicated DenomAuth
	require.Error(t, err)

	// assert that the error message matches the expected error message
	require.EqualError(t, err, "duplicated index for denomAuth")
}

func TestGenesisState_Validate_InvalidCreatorInDenomList(t *testing.T) {
	// Test case: try to validate a GenesisState with an invalid Denom creator

	creator := sample.AccAddress()
	invalidCreator := "invalid_address"
	fullDenom := "coin" + types.FactoryDenomDelimiterChar + invalidCreator + types.FactoryDenomDelimiterChar + "abc"

	invalidCreatorGenesis := &types.GenesisState{
		DenomList: []types.Denom{
			{Creator: invalidCreator, Denom: fullDenom, MintingCap: cosmosmath.NewUint(1000), Minted: cosmosmath.NewUint(0)},
		},
		DenomAuthList: []types.DenomAuth{
			{Denom: fullDenom, BankAdmin: creator, MetadataAdmin: creator},
		},
		Params: types.DefaultParams(),
	}

	err := invalidCreatorGenesis.Validate()
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidDenom)
	require.Contains(t, err.Error(), "Denom creator invalid")
}

func TestGenesisState_Validate_InvalidDenomInDenomList(t *testing.T) {
	// Test case: invalid Denom in DenomList (e.g., Denom fails validation)

	creator := sample.AccAddress()

	invalidDenomGenesis := &types.GenesisState{
		DenomList: []types.Denom{
			{Creator: creator, Denom: "invalid_denom"},
		},
		Params: types.DefaultParams(),
	}

	err := invalidDenomGenesis.Validate()
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidDenom)
	require.EqualError(t, err, "(invalid_denom) : invalid denom: invalid_denom: Factory Denom name is not valid")
}

func TestGenesisState_Validate_EmptyDenomList(t *testing.T) {
	// Test case: check the Validate method of a GenesisState with an empty DenomAuthList

	creator := sample.AccAddress()
	fullDenomAbc := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + "abc"
	fullDenomUsdt := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + "usdt"

	genState := &types.GenesisState{
		DenomList: []types.Denom{},
		DenomAuthList: []types.DenomAuth{
			{Denom: fullDenomAbc, BankAdmin: creator, MetadataAdmin: creator},
			{Denom: fullDenomUsdt, BankAdmin: creator, MetadataAdmin: creator},
		},
		Params: types.DefaultParams(),
	}

	err := genState.Validate()
	require.Error(t, err)
	require.EqualError(t, err, fmt.Sprintf("missing Denom for DenomAuth '%s'", fullDenomAbc))
}

func TestGenesisState_Validate_MissingDenomList(t *testing.T) {
	// Test case: check the Validate method of a GenesisState with a missing DenomList and non-empty DenomAuthList
	// it checks if the validation of a GenesisState with a missing DenomList and non-empty DenomAuthList
	// and default Params returns error for missing DenomList

	creator := sample.AccAddress()
	fullDenomAbc := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + "abc"
	fullDenomUsdt := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + "usdt"

	validGenesis := &types.GenesisState{
		DenomAuthList: []types.DenomAuth{
			{Denom: fullDenomAbc, BankAdmin: creator, MetadataAdmin: creator},
			{Denom: fullDenomUsdt, BankAdmin: creator, MetadataAdmin: creator},
		},
		Params: types.DefaultParams(),
	}

	// call the Validate method on the GenesisState
	err := validGenesis.Validate()

	// assertion: the Validate method should return an error as there is no DenomList for the denoms
	require.Error(t, err)
	require.EqualError(t, err, fmt.Sprintf("missing Denom for DenomAuth '%s'", fullDenomAbc))
}

func TestGenesisState_Validate_MissingOneDenomInDenomList(t *testing.T) {
	// Test case: check the Validate method of a GenesisState with a missing Denom in DenomList
	// but existing in DenomAuthList

	creator1 := sample.AccAddress()
	creator2 := sample.AccAddress()
	fullDenomAbc := "coin" + types.FactoryDenomDelimiterChar + creator1 + types.FactoryDenomDelimiterChar + "abc"
	fullDenomUsdt := "coin" + types.FactoryDenomDelimiterChar + creator2 + types.FactoryDenomDelimiterChar + "usdt"

	validGenesis := &types.GenesisState{
		DenomList: []types.Denom{
			{Denom: fullDenomAbc, Creator: creator1, MintingCap: cosmosmath.NewUint(1000), Minted: cosmosmath.NewUint(0)},
		},
		DenomAuthList: []types.DenomAuth{
			{Denom: fullDenomAbc, BankAdmin: creator1, MetadataAdmin: creator1},
			{Denom: fullDenomUsdt, BankAdmin: creator2, MetadataAdmin: creator2},
		},
		Params: types.DefaultParams(),
	}

	// call the Validate method on the GenesisState
	err := validGenesis.Validate()

	// assertion: the Validate method should return an error as there is no DenomList for the denoms
	require.Error(t, err)
	require.EqualError(t, err, fmt.Sprintf("missing Denom for DenomAuth '%s'", fullDenomUsdt))
}

func TestGenesisState_Validate_NilDenomAuthList(t *testing.T) {
	// Test case: check the Validate method of a GenesisState with a Nil DenomAuthList

	creator := sample.AccAddress()
	fullDenom := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + "abc"

	genState := &types.GenesisState{
		DenomList:     []types.Denom{{Creator: creator, Denom: fullDenom, MintingCap: cosmosmath.NewUint(1000), Minted: cosmosmath.NewUint(0)}},
		DenomAuthList: nil,
		Params:        types.DefaultParams(),
	}

	err := genState.Validate()
	require.Error(t, err)
	require.EqualError(t, err, fmt.Sprintf("missing DenomAuth for denom '%s'", fullDenom))
}

func TestGenesisState_Validate_EmptyDenomAuthList(t *testing.T) {
	// Test case: check the Validate method of a GenesisState with an empty DenomAuthList

	creator := sample.AccAddress()
	fullDenom := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + "abc"

	genState := &types.GenesisState{
		DenomList:     []types.Denom{{Creator: creator, Denom: fullDenom, MintingCap: cosmosmath.NewUint(1000), Minted: cosmosmath.NewUint(0)}},
		DenomAuthList: []types.DenomAuth{},
		Params:        types.DefaultParams(),
	}

	err := genState.Validate()
	require.Error(t, err)
	require.EqualError(t, err, fmt.Sprintf("missing DenomAuth for denom '%s'", fullDenom))
}

func TestGenesisState_Validate_MissingDenomAuthList(t *testing.T) {
	// Test case: check the Validate method of a GenesisState with a non-empty DenomList and missing DenomAuthList

	creator := sample.AccAddress()
	fullDenomAbc := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + "abc"
	fullDenomUsdt := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + "usdt"

	validGenesis := &types.GenesisState{
		DenomList: []types.Denom{
			{Creator: creator, Denom: fullDenomAbc, MintingCap: cosmosmath.NewUint(1000), Minted: cosmosmath.NewUint(0)},
			{Creator: creator, Denom: fullDenomUsdt, MintingCap: cosmosmath.NewUint(1000), Minted: cosmosmath.NewUint(0)},
		},
		Params: types.DefaultParams(),
	}

	// call the Validate method on the GenesisState
	err := validGenesis.Validate()

	// assertion: the Validate method should return an error as there is no DenomAuth for the denoms
	require.Error(t, err)
	require.EqualError(t, err, fmt.Sprintf("missing DenomAuth for denom '%s'", fullDenomAbc))
}

func TestGenesisState_Validate_MissingOneDenomAuthInDenomAuthList(t *testing.T) {
	// Test case: error as missing DenomAuth in DenomAuth but existing in DenomAuthList

	creator1 := sample.AccAddress()
	creator2 := sample.AccAddress()
	fullDenomAbc := "coin" + types.FactoryDenomDelimiterChar + creator1 + types.FactoryDenomDelimiterChar + "abc"
	fullDenomUsdt := "coin" + types.FactoryDenomDelimiterChar + creator2 + types.FactoryDenomDelimiterChar + "usdt"

	validGenesis := &types.GenesisState{
		DenomList: []types.Denom{
			{Denom: fullDenomAbc, Creator: creator1, MintingCap: cosmosmath.NewUint(1000), Minted: cosmosmath.NewUint(0)},
			{Denom: fullDenomUsdt, Creator: creator2, MintingCap: cosmosmath.NewUint(1000), Minted: cosmosmath.NewUint(0)},
		},
		DenomAuthList: []types.DenomAuth{
			{Denom: fullDenomAbc, BankAdmin: creator1, MetadataAdmin: creator1},
		},
		Params: types.DefaultParams(),
	}

	// call the Validate method on the GenesisState
	err := validGenesis.Validate()

	// assertion: the Validate method should return an error as there is no DenomList for the denoms
	require.Error(t, err)
	require.EqualError(t, err, fmt.Sprintf("missing DenomAuth for denom '%s'", fullDenomUsdt))
}

func TestGenesisState_Validate_InvalidBankAdminInDenomAuthList(t *testing.T) {
	// Test case: invalid BankAdmin address in DenomAuthList

	creator := sample.AccAddress()
	invalidAdmin := "invalid_address"
	fullDenom := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + "abc"

	invalidBankAdminGenesis := &types.GenesisState{
		DenomAuthList: []types.DenomAuth{
			{Denom: fullDenom, BankAdmin: invalidAdmin, MetadataAdmin: creator},
		},
		Params: types.DefaultParams(),
	}

	err := invalidBankAdminGenesis.Validate()
	require.Error(t, err)
	require.ErrorIs(t, err, sdkerrors.ErrInvalidAddress)
	require.Contains(t, err.Error(), "DenomAuth BankAdmin invalid")
}

func TestGenesisState_Validate_InvalidMetadataAdminInDenomAuthList(t *testing.T) {
	// Test case: invalid MetadataAdmin address in DenomAuthList

	creator := sample.AccAddress()
	invalidAdmin := "invalid_address"
	fullDenom := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + "abc"

	invalidMetadataAdminGenesis := &types.GenesisState{
		DenomAuthList: []types.DenomAuth{
			{Denom: fullDenom, BankAdmin: creator, MetadataAdmin: invalidAdmin},
		},
		Params: types.DefaultParams(),
	}

	err := invalidMetadataAdminGenesis.Validate()
	require.Error(t, err)
	require.ErrorIs(t, err, sdkerrors.ErrInvalidAddress)
	require.Contains(t, err.Error(), "DenomAuth MetadataAdmin invalid")
}

func TestDenom_Validate_EmptyCreator(t *testing.T) {
	// Test case: check the Validate method of a Denom with an empty creator

	fullDenom := "coin" + types.FactoryDenomDelimiterChar + types.FactoryDenomDelimiterChar + "abc"

	invalidDenom := types.Denom{
		Creator:    "",
		Denom:      fullDenom,
		Minted:     cosmosmath.NewUint(0),
		MintingCap: cosmosmath.NewUint(1000),
	}

	err := invalidDenom.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "Denom creator invalid")
}

func TestDenom_Validate_EmptyDenom(t *testing.T) {
	// Test case: check the Validate method of a Denom with an empty denom

	creator := sample.AccAddress()

	invalidDenom := types.Denom{
		Creator:    creator,
		Denom:      "",
		Minted:     cosmosmath.NewUint(0),
		MintingCap: cosmosmath.NewUint(1000),
	}

	err := invalidDenom.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid denom")
}

func TestDenom_Validate_MintedExceedsMintingCap(t *testing.T) {
	// Test case: minted amount exceeds minting cap

	creator := sample.AccAddress()
	fullDenom := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + "abc"

	invalidDenom := types.Denom{
		Creator:    creator,
		Denom:      fullDenom,
		Minted:     cosmosmath.NewUint(1001),
		MintingCap: cosmosmath.NewUint(1000),
	}

	err := invalidDenom.Validate()
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidDenom)
	require.EqualError(
		t,
		err,
		"Denom minted (1001) exceeds minting cap (1000): Factory Denom name is not valid",
	)
}

func TestDenom_Validate_ZeroMintingCap(t *testing.T) {
	// Test case: minting cap is zero

	creator := sample.AccAddress()
	fullDenom := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + "abc"

	invalidDenom := types.Denom{
		Creator:    creator,
		Denom:      fullDenom,
		Minted:     cosmosmath.NewUint(100),
		MintingCap: cosmosmath.NewUint(0),
	}

	err := invalidDenom.Validate()
	require.Error(t, err)
	require.EqualError(
		t,
		err,
		"Minting Cap 0 must be greater than 0: Factory Denom name is not valid",
	)
}

func TestDenom_Validate_CreatorMismatch(t *testing.T) {
	// Test case: denom's creator does not match extracted creator

	creator := sample.AccAddress()
	creatorMismatch := sample.AccAddress()
	mismatchedDenom := "coin" + types.FactoryDenomDelimiterChar + creatorMismatch + types.FactoryDenomDelimiterChar + "abc"

	invalidDenom := types.Denom{
		Creator:    creator,
		Denom:      mismatchedDenom,
		Minted:     cosmosmath.NewUint(0),
		MintingCap: cosmosmath.NewUint(1000),
	}

	err := invalidDenom.Validate()
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidDenom)
	require.EqualError(
		t,
		err,
		fmt.Sprintf(
			"Denom creator (%s) does not match denom creator (%s): Factory Denom name is not valid",
			creatorMismatch,
			creator,
		),
	)
}

func TestDenomAuth_Validate_EmptyBankAdmin(t *testing.T) {
	// Test case: check the Validate method of a DenomAuth with an empty BankAdmin

	creator := sample.AccAddress()
	fullDenom := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + "abc"

	invalidDenomAuth := types.DenomAuth{
		Denom:         fullDenom,
		BankAdmin:     "",
		MetadataAdmin: creator,
	}

	err := invalidDenomAuth.Validate()
	require.NoError(t, err, "DenomAuth BankAdmin disabled therefore should be valid")
}

func TestDenomAuth_Validate_EmptyMetadataAdmin(t *testing.T) {
	// Test case: check the Validate method of a DenomAuth with an empty MetadataAdmin

	creator := sample.AccAddress()
	fullDenom := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + "abc"

	invalidDenomAuth := types.DenomAuth{
		Denom:         fullDenom,
		BankAdmin:     creator,
		MetadataAdmin: "",
	}

	err := invalidDenomAuth.Validate()
	require.NoError(t, err, "DenomAuth MetadataAdmin disabled therefore should be valid")
}

func TestDenomAuth_Validate_InvalidDenom(t *testing.T) {
	// Test case: invalid Denom in DenomAuth

	invalidDenomAuth := types.DenomAuth{
		Denom:         "invalid_denom",
		BankAdmin:     sample.AccAddress(),
		MetadataAdmin: sample.AccAddress(),
	}

	err := invalidDenomAuth.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid denom")
}

func TestDenom_Validate_NilMintingCap(t *testing.T) {
	// Test case: check the Validate method of a Denom with a nil MintingCap

	creator := sample.AccAddress()
	fullDenom := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + "abc"

	invalidDenom := types.Denom{
		Creator:    creator,
		Denom:      fullDenom,
		Minted:     cosmosmath.NewUint(0),
		MintingCap: cosmosmath.Uint{},
	}

	err := invalidDenom.Validate()
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidDenom)
	require.EqualError(
		t,
		err,
		"Minting Cap <nil> is nil: Factory Denom name is not valid",
	)
}
