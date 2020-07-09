package order

import (
	"fmt"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"testing"

	"github.com/okex/okchain/x/common/monitor"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/okex/okchain/x/staking/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/okex/okchain/x/common"
	"github.com/okex/okchain/x/dex"
	"github.com/okex/okchain/x/token"
)

type MockApp struct {
	*mock.App

	keyOrder     *sdk.KVStoreKey
	keyToken     *sdk.KVStoreKey
	keyLock      *sdk.KVStoreKey
	keyDex       *sdk.KVStoreKey
	keyTokenPair *sdk.KVStoreKey

	keyBank *sdk.KVStoreKey
	keyAcc  *sdk.KVStoreKey

	orderKeeper  Keeper
	tokenKeeper  token.Keeper
	dexKeeper    dex.Keeper
}

func registerCodec(cdc *codec.Codec) {
	RegisterCodec(cdc)
	token.RegisterCodec(cdc)
	authtypes.RegisterCodec(cdc)
}

func getMockApp(t *testing.T, numGenAccs int) (mockApp *MockApp, addrKeysSlice mock.AddrKeysSlice) {
	return getMockAppWithBalance(t, numGenAccs, 100)
}

// initialize the mock application for this module
func getMockAppWithBalance(t *testing.T, numGenAccs int, balance int64) (mockApp *MockApp,
	addrKeysSlice mock.AddrKeysSlice) {
	mapp := mock.NewApp()
	registerCodec(mapp.Cdc)

	mockApp = &MockApp{
		App:      mapp,
		keyOrder: sdk.NewKVStoreKey(OrderStoreKey),

		keyToken:     sdk.NewKVStoreKey(token.StoreKey),
		keyLock:      sdk.NewKVStoreKey(token.KeyLock),
		keyDex:       sdk.NewKVStoreKey(dex.StoreKey),
		keyTokenPair: sdk.NewKVStoreKey(dex.TokenPairStoreKey),

		keyBank: sdk.NewKVStoreKey(banktypes.StoreKey),
		keyAcc:  sdk.NewKVStoreKey(authtypes.StoreKey),
	}

	feeCollector := authtypes.NewEmptyModuleAccount(authtypes.FeeCollectorName)
	blacklistedAddrs := make(map[string]bool)
	blacklistedAddrs[feeCollector.String()] = true

	mockApp.BankKeeper = bankkeeper.NewBaseKeeper(mockApp.AppCodec, mockApp.keyBank, mockApp.AccountKeeper,
		mockApp.ParamsKeeper.Subspace(banktypes.ModuleName), blacklistedAddrs)

	maccPerms := map[string][]string{
		authtypes.FeeCollectorName: nil,
		token.ModuleName:      {authtypes.Minter, authtypes.Burner},
	}
	mockApp.AccountKeeper = authkeeper.NewAccountKeeper(mockApp.AppCodec, mockApp.keyAcc,
		mockApp.ParamsKeeper.Subspace(authtypes.ModuleName), authtypes.ProtoBaseAccount, maccPerms)

	mockApp.tokenKeeper = token.NewKeeper(
		mockApp.BankKeeper,
		mockApp.ParamsKeeper.Subspace(token.DefaultParamspace),
		authtypes.FeeCollectorName,
		mockApp.AccountKeeper,
		mockApp.keyToken,
		mockApp.keyLock,
		mockApp.Cdc,
		true)

	mockApp.dexKeeper = dex.NewKeeper(
		authtypes.FeeCollectorName,
		mockApp.AccountKeeper,
		mockApp.ParamsKeeper.Subspace(dex.DefaultParamspace),
		mockApp.tokenKeeper,
		nil,
		mockApp.BankKeeper,
		mockApp.keyDex,
		mockApp.keyTokenPair,
		mockApp.Cdc)

	mockApp.orderKeeper = NewKeeper(
		mockApp.tokenKeeper,
		mockApp.AccountKeeper,
		mockApp.BankKeeper,
		mockApp.dexKeeper,
		mockApp.ParamsKeeper.Subspace(DefaultParamspace),
		authtypes.FeeCollectorName,
		mockApp.keyOrder,
		mockApp.Cdc,
		true,
		monitor.NopOrderMetrics())

	mockApp.Router().AddRoute(sdk.NewRoute(RouterKey, NewOrderHandler(mockApp.orderKeeper)))
	mockApp.QueryRouter().AddRoute(QuerierRoute, NewQuerier(mockApp.orderKeeper))

	decCoins, err := sdk.ParseDecCoins(fmt.Sprintf("%d%s,%d%s",
		balance, common.NativeToken, balance, common.TestToken))
	require.Nil(t, err)
	coins := decCoins
	keysSlice, genAccs, genBals := CreateGenAccounts(numGenAccs, coins)
	addrKeysSlice = keysSlice

	mockApp.SetBeginBlocker(getBeginBlocker(mockApp.orderKeeper))
	mockApp.SetEndBlocker(getEndBlocker(mockApp.orderKeeper))
	mockApp.SetInitChainer(getInitChainer(mockApp.App, mockApp.AccountKeeper,
		[]authtypes.ModuleAccountI{feeCollector}))

	// todo: checkTx in mock app
	mockApp.SetAnteHandler(nil)

	app := mockApp
	require.NoError(t, app.CompleteSetup(
		app.keyOrder,
		app.keyToken,
		app.keyDex,
		app.keyTokenPair,
		app.keyLock,
		app.keyBank,
		app.keyAcc,
	))
	mock.SetGenesis(mockApp.App, genAccs, genBals)

	for i := 0; i < numGenAccs; i++ {
		mock.CheckBalance(t, app.App, keysSlice[i].Address, coins)
		mockApp.TotalCoinsSupply = mockApp.TotalCoinsSupply.Add(coins...)
	}

	return mockApp, addrKeysSlice
}

