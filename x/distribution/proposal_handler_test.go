package distribution

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/okex/okchain/x/distribution/keeper"
	"github.com/okex/okchain/x/distribution/types"
	govtypes "github.com/okex/okchain/x/gov/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

var (
	delPk1   = ed25519.GenPrivKey().PubKey()
	delAddr1 = sdk.AccAddress(delPk1.Address())

	amount = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1)))
)

func testProposal(recipient sdk.AccAddress, amount sdk.Coins) govtypes.Proposal {
	any, err := govtypes.ContentToAny(types.NewCommunityPoolSpendProposal("Test", "description",
		recipient, amount))
	if err != nil {
		panic(err)
	}

	return govtypes.Proposal{Content: any}
}

func TestProposalHandlerPassed(t *testing.T) {
	ctx, accountKeeper, k, _, bankKeeper := keeper.CreateTestInputDefault(t, false, 10)
	recipient := delAddr1

	// add coins to the module account
	macc := k.GetDistributionAccount(ctx)
	newCoins := bankKeeper.GetAllBalances(ctx, macc.GetAddress()).Add(amount...)
	err := bankKeeper.SetBalances(ctx, macc.GetAddress(), newCoins)
	require.NoError(t, err)

	account := accountKeeper.NewAccountWithAddress(ctx, recipient)
	require.True(t, bankKeeper.GetAllBalances(ctx, account.GetAddress()).IsZero())
	accountKeeper.SetAccount(ctx, account)

	feePool := k.GetFeePool(ctx)
	feePool.CommunityPool = sdk.NewDecCoins(amount...)
	k.SetFeePool(ctx, feePool)

	tp := testProposal(recipient, amount)
	hdlr := NewCommunityPoolSpendProposalHandler(k)
	require.NoError(t, hdlr(ctx, &tp))
	require.Equal(t, bankKeeper.GetAllBalances(ctx, accountKeeper.GetAccount(ctx, recipient).GetAddress()), amount)
}

func TestProposalHandlerFailed(t *testing.T) {
	ctx, accountKeeper, k, _, bankKpeeper := keeper.CreateTestInputDefault(t, false, 10)
	recipient := delAddr1

	account := accountKeeper.NewAccountWithAddress(ctx, recipient)
	require.True(t, bankKpeeper.GetAllBalances(ctx, account.GetAddress()).IsZero())
	accountKeeper.SetAccount(ctx, account)

	tp := testProposal(recipient, amount)
	hdlr := NewCommunityPoolSpendProposalHandler(k)
	require.Error(t, hdlr(ctx, &tp))
	require.True(t, bankKpeeper.GetAllBalances(ctx, accountKeeper.GetAccount(ctx, recipient).GetAddress()).IsZero())
}
