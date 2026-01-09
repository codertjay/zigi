package types

import (
	"fmt"
	"zigchain/zutils/validators"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DefaultIndex is the default global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		PoolList:     []Pool{},
		PoolsMeta:    nil,
		PoolUidsList: []PoolUids{},
		// this line is used by starport scaffolding # genesis/types/default
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Check for duplicated index in pool
	poolIndexMap := make(map[string]struct{})

	for _, pool := range gs.PoolList {
		index := string(PoolKey(pool.PoolId))
		if _, ok := poolIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for pool")
		}
		poolIndexMap[index] = struct{}{}

		if err := pool.Validate(); err != nil {
			return err
		}

	}

	// check if gs.PoolsMeta.NextPoolId exits and if it's equal to the number of pools in the PoolList
	if gs.PoolsMeta != nil {
		// #nosec G115 - let is int32, and we are comparing it with uint64
		// given that this is config value, probably never used, but if used it's not expected to be large, it's safe to ignore
		expectedNextPoolId := uint64(len(gs.PoolList) + 1)
		if gs.PoolsMeta.NextPoolId != expectedNextPoolId {
			return fmt.Errorf(
				"NextPoolId: (%d) must be equal to the number of pools in the PoolList (%d + 1 = %d total) ",
				gs.PoolsMeta.NextPoolId,
				expectedNextPoolId-1,
				expectedNextPoolId,
			)
		}
	} else {
		if len(gs.PoolList) > 0 {
			return fmt.Errorf("PoolsMeta is nil but PoolList is not empty")
		}
	}

	// Check for duplicated index in poolUids
	poolUidsIndexMap := make(map[string]struct{})

	for _, poolUids := range gs.PoolUidsList {
		index := string(PoolUidsKey(poolUids.PoolUid))
		if _, ok := poolUidsIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for poolUids")
		}
		poolUidsIndexMap[index] = struct{}{}

		if err := validators.CheckPoolId(poolUids.PoolId); err != nil {
			return err
		}

	}

	// Validate that every pool has a corresponding PoolUid
	for _, pool := range gs.PoolList {
		poolUid := PoolUidsKey(GetPoolUidString(pool))
		if _, exists := poolUidsIndexMap[string(poolUid)]; !exists {
			return fmt.Errorf("missing PoolUidString '%s' in PoolUidsList", string(poolUid))
		}
	}

	// Validate that every PoolUid corresponds to a Pool
	for _, poolUids := range gs.PoolUidsList {
		poolUid := PoolKey(poolUids.PoolId)
		if _, exists := poolIndexMap[string(poolUid)]; !exists {
			return fmt.Errorf("missing PoolId '%s' from PoolUidsList in PoolList", poolUids.PoolId)
		}
	}
	// this line is used by starport scaffolding # genesis/types/validate

	return gs.Params.Validate()
}

// ValidatePool checks if a pool is valid
func (p Pool) Validate() error {

	//	LpToken types.Coin   `protobuf:"bytes,2,opt,name=lpToken,proto3" json:"lpToken"`
	//	Creator string       `protobuf:"bytes,3,opt,name=creator,proto3" json:"creator,omitempty"`
	//	Fee     uint32       `protobuf:"varint,4,opt,name=fee,proto3" json:"fee,omitempty"`
	//	Formula string       `protobuf:"bytes,5,opt,name=formula,proto3" json:"formula,omitempty"`
	//	Coins   []types.Coin `protobuf:"bytes,6,rep,name=coins,proto3" json:"coins"`
	//	Address string       `protobuf:"bytes,7,opt,name=address,proto3" json:"address,omitempty"`

	if err := validators.CheckPoolId(p.PoolId); err != nil {
		return err
	}

	if err := validators.AddressCheck("creator", p.Creator); err != nil {
		return err
	}

	if err := validateNewPoolFeePct(p.Fee); err != nil {
		return err
	}

	if !IsValidFormula(p.Formula) {
		return fmt.Errorf("invalid formula: %s", p.Formula)
	}

	if len(p.Coins) < 2 {
		return fmt.Errorf("pool must have at least 2 coins")
	}
	for _, coin := range p.Coins {
		if err := validators.CoinCheck(coin, false); err != nil {
			return err
		}
	}

	if err := validators.CoinCheck(p.LpToken, false); err != nil {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidCoins,
			"lpToken: %s",
			err,
		)
	}

	if err := validators.AddressCheck("address", p.Address); err != nil {
		return err
	}

	// Check that the address equals the expected module address generated from the pool ID
	expectedAddress := GetPoolAddress(p.PoolId).String()
	if p.Address != expectedAddress {
		return errorsmod.Wrapf(
			ErrInvalidPoolAddress,
			"address '%s' does not match expected module address '%s' for pool ID '%s'",
			p.Address,
			expectedAddress,
			p.PoolId,
		)
	}

	return nil

}
