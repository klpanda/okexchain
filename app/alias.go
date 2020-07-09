package app

import (
	"github.com/okex/okchain/app/protocol"
)

var (
	// MakeCodec is the function alias for codec maker
	MakeEncodingConfig = protocol.MakeEncodingConfig
	// ModuleBasics is the variable alias for NewBasicManager
	ModuleBasics = protocol.ModuleBasics
	// DefaultCLIHome is the directory for okchaincli
	DefaultCLIHome = protocol.DefaultCLIHome
	// DefaultNodeHome is the directory for okchaind
	DefaultNodeHome = protocol.DefaultNodeHome
)
