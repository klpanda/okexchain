package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkGovTypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/okex/okchain/x/gov/types"
)

// SubmitProposal creates new proposal given a content
func (keeper Keeper) SubmitProposal(ctx sdk.Context, content sdkGovTypes.Content) (sdkGovTypes.Proposal, error) {
	if !keeper.Router().HasRoute(content.ProposalRoute()) {
		return sdkGovTypes.Proposal{}, types.ErrNoProposalHandlerExists
	}

	proposalID, err := keeper.GetProposalID(ctx)
	if err != nil {
		return sdkGovTypes.Proposal{}, err
	}
	// get the time now as the submit time
	submitTime := ctx.BlockHeader().Time
	// get params for special proposal
	var depositPeriod time.Duration
	if !keeper.proposalHandlerRouter.HasRoute(content.ProposalRoute()) {
		depositPeriod = keeper.GetDepositParams(ctx).MaxDepositPeriod
	} else {
		proposalParams := keeper.proposalHandlerRouter.GetRoute(content.ProposalRoute())
		depositPeriod = proposalParams.GetMaxDepositPeriod(ctx, content)
	}

	proposal, err := sdkGovTypes.NewProposal(ctx, keeper.totalPower(ctx), content, proposalID, submitTime,
		submitTime.Add(depositPeriod))
	if err != nil {
		return sdkGovTypes.Proposal{}, err
	}

	keeper.SetProposal(ctx, proposal)
	keeper.InsertInactiveProposalQueue(ctx, proposalID, proposal.DepositEndTime)
	keeper.SetProposalID(ctx, proposalID+1)

	if keeper.proposalHandlerRouter.HasRoute(content.ProposalRoute()) {
		keeper.proposalHandlerRouter.GetRoute(content.ProposalRoute()).AfterSubmitProposalHandler(ctx, proposal)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitProposal,
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposalID)),
		),
	)

	return proposal, nil
}

// GetProposalsFiltered get Proposals from store by ProposalID
// voterAddr will filter proposals by whether or not that address has voted on them
// depositorAddr will filter proposals by whether or not that address has deposited to them
// status will filter proposals by status
// numLatest will fetch a specified number of the most recent proposals, or 0 for all proposals
func (keeper Keeper) GetProposalsFiltered(
	ctx sdk.Context, voterAddr sdk.AccAddress, depositorAddr sdk.AccAddress, status types.ProposalStatus,
	numLatest int,
) []sdkGovTypes.Proposal {

	maxProposalID, err := keeper.GetProposalID(ctx)
	if err != nil {
		return []sdkGovTypes.Proposal{}
	}

	matchingProposals := []sdkGovTypes.Proposal{}

	if numLatest == 0 {
		numLatest = int(maxProposalID)
	}

	for proposalID := maxProposalID - uint64(numLatest); proposalID < maxProposalID; proposalID++ {
		if voterAddr != nil && len(voterAddr) != 0 {
			_, found := keeper.GetVote(ctx, proposalID, voterAddr)
			if !found {
				continue
			}
		}

		if depositorAddr != nil && len(depositorAddr) != 0 {
			_, found := keeper.GetDeposit(ctx, proposalID, depositorAddr)
			if !found {
				continue
			}
		}

		proposal, ok := keeper.GetProposal(ctx, proposalID)
		if !ok {
			continue
		}

		if types.ValidProposalStatus(status) && proposal.Status != status {
			continue
		}

		matchingProposals = append(matchingProposals, proposal)
	}
	return matchingProposals
}


func (keeper Keeper) activateVotingPeriod(ctx sdk.Context, proposal *sdkGovTypes.Proposal) {
	proposal.VotingStartTime = ctx.BlockHeader().Time
	var votingPeriod time.Duration
	if !keeper.proposalHandlerRouter.HasRoute(proposal.ProposalRoute()) {
		votingPeriod = keeper.GetVotingPeriod(ctx, proposal.GetContent())
	} else {
		phr := keeper.proposalHandlerRouter.GetRoute(proposal.ProposalRoute())
		votingPeriod = phr.GetVotingPeriod(ctx, proposal.GetContent())
	}
	// calculate the end time of voting
	proposal.VotingEndTime = proposal.VotingStartTime.Add(votingPeriod)
	proposal.Status = types.StatusVotingPeriod

	keeper.RemoveFromInactiveProposalQueue(ctx, proposal.ProposalID, proposal.DepositEndTime)
	keeper.InsertActiveProposalQueue(ctx, proposal.ProposalID, proposal.VotingEndTime)
}
