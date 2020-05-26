package types

import (
	"fmt"
	"github.com/okex/okchain/x/params"
)

// const
const (
	defaultParam = "default param"
)

// nolint - Keys for parameter access
var (
	_        params.ParamSet = (*Params)(nil)
	KeyParam                 = []byte("Param")
)

// Params defines the high level settings for maker
type Params struct {
	Param string `json:"param" yaml:"param"`
}

// NewParams creates a new Params instance
func NewParams(param string) Params {
	return Params{
		Param: param,
	}
}

// ParamSetPairs is the implements params.ParamSet
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{Key: KeyParam, Value: &p.Param},
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(defaultParam)
}

// String returns a human readable string representation of the Params
func (p Params) String() string {
	return fmt.Sprintf(`Params:
  Param:    				%s`,
		p.Param)
}
