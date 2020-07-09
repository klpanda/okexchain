package upgrade

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/gogo/protobuf/grpc"

	"github.com/okex/okchain/x/upgrade/keeper"
	"github.com/okex/okchain/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/okex/okchain/x/upgrade/client/cli"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
)

// check the implementation of the interface
var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic is a struct of app module basics object
type AppModuleBasic struct{}

// Name returns module name
func (AppModuleBasic) Name() string {
	return ModuleName
}

// RegisterCodec registers module codec
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	types.RegisterCodec(cdc)
}

// DefaultGenesis returns default genesis state
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONMarshaler) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// GetQueryCmd gets the root query command of this module
func (AppModuleBasic) GetQueryCmd(ctx client.Context) *cobra.Command {
	return cli.GetQueryCmd(StoreKey, ctx.Codec)
}

// nolint
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONMarshaler, _ json.RawMessage) error                { return nil }
func (AppModuleBasic) RegisterRESTRoutes(ctx client.Context, _ *mux.Router) { return }
func (AppModuleBasic) GetTxCmd(ctx client.Context) *cobra.Command               { return nil }

// AppModule is a struct of app module
type AppModule struct {
	AppModuleBasic
	keeper Keeper
}

// NewAppModule creates a new AppModule object for upgrade module
func NewAppModule(keeper Keeper) AppModule {
	return AppModule{
		AppModuleBasic{},
		keeper,
	}
}

func (am AppModule) RegisterQueryService(grpc.Server) {}

// InitGenesis initializes module genesis
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	InitGenesis(ctx, am.keeper, genesisState)
	return []abci.ValidatorUpdate{}
}

// Route returns module message route name
func (AppModule) Route() sdk.Route {
	return sdk.NewRoute(RouterKey, nil)
}

// QuerierRoute returns module querier route name
func (AppModule) QuerierRoute() string {
	return QuerierRoute
}

// NewQuerierHandler returns module querier
func (am AppModule) NewQuerierHandler() sdk.Querier {
	return keeper.NewQuerier(am.keeper)
}

// EndBlock is invoked on the end of each block
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	EndBlocker(ctx, am.keeper)
	return nil
}

// nolint
func (AppModule) NewHandler() sdk.Handler                            { return nil }
func (AppModule) ExportGenesis(_ sdk.Context, cdc codec.JSONMarshaler) json.RawMessage        { return nil }
func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry)         { return }
func (AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}
