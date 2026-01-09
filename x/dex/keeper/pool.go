package keeper

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"strconv"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	errorsmod "cosmossdk.io/errors"

	"zigchain/x/dex/types"
	"zigchain/zutils/constants"
)

// SetPool set a specific pool in the store from its index
func (k Keeper) SetPool(ctx context.Context, pool types.Pool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.PoolKeyPrefix))
	b := k.cdc.MustMarshal(&pool)
	store.Set(types.PoolKey(
		pool.PoolId,
	), b)
}

// GetPool returns a pool from its index
func (k Keeper) GetPool(
	ctx context.Context,
	poolIDString string,

) (val types.Pool, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.PoolKeyPrefix))

	b := store.Get(types.PoolKey(
		poolIDString,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// GetAllPool returns all pool
func (k Keeper) GetAllPool(ctx context.Context) (list []types.Pool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.PoolKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer func(iterator storetypes.Iterator) {
		err := iterator.Close()
		if err != nil {
			k.logger.Error("failed to close iterator", "error", err)
		}
	}(iterator)

	for ; iterator.Valid(); iterator.Next() {
		var val types.Pool
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// GetNextPoolId returns the next pool id.
func (k Keeper) GetNextPoolID(ctx context.Context) (val uint64) {
	meta, found := k.GetPoolsMeta(ctx)

	if !found {
		return 1
	}

	return meta.NextPoolId
}

func (k Keeper) GetNextPoolIDString(ctx context.Context) string {
	poolId := k.GetNextPoolID(ctx)
	poolIDString := constants.PoolPrefix + strconv.FormatUint(poolId, 10)
	return poolIDString
}

// SetNextPoolID sets the next pool ID.
func (k Keeper) SetNextPoolID(ctx context.Context) {

	count := k.GetNextPoolID(ctx)
	if count == math.MaxUint64 {
		// Handle the overflow case
		panic("overflow on pool id")
	}

	k.SetPoolsMeta(ctx, types.PoolsMeta{
		NextPoolId: count + 1,
	},
	)
}

// GetAndSetNextPoolID returns the next pool id number and increments the pool id.
func (k Keeper) GetAndSetNextPoolID(ctx context.Context) uint64 {
	poolID := k.GetNextPoolID(ctx)
	k.SetNextPoolID(ctx)
	return poolID
}

// GetAndSetNextPoolIDString returns the next pool id with string prefix string and increments the pool id.
func (k Keeper) GetAndSetNextPoolIDString(ctx context.Context) string {
	poolId := k.GetNextPoolID(ctx)
	poolIDString := constants.PoolPrefix + strconv.FormatUint(poolId, 10)
	k.SetNextPoolID(ctx)
	return poolIDString
}

// CanOverwriteAccountTypes is a map of extra account types that can be overridden.
// This is defined as a global variable, so it can be modified in the chain's app.go and used here without
// having to import the chain.
var CanOverwriteAccountTypes map[reflect.Type]struct{}

// CanCreateModuleAccountAtAddr tells us if we can safely make a module account at
// a given address. By collision resistance of the address (given API safe construction),
// the only way for an account to be already be at this address is if its claimed by the same
// pre-image from the correct module,
// or some SDK command breaks assumptions and creates an account at designated address.
// This function checks if there is an account at that address, and runs some safety checks
// to be extra-sure its not a user account (e.g. non-zero sequence, pubkey, of fore-seen account types).
// If there is no account, or if we believe its not a user-spendable account, we allow module account
// creation at the address.
// else, we do not.
//
// TODO: This is generally from an SDK design flaw
// code based off wasmd code: https://github.com/CosmWasm/wasmd/pull/996
// Its _mandatory_ that the caller do the API safe construction to generate a module account addr,
// namely, address.Module(ModuleName, {key})
// CanCreateModuleAccountAtAddr checks if a module account can be created at the given address.
// Returns nil if creation is allowed, or an error if not.
// CanCreateModuleAccountAtAddr checks if a module account can be created at the given address.
// Returns nil if creation is allowed, or an error if not.
func (k Keeper) CanCreateModuleAccountAtAddr(ctx sdk.Context, addr sdk.AccAddress) error {

	// Get an account at the given address
	existingAcct := k.accountKeeper.GetAccount(ctx, addr)
	// If no account exists, creation is allowed
	if existingAcct == nil {
		return nil
	}

	// Do not allow creation if an account has sent transactions or has a public key
	// If a sequence is not zero or public key is not nil, it indicates that the account has sent transactions
	if existingAcct.GetSequence() != 0 || existingAcct.GetPubKey() != nil {
		return fmt.Errorf("cannot create module account at %s: account has sent transactions", existingAcct.GetAddress())
	}

	// Define overridable account types (that were not yet used)
	overridableTypes := map[reflect.Type]struct{}{
		reflect.TypeOf(&authtypes.BaseAccount{}):                 {},
		reflect.TypeOf(&vestingtypes.BaseVestingAccount{}):       {},
		reflect.TypeOf(&vestingtypes.ContinuousVestingAccount{}): {},
		reflect.TypeOf(&vestingtypes.DelayedVestingAccount{}):    {},
		reflect.TypeOf(&vestingtypes.PeriodicVestingAccount{}):   {},
		reflect.TypeOf(&vestingtypes.PermanentLockedAccount{}):   {},
	}

	// Add extra overridable types from CanOverwriteAccountTypes
	for typ := range CanOverwriteAccountTypes {
		overridableTypes[typ] = struct{}{}
	}

	// Allow creation if an account type is overridable
	if _, isOverridable := overridableTypes[reflect.TypeOf(existingAcct)]; isOverridable {
		return nil
	}

	// Block creation for non-overridable account types, include an account type in error
	return fmt.Errorf("cannot create module account at %s: existing account of type %T is not an overridable type",
		existingAcct.GetAddress(), reflect.TypeOf(existingAcct))
}

// CreatePoolAccount creates a pool account (and saves it to store) given pool ID String
func (k Keeper) CreatePoolAccount(ctx sdk.Context, poolIDString string) (sdk.AccAddress, error) {

	// get address from the pool ID string like zp1, zp2, zp3, ...
	addr := types.GetPoolAddress(poolIDString)

	// Check if the pool account already exists, accounts can exist in other realms wasm, staking etc
	if err := k.CanCreateModuleAccountAtAddr(ctx, addr); err != nil {
		return nil, errorsmod.Wrapf(
			err,
			"CreatePoolAccount: Failed to create module account at address %s",
			addr.String(),
		)
	}
	// generate the pool account (in memory only)
	poolAccount := k.accountKeeper.NewAccount(
		ctx,
		// From a module account (adds name and permissions)
		authtypes.NewModuleAccount(
			// from the base account provided by the address (has only address)
			authtypes.NewBaseAccountWithAddress(addr),
			addr.String(),
		),
	)
	// Save the account to the store
	k.accountKeeper.SetAccount(ctx, poolAccount)
	return poolAccount.GetAddress(), nil
}

// SendFromAddressToPool deposits the specified coins from sender the pool account
func (k Keeper) SendFromAddressToPool(ctx sdk.Context, sender sdk.AccAddress, poolIDString string, coins sdk.Coins) error {

	// Check if coins are empty
	if coins.Empty() {
		return errorsmod.Wrapf(
			errors.ErrInvalidRequest,
			"SendFromAddressToPool: coins cannot be empty",
		)
	}

	// Get address from the pool ID string like zp1, zp2, zp3, ...
	poolAddress := types.GetPoolAddress(poolIDString)

	// Check if the sender has enough balances for each coin
	for _, coin := range coins {
		// Check if the coin is valid
		if coin.IsNegative() {
			return errorsmod.Wrapf(
				errors.ErrInvalidRequest,
				"SendFromAddressToPool: coins cannot be negative",
			)
		}
		if coin.IsZero() {
			return errorsmod.Wrapf(
				errors.ErrInvalidRequest,
				"SendFromAddressToPool: coins cannot be zero",
			)
		}
		// Check if the sender has enough coins
		if !k.bankKeeper.HasBalance(ctx, sender, coin) {
			return errorsmod.Wrapf(
				errors.ErrInsufficientFunds,
				"SendFromAddressToPool: Insufficient funds in sender %s to send %s coins to poolID %s (address: %s)",
				sender.String(),
				coin.String(),
				poolIDString,
				poolAddress,
			)
		}
	}

	// Check if the pool account exists
	poolAccount := k.accountKeeper.GetAccount(ctx, poolAddress)
	if poolAccount == nil {
		return errorsmod.Wrapf(
			types.ErrPoolAccountNotFound,
			"Pool ID: %s account does not exist",
			poolIDString,
		)
	}

	// Send coins from a sender to a pool account
	err := k.bankKeeper.SendCoins(ctx, sender, poolAddress, coins)
	// Check if the sending was successful
	if err != nil {
		return errorsmod.Wrapf(
			err,
			"SendFromAddressToPool: Failed to send %s coins from sender %s to poolID %s (address: %s)",
			coins.String(),
			sender.String(),
			poolIDString,
			poolAccount.String(),
		)
	}

	// On success, return nil
	return nil
}

// SendFromPoolToAddress withdraws the specified coins from the pool account to receiver
func (k Keeper) SendFromPoolToAddress(ctx sdk.Context, poolIDString string, receiver sdk.AccAddress, coins sdk.Coins) error {

	if coins.Empty() {
		return errorsmod.Wrapf(
			errors.ErrInvalidRequest,
			"SendFromPoolToAddress: No coins to send from poolID %s (address: %s) to receiver %s",
			poolIDString,
			types.GetPoolAddress(poolIDString).String(),
			receiver.String(),
		)
	}

	poolAddress := types.GetPoolAddress(poolIDString)

	// Check if the pool account exists
	poolAccount := k.accountKeeper.GetAccount(ctx, poolAddress)
	if poolAccount == nil {
		return errorsmod.Wrapf(
			types.ErrPoolAccountNotFound,
			"Pool ID: %s account already does not exists",
			poolIDString,
		)
	}

	// Check if the pool account has enough balances
	for _, coin := range coins {
		if !k.bankKeeper.HasBalance(ctx, poolAddress, coin) {
			return errorsmod.Wrapf(
				errors.ErrInsufficientFunds,
				"SendFromPoolToAddress: Insufficient funds in poolID %s (address: %s) to send %s coins to receiver %s",
				poolIDString,
				poolAddress,
				coin.String(),
				receiver.String(),
			)
		}
	}

	err := k.bankKeeper.SendCoins(ctx, poolAddress, receiver, coins)
	if err != nil {
		return errorsmod.Wrapf(
			err,
			"SendFromPoolToAddress: Failed to send %s coins from poolID %s (address: %s) to receiver %s",
			coins.String(),
			poolIDString,
			poolAddress,
			receiver.String(),
		)
	}

	return nil
}
