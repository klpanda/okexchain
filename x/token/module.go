package token

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/gogo/protobuf/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/okex/okchain/x/common/version"
	tokenTypes "github.com/okex/okchain/x/token/types"
)

var (
	_ module.AppModule = AppModule{}
)

// AppModule app module
type AppModule struct {
	AppModuleBasic
	keeper       Keeper
	version      version.ProtocolVersionType
}

// NewAppModule creates a new AppModule object
func NewAppModule(v version.ProtocolVersionType, keeper Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         keeper,
		version:        v,
	}
}

// nolint
func (AppModule) Name() string {
	return tokenTypes.ModuleName
}

// nolint
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
}

// Route module message route name
func (am AppModule) Route() sdk.Route {
	return sdk.NewRoute(tokenTypes.RouterKey, NewTokenHandler(am.keeper, am.version))
}

// nolint
func (am AppModule) NewHandler() sdk.Handler {
	return NewTokenHandler(am.keeper, am.version)
}

// nolint
func (AppModule) QuerierRoute() string {
	return tokenTypes.QuerierRoute
}

// nolint
func (am AppModule) NewQuerierHandler() sdk.Querier {
	return NewQuerier(am.keeper)
}

func (am AppModule) RegisterQueryService(grpc.Server) {}

// nolint
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState GenesisState
	tokenTypes.ModuleCdc.MustUnmarshalJSON(data, &genesisState)
	initGenesis(ctx, am.keeper, genesisState)
	return []abci.ValidatorUpdate{}
}

// nolint
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONMarshaler) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return tokenTypes.ModuleCdc.MustMarshalJSON(gs)
}

// nolint
func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	beginBlocker(ctx, am.keeper)
}

// nolint
func (AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
