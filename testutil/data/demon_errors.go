package errorPacks

import (
	"fmt"
	"strings"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/zutils/constants"
)

// longCoinDenom - example of Coin with too long denom name
var longCoinDenom = sdk.Coin{Denom: strings.Repeat("a", 130), Amount: math.NewInt(10)}
var longCoinSubDenom = sdk.Coin{Denom: strings.Repeat("a", 60), Amount: math.NewInt(10)}

// InvalidDenomName - denom name tests
var InvalidDenomName = []struct {
	// Name/description of the test case
	TestName string
	// Field value
	FieldValue sdk.Coin
	// expected error message for the test case
	ErrorText string
}{
	{
		TestName:   "empty_denom_name",
		FieldValue: sdk.Coin{Denom: "", Amount: math.NewInt(10)},
		ErrorText:  "invalid coin: denomination 10 cannot be empty (e.g., 10uzig)",
	},
	{
		TestName:   "too short denom name",
		FieldValue: sdk.Coin{Denom: "lp", Amount: math.NewInt(10)},
		ErrorText:  "invalid coin: 'lp' denom name is too short for 10lp, minimum 3 characters e.g. 10uzig",
	},
	{
		TestName:   "invalid denom format: special characters exclamation sign",
		FieldValue: sdk.Coin{Denom: "btc!", Amount: math.NewInt(10)},
		ErrorText:  "invalid coin: '10btc!' only letters (a-z, A-Z), numbers (0-9), dots (.) and forward slashes (/) are allowed e.g. 10uzig",
	},
	{
		TestName:   "invalid denom format: special characters underscore",
		FieldValue: sdk.Coin{Denom: "btc_", Amount: math.NewInt(10)},
		ErrorText:  "invalid coin: '10btc_' only letters (a-z, A-Z), numbers (0-9), dots (.) and forward slashes (/) are allowed e.g. 10uzig",
	},
	{
		TestName:   "invalid denom format: special characters question mark",
		FieldValue: sdk.Coin{Denom: "btc?eth", Amount: math.NewInt(10)},
		ErrorText:  "invalid coin: '10btc?eth' only letters (a-z, A-Z), numbers (0-9), dots (.) and forward slashes (/) are allowed e.g. 10uzig",
	},
	{
		TestName:   "invalid denom format: space in the middle",
		FieldValue: sdk.Coin{Denom: "btc eth", Amount: math.NewInt(10)},
		ErrorText:  "invalid coin: '10btc eth' only letters (a-z, A-Z), numbers (0-9), dots (.) and forward slashes (/) are allowed e.g. 10uzig",
	},
	{
		TestName:   "invalid denom format: trailing space",
		FieldValue: sdk.Coin{Denom: "btc ", Amount: math.NewInt(10)},
		ErrorText:  "invalid coin: '10btc ' only letters (a-z, A-Z), numbers (0-9), dots (.) and forward slashes (/) are allowed e.g. 10uzig",
	},
	{
		TestName:   "invalid denom format: starting space",
		FieldValue: sdk.Coin{Denom: " btc", Amount: math.NewInt(10)},
		ErrorText:  "invalid coin: '10 btc' only letters (a-z, A-Z), numbers (0-9), dots (.) and forward slashes (/) are allowed e.g. 10uzig",
	},
	{
		TestName:   "too long denom name",
		FieldValue: longCoinDenom,
		ErrorText: fmt.Sprintf(
			"invalid coin: '%s' denom name is too long (%d) for %s, maximum %d characters e.g. uzig",
			longCoinDenom.Denom,
			len(longCoinDenom.Denom),
			longCoinDenom.String(),
			constants.MaxDenomLength,
		),
	},
	{
		TestName:   "invalid denom format: starting with number",
		FieldValue: sdk.Coin{Denom: "1btc", Amount: math.NewInt(10)},
		ErrorText:  "invalid coin: '101btc' only letters (a-z, A-Z), numbers (0-9), dots (.) and forward slashes (/) are allowed e.g. 10uzig",
	},
	{
		TestName:   "invalid denom format: starting with special character dot",
		FieldValue: sdk.Coin{Denom: ".btc", Amount: math.NewInt(10)},
		ErrorText:  "invalid coin: '10.btc' only letters (a-z, A-Z), numbers (0-9), dots (.) and forward slashes (/) are allowed e.g. 10uzig",
	},
	{
		TestName:   "invalid denom format: starting with special character /",
		FieldValue: sdk.Coin{Denom: "/btc", Amount: math.NewInt(10)},
		ErrorText:  "invalid coin: '10/btc' only letters (a-z, A-Z), numbers (0-9), dots (.) and forward slashes (/) are allowed e.g. 10uzig",
	},
	{
		TestName:   "invalid denom format: with non ASCII characters unicod",
		FieldValue: sdk.Coin{Denom: "uatóm", Amount: math.NewInt(10)},
		ErrorText:  "invalid coin: '10uatóm' only letters (a-z, A-Z), numbers (0-9), dots (.) and forward slashes (/) are allowed e.g. 10uzig",
	},
	{
		TestName:   "invalid denom format: with non ASCII characters emoji",
		FieldValue: sdk.Coin{Denom: "u🚀zig", Amount: math.NewInt(10)},
		ErrorText:  "invalid coin: '10u🚀zig' only letters (a-z, A-Z), numbers (0-9), dots (.) and forward slashes (/) are allowed e.g. 10uzig",
	},
	{
		TestName:   "invalid denom format: with non ASCII characters chinese characters",
		FieldValue: sdk.Coin{Denom: "u币zig", Amount: math.NewInt(10)},
		ErrorText:  "invalid coin: '10u币zig' only letters (a-z, A-Z), numbers (0-9), dots (.) and forward slashes (/) are allowed e.g. 10uzig",
	},
	{
		TestName:   "invalid denom format: with tab character",
		FieldValue: sdk.Coin{Denom: "u\tzig", Amount: math.NewInt(10)},
		ErrorText:  "invalid coin: '10u\tzig' only letters (a-z, A-Z), numbers (0-9), dots (.) and forward slashes (/) are allowed e.g. 10uzig",
	},
	{
		TestName:   "invalid denom format: with new line character",
		FieldValue: sdk.Coin{Denom: "u\nzig", Amount: math.NewInt(10)},
		ErrorText:  "invalid coin: '10u\nzig' only letters (a-z, A-Z), numbers (0-9), dots (.) and forward slashes (/) are allowed e.g. 10uzig",
	},
	// {
	// 	TestName:   "invalid denom format: uppercase letters embedded",
	// 	FieldValue: sdk.Coin{Denom: "bTc123", Amount: math.NewInt(10)},
	// 	ErrorText:  "invalid coin: '10bTc123' only lowercase letters (a-z) followed by lowercase letters (a-z), and numbers (0-9) are allowed e.g. 10uzig",
	// },
}

