package types

import "zigchain/testutil/sample"

func DefaultOperatorAddress() string {
	return sample.ZeroAccAddress()
}

func DefaultPauserAddresses() []string {
	return []string{}
}

func DefaultEnabled() bool {
	return false
}

func DefaultNativeClientId() string {
	return "07-tendermint-0"
}

func DefaultCounterpartyClientId() string {
	return "07-tendermint-0"
}

func DefaultNativePort() string {
	return "transfer"
}

func DefaultCounterpartyPort() string {
	return "transfer"
}

func DefaultNativeChannel() string {
	return "channel-0"
}

func DefaultCounterpartyChannel() string {
	return "channel-0"
}

func DefaultDenom() string {
	return "uzig"
}

func DefaultDecimalDifference() uint32 {
	return 0
}
