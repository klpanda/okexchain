package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/okex/okchain/x/dex/types"
)

// RegisterInvariants registers all dex invariants
func RegisterInvariants(ir sdk.InvariantRegistry, keeper IKeeper, accKeeper AccountKeeper, bankKeeper BankKeeper) {
	ir.RegisterRoute(types.ModuleName, "module-account", ModuleAccountInvariant(keeper, accKeeper, bankKeeper))
}

// ModuleAccountInvariant checks that the module account coins reflects the sum of
// locks amounts held on store
func ModuleAccountInvariant(keeper IKeeper, accKeeper AccountKeeper, bankKeeper BankKeeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var depositsCoins, withdrawCoins sdk.DecCoins

		// get product deposits
		for _, product := range keeper.GetTokenPairs(ctx) {
			if product == nil {
				panic("the nil pointer is not expected")
			}
			depositsCoins = depositsCoins.Add(sdk.DecCoins{product.Deposits}...)
		}

		keeper.IterateWithdrawInfo(ctx, func(_ int64, withdrawInfo types.WithdrawInfo) (stop bool) {
			withdrawCoins = withdrawCoins.Add(sdk.DecCoins{withdrawInfo.Deposits}...)
			return false
		})

		moduleAcc := accKeeper.GetModuleAccount(ctx, types.ModuleName)

		bal := bankKeeper.GetAllBalances(ctx, moduleAcc.GetAddress())
		broken := !bal.IsEqual(depositsCoins.Add(withdrawCoins...))

		return sdk.FormatInvariant(types.ModuleName, "module coins",
			fmt.Sprintf("\tdex ModuleAccount coins: %s\n\tsum of deposits coins: %s\tsum of withdraw coins: %s\n",
				bal, depositsCoins, withdrawCoins)), broken
	}
}
