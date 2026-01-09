package keeper

import (
	"fmt"
	"math/big"
	"strconv"

	"zigchain/x/tokenwrapper/types"
	"zigchain/zutils/constants"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService store.KVStoreService
		logger       log.Logger

		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string

		bankKeeper types.BankKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	logger log.Logger,
	authority string,

	bankKeeper types.BankKeeper,
) Keeper {
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	return Keeper{
		cdc:          cdc,
		storeService: storeService,
		authority:    authority,
		logger:       logger,

		bankKeeper: bankKeeper,
	}
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger() log.Logger {
	return k.logger.With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// CheckModuleBalance checks if the module has enough amount of the given coins
func (k Keeper) CheckModuleBalance(ctx sdk.Context, coins sdk.Coins) error {
	for _, coin := range coins {
		if !k.bankKeeper.HasBalance(ctx, authtypes.NewModuleAddress(types.ModuleName), coin) {
			return fmt.Errorf("module does not have enough balance of %s", coin.String())
		}
	}
	return nil
}

// CheckAccountBalance checks if the account has enough amount of the given coins
func (k Keeper) CheckAccountBalance(ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) error {
	for _, coin := range coins {
		if !k.bankKeeper.HasBalance(ctx, addr, coin) {
			return fmt.Errorf("account %s does not have enough balance of %s", addr.String(), coin.String())
		}
	}
	return nil
}

// LockTokens locks the specified amount of tokens for the given address.
func (k Keeper) LockTokens(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) error {
	if err := k.CheckAccountBalance(ctx, addr, amt); err != nil {
		return err
	}

	// Send coins from account to module account
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, amt); err != nil {
		return err
	}

	return nil
}

// UnlockTokens unlocks the specified amount of tokens for the given address.
func (k Keeper) UnlockTokens(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) error {
	if err := k.CheckModuleBalance(ctx, amt); err != nil {
		return err
	}

	// Send coins from module account to account
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, amt); err != nil {
		return err
	}

	return nil
}

// BurnIbcTokens burns the specified amount of IBC tokens
func (k Keeper) BurnIbcTokens(ctx sdk.Context, amt sdk.Coins) error {
	// Check if the module has enough balance to burn
	if err := k.CheckModuleBalance(ctx, amt); err != nil {
		return err
	}

	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, amt); err != nil {
		return err
	}

	return nil
}

// GetModuleWalletBalances returns the balances of the module wallet
func (k Keeper) GetModuleWalletBalances(ctx sdk.Context) (string, sdk.Coins) {
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	balances := k.bankKeeper.GetAllBalances(ctx, moduleAddr)
	return moduleAddr.String(), balances
}

// GetTotalTransferredIn returns the total amount of ZIG tokens transferred in
func (k Keeper) GetTotalTransferredIn(ctx sdk.Context) sdkmath.Int {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.TotalTransferredInKey)
	if bz == nil {
		return sdkmath.ZeroInt()
	}
	return sdkmath.NewIntFromBigInt(new(big.Int).SetBytes(bz))
}

// SetTotalTransferredIn sets the total amount of ZIG tokens transferred in
func (k Keeper) SetTotalTransferredIn(ctx sdk.Context, amount sdkmath.Int) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store.Set(types.TotalTransferredInKey, amount.BigInt().Bytes())
}

// AddToTotalTransferredIn adds an amount to the total ZIG tokens transferred in
func (k Keeper) AddToTotalTransferredIn(ctx sdk.Context, amount sdkmath.Int) {
	current := k.GetTotalTransferredIn(ctx)
	k.SetTotalTransferredIn(ctx, current.Add(amount))
}

// GetTotalTransferredOut returns the total amount of ZIG tokens transferred out
func (k Keeper) GetTotalTransferredOut(ctx sdk.Context) sdkmath.Int {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.TotalTransferredOutKey)
	if bz == nil {
		return sdkmath.ZeroInt()
	}
	return sdkmath.NewIntFromBigInt(new(big.Int).SetBytes(bz))
}

