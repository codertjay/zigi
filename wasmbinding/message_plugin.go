package wasmbinding

import (
	"encoding/json"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	errorsmod "cosmossdk.io/errors"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"zigchain/wasmbinding/bindings"
	dextypes "zigchain/x/dex/types"
	factorykeeper "zigchain/x/factory/keeper"
	factorytypes "zigchain/x/factory/types"

	dexkeeper "zigchain/x/dex/keeper"
)

// CustomMessageDecorator returns decorator for custom CosmWasm bindings messages
// we can pass in here any keepers we need to handle custom messages that are allowed by the contract
func CustomMessageDecorator(
	bank *bankkeeper.BaseKeeper,
	tokenFactory *factorykeeper.Keeper,
	dexFactory *dexkeeper.Keeper,
) func(wasmkeeper.Messenger) wasmkeeper.Messenger {
	// 	DispatchMsg(
	//	ctx sdk.Context,
	//	contractAddr sdk.AccAddress,
	//	contractIBCPortID string, m
	//	sg wasmvmtypes.CosmosMsg
	//	) (events []sdk.Event, data [][]byte, msgResponses [][]*codectypes.Any, err error)
	return func(old wasmkeeper.Messenger) wasmkeeper.Messenger {
		return &CustomMessenger{
			wrapped:      old,
			bank:         bank,
			tokenFactory: tokenFactory,
			dexFactory:   dexFactory,
		}
	}
}

// CustomMessenger is a custom messenger for CosmWasm bindings messages
// passing our module keepers allows us to interact with modules
// and deal with custom messages that are allowed by the contract
type CustomMessenger struct {
	wrapped      wasmkeeper.Messenger
	bank         *bankkeeper.BaseKeeper
	tokenFactory *factorykeeper.Keeper
	dexFactory   *dexkeeper.Keeper
}

// Assert CustomMessenger implements wasmkeeper.Messenger
var _ wasmkeeper.Messenger = (*CustomMessenger)(nil)

// DispatchMsg executes on the contractMsg.
// DispatchMsg executes on the contractMsg.
func (m *CustomMessenger) DispatchMsg(
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	contractIBCPortID string,
	msg wasmvmtypes.CosmosMsg,
) (
	events []sdk.Event,
	data [][]byte,
	msgResponses [][]*codectypes.Any,
	err error,
) {
	// verify that the message is a custom message
	if msg.Custom != nil {
		// only handle the path that is defined
		// leave everything else for the wrapped version

		// set the type of the message to be a ZIGChain message
		var contractMsg bindings.ZMsg

		// checked: unmarshal the message into the contractMsg
		if err := json.Unmarshal(msg.Custom, &contractMsg); err != nil {
			return nil, nil, nil, errorsmod.Wrap(err, "zigchain msg")
		}

		// Handle custom messages
		switch {
		case contractMsg.CreateDenom != nil:
			return m.createDenom(ctx, contractAddr, contractMsg.CreateDenom)
		case contractMsg.SetDenomMetadata != nil:

			return m.setDenomMetadata(ctx, contractAddr, contractMsg.SetDenomMetadata)
		case contractMsg.MintAndSendTokens != nil:

			return m.mintAndSendTokens(ctx, contractAddr, contractMsg.MintAndSendTokens)
		case contractMsg.ProposeDenomAdmin != nil:

			return m.proposeDenomAuth(ctx, contractAddr, contractMsg.ProposeDenomAdmin)

		case contractMsg.UpdateDenomMetadataAuth != nil:

			return m.updateDenomMetadataAuth(ctx, contractAddr, contractMsg.UpdateDenomMetadataAuth)
		case contractMsg.UpdateDenomURI != nil:

			return m.updateDenomURI(ctx, contractAddr, contractMsg.UpdateDenomURI)
		case contractMsg.UpdateDenomMintingCap != nil:

			return m.updateDenomMintingCap(ctx, contractAddr, contractMsg.UpdateDenomMintingCap)
		case contractMsg.BurnTokens != nil:

			return m.burnTokens(ctx, contractAddr, contractMsg.BurnTokens)
		case contractMsg.CreatePool != nil:

			return m.createPool(ctx, contractAddr, contractMsg.CreatePool)
		case contractMsg.AddLiquidity != nil:

			return m.addLiquidity(ctx, contractAddr, contractMsg.AddLiquidity)
		case contractMsg.RemoveLiquidity != nil:

			return m.removeLiquidity(ctx, contractAddr, contractMsg.RemoveLiquidity)

		case contractMsg.SwapExactIn != nil:
			return m.swapExactIn(ctx, contractAddr, contractMsg.SwapExactIn)

		case contractMsg.SwapExactOut != nil:
			return m.swapExactOut(ctx, contractAddr, contractMsg.SwapExactOut)

		default:
			// Return an error for unsupported messages
			return nil, nil, nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "unsupported custom message")
		}
	}

	// Delegate to the wrapped messenger if not a custom message
	return m.wrapped.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
}

