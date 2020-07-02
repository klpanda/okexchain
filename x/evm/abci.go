package evm

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/okex/okchain/x/evm/keeper"

	abci "github.com/tendermint/tendermint/abci/types"
)

func EndBlocker(ctx sdk.Context, keeper keeper.Keeper) []abci.ValidatorUpdate {
	// Gas costs are handled within msg handler so costs should be ignored
	ctx = ctx.WithBlockGasMeter(sdk.NewInfiniteGasMeter())

	// Update account balances before committing other parts of state
	keeper.StateDB.WithContext(ctx).UpdateAccounts()

	// Commit state objects to KV store
	_, err := keeper.StateDB.Commit(true)
	if err != nil {
		panic(err)
	}

	// Clear accounts cache after account data has been committed
	keeper.StateDB.ClearStateObjects()

	return []abci.ValidatorUpdate{}
}
