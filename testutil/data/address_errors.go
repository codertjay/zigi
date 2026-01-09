package errorPacks

import (
	"fmt"
	"strings"
)

// longAddress - example of address with too long denom name
var longAddress = "zig1" + strings.Repeat("a", 100)

// InvalidSignerAddress create a slice of subtests (define test cases with different invalid Signer address)
var InvalidSignerAddress = []struct {
	// Name/description of the test case
	TestName string
	// Field value
	FieldValue string
	// expected error message for the test case
	ErrorText string
}{
	{
		TestName:   "invalid address format",
		FieldValue: "invalid_address",
		ErrorText:  "SIGNER ADDRESS: 'invalid_address' (decoding bech32 failed: invalid separator index -1)",
	},
	{
		TestName:   "empty address",
		FieldValue: "",
		ErrorText:  "SIGNER ADDRESS: '' (empty address string is not allowed)",
	},
	{
		TestName:   "too short address",
		FieldValue: "zig123",
		ErrorText:  "SIGNER ADDRESS: 'zig123' (decoding bech32 failed: invalid bech32 string length 6)",
	},
	{
		TestName: "too long address",
		// checked: used for testing
		FieldValue: "zig1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
		ErrorText: "SIGNER ADDRESS: " +
			// checked: used for testing
			"'zig1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890' " +
			"(decoding bech32 failed: invalid checksum (expected yurny3 got 567890))",
	},
	{
		TestName:   "address is not all lowercase or all uppercase",
		FieldValue: "Zig12z7zhe2d03wn03e8lkw8At6zjxa7y65q5xnp5t",
		// checked: used for testing
		ErrorText: "SIGNER ADDRESS: 'Zig12z7zhe2d03wn03e8lkw8At6zjxa7y65q5xnp5t' " +
			"(decoding bech32 failed: string not all lowercase or all uppercase)",
	},
	{
		TestName:   "address contains invalid characters",
		FieldValue: "zi!12z7zhe2d03wn03e8lkw8at6zjxa7y65q5xnp5%",
		ErrorText: "SIGNER ADDRESS: 'zi!12z7zhe2d03wn03e8lkw8at6zjxa7y65q5xnp5%' " +
			"(decoding bech32 failed: invalid character not part of charset: 37)",
	},
	{
		TestName:   "address contains only numbers",
		FieldValue: "123121712321032203281238456456771651512351",
		// checked: used for testing
		ErrorText: "SIGNER ADDRESS: '123121712321032203281238456456771651512351' " +
			"(decoding bech32 failed: invalid separator index 41)",
	},
	{
		TestName:   "address contains space in string",
		FieldValue: "abcdefghijabcdefghijabcde ghijabcdefghijabcdefghij",
		ErrorText: "SIGNER ADDRESS: 'abcdefghijabcdefghijabcde ghijabcdefghijabcdefghij' " +
			"(decoding bech32 failed: invalid character in string: ' ')",
	},
	{
		TestName:   "address too short",
		FieldValue: "a",
		ErrorText:  "SIGNER ADDRESS: 'a' (decoding bech32 failed: invalid bech32 string length 1)",
	},
	{
		TestName: "address has invalid prefix",
		// checked: used for testing
		FieldValue: "yoy193fxruxcm8y32njzt23c86en7hd8tajma79tt3",
		// checked: used for testing
		ErrorText: "SIGNER ADDRESS: 'yoy193fxruxcm8y32njzt23c86en7hd8tajma79tt3' " +
			"(decoding bech32 failed: invalid checksum (expected z6kscw got a79tt3))",
	},
	{
		TestName: "address has spec chars",
		// checked: used for testing
		FieldValue: "zig1/\\.%&?32njzt23c86en7hd8tajma79tt3",
		// checked: used for testing
		ErrorText: "SIGNER ADDRESS: 'zig1/\\.%&?32njzt23c86en7hd8tajma79tt3' (decoding bech32 failed: invalid character not part of charset: 47)",
	},
}

