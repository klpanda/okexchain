package keeper

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/okex/okchain/x/staking/types"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/okex/okchain/x/params"
	"github.com/okex/okchain/x/staking"

	"github.com/okex/okchain/x/distribution/types"
)

//nolint: deadcode unused
var (
	delPk1   = ed25519.GenPrivKey().PubKey()
	delPk2   = ed25519.GenPrivKey().PubKey()
	delPk3   = ed25519.GenPrivKey().PubKey()
	delPk4   = ed25519.GenPrivKey().PubKey()
	delAddr1 = sdk.AccAddress(delPk1.Address())
	delAddr2 = sdk.AccAddress(delPk2.Address())
	delAddr3 = sdk.AccAddress(delPk3.Address())
	delAddr4 = sdk.AccAddress(delPk4.Address())

	valOpPk1    = ed25519.GenPrivKey().PubKey()
	valOpPk2    = ed25519.GenPrivKey().PubKey()
	valOpPk3    = ed25519.GenPrivKey().PubKey()
	valOpPk4    = ed25519.GenPrivKey().PubKey()
	valOpAddr1  = sdk.ValAddress(valOpPk1.Address())
	valOpAddr2  = sdk.ValAddress(valOpPk2.Address())
	valOpAddr3  = sdk.ValAddress(valOpPk3.Address())
	valOpAddr4  = sdk.ValAddress(valOpPk4.Address())
	valAccAddr1 = sdk.AccAddress(valOpPk1.Address()) // generate acc addresses for these validator keys too
	valAccAddr2 = sdk.AccAddress(valOpPk2.Address())
	valAccAddr3 = sdk.AccAddress(valOpPk3.Address())
	valAccAddr4 = sdk.AccAddress(valOpPk4.Address())

	valConsPk1   = ed25519.GenPrivKey().PubKey()
	valConsPk2   = ed25519.GenPrivKey().PubKey()
	valConsPk3   = ed25519.GenPrivKey().PubKey()
	valConsPk4   = ed25519.GenPrivKey().PubKey()
	valConsAddr1 = sdk.ConsAddress(valConsPk1.Address())
	valConsAddr2 = sdk.ConsAddress(valConsPk2.Address())
	valConsAddr3 = sdk.ConsAddress(valConsPk3.Address())
	valConsAddr4 = sdk.ConsAddress(valConsPk4.Address())

	// TODO move to common testing package for all modules
	// test addresses
	TestAddrs = []sdk.AccAddress{
		delAddr1, delAddr2, delAddr3, delAddr4,
		valAccAddr1, valAccAddr2, valAccAddr3, valAccAddr4,
	}

	distrAcc = authtypes.NewEmptyModuleAccount(types.ModuleName)
)

// GetTestAddrs returns valOpAddrs, valConsPks, valConsAddrs for test
func GetTestAddrs() ([]sdk.ValAddress, []crypto.PubKey, []sdk.ConsAddress) {
	valOpAddrs := []sdk.ValAddress{valOpAddr1, valOpAddr2, valOpAddr3, valOpAddr4}
	valConsPks := []crypto.PubKey{valConsPk1, valConsPk2, valConsPk3, valConsPk4}
	valConsAddrs := []sdk.ConsAddress{valConsAddr1, valConsAddr2, valConsAddr3, valConsAddr4}
	return valOpAddrs, valConsPks, valConsAddrs
}

// NewTestDecCoins returns dec coins
func NewTestDecCoins(i int64, precison int64) sdk.DecCoins {
	return sdk.DecCoins{NewTestDecCoin(i, precison)}
}

// NewTestDecCoin returns one dec coin
func NewTestDecCoin(i int64, precison int64) sdk.DecCoin {
	return sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDecWithPrec(i, precison))
}

// MakeTestCodec creates a codec used only for testing
func MakeTestCodec() *codec.Codec {
	var cdc = codec.New()
	banktypes.RegisterCodec(cdc)
	stakingtypes.RegisterCodec(cdc)
	authtypes.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	cryptocodec.RegisterCrypto(cdc)

	types.RegisterCodec(cdc) // distr
	return cdc
}

// CreateTestInputDefault test input with default values
func CreateTestInputDefault(t *testing.T, isCheckTx bool, initPower int64) (
	sdk.Context, authkeeper.AccountKeeper, Keeper, staking.Keeper, types.BankKeeper) {
	communityTax := sdk.NewDecWithPrec(2, 2)
	ctx, ak, bk, dk, sk, _ := CreateTestInputAdvanced(t, isCheckTx, initPower, communityTax)
	sh := staking.NewHandler(sk)
	valOpAddrs, valConsPks, _ := GetTestAddrs()
	// create four validators
	for i := int64(0); i < 4; i++ {
		msg := staking.NewMsgCreateValidator(valOpAddrs[i], valConsPks[i],
			staking.Description{}, NewTestDecCoin(i+1, 0))
		_, err := sh(ctx, &msg)
		require.Nil(t, err)
		// assert initial state: zero current rewards
		require.True(t, dk.GetValidatorAccumulatedCommission(ctx, valOpAddrs[i]).IsZero())
	}
	return ctx, ak, dk, sk, bk
}

