package keeper

import (
	"context"
	"fmt"
	"math/big"
	"strconv"

	log "github.com/sirupsen/logrus"

	errorsmod "cosmossdk.io/errors"
	cosmosmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"zigchain/x/dex/events"
	"zigchain/x/dex/types"
	"zigchain/zutils/constants"
)

func (k msgServer) CreatePool(
	goCtx context.Context,
	msg *types.MsgCreatePool,
) (
	*types.MsgCreatePoolResponse,
	error,
) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get the sender address we will need it to transfer funds around
	sender, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil,
			errorsmod.Wrapf(
				sdkerrors.ErrInvalidAddress,
				"Invalid address: %s",
				msg.Creator,
			)
	}

	// Using coins will ensure sorting, also check for duplicates and future prof for multi coin pools
	poolUid, found := k.GetPoolUidsFromCoins(ctx, sdk.NewCoins(msg.Base, msg.Quote))

	if found {
		return nil,
			errorsmod.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"Pool with id: %s already exists with uid: %s",
				poolUid.PoolId,
				poolUid.PoolUid,
			)

	}

	// Check if the sender has enough base tokens to create the pool (comparing to what was in the message)
	// Double check as it will check again on send - check error message and maybe remove
	if !k.bankKeeper.HasBalance(ctx, sender, msg.Base) {
		return nil,
			errorsmod.Wrapf(
				sdkerrors.ErrInsufficientFunds,
				"Signer wallet does not have %s tokens",
				msg.Base.String(),
			)
	}

	// Check if the sender has enough quote tokens to create the pool (comparing to what was in the message)
	// Double check as it will check again on send - check error message and maybe remove
	if !k.bankKeeper.HasBalance(ctx, sender, msg.Quote) {
		return nil,
			errorsmod.Wrapf(
				sdkerrors.ErrInsufficientFunds,
				"Signer wallet does not have %s tokens",
				msg.Quote.String(),
			)
	}

	params := k.GetParams(ctx)

	if params.CreationFee > 0 {
		if !k.bankKeeper.HasBalance(ctx, sender, sdk.NewCoin(constants.BondDenom, cosmosmath.NewInt(int64(params.CreationFee)))) {
			return nil,
				errorsmod.Wrapf(
					sdkerrors.ErrInsufficientFunds,
					"Signer wallet does not have %s tokens",
					sdk.NewCoin(constants.BondDenom, cosmosmath.NewInt(int64(params.CreationFee))).String(),
				)
		}
	}

	// if the creation fee is greater than 0, we will charge the sender
	if params.CreationFee > 0 {

		// beneficiary receives the fee if set, otherwise it is burned
		if params.Beneficiary != "" {
			fee := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, cosmosmath.NewInt(int64(params.CreationFee))))
			beneficiary, err := sdk.AccAddressFromBech32(params.Beneficiary)
			if err != nil {
				return nil,
					errorsmod.Wrapf(
						sdkerrors.ErrInvalidAddress,
						"Invalid address: %s",
						params.Beneficiary,
					)
			}

			// Charge for creating a pool first
			err = k.bankKeeper.SendCoins(
				ctx,
				sender,
				beneficiary,
				fee,
			)

			if err != nil {
				return nil,
					errorsmod.Wrapf(
						sdkerrors.ErrInsufficientFunds,
						"Error while sending coins %s from account: %s to beneficiary: %s",
						fee.String(),
						sender,
						params.Beneficiary,
					)
			}
		} else {
			// Charge for creating a pool first
			err = k.bankKeeper.SendCoinsFromAccountToModule(
				ctx,
				sender,
				types.ModuleName,
				sdk.NewCoins(sdk.NewCoin(constants.BondDenom, cosmosmath.NewInt(int64(params.CreationFee)))),
			)

			if err != nil {
				return nil,
					errorsmod.Wrapf(
						sdkerrors.ErrInsufficientFunds,
						"Error while sending coins %s from account: %s to module: %s",
						sdk.NewCoin(constants.BondDenom, cosmosmath.NewInt(int64(params.CreationFee))).String(),
						sender,
						types.ModuleName,
					)
			}

			err = k.bankKeeper.BurnCoins(
				ctx,
				types.ModuleName,
				sdk.NewCoins(
					sdk.NewCoin(constants.BondDenom,
						cosmosmath.NewInt(int64(params.CreationFee)),
					),
				),
			)
			if err != nil {
				msg := fmt.Sprintf(
					"CreatePool: BurnCoins Failed in burning %s coins from module %s",
					sdk.NewCoin(constants.BondDenom, cosmosmath.NewInt(int64(params.CreationFee))).String(),
					types.ModuleName,
				)
				log.Error(msg)
				return nil,
					errorsmod.Wrapf(
						sdkerrors.ErrInsufficientFunds,
						"%s", msg,
					)
			}

		}
	}

	// Get the pool id by auto-incrementing the pool count
	poolID := k.GetAndSetNextPoolID(ctx)

	// Create the pool id string it is the next incremented pool id prefixed with zp, e.g., zp1, zp2, zp3, ...
	poolIDString := constants.PoolPrefix + strconv.FormatUint(poolID, 10)

	// Check if the poll with that ID already exists (just in case)
	_, isFound := k.GetPool(
		ctx,
		poolIDString,
	)

	// now that we have that, we can report an error from above if the pool already exists
	if isFound {
		return nil,
			errorsmod.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"Pool with id: %s already exists",
				poolIDString,
			)
	}

	// This will create a pool account
	var poolAddress sdk.AccAddress
	poolAddress, err = k.CreatePoolAccount(ctx, poolIDString)

	if err != nil {
		return nil, err
	}

	// If in the request, the receiver is sent to receive the LP tokens,
	// we will use that address, otherwise we will use the sender address
	var receiver sdk.AccAddress
	if msg.Receiver != "" {
		receiver, err = sdk.AccAddressFromBech32(msg.Receiver)
		if err != nil {
			return nil, err
		}
	} else {
		receiver = sender
	}

	// Transfer the base and quote coins from the sender to the module account
	err = k.SendFromAddressToPool(
		ctx,
		sender,
		poolIDString,
		sdk.NewCoins(msg.Base, msg.Quote),
	)

	// Check if there was an error while transferring the coins
	if err != nil {
		return nil,
			errorsmod.Wrapf(
				err,
				"CreatePool: SendFromAddressToPool error while sending coins %s and %s",
				msg.Base.String(),
				msg.Quote.String(),
			)
	}

	// Let's create the LP token for the pool
	lpAmount, userShares, err := k.initialLiquidityShares(ctx, poolIDString, msg.Base, msg.Quote)

	if err != nil {
		return nil, err
	}

	// Mint the total LP tokens to the module account
	err = k.bankKeeper.MintCoins(
		ctx,
		types.ModuleName,
		sdk.NewCoins(userShares),
	)

	if err != nil {
		return nil,
			errorsmod.Wrapf(
				sdkerrors.ErrInsufficientFunds,
				"Error in minting coins",
			)
	}

	// Send the LP tokens to the receiver
	err = k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		receiver,
		sdk.NewCoins(userShares),
	)
	if err != nil {
		msg := fmt.Sprintf(
			"CreatePool: SendCoinsFromModuleToAccount Failed in sending %s coins from module %s to account: %s",
			userShares,
			types.ModuleName,
			receiver,
		)
		log.Error(msg)
		return nil,
			errorsmod.Wrapf(
				sdkerrors.ErrInsufficientFunds,
				"%s", msg,
			)
	}

	// Generate the pool object
	var pool = types.Pool{
		Creator: msg.Creator,
		PoolId:  poolIDString,
		LpToken: lpAmount,
		Fee:     params.NewPoolFeePct, // e.g. 0.5%
		Formula: types.FormulaConstantProduct,
		Coins:   sdk.NewCoins(msg.Base, msg.Quote),
		Address: poolAddress.String(),
	}

	// Save the pool to the store
	k.SetPool(
		ctx,
		pool,
	)

	// Set a secondary index for the pool
	k.SetPoolUidFromPool(ctx, pool)

	events.EmitPoolCreateEvent(ctx, sender, &pool, receiver)

	return &types.MsgCreatePoolResponse{
		PoolId:  poolIDString,
		Base:    msg.Base,
		Quote:   msg.Quote,
		LpToken: userShares,
	}, nil
}