// InvalidReceiverAddress create a slice of subtests (define test cases with different invalid Receiver address)
var InvalidReceiverAddress = []struct {
	// Name/description of the test case
	TestName string
	// Field value
	FieldValue string
	// expected error message for the test case
	ErrorText string
}{
	{
		TestName:   "invalid address format",
		FieldValue: "bad_address",
		ErrorText:  "Invalid receiver address: bad_address",
	},
	{
		TestName:   "too short address",
		FieldValue: "zig123",
		ErrorText:  "Invalid receiver address: zig123",
	},
	{
		TestName: "too long address",
		// checked: used for testing
		FieldValue: "zig1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
		ErrorText: "Invalid receiver address: " +
			"zig1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890", // checked: used for testing

	},
	{
		TestName:   "address is not all lowercase or all uppercase",
		FieldValue: "Zig12z7zhe2d03wn03e8lkw8At6zjxa7y65q5xnp5t",
		ErrorText:  "Invalid receiver address: Zig12z7zhe2d03wn03e8lkw8At6zjxa7y65q5xnp5t",
	},
	{
		TestName:   "address contains invalid characters",
		FieldValue: "zi!12z7zhe2d03wn03e8lkw8at6zjxa7y65q5xnp5%",
		ErrorText:  "Invalid receiver address: zi!12z7zhe2d03wn03e8lkw8at6zjxa7y65q5xnp5%",
	},
	{
		TestName:   "address contains only numbers",
		FieldValue: "123121712321032203281238456456771651512351",
		ErrorText:  "Invalid receiver address: 123121712321032203281238456456771651512351",
	},
	{
		TestName:   "address contains space in string",
		FieldValue: "abcdefghijabcdefghijabcde ghijabcdefghijabcdefghij",
		ErrorText:  "Invalid receiver address: abcdefghijabcdefghijabcde ghijabcdefghijabcdefghij",
	},
	{
		TestName: "address has invalid prefix",
		// checked: used for testing
		FieldValue: "yoy193fxruxcm8y32njzt23c86en7hd8tajma79tt3",
		// checked: used for testing
		ErrorText: "Invalid receiver address: yoy193fxruxcm8y32njzt23c86en7hd8tajma79tt3",
	},
	{
		TestName: "address has spec chars",
		// checked: used for testing
		FieldValue: "zig1/\\.%&?32njzt23c86en7hd8tajma79tt3",
		// checked: used for testing
		ErrorText: "Invalid receiver address: zig1/\\.%&?32njzt23c86en7hd8tajma79tt3",
	},
}

