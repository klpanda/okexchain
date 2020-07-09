package token

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// BankKeeper defines the expected supply Keeper (noalias)
type AccountKeeper interface {
	GetModuleAccount(ctx sdk.Context, name string) authexported.ModuleAccountI

	// TODO remove with genesis 2-phases refactor https://github.com/cosmos/cosmos-sdk/issues/2862
	SetModuleAccount(sdk.Context, authexported.ModuleAccountI)

	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authexported.AccountI
}

// StakingKeeper defines the expected staking Keeper (noalias)
type StakingKeeper interface {
	IsValidator(ctx sdk.Context, addr sdk.AccAddress) bool
}