// SetTotalTransferredOut sets the total amount of ZIG tokens transferred out
func (k Keeper) SetTotalTransferredOut(ctx sdk.Context, amount sdkmath.Int) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store.Set(types.TotalTransferredOutKey, amount.BigInt().Bytes())
}

// AddToTotalTransferredOut adds an amount to the total ZIG tokens transferred out
func (k Keeper) AddToTotalTransferredOut(ctx sdk.Context, amount sdkmath.Int) {
	current := k.GetTotalTransferredOut(ctx)
	k.SetTotalTransferredOut(ctx, current.Add(amount))
}

// GetOperatorAddress returns the current operator address
func (k Keeper) GetOperatorAddress(ctx sdk.Context) string {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.OperatorAddressKey)
	if bz == nil {
		return ""
	}
	return string(bz)
}

// SetOperatorAddress sets the operator address
func (k Keeper) SetOperatorAddress(ctx sdk.Context, address string) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store.Set(types.OperatorAddressKey, []byte(address))
}

// IsEnabled returns whether the token wrapper is enabled
func (k Keeper) IsEnabled(ctx sdk.Context) bool {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.EnabledKey)
	if bz == nil {
		return true // enabled by default
	}
	return bz[0] == 1
}

// SetEnabled sets whether the token wrapper is enabled
func (k Keeper) SetEnabled(ctx sdk.Context, enabled bool) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	if enabled {
		store.Set(types.EnabledKey, []byte{1})
	} else {
		store.Set(types.EnabledKey, []byte{0})
	}
}

// GetNativeClientId returns the native client ID for IBC transfers
func (k Keeper) GetNativeClientId(ctx sdk.Context) string {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.NativeClientIdKey)
	if bz == nil {
		return ""
	}
	return string(bz)
}

// SetNativeClientId sets the native client ID for IBC transfers
func (k Keeper) SetNativeClientId(ctx sdk.Context, clientId string) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store.Set(types.NativeClientIdKey, []byte(clientId))
}

// GetCounterpartyClientId returns the counterparty client ID for IBC transfers
func (k Keeper) GetCounterpartyClientId(ctx sdk.Context) string {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.CounterpartyClientIdKey)
	if bz == nil {
		return ""
	}
	return string(bz)
}

// SetCounterpartyClientId sets the counterparty client ID for IBC transfers
func (k Keeper) SetCounterpartyClientId(ctx sdk.Context, clientId string) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store.Set(types.CounterpartyClientIdKey, []byte(clientId))
}

// GetNativePort returns the native port for IBC transfers
func (k Keeper) GetNativePort(ctx sdk.Context) string {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.NativePortKey)
	if bz == nil {
		return ""
	}
	return string(bz)
}

// SetNativePort sets the native port for IBC transfers
func (k Keeper) SetNativePort(ctx sdk.Context, port string) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store.Set(types.NativePortKey, []byte(port))
}

// GetCounterpartyPort returns the counterparty port for IBC transfers
func (k Keeper) GetCounterpartyPort(ctx sdk.Context) string {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.CounterpartyPortKey)
	if bz == nil {
		return ""
	}
	return string(bz)
}

// SetCounterpartyPort sets the counterparty port for IBC transfers
func (k Keeper) SetCounterpartyPort(ctx sdk.Context, port string) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store.Set(types.CounterpartyPortKey, []byte(port))
}

// GetNativeChannel returns the native channel for IBC transfers
func (k Keeper) GetNativeChannel(ctx sdk.Context) string {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.NativeChannelKey)
	if bz == nil {
		return ""
	}
	return string(bz)
}

// SetNativeChannel sets the native channel for IBC transfers
func (k Keeper) SetNativeChannel(ctx sdk.Context, channel string) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store.Set(types.NativeChannelKey, []byte(channel))
}

// GetCounterpartyChannel returns the counterparty channel for IBC transfers
func (k Keeper) GetCounterpartyChannel(ctx sdk.Context) string {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.CounterpartyChannelKey)
	if bz == nil {
		return ""
	}
	return string(bz)
}

