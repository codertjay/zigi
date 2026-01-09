package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParams_Validate_Positive(t *testing.T) {
	params := DefaultParams()

	err := params.Validate()
	require.NoError(t, err, "expected default params to pass validation")
}