// createDenom creates a new token denom
func (m *CustomMessenger) createDenom(
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	createDenom *bindings.CreateDenom,
) ([]sdk.Event, [][]byte, [][]*codectypes.Any, error) {
	err := PerformCreateDenom(m.tokenFactory, m.bank, ctx, contractAddr, createDenom)
	if err != nil {

		return nil, nil, nil, errorsmod.Wrap(err, "perform create denom")
	}
	return nil, nil, nil, nil
}

// PerformCreateDenom is used with createDenom to create a token denom; validates the msgCreateDenom.
func PerformCreateDenom(
	f *factorykeeper.Keeper,
	_ *bankkeeper.BaseKeeper,
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	createDenom *bindings.CreateDenom,
) error {

	if createDenom == nil {

		return wasmvmtypes.InvalidRequest{Err: "create denom null create denom"}
	}

	msgServer := factorykeeper.NewMsgServerImpl(*f)

	msgCreateDenom := factorytypes.NewMsgCreateDenom(
		contractAddr.String(),
		createDenom.Denom,
		createDenom.MintingCap,
		createDenom.CanChangeMintingCap,
		createDenom.URI,
		createDenom.URIHash,
	)

	if err := msgCreateDenom.ValidateBasic(); err != nil {
		return errorsmod.Wrap(err, "failed validating MsgCreateDenom")
	}

	// Create denom
	_, err := msgServer.CreateDenom(
		ctx,
		msgCreateDenom,
	)
	if err != nil {
		return errorsmod.Wrap(err, "creating denom")
	}
	return nil
}

// setDenomMetadata sets the metadata of a denom
func (m *CustomMessenger) setDenomMetadata(
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	setDenomMetadata *bindings.SetDenomMetadata,
) ([]sdk.Event, [][]byte, [][]*codectypes.Any, error) {
	err := PerformSetDenomMetadata(m.tokenFactory, ctx, contractAddr, setDenomMetadata)
	if err != nil {
		return nil, nil, nil, errorsmod.Wrap(err, "perform set denom metadata")
	}
	return nil, nil, nil, nil
}

func PerformSetDenomMetadata(
	f *factorykeeper.Keeper,
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	setDenomMetadata *bindings.SetDenomMetadata,
) error {
	if setDenomMetadata == nil {
		return wasmvmtypes.InvalidRequest{Err: "set denom metadata null set denom metadata"}
	}

	msgServer := factorykeeper.NewMsgServerImpl(*f)

	msgSetDenomMetadata := factorytypes.NewMsgSetDenomMetadata(
		contractAddr.String(),
		setDenomMetadata.Metadata,
	)

	if err := msgSetDenomMetadata.ValidateBasic(); err != nil {
		return errorsmod.Wrap(err, "failed validating MsgSetDenomMetadata")
	}

	_, err := msgServer.SetDenomMetadata(
		ctx,
		msgSetDenomMetadata,
	)
	if err != nil {
		return errorsmod.Wrap(err, "setting denom metadata")
	}
	return nil
}

// mintAndSendTokens mints and sends tokens to a recipient
func (m *CustomMessenger) mintAndSendTokens(
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	mintAndSendTokens *bindings.MintAndSendTokens,
) ([]sdk.Event, [][]byte, [][]*codectypes.Any, error) {
	err := PerformMintAndSendTokens(m.tokenFactory, ctx, contractAddr, mintAndSendTokens)
	if err != nil {
		return nil, nil, nil, errorsmod.Wrap(err, "perform mint and send tokens")
	}
	return nil, nil, nil, nil
}