// SetCounterpartyChannel sets the counterparty channel for IBC transfers
func (k Keeper) SetCounterpartyChannel(ctx sdk.Context, channel string) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store.Set(types.CounterpartyChannelKey, []byte(channel))
}

// GetDenom returns the denom for wrapped tokens
func (k Keeper) GetDenom(ctx sdk.Context) string {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.DenomKey)
	if bz == nil {
		return ""
	}
	return string(bz)
}

// SetDenom sets the denom for wrapped tokens
func (k Keeper) SetDenom(ctx sdk.Context, denom string) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store.Set(types.DenomKey, []byte(denom))
}

// GetDecimalDifference returns the decimal difference for IBC transfers
func (k Keeper) GetDecimalDifference(ctx sdk.Context) uint32 {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.DecimalDifferenceKey)
	if bz == nil {
		return 0
	}
	value, err := strconv.ParseUint(string(bz), 10, 32)
	if err != nil {
		k.Logger().Error("failed to parse decimal difference", "error", err)
		return 0
	}
	return uint32(value)
}

// SetDecimalDifference sets the decimal difference for IBC transfers
func (k Keeper) SetDecimalDifference(ctx sdk.Context, decimalDifference uint32) error {
	if decimalDifference > 18 {
		return fmt.Errorf("decimal difference must be between 0 and 18, got %d", decimalDifference)
	}
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store.Set(types.DecimalDifferenceKey, []byte(strconv.FormatUint(uint64(decimalDifference), 10)))
	return nil
}

// GetDecimalConversionFactor returns the conversion factor based on the decimal difference
func (k Keeper) GetDecimalConversionFactor(ctx sdk.Context) sdkmath.Int {
	decimalDifference := k.GetDecimalDifference(ctx)
	return sdkmath.NewIntFromBigInt(big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(decimalDifference)), nil))
}

// GetPauserAddresses returns the list of pauser addresses
func (k Keeper) GetPauserAddresses(ctx sdk.Context) []string {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.PauserAddressesKey)
	if bz == nil {
		return []string{}
	}
	var pauserAddresses types.PauserAddresses
	k.cdc.MustUnmarshal(bz, &pauserAddresses)
	if pauserAddresses.Addresses == nil {
		return []string{}
	}
	return pauserAddresses.Addresses
}

// SetPauserAddresses sets the list of pauser addresses
func (k Keeper) SetPauserAddresses(ctx sdk.Context, addresses []string) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	pauserAddresses := types.PauserAddresses{Addresses: addresses}
	bz := k.cdc.MustMarshal(&pauserAddresses)
	store.Set(types.PauserAddressesKey, bz)
}

// AddPauserAddress adds a new pauser address to the list
func (k Keeper) AddPauserAddress(ctx sdk.Context, address string) {
	addresses := k.GetPauserAddresses(ctx)
	// Check if the address already exists
	for _, addr := range addresses {
		if addr == address {
			return
		}
	}
	addresses = append(addresses, address)
	k.SetPauserAddresses(ctx, addresses)
}

// RemovePauserAddress removes a pauser address from the list
func (k Keeper) RemovePauserAddress(ctx sdk.Context, address string) {
	addresses := k.GetPauserAddresses(ctx)
	for i, addr := range addresses {
		if addr == address {
			addresses = append(addresses[:i], addresses[i+1:]...)
			break
		}
	}
	k.SetPauserAddresses(ctx, addresses)
}

// IsPauserAddress checks if an address is a pauser
func (k Keeper) IsPauserAddress(ctx sdk.Context, address string) bool {
	addresses := k.GetPauserAddresses(ctx)
	for _, addr := range addresses {
		if addr == address {
			return true
		}
	}
	return false
}

// GetProposedOperatorAddress returns the proposed operator address
func (k Keeper) GetProposedOperatorAddress(ctx sdk.Context) string {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.ProposedOperatorAddressKey)
	if bz == nil {
		return ""
	}
	return string(bz)
}

