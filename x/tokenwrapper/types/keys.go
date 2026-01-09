package types

const (
	// ModuleName defines the module name
	ModuleName = "tokenwrapper"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_tokenwrapper"

	// Version defines the current version the IBC module supports
	Version = "tokenwrapper-1"
)

var (
	ParamsKey = []byte("p_tokenwrapper")

	// Store keys for tracking ZIG token transfers
	TotalTransferredInKey  = []byte("total_transferred_in")
	TotalTransferredOutKey = []byte("total_transferred_out")

	// Store keys for operator and enabled state
	OperatorAddressKey         = []byte("operator_address")
	ProposedOperatorAddressKey = []byte("proposed_operator_address")
	EnabledKey                 = []byte("enabled")
	PauserAddressesKey         = []byte("pauser_addresses")

	// Store keys for IBC parameters
	NativeClientIdKey       = []byte("native_client_id")
	CounterpartyClientIdKey = []byte("counterparty_client_id")
	NativePortKey           = []byte("native_port")
	CounterpartyPortKey     = []byte("counterparty_port")
	NativeChannelKey        = []byte("native_channel")
	CounterpartyChannelKey  = []byte("counterparty_channel")
	DenomKey                = []byte("denom")
	DecimalDifferenceKey    = []byte("decimal_difference")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