func PerformMintAndSendTokens(
	f *factorykeeper.Keeper,
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	mintAndSendTokens *bindings.MintAndSendTokens,
) error {
	if mintAndSendTokens == nil {
		return wasmvmtypes.InvalidRequest{Err: "mint tokens null mint tokens"}
	}

	msgServer := factorykeeper.NewMsgServerImpl(*f)

	msgMintAndSendTokens := factorytypes.NewMsgMintAndSendTokens(
		contractAddr.String(),
		mintAndSendTokens.Token,
		mintAndSendTokens.Recipient,
	)

	if err := msgMintAndSendTokens.ValidateBasic(); err != nil {
		return errorsmod.Wrap(err, "failed validating MsgMintAndSendTokens")
	}

	_, err := msgServer.MintAndSendTokens(
		ctx,
		msgMintAndSendTokens,
	)
	if err != nil {
		return errorsmod.Wrap(err, "minting and sending tokens")
	}
	return nil
}

// proposeDenomAuth proposes the admins address: bank, and metadata of a denom, only bank admin can do it
func (m *CustomMessenger) proposeDenomAuth(
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	proposeDenomAdmin *bindings.ProposeDenomAdmin,
) ([]sdk.Event, [][]byte, [][]*codectypes.Any, error) {
	// Perform update denom auth
	err := PerformProposeDenomAuth(
		m.tokenFactory,
		contractAddr,
		ctx,
		proposeDenomAdmin)
	// Check if there is an error
	if err != nil {
		return nil, nil, nil, errorsmod.Wrap(err, "perform update denom auth")
	}
	// Returns on success
	return nil, nil, nil, nil
}

func PerformProposeDenomAuth(
	f *factorykeeper.Keeper,
	contractAddr sdk.AccAddress,
	ctx sdk.Context,
	proposeDenomAdmin *bindings.ProposeDenomAdmin,
) error {
	// Make sure that there is a value on the pointer
	if proposeDenomAdmin == nil {
		return wasmvmtypes.InvalidRequest{Err: "update denom auth null update denom auth"}

	}

	// Create a new message to update denom auth
	msgProposeDenomAuth := factorytypes.NewMsgProposeDenomAdmin(
		contractAddr.String(),
		proposeDenomAdmin.Denom,
		proposeDenomAdmin.BankAdmin,
		proposeDenomAdmin.MetadataAdmin,
	)

	if err := msgProposeDenomAuth.ValidateBasic(); err != nil {
		return errorsmod.Wrap(err, "failed validating MsgProposeDenomAdmin")
	}

	// Create a factory message server
	msgServer := factorykeeper.NewMsgServerImpl(*f)

	_, err := msgServer.ProposeDenomAdmin(
		ctx,
		msgProposeDenomAuth,
	)

	if err != nil {
		return errorsmod.Wrap(err, "updating denom auth")
	}

	return nil
}

// updateDenomMetadataAuth updates the metadata admin of a denom
func (m *CustomMessenger) updateDenomMetadataAuth(
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	updateDenomMetadataAuth *bindings.UpdateDenomMetadataAuth,
) ([]sdk.Event, [][]byte, [][]*codectypes.Any, error) {
	err := PerformUpdateDenomMetadataAuth(m.tokenFactory, ctx, contractAddr, updateDenomMetadataAuth)
	if err != nil {
		return nil, nil, nil, errorsmod.Wrap(err, "perform update denom metadata auth")
	}
	return nil, nil, nil, nil
}

