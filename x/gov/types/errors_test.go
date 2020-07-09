package types

import (
	sdkerror "github.com/cosmos/cosmos-sdk/types/errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrors(t *testing.T) {
	err := ErrInitialDepositNotEnough(DefaultCodespace, "")
	require.Equal(t, CodeInitialDepositNotEnough, err.(sdkerror.Error).ABCICode())

	err = ErrUnknownProposal(DefaultCodespace, 0)
	require.Equal(t, CodeUnknownProposal, err.(sdkerror.Error).ABCICode())

	err = ErrInvalidateProposalStatus(DefaultCodespace, "")
	require.Equal(t, CodeInvalidProposalStatus, err.(sdkerror.Error).ABCICode())

	err = ErrInvalidHeight(DefaultCodespace, 100, 100, 100)
	require.Equal(t, CodeInvalidHeight, err.(sdkerror.Error).ABCICode())

	err = ErrInvalidProposer(DefaultCodespace, "")
	require.Equal(t, CodeInvalidProposer, err.(sdkerror.Error).ABCICode())

}