// SetProposedOperatorAddress sets the proposed operator address
func (k Keeper) SetProposedOperatorAddress(ctx sdk.Context, address string) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store.Set(types.ProposedOperatorAddressKey, []byte(address))
}

// GetIBCRecvDenom gets the IBC denom for the received token
func (k Keeper) GetIBCRecvDenom(ctx sdk.Context, denom string) string {
	// Get module IBC settings
	moduleNativePort := k.GetNativePort(ctx)
	moduleNativeChannel := k.GetNativeChannel(ctx)

	// Prepare denom for IBC token
	sourcePrefix := transfertypes.NewHop(moduleNativePort, moduleNativeChannel)
	prefixedDenom := fmt.Sprintf("%s/%s", sourcePrefix.String(), denom)
	return transfertypes.ExtractDenomFromPath(prefixedDenom).IBCDenom()
}

// ScaleDownTokenPrecision scales down the token precision from 18 to 6 decimals
func (k Keeper) ScaleDownTokenPrecision(ctx sdk.Context, amount sdkmath.Int) (sdkmath.Int, error) {
	// Convert from 18 decimals to 6 decimals
	// Divide by conversion factor to convert from 18 decimals to 6 decimals
	// Note: Quo performs integer division which rounds down
	conversionFactor := k.GetDecimalConversionFactor(ctx)
	convertedAmount := amount.Quo(conversionFactor)

	// Check if the converted amount is zero or negative
	if convertedAmount.IsZero() || convertedAmount.IsNegative() {
		return sdkmath.Int{}, fmt.Errorf("converted amount is zero or negative: %s", convertedAmount.String())
	}

	return convertedAmount, nil
}

// ScaleUpTokenPrecision scales up the token precision from 6 to 18 decimals
func (k Keeper) ScaleUpTokenPrecision(ctx sdk.Context, amount sdkmath.Int) (sdkmath.Int, error) {
	// Convert from 6 decimals to 18 decimals
	// Multiply by conversion factor to convert from 6 decimals to 18 decimals
	conversionFactor := k.GetDecimalConversionFactor(ctx)
	convertedAmount := amount.Mul(conversionFactor)

	// Check if the converted amount is zero or negative
	if convertedAmount.IsZero() || convertedAmount.IsNegative() {
		return sdkmath.Int{}, fmt.Errorf("converted amount is zero or negative: %s", convertedAmount.String())
	}

	return convertedAmount, nil
}

// CheckBalances checks if the account and module have enough balance
func (k Keeper) CheckBalances(ctx sdk.Context, receiver sdk.AccAddress, amount sdkmath.Int, recvDenom string, convertedAmount sdkmath.Int) error {
	ibcCoins := sdk.NewCoins(sdk.NewCoin(recvDenom, amount))
	if err := k.CheckAccountBalance(ctx, receiver, ibcCoins); err != nil {
		return err
	}
	nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, convertedAmount))
	if err := k.CheckModuleBalance(ctx, nativeCoins); err != nil {
		return err
	}
	return nil
}

// LockIBCTokens locks the IBC tokens
func (k Keeper) LockIBCTokens(ctx sdk.Context, receiver sdk.AccAddress, amount sdkmath.Int, recvDenom string) (sdk.Coins, error) {
	ibcCoins := sdk.NewCoins(sdk.NewCoin(recvDenom, amount))
	if err := k.LockTokens(ctx, receiver, ibcCoins); err != nil {
		return sdk.Coins{}, fmt.Errorf("failed to lock tokens: %w", err)
	}
	return ibcCoins, nil
}

