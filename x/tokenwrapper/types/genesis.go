package types

import (
	"fmt"
	"zigchain/zutils/validators"

	"cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// this line is used by starport scaffolding # genesis/types/import

// DefaultIndex is the default global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		// this line is used by starport scaffolding # genesis/types/default
		Params:                  DefaultParams(),
		TotalTransferredIn:      math.ZeroInt(),
		TotalTransferredOut:     math.ZeroInt(),
		OperatorAddress:         DefaultOperatorAddress(),
		ProposedOperatorAddress: DefaultOperatorAddress(),
		PauserAddresses:         DefaultPauserAddresses(),
		Enabled:                 DefaultEnabled(),
		NativeClientId:          DefaultNativeClientId(),
		CounterpartyClientId:    DefaultCounterpartyClientId(),
		NativePort:              DefaultNativePort(),
		CounterpartyPort:        DefaultCounterpartyPort(),
		NativeChannel:           DefaultNativeChannel(),
		CounterpartyChannel:     DefaultCounterpartyChannel(),
		Denom:                   DefaultDenom(),
		DecimalDifference:       DefaultDecimalDifference(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Validate Params
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	// Validate TotalTransferredIn and TotalTransferredOut are not negative
	if gs.TotalTransferredIn.IsNegative() {
		return sdkerrors.ErrInvalidCoins.Wrapf("total transferred in cannot be negative: %s", gs.TotalTransferredIn.String())
	}
	if gs.TotalTransferredOut.IsNegative() {
		return sdkerrors.ErrInvalidCoins.Wrapf("total transferred out cannot be negative: %s", gs.TotalTransferredOut.String())
	}

	// Validate OperatorAddress
	if err := validators.AddressCheck("operator", gs.OperatorAddress); err != nil {
		return err
	}
	if err := validators.AddressCheck("proposed_operator", gs.ProposedOperatorAddress); err != nil {
		return err
	}

	// Validate PauserAddresses
	for i, addr := range gs.PauserAddresses {
		if err := validators.AddressCheck(fmt.Sprintf("pauser[%d]", i), addr); err != nil {
			return err
		}
	}

	// Validate Client IDs
	if err := validators.ValidateClientId(gs.NativeClientId); err != nil {
		return err
	}
	if err := validators.ValidateClientId(gs.CounterpartyClientId); err != nil {
		return err
	}

	// Validate Ports
	if err := validators.ValidatePort(gs.NativePort); err != nil {
		return err
	}
	if err := validators.ValidatePort(gs.CounterpartyPort); err != nil {
		return err
	}

	// Validate Channels
	if err := validators.ValidateChannel(gs.NativeChannel); err != nil {
		return err
	}
	if err := validators.ValidateChannel(gs.CounterpartyChannel); err != nil {
		return err
	}

	// Validate Denom
	if err := validators.ValidateDenom(gs.Denom); err != nil {
		return err
	}

	// Validate DecimalDifference
	if err := validators.ValidateDecimalDifference(gs.DecimalDifference); err != nil {
		return err
	}

	return nil
}
