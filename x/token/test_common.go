package token

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/okex/okchain/x/params"
	"github.com/okex/okchain/x/staking"
	"github.com/okex/okchain/x/token/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tm-db"
)

// CreateParam create okchain parm for test
func CreateParam(t *testing.T, isCheckTx bool) (sdk.Context, Keeper, *sdk.KVStoreKey, []byte) {
	keyStaking := sdk.NewKVStoreKey(staking.StoreKey)
	tkeyStaking := sdk.NewTransientStoreKey(staking.TStoreKey)
	keyAcc := sdk.NewKVStoreKey(authtypes.StoreKey)
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)
	keyBank := sdk.NewKVStoreKey(banktypes.StoreKey)

	keyToken := sdk.NewKVStoreKey("token")
	keyLock := sdk.NewKVStoreKey("lock")

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(tkeyStaking, sdk.StoreTypeTransient, nil)
	ms.MountStoreWithDB(keyStaking, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.MountStoreWithDB(keyToken, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyLock, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyBank, sdk.StoreTypeIAVL, db)

	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "foochainid"}, isCheckTx, nil)

	cdc := codec.New()
	RegisterCodec(cdc)
	cryptocodec.RegisterCrypto(cdc)
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	appCodec := codec.NewHybridCodec(cdc, interfaceRegistry)


	pk := params.NewKeeper(appCodec, keyParams, tkeyParams)
	//feeCollectorAcc := supply.NewEmptyModuleAccount(auth.FeeCollectorName)
	blacklistedAddrs := make(map[string]bool)
	//blacklistedAddrs[feeCollectorAcc.String()] = true
	maccPerms := map[string][]string{
		authtypes.FeeCollectorName: nil,
		types.ModuleName:      nil,
	}

	accountKeeper := authkeeper.NewAccountKeeper(
		appCodec,    // amino codec
		keyAcc, // target store
		pk.Subspace(authtypes.ModuleName),
		authtypes.ProtoBaseAccount, // prototype
		maccPerms,
	)

	bk := bankkeeper.NewBaseKeeper( appCodec, keyBank,
		accountKeeper,
		pk.Subspace(banktypes.ModuleName),
		blacklistedAddrs,
	)

	tk := NewKeeper(bk,
		pk.Subspace(DefaultParamspace),
		authtypes.FeeCollectorName,
		accountKeeper,
		keyToken,
		keyLock,
		cdc,
		true)
	tk.SetParams(ctx, types.DefaultParams())

	return ctx, tk, keyParams, []byte("testToken")
}