func PerformUpdateDenomMetadataAuth(
	f *factorykeeper.Keeper,
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	updateDenomMetadataAuth *bindings.UpdateDenomMetadataAuth,
) error {
	if updateDenomMetadataAuth == nil {
		return wasmvmtypes.InvalidRequest{Err: "update denom metadata auth null update denom metadata auth"}
	}

	msgUpdateDenomMetadataAuth := factorytypes.NewMsgUpdateDenomMetadataAuth(
		contractAddr.String(),
		updateDenomMetadataAuth.Denom,
		updateDenomMetadataAuth.MetadataAdmin,
	)
	if err := msgUpdateDenomMetadataAuth.ValidateBasic(); err != nil {
		return errorsmod.Wrap(err, "failed validating MsgUpdateDenomMetadataAuth")
	}

	msgServer := factorykeeper.NewMsgServerImpl(*f)

	_, err := msgServer.UpdateDenomMetadataAuth(
		ctx,
		msgUpdateDenomMetadataAuth,
	)
	if err != nil {
		return errorsmod.Wrap(err, "updating denom metadata auth")
	}
	return nil
}

// updateDenomURI updates the URI of a denom and its sha256 hash
func (m *CustomMessenger) updateDenomURI(
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	updateDenomURI *bindings.UpdateDenomURI,
) ([]sdk.Event, [][]byte, [][]*codectypes.Any, error) {
	err := PerformUpdateDenomURI(m.tokenFactory, ctx, contractAddr, updateDenomURI)
	if err != nil {
		return nil, nil, nil, errorsmod.Wrap(err, "perform update denom uri")
	}
	return nil, nil, nil, nil
}

func PerformUpdateDenomURI(
	f *factorykeeper.Keeper,
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	updateDenomURI *bindings.UpdateDenomURI,
) error {
	if updateDenomURI == nil {
		return wasmvmtypes.InvalidRequest{Err: "update denom uri null update denom uri"}
	}

	msgUpdateDenomURI := factorytypes.NewMsgUpdateDenomURI(
		contractAddr.String(),
		updateDenomURI.Denom,
		updateDenomURI.URI,
		updateDenomURI.URIHash,
	)
	if err := msgUpdateDenomURI.ValidateBasic(); err != nil {
		return errorsmod.Wrap(err, "failed validating MsgUpdateDenomURI")
	}

	msgServer := factorykeeper.NewMsgServerImpl(*f)

	_, err := msgServer.UpdateDenomURI(
		ctx,
		msgUpdateDenomURI,
	)
	if err != nil {
		return errorsmod.Wrap(err, "updating denom uri")
	}
	return nil
}

// updateDenomMintingCap updates the minting cap and options to lock minting cap changes on a denom
func (m *CustomMessenger) updateDenomMintingCap(
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	updateDenomMintingCap *bindings.UpdateDenomMintingCap,
) ([]sdk.Event, [][]byte, [][]*codectypes.Any, error) {
	err := PerformUpdateDenomMintingCap(m.tokenFactory, ctx, contractAddr, updateDenomMintingCap)
	if err != nil {
		return nil, nil, nil, errorsmod.Wrap(err, "perform update denom minting cap")
	}
	return nil, nil, nil, nil
}

func PerformUpdateDenomMintingCap(
	f *factorykeeper.Keeper,
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	updateDenomMintingCap *bindings.UpdateDenomMintingCap,
) error {
	if updateDenomMintingCap == nil {
		return wasmvmtypes.InvalidRequest{Err: "update denom minting cap null update denom minting cap"}
	}

	msgUpdateDenomMintingCap := factorytypes.NewMsgUpdateDenomMintingCap(
		contractAddr.String(),
		updateDenomMintingCap.Denom,
		updateDenomMintingCap.MintingCap,
		updateDenomMintingCap.CanChangeMintingCap,
	)
	if err := msgUpdateDenomMintingCap.ValidateBasic(); err != nil {
		return errorsmod.Wrap(err, "failed validating MsgUpdateDenomMintingCap")
	}

	msgServer := factorykeeper.NewMsgServerImpl(*f)

	_, err := msgServer.UpdateDenomMintingCap(
		ctx,
		msgUpdateDenomMintingCap,
	)
	if err != nil {
		return errorsmod.Wrap(err, "updating denom minting cap")
	}
	return nil
}

