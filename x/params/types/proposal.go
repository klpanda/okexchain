package types

import (
	"fmt"
	sdkerror "github.com/cosmos/cosmos-sdk/types/errors"
	"strings"

	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"

	govtypes "github.com/okex/okchain/x/gov/types"

	sdkparams "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Assert ParameterChangeProposal implements govtypes.Content at compile-time
var _ govtypes.Content = ParameterChangeProposal{}

func init() {
	govtypes.RegisterProposalType(proposal.ProposalTypeChange)
	govtypes.RegisterProposalTypeCodec(ParameterChangeProposal{}, "okchain/params/ParameterChangeProposal")
}

// ParameterChangeProposal is the struct of param change proposal
type ParameterChangeProposal struct {
	*proposal.ParameterChangeProposal
	Height uint64 `json:"height" yaml:"height"`
}

// NewParameterChangeProposal creates a new instance of ParameterChangeProposal
func NewParameterChangeProposal(title, description string, changes []proposal.ParamChange, height uint64,
) ParameterChangeProposal {
	return ParameterChangeProposal{
		ParameterChangeProposal: proposal.NewParameterChangeProposal(title, description, changes),
		Height:                  height,
	}
}

// ValidateBasic validates the parameter change proposal
func (pcp ParameterChangeProposal) ValidateBasic() error {
	if len(strings.TrimSpace(pcp.Title)) == 0 {
		return sdkerror.Wrap(govtypes.ErrInvalidProposalContent, "proposal title cannot be blank")
	}
	if len(pcp.Title) > govtypes.MaxTitleLength {
		return sdkerror.Wrap(govtypes.ErrInvalidProposalContent,
			fmt.Sprintf("proposal title is longer than max length of %d", govtypes.MaxTitleLength))
	}

	if len(pcp.Description) == 0 {
		return sdkerror.Wrap(govtypes.ErrInvalidProposalContent, "proposal description cannot be blank")
	}

	if len(pcp.Description) > govtypes.MaxDescriptionLength {
		return sdkerror.Wrap(govtypes.ErrInvalidProposalContent,
			fmt.Sprintf("proposal description is longer than max length of %d", govtypes.MaxDescriptionLength))
	}

	if pcp.ProposalType() != proposal.ProposalTypeChange {
		return sdkerror.Wrap(govtypes.ErrInvalidProposalType, pcp.ProposalType())
	}

	if len(pcp.Changes) != 1 {
		return ErrInvalidMaxProposalNum(sdkparams.ModuleName, fmt.Sprintf("one proposal can only change one pair of parameter"))
	}

	return proposal.ValidateChanges(pcp.Changes)
}
