package types

import (
	"fmt"
	sdkerror "github.com/cosmos/cosmos-sdk/types/errors"
)

// const CodeType
const (
	codeInvalidProduct      uint32 = 1
	codeTokenPairNotFound   uint32 = 2
	codeDelistOwnerNotMatch uint32 = 3

	codeInvalidBalanceNotEnough uint32 = 4
	codeInvalidAsset            uint32 = 5
)

// CodeType to Message
func codeToDefaultMsg(code uint32) string {
	switch code {
	case codeInvalidProduct:
		return "invalid product"
	case codeTokenPairNotFound:
		return "tokenpair not found"
	case codeDelistOwnerNotMatch:
		return "tokenpair delistor should be it's owner "
	default:
		return fmt.Sprintf("unknown code %d", code)
	}
}

// ErrInvalidProduct returns invalid product error
func ErrInvalidProduct(msg string) *sdkerror.Error {
	return sdkerror.New(DefaultCodespace, codeInvalidProduct, codeToDefaultMsg(codeInvalidProduct)+": %s" + msg)
}

// ErrTokenPairNotFound returns token pair not found error
func ErrTokenPairNotFound(msg string) *sdkerror.Error {
	return sdkerror.New(DefaultCodespace, codeTokenPairNotFound, codeToDefaultMsg(codeTokenPairNotFound)+": %s" + msg)
}

// ErrDelistOwnerNotMatch returns delist owner not match error
func ErrDelistOwnerNotMatch(msg string) *sdkerror.Error {
	return sdkerror.New(DefaultCodespace, codeDelistOwnerNotMatch, codeToDefaultMsg(codeDelistOwnerNotMatch)+": %s" + msg)
}

// ErrInvalidBalanceNotEnough returns invalid balance not enough error
func ErrInvalidBalanceNotEnough(message string) *sdkerror.Error {
	return sdkerror.New(DefaultCodespace, codeInvalidBalanceNotEnough, message)
}

// ErrInvalidAsset returns invalid asset error
func ErrInvalidAsset(message string) *sdkerror.Error {
	return sdkerror.New(DefaultCodespace, codeInvalidAsset, message)
}

// ErrTokenPairExisted returns an error when the token pair is existing during the process of listing
func ErrTokenPairExisted(baseAsset, quoteAsset string) *sdkerror.Error {
	return sdkerror.New(DefaultCodespace, codeInvalidAsset,
		fmt.Sprintf("failed. the token pair exists with %s and %s", baseAsset, quoteAsset))
}