func getBeginBlocker(keeper Keeper) sdk.BeginBlocker {
	return func(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
		BeginBlocker(ctx, keeper)
		return abci.ResponseBeginBlock{}
	}
}

func getEndBlocker(keeper Keeper) sdk.EndBlocker {
	return func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
		EndBlocker(ctx, keeper)
		return abci.ResponseEndBlock{}
	}
}

func getInitChainer(mapp *mock.App, accKeeper types.AccountKeeper,
	blacklistedAddrs []authtypes.ModuleAccountI) sdk.InitChainer {
	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		mapp.InitChainer(ctx, req)
		// set module accounts
		for _, macc := range blacklistedAddrs {
			accKeeper.SetModuleAccount(ctx, macc)
		}
		return abci.ResponseInitChain{}
	}
}

//func produceOrderTxs(app *MockApp, ctx sdk.Context, numToGenerate int, addrKeys mock.AddrKeys,
//	orderMsg *MsgNewOrder) []auth.StdTx {
//	txs := make([]auth.StdTx, numToGenerate)
//	orderMsg.Sender = addrKeys.Address
//	for i := 0; i < numToGenerate; i++ {
//		txs[i] = buildTx(app, ctx, addrKeys, *orderMsg)
//	}
//	return txs
//}

//func buildTx(app *MockApp, ctx sdk.Context, addrKeys mock.AddrKeys, msg sdk.Msg) auth.StdTx {
//	accs := app.AccountKeeper.GetAccount(ctx, addrKeys.Address)
//	accNum := accs.GetAccountNumber()
//	seqNum := accs.GetSequence()
//
//	tx := mock.GenTx(
//		[]sdk.Msg{msg}, []uint64{uint64(accNum)}, []uint64{uint64(seqNum)}, addrKeys.PrivKey)
//	res := app.Check(tx)
//	if !res.IsOK() {
//		panic(fmt.Sprintf("something wrong in checking transaction: %v", res))
//	}
//	return tx
//}
//
//func mockApplyBlock(app *MockApp, blockHeight int64, txs []auth.StdTx) {
//	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: blockHeight}})
//
//	newCtx := app.NewContext(false, abci.Header{})
//	param := DefaultTestParams()
//	app.orderKeeper.SetParams(newCtx, &param)
//	for _, tx := range txs {
//		app.Deliver(tx)
//	}
//
//	app.EndBlock(abci.RequestEndBlock{Height: blockHeight})
//	app.Commit()
//}

func CreateGenAccounts(numAccs int, genCoins sdk.Coins) (addrKeysSlice mock.AddrKeysSlice,
	genAccs []authtypes.BaseAccount, genBals []banktypes.Balance) {
	for i := 0; i < numAccs; i++ {
		privKey := secp256k1.GenPrivKey()
		pubKey := privKey.PubKey()
		addr := sdk.AccAddress(pubKey.Address())

		addrKeys := mock.NewAddrKeys(addr, pubKey, privKey)
		account := authtypes.BaseAccount{
			Address: addr,
		}
		genAccs = append(genAccs, account)
		bal := banktypes.Balance{
			Address: addr,
			Coins: genCoins,
		}
		genBals = append(genBals, bal)
		addrKeysSlice = append(addrKeysSlice, addrKeys)
	}
	return
}
