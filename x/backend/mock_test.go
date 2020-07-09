package backend

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server/api"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/okex/okchain/x/backend/types"
	"os"
	"testing"
	"time"

	"github.com/okex/okchain/x/backend/client/cli"
	"github.com/okex/okchain/x/backend/config"
	"github.com/okex/okchain/x/backend/orm"
	"github.com/okex/okchain/x/common/monitor"
	"github.com/okex/okchain/x/common/version"
	"github.com/okex/okchain/x/order/keeper"
	ordertypes "github.com/okex/okchain/x/order/types"
	tokentypes "github.com/okex/okchain/x/token/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/okex/okchain/x/common"

	//"github.com/okex/okchain/x/gov"
	"github.com/okex/okchain/x/dex"
	"github.com/okex/okchain/x/order"

	//"github.com/okex/okchain/x/staking"
	"github.com/okex/okchain/x/token"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

type MockApp struct {
	*mock.App

	keyOrder *sdk.KVStoreKey

	keyToken     *sdk.KVStoreKey
	keyLock      *sdk.KVStoreKey
	keyDex       *sdk.KVStoreKey
	keyTokenPair *sdk.KVStoreKey

	keyAcc  *sdk.KVStoreKey
	keyBank *sdk.KVStoreKey

	orderKeeper   keeper.Keeper
	dexKeeper     dex.Keeper
	tokenKeeper   token.Keeper
	backendKeeper Keeper
}

func registerCdc(cdc *codec.Codec) {
	authtypes.RegisterCodec(cdc)
}

// initialize the mock application for this module
func getMockApp(t *testing.T, numGenAccs int, enableBackend bool, dbDir string) (mockApp *MockApp, addrKeysSlice mock.AddrKeysSlice) {
	mapp := mock.NewApp()
	registerCdc(mapp.Cdc)

	mockApp = &MockApp{
		App:          mapp,
		keyOrder:     sdk.NewKVStoreKey(ordertypes.OrderStoreKey),
		keyToken:     sdk.NewKVStoreKey(tokentypes.ModuleName),
		keyLock:      sdk.NewKVStoreKey(tokentypes.KeyLock),
		keyDex:       sdk.NewKVStoreKey(dex.StoreKey),
		keyTokenPair: sdk.NewKVStoreKey(dex.TokenPairStoreKey),
		keyAcc:       sdk.NewKVStoreKey(authtypes.StoreKey),
		keyBank:      sdk.NewKVStoreKey(banktypes.StoreKey),
	}

	feeCollector := authtypes.NewEmptyModuleAccount(authtypes.FeeCollectorName)
	blacklistedAddrs := make(map[string]bool)
	blacklistedAddrs[feeCollector.String()] = true

	maccPerms := map[string][]string{
		authtypes.FeeCollectorName: nil,
		token.ModuleName:           {authtypes.Minter, authtypes.Burner},
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
		//mockApp.keyTokenPair,
		mockApp.Cdc,
		true,
	)

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

	mockApp.orderKeeper = keeper.NewKeeper(
		mockApp.tokenKeeper,
		mockApp.AccountKeeper,
		mockApp.BankKeeper,
		mockApp.dexKeeper,
		mockApp.ParamsKeeper.Subspace(ordertypes.DefaultParamspace),
		authtypes.FeeCollectorName,
		mockApp.keyOrder,
		mockApp.Cdc,
		true,
		monitor.NopOrderMetrics())

	// CleanUp data
	cfg, err := config.SafeLoadMaintainConfig(config.DefaultTestConfig)
	require.Nil(t, err)
	cfg.EnableBackend = enableBackend
	cfg.EnableMktCompute = enableBackend
	cfg.OrmEngine.EngineType = orm.EngineTypeSqlite
	cfg.OrmEngine.ConnectStr = config.DefaultTestDataHome + "/sqlite3/backend.db"
	if dbDir == "" {
		path := config.DefaultTestDataHome + "/sqlite3"
		if err := os.RemoveAll(path); err != nil {
			mockApp.Logger().Debug(err.Error())
		}
	} else {
		cfg.LogSQL = false
		cfg.OrmEngine.ConnectStr = dbDir + "/backend.db"
	}

	mockApp.backendKeeper = NewKeeper(
		mockApp.orderKeeper,
		mockApp.tokenKeeper,
		&mockApp.dexKeeper,
		nil,
		mockApp.Cdc,
		mockApp.Logger(),
		cfg)

	mockApp.Router().AddRoute(sdk.NewRoute(ordertypes.RouterKey, order.NewOrderHandler(mockApp.orderKeeper)))
	mockApp.QueryRouter().AddRoute(ordertypes.QuerierRoute, keeper.NewQuerier(mockApp.orderKeeper))
	//mockApp.Router().AddRoute(token.RouterKey, token.NewHandler(mockApp.tokenKeeper))
	mockApp.Router().AddRoute(sdk.NewRoute(token.RouterKey, token.NewTokenHandler(mockApp.tokenKeeper, version.ProtocolVersionV0)))
	mockApp.QueryRouter().AddRoute(token.QuerierRoute, token.NewQuerier(mockApp.tokenKeeper))

	intQuantity := 100000
	coins, _ := sdk.ParseDecCoins(fmt.Sprintf("%d%s,%d%s",
		intQuantity, common.NativeToken, intQuantity, common.TestToken))
	keysSlice, genAccs, genBals := CreateGenAccounts(numGenAccs, coins)
	mockApp.SetEndBlocker(getEndBlocker(mockApp.orderKeeper, mockApp.backendKeeper))
	mockApp.SetInitChainer(getInitChainer(mockApp.App, mockApp.AccountKeeper,
		[]authtypes.ModuleAccountI{feeCollector}))

	addrKeysSlice = keysSlice

	// todo: checkTx in mock app
	mockApp.SetAnteHandler(nil)

	app := mockApp
	mockApp.MountStores(
		//app.keyOrder,
		app.keyToken,
		app.keyTokenPair,
		app.keyLock,
		app.keyBank,
		app.keyDex,
	)

	require.NoError(t, mockApp.CompleteSetup(mockApp.keyOrder))
	mock.SetGenesis(mockApp.App, genAccs, genBals)

	for i := 0; i < numGenAccs; i++ {
		mock.CheckBalance(t, app.App, keysSlice[i].Address, coins)
		mockApp.TotalCoinsSupply = mockApp.TotalCoinsSupply.Add(coins...)
	}
	return
}

