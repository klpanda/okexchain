package keeper

import (
	"bytes"
	"encoding/hex"
	"fmt"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"strconv"
	"testing"
	"time"

	"github.com/okex/okchain/x/staking/types"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/okex/okchain/x/params"
	//distr "github.com/okex/okchain/x/distribution"
)

// dummy addresses used for testing
// nolint: unused deadcode
var (
	Addrs = createTestAddrs(500)
	PKs   = createTestPubKeys(500)

	addrDels = []sdk.AccAddress{
		Addrs[0],
		Addrs[1],
		Addrs[2],
	}

	addrVals = []sdk.ValAddress{
		sdk.ValAddress(Addrs[3]),
		sdk.ValAddress(Addrs[4]),
		sdk.ValAddress(Addrs[5]),
		sdk.ValAddress(Addrs[6]),
		sdk.ValAddress(Addrs[7]),
	}

	SufficientInitBalance = int64(10000)
	InitMsd2000           = sdk.NewDec(2000)
	TestChainID           = "stkchainid"
)

type MockStakingKeeper struct {
	Keeper
	StoreKey     sdk.StoreKey
	TkeyStoreKey sdk.StoreKey
	BankKeeper   bankkeeper.BaseKeeper
	MountedStore store.MultiStore
	AccKeeper    authkeeper.AccountKeeper
}

func NewMockStakingKeeper(k Keeper, keyStoreKey, tkeyStoreKey sdk.StoreKey, bKeeper bankkeeper.BaseKeeper,
	ms store.MultiStore, accKeeper authkeeper.AccountKeeper) MockStakingKeeper {
	return MockStakingKeeper{
		k,
		keyStoreKey,
		tkeyStoreKey,
		bKeeper,
		ms,
		accKeeper,
	}
}

//_______________________________________________________________________________________

// MakeTestCodec creates a codec used only for testing
func MakeTestCodec() *codec.Codec {
	var cdc = codec.New()

	// Register Msgs
	cdc.RegisterInterface((*sdk.Msg)(nil), nil)
	cdc.RegisterConcrete(banktypes.MsgSend{}, "test/staking/Send", nil)
	cdc.RegisterConcrete(types.MsgCreateValidator{}, "test/staking/CreateValidator", nil)
	cdc.RegisterConcrete(types.MsgDestroyValidator{}, "test/staking/DestroyValidator", nil)
	cdc.RegisterConcrete(types.MsgEditValidator{}, "test/staking/EditValidator", nil)
	cdc.RegisterConcrete(types.MsgWithdraw{}, "test/staking/MsgWithdraw", nil)
	cdc.RegisterConcrete(types.MsgAddShares{}, "test/staking/MsgAddShares", nil)

	// Register AppAccount
	cdc.RegisterInterface((*authtypes.AccountI)(nil), nil)
	cdc.RegisterConcrete(&authtypes.BaseAccount{}, "test/staking/BaseAccount", nil)
	banktypes.RegisterCodec(cdc)
	cryptocodec.RegisterCrypto(cdc)

	return cdc
}

// CreateTestInput returns all sorts of input required for testing
// initBalance is converted to an amount of tokens.
func CreateTestInput(t *testing.T, isCheckTx bool, initBalance int64) (sdk.Context, authkeeper.AccountKeeper, MockStakingKeeper) {

	// init storage
	keyStaking := sdk.NewKVStoreKey(types.StoreKey)
	tkeyStaking := sdk.NewTransientStoreKey(types.TStoreKey)
	keyAcc := sdk.NewKVStoreKey(authtypes.StoreKey)
	keyParams := sdk.NewKVStoreKey(paramstypes.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(paramstypes.TStoreKey)
	keyBank := sdk.NewKVStoreKey(banktypes.StoreKey)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(tkeyStaking, sdk.StoreTypeTransient, nil)
	ms.MountStoreWithDB(keyStaking, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.MountStoreWithDB(keyBank, sdk.StoreTypeIAVL, db)
	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	// init context
	ctx := sdk.NewContext(ms, abci.Header{ChainID: TestChainID}, isCheckTx, log.NewNopLogger())
	ctx = ctx.WithConsensusParams(
		&abci.ConsensusParams{
			Validator: &abci.ValidatorParams{
				PubKeyTypes: []string{tmtypes.ABCIPubKeyTypeEd25519},
			},
		},
	)
	ctx = ctx.WithBlockTime(time.Now())
	cdc := MakeTestCodec()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	appCodec := codec.NewHybridCodec(cdc, interfaceRegistry)

	feeCollectorAcc := authtypes.NewEmptyModuleAccount(authtypes.FeeCollectorName)
	notBondedPool := authtypes.NewEmptyModuleAccount(types.NotBondedPoolName, authtypes.Burner, authtypes.Staking)
	bondPool := authtypes.NewEmptyModuleAccount(types.BondedPoolName, authtypes.Burner, authtypes.Staking)

	blacklistedAddrs := make(map[string]bool)
	blacklistedAddrs[feeCollectorAcc.String()] = true
	blacklistedAddrs[notBondedPool.String()] = true
	blacklistedAddrs[bondPool.String()] = true

	// init module keepers
	pk := params.NewKeeper(appCodec, keyParams, tkeyParams)

	maccPerms := map[string][]string{
		authtypes.FeeCollectorName:   nil,
		types.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		types.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
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

	initTokens := sdk.NewInt(initBalance)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens))
	totalSupply := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens.MulRaw(int64(len(Addrs)))))

	bk.SetSupply(ctx, banktypes.NewSupply(totalSupply))

	keeper := NewKeeper(appCodec, keyStaking, tkeyStaking, accountKeeper, bk, pk.Subspace(DefaultParamspace), types.DefaultCodespace)
	keeper.SetParams(ctx, types.DefaultParams())

	// set module accounts
	accountKeeper.SetModuleAccount(ctx, feeCollectorAcc)
	accountKeeper.SetModuleAccount(ctx, bondPool)

	// fill all the addresses with some coins, set the loose pool tokens simultaneously
	for _, addr := range Addrs {
		_, err := bk.AddCoins(ctx, addr, initCoins)
		if err != nil {
			panic(err)
		}
	}

	distrKeeper := mockDistributionKeeper{}
	hooks := types.NewMultiStakingHooks(distrKeeper.Hooks())
	keeper.SetHooks(hooks)

	mockKeeper := NewMockStakingKeeper(keeper, keyStaking, tkeyStaking,
		bk, ms, accountKeeper)

	return ctx, accountKeeper, mockKeeper
}

