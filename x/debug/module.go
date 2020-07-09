package debug

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/gogo/protobuf/grpc"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/okex/okchain/x/debug/types"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
)

// check the implementation of the interface
var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// app module basics object
type AppModuleBasic struct{}

// module name
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// register module codec
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	RegisterCodec(cdc)
}

// default genesis state
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONMarshaler) json.RawMessage { return nil }

// module validate genesis
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONMarshaler, _ json.RawMessage) error { return nil }

// register rest routes
func (AppModuleBasic) RegisterRESTRoutes(ctx client.Context, _ *mux.Router) {}

// get the root tx command of this module
func (AppModuleBasic) GetTxCmd(ctx client.Context) *cobra.Command {
	return nil
}

// get the root query command of this module
func (AppModuleBasic) GetQueryCmd(ctx client.Context) *cobra.Command {
	return nil
}

// app module
type AppModule struct {
	AppModuleBasic
	keeper Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(keeper Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         keeper,
	}
}

func (am AppModule) RegisterQueryService(grpc.Server) {}

// module init-genesis
func (AppModule) InitGenesis(sdk.Context, codec.JSONMarshaler, json.RawMessage) []abci.ValidatorUpdate { return nil }

// module export genesis
func (AppModule) ExportGenesis(sdk.Context, codec.JSONMarshaler) json.RawMessage { return nil }

// register invariants
func (AppModule) RegisterInvariants(sdk.InvariantRegistry) {}

// module message route name
func (AppModule) Route() sdk.Route {
	return sdk.NewRoute(RouterKey, nil)
}

// module handler
func (am AppModule) NewHandler() sdk.Handler {
	return nil
}

// module querier route name
func (AppModule) QuerierRoute() string {
	return QuerierRoute
}

// module querier
func (am AppModule) NewQuerierHandler() sdk.Querier {
	return NewDebugger(am.keeper)
}

// module begin-block
func (am AppModule) BeginBlock(ctx sdk.Context, _ abci.RequestBeginBlock) {

}

// module end-block
func (AppModule) EndBlock(sdk.Context, abci.RequestEndBlock) []abci.ValidatorUpdate { return nil }