func getEndBlocker(orderKeeper keeper.Keeper, backendKeeper Keeper) sdk.EndBlocker {
	return func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
		order.EndBlocker(ctx, orderKeeper)
		EndBlocker(ctx, backendKeeper)
		return abci.ResponseEndBlock{}
	}
}

func getInitChainer(mapp *mock.App, accKeeper authkeeper.AccountKeeper,
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

func buildTx(app *MockApp, ctx sdk.Context, addrKeys mock.AddrKeys, msg []sdk.Msg) authtypes.StdTx {
	accs := app.AccountKeeper.GetAccount(ctx, addrKeys.Address)
	accNum := accs.GetAccountNumber()
	seqNum := accs.GetSequence()

	tx := mock.GenTx(msg, []uint64{uint64(accNum)}, []uint64{uint64(seqNum)}, addrKeys.PrivKey)
	_, _, err := app.Check(tx)
	if err != nil {
		panic("something wrong in checking transaction")
	}
	return tx
}

func mockApplyBlock(app *MockApp, ctx sdk.Context, txs []authtypes.StdTx) {
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: ctx.BlockHeight()}})

	orderParam := ordertypes.DefaultParams()
	app.orderKeeper.SetParams(ctx, &orderParam)
	tokenParam := tokentypes.DefaultParams()
	app.tokenKeeper.SetParams(ctx, tokenParam)
	for i, tx := range txs {
		_, response, err := app.Deliver(tx)
		if err == nil {
			txBytes, _ := authtypes.DefaultTxEncoder(app.Cdc)(tx)
			txHash := fmt.Sprintf("%X", tmhash.Sum(txBytes))
			app.Logger().Info(fmt.Sprintf("[Sync Tx(%s) to backend module]", txHash))
			app.backendKeeper.SyncTx(ctx, &txs[i], txHash, ctx.BlockHeader().Time.Unix()) // do not use tx
		} else {
			app.Logger().Error(fmt.Sprintf("DeliverTx failed: %v", response))
		}
	}

	app.EndBlock(abci.RequestEndBlock{Height: ctx.BlockHeight()})
	app.Commit()
}

