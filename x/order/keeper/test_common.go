package keeper

import (
	"fmt"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"os"
	"testing"
	"time"

	"github.com/okex/okchain/x/common"

	"github.com/okex/okchain/x/common/monitor"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/okex/okchain/x/params"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/okex/okchain/x/dex"
	"github.com/okex/okchain/x/order/types"
	"github.com/okex/okchain/x/token"
)

var mockOrder = types.MockOrder

// TestInput stores some variables for testing
type TestInput struct {
	Ctx       sdk.Context
	Cdc       *codec.Codec
	TestAddrs []sdk.AccAddress

	OrderKeeper   Keeper
	TokenKeeper   token.Keeper
	AccountKeeper authkeeper.AccountKeeper
	BankKeeper    bankkeeper.BaseKeeper
	DexKeeper     dex.Keeper
}

// MakeTestCodec creates a codec used only for testing
func MakeTestCodec() *codec.Codec {
	var cdc = codec.New()
	banktypes.RegisterCodec(cdc)
	authtypes.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	dex.RegisterCodec(cdc)
	cryptocodec.RegisterCrypto(cdc)

	types.RegisterCodec(cdc) // order
	token.RegisterCodec(cdc) // token
	return cdc
}

// CreateTestInputWithBalance creates TestInput with the number of account and the quantity
func CreateTestInputWithBalance(t *testing.T, numAddrs, initQuantity int64) TestInput {

	db := dbm.NewMemDB()

	keyAcc := sdk.NewKVStoreKey(authtypes.StoreKey)
	keyBank := sdk.NewKVStoreKey(banktypes.StoreKey)
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)

	// order module
	keyOrder := sdk.NewKVStoreKey(types.OrderStoreKey)

	// token module
	keyToken := sdk.NewKVStoreKey(token.StoreKey)
	keyLock := sdk.NewKVStoreKey(token.KeyLock)
	//keyTokenPair := sdk.NewKVStoreKey(token.KeyTokenPair)

	// dex module
	storeKey := sdk.NewKVStoreKey(dex.StoreKey)
	keyTokenPair := sdk.NewKVStoreKey(dex.TokenPairStoreKey)

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyBank, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)

	ms.MountStoreWithDB(keyOrder, sdk.StoreTypeIAVL, db)

	ms.MountStoreWithDB(keyToken, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyLock, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(storeKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyTokenPair, sdk.StoreTypeIAVL, db)

	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	ctx := sdk.NewContext(ms, abci.Header{Time: time.Unix(0, 0)}, false, log.NewTMLogger(os.Stdout))
	cdc := MakeTestCodec()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	appCodec := codec.NewHybridCodec(cdc, interfaceRegistry)

	feeCollectorAcc := authtypes.NewEmptyModuleAccount(authtypes.FeeCollectorName)

	blacklistedAddrs := make(map[string]bool)
	blacklistedAddrs[feeCollectorAcc.String()] = true

	paramsKeeper := params.NewKeeper(appCodec, keyParams, tkeyParams)
	maccPerms := map[string][]string{
		authtypes.FeeCollectorName: nil,
		token.ModuleName:           {authtypes.Minter, authtypes.Burner},
	}
	accountKeeper := authkeeper.NewAccountKeeper(appCodec, keyAcc,
		paramsKeeper.Subspace(authtypes.ModuleName), authtypes.ProtoBaseAccount, maccPerms)
	bankKeeper := bankkeeper.NewBaseKeeper(appCodec, keyBank, accountKeeper, paramsKeeper.Subspace(banktypes.ModuleName), blacklistedAddrs)

	bankKeeper.SetSupply(ctx, banktypes.NewSupply(sdk.Coins{}))

	// set module accounts
	accountKeeper.SetModuleAccount(ctx, feeCollectorAcc)

	// token keeper
	tokenKeepr := token.NewKeeper(bankKeeper, paramsKeeper.Subspace(token.DefaultParamspace),
		authtypes.FeeCollectorName, accountKeeper, keyToken, keyLock, cdc, true)

	// dex keeper
	paramsSubspace := paramsKeeper.Subspace(dex.DefaultParamspace)
	dexKeeper := dex.NewKeeper(authtypes.FeeCollectorName, accountKeeper, paramsSubspace, tokenKeepr, nil, bankKeeper, storeKey, keyTokenPair, cdc)

	// order keeper
	orderKeeper := NewKeeper(tokenKeepr, accountKeeper, bankKeeper, dexKeeper,
		paramsKeeper.Subspace(types.DefaultParamspace), authtypes.FeeCollectorName, keyOrder,
		cdc, true, monitor.NopOrderMetrics())

	defaultParams := types.DefaultTestParams()
	orderKeeper.SetParams(ctx, &defaultParams)

	// init account tokens
	decCoins, err := sdk.ParseDecCoins(fmt.Sprintf("%d%s,%d%s",
		initQuantity, common.NativeToken, initQuantity, common.TestToken))
	require.Nil(t, err)

	initCoins := decCoins

	var testAddrs []sdk.AccAddress
	for i := int64(0); i < numAddrs; i++ {
		pk := ed25519.GenPrivKey().PubKey()
		addr := sdk.AccAddress(pk.Address())
		testAddrs = append(testAddrs, addr)
		//_, err := bankKeeper.AddCoins(ctx, addr, initCoins)
		err := orderKeeper.bankKeeper.MintCoins(ctx, token.ModuleName, initCoins)
		require.Nil(t, err)
		err = orderKeeper.bankKeeper.SendCoinsFromModuleToAccount(ctx, token.ModuleName, addr, initCoins)
		require.Nil(t, err)
	}

	return TestInput{ctx, cdc, testAddrs, orderKeeper, tokenKeepr,
		accountKeeper, bankKeeper, dexKeeper}
}

// CreateTestInput creates TestInput with default params
func CreateTestInput(t *testing.T) TestInput {
	return CreateTestInputWithBalance(t, 2, 100)
}
