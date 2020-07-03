package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultCodespace sdk.CodespaceType = ModuleName
)

// const CodeType
const (
	codeNoPayload                sdk.CodeType = 1
	codeOutOfGas                 sdk.CodeType = 2
	codeCodeStoreOutOfGas        sdk.CodeType = 3
	codeDepth                    sdk.CodeType = 4
	codeTraceLimitReached        sdk.CodeType = 5
	codeNoCompatibleInterpreter  sdk.CodeType = 6
	codeEmptyInputs              sdk.CodeType = 7
	codeInsufficientBalance      sdk.CodeType = 8
	codeContractAddressCollision sdk.CodeType = 9
	codeNoCodeExist              sdk.CodeType = 10
	codeMaxCodeSizeExceeded      sdk.CodeType = 11
	codeWriteProtection          sdk.CodeType = 12
	codeReturnDataOutOfBounds    sdk.CodeType = 13
	codeExecutionReverted        sdk.CodeType = 14
	codeInvalidJump              sdk.CodeType = 15
	codeGasUintOverflow          sdk.CodeType = 16
	codeWrongCtx                 sdk.CodeType = 17
)

// CodeType to Message
func codeToDefaultMsg(code sdk.CodeType) string {
	switch code {
	case codeNoPayload:
		return "no payload"
	case codeOutOfGas:
		return "out of gas"
	case codeCodeStoreOutOfGas:
		return "contract creation code storage out of gas"
	case codeDepth:
		return "max call depth exceeded"
	case codeTraceLimitReached:
		return "the number of logs reached the specified limit"
	case codeNoCompatibleInterpreter:
		return "no compatible interpreter"
	case codeEmptyInputs:
		return "empty input"
	case codeInsufficientBalance:
		return "insufficient balance for transfer"
	case codeContractAddressCollision:
		return "contract address collision"
	case codeNoCodeExist:
		return "no code exists"
	case codeMaxCodeSizeExceeded:
		return "evm: max code size exceeded"
	case codeWriteProtection:
		return "vm: write protection"
	case codeReturnDataOutOfBounds:
		return "evm: return data out of bounds"
	case codeExecutionReverted:
		return "evm: execution reverted"
	case codeInvalidJump:
		return "evm: invalid jump destination"
	case codeGasUintOverflow:
		return "gas uint64 overflow"
	case codeWrongCtx:
		return "must be simulate mode when gas limit is 0"

	default:
		return fmt.Sprintf("unknown code %d", code)
	}
}

// ErrNoPayload returns no payload error
func ErrNoPayload() sdk.Error {
	return sdk.NewError(DefaultCodespace, codeNoPayload, codeToDefaultMsg(codeNoPayload))
}

// ErrOutOfGas returns out of gas error
func ErrOutOfGas() sdk.Error {
	return sdk.NewError(DefaultCodespace, codeOutOfGas, codeToDefaultMsg(codeOutOfGas))
}

// ErrCodeStoreOutOfGas returns code storage out of gas error
func ErrCodeStoreOutOfGas() sdk.Error {
	return sdk.NewError(DefaultCodespace, codeCodeStoreOutOfGas, codeToDefaultMsg(codeCodeStoreOutOfGas))
}

// ErrDepth returns depth error
func ErrDepth() sdk.Error {
	return sdk.NewError(DefaultCodespace, codeDepth, codeToDefaultMsg(codeDepth))
}

// ErrTraceLimitReached
func ErrTraceLimitReached() sdk.Error {
	return sdk.NewError(DefaultCodespace, codeTraceLimitReached, codeToDefaultMsg(codeTraceLimitReached))
}

// ErrNoCompatibleInterpreter
func ErrNoCompatibleInterpreter() sdk.Error {
	return sdk.NewError(DefaultCodespace, codeTraceLimitReached, codeToDefaultMsg(codeTraceLimitReached))
}

// ErrEmptyInputs returns empty input error
func ErrEmptyInputs() sdk.Error {
	return sdk.NewError(DefaultCodespace, codeEmptyInputs, codeToDefaultMsg(codeEmptyInputs))
}

// ErrInsufficientBalance returns insufficient balance error
func ErrInsufficientBalance() sdk.Error {
	return sdk.NewError(DefaultCodespace, codeInsufficientBalance, codeToDefaultMsg(codeInsufficientBalance))
}

// ErrContractAddressCollision
func ErrContractAddressCollision() sdk.Error {
	return sdk.NewError(DefaultCodespace, codeContractAddressCollision, codeToDefaultMsg(codeContractAddressCollision))
}

// ErrNoCodeExist
func ErrNoCodeExist() sdk.Error {
	return sdk.NewError(DefaultCodespace, codeNoCodeExist, codeToDefaultMsg(codeNoCodeExist))
}

// ErrWriteProtection
func ErrWriteProtection() sdk.Error {
	return sdk.NewError(DefaultCodespace, codeWriteProtection, codeToDefaultMsg(codeWriteProtection))
}

// ErrReturnDataOutOfBounds
func ErrReturnDataOutOfBounds() sdk.Error {
	return sdk.NewError(DefaultCodespace, codeMaxCodeSizeExceeded, codeToDefaultMsg(codeMaxCodeSizeExceeded))
}

// ErrExecutionReverted
func ErrExecutionReverted() sdk.Error {
	return sdk.NewError(DefaultCodespace, codeExecutionReverted, codeToDefaultMsg(codeExecutionReverted))
}

// ErrInvalidJump
func ErrInvalidJump() sdk.Error {
	return sdk.NewError(DefaultCodespace, codeInvalidJump, codeToDefaultMsg(codeInvalidJump))
}

// ErrGasUintOverflow
func ErrGasUintOverflow() sdk.Error {
	return sdk.NewError(DefaultCodespace, codeGasUintOverflow, codeToDefaultMsg(codeGasUintOverflow))
}

// ErrWrongCtx
func ErrWrongCtx() sdk.Error {
	return sdk.NewError(DefaultCodespace, codeWrongCtx, codeToDefaultMsg(codeWrongCtx))
}

// ErrMaxCodeSizeExceeded
func ErrMaxCodeSizeExceeded() sdk.Error {
	return sdk.NewError(DefaultCodespace, codeMaxCodeSizeExceeded, codeToDefaultMsg(codeMaxCodeSizeExceeded))
}
