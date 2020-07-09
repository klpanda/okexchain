package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/okex/okchain/x/params"
)

// Staking params default values
const (
	// DefaultUnbondingTime reflects three weeks in seconds as the default
	// unbonding time.
	// TODO: Justify our choice of default here.
	DefaultUnbondingTime time.Duration = time.Hour * 24 * 7 * 2

	// Default maximum number of bonded validators
	DefaultMaxValidators uint32 = 21

	DefaultEpoch              uint32 = 252

	// Default maximum number of validators to vote
	DefaultMaxValsToAddShares uint32 = 30

	// Default validate rate update interval by hours
	DefaultValidateRateUpdateInterval = 24
)

var (
	// DefaultMinDelegation is the limit value of delegation or undelegation
	DefaultMinDelegation = sdk.NewDecWithPrec(1, 4)
	// DefaultMinSelfDelegation is the default value of each validator's msd (hard code)
	DefaultMinSelfDelegation = sdk.NewDec(10000)
)

// nolint - Keys for parameter access
var (
	KeyUnbondingTime     = []byte("UnbondingTime")
	KeyMaxValidators     = []byte("MaxValidators")
	KeyBondDenom         = []byte("BondDenom")
	KeyEpoch             = []byte("BlocksPerEpoch")    // how many blocks each epoch has
	KeyTheEndOfLastEpoch = []byte("TheEndOfLastEpoch") // a block height that is the end of last epoch

	KeyMaxValsToAddShares = []byte("MaxValsToAddShares")
	KeyMinDelegation      = []byte("MinDelegation")
	KeyMinSelfDelegation  = []byte("MinSelfDelegation")
)

var _ params.ParamSet = (*Params)(nil)

//// Params defines the high level settings for staking
//type Params struct {
//	// time duration of unbonding
//	UnbondingTime time.Duration `json:"unbonding_time" yaml:"unbonding_time"`
//	// note: we need to be a bit careful about potential overflow here, since this is user-determined
//	// maximum number of validators (max uint16 = 65535)
//	MaxValidators uint32 `json:"max_bonded_validators" yaml:"max_bonded_validators"`
//	// epoch for validator update
//	Epoch              uint16 `json:"epoch" yaml:"epoch"`
//	MaxValsToAddShares uint16 `json:"max_validators_to_add_shares" yaml:"max_validators_to_add_shares"`
//	// bondable coin denomination
//	BondDenom string `json:"bond_denom" yaml:"bond_denom"`
//	// limited amount of delegate
//	MinDelegation sdk.Dec `json:"min_delegation" yaml:"min_delegation"`
//	// validator's self declared minimum self delegation
//	MinSelfDelegation sdk.Dec `json:"min_self_delegation" yaml:"min_self_delegation"`
//}

// NewParams creates a new Params instance
func NewParams(unbondingTime time.Duration, maxValidators uint32, bondDenom string, epoch uint32, maxValsToAddShares uint32, minDelegation sdk.Dec, minSelfDelegation sdk.Dec) Params {
	return Params{
		UnbondingTime:      unbondingTime,
		MaxValidators:      maxValidators,
		BondDenom:          bondDenom,
		Epoch:              epoch,
		MaxValsToAddShares: maxValsToAddShares,
		MinDelegation:      minDelegation,
		MinSelfDelegation:  minSelfDelegation,
	}
}

func tmpValidate(value interface{}) error {
	return nil
}

// ParamSetPairs is the implements params.ParamSet
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{KeyUnbondingTime, &p.UnbondingTime, tmpValidate},
		{KeyMaxValidators,  &p.MaxValidators, tmpValidate},
		{KeyBondDenom,  &p.BondDenom, tmpValidate},
		{KeyEpoch,  &p.Epoch, tmpValidate},
		{KeyMaxValsToAddShares,  &p.MaxValsToAddShares, tmpValidate},
		{KeyMinDelegation,  &p.MinDelegation, tmpValidate},
		{KeyMinSelfDelegation,  &p.MinSelfDelegation, tmpValidate},
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultUnbondingTime, DefaultMaxValidators,
		sdk.DefaultBondDenom, DefaultEpoch,
		DefaultMaxValsToAddShares, DefaultMinDelegation,
		DefaultMinSelfDelegation,
	)
}

// String returns a human readable string representation of the Params
func (p Params) String() string {
	return fmt.Sprintf(`Params:
  Unbonding Time:    		%s
  Max Validators:   	 	%d
  Epoch: 					%d
  Bonded Coin Denom: 		%s
  MaxValsToAddShares:       %d
  MinDelegation				%d
  MinSelfDelegation         %d`,
		p.UnbondingTime, p.MaxValidators, p.Epoch, p.BondDenom, p.MaxValsToAddShares, p.MinDelegation, p.MinSelfDelegation)
}

// Validate gives a quick validity check for a set of params
func (p Params) Validate() error {
	if p.BondDenom == "" {
		return fmt.Errorf("staking parameter BondDenom can't be an empty string")
	}
	if p.MaxValidators == 0 {
		return fmt.Errorf("staking parameter MaxValidators must be a positive integer")
	}
	if p.Epoch == 0 {
		return fmt.Errorf("staking parameter Epoch must be a positive integer")
	}
	if p.MaxValsToAddShares == 0 {
		return fmt.Errorf("staking parameter MaxValsToAddShares must be a positive integer")
	}

	return nil
}
