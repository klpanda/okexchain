package dex

import (
	"github.com/cosmos/cosmos-sdk/client"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"testing"

	"github.com/okex/okchain/x/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/okex/okchain/x/dex/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func getMockTestCaseEvn(t *testing.T) (mApp *mockApp,
	tkKeeper *mockTokenKeeper, spKeeper *mockBankKeeper, dexKeeper *mockDexKeeper, testContext sdk.Context, cliCtx client.Context) {
	fakeTokenKeeper := newMockTokenKeeper()
	fakeBankKeeper := newMockBankKeeper()

	mApp, mockDexKeeper, err := newMockApp(fakeTokenKeeper, fakeBankKeeper, 10)
	require.Nil(t, err)

	mApp.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 2}})
	ctx := mApp.BaseApp.NewContext(false, abci.Header{})
	clientCtx := client.NewContext().WithJSONMarshaler(mApp.AppCodec).
		WithAccountRetriever(authtypes.NewAccountRetriever(mApp.AppCodec)).
		WithCodec(mApp.Cdc)

	return mApp, fakeTokenKeeper, fakeBankKeeper, mockDexKeeper, ctx, clientCtx
}

func TestHandler_HandleMsgList(t *testing.T) {
	mApp, tkKeeper, spKeeper, mDexKeeper, ctx, _ := getMockTestCaseEvn(t)

	address := mApp.GenesisAccounts[0].GetAddress()
	listMsg := NewMsgList(address, "btc", common.NativeToken, sdk.NewDec(10))

	handlerFunctor := NewHandler(mApp.dexKeeper)

	// fail case : failed to list because token is invalid
	tkKeeper.exist = false
	badResult, err  := handlerFunctor(ctx, &listMsg)
	require.NotNil(t, err)

	// fail case : failed to list because tokenpair has been exist
	tkKeeper.exist = true
	badResult, err = handlerFunctor(ctx, &listMsg)
	require.NotNil(t, err)
	require.True(t, badResult.Events == nil)

	// fail case : failed to list because SendCoinsFromModuleToAccount return error
	tkKeeper.exist = true
	mDexKeeper.getFakeTokenPair = false
	spKeeper.behaveEvil = true
	badResult, err = handlerFunctor(ctx, &listMsg)
	require.NotNil(t, err)
	// successful case
	tkKeeper.exist = true
	spKeeper.behaveEvil = false
	mDexKeeper.getFakeTokenPair = false
	goodResult, err := handlerFunctor(ctx, &listMsg)
	require.Nil(t, err)
	require.True(t, goodResult.Events != nil)
}

func TestHandler_HandleMsgDeposit(t *testing.T) {
	mApp, _, _, mDexKeeper, ctx, _ := getMockTestCaseEvn(t)
	builtInTP := GetBuiltInTokenPair()
	depositMsg := NewMsgDeposit(builtInTP.Name(),
		sdk.NewDecCoin(builtInTP.QuoteAssetSymbol, sdk.NewInt(100)), builtInTP.Owner)

	handlerFunctor := NewHandler(mApp.dexKeeper)

	// Case1: failed to deposit
	mDexKeeper.failToDeposit = true
	_, err  := handlerFunctor(ctx, &depositMsg)
	require.NotNil(t, err)

	// Case2: success to deposit
	mDexKeeper.failToDeposit = false
	good1, err := handlerFunctor(ctx, &depositMsg)
	require.Nil(t, err)
	require.True(t, good1.Events != nil)
}

func TestHandler_HandleMsgWithdraw(t *testing.T) {
	mApp, _, _, mDexKeeper, ctx, _ := getMockTestCaseEvn(t)
	builtInTP := GetBuiltInTokenPair()
	withdrawMsg := NewMsgWithdraw(builtInTP.Name(),
		sdk.NewDecCoin(builtInTP.QuoteAssetSymbol, sdk.NewInt(100)), builtInTP.Owner)

	handlerFunctor := NewHandler(mApp.dexKeeper)

	// Case1: failed to deposit
	mDexKeeper.failToWithdraw = true
	_, err := handlerFunctor(ctx, &withdrawMsg)
	require.NotNil(t, err)
	// Case2: success to deposit
	mDexKeeper.failToWithdraw = false
	good1 , err := handlerFunctor(ctx, &withdrawMsg)
	require.Nil(t, err)
	require.True(t, good1.Events != nil)
}

func TestHandler_HandleMsgBad(t *testing.T) {
	mApp, _, _, _, ctx, _ := getMockTestCaseEvn(t)
	handlerFunctor := NewHandler(mApp.dexKeeper)

	_, err := handlerFunctor(ctx, sdk.NewTestMsg())
	require.Nil(t, err)
}

func TestHandler_handleMsgTransferOwnership(t *testing.T) {
	mApp, _, spKeeper, mDexKeeper, ctx, _ := getMockTestCaseEvn(t)

	tokenPair := GetBuiltInTokenPair()
	err := mDexKeeper.SaveTokenPair(ctx, tokenPair)
	require.Nil(t, err)
	handlerFunctor := NewHandler(mApp.dexKeeper)
	to := mApp.GenesisAccounts[0].GetAddress()

	// successful case
	msgTransferOwnership := types.NewMsgTransferOwnership(tokenPair.Owner, to, tokenPair.Name(), nil, nil)
	spKeeper.behaveEvil = false
	handlerFunctor(ctx, &msgTransferOwnership)

	// fail case : failed to TransferOwnership because product is not exist
	msgFailedTransferOwnership := types.NewMsgTransferOwnership(tokenPair.Owner, to, "no-product", nil, nil)
	spKeeper.behaveEvil = false
	handlerFunctor(ctx, &msgFailedTransferOwnership)

	// fail case : failed to SendCoinsFromModuleToAccount return error
	msgFailedTransferOwnership = types.NewMsgTransferOwnership(tokenPair.Owner, to, tokenPair.Name(), nil, nil)
	spKeeper.behaveEvil = true
	handlerFunctor(ctx, &msgFailedTransferOwnership)
}