// burnTokens burns tokens from the signer's account
func (m *CustomMessenger) burnTokens(
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	burnTokens *bindings.BurnTokens,
) ([]sdk.Event, [][]byte, [][]*codectypes.Any, error) {
	err := PerformBurnTokens(m.tokenFactory, ctx, contractAddr, burnTokens)
	if err != nil {
		return nil, nil, nil, errorsmod.Wrap(err, "perform burn tokens")
	}
	return nil, nil, nil, nil
}

func PerformBurnTokens(
	f *factorykeeper.Keeper,
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	burnTokens *bindings.BurnTokens,
) error {
	// Make sure a message is not nil
	if burnTokens == nil {
		return wasmvmtypes.InvalidRequest{Err: "burn tokens null burn tokens"}
	}
	// Make a new message to burn tokens
	msgBurnTokens := factorytypes.NewMsgBurnTokens(
		contractAddr.String(),
		burnTokens.Token,
	)
	// Validate the message
	if err := msgBurnTokens.ValidateBasic(); err != nil {
		return errorsmod.Wrap(err, "failed validating MsgBurnTokens")
	}
	// Create a factory message server
	msgServer := factorykeeper.NewMsgServerImpl(*f)
	// Call burn RPC on server with a burn message
	_, err := msgServer.BurnTokens(
		ctx,
		msgBurnTokens,
	)
	// Check if there is an error
	if err != nil {
		return errorsmod.Wrap(err, "burning tokens")
	}
	return nil
}

// createPool creates a new pool
func (m *CustomMessenger) createPool(
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	createPool *bindings.CreatePool,

) ([]sdk.Event, [][]byte, [][]*codectypes.Any, error) {
	err := PerformCreatePool(m.dexFactory, ctx, contractAddr, createPool)
	if err != nil {
		return nil, nil, nil, errorsmod.Wrap(err, "perform create pool")
	}
	return nil, nil, nil, nil
}

func PerformCreatePool(
	f *dexkeeper.Keeper,
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	createPool *bindings.CreatePool,
) error {
	if createPool == nil {
		return wasmvmtypes.InvalidRequest{Err: "create pool null create pool"}
	}
	msgCreatePool := dextypes.NewMsgCreatePool(
		contractAddr.String(),
		createPool.Base,
		createPool.Quote,
		createPool.Receiver,
	)
	if err := msgCreatePool.ValidateBasic(); err != nil {
		return errorsmod.Wrap(err, "failed validating MsgCreatePool")
	}
	msgServer := dexkeeper.NewMsgServerImpl(*f)
	_, err := msgServer.CreatePool(
		ctx,
		msgCreatePool,
	)
	if err != nil {
		return errorsmod.Wrap(err, "creating pool")
	}
	return nil
}

// SwapExactIn swaps tokens in a pool
func (m *CustomMessenger) swapExactIn(
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	swapExactIn *bindings.SwapExactIn,
) ([]sdk.Event, [][]byte, [][]*codectypes.Any, error) {
	err := PerformSwapExactIn(m.dexFactory, ctx, contractAddr, swapExactIn)
	if err != nil {
		return nil, nil, nil, errorsmod.Wrap(err, "perform swapExactIn")
	}
	return nil, nil, nil, nil
}

func PerformSwapExactIn(
	f *dexkeeper.Keeper,
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	swapExactIn *bindings.SwapExactIn,
) error {
	if swapExactIn == nil {
		return wasmvmtypes.InvalidRequest{Err: "swapExactIn null"}
	}
	msgSwapExactIn := dextypes.NewMsgSwapExactIn(
		contractAddr.String(),
		swapExactIn.Incoming,
		swapExactIn.PoolID,
		swapExactIn.Receiver,
		swapExactIn.OutgoingMin,
	)
	if err := msgSwapExactIn.ValidateBasic(); err != nil {
		return errorsmod.Wrap(err, "failed validating msgSwapExactIn")
	}
	msgServer := dexkeeper.NewMsgServerImpl(*f)
	_, err := msgServer.SwapExactIn(
		ctx,
		msgSwapExactIn,
	)
	if err != nil {
		return errorsmod.Wrap(err, "swapping failed")
	}
	return nil
}

