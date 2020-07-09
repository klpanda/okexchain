//nolint
package types

import (
	"fmt"
	sdkerror "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	DefaultCodespace string = "gov"

	CodeUnknownProposal         uint32 = 1
	CodeInvalidProposalStatus   uint32 = 12
	CodeInitialDepositNotEnough uint32 = 13
	CodeInvalidProposer         uint32 = 14
	CodeInvalidHeight           uint32 = 15
)

func ErrUnknownProposal(codespace string, proposalID uint64) error {
	return sdkerror.New(codespace, CodeUnknownProposal, fmt.Sprintf("unknown proposal with id %d", proposalID))
}

func ErrInvalidateProposalStatus(codespace string, msg string) error {
	return sdkerror.New(codespace, CodeInvalidProposalStatus, msg)
}

func ErrInitialDepositNotEnough(codespace string, initDeposit string) error {
	return sdkerror.New(codespace, CodeInitialDepositNotEnough,
		fmt.Sprintf("InitialDeposit must be greater than or equal to %s", initDeposit))
}

func ErrInvalidProposer(codespace string, message string) error {
	return sdkerror.New(codespace, CodeInvalidProposer, message)
}

func ErrInvalidHeight(codespace string, h, ch, max uint64) error {
	return sdkerror.New(codespace, CodeInvalidHeight,
		fmt.Sprintf("Height %d must be greater than current block height %d and less than %d + %d.",
			h, ch, ch, max))
}
