package poolswap

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/std"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/okex/okchain/x/poolswap/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestAppModule(t *testing.T) {
	mapp, _ := getMockApp(t, 1)
	keeper := mapp.swapKeeper
	mapp.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 2}})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{}).WithBlockHeight(10)
	module := NewAppModule(keeper)

	require.EqualValues(t, ModuleName, module.Name())
	require.EqualValues(t, RouterKey, module.Route())
	require.EqualValues(t, QuerierRoute, module.QuerierRoute())

	cdc := ModuleCdc
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	appCodec := codec.NewHybridCodec(cdc, interfaceRegistry)
	std.RegisterInterfaces(interfaceRegistry)

	cliCtx := client.NewContext().WithCodec(cdc).WithJSONMarshaler(appCodec)

	msg := module.DefaultGenesis(appCodec)
	require.Nil(t, module.ValidateGenesis(appCodec, msg))
	require.NotNil(t, module.ValidateGenesis(appCodec, []byte{}))

	module.InitGenesis(ctx, appCodec, msg)
	params := keeper.GetParams(ctx)
	require.EqualValues(t, types.DefaultParams().FeeRate, params.FeeRate)
	exportMsg := module.ExportGenesis(ctx, appCodec)

	var gs GenesisState
	ModuleCdc.MustUnmarshalJSON(exportMsg, &gs)
	require.EqualValues(t, msg, json.RawMessage(ModuleCdc.MustMarshalJSON(gs)))

	// for coverage
	module.BeginBlock(ctx, abci.RequestBeginBlock{})
	module.EndBlock(ctx, abci.RequestEndBlock{})
	module.GetQueryCmd(cliCtx)
	module.GetTxCmd(cliCtx)
	module.NewQuerierHandler()
	module.NewHandler()
	rs := api.New(cliCtx, nil)
	module.RegisterRESTRoutes(rs.ClientCtx, rs.Router)
	module.RegisterInvariants(nil)
	module.RegisterCodec(codec.New())
}
