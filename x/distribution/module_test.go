package distribution

import (
	"github.com/cosmos/cosmos-sdk/client"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/okex/okchain/x/distribution/keeper"
	"github.com/okex/okchain/x/distribution/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestAppModule(t *testing.T) {
	ctx, _, k, _, supplyKeeper := keeper.CreateTestInputDefault(t, false, 1000)

	module := NewAppModule(k, supplyKeeper)
	require.EqualValues(t, ModuleName, module.AppModuleBasic.Name())
	require.EqualValues(t, ModuleName, module.Name())
	require.EqualValues(t, RouterKey, module.Route())
	require.EqualValues(t, QuerierRoute, module.QuerierRoute())

	cdc := codec.New()
	module.RegisterCodec(cdc)
	cliCtx := client.NewContext().WithCodec(cdc)

	msg := module.DefaultGenesis(cdc)
	require.Nil(t, module.ValidateGenesis(cdc, msg))
	require.NotNil(t, module.ValidateGenesis(cdc, []byte{}))
	module.InitGenesis(ctx, cdc, msg)
	exportMsg := module.ExportGenesis(ctx, cdc)

	var gs GenesisState
	require.NotPanics(t, func() {
		types.ModuleCdc.MustUnmarshalJSON(exportMsg, &gs)
	})

	// for coverage
	module.BeginBlock(ctx, abci.RequestBeginBlock{})
	module.EndBlock(ctx, abci.RequestEndBlock{})
	module.GetQueryCmd(cliCtx)
	module.GetTxCmd(cliCtx)
	module.NewQuerierHandler()
	module.NewHandler()
}
