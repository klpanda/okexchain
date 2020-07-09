package types

import (
	"fmt"
	sdkerror "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govTypes "github.com/okex/okchain/x/gov/types"
)

// const
const (
	DefaultCodespace          string = "upgrade"
	CodeInvalidMsgType        uint32      = 100
	CodeUnSupportedMsgType    uint32      = 101
	CodeNotCurrentProposal    uint32      = 102
	CodeNotValidator          uint32      = 103
	CodeDoubleSwitch          uint32      = 104
	CodeNoUpgradeConfig       uint32      = 105
	CodeInvalidUpgradeParams  uint32      = 107
	CodeInvalidSoftWareDescri uint32      = 108
	CodeInvalidVersion        uint32      = 109
	CodeSwitchPeriodInProcess uint32      = 110
)

func codeToDefaultMsg(code uint32) string {
	switch code {
	case CodeInvalidMsgType:
		return "Invalid msg type"
	case CodeUnSupportedMsgType:
		return "current version software doesn't support the msg type"
	default:
		return fmt.Sprintf("unknown code %d", code)
	}
}

// NewError returns a new error with a specific msg
func NewError(codespace string, code uint32, msg string) error {
	msg = msgOrDefaultMsg(msg, code)
	return sdkerror.New(codespace, code, msg)
}

func msgOrDefaultMsg(msg string, code uint32) string {
	if msg != "" {
		return msg
	}
	return codeToDefaultMsg(code)
}

// ErrInvalidVersion returns an error when the version is invalid
func ErrInvalidVersion(codespace string, version uint64) error {
	return sdkerror.New(codespace, CodeInvalidVersion,
		fmt.Sprintf("failed. version [%v] in AppUpgradeProposal is invalid", version))
}

// ErrInvalidSwitchHeight returns an error when the switch height for upgrade is invalid
func ErrInvalidSwitchHeight(codespace string, blockHeight uint64, switchHeight uint64) error {
	return sdkerror.New(codespace, govTypes.CodeInvalidHeight,
		fmt.Sprintf("failed. protocol switchHeight [%v] in AppUpgradeProposal isn't large than current block height [%v]",
			switchHeight, blockHeight))
}

// ErrSwitchPeriodInProcess returns an error when the UpgradeConfig has existed
func ErrSwitchPeriodInProcess(codespace string) error {
	return sdkerror.New(codespace, CodeSwitchPeriodInProcess, "failed. app upgrade switch period is in process")
}

func errZeroSwitchHeight(codespace string) error {
	return sdkerror.New(codespace, govTypes.CodeInvalidHeight,
		fmt.Sprintf("failed. protocol switchHeight in AppUpgradeProposal isn't allowed to be 0"))
}

func errInvalidUpgradeThreshold(codespace string, Threshold sdk.Dec) error {
	return sdkerror.New(codespace, CodeInvalidUpgradeParams,
		fmt.Sprintf("failed. invalid Upgrade Threshold( "+Threshold.String()+" ) should be [0.75, 1)"))
}

func errInvalidLength(codespace string, descriptor string, got, max int) error {
	msg := fmt.Sprintf("failed. bad length for %v, got length %v, max is %v", descriptor, got, max)
	return sdkerror.New(codespace, CodeInvalidSoftWareDescri, msg)
}
