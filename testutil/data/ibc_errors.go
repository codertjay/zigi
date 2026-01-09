package errorPacks

import (
	"fmt"
	"strings"

	"zigchain/zutils/constants"
)

// InvalidClientID defines a slice of subtests for invalid ClientId values
var InvalidClientID = []struct {
	// Name/description of the test case
	TestName string
	// The value to test
	FieldValue string
	// The expected error message
	ErrorText string
}{
	{
		TestName:   "empty client ID",
		FieldValue: "",
		ErrorText:  "light client is invalid: client ID cannot be empty",
	},
	{
		TestName:   "too short client ID",
		FieldValue: "ab",
		ErrorText:  "light client is invalid: invalid client ID format",
	},
	{
		TestName:   "too long client ID",
		FieldValue: "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmno", // 65 chars
		ErrorText:  "light client is invalid: invalid client ID format",
	},
	{
		TestName:   "invalid characters - space",
		FieldValue: "client id with space",
		ErrorText:  "light client is invalid: invalid client ID format",
	},
	{
		TestName:   "invalid characters - symbol %",
		FieldValue: "client%id",
		ErrorText:  "light client is invalid: invalid client ID format",
	},
	{
		TestName:   "invalid characters - emoji",
		FieldValue: "client😀id",
		ErrorText:  "light client is invalid: invalid client ID format",
	},
	{
		TestName:   "only special characters",
		FieldValue: "!@#$%^&*()",
		ErrorText:  "light client is invalid: invalid client ID format",
	},
}

// InvalidClientIDTypes defines a slice of subtests for invalid ClientId types
var InvalidClientIDTypes = []struct {
	TestName  string
	Input     interface{}
	ErrorText string
}{
	{
		TestName:  "int instead of string",
		Input:     123,
		ErrorText: "invalid parameter type: int",
	},
	{
		TestName:  "bool instead of string",
		Input:     true,
		ErrorText: "invalid parameter type: bool",
	},
	{
		TestName:  "struct instead of string",
		Input:     struct{}{},
		ErrorText: "invalid parameter type: struct {}",
	},
}

// InvalidPort defines a slice of subtests for invalid Port values
var InvalidPort = []struct {
	TestName   string
	FieldValue string
	ErrorText  string
}{
	{
		TestName:   "empty port",
		FieldValue: "",
		ErrorText:  "invalid port: port cannot be empty",
	},
	{
		TestName:   "too short port",
		FieldValue: "a",
		ErrorText:  "invalid port: port length must be between 2 and 128 characters",
	},
	{
		TestName:   "too long port",
		FieldValue: strings.Repeat("p", 129),
		ErrorText:  "invalid port: port length must be between 2 and 128 characters",
	},
	{
		TestName:   "port with space",
		FieldValue: "port name",
		ErrorText:  "invalid port: port contains invalid characters. Only alphanumeric, ., _, +, -, #, [, ], <, > are allowed",
	},
	{
		TestName:   "port with invalid symbol %",
		FieldValue: "port%name",
		ErrorText:  "invalid port: port contains invalid characters. Only alphanumeric, ., _, +, -, #, [, ], <, > are allowed",
	},
	{
		TestName:   "only invalid characters",
		FieldValue: "!@#$",
		ErrorText:  "invalid port: port contains invalid characters. Only alphanumeric, ., _, +, -, #, [, ], <, > are allowed",
	},
}

// InvalidPortTypes includes non-string values for Port
var InvalidPortTypes = []struct {
	TestName  string
	Input     interface{}
	ErrorText string
}{
	{
		TestName:  "int instead of string",
		Input:     123,
		ErrorText: "invalid parameter type: int",
	},
	{
		TestName:  "bool instead of string",
		Input:     false,
		ErrorText: "invalid parameter type: bool",
	},
	{
		TestName:  "nil instead of string",
		Input:     nil,
		ErrorText: "invalid parameter type: <nil>",
	},
	{
		TestName:  "slice instead of string",
		Input:     []string{"port1", "port2"},
		ErrorText: "invalid parameter type: []string",
	},
}

// InvalidChannel defines invalid values for the Channel field
var InvalidChannel = []struct {
	TestName   string
	FieldValue string
	ErrorText  string
}{
	{
		TestName:   "empty channel",
		FieldValue: "",
		ErrorText:  "invalid channel identifier: channel cannot be empty",
	},
	{
		TestName:   "too short channel",
		FieldValue: "chan123",
		ErrorText:  "invalid channel identifier: invalid channel ID format",
	},
	{
		TestName:   "too long channel",
		FieldValue: "channel" + strings.Repeat("x", 60), // total length = 67
		ErrorText:  "invalid channel identifier: invalid channel ID format",
	},
	{
		TestName:   "channel with space",
		FieldValue: "chan nel01",
		ErrorText:  "invalid channel identifier: invalid channel ID format",
	},
	{
		TestName:   "channel with % symbol",
		FieldValue: "chan%nel01",
		ErrorText:  "invalid channel identifier: invalid channel ID format",
	},
	{
		TestName:   "channel only invalid characters",
		FieldValue: "@#$%^&*()",
		ErrorText:  "invalid channel identifier: invalid channel ID format",
	},
}

