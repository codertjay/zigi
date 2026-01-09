package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	errorPacks "zigchain/testutil/data"
	types "zigchain/testutil/helpers"
	"zigchain/testutil/sample"
	"zigchain/zutils/validators"
)

var SampleMsgSetDenomMetadata = MsgSetDenomMetadata{
	Signer: sample.AccAddress(),
	Metadata: banktypes.Metadata{
		Description: "Native token of ZIGChain",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    "uzig",
				Exponent: 0,
				Aliases: []string{
					"microzig",
				},
			},
			{
				Denom:    "mzig",
				Exponent: 3,
				Aliases: []string{
					"millizig",
				},
			},
			{
				Denom:    "zig",
				Exponent: 6,
				Aliases: []string{
					"zig",
				},
			},
		},
		Base:    "uzig",
		Display: "zig",
		Name:    "Zig Zag",
		Symbol:  "ZIG",
		URI:     "https://example.com",
		URIHash: validators.SHA256HashOfURL("https://example.com"),
	},
}

// Positive test cases

func TestMsgSetDenomMetadata_NewMsgSetDenomMetadata_Positive(t *testing.T) {
	// Test case: set denom metadata with walid input data

	signer := sample.AccAddress()
	metadata := banktypes.Metadata{
		Description: "Native token of ZIGChain",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    "uzig",
				Exponent: 0,
			},
		},
		Base:    "uzig",
		Display: "zig",
		Name:    "Zig Zag",
		Symbol:  "ZIG",
		URI:     "https://zigdao.com/jsonschema/zig.json",
		URIHash: validators.SHA256HashOfURL("https://zigdao.com/jsonschema/zig.json"),
	}

	// create a new MsgSetDenomMetadata instance
	msg := NewMsgSetDenomMetadata(signer, metadata)

	// validate the fields
	require.NotNil(t, msg, "expected the message to be non-nil")
	require.Equal(t, signer, msg.Signer, "expected the signer to match the input signer")
	require.Equal(t, metadata, msg.Metadata, "expected the metadata to match the input metadata")
}

func TestMsgSetDenomMetadata_Positive(t *testing.T) {
	// Test case: validate basic properties of MsgSetDenomMetadata

	err := SampleMsgSetDenomMetadata.ValidateBasic()
	require.NoError(t, err)
}

// Negative test cases

func TestMsgSetDenomMetadata_InvalidAdminAddress(t *testing.T) {
	// Test case: set an invalid admin address and check for errors
	msg := SampleMsgSetDenomMetadata

	// set the owner address to an invalid address and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Signer,
		&errorPacks.InvalidSignerAddress,
		sdkerrors.ErrInvalidAddress,
	)
}

func TestMsgSetDenomMetadata_InvalidMetadataDescription(t *testing.T) {
	// Test case: set an invalid metadata description and check for errors

	msg := SampleMsgSetDenomMetadata

	// set the metadata description to an invalid value and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Metadata.Description,
		&errorPacks.InvalidDenomMetaDescription,
		ErrInvalidMetadata,
	)
}

func TestMsgSetDenomMetadata_InvalidDenomMetaBase(t *testing.T) {
	// Test case: set an invalid metadata base and check for errors

	msg := SampleMsgSetDenomMetadata

	// set the metadata base to an invalid value and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Metadata.Base,
		&errorPacks.InvalidDenomMetaBase,
		ErrInvalidMetadata,
	)
}

func TestMsgSetDenomMetadata_InvalidDenomMetaDisplay(t *testing.T) {
	// Test case: set an invalid metadata display and check for errors

	msg := SampleMsgSetDenomMetadata

	// set the metadata display to an invalid value and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Metadata.Display,
		&errorPacks.InvalidDenomMetaDisplay,
		ErrInvalidMetadata,
	)
}

