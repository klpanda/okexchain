package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/okex/okchain/x/distribution/types"
)

// GetDistributionAccount returns the distribution ModuleAccount
func (k Keeper) GetDistributionAccount(ctx sdk.Context) authexported.ModuleAccountI {
	return k.accKeeper.GetModuleAccount(ctx, types.ModuleName)
}
