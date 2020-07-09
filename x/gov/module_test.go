package gov

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server/api"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkGovClient "github.com/cosmos/cosmos-sdk/x/gov/client"
	sdkGovRest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"
	"github.com/okex/okchain/x/common/version"
	"github.com/okex/okchain/x/gov/keeper"
	"github.com/okex/okchain/x/gov/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestAppModule_BeginBlock(t *testing.T) {

}

func getCmdSubmitProposal(cliCtx client.Context) *cobra.Command {
	return &cobra.Command{}
}

func proposalRESTHandler(cliCtx client.Context) sdkGovRest.ProposalRESTHandler {
	return sdkGovRest.ProposalRESTHandler{}
}

func TestNewAppModuleBasic(t *testing.T) {
	ctx, ak, gk, _, crisisKeeper := keeper.CreateTestInput(t, false, 1000)

	moduleBasic := NewAppModuleBasic(sdkGovClient.ProposalHandler{
		CLIHandler:  getCmdSubmitProposal,
		RESTHandler: proposalRESTHandler,
	})

	require.Equal(t, types.ModuleName, moduleBasic.Name())

	cdc := codec.New()
	moduleBasic.RegisterCodec(cdc)
	bz, err := cdc.MarshalBinaryBare(types.MsgSubmitProposal{})
	require.NotNil(t, bz)
	require.Nil(t, err)

	cliCtx := client.NewContext().WithCodec(cdc)
	jsonMsg := moduleBasic.DefaultGenesis(cdc)
	err = moduleBasic.ValidateGenesis(cdc, jsonMsg)
	require.Nil(t, err)
	err = moduleBasic.ValidateGenesis(cdc, jsonMsg[:len(jsonMsg)-1])
	require.NotNil(t, err)

	rs := api.New(cliCtx, nil)
	moduleBasic.RegisterRESTRoutes(rs.ClientCtx, rs.Router)

	// todo: check diff after GetTxCmd
	moduleBasic.GetTxCmd(cliCtx)

	// todo: check diff after GetQueryCmd
	moduleBasic.GetQueryCmd(cliCtx)

	appModule := NewAppModule(version.CurrentProtocolVersion, gk, ak, gk.BankKeeper())
	require.Equal(t, types.ModuleName, appModule.Name())

	// todo: check diff after RegisterInvariants
	appModule.RegisterInvariants(&crisisKeeper)

	require.Equal(t, RouterKey, appModule.Route())

	require.IsType(t, NewHandler(gk), appModule.NewHandler())

	require.Equal(t, QuerierRoute, appModule.QuerierRoute())

	require.IsType(t, NewQuerier(gk.Keeper), appModule.NewQuerierHandler())

	require.Equal(t, []abci.ValidatorUpdate{}, appModule.InitGenesis(ctx, cliCtx.Codec, jsonMsg))

	require.Equal(t, jsonMsg, appModule.ExportGenesis(ctx, cliCtx.Codec))

	appModule.BeginBlock(ctx, abci.RequestBeginBlock{})

	appModule.EndBlock(ctx, abci.RequestEndBlock{})
}
