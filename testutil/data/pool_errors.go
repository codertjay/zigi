package errorPacks

import (
	"fmt"
	"strings"

	"zigchain/zutils/constants"
)

// longPoolId - example of PoolId with too long name
var longPoolId = "zp" + strings.Repeat("1", 100)

// InvalidPoolId - poolId name tests
var InvalidPoolId = []struct {
	// Name/description of the test case
	TestName string
	// Field value
	FieldValue string
	// expected error message for the test case
	ErrorText string
}{
	{
		TestName:   "empty poolId",
		FieldValue: "",
		ErrorText:  "Invalid pool id: pool id is empty",
	},
	{
		TestName:   "invalid poolId format: only numbers",
		FieldValue: "123",
		ErrorText:  "Invalid pool id: '123', pool id has to start with 'zp' followed by numbers e.g. zp123",
	},
	{
		TestName:   "invalid poolId format: only zp, missing numbers",
		FieldValue: "zp",
		ErrorText:  "Invalid pool id: 'zp' pool id is too short, minimum 3 characters",
	},
	{
		TestName:   "invalid poolId format: zp and letters",
		FieldValue: "zpdagada",
		ErrorText:  "Invalid pool id: 'zpdagada', pool id has to start with 'zp' followed by numbers e.g. zp123",
	},
	{
		TestName:   "invalid poolId format: wrong starting letters",
		FieldValue: "nk123",
		ErrorText:  "Invalid pool id: 'nk123', pool id has to start with 'zp' followed by numbers e.g. zp123",
	},
	{
		TestName:   "invalid poolId format: uppercase ZP",
		FieldValue: "ZP123",
		ErrorText:  "Invalid pool id: 'ZP123', pool id has to start with 'zp' followed by numbers e.g. zp123",
	},
	{
		TestName:   "invalid poolId format: space between letters and numbers",
		FieldValue: "zp 123",
		ErrorText:  "Invalid pool id: 'zp 123', pool id has to start with 'zp' followed by numbers e.g. zp123",
	},
	{
		TestName:   "invalid poolId format: too long",
		FieldValue: longPoolId,
		ErrorText: fmt.Sprintf("Invalid pool id: '%s' pool id is too long (102), maximum %d characters",
			longPoolId,
			constants.MaxSubDenomLength,
		),
	},
}
