package types

import (
	"fmt"
	sdkerror "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	CodeInvalidPriceDigit       uint32 = 1
	CodeInvalidMinTradeSize     uint32 = 2
	CodeInvalidDexList          uint32 = 3
	CodeInvalidBalanceNotEnough uint32 = 4
	CodeInvalidHeight           uint32 = 5
	CodeInvalidAsset            uint32 = 6
	CodeInvalidCommon           uint32 = 7
)

func ErrInvalidDexList(codespace string, message string) error {
	return sdkerror.New(codespace, CodeInvalidDexList, message)
}

func ErrInvalidBalanceNotEnough(codespace string, message string) error {
	return sdkerror.New(codespace, CodeInvalidBalanceNotEnough, message)
}

func ErrInvalidHeight(codespace string, h, ch, max int64) error {
	return sdkerror.New(codespace, CodeInvalidHeight, fmt.Sprintf("Height %d must be greater than current block height %d and less than %d + %d.", h, ch, ch, max))
}

func ErrInvalidCommon(codespace string, message string) error {
	return sdkerror.New(codespace, CodeInvalidCommon, message)
}