// UnlockNativeTokens unlocks the native tokens
func (k Keeper) UnlockNativeTokens(ctx sdk.Context, receiver sdk.AccAddress, amount sdkmath.Int, ibcCoins sdk.Coins) (sdk.Coins, error) {
	nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, amount))
	if err := k.UnlockTokens(ctx, receiver, nativeCoins); err != nil {
		// Unlock previously locked IBC tokens on error
		if unlockErr := k.UnlockTokens(ctx, receiver, ibcCoins); unlockErr != nil {
			k.Logger().Error(fmt.Sprintf("failed to unlock previously locked IBC tokens: %v", unlockErr))
		}
		return sdk.Coins{}, fmt.Errorf("failed to unlock tokens: %w", err)
	}
	return nativeCoins, nil
}

// RecoverZig attempts to recover native zig from address
func (k Keeper) RecoverZig(ctx sdk.Context, address sdk.AccAddress) (sdk.AccAddress, sdk.Coin, sdk.Coin, error) {
	// Check if module is enabled
	if !k.IsEnabled(ctx) {
		err := fmt.Errorf("module disabled: %v", types.ErrModuleDisabled)
		types.EmitTokenWrapperErrorEvent(ctx, err)
		k.Logger().Error(err.Error())
		return sdk.AccAddress{}, sdk.Coin{}, sdk.Coin{}, err
	}

	// Prevent recovery on the operator address
	if address.String() == k.GetOperatorAddress(ctx) {
		err := fmt.Errorf("recovery not allowed on operator address: %v", types.ErrRecoveryNotAllowedOnOperatorAddress)
		types.EmitTokenWrapperErrorEvent(ctx, err)
		k.Logger().Error(err.Error())
		return sdk.AccAddress{}, sdk.Coin{}, sdk.Coin{}, err
	}

	// Get module denom
	moduleDenom := k.GetDenom(ctx)

	// Get IBC denom for the received token
	recvDenom := k.GetIBCRecvDenom(ctx, moduleDenom)

	// Get IBC vouchers amount available in address
	amount := k.bankKeeper.GetBalance(ctx, address, recvDenom)

	// Amount not positive, return an error
	if !amount.IsPositive() {
		err := fmt.Errorf("no IBC vouchers available in address: %v", types.ErrNoIBCVouchersAvailableInAddress)
		types.EmitTokenWrapperErrorEvent(ctx, err)
		k.Logger().Error(err.Error())
		return sdk.AccAddress{}, sdk.Coin{}, sdk.Coin{}, err
	}

	// Handle token conversion
	convertedAmount, err := k.ScaleDownTokenPrecision(ctx, amount.Amount)
	if err != nil {
		types.EmitTokenWrapperErrorEvent(ctx, err)
		k.Logger().Error(fmt.Sprintf("token conversion failed: %v", err))
		return sdk.AccAddress{}, sdk.Coin{}, sdk.Coin{}, err
	}

	// Check balances
	if err := k.CheckBalances(ctx, address, amount.Amount, recvDenom, convertedAmount); err != nil {
		types.EmitTokenWrapperErrorEvent(ctx, err)
		k.Logger().Error(fmt.Sprintf("balances check failed: %v", err))
		return sdk.AccAddress{}, sdk.Coin{}, sdk.Coin{}, err
	}

	// Lock IBC tokens
	ibcCoins, err := k.LockIBCTokens(ctx, address, amount.Amount, recvDenom)
	if err != nil {
		types.EmitTokenWrapperErrorEvent(ctx, err)
		k.Logger().Error(fmt.Sprintf("failed to lock IBC tokens: %v", err))
		return sdk.AccAddress{}, sdk.Coin{}, sdk.Coin{}, err
	}

	// Unlock native tokens
	nativeCoins, err := k.UnlockNativeTokens(ctx, address, convertedAmount, ibcCoins)
	if err != nil {
		types.EmitTokenWrapperErrorEvent(ctx, err)
		k.Logger().Error(fmt.Sprintf("failed to unlock native tokens: %v", err))
		return sdk.AccAddress{}, sdk.Coin{}, sdk.Coin{}, err
	}

	// Track transferred amount
	k.AddToTotalTransferredIn(ctx, convertedAmount)

	// Return address and token coins
	return address, ibcCoins[0], nativeCoins[0], nil
}
