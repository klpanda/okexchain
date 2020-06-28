package types

const (
	ModuleName    = "evm"
	StoreKey      = ModuleName
	CodeKey       = StoreKey + "_code"
	LogKey        = StoreKey + "_log"
	StoreDebugKey = StoreKey + "_debug"
	QuerierRoute  = ModuleName
	RouterKey     = ModuleName
)

var (
	LogIndexKey = []byte("logIndexKey")
)
