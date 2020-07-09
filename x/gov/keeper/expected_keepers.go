package keeper

import (
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
	stakingexported "github.com/okex/okchain/x/staking/exported"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// StakingKeeper defines expected staking keeper (Validator and Delegator sets)
type StakingKeeper interface {
	// iterate through bonded validators by operator address, execute func for each validator
	// gov use it for getting votes of validator
	IterateBondedValidatorsByPower(sdk.Context,
		func(index int64, validator stakingexported.ValidatorI) (stop bool))

	// gov use it for getting votes of delegator which has been voted to validator
	Delegator(ctx sdk.Context, delAddr sdk.AccAddress) stakingexported.DelegatorI

	TotalBondedTokens(context sdk.Context) sdk.Dec
	IterateDelegations(ctx sdk.Context, delegator sdk.AccAddress, fn func(index int64, delegation exported.DelegationI) (stop bool))
}
