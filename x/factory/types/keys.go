package types

const (
	// ModuleName defines the module name
	ModuleName = "factory"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	//RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	//MemStoreKey = "mem_factory"
)

var (
	ParamsKey = []byte("p_factory")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
