package upgrade

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

	"github.com/okex/okchain/x/common/proto"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/okex/okchain/x/params"
	"github.com/okex/okchain/x/staking"
	"github.com/okex/okchain/x/staking/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

var (
	pubKeys = []crypto.PubKey{
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB50"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB51"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB52"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB53"),
	}

	accAddrs = []sdk.AccAddress{
		sdk.AccAddress(pubKeys[0].Address()),
		sdk.AccAddress(pubKeys[1].Address()),
		sdk.AccAddress(pubKeys[2].Address()),
		sdk.AccAddress(pubKeys[3].Address()),
	}

	maccPerms = map[string][]string{
		staking.BondedPoolName:    {authtypes.Staking},
		staking.NotBondedPoolName: {authtypes.Staking},
	}
)

func testPrepare(t *testing.T) (ctx sdk.Context, keeper Keeper, stakingKeeper staking.Keeper, paramsKeeper params.Keeper) {
	skMap := sdk.NewKVStoreKeys(
		"main",
		authtypes.StoreKey,
		banktypes.StoreKey,

		// for staking/distr rollback to cosmos-sdk
		//staking.StoreKey, staking.DelegatorPoolKey, staking.RedelegationKeyM, staking.RedelegationActonKey, staking.UnbondingKey,
		staking.StoreKey,
		params.StoreKey,
		StoreKey,
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

	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	ctx = sdk.NewContext(ms, abci.Header{}, false, log.NewTMLogger(os.Stdout))
	cdc := getTestCodec()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	appCodec := codec.NewHybridCodec(cdc, interfaceRegistry)

	paramsKeeper = params.NewKeeper(appCodec, skMap[params.StoreKey], tskMap[params.TStoreKey])
	accountKeeper := authkeeper.NewAccountKeeper(appCodec, skMap[authtypes.StoreKey],
		paramsKeeper.Subspace(authtypes.ModuleName), authtypes.ProtoBaseAccount, maccPerms)
	bankKeeper := bankkeeper.NewBaseKeeper(appCodec, skMap[banktypes.StoreKey], accountKeeper, paramsKeeper.Subspace(banktypes.ModuleName), nil)

	// for staking/distr rollback to cosmos-sdk
	//stakingKeeper = staking.NewKeeper(
	//	cdc, skMap[staking.StoreKey], skMap[staking.DelegatorPoolKey], skMap[staking.RedelegationKeyM], skMap[staking.RedelegationActonKey], skMap[staking.UnbondingKey], tskMap[staking.TStoreKey],
	//	supplyKeeper, paramsKeeper.Subspace(staking.DefaultParamspace), staking.DefaultCodespace)
	stakingKeeper = staking.NewKeeper(
		appCodec, skMap[staking.StoreKey], tskMap[staking.TStoreKey],
		accountKeeper, bankKeeper, paramsKeeper.Subspace(staking.DefaultParamspace), staking.DefaultCodespace)

	stakingKeeper.SetParams(ctx, types.DefaultParams())
	protocolKeeper := proto.NewProtocolKeeper(skMap["main"])
	keeper = NewKeeper(cdc, skMap[StoreKey], protocolKeeper, stakingKeeper, bankKeeper, paramsKeeper.Subspace(DefaultParamspace))
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
