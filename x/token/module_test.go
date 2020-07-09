package token

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server/api"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/okex/okchain/x/common/version"
	"github.com/okex/okchain/x/token/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestAppModule_InitGenesis(t *testing.T) {
	app, tokenKeeper, _ := getMockDexAppEx(t, 0)
	module := NewAppModule(version.ProtocolVersionV0, tokenKeeper)
	ctx := app.NewContext(true, abci.Header{})
	gs := defaultGenesisState()
	gs.Tokens = nil
	gsJSON := types.ModuleCdc.MustMarshalJSON(gs)
	cliCtx := client.NewContext().WithCodec(app.Cdc).WithJSONMarshaler(app.AppCodec)

	err := module.ValidateGenesis(app.AppCodec, gsJSON)
	require.NoError(t, err)

	vu := module.InitGenesis(ctx, app.AppCodec, gsJSON)
	params := tokenKeeper.GetParams(ctx)
	require.Equal(t, gs.Params, params)
	require.Equal(t, vu, []abci.ValidatorUpdate{})

	export := module.ExportGenesis(ctx, app.AppCodec)
	require.EqualValues(t, gsJSON, []byte(export))

	require.EqualValues(t, types.ModuleName, module.Name())
	require.EqualValues(t, types.ModuleName, module.AppModuleBasic.Name())
	require.EqualValues(t, types.RouterKey, module.Route())
	require.EqualValues(t, types.QuerierRoute, module.QuerierRoute())
	module.NewHandler()
	module.GetQueryCmd(cliCtx)
	module.GetTxCmd(cliCtx)
	module.NewQuerierHandler()
	rs := api.New(cliCtx, nil)
	module.RegisterRESTRoutes(rs.ClientCtx, rs.Router)
	module.BeginBlock(ctx, abci.RequestBeginBlock{})
	module.EndBlock(ctx, abci.RequestEndBlock{})
	module.DefaultGenesis(app.AppCodec)
	module.RegisterCodec(codec.New())

	gsJSON = []byte("[[],{}]")
	err = module.ValidateGenesis(app.AppCodec, gsJSON)
	require.NotNil(t, err)
}
