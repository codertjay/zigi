package migrations

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (m Migrator) V2Migration(ctx sdk.Context) error {
	// set new ibc settings fields based on zig-test-2 <> Axelar testnet IBC connections
	m.keeper.SetNativeClientId(ctx, "07-tendermint-0")
	m.keeper.SetCounterpartyClientId(ctx, "07-tendermint-1163")
	m.keeper.SetNativePort(ctx, "transfer")
	m.keeper.SetCounterpartyPort(ctx, "transfer")
	m.keeper.SetNativeChannel(ctx, "channel-0")
	m.keeper.SetCounterpartyChannel(ctx, "channel-612")
	m.keeper.SetDenom(ctx, "uaxl")

	if err := m.keeper.SetDecimalDifference(ctx, 0); err != nil {
		return err
	}

	// initialize pauser addresses
	m.keeper.SetPauserAddresses(ctx, []string{})

	return nil
}
