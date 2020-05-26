package types

// GenesisState - all maker state that must be provided at genesis
type GenesisState struct {
	Params Params `json:"params" yaml:"params"`
}

// DefaultGenesisState gets the default genesis state
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params: DefaultParams(),
	}
}
