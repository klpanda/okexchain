package keeper

import (
	"bytes"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"strconv"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/okex/okchain/x/gov/types"
	"github.com/okex/okchain/x/params"
	"github.com/okex/okchain/x/staking"
)

var (
	// Addrs store generated addresses for test
	Addrs = createTestAddrs(500)

	DefaultMSD = sdk.NewDecWithPrec(1, 3)
)

var (
	pubkeys = []crypto.PubKey{
		ed25519.GenPrivKey().PubKey(), ed25519.GenPrivKey().PubKey(),
		ed25519.GenPrivKey().PubKey(), ed25519.GenPrivKey().PubKey(),
	}

	testDescription = staking.NewDescription("T", "E", "S", "T")
)

// nolint: unparam
func createTestAddrs(numAddrs int) []sdk.AccAddress {
	var addresses []sdk.AccAddress
	var buffer bytes.Buffer

	// start at 100 so we can make up to 999 test addresses with valid test addresses
	for i := 100; i < (numAddrs + 100); i++ {
		numString := strconv.Itoa(i)
		_, err := buffer.WriteString("A58856F0FD53BF058B4909A21AEC019107BA6") //base address string
		if err != nil {
			panic(err)
		}

		_, err = buffer.WriteString(numString) //adding on final two digits to make addresses unique
		if err != nil {
			panic(err)
		}
		res, err := sdk.AccAddressFromHex(buffer.String())
		if err != nil {
			panic(err)
		}
		addresses = append(addresses, res)
		buffer.Reset()
	}
	return addresses
}

// CreateValidators creates validators according to arguments
func CreateValidators(
	t *testing.T, stakingHandler sdk.Handler, ctx sdk.Context, addrs []sdk.ValAddress, powerAmt []int64,
) {
	require.True(t, len(addrs) <= len(pubkeys), "Not enough pubkeys specified at top of file.")

	for i := 0; i < len(addrs); i++ {
		valCreateMsg := staking.NewMsgCreateValidator(
			addrs[i], pubkeys[i],
			testDescription,
			sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, DefaultMSD),
		)

		_, err := stakingHandler(ctx, &valCreateMsg)
		require.Nil(t, err)
	}
}

