package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/okex/okchain/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
)

// GetBondedPool returns the bonded tokens pool's module account
func (k Keeper) GetBondedPool(ctx sdk.Context) (bondedPool authtypes.ModuleAccountI) {
	return k.accKeeper.GetModuleAccount(ctx, types.BondedPoolName)
}

// GetNotBondedPool returns the not bonded tokens pool's module account
func (k Keeper) GetNotBondedPool(ctx sdk.Context) (notBondedPool authtypes.ModuleAccountI) {
	return k.accKeeper.GetModuleAccount(ctx, types.NotBondedPoolName)
}

// bondedTokensToNotBonded transfers coins from the bonded to the not bonded pool within staking
func (k Keeper) bondedTokensToNotBonded(ctx sdk.Context, tokens sdk.DecCoin) {

	coins := tokens.ToCoins()
	err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.BondedPoolName, types.NotBondedPoolName, coins)
	if err != nil {
		panic(err)
	}
}

// TotalBondedTokens total staking tokens supply which is bonded
// TODO:No usages found in project files,remove it later
func (k Keeper) TotalBondedTokens(ctx sdk.Context) sdk.Dec {
	bondedPool := k.GetBondedPool(ctx)
	return k.bankKeeper.GetAllBalances(ctx, bondedPool.GetAddress()).AmountOf(k.BondDenom(ctx))
}

// StakingTokenSupply staking tokens from the total supply
func (k Keeper) StakingTokenSupply(ctx sdk.Context) sdk.Dec {
	return k.bankKeeper.GetSupply(ctx).GetTotal().AmountOf(k.BondDenom(ctx))
}

// BondedRatio the fraction of the staking tokens which are currently bonded
func (k Keeper) BondedRatio(ctx sdk.Context) sdk.Dec {
	bondedPool := k.GetBondedPool(ctx)

	stakeSupply := k.StakingTokenSupply(ctx)
	if stakeSupply.IsPositive() {
		return k.bankKeeper.GetAllBalances(ctx, bondedPool.GetAddress()).AmountOf(k.BondDenom(ctx)).Quo(stakeSupply)
	}
	return sdk.ZeroDec()
}


func (s Keeper) IterateDelegations(ctx sdk.Context, delegator sdk.AccAddress,
	fn func(index int64, delegation exported.DelegationI) (stop bool),
) {}

