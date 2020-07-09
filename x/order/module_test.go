package order

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"testing"

	"github.com/okex/okchain/x/common/version"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/okex/okchain/x/order/keeper"
	"github.com/okex/okchain/x/order/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestAppModule(t *testing.T) {
	testInput := keeper.CreateTestInput(t)
	keeper := testInput.OrderKeeper
	ctx := testInput.Ctx
	module := NewAppModule(version.CurrentProtocolVersion, testInput.OrderKeeper)

	require.EqualValues(t, ModuleName, module.Name())
	require.EqualValues(t, RouterKey, module.Route())
	require.EqualValues(t, QuerierRoute, module.QuerierRoute())

	cdc := codec.New()
	module.RegisterCodec(cdc)
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	appCodec := codec.NewHybridCodec(cdc, interfaceRegistry)
	cliCtx := client.NewContext().WithCodec(cdc).WithJSONMarshaler(appCodec)

	msg := module.DefaultGenesis(appCodec)
	require.Nil(t, module.ValidateGenesis(appCodec, msg))
	require.NotNil(t, module.ValidateGenesis(appCodec, []byte{}))

	module.InitGenesis(ctx, appCodec, msg)
	params := keeper.GetParams(ctx)
	require.EqualValues(t, 259200, params.OrderExpireBlocks)
	exportMsg := module.ExportGenesis(ctx, appCodec)

	var gs GenesisState
	types.ModuleCdc.MustUnmarshalJSON(exportMsg, &gs)
	require.EqualValues(t, msg, types.ModuleCdc.MustMarshalJSON(gs))

	// for coverage
	module.BeginBlock(ctx, abci.RequestBeginBlock{})
	module.EndBlock(ctx, abci.RequestEndBlock{})
	module.GetQueryCmd(cliCtx)
	module.GetTxCmd(cliCtx)
	module.NewQuerierHandler()
	module.NewHandler()
}
