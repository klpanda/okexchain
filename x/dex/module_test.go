package dex

import (
	"github.com/cosmos/cosmos-sdk/server/api"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/okex/okchain/x/common/version"
	"github.com/okex/okchain/x/dex/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestAppModule_Smoke(t *testing.T) {
	_, _, spKeeper, dexKeeper, ctx, cliCtx := getMockTestCaseEvn(t)

	//func NewAppModule(version ProtocolVersionType, keeper Keeper, supplyKeeper BankKeeper) AppModule {
	appModule := NewAppModule(version.CurrentProtocolVersion, dexKeeper, spKeeper, spKeeper)

	// Const Info
	require.True(t, appModule.Name() == ModuleName)
	require.True(t, appModule.Route().Path() == RouterKey)
	require.True(t, appModule.QuerierRoute() == QuerierRoute)
	require.True(t, appModule.GetQueryCmd(cliCtx) != nil)
	require.True(t, appModule.GetTxCmd(cliCtx) != nil)

	// RegisterCodec
	appModule.RegisterCodec(codec.New())

	appModule.RegisterInvariants(MockInvariantRegistry{})
	rs := api.New(cliCtx, nil)
	appModule.RegisterRESTRoutes(rs.ClientCtx, rs.Router)
	handler := appModule.NewHandler()
	require.True(t, handler != nil)
	querior := appModule.NewQuerierHandler()
	require.True(t, querior != nil)

	// Initialization for genesis
	defaultGen := appModule.DefaultGenesis(cliCtx.Codec)
	err := appModule.ValidateGenesis(cliCtx.Codec, defaultGen)
	require.True(t, err == nil)

	illegalData := []byte{}
	err = appModule.ValidateGenesis(cliCtx.Codec, illegalData)
	require.Error(t, err)

	validatorUpdates := appModule.InitGenesis(ctx, cliCtx.Codec, defaultGen)
	require.True(t, len(validatorUpdates) == 0)
	exportedGenesis := appModule.ExportGenesis(ctx, cliCtx.Codec)
	require.True(t, exportedGenesis != nil)

	// Begin Block
	appModule.BeginBlock(ctx, abci.RequestBeginBlock{})

	// EndBlock : add data for execute in EndBlock
	tokenPair := GetBuiltInTokenPair()
	withdrawInfo := types.WithdrawInfo{
		Owner:    tokenPair.Owner,
		Deposits: tokenPair.Deposits,
	}
	dexKeeper.SetWithdrawInfo(ctx, withdrawInfo)
	dexKeeper.SetWithdrawCompleteTimeAddress(ctx, ctx.BlockHeader().Time, tokenPair.Owner)

	// fail case : failed to SendCoinsFromModuleToAccount return error
	spKeeper.behaveEvil = true
	appModule.EndBlock(ctx, abci.RequestEndBlock{})

	// successful case : success to SendCoinsFromModuleToAccount which return nil
	spKeeper.behaveEvil = false
	appModule.EndBlock(ctx, abci.RequestEndBlock{})
}

type MockInvariantRegistry struct{}

func (ir MockInvariantRegistry) RegisterRoute(moduleName, route string, invar sdk.Invariant) {}
