package poolswap

import (
	"fmt"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/okex/okchain/x/poolswap/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/okex/okchain/x/token"
)

type MockApp struct {
	*mock.App

	keySwap  *sdk.KVStoreKey
	keyToken *sdk.KVStoreKey
	keyLock  *sdk.KVStoreKey
	keyAcc   *sdk.KVStoreKey
	keyBank  *sdk.KVStoreKey

	swapKeeper   Keeper
	tokenKeeper  token.Keeper
}

func registerCodec(cdc *codec.Codec) {
	RegisterCodec(cdc)
	token.RegisterCodec(cdc)
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
		keySwap:  sdk.NewKVStoreKey(StoreKey),
		keyToken: sdk.NewKVStoreKey(token.StoreKey),
		keyLock:  sdk.NewKVStoreKey(token.KeyLock),
		keyAcc:   sdk.NewKVStoreKey(authtypes.StoreKey),
		keyBank:  sdk.NewKVStoreKey(banktypes.StoreKey),
	}

	feeCollector := authtypes.NewEmptyModuleAccount(authtypes.FeeCollectorName)
	blacklistedAddrs := make(map[string]bool)
	blacklistedAddrs[feeCollector.String()] = true

	maccPerms := map[string][]string{
		authtypes.FeeCollectorName: nil,
		token.ModuleName:      {authtypes.Minter, authtypes.Burner},
		ModuleName:            {authtypes.Minter, authtypes.Burner},
	}

	mockApp.AccountKeeper = authkeeper.NewAccountKeeper(mockApp.AppCodec, mockApp.keyAcc,
		mockApp.ParamsKeeper.Subspace(authtypes.ModuleName),
		authtypes.ProtoBaseAccount, maccPerms)

	mockApp.BankKeeper = bankkeeper.NewBaseKeeper(mockApp.AppCodec, mockApp.keyBank, mockApp.AccountKeeper,
		mockApp.ParamsKeeper.Subspace(banktypes.ModuleName), blacklistedAddrs)

	mockApp.tokenKeeper = token.NewKeeper(
		mockApp.BankKeeper,
		mockApp.ParamsKeeper.Subspace(token.DefaultParamspace),
		authtypes.FeeCollectorName,
		mockApp.AccountKeeper,
		mockApp.keyToken,
		mockApp.keyLock,
		mockApp.Cdc,
		true)

	mockApp.swapKeeper = NewKeeper(
		mockApp.BankKeeper,
		mockApp.tokenKeeper,
		mockApp.Cdc,
		mockApp.keySwap,
		mockApp.ParamsKeeper.Subspace(DefaultParamspace),
	)

	mockApp.Router().AddRoute(sdk.NewRoute(RouterKey, NewHandler(mockApp.swapKeeper)))
	mockApp.QueryRouter().AddRoute(QuerierRoute, NewQuerier(mockApp.swapKeeper))

	mockApp.SetBeginBlocker(getBeginBlocker(mockApp.swapKeeper))
	mockApp.SetEndBlocker(getEndBlocker(mockApp.swapKeeper))

	decCoins, err := sdk.ParseDecCoins(fmt.Sprintf("%d%s,%d%s,%d%s,%d%s",
		balance, types.TestQuotePooledToken, balance, types.TestBasePooledToken, balance, types.TestBasePooledToken2, balance, types.TestBasePooledToken3))
	require.Nil(t, err)
	coins := decCoins

	keysSlice, genAccs, genBals := CreateGenAccounts(numGenAccs, coins)
	addrKeysSlice = keysSlice

	mockApp.SetInitChainer(getInitChainer(mockApp.App, []authtypes.ModuleAccountI{feeCollector}))

	// todo: checkTx in mock app
	mockApp.SetAnteHandler(nil)

	app := mockApp
	require.NoError(t, app.CompleteSetup(
		app.keySwap,
		app.keyToken,
		app.keyLock,
		app.keyBank,
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

func getInitChainer(mapp *mock.App,
	blacklistedAddrs []authtypes.ModuleAccountI) sdk.InitChainer {
	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		mapp.InitChainer(ctx, req)
		// set module accounts
		for _, macc := range blacklistedAddrs {
			mapp.AccountKeeper.SetModuleAccount(ctx, macc)
		}
		return abci.ResponseInitChain{}
	}
}

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
