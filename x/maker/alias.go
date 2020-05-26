package maker

import "github.com/okex/okchain/x/maker/types"

const (
	ModuleName = types.ModuleName
	StoreKey   = types.StoreKey
)

var (
	// functions aliases
	RegisterCodec       = types.RegisterCodec
	DefaultGenesisState = types.DefaultGenesisState

	// variable aliases
	ModuleCdc = types.ModuleCdc
)

type (
	GenesisState = types.GenesisState
)
