//+build !stream

package stream

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/gogo/protobuf/grpc"
)

const (
	ModuleName = "stream"
)

// type check to ensure the interface is properly implemented
var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// app module Basics object
type AppModuleBasic struct{}

func (AppModuleBasic) Name() string { return ModuleName }

func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {}

func (AppModuleBasic) DefaultGenesis(cdc codec.JSONMarshaler) json.RawMessage { return nil }

// Validation check of the Genesis
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONMarshaler, bz json.RawMessage) error { return nil }

// Register rest routes
func (AppModuleBasic) RegisterRESTRoutes(ctx client.Context, rtr *mux.Router) {}

// Get the root query command of this module
func (AppModuleBasic) GetQueryCmd(ctx client.Context) *cobra.Command { return nil }

// Get the root tx command of this module
func (AppModuleBasic) GetTxCmd(ctx client.Context) *cobra.Command { return nil }

type AppModule struct {
	AppModuleBasic
}

// NewAppModule creates a new AppModule Object
func NewAppModule(k Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
	}
}

func (AppModule) Name() string { return ModuleName }

func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {}

func (am AppModule) Route() sdk.Route { return sdk.NewRoute(ModuleName, am.NewHandler()) }

func (am AppModule) NewHandler() sdk.Handler { return nil }
func (am AppModule) QuerierRoute() string    { return "" }

func (am AppModule) NewQuerierHandler() sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		return nil, nil
	}
}

func (am AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return nil
}

func (am AppModule) RegisterQueryService(grpc.Server) {}

func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, data json.RawMessage) []abci.ValidatorUpdate {
	return nil
}

func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONMarshaler) json.RawMessage { return nil }