// InvalidChannelTypes includes non-string values for Channel
var InvalidChannelTypes = []struct {
	TestName  string
	Input     interface{}
	ErrorText string
}{
	{
		TestName:  "int instead of string",
		Input:     456,
		ErrorText: "invalid parameter type: int",
	},
	{
		TestName:  "bool instead of string",
		Input:     true,
		ErrorText: "invalid parameter type: bool",
	},
	{
		TestName:  "nil instead of string",
		Input:     nil,
		ErrorText: "invalid parameter type: <nil>",
	},
	{
		TestName:  "map instead of string",
		Input:     map[string]int{"chan": 1},
		ErrorText: "invalid parameter type: map[string]int",
	},
	{
		TestName:  "slice instead of string",
		Input:     []string{"chan1", "chan2"},
		ErrorText: "invalid parameter type: []string",
	},
}

// InvalidDenom defines a slice of subtests for invalid Denom values
var InvalidDenom = []struct {
	TestName   string
	FieldValue string
	ErrorText  string
}{
	{
		TestName:   "empty denom",
		FieldValue: "",
		ErrorText:  "denom cannot be empty",
	},
	{
		TestName:   "denom with space",
		FieldValue: "u zig",
		ErrorText:  "denom contains invalid characters. Only alphanumeric, ., _, +, -, #, [, ], <, > are allowed",
	},
	{
		TestName:   "denom with invalid symbol %",
		FieldValue: "u%zig",
		ErrorText:  "denom contains invalid characters. Only alphanumeric, ., _, +, -, #, [, ], <, > are allowed",
	},
	{
		TestName:   "denom with emoji",
		FieldValue: "u🚀zig",
		ErrorText:  "denom contains invalid characters. Only alphanumeric, ., _, +, -, #, [, ], <, > are allowed",
	},
	{
		TestName:   "too short denom",
		FieldValue: "zz",
		ErrorText:  fmt.Sprintf("invalid coin: 'zz' denom name is too short, minimum %d characters e.g. 10uzig: invalid coins", constants.MinSubDenomLength),
	},
	{
		TestName:   "too long denom",
		FieldValue: "u" + strings.Repeat("z", constants.MaxDenomLength), // length = MaxDenomLength + 1
		ErrorText:  fmt.Sprintf("invalid coin: '%s' denom name is too long (%d), maximum %d characters e.g. uzig: invalid coins", "u"+strings.Repeat("z", constants.MaxDenomLength), constants.MaxDenomLength+1, constants.MaxDenomLength),
	},
}

// InvalidDenomTypes includes non-string values for Denom
var InvalidDenomTypes = []struct {
	TestName  string
	Input     interface{}
	ErrorText string
}{
	{
		TestName:  "int instead of string",
		Input:     999,
		ErrorText: "invalid parameter type: int",
	},
	{
		TestName:  "bool instead of string",
		Input:     false,
		ErrorText: "invalid parameter type: bool",
	},
	{
		TestName:  "nil instead of string",
		Input:     nil,
		ErrorText: "invalid parameter type: <nil>",
	},
	{
		TestName:  "slice instead of string",
		Input:     []string{"uzig", "uatom"},
		ErrorText: "invalid parameter type: []string",
	},
}

// InvalidDecimalDifference defines a slice of subtests for invalid DecimalDifference values
var InvalidDecimalDifference = []struct {
	TestName   string
	FieldValue uint32
	ErrorText  string
}{
	{
		TestName:   "decimal difference above max",
		FieldValue: 19,
		ErrorText:  "decimal difference cannot be greater than 18",
	},
	{
		TestName:   "decimal difference far above max",
		FieldValue: 255,
		ErrorText:  "decimal difference cannot be greater than 18",
	},
}

// InvalidDecimalDifference defines a slice of subtests for invalid DecimalDifference values
var InvalidDecimalDifferenceTypes = []struct {
	TestName  string
	Input     interface{}
	ErrorText string
}{
	{
		TestName:  "string instead of uint32",
		Input:     "18",
		ErrorText: "invalid parameter type: string",
	},
	{
		TestName:  "float instead of uint32",
		Input:     18.0,
		ErrorText: "invalid parameter type: float64",
	},
	{
		TestName:  "bool instead of uint32",
		Input:     true,
		ErrorText: "invalid parameter type: bool",
	},
	{
		TestName:  "int (signed) instead of uint32",
		Input:     -1,
		ErrorText: "invalid parameter type: int",
	},
	{
		TestName:  "nil input",
		Input:     nil,
		ErrorText: "invalid parameter type: <nil>",
	},
}
