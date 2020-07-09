package genutil

import (
	"encoding/json"

	"github.com/okex/okchain/x/genutil/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	_ module.AppModuleGenesis = AppModule{}
	_ module.AppModuleBasic   = AppModuleBasic{}
)

// AppModuleBasic is the struct of app module basics object
type AppModuleBasic struct{}

// Name returns the module name
func (AppModuleBasic) Name() string {
	return ModuleName
}

// DefaultGenesis returns the default genesis state in json raw message
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONMarshaler) json.RawMessage {
	return cdc.MustMarshalJSON(GenesisState{})
}

// ValidateGenesis gives a quick validity check for module genesis
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONMarshaler, bz json.RawMessage) error {
	var data GenesisState
	err := cdc.UnmarshalJSON(bz, &data)
	if err != nil {
		return err
	}
	return ValidateGenesis(data)
}

// nolint
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec)                         {}
func (AppModuleBasic) RegisterRESTRoutes(_ client.Context, _ *mux.Router) {}
func (AppModuleBasic) GetTxCmd(_ client.Context) *cobra.Command                 { return nil }
func (AppModuleBasic) GetQueryCmd(_ client.Context) *cobra.Command              { return nil }

// AppModule is the struct of this app module
type AppModule struct {
	AppModuleBasic
	accountKeeper types.AccountKeeper
	stakingKeeper types.StakingKeeper
	deliverTx     deliverTxfn
}

// NewAppModule creates a new AppModule object
func NewAppModule(accountKeeper types.AccountKeeper,
	stakingKeeper types.StakingKeeper, deliverTx deliverTxfn) module.AppModule {

	return module.NewGenesisOnlyAppModule(AppModule{
		AppModuleBasic: AppModuleBasic{},
		accountKeeper:  accountKeeper,
		stakingKeeper:  stakingKeeper,
		deliverTx:      deliverTx,
	})
}

// InitGenesis initializes the module genesis state
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	return InitGenesis(ctx, ModuleCdc, am.stakingKeeper, am.deliverTx, genesisState)
}

// ExportGenesis exports the module genesis state
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONMarshaler) json.RawMessage {
	return nil
}