// CreateTestInput returns keepers for test
func CreateTestInput(
	t *testing.T, isCheckTx bool, initBalance int64,
) (sdk.Context, authkeeper.AccountKeeper, Keeper, staking.Keeper, crisiskeeper.Keeper) {
	stakingSk := sdk.NewKVStoreKey(staking.StoreKey)

	stakingTkSk := sdk.NewTransientStoreKey(staking.TStoreKey)

	keyAcc := sdk.NewKVStoreKey(authtypes.StoreKey)
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)
	keyBank := sdk.NewKVStoreKey(banktypes.StoreKey)
	keyGov := sdk.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(stakingTkSk, sdk.StoreTypeTransient, nil)
	ms.MountStoreWithDB(stakingSk, sdk.StoreTypeIAVL, db)

	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.MountStoreWithDB(keyBank, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyGov, sdk.StoreTypeIAVL, db)
	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "okchain"}, isCheckTx, log.NewNopLogger())
	ctx = ctx.WithConsensusParams(
		&abci.ConsensusParams{
			Validator: &abci.ValidatorParams{
				PubKeyTypes: []string{tmtypes.ABCIPubKeyTypeEd25519},
			},
		},
	)
	cdc := MakeTestCodec()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	appCodec := codec.NewHybridCodec(cdc, interfaceRegistry)

	feeCollectorAcc := authtypes.NewEmptyModuleAccount(authtypes.FeeCollectorName)
	notBondedPool := authtypes.NewEmptyModuleAccount(staking.NotBondedPoolName, authtypes.Staking)
	bondPool := authtypes.NewEmptyModuleAccount(staking.BondedPoolName, authtypes.Staking)
	govAcc := authtypes.NewEmptyModuleAccount(types.ModuleName, authtypes.Staking)

	blacklistedAddrs := make(map[string]bool)
	blacklistedAddrs[feeCollectorAcc.String()] = true
	blacklistedAddrs[notBondedPool.String()] = true
	blacklistedAddrs[bondPool.String()] = true

	pk := params.NewKeeper(appCodec, keyParams, tkeyParams)
	pk.SetParams(ctx, params.DefaultParams())

	maccPerms := map[string][]string{
		authtypes.FeeCollectorName:     nil,
		staking.NotBondedPoolName: {authtypes.Staking},
		staking.BondedPoolName:    {authtypes.Staking},
		types.ModuleName:          nil,
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
	pk.SetBankKeeper(bk)


	initCoins := sdk.NewCoins(sdk.NewInt64DecCoin(sdk.DefaultBondDenom, initBalance))
	totalSupply := sdk.NewCoins(sdk.NewInt64DecCoin(sdk.DefaultBondDenom, initBalance*(int64(len(Addrs)))))

	bk.SetSupply(ctx, banktypes.NewSupply(totalSupply))

	// for staking/distr rollback to cosmos-sdk
	stakingKeeper := staking.NewKeeper(appCodec, stakingSk, stakingTkSk, accountKeeper, bk,
		pk.Subspace(staking.DefaultParamspace), staking.DefaultCodespace)

	stakingKeeper.SetParams(ctx, staking.DefaultParams())
	pk.SetStakingKeeper(stakingKeeper)

	// set module accounts
	err = bk.SetBalances(ctx, notBondedPool.GetAddress(), totalSupply)
	require.NoError(t, err)

	accountKeeper.SetModuleAccount(ctx, feeCollectorAcc)
	accountKeeper.SetModuleAccount(ctx, bondPool)
	accountKeeper.SetModuleAccount(ctx, notBondedPool)
	accountKeeper.SetModuleAccount(ctx, govAcc)

	// fill all the addresses with some coins, set the loose pool tokens simultaneously
	for _, addr := range Addrs {
		_, err := bk.AddCoins(ctx, addr, initCoins)
		if err != nil {
			panic(err)
		}
	}

	govSubspace := pk.Subspace(types.DefaultParamspace)
	govRouter := NewRouter()
	govRouter.AddRoute(types.RouterKey, types.ProposalHandler).
		AddRoute(params.RouterKey, params.NewParamChangeProposalHandler(&pk))
	govProposalHandlerRouter := NewProposalHandlerRouter()
	govProposalHandlerRouter.AddRoute(params.RouterKey, pk)
	keeper := NewKeeper(appCodec, keyGov, govSubspace, accountKeeper, bk, stakingKeeper,
			govRouter, govProposalHandlerRouter, authtypes.FeeCollectorName)
	pk.SetGovKeeper(keeper)

	minDeposit := sdk.NewDecCoinsFromDec(sdk.DefaultBondDenom, sdk.NewDec(100))
	depositParams := types.DepositParams{
		MinDeposit:       minDeposit,
		MaxDepositPeriod: time.Hour * 24,
	}
	votingParams := types.VotingParams{
		VotingPeriod: time.Hour * 72,
	}
	tallyParams := types.TallyParams{
		Quorum:          sdk.NewDecWithPrec(334, 3),
		Threshold:       sdk.NewDecWithPrec(5, 1),
		Veto:            sdk.NewDecWithPrec(334, 3),
		YesInVotePeriod: sdk.NewDecWithPrec(667, 3),
	}
	keeper.SetProposalID(ctx, 1)
	keeper.SetDepositParams(ctx, depositParams)
	keeper.SetVotingParams(ctx, votingParams)
	keeper.SetTallyParams(ctx, tallyParams)

	crisisKeeper := crisiskeeper.NewKeeper(pk.Subspace(crisistypes.ModuleName), 0,
		bk, authtypes.FeeCollectorName)
	return ctx, accountKeeper, keeper, stakingKeeper, crisisKeeper
}

// MakeTestCodec creates a codec used only for testing
func MakeTestCodec() *codec.Codec {
	var cdc = codec.New()

	// Register Msgs
	cdc.RegisterInterface((*sdk.Msg)(nil), nil)
	cdc.RegisterConcrete(types.MsgSubmitProposal{}, "test/gov/MsgSubmitProposal", nil)
	cdc.RegisterConcrete(types.MsgDeposit{}, "test/gov/MsgDeposit", nil)
	cdc.RegisterConcrete(types.MsgVote{}, "test/gov/MsgVote", nil)

	cdc.RegisterInterface((*types.Content)(nil), nil)
	cdc.RegisterConcrete(types.TextProposal{}, "test/gov/TextProposal", nil)
	cdc.RegisterConcrete(params.ParameterChangeProposal{}, "test/params/ParameterChangeProposal", nil)
	cdc.RegisterConcrete(types.Proposal{}, "test/gov/Proposal", nil)

	// Register AppAccount
	cdc.RegisterInterface((*authtypes.AccountI)(nil), nil)
	cdc.RegisterConcrete(&authtypes.BaseAccount{}, "test/gov/BaseAccount", nil)
	banktypes.RegisterCodec(cdc)
	cryptocodec.RegisterCrypto(cdc)

	return cdc
}
