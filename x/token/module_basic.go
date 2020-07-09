package token

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/client"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"github.com/okex/okchain/x/token/client/cli"
	"github.com/okex/okchain/x/token/client/rest"
	tokenTypes "github.com/okex/okchain/x/token/types"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
)

// nolint
type AppModuleBasic struct{}

// nolint
func (AppModuleBasic) Name() string {
	return tokenTypes.ModuleName
}

// nolint
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	RegisterCodec(cdc)
}

// nolint
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONMarshaler) json.RawMessage {
	return cdc.MustMarshalJSON(defaultGenesisState())
}

// validateGenesis module validate genesis from json raw message
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONMarshaler, bz json.RawMessage) error {
	var data GenesisState
	err := cdc.UnmarshalJSON(bz, &data)
	if err != nil {
		return err
	}
	return validateGenesis(data)
}

// RegisterRESTRoutes register rest routes
func (AppModuleBasic) RegisterRESTRoutes(ctx client.Context, rtr *mux.Router) {
	rest.RegisterRoutes(ctx, rtr, ModuleName)
}

// GetTxCmd gets the root tx command of this module
func (AppModuleBasic) GetTxCmd(ctx client.Context) *cobra.Command {
	return cli.GetTxCmd(tokenTypes.StoreKey, ctx.Codec)
}

// GetQueryCmd gets the root query command of this module
func (AppModuleBasic) GetQueryCmd(ctx client.Context) *cobra.Command {
	return cli.GetQueryCmd(tokenTypes.StoreKey, ctx.Codec)
}
