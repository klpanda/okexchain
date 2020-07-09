package gov

import (
	"fmt"
	sdkerror "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/okex/okchain/x/common"
	"github.com/okex/okchain/x/gov/keeper"
	"github.com/okex/okchain/x/gov/types"
)

// NewHandler handle all "gov" type messages.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *MsgDeposit:
			return handleMsgDeposit(ctx, keeper, *msg)

		case *MsgSubmitProposal:
			return handleMsgSubmitProposal(ctx, keeper, *msg)

		case *MsgVote:
			return handleMsgVote(ctx, keeper, *msg)

		default:
			errMsg := fmt.Sprintf("unrecognized gov message type: %T", msg)
			return nil, sdkerror.Wrap(sdkerror.ErrUnknownRequest, errMsg)
		}
	}
}

func handleMsgSubmitProposal(ctx sdk.Context, keeper keeper.Keeper, msg MsgSubmitProposal) (*sdk.Result, error) {
	err := hasOnlyDefaultBondDenom(msg.InitialDeposit)
	if err != nil {
		return nil, err
	}

	// use ctx directly
	if !keeper.ProposalHandlerRouter().HasRoute(msg.GetContent().ProposalRoute()) {
		err = keeper.CheckMsgSubmitProposal(ctx, msg)
	} else {
		proposalHandler := keeper.ProposalHandlerRouter().GetRoute(msg.GetContent().ProposalRoute())
		err = proposalHandler.CheckMsgSubmitProposal(ctx, msg)
	}
	if err != nil {
		return nil, err
	}

	proposal, err := keeper.SubmitProposal(ctx, msg.GetContent())
	if err != nil {
		return nil, err
	}

	err = keeper.AddDeposit(ctx, proposal.ProposalID, msg.Proposer,
		msg.InitialDeposit, types.EventTypeSubmitProposal)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Proposer.String()),
		),
	)

	return &sdk.Result{
		Data:   types.GetProposalIDBytes(proposal.ProposalID),
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}

func handleMsgDeposit(ctx sdk.Context, keeper keeper.Keeper, msg MsgDeposit) (*sdk.Result, error) {
	if err := hasOnlyDefaultBondDenom(msg.Amount); err != nil {
		return nil, err
	}
	// check depositor has sufficient coins
	err := common.HasSufficientCoins(msg.Depositor, keeper.BankKeeper().GetAllBalances(ctx, msg.Depositor),
		msg.Amount)
	if err != nil {
		return nil, sdkerror.Wrapf(sdkerror.ErrInsufficientFunds, err.Error())
	}

	err = keeper.AddDeposit(ctx, msg.ProposalID, msg.Depositor,
		msg.Amount, types.EventTypeProposalDeposit)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Depositor.String()),
		),
	)

	return &sdk.Result{Events: ctx.EventManager().Events().ToABCIEvents()}, nil
}

func handleMsgVote(ctx sdk.Context, k keeper.Keeper, msg MsgVote) (*sdk.Result, error) {
	proposal, ok := k.GetProposal(ctx, msg.ProposalID)
	if !ok {
		return nil, types.ErrUnknownProposal(types.DefaultCodespace, msg.ProposalID)
	}

	err, _ := k.AddVote(ctx, msg.ProposalID, msg.Voter, msg.Option)
	if err != nil {
		return nil, err
	}

	status, distribute, tallyResults := keeper.Tally(ctx, k, proposal, false)
	// update tally results after vote every time
	proposal.FinalTallyResult = tallyResults

	// this vote makes the votingPeriod end
	if status != StatusVotingPeriod {
		handleProposalAfterTally(ctx, k, &proposal, distribute, status)
		k.RemoveFromActiveProposalQueue(ctx, proposal.ProposalID, proposal.VotingEndTime)
		proposal.VotingEndTime = ctx.BlockHeader().Time
		k.DeleteVotes(ctx, proposal.ProposalID)
	}
	k.SetProposal(ctx, proposal)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Voter.String()),
			sdk.NewAttribute(types.AttributeKeyProposalStatus, proposal.Status.String()),
		),
	)

	return &sdk.Result{Events: ctx.EventManager().Events().ToABCIEvents()}, nil
}

func handleProposalAfterTally(
	ctx sdk.Context, k keeper.Keeper, proposal *types.Proposal, distribute bool, status ProposalStatus,
) (string, string) {
	if distribute {
		k.DistributeDeposits(ctx, proposal.ProposalID)
	} else {
		k.RefundDeposits(ctx, proposal.ProposalID)
	}

	if status == StatusPassed {
		handler := k.Router().GetRoute(proposal.ProposalRoute())
		cacheCtx, writeCache := ctx.CacheContext()

		// The proposal handler may execute state mutating logic depending
		// on the proposal content. If the handler fails, no state mutation
		// is written and the error message is logged.
		err := handler(cacheCtx, proposal)
		if err == nil {
			proposal.Status = StatusPassed
			// write state to the underlying multi-store
			writeCache()
			return types.AttributeValueProposalPassed, "passed"
		}

		proposal.Status = StatusFailed
		return types.AttributeValueProposalFailed, fmt.Sprintf("passed, but failed on execution: %s",
			err.Error())
	} else if status == StatusRejected {
		if k.ProposalHandlerRouter().HasRoute(proposal.ProposalRoute()) {
			k.ProposalHandlerRouter().GetRoute(proposal.ProposalRoute()).RejectedHandler(ctx, proposal.GetContent())
		}
		proposal.Status = StatusRejected
		return types.AttributeValueProposalRejected, "rejected"
	}
	return "", ""
}

func hasOnlyDefaultBondDenom(decCoins sdk.DecCoins) error {
	if len(decCoins) != 1 || decCoins[0].Denom != sdk.DefaultBondDenom || !decCoins.IsValid() {
		return sdkerror.Wrap(sdkerror.ErrInvalidCoins, fmt.Sprintf("must deposit %s but got %s", sdk.DefaultBondDenom, decCoins.String()))
	}
	return nil
}
