package keeper

import (
	"encoding/hex"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"os"
	"testing"

	//"github.com/okex/okchain/x/staking/util"
	"github.com/tendermint/tendermint/crypto/ed25519"

	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/okex/okchain/x/common/proto"
	"github.com/okex/okchain/x/params"

	//"github.com/okex/okchain/x/staking"
	"github.com/okex/okchain/x/staking"
	"github.com/okex/okchain/x/upgrade/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/log"
)

var (
	pubKeys = []crypto.PubKey{
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB50"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB51"),
	}

	accAddrs = []sdk.AccAddress{
		sdk.AccAddress(pubKeys[0].Address()),
		sdk.AccAddress(pubKeys[1].Address()),
	}

	maccPerms = map[string][]string{
		staking.BondedPoolName:    {authtypes.Staking},
		staking.NotBondedPoolName: {authtypes.Staking},
	}
)

func testPrepare(t *testing.T) (ctx sdk.Context, keeper Keeper) {
	skMap := sdk.NewKVStoreKeys(
		"main",
		authtypes.StoreKey,
		banktypes.StoreKey,
		staking.StoreKey,
		params.StoreKey,
		types.StoreKey,
	)
	tskMap := sdk.NewTransientStoreKeys(staking.TStoreKey, params.TStoreKey)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)

	for _, v := range skMap {
		ms.MountStoreWithDB(v, sdk.StoreTypeIAVL, db)
	}

	for _, v := range tskMap {
		ms.MountStoreWithDB(v, sdk.StoreTypeTransient, db)
	}

	require.NoError(t, ms.LoadLatestVersion())

	ctx = sdk.NewContext(ms, abci.Header{}, false, log.NewTMLogger(os.Stdout))
	cdc := getTestCodec()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	appCodec := codec.NewHybridCodec(cdc, interfaceRegistry)
	paramsKeeper := params.NewKeeper(appCodec, skMap[params.StoreKey], tskMap[params.TStoreKey])
	accountKeeper := authkeeper.NewAccountKeeper(appCodec, skMap[authtypes.StoreKey],
		paramsKeeper.Subspace(authtypes.ModuleName), authtypes.ProtoBaseAccount, maccPerms)
	bankKeeper := bankkeeper.NewBaseKeeper(appCodec, skMap[banktypes.StoreKey], accountKeeper, paramsKeeper.Subspace(banktypes.ModuleName), nil)

	stakingKeeper := staking.NewKeeper(
		appCodec, skMap[staking.StoreKey], tskMap[staking.TStoreKey],
		accountKeeper, bankKeeper, paramsKeeper.Subspace(staking.DefaultParamspace), staking.DefaultCodespace)
	stakingKeeper.SetParams(ctx, staking.DefaultParams())
	protocolKeeper := proto.NewProtocolKeeper(skMap["main"])
	keeper = NewKeeper(cdc, skMap[types.StoreKey], protocolKeeper, stakingKeeper, bankKeeper, paramsKeeper.Subspace(types.DefaultParamspace))
	return
}

func getTestCodec() *codec.Codec {
	cdc := codec.New()
	sdk.RegisterCodec(cdc)
	authtypes.RegisterCodec(cdc)
	banktypes.RegisterCodec(cdc)
	staking.RegisterCodec(cdc)
	cryptocodec.RegisterCrypto(cdc)
	cdc.Seal()
	return cdc
}

func newPubKey(pubKey string) (res crypto.PubKey) {
	pubKeyBytes, err := hex.DecodeString(pubKey)
	if err != nil {
		panic(err)
	}
	var pubKeyEd25519 ed25519.PubKeyEd25519
	copy(pubKeyEd25519[:], pubKeyBytes[:])
	return pubKeyEd25519
}