// integerSqrt computes the integer square root of an cosmos.Int value deterministically.
func integerSqrt(value cosmosmath.Int) cosmosmath.Int {
	if value.IsNegative() {
		panic("cannot compute square root of a negative number")
	}

	// Use math/big for precise square root computation
	bigValue := value.BigInt()          // Convert cosmos.Int to big.Int
	root := new(big.Int).Sqrt(bigValue) // Compute square root

	return cosmosmath.NewIntFromBigInt(root) // Convert back to cosmos.Int
}

// initialLiquidityShares calculates the number of shares to mint based on the added liquidity.
// It also implements a minimal liquidity lock to prevent share inflation attacks.
func (k msgServer) initialLiquidityShares(ctx sdk.Context, poolId string, base sdk.Coin, quote sdk.Coin) (sdk.Coin, sdk.Coin, error) {
	// Ensure the amounts are non-negative
	if !base.Amount.IsPositive() || !quote.Amount.IsPositive() {
		return sdk.Coin{}, sdk.Coin{}, types.ErrNonPositiveAmounts
	}

	// Compute the product of the base and quote amounts
	product := base.Amount.Mul(quote.Amount)

	// Compute the integer square root deterministically
	lpAmount := integerSqrt(product)

	// Get minimal liquidity lock from params
	// Minimal liquidity lock is used to prevent share inflation attacks
	params := k.GetParams(ctx)
	minimalLock := cosmosmath.NewIntFromUint64(uint64(params.MinimalLiquidityLock))

	// lpAmount must be greater than minimal lock
	if lpAmount.LT(minimalLock) {
		return sdk.Coin{}, sdk.Coin{}, types.ErrInsufficientLiquidityLock
	}

	userShares := lpAmount.Sub(minimalLock)

	// Return both the total LP tokens and the user shares
	return sdk.NewCoin(poolId, lpAmount), sdk.NewCoin(poolId, userShares), nil
}