func InvalidAddressErrors(field string) []struct {
	// Name/description of the test case
	TestName string
	// Field value
	FieldValue string
	// expected error message for the test case
	ErrorText string
} {
	// InvalidSignerAddress create a slice of subtests (define test cases with different invalid Signer address)
	return []struct {
		// Name/description of the test case
		TestName string
		// Field value
		FieldValue string
		// expected error message for the test case
		ErrorText string
	}{
		{
			TestName:   fmt.Sprintf("Invalid %s address format", field),
			FieldValue: "zig_invalid_address",
			ErrorText: fmt.Sprintf(
				"%s address: 'zig_invalid_address' (decoding bech32 failed: invalid separator index -1)",
				field,
			),
		},
		{
			TestName:   fmt.Sprintf("%s empty address", field),
			FieldValue: "",
			ErrorText: fmt.Sprintf(
				"%s address: cannot be empty",
				field,
			),
		},
		{
			TestName:   fmt.Sprintf("%s: too short address", field),
			FieldValue: "zig123",
			ErrorText: fmt.Sprintf(
				"%s address: 'zig123' (decoding bech32 failed: invalid bech32 string length 6)",
				field,
			),
		},
		{
			TestName:   fmt.Sprintf("%s: too long address", field),
			FieldValue: longAddress,
			ErrorText: fmt.Sprintf(
				"%s address: '%s' (decoding bech32 failed: invalid checksum (expected 4zzudg got aaaaaa))",
				field,
				longAddress,
			),
		},
		{
			TestName:   fmt.Sprintf("%s: address is not all lowercase or all uppercase", field),
			FieldValue: "Zig12z7zhe2d03wn03e8lkw8At6zjxa7y65q5xnp5t",
			ErrorText: fmt.Sprintf(
				// checked: used for testing
				"%s address: 'Zig12z7zhe2d03wn03e8lkw8At6zjxa7y65q5xnp5t' "+
					"(decoding bech32 failed: string not all lowercase or all uppercase)",
				field,
			),
		},
		{
			TestName:   fmt.Sprintf("%s: address contains invalid characters", field),
			FieldValue: "zi!12z7zhe2d03wn03e8lkw8at6zjxa7y65q5xnp5%",
			ErrorText: fmt.Sprintf(
				// checked: used for testing
				"%s address: 'zi!12z7zhe2d03wn03e8lkw8at6zjxa7y65q5xnp5%%' "+
					"(decoding bech32 failed: invalid character not part of charset: 37)",
				field,
			),
		},
		{
			TestName:   fmt.Sprintf("%s: address contains only numbers", field),
			FieldValue: "123121712321032203281238456456771651512351",
			ErrorText: fmt.Sprintf(
				// checked: used for testing
				"%s address: '123121712321032203281238456456771651512351' "+
					"(decoding bech32 failed: invalid separator index 41)",
				field,
			),
		},
		{
			TestName:   fmt.Sprintf("%s: address contains space in string", field),
			FieldValue: "zig1abcdefghijabcdefghijabcde ghijabcdefghijabcdefghij",
			ErrorText: fmt.Sprintf(
				"%s address: 'zig1abcdefghijabcdefghijabcde ghijabcdefghijabcdefghij' "+
					"(decoding bech32 failed: invalid character in string: ' ')",
				field,
			),
		},
		{
			TestName:   fmt.Sprintf("%s: address too short", field),
			FieldValue: "a",
			ErrorText: fmt.Sprintf(
				"%s address: 'a' (decoding bech32 failed: invalid bech32 string length 1)",
				field,
			),
		},
		{
			TestName:   fmt.Sprintf("%s: address has invalid prefix", field),
			FieldValue: "yoy193fxruxcm8y32njzt23c86en7hd8tajma79tt3",
			ErrorText: fmt.Sprintf(
				// checked: used for testing
				"%s address: 'yoy193fxruxcm8y32njzt23c86en7hd8tajma79tt3' "+
					"(decoding bech32 failed: invalid checksum (expected z6kscw got a79tt3))",
				field,
			),
		},
	}
}

func InvalidAdminAddressErrors(field string) []struct {
	// Name/description of the test case
	TestName string
	// Field value
	FieldValue string
	// expected error message for the test case
	ErrorText string
} {
	// InvalidAdminAddress create a slice of subtests (define test cases with different invalid address)
	return []struct {
		// Name/description of the test case
		TestName string
		// Field value
		FieldValue string
		// expected error message for the test case
		ErrorText string
	}{
		{
			TestName:   fmt.Sprintf("Invalid %s address format empty", field),
			FieldValue: "",
			ErrorText: fmt.Sprintf(
				"%s address ()",
				field,
			),
		},
		{
			TestName:   fmt.Sprintf("Invalid %s address format", field),
			FieldValue: "invalid_address",
			ErrorText: fmt.Sprintf(
				"%s address (invalid_address)",
				field,
			),
		},
		{
			TestName:   fmt.Sprintf("%s: too short address", field),
			FieldValue: "zig123",
			ErrorText: fmt.Sprintf(
				"%s address (zig123)",
				field,
			),
		},
		{
			TestName:   fmt.Sprintf("%s: too long address", field),
			FieldValue: longAddress,
			ErrorText: fmt.Sprintf(
				"%s address (%s)",
				field,
				longAddress,
			),
		},
		{
			TestName:   fmt.Sprintf("%s: address is not all lowercase or all uppercase", field),
			FieldValue: "Zig12z7zhe2d03wn03e8lkw8At6zjxa7y65q5xnp5t",
			ErrorText: fmt.Sprintf(
				"%s address (Zig12z7zhe2d03wn03e8lkw8At6zjxa7y65q5xnp5t)",
				field,
			),
		},
		{
			TestName:   fmt.Sprintf("%s: address contains invalid characters", field),
			FieldValue: "zi!12z7zhe2d03wn03e8lkw8at6zjxa7y65q5xnp5%",
			ErrorText: fmt.Sprintf(
				"%s address (zi!12z7zhe2d03wn03e8lkw8at6zjxa7y65q5xnp5%%)",
				field,
			),
		},
		{
			TestName:   fmt.Sprintf("%s: address contains only numbers", field),
			FieldValue: "123121712321032203281238456456771651512351",
			ErrorText: fmt.Sprintf(
				"%s address (123121712321032203281238456456771651512351)",
				field,
			),
		},
		{
			TestName:   fmt.Sprintf("%s: address contains space in string", field),
			FieldValue: "abcdefghijabcdefghijabcde ghijabcdefghijabcdefghij",
			ErrorText: fmt.Sprintf(
				"%s address (abcdefghijabcdefghijabcde ghijabcdefghijabcdefghij)",
				field,
			),
		},
		{
			TestName:   fmt.Sprintf("%s: address has bad prefix", field),
			FieldValue: "yoy193fxruxcm8y32njzt23c86en7hd8tajma79tt3",
			ErrorText: fmt.Sprintf(
				"%s address (yoy193fxruxcm8y32njzt23c86en7hd8tajma79tt3)",
				field,
			),
		},
	}
}