// SwapExactIn swaps tokens in a pool
func (m *CustomMessenger) swapExactOut(
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	swapExactOut *bindings.SwapExactOut,
) ([]sdk.Event, [][]byte, [][]*codectypes.Any, error) {
	err := PerformSwapExactOut(m.dexFactory, ctx, contractAddr, swapExactOut)
	if err != nil {
		return nil, nil, nil, errorsmod.Wrap(err, "perform swapExactOut")
	}
	return nil, nil, nil, nil
}

func PerformSwapExactOut(
	f *dexkeeper.Keeper,
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	swapExactOut *bindings.SwapExactOut,
) error {
	if swapExactOut == nil {
		return wasmvmtypes.InvalidRequest{Err: "Message null in swapExactOut"}
	}
	msgSwapExactOut := dextypes.NewMsgSwapExactOut(
		contractAddr.String(),
		swapExactOut.Outgoing,
		swapExactOut.PoolID,
		swapExactOut.Receiver,
		swapExactOut.IncomingMax,
	)
	if err := msgSwapExactOut.ValidateBasic(); err != nil {
		return errorsmod.Wrap(err, "failed validating msgSwapExactOut")
	}
	msgServer := dexkeeper.NewMsgServerImpl(*f)
	_, err := msgServer.SwapExactOut(
		ctx,
		msgSwapExactOut,
	)
	if err != nil {
		return errorsmod.Wrap(err, "swapping exact out failed")
	}
	return nil
}

func (m *CustomMessenger) addLiquidity(
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	addLiquidity *bindings.AddLiquidity,
) ([]sdk.Event, [][]byte, [][]*codectypes.Any, error) {
	err := PerformAddLiquidity(m.dexFactory, ctx, contractAddr, addLiquidity)
	if err != nil {
		return nil, nil, nil, errorsmod.Wrap(err, "perform add liquidity")
	}
	return nil, nil, nil, nil
}

func PerformAddLiquidity(
	f *dexkeeper.Keeper,
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	addLiquidity *bindings.AddLiquidity,
) error {
	if addLiquidity == nil {
		return wasmvmtypes.InvalidRequest{Err: "add liquidity null add liquidity"}
	}
	msgAddLiquidity := dextypes.NewMsgAddLiquidity(
		contractAddr.String(),
		addLiquidity.PoolID,
		addLiquidity.Base,
		addLiquidity.Quote,
		addLiquidity.Receiver,
	)

	if err := msgAddLiquidity.ValidateBasic(); err != nil {
		return errorsmod.Wrap(err, "failed validating MsgAddLiquidity")
	}

	msgServer := dexkeeper.NewMsgServerImpl(*f)
	_, err := msgServer.AddLiquidity(
		ctx,
		msgAddLiquidity,
	)

	if err != nil {
		return errorsmod.Wrap(err, "adding liquidity")
	}

	return nil
}

func (m *CustomMessenger) removeLiquidity(

	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	removeLiquidity *bindings.RemoveLiquidity,
) ([]sdk.Event, [][]byte, [][]*codectypes.Any, error) {
	err := PerformRemoveLiquidity(m.dexFactory, ctx, contractAddr, removeLiquidity)
	if err != nil {
		return nil, nil, nil, errorsmod.Wrap(err, "perform remove liquidity")
	}
	return nil, nil, nil, nil
}

func PerformRemoveLiquidity(
	f *dexkeeper.Keeper,
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	removeLiquidity *bindings.RemoveLiquidity,
) error {
	if removeLiquidity == nil {
		return wasmvmtypes.InvalidRequest{Err: "remove liquidity null remove liquidity"}
	}

	msgRemoveLiquidity := dextypes.NewMsgRemoveLiquidity(
		contractAddr.String(),
		removeLiquidity.LPToken,
		removeLiquidity.Receiver,
	)

	if err := msgRemoveLiquidity.ValidateBasic(); err != nil {
		return errorsmod.Wrap(err, "failed validating MsgRemoveLiquidity")
	}

	msgServer := dexkeeper.NewMsgServerImpl(*f)

	_, err := msgServer.RemoveLiquidity(
		ctx,
		msgRemoveLiquidity,
	)

	if err != nil {
		return errorsmod.Wrap(err, "removing liquidity")
	}

	return nil
}
