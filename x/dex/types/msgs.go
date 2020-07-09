package types

import (
	"fmt"
	sdkerror "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	typeMsgDeposit           = "deposit"
	typeMsgWithdraw          = "withdraw"
	typeMsgTransferOwnership = "transferOwnership"
)

// MsgList - high level transaction of the dex module
var _ sdk.Msg = &MsgList{}

// NewMsgList creates a new MsgList
func NewMsgList(owner sdk.AccAddress, listAsset, quoteAsset string, initPrice sdk.Dec) MsgList {
	return MsgList{
		Owner:      owner,
		ListAsset:  listAsset,
		QuoteAsset: quoteAsset,
		InitPrice:  initPrice,
	}
}

// Route Implements Msg
func (msg MsgList) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgList) Type() string { return "list" }

// ValidateBasic Implements Msg
func (msg MsgList) ValidateBasic() error {
	if msg.ListAsset == msg.QuoteAsset {
		return sdkerror.Wrap(sdkerror.ErrInvalidCoins, fmt.Sprintf("failed to list product because base asset is same as quote asset"))
	}

	if !msg.InitPrice.IsPositive() {
		return sdkerror.Wrapf(sdkerror.ErrUnknownRequest, "invalid init price")
	}

	if msg.Owner.Empty() {
		return sdkerror.Wrapf(sdkerror.ErrInvalidAddress, "missing owner address")
	}
	return nil
}

// GetSignBytes Implements Msg
func (msg MsgList) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners Implements Msg
func (msg MsgList) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Owner}
}

// MsgDeposit - high level transaction of the dex module
var _ sdk.Msg = &MsgDeposit{}

// NewMsgDeposit creates a new MsgDeposit
func NewMsgDeposit(product string, amount sdk.DecCoin, depositor sdk.AccAddress) MsgDeposit {
	return MsgDeposit{product, amount, depositor}
}

// Route Implements Msg
func (msg MsgDeposit) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgDeposit) Type() string { return typeMsgDeposit }

// ValidateBasic Implements Msg
func (msg MsgDeposit) ValidateBasic() error {
	if msg.Depositor.Empty() {
		return sdkerror.Wrapf(sdkerror.ErrInvalidAddress, msg.Depositor.String())
	}
	if !msg.Amount.IsValid() || !msg.Amount.IsPositive() {
		return sdkerror.Wrapf(sdkerror.ErrInvalidCoins, msg.Amount.String())
	}

	return nil
}

// GetSignBytes Implements Msg
func (msg MsgDeposit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners Implements Msg
func (msg MsgDeposit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Depositor}
}

// MsgWithdraw - high level transaction of the dex module
var _ sdk.Msg = &MsgWithdraw{}

// NewMsgWithdraw creates a new MsgWithdraw
func NewMsgWithdraw(product string, amount sdk.DecCoin, depositor sdk.AccAddress) MsgWithdraw {
	return MsgWithdraw{product, amount, depositor}
}

// Route Implements Msg
func (msg MsgWithdraw) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgWithdraw) Type() string { return typeMsgWithdraw }

// ValidateBasic Implements Msg
func (msg MsgWithdraw) ValidateBasic() error {
	if msg.Depositor.Empty() {
		return sdkerror.Wrapf(sdkerror.ErrInvalidAddress, msg.Depositor.String())
	}
	if !msg.Amount.IsValid() || !msg.Amount.IsPositive() {
		return sdkerror.Wrapf(sdkerror.ErrInvalidCoins, msg.Amount.String())
	}

	return nil
}

// GetSignBytes Implements Msg
func (msg MsgWithdraw) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners Implements Msg
func (msg MsgWithdraw) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Depositor}
}

// MsgTransferOwnership - high level transaction of the dex module
var _ sdk.Msg = &MsgTransferOwnership{}

// NewMsgTransferOwnership create a new MsgTransferOwnership
func NewMsgTransferOwnership(from, to sdk.AccAddress, product string, pub, sig []byte) MsgTransferOwnership {
	return MsgTransferOwnership{
		FromAddress: from,
		ToAddress:   to,
		Product:     product,
		Pubkey:      pub,
		ToSignature: sig,
	}
}

// Route Implements Msg
func (msg MsgTransferOwnership) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgTransferOwnership) Type() string { return typeMsgTransferOwnership }

// ValidateBasic Implements Msg
func (msg MsgTransferOwnership) ValidateBasic() error {
	if msg.FromAddress.Empty() {
		return sdkerror.Wrapf(sdkerror.ErrInvalidAddress, "missing sender address")
	}

	if msg.ToAddress.Empty() {
		return sdkerror.Wrapf(sdkerror.ErrInvalidAddress, "missing recipient address")
	}

	if msg.Product == "" {
		return sdkerror.Wrapf(sdkerror.ErrUnknownRequest, "product cannot be empty")
	}

	if !msg.checkMultiSign() {
		return sdkerror.Wrapf(sdkerror.ErrUnauthorized, "invalid multi signature")
	}
	return nil
}

// GetSignBytes Implements Msg
func (msg MsgTransferOwnership) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners Implements Msg
func (msg MsgTransferOwnership) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.FromAddress}
}

func (msg MsgTransferOwnership) checkMultiSign() bool {
	// check pubkey
	if msg.Pubkey == nil {
		return false
	}

	stdSign := authtypes.StdSignature {
		msg.Pubkey,
		msg.ToSignature,
	}

	if !sdk.AccAddress(stdSign.GetPubKey().Address()).Equals(msg.ToAddress) {
		return false
	}

	// check multisign
	toValid := stdSign.GetPubKey().VerifyBytes(msg.GetSignBytes(), msg.ToSignature)
	return toValid
}
