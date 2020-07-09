package keeper

import (
	"fmt"
	sdkerror "github.com/cosmos/cosmos-sdk/types/errors"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/okex/okchain/x/common"
	"github.com/okex/okchain/x/common/proto"
	"github.com/okex/okchain/x/gov"
	"github.com/okex/okchain/x/token"
	"github.com/okex/okchain/x/upgrade/types"
)

// GetMinDeposit implements ProposalHandler interface
func (k Keeper) GetMinDeposit(ctx sdk.Context, content gov.Content) (minDeposit sdk.DecCoins) {
	switch content.(type) {
	case types.AppUpgradeProposal:
		minDeposit = k.GetParams(ctx).AppUpgradeMinDeposit
	}

	return
}

// GetMaxDepositPeriod implements ProposalHandler interface
func (k Keeper) GetMaxDepositPeriod(ctx sdk.Context, content gov.Content) (maxDepositPeriod time.Duration) {
	switch content.(type) {
	case types.AppUpgradeProposal:
		maxDepositPeriod = k.GetParams(ctx).AppUpgradeMaxDepositPeriod
	}

	return
}

// GetVotingPeriod implements ProposalHandler interface
func (k Keeper) GetVotingPeriod(ctx sdk.Context, content gov.Content) (votingPeriod time.Duration) {
	switch content.(type) {
	case types.AppUpgradeProposal:
		votingPeriod = k.GetParams(ctx).AppUpgradeVotingPeriod
	}

	return
}

// CheckMsgSubmitProposal implements ProposalHandler interface
func (k Keeper) CheckMsgSubmitProposal(ctx sdk.Context, msg gov.MsgSubmitProposal) error {
	// check message sender is current validator
	if !k.stakingKeeper.IsValidator(ctx, msg.Proposer) {
		return gov.ErrInvalidProposer(types.DefaultCodespace,
			fmt.Sprintf("proposer of App Upgrade Proposal must be validator"))
	}
	// check initial deposit more than or equal to ratio of MinDeposit
	initDeposit := k.GetParams(ctx).AppUpgradeMinDeposit.MulDec(sdk.NewDecWithPrec(1, 1))
	if err := common.HasSufficientCoins(msg.Proposer, msg.InitialDeposit, initDeposit); err != nil {
		return sdkerror.New(types.DefaultCodespace, token.CodeInvalidAsset, err.Error())
	}
	// check proposer has sufficient coins
	if err := common.HasSufficientCoins(msg.Proposer, k.bankKeeper.GetAllBalances(ctx, msg.Proposer), msg.InitialDeposit); err != nil {
		return sdkerror.New(types.DefaultCodespace, token.CodeInvalidAsset, err.Error())
	}

	upgradeProposal := msg.GetContent().(types.AppUpgradeProposal)
	if !k.protocolKeeper.IsValidVersion(ctx, upgradeProposal.ProtocolDefinition.Version) {
		return types.ErrInvalidVersion(types.DefaultCodespace, upgradeProposal.ProtocolDefinition.Version)
	}

	if uint64(ctx.BlockHeight()) > upgradeProposal.ProtocolDefinition.Height {
		return types.ErrInvalidSwitchHeight(types.DefaultCodespace, uint64(ctx.BlockHeight()),
			upgradeProposal.ProtocolDefinition.Height)
	}

	if _, ok := k.protocolKeeper.GetUpgradeConfig(ctx); ok {
		return types.ErrSwitchPeriodInProcess(types.DefaultCodespace)
	}
	return nil
}

// nolint
func (Keeper) VoteHandler(_ sdk.Context, _ gov.Proposal, _ gov.Vote) (string, error) {
	return "", nil
}
func (Keeper) AfterSubmitProposalHandler(_ sdk.Context, _ gov.Proposal) {}
func (Keeper) AfterDepositPeriodPassed(_ sdk.Context, _ gov.Proposal)   {}
func (Keeper) RejectedHandler(_ sdk.Context, _ gov.Content)             {}

// NewAppUpgradeProposalHandler creates a new upgrade handler for gov module
func NewAppUpgradeProposalHandler(k *Keeper) gov.Handler {
	return func(ctx sdk.Context, proposal *gov.Proposal) error {
		switch c := proposal.GetContent().(type) {
		case types.AppUpgradeProposal:
			return handleAppUpgradeProposal(ctx, k, proposal)

		default:
			errMsg := fmt.Sprintf("unrecognized param proposal content type: %s", c.ProposalType())
			return sdkerror.Wrap(sdkerror.ErrUnknownRequest, errMsg)
		}
	}
}

func handleAppUpgradeProposal(ctx sdk.Context, k *Keeper, proposal *gov.Proposal) error {
	logger := ctx.Logger().With("module", types.ModuleName)
	logger.Info("Begin to Execute AppUpgradeProposal")
	upgradeProposal := proposal.GetContent().(types.AppUpgradeProposal)
	if _, found := k.protocolKeeper.GetUpgradeConfig(ctx); found {
		logger.Error("Execute AppUpgradeProposal Failure", "info",
			fmt.Sprintf("App Upgrade Switch Period is in process."))
		return nil
	}

	if !k.protocolKeeper.IsValidVersion(ctx, upgradeProposal.ProtocolDefinition.Version) {
		logger.Error("Execute AppUpgradeProposal Failure", "info",
			fmt.Sprintf("version [%d] in AppUpgradeProposal is NOT valid", upgradeProposal.ProtocolDefinition.Version))
		return nil
	}

	if uint64(ctx.BlockHeight())+1 >= upgradeProposal.ProtocolDefinition.Height {
		logger.Error("Execute AppUpgradeProposal Failure", "info",
			fmt.Sprintf("switch height [%d] in AppUpgradeProposal must be more than current block height",
				upgradeProposal.ProtocolDefinition.Height))
		return nil
	}

	k.protocolKeeper.SetUpgradeConfig(ctx,
		proto.NewAppUpgradeConfig(proposal.ProposalID, upgradeProposal.ProtocolDefinition))
	logger.Info("Execute AppUpgradeProposal Success")

	return nil
}
