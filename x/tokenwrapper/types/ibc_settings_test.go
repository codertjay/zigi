package types

import (
	"testing"

	"zigchain/testutil/sample"

	"github.com/stretchr/testify/require"
)

func TestDefaultOperatorAddress(t *testing.T) {
	require.Equal(t, sample.ZeroAccAddress(), DefaultOperatorAddress())
}

func TestDefaultEnabled(t *testing.T) {
	require.Equal(t, false, DefaultEnabled())
}

func TestDefaultNativeClientId(t *testing.T) {
	require.Equal(t, "07-tendermint-0", DefaultNativeClientId())
}

func TestDefaultCounterpartyClientId(t *testing.T) {
	require.Equal(t, "07-tendermint-0", DefaultCounterpartyClientId())
}

func TestDefaultNativePort(t *testing.T) {
	require.Equal(t, "transfer", DefaultNativePort())
}

func TestDefaultCounterpartyPort(t *testing.T) {
	require.Equal(t, "transfer", DefaultCounterpartyPort())
}

func TestDefaultNativeChannel(t *testing.T) {
	require.Equal(t, "channel-0", DefaultNativeChannel())
}

func TestDefaultCounterpartyChannel(t *testing.T) {
	require.Equal(t, "channel-0", DefaultCounterpartyChannel())
}

func TestDefaultDenom(t *testing.T) {
	require.Equal(t, "uzig", DefaultDenom())
}

func TestDefaultDecimalDifference(t *testing.T) {
	require.Equal(t, uint32(0), DefaultDecimalDifference())
}