func TestMsgSetDenomMetadata_InvalidDenomMetaName(t *testing.T) {
	// Test case: set an invalid metadata name and check for errors

	msg := SampleMsgSetDenomMetadata

	// set the metadata name to an invalid value and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Metadata.Name,
		&errorPacks.InvalidDenomMetaName,
		ErrInvalidMetadata,
	)
}

func TestMsgSetDenomMetadata_InvalidDenomMetaSymbol(t *testing.T) {
	// Test case: set an invalid metadata symbol and check for errors

	msg := SampleMsgSetDenomMetadata

	// set the metadata symbol to an invalid value and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Metadata.Symbol,
		&errorPacks.InvalidDenomMetaSymbol,
		ErrInvalidMetadata,
	)
}

func TestMsgSetDenomMetadata_InvalidDenomMetaUnit(t *testing.T) {
	// Test case: set an empty metadata unit and check for errors

	// create a slice of subtests (define test cases with different invalid metadata units)
	tests := []struct {
		// name/description of the test case
		name string
		// metadata unit that will be created on MsgSetDenomMetadata
		unit *banktypes.DenomUnit
		// expected error message for the test case
		errorMessage string
	}{
		{
			name: "Empty Denom",
			unit: &banktypes.DenomUnit{
				Denom:    "",
				Exponent: 0,
				Aliases: []string{
					"uzig",
				},
			},
			errorMessage: "metadata's first denomination unit must be the one with base denom 'uzig': Metadata is not valid",
		},
		{
			name: "Base Exponent must be 0",
			unit: &banktypes.DenomUnit{
				Denom:    "uzig",
				Exponent: 1,
				Aliases: []string{
					"uzig",
				},
			},
			errorMessage: "the exponent for base denomination unit uzig must be 0: Metadata is not valid",
		},
		{
			name: "Must contain a denomination unit",
			unit: &banktypes.DenomUnit{
				Denom:    "uzig",
				Exponent: 0,
				Aliases: []string{
					"uzig",
				},
			},
			errorMessage: "metadata must contain a denomination unit with display denom 'zig': Metadata is not valid",
		},
		{
			name: "Alias cannot be blank",
			unit: &banktypes.DenomUnit{
				Denom:    "uzig",
				Exponent: 0,
				Aliases: []string{
					"",
				},
			},
			errorMessage: "alias for denom unit uzig cannot be blank: Metadata is not valid",
		},
	}

	// make a copy of sample message
	msg := SampleMsgSetDenomMetadata

	// iterate through each test case
	for _, test := range tests {
		t.Run("MsgSetDenomMetadata.DenomUnit "+test.name, func(t *testing.T) {
			// set the metadata unit to an invalid value and check for errors
			msg.Metadata.DenomUnits = []*banktypes.DenomUnit{test.unit}
			// validate the message
			err := msg.ValidateBasic()
			// check if there was an error
			require.Error(t, err)
			// assert that the error is of type ErrInvalidMetadata
			require.ErrorIs(t, err, ErrInvalidMetadata)
			// assert that the error message matches the expected error message
			require.EqualError(t, err, test.errorMessage)
		})
	}
}

func TestMsgSetDenomMetadata_ValidateBasic_InvalidURI(t *testing.T) {
	// Test case: invalid URI

	// make a copy of sample message
	msg := SampleMsgSetDenomMetadata

	// set invalid uri and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Metadata.URI,
		&errorPacks.InvalidDenomURI,
		ErrInvalidMetadata,
	)
}

func TestMsgSetDenomMetadata_ValidateBasic_InvalidURIHash(t *testing.T) {
	// Test case: invalid URI hash

	// make a copy of sample message
	msg := SampleMsgSetDenomMetadata

	// set invalid uri hash and check for errors
	types.CheckMessageStringFieldErrors(
		t,
		&msg,
		&msg.Metadata.URIHash,
		&errorPacks.InvalidDenomURIHash,
		ErrInvalidMetadata,
	)
}
