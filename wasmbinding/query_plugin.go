package wasmbinding

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/gogoproto/proto"

	errorsmod "cosmossdk.io/errors"
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/wasmbinding/bindings"
	dexkeeper "zigchain/x/dex/keeper"
	dextypes "zigchain/x/dex/types"
	factoryTypes "zigchain/x/factory/types"
	"zigchain/zutils/validators"
)

// StargateQuerier dispatches whitelisted stargate queries
func StargateQuerier(
	queryRouter baseapp.GRPCQueryRouter,
	cdc codec.Codec,
) func(ctx sdk.Context, request *wasmvmtypes.StargateQuery) ([]byte, error) {
	return func(ctx sdk.Context, request *wasmvmtypes.StargateQuery) ([]byte, error) {
		protoResponseType, err := getWhitelistedQuery(request.Path)
		if err != nil {
			return nil, err
		}

		// no matter what happens after this point, we must return
		// the response type to prevent sync.Pool from leaking.
		defer returnStargateResponseToPool(request.Path, protoResponseType)

		route := queryRouter.Route(request.Path)
		if route == nil {
			return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("No route to query '%s'", request.Path)}
		}

		res, err := route(ctx, &abci.RequestQuery{
			Data: request.Data,
			Path: request.Path,
		})
		if err != nil {
			return nil, err
		}

		if res.Value == nil {
			return nil, fmt.Errorf("res returned from abci query route is nil")
		}

		bz, err := ConvertProtoToJSONMarshal(protoResponseType, res.Value, cdc)
		if err != nil {
			return nil, err
		}

		return bz, nil
	}
}

// CustomQuerier dispatches custom CosmWasm bindings queries.
func CustomQuerier(qp *QueryPlugin) func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	return func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
		var contractQuery bindings.ZQuery
		// checked: unmarshal the request into the contract query struct
		if err := json.Unmarshal(request, &contractQuery); err != nil {
			return nil, errorsmod.Wrap(err, "zigchain query")
		}

		switch {
		case contractQuery.Denom != nil:
			denom := contractQuery.Denom.Denom

			kDenom, found := qp.factoryKeeper.GetDenom(ctx, denom)

			if !found {
				return nil,
					errorsmod.Wrapf(
						factoryTypes.ErrDenomNotFound,
						"denom '%s' not found",
						denom,
					)
			}

			denomAuth, found := qp.factoryKeeper.GetDenomAuth(ctx, denom)

			if !found {
				return nil,
					errorsmod.Wrapf(
						factoryTypes.ErrDenomAuthNotFound,
						"denomAuth '%s' not found",
						denom,
					)
			}

			if kDenom.Denom != denomAuth.Denom {
				return nil,
					errorsmod.Wrapf(
						factoryTypes.ErrDenomNotFound,
						"denom '%s' not found",
						denom,
					)
			}

			res := bindings.DenomResponse{
				Denom:               denom,
				Minted:              kDenom.Minted,
				MintingCap:          kDenom.MintingCap,
				CanChangeMintingCap: kDenom.CanChangeMintingCap,
				Creator:             kDenom.Creator,
				BankAdmin:           denomAuth.BankAdmin,
				MetadataAdmin:       denomAuth.MetadataAdmin,
			}

			bz, err := json.Marshal(res)
			if err != nil {
				return nil, errorsmod.Wrap(err, "denom query response")
			}

			return bz, nil

		case contractQuery.Pool != nil:
			poolID := contractQuery.Pool.PoolID

			kPool, found := qp.dexKeeper.GetPool(ctx, poolID)

			if !found {
				return nil,
					errorsmod.Wrapf(
						dextypes.ErrPoolNotFound,
						"pool '%s' not found",
						poolID,
					)
			}

			res := bindings.PoolResponse{
				PoolID:  kPool.PoolId,
				LPToken: kPool.LpToken,
				Creator: kPool.Creator,
				Fee:     kPool.Fee,
				Formula: kPool.Formula,
				Coins:   kPool.Coins,
			}

			bz, err := json.Marshal(res)
			if err != nil {
				return nil, errorsmod.Wrap(err, "pool query response")
			}

			return bz, nil

		case contractQuery.SwapIn != nil:
			poolID := contractQuery.SwapIn.PoolID
			coinIn := contractQuery.SwapIn.CoinIn

			if err := validators.CheckPoolId(poolID); err != nil {
				return nil, err
			}

			kPool, found := qp.dexKeeper.GetPool(ctx, poolID)

			if !found {
				return nil,
					errorsmod.Wrapf(
						dextypes.ErrPoolNotFound,
						"pool '%s' not found",
						poolID,
					)
			}

			coinOut, feeCoin, err := dexkeeper.CalculateSwapAmount(&kPool, coinIn)

			if err != nil {
				return nil, err
			}

			res := bindings.SwapInResponse{
				CoinOut: coinOut,
				Fee:     feeCoin,
			}

			bz, err := json.Marshal(res)
			if err != nil {
				return nil, errorsmod.Wrap(err, "swap in query response")
			}

			return bz, nil

		default:
			return nil, wasmvmtypes.UnsupportedRequest{
				Kind: "unknown zigchain query variant",
			}
		}
	}
}

// ConvertProtoToJsonMarshal  unmarshals the given bytes into a proto message and then marshals it to json.
// This is done so that clients calling stargate queries do not need to define their own proto unmarshalers,
// being able to use response directly by json marshalling, which is supported in cosmwasm.
func ConvertProtoToJSONMarshal(protoResponseType proto.Message, bz []byte, cdc codec.Codec) ([]byte, error) {
	// unmarshal binary into stargate response data structure
	err := cdc.Unmarshal(bz, protoResponseType)
	if err != nil {
		return nil, err
	}

	bz, err = cdc.MarshalJSON(protoResponseType)
	if err != nil {
		return nil, err
	}

	protoResponseType.Reset()

	return bz, nil
}
