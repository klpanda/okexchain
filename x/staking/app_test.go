package staking

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server/api"
	"testing"

	"github.com/okex/okchain/x/staking/keeper"
	"github.com/okex/okchain/x/staking/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mock"
)

// getMockApp returns an initialized mock application for this module.
func getMockApp(t *testing.T) (*mock.App, keeper.MockStakingKeeper) {
	mApp := mock.NewApp()

	//RegisterCodec(mApp.Cdc)

	_, accKeeper, mKeeper := CreateTestInput(t, false, SufficientInitPower)

	mApp.Router().AddRoute(sdk.NewRoute(RouterKey, NewHandler(mKeeper.Keeper)))
	mApp.SetEndBlocker(getEndBlocker(mKeeper.Keeper))
	mApp.SetInitChainer(getInitChainer(mApp, mKeeper.Keeper, accKeeper, mKeeper.BankKeeper))

	require.NoError(t, mApp.CompleteSetup(mKeeper.StoreKey, mKeeper.TkeyStoreKey))
	return mApp, mKeeper
}

// getEndBlocker returns a staking endblocker.
func getEndBlocker(keeper Keeper) sdk.EndBlocker {
	return func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
		validatorUpdates := EndBlocker(ctx, keeper)

		return abci.ResponseEndBlock{
			ValidatorUpdates: validatorUpdates,
		}
	}
}

// getInitChainer initializes the chainer of the mock app and sets the genesis
// state. It returns an empty ResponseInitChain.
func getInitChainer(mapp *mock.App, keeper Keeper, accKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper) sdk.InitChainer {
	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		mapp.InitChainer(ctx, req)

		stakingGenesis := DefaultGenesisState()
		validators := InitGenesis(ctx, keeper, accKeeper, bankKeeper, stakingGenesis)

		return abci.ResponseInitChain{
			Validators: validators,
		}
	}
}

type MockInvariantRegistry struct{}

func (ir MockInvariantRegistry) RegisterRoute(moduleName, route string, invar sdk.Invariant) {}

//__________________________________________________________________________________________

func TestAppSmoke(t *testing.T) {
	mApp, mKeeper := getMockApp(t)
	appModule := NewAppModule(mKeeper.Keeper, mKeeper.AccKeeper, mKeeper.BankKeeper)

	cliCtx := client.NewContext().WithCodec(mApp.Cdc).WithJSONMarshaler(mApp.AppCodec)
	// Const Info
	require.True(t, appModule.Name() == ModuleName)
	require.True(t, appModule.Route().Path() == RouterKey)
	require.True(t, appModule.QuerierRoute() == QuerierRoute)
	require.True(t, appModule.GetQueryCmd(cliCtx) != nil)
	require.True(t, appModule.GetTxCmd(cliCtx) != nil)

	appModule.RegisterCodec(mApp.Cdc)
	appModule.RegisterInvariants(MockInvariantRegistry{})
	rs := api.New(cliCtx, nil)
	appModule.RegisterRESTRoutes(rs.ClientCtx, rs.Router)
	handler := appModule.NewHandler()
	require.True(t, handler != nil)
	querior := appModule.NewQuerierHandler()
	require.True(t, querior != nil)

	// Extra Helper
	appModule.CreateValidatorMsgHelpers("0.0.0.0")
	//txBldr := authtypes.NewTxBuilderFromCLI().WithTxEncoder(authclient.GetTxEncoder(mApp.Cdc))
	//appModule.BuildCreateValidatorMsg(cliCtx, txBldr)

	// Initialization for genesis
	defaultGen := appModule.DefaultGenesis(mApp.AppCodec)
	err := appModule.ValidateGenesis(mApp.AppCodec, defaultGen)
	require.True(t, err == nil)

	illegalData := []byte{}
	err = appModule.ValidateGenesis(mApp.AppCodec, illegalData)
	require.Error(t, err)

	// Basic abci test
	header := abci.Header{ChainID: keeper.TestChainID, Height: 0}
	ctx := sdk.NewContext(mKeeper.MountedStore, header, false, log.NewNopLogger())
	validatorUpdates := appModule.InitGenesis(ctx, mApp.AppCodec, defaultGen)
	require.True(t, len(validatorUpdates) == 0)
	exportedGenesis := appModule.ExportGenesis(ctx, mApp.AppCodec)
	require.True(t, exportedGenesis != nil)

	// Begin & End Block
	appModule.BeginBlock(ctx, abci.RequestBeginBlock{})
	appModule.EndBlock(ctx, abci.RequestEndBlock{})

}