func CreateGenAccounts(numAccs int, genCoins sdk.Coins) (addrKeysSlice mock.AddrKeysSlice, genAccs []authtypes.BaseAccount, genBals []banktypes.Balance) {
	for i := 0; i < numAccs; i++ {
		privKey := secp256k1.GenPrivKey()
		pubKey := privKey.PubKey()
		addr := sdk.AccAddress(pubKey.Address())

		addrKeys := mock.NewAddrKeys(addr, pubKey, privKey)
		account := authtypes.BaseAccount{
			Address: addr,
		}
		genAccs = append(genAccs, account)
		genBals = append(genBals, banktypes.Balance{
			Address: addr,
			Coins:   genCoins,
		})
		addrKeysSlice = append(addrKeysSlice, addrKeys)
	}
	return
}

func mockOrder(orderID, product, side, price, quantity string) *ordertypes.Order {
	return &ordertypes.Order{
		OrderID:           orderID,
		Product:           product,
		Side:              side,
		Price:             sdk.MustNewDecFromStr(price),
		FilledAvgPrice:    sdk.ZeroDec(),
		Quantity:          sdk.MustNewDecFromStr(quantity),
		RemainQuantity:    sdk.MustNewDecFromStr(quantity),
		Status:            ordertypes.OrderStatusOpen,
		OrderExpireBlocks: ordertypes.DefaultOrderExpireBlocks,
		FeePerBlock:       ordertypes.DefaultFeePerBlock,
	}
}

func FireEndBlockerPeriodicMatch(t *testing.T, enableBackend bool) (mockDexApp *MockApp, orders []*ordertypes.Order) {
	mapp, addrKeysSlice := getMockApp(t, 2, enableBackend, "")
	mapp.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 2}})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{Time: time.Now()}).WithBlockHeight(10)
	mapp.BankKeeper.SetSupply(ctx, banktypes.NewSupply(mapp.TotalCoinsSupply))
	feeParams := ordertypes.DefaultParams()
	mapp.orderKeeper.SetParams(ctx, &feeParams)
	tokenPair := dex.GetBuiltInTokenPair()

	err := mapp.dexKeeper.SaveTokenPair(ctx, tokenPair)
	require.Nil(t, err)
	// mock orders
	orders = []*ordertypes.Order{
		mockOrder("", types.TestTokenPair, types.BuyOrder, "10.0", "1.0"),
		mockOrder("", types.TestTokenPair, types.SellOrder, "10.0", "1.5"),
	}
	orders[0].Sender = addrKeysSlice[0].Address
	orders[1].Sender = addrKeysSlice[1].Address
	for i := 0; i < 2; i++ {
		err := mapp.orderKeeper.PlaceOrder(ctx, orders[i])
		require.NoError(t, err)
	}

	// call EndBlocker to execute periodic match

	order.EndBlocker(ctx, mapp.orderKeeper)
	EndBlocker(ctx, mapp.backendKeeper)
	return mapp, orders
}

func TestAppModule(t *testing.T) {
	mapp, _ := getMockApp(t, 2, false, "")
	mapp.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 2}})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{Time: time.Now()}).WithBlockHeight(10)
	clientCtx := client.NewContext().WithJSONMarshaler(mapp.AppCodec).
		WithAccountRetriever(authtypes.NewAccountRetriever(mapp.AppCodec)).
		WithCodec(mapp.Cdc)

	app := NewAppModule(mapp.backendKeeper)

	require.Equal(t, true, app.InitGenesis(ctx, mapp.AppCodec, nil) == nil)
	require.Equal(t, nil, app.ValidateGenesis(mapp.AppCodec, nil))
	require.Equal(t, true, app.DefaultGenesis(mapp.AppCodec) == nil)
	require.Equal(t, true, app.ExportGenesis(ctx, mapp.AppCodec) == nil)
	require.Equal(t, true, app.NewHandler() == nil)
	require.Equal(t, true, app.GetTxCmd(clientCtx) == nil)
	require.EqualValues(t, cli.GetQueryCmd(QuerierRoute, mapp.Cdc).Name(), app.GetQueryCmd(clientCtx).Name())
	require.Equal(t, ModuleName, app.Name())
	require.Equal(t, ModuleName, app.AppModuleBasic.Name())
	require.Equal(t, true, app.NewQuerierHandler() != nil)
	require.Equal(t, RouterKey, app.Route())
	require.Equal(t, QuerierRoute, app.QuerierRoute())
	require.Equal(t, true, app.EndBlock(ctx, abci.RequestEndBlock{}) == nil)

	rs := api.New(clientCtx, nil)
	app.RegisterRESTRoutes(rs.ClientCtx, rs.Router)
}