// InvalidDenomNameString - denom name string tests
var InvalidDenomNameString = []struct {
	// Name/description of the test case
	TestName string
	// Field value
	FieldValue string
	// expected error message for the test case
	ErrorText string
}{
	{
		TestName:   "empty_denom_name",
		FieldValue: "",
		ErrorText:  "invalid coin: denomination '' cannot be empty (e.g., 10uzig)",
	},
	{
		TestName:   "too short denom name",
		FieldValue: "lp",
		ErrorText:  "invalid coin: 'lp' denom name is too short, minimum 3 characters e.g. 10uzig",
	},
	{
		TestName:   "invalid denom format: special characters exclamation sign",
		FieldValue: "btc!",
		ErrorText:  "invalid coin: 'btc!' only letters (a-z, A-Z), numbers (0-9), dots (.) and forward slashes (/) are allowed e.g. 10uzig",
	},
	{
		TestName:   "invalid denom format: space",
		FieldValue: "btc eth",
		ErrorText:  "invalid coin: 'btc eth' only letters (a-z, A-Z), numbers (0-9), dots (.) and forward slashes (/) are allowed e.g. 10uzig",
	},
	{
		TestName:   "too long denom name",
		FieldValue: strings.Repeat("a", 130),
		ErrorText: fmt.Sprintf(
			"invalid coin: '%s' denom name is too long (%d), maximum %d characters e.g. uzig",
			strings.Repeat("a", 130),
			len(strings.Repeat("a", 130)),
			constants.MaxDenomLength,
		),
	},
	// {
	// 	TestName:   "invalid denom format: uppercase letters embedded",
	// 	FieldValue: sdk.Coin{Denom: "bTc123", Amount: math.NewInt(10)},
	// 	ErrorText:  "invalid coin: '10bTc123' only lowercase letters (a-z), numbers (0-9), '/' and ':' are allowed e.g. 10uzig",
	// },
	{
		TestName:   "invalid denom format: question mark",
		FieldValue: "btc?eth",
		ErrorText:  "invalid coin: 'btc?eth' only letters (a-z, A-Z), numbers (0-9), dots (.) and forward slashes (/) are allowed e.g. 10uzig",
	},
	{
		TestName:   "starting with number",
		FieldValue: "1btc",
		ErrorText:  "invalid coin: '1btc' only letters (a-z, A-Z), numbers (0-9), dots (.) and forward slashes (/) are allowed e.g. 10uzig",
	},
}

