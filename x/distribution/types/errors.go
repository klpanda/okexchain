// nolint
package types

import (
	sdkerror "github.com/cosmos/cosmos-sdk/types/errors"
)

type CodeType = uint32

const (
	DefaultCodespace            = ModuleName
	CodeInvalidInput            CodeType          = 103
	CodeNoValidatorCommission   CodeType          = 105
	CodeSetWithdrawAddrDisabled CodeType          = 106
)

func ErrNilDelegatorAddr(codespace string) error {
	return sdkerror.New(codespace, CodeInvalidInput, "delegator address is nil")
}
func ErrNilWithdrawAddr(codespace string) error {
	return sdkerror.New(codespace, CodeInvalidInput, "withdraw address is nil")
}
func ErrNilValidatorAddr(codespace string) error {
	return sdkerror.New(codespace, CodeInvalidInput, "validator address is nil")
}
func ErrNoValidatorCommission(codespace string) error {
	return sdkerror.New(codespace, CodeNoValidatorCommission, "no validator commission to withdraw")
}
func ErrSetWithdrawAddrDisabled(codespace string) error {
	return sdkerror.New(codespace, CodeSetWithdrawAddrDisabled, "set withdraw address disabled")
}
func ErrBadDistribution(codespace string) error {
	return sdkerror.New(codespace, CodeInvalidInput, "community pool does not have sufficient coins to distribute")
}
func ErrInvalidProposalAmount(codespace string) error {
	return sdkerror.New(codespace, CodeInvalidInput, "invalid community pool spend proposal amount")
}
func ErrEmptyProposalRecipient(codespace string) error {
	return sdkerror.New(codespace, CodeInvalidInput, "invalid community pool spend proposal recipient")
}
