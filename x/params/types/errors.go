package types

import (
	sdkerror "github.com/cosmos/cosmos-sdk/types/errors"
)

// const
const (
	CodeInvalidMaxProposalNum uint32 = 4
)

// ErrInvalidMaxProposalNum returns error when the number of params to change are out of limit
func ErrInvalidMaxProposalNum(codespace string, msg string) error {
	return sdkerror.New(codespace, CodeInvalidMaxProposalNum, msg)
}
