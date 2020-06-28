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
func ErrNoPayload(msg string) sdk.Error {
	return sdk.NewError(DefaultCodespace, codeNoPayload, codeToDefaultMsg(codeNoPayload)+": %s", msg)
}

// ErrOutOfGas returns out of gas error
func ErrOutOfGas(msg string) sdk.Error {
	return sdk.NewError(DefaultCodespace, codeOutOfGas, codeToDefaultMsg(codeOutOfGas)+": %s", msg)
}

// ErrCodeStoreOutOfGas returns code storage out of gas error
func ErrCodeStoreOutOfGas(msg string) sdk.Error {
	return sdk.NewError(DefaultCodespace, codeCodeStoreOutOfGas, codeToDefaultMsg(codeCodeStoreOutOfGas)+": %s", msg)
}

// ErrDepth returns depth error
func ErrDepth(msg string) sdk.Error {
	return sdk.NewError(DefaultCodespace, codeDepth, codeToDefaultMsg(codeDepth)+": %s", msg)
}

// ErrTraceLimitReached returns trace limit error
func ErrTraceLimitReached(msg string) sdk.Error {
	return sdk.NewError(DefaultCodespace, codeTraceLimitReached, codeToDefaultMsg(codeTraceLimitReached)+": %s", msg)
}

// ErrNoCompatibleInterpreter returns trace limit error
func ErrNoCompatibleInterpreter(msg string) sdk.Error {
	return sdk.NewError(DefaultCodespace, codeTraceLimitReached, codeToDefaultMsg(codeTraceLimitReached)+": %s", msg)
}

/*

ErrNoCompatibleInterpreter
ErrEmptyInputs
ErrInsufficientBalance
ErrContractAddressCollision
ErrNoCodeExist
ErrMaxCodeSizeExceeded
ErrWriteProtection
ErrReturnDataOutOfBounds
ErrExecutionReverted
ErrInvalidJump
ErrGasUintOverflow
ErrWrongCtx
*/