// CreateTestInputAdvanced hogpodge of all sorts of input required for testing
func CreateTestInputAdvanced(t *testing.T, isCheckTx bool, initPower int64, communityTax sdk.Dec) (
	sdk.Context, authkeeper.AccountKeeper, bankkeeper.BaseKeeper, Keeper, staking.Keeper, params.Keeper) {

	initTokens := sdk.TokensFromConsensusPower(initPower)

	keyDistr := sdk.NewKVStoreKey(types.StoreKey)
	keyStaking := sdk.NewKVStoreKey(staking.StoreKey)
	tkeyStaking := sdk.NewTransientStoreKey(staking.TStoreKey)
	keyAcc := sdk.NewKVStoreKey(authtypes.StoreKey)
	keyBank := sdk.NewKVStoreKey(banktypes.StoreKey)
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)

	ms.MountStoreWithDB(keyDistr, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyStaking, sdk.StoreTypeTransient, nil)
	ms.MountStoreWithDB(keyStaking, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyBank, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)

	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	feeCollectorAcc := authtypes.NewEmptyModuleAccount(authtypes.FeeCollectorName)
	notBondedPool := authtypes.NewEmptyModuleAccount(staking.NotBondedPoolName, authtypes.Burner, authtypes.Staking)
	bondPool := authtypes.NewEmptyModuleAccount(staking.BondedPoolName, authtypes.Burner, authtypes.Staking)

	blacklistedAddrs := make(map[string]bool)
	blacklistedAddrs[feeCollectorAcc.GetAddress().String()] = true
	blacklistedAddrs[notBondedPool.GetAddress().String()] = true
	blacklistedAddrs[bondPool.GetAddress().String()] = true
	blacklistedAddrs[distrAcc.GetAddress().String()] = true

	cdc := MakeTestCodec()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	appCodec := codec.NewHybridCodec(cdc, interfaceRegistry)
	pk := params.NewKeeper(appCodec, keyParams, tkeyParams)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "foochainid"}, isCheckTx, log.NewNopLogger())
	maccPerms := map[string][]string{
		authtypes.FeeCollectorName:     nil,
		types.ModuleName:          nil,
		staking.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		staking.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
	}
	accountKeeper := authkeeper.NewAccountKeeper(appCodec, keyAcc, pk.Subspace(authtypes.ModuleName), authtypes.ProtoBaseAccount, maccPerms)
	bankKeeper := bankkeeper.NewBaseKeeper(appCodec, keyBank, accountKeeper, pk.Subspace(banktypes.ModuleName), blacklistedAddrs)

	sk := staking.NewKeeper(appCodec, keyStaking, tkeyStaking, accountKeeper, bankKeeper,
		pk.Subspace(staking.DefaultParamspace), staking.DefaultCodespace)
	sk.SetParams(ctx, staking.DefaultParams())

	keeper := NewKeeper(cdc, keyDistr, pk.Subspace(DefaultParamspace), sk, accountKeeper, bankKeeper,
		types.DefaultCodespace, authtypes.FeeCollectorName, blacklistedAddrs)

	keeper.SetWithdrawAddrEnabled(ctx, true)
	initCoins := sdk.NewCoins(sdk.NewCoin(sk.BondDenom(ctx), initTokens))
	totalSupply := sdk.NewCoins(sdk.NewCoin(sk.BondDenom(ctx), initTokens.MulRaw(int64(len(TestAddrs)))))
	bankKeeper.SetSupply(ctx, banktypes.NewSupply(totalSupply))

	// fill all the addresses with some coins, set the loose pool tokens simultaneously
	for _, addr := range TestAddrs {
		_, err := bankKeeper.AddCoins(ctx, addr, initCoins)
		require.Nil(t, err)
	}

	// set module accounts
	keeper.accKeeper.SetModuleAccount(ctx, feeCollectorAcc)
	keeper.accKeeper.SetModuleAccount(ctx, notBondedPool)
	keeper.accKeeper.SetModuleAccount(ctx, bondPool)
	keeper.accKeeper.SetModuleAccount(ctx, distrAcc)

	// set the distribution hooks on staking
	sk.SetHooks(keeper.Hooks())

	// set genesis items required for distribution
	keeper.SetFeePool(ctx, types.InitialFeePool())
	keeper.SetCommunityTax(ctx, communityTax)

	return ctx, accountKeeper, bankKeeper, keeper, sk, pk
}
