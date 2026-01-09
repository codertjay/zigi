package wasmbinding

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	dexkeeper "zigchain/x/dex/keeper"
	factorykeeper "zigchain/x/factory/keeper"
)

func RegisterCustomPlugins(
	bank *bankkeeper.BaseKeeper,
	factory *factorykeeper.Keeper,
	dex *dexkeeper.Keeper,

) []wasmkeeper.Option {

	wasmQueryPlugin := NewQueryPlugin(
		bank,
		factory,
		dex,
	)

	queryPluginOpt := wasmkeeper.WithQueryPlugins(&wasmkeeper.QueryPlugins{
		Custom: CustomQuerier(wasmQueryPlugin),
	})

	messengerDecoratorOpt := wasmkeeper.WithMessageHandlerDecorator(
		CustomMessageDecorator(
			bank,
			factory,
			dex,
		),
	)

	return []wasmkeeper.Option{
		queryPluginOpt,
		messengerDecoratorOpt,
	}
}

func RegisterStargateQueries(
	queryRouter baseapp.GRPCQueryRouter,
	codec codec.Codec,
) []wasmkeeper.Option {
	queryPluginOpt := wasmkeeper.WithQueryPlugins(
		&wasmkeeper.QueryPlugins{
			Stargate: StargateQuerier(queryRouter, codec),
		})

	return []wasmkeeper.Option{
		queryPluginOpt,
	}
}