func InvalidMetadataAdminAddressErrors(field string) []struct {
	// Name/description of the test case
	TestName string
	// Field value
	FieldValue string
	// expected error message for the test case
	ErrorText string
} {
	// InvalidMetadataAdminAddress create a slice of subtests (define test cases with different invalid address)
	return []struct {
		// Name/description of the test case
		TestName string
		// Field value
		FieldValue string
		// expected error message for the test case
		ErrorText string
	}{
		{
			TestName:   fmt.Sprintf("Invalid %s address format", field),
			FieldValue: "invalid_address",
			ErrorText: fmt.Sprintf(
				"%s address (invalid_address)",
				field,
			),
		},
		{
			TestName:   fmt.Sprintf("%s: too short address", field),
			FieldValue: "zig123",
			ErrorText: fmt.Sprintf(
				"%s address (zig123)",
				field,
			),
		},
		{
			TestName:   fmt.Sprintf("%s: too long address", field),
			FieldValue: longAddress,
			ErrorText: fmt.Sprintf(
				"%s address (%s)",
				field,
				longAddress,
			),
		},
		{
			TestName:   fmt.Sprintf("%s: address is not all lowercase or all uppercase", field),
			FieldValue: "Zig12z7zhe2d03wn03e8lkw8At6zjxa7y65q5xnp5t",
			ErrorText: fmt.Sprintf(
				"%s address (Zig12z7zhe2d03wn03e8lkw8At6zjxa7y65q5xnp5t)",
				field,
			),
		},
		{
			TestName:   fmt.Sprintf("%s: address contains invalid characters", field),
			FieldValue: "zi!12z7zhe2d03wn03e8lkw8at6zjxa7y65q5xnp5%",
			ErrorText: fmt.Sprintf(
				"%s address (zi!12z7zhe2d03wn03e8lkw8at6zjxa7y65q5xnp5%%)",
				field,
			),
		},
		{
			TestName:   fmt.Sprintf("%s: address contains only numbers", field),
			FieldValue: "123121712321032203281238456456771651512351",
			ErrorText: fmt.Sprintf(
				"%s address (123121712321032203281238456456771651512351)",
				field,
			),
		},
		{
			TestName:   fmt.Sprintf("%s: address contains space in string", field),
			FieldValue: "abcdefghijabcdefghijabcde ghijabcdefghijabcdefghij",
			ErrorText: fmt.Sprintf(
				"%s address (abcdefghijabcdefghijabcde ghijabcdefghijabcdefghij)",
				field,
			),
		},
		{
			TestName:   fmt.Sprintf("%s: address has bad prefix", field),
			FieldValue: "yoy193fxruxcm8y32njzt23c86en7hd8tajma79tt3",
			ErrorText: fmt.Sprintf(
				"%s address (yoy193fxruxcm8y32njzt23c86en7hd8tajma79tt3)",
				field,
			),
		},
	}
}