func NewPubKey(pk string) (res crypto.PubKey) {
	pkBytes, err := hex.DecodeString(pk)
	if err != nil {
		panic(err)
	}
	//res, err = crypto.PubKeyFromBytes(pkBytes)
	var pkEd ed25519.PubKeyEd25519
	copy(pkEd[:], pkBytes[:])
	return pkEd
}

// TestAddr is designed for incode address generation
func TestAddr(addr string, bech string) sdk.AccAddress {

	res, err := sdk.AccAddressFromHex(addr)
	if err != nil {
		panic(err)
	}
	bechexpected := res.String()
	if bech != bechexpected {
		panic("Bech encoding doesn't match reference")
	}

	bechres, err := sdk.AccAddressFromBech32(bech)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(bechres, res) {
		panic("Bech decode and hex decode don't match")
	}

	return res
}

func createTestAddrs(numAddrs int) []sdk.AccAddress {
	var addresses []sdk.AccAddress
	var buffer bytes.Buffer

	// start at 100 so we can make up to 999 test addresses with valid test addresses
	for i := 100; i < (numAddrs + 100); i++ {
		numString := strconv.Itoa(i)
		buffer.WriteString("A58856F0FD53BF058B4909A21AEC019107BA6") //base address string

		buffer.WriteString(numString) //adding on final two digits to make addresses unique
		res, err := sdk.AccAddressFromHex(buffer.String())
		if err != nil {
			fmt.Print("error")
		}
		bech := res.String()
		addresses = append(addresses, TestAddr(buffer.String(), bech))
		buffer.Reset()
	}
	return addresses
}

// nolint: unparam
func createTestPubKeys(numPubKeys int) []crypto.PubKey {
	var publicKeys []crypto.PubKey
	var buffer bytes.Buffer

	//start at 10 to avoid changing 1 to 01, 2 to 02, etc
	for i := 100; i < (numPubKeys + 100); i++ {
		numString := strconv.Itoa(i)
		buffer.WriteString("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AF") //base pubkey string
		buffer.WriteString(numString)                                                       //adding on final two digits to make pubkeys unique
		publicKeys = append(publicKeys, NewPubKey(buffer.String()))
		buffer.Reset()
	}
	return publicKeys
}

//_____________________________________________________________________________________

// ValidatorByPowerIndexExists checks whether a certain by-power index record exist
func ValidatorByPowerIndexExists(ctx sdk.Context, keeper MockStakingKeeper, power []byte) bool {
	store := ctx.KVStore(keeper.StoreKey)
	return store.Has(power)
}

func NewTestMsgCreateValidator(address sdk.ValAddress, pubKey crypto.PubKey, msdAmt sdk.Dec) types.MsgCreateValidator {
	msd := sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, msdAmt)

	return types.NewMsgCreateValidator(address, pubKey,
		types.NewDescription("my moniker", "my identity", "my website", "my details"), msd,
	)
}

func NewTestMsgDeposit(delAddr sdk.AccAddress, valAddr sdk.ValAddress, amt sdk.Dec) types.MsgDeposit {
	amount := sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, amt)
	return types.NewMsgDeposit(delAddr, amount)
}

func SimpleCheckValidator(t *testing.T, ctx sdk.Context, stkKeeper Keeper, vaAddr sdk.ValAddress,
	expMsd sdk.Dec, expStatus sdk.BondStatus, expDlgShares sdk.Dec, expJailed bool) *types.Validator {
	val, ok := stkKeeper.GetValidator(ctx, vaAddr)
	require.True(t, ok)
	require.True(t, val.GetMinSelfDelegation().Equal(expMsd), val.MinSelfDelegation.String(), expMsd.String())
	require.True(t, val.GetStatus().Equal(expStatus), val.GetStatus().String(), expStatus.String())
	require.True(t, val.GetDelegatorShares().Equal(expDlgShares), val.GetDelegatorShares().String(), expDlgShares.String())
	require.True(t, val.IsJailed() == expJailed)

	return &val
}

// mockDistributionKeeper is supported to test Hooks
type mockDistributionKeeper struct{}

func (dk mockDistributionKeeper) Hooks() types.StakingHooks                                       { return dk }
func (dk mockDistributionKeeper) AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress)   {}
func (dk mockDistributionKeeper) BeforeValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress) {}
func (dk mockDistributionKeeper) AfterValidatorRemoved(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
}
func (dk mockDistributionKeeper) AfterValidatorBonded(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
}
func (dk mockDistributionKeeper) AfterValidatorBeginUnbonding(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
}
func (dk mockDistributionKeeper) AfterValidatorDestroyed(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
}