// InvalidDenomAmountZeroOK it means we need to test only for negative amount
var InvalidDenomAmountZeroOK = []struct {
	// Name/description of the test case
	TestName string
	// Field value
	FieldValue sdk.Coin
	// expected error message for the test case
	ErrorText string
}{
	{
		TestName:   "invalid denom amount: check for negative",
		FieldValue: sdk.Coin{Denom: "lpk", Amount: math.NewInt(-10)},
		ErrorText:  "invalid coin amount: -10 cannot be negative (-10lpk)",
	},
}

// InvalidDenomAmountZeroNotOK checks if the amount is zero (will merge with InvalidDenomAmountZeroOK bellow)
var InvalidDenomAmountZeroNotOK = []struct {
	// Name/description of the test case
	TestName string
	// Field value
	FieldValue sdk.Coin
	// expected error message for the test case
	ErrorText string
}{
	{
		TestName:   "invalid denom amount: zero",
		FieldValue: sdk.Coin{Denom: "lpk", Amount: math.NewInt(0)},
		ErrorText:  "invalid coin amount: 0 has to be positive (0lpk)",
	},
}

// InvalidSubDenom checks for sub denom name
var InvalidSubDenomString = []struct {
	// Name/description of the test case
	TestName string
	// Field value
	FieldValue string
	// expected error message for the test case
	ErrorText string
}{
	{
		TestName:   "empty_subdenom_name",
		FieldValue: "",
		ErrorText:  "Invalid subdenom name: denom name is empty e.g. uzig",
	},
	{
		TestName:   "too_short_subdenom_name",
		FieldValue: "lp",
		ErrorText:  "invalid coin: 'lp' denom name is too short, minimum 3 characters e.g. uzig",
	},
	{
		TestName:   "invalid subdenom format: special characters exclamation sign",
		FieldValue: "btc!",
		ErrorText:  "invalid coin: 'btc!' only lowercase letters (a-z) and numbers (0-9) are allowed e.g. uzig123",
	},
	{
		TestName:   "invalid denom format: space",
		FieldValue: "btc eth",
		ErrorText:  "invalid coin: 'btc eth' only lowercase letters (a-z) and numbers (0-9) are allowed e.g. uzig123",
	},
	{
		TestName:   "too long subdenom name",
		FieldValue: longCoinSubDenom.Denom,
		ErrorText: fmt.Sprintf(
			"invalid coin: '%s' denom name is too long (%d), maximum %d characters e.g. uzig",
			longCoinSubDenom.Denom,
			len(longCoinSubDenom.Denom),
			constants.MaxSubDenomLength,
		),
	},
	{
		TestName:   "invalid subdenom format: uppercase letters embedded",
		FieldValue: "bTc123",
		ErrorText:  "invalid coin: 'bTc123' only lowercase letters (a-z) and numbers (0-9) are allowed e.g. uzig123",
	},
	{
		TestName:   "invalid subdenom format: question mark",
		FieldValue: "btc?eth",
		ErrorText:  "invalid coin: 'btc?eth' only lowercase letters (a-z) and numbers (0-9) are allowed e.g. uzig123",
	},
	{
		TestName:   "invalid subdenom format: symbol not allowed in our denom but allowed in default denom :_-",
		FieldValue: "btc:_-eth",
		ErrorText:  "invalid coin: 'btc:_-eth' only lowercase letters (a-z) and numbers (0-9) are allowed e.g. uzig123",
	},
	{
		TestName:   "invalid subdenom format: no starting with letter but number",
		FieldValue: "2btc",
		ErrorText:  "invalid coin: '2btc' denom name has to start with a lowercase letter e.g. uzig",
	},
	{
		TestName:   "invalid subdenom format: no starting with letter but slash",
		FieldValue: "/btc",
		ErrorText:  "invalid coin: '/btc' denom name has to start with a lowercase letter e.g. uzig",
	},
}

// Merge names first

var InvalidDenomZeroAmountOK = append(InvalidDenomAmountZeroOK, InvalidDenomName...)
var InvalidDenomZeroAmountNotOK = append(InvalidDenomAmountZeroNotOK, InvalidDenomZeroAmountOK...)
var InvalidDenomZeroAmountNotOKOnlyAmounts = append(InvalidDenomAmountZeroOK, InvalidDenomAmountZeroNotOK...)

// InvalidDenomMetaDescription - denom description tests
var InvalidDenomMetaDescription = []struct {
	// Name/description of the test case
	TestName string
	// Field value
	FieldValue string
	// expected error message for the test case
	ErrorText string
}{
	{
		TestName:   "too short denom meta description",
		FieldValue: "ab",
		ErrorText:  "Description length must be between 3 and 255",
	},
	{
		TestName:   "too long denom meta description",
		FieldValue: strings.Repeat("a", 256),
		ErrorText:  "Description length must be between 3 and 255",
	},
}

