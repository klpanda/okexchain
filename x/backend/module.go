package backend

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/gogo/protobuf/grpc"

	"github.com/okex/okchain/x/backend/client/cli"
	"github.com/okex/okchain/x/backend/client/rest"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// type check to ensure the interface is properly implemented
var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic app module Basics object
type AppModuleBasic struct{}

// Name return ModuleName
func (AppModuleBasic) Name() string {
	return ModuleName
}

// RegisterCodec registers module codec
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
}

// DefaultGenesis returns nil
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONMarshaler) json.RawMessage {
	return nil
}

// ValidateGenesis  Validation check of the Genesis
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONMarshaler, bz json.RawMessage) error {
	return nil
}

// RegisterRESTRoutes register rest routes
func (AppModuleBasic) RegisterRESTRoutes(ctx client.Context, rtr *mux.Router) {

	rest.RegisterRoutes(ctx, rtr)
	rest.RegisterRoutesV2(ctx, rtr)
}

// GetQueryCmd return the root query command of this module
func (AppModuleBasic) GetQueryCmd(ctx client.Context) *cobra.Command {
	return cli.GetQueryCmd(QuerierRoute, ctx.Codec)
}

// GetTxCmd return the root tx command of this module
func (AppModuleBasic) GetTxCmd(ctx client.Context) *cobra.Command {
	return nil
}

// AppModule is a struct of app module
type AppModule struct {
	AppModuleBasic
	keeper Keeper
}

// NewAppModule creates a new AppModule Object
func NewAppModule(k Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         k,
	}
}

// Name returns module name
func (AppModule) Name() string {
	return ModuleName
}

// RegisterInvariants registers invariants
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {}

// Route returns module message route name
func (am AppModule) Route() sdk.Route {
	return sdk.NewRoute(RouterKey, am.NewHandler())
}

// NewHandler returns module handler
func (am AppModule) NewHandler() sdk.Handler {
	return nil
}

// QuerierRoute returns module querier route name
func (am AppModule) QuerierRoute() string {
	return QuerierRoute
}

// NewQuerierHandler returns module querier
func (am AppModule) NewQuerierHandler() sdk.Querier {
	return NewQuerier(am.keeper)
}

// BeginBlock is invoked on the beginning of each block
func (am AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {
}

// EndBlock is invoked on the end of each block, start to execute backend logic
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	EndBlocker(ctx, am.keeper)
	return nil
}

func (am AppModule) RegisterQueryService(grpc.Server) {}

// InitGenesis initialize module genesis
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, data json.RawMessage) []abci.ValidatorUpdate {
	return nil
}

// ExportGenesis exports module genesis
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONMarshaler) json.RawMessage {
	return nil
}