// InvalidDenomMetaBase - denom base tests
var InvalidDenomMetaBase = []struct {
	// Name/description of the test case
	TestName string
	// Field value
	FieldValue string
	// expected error message for the test case
	ErrorText string
}{
	{
		TestName:   "empty_denom_meta_base",
		FieldValue: "",
		ErrorText:  "invalid metadata base denom: invalid denom: ",
	},
	{
		TestName:   "invalid_denom_meta_base",
		FieldValue: "btc",
		ErrorText:  "metadata's first denomination unit must be the one with base denom 'btc'",
	},
}

// InvalidDenomMetaDisplay - denom display tests
var InvalidDenomMetaDisplay = []struct {
	// Name/description of the test case
	TestName string
	// Field value
	FieldValue string
	// expected error message for the test case
	ErrorText string
}{
	{
		TestName:   "empty_denom_meta_display",
		FieldValue: "",
		ErrorText:  "invalid metadata display denom: invalid denom: ",
	},
	{
		TestName:   "invalid_denom_meta_display_symbols",
		FieldValue: "#$%§",
		ErrorText:  "invalid metadata display denom: invalid denom: #$%§",
	},
	{
		TestName:   "invalid_denom_meta_display",
		FieldValue: "btc",
		ErrorText:  "metadata must contain a denomination unit with display denom 'btc'",
	},
}

// Invalid InvalidDenomDenomMetaName - denom name tests
var InvalidDenomMetaName = []struct {
	// Name/description of the test case
	TestName string
	// Field value
	FieldValue string
	// expected error message for the test case
	ErrorText string
}{
	{
		TestName:   "empty_denom_meta_name",
		FieldValue: "",
		ErrorText:  "name field cannot be blank",
	},
}

// InvalidDenomMetaSymbol - denom symbol tests
var InvalidDenomMetaSymbol = []struct {
	// Name/description of the test case
	TestName string
	// Field value
	FieldValue string
	// expected error message for the test case
	ErrorText string
}{
	{
		TestName:   "empty_denom_meta_symbol",
		FieldValue: "",
		ErrorText:  "symbol field cannot be blank",
	},
}

// InvalidDenomURI
var InvalidDenomURI = []struct {
	// Name/description of the test case
	TestName string
	// Field value
	FieldValue string
	// expected error message for the test case
	ErrorText string
}{
	{
		TestName:   "invalid_uri",
		FieldValue: "invalid_uri_invalid_uri_invalid_uri",
		ErrorText:  "URI must be a valid URL",
	},
	{
		TestName:   "invalid_uri_too_long",
		FieldValue: "https://www.example.com/" + strings.Repeat("a", 250),
		ErrorText: fmt.Sprintf(
			"URI: %s length must be between 15 and %d characters",
			"https://www.example.com/"+strings.Repeat("a", 250), constants.MaxURILength),
	},
	{
		TestName:   "invalid_uri_too_short",
		FieldValue: "https://",
		ErrorText: fmt.Sprintf(
			"URI: %s length must be between 15 and %d characters",
			"https://", constants.MaxURILength),
	},
}

// InvalidDenomURIHash
var InvalidDenomURIHash = []struct {
	// Name/description of the test case
	TestName string
	// Field value
	FieldValue string
	// expected error message for the test case
	ErrorText string
}{
	{
		TestName:   "invalid_uri_hash_too_long",
		FieldValue: strings.Repeat("a2x", 25),
		ErrorText: fmt.Sprintf(
			"URIHash: %s (len of %d) for URL: %s must be a valid SHA256 hash string of 64 alpha numeric characters",
			strings.Repeat("a2x", 25),
			len(strings.Repeat("a2x", 25)),
			"https://example.com",
		),
	},
	{
		TestName:   "invalid_uri_hash_too_short",
		FieldValue: "a2x",
		ErrorText: fmt.Sprintf(
			"URIHash: %s (len of %d) for URL: %s must be a valid SHA256 hash string of 64 alpha numeric characters",
			"a2x",
			len("a2x"),
			"https://example.com",
		),
	},
	{
		TestName:   "invalid_uri_hash",
		FieldValue: strings.Repeat("a_21", 16),
		ErrorText: fmt.Sprintf(
			"URIHash: %s (len of %d) for URL: %s must be a valid SHA256 hash string of 64 alpha numeric characters",
			strings.Repeat("a_21", 16),
			len(strings.Repeat("a_21", 16)),
			"https://example.com",
		),
	},
}
