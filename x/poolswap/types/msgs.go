package types

import (
	"fmt"
	sdkerror "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/okex/okchain/x/common"
)

// PoolSwap message types and routes
const (
	TypeMsgAddLiquidity = "add_liquidity"
	TypeMsgTokenSwap    = "token_swap"
)

// MsgAddLiquidity Deposit quote_amount and base_amount at current ratio to mint pool tokens.
var _ sdk.Msg = &MsgAddLiquidity{}

// NewMsgAddLiquidity is a constructor function for MsgAddLiquidity
func NewMsgAddLiquidity(minLiquidity sdk.Dec, maxBaseAmount, quoteAmount sdk.DecCoin, deadline int64, sender sdk.AccAddress) MsgAddLiquidity {
	return MsgAddLiquidity{
		MinLiquidity:  minLiquidity,
		MaxBaseAmount: maxBaseAmount,
		QuoteAmount:   quoteAmount,
		Deadline:      deadline,
		Sender:        sender,
	}
}

// Route should return the name of the module
func (msg MsgAddLiquidity) Route() string { return RouterKey }

// Type should return the action
func (msg MsgAddLiquidity) Type() string { return "add_liquidity" }

// ValidateBasic runs stateless checks on the message
func (msg MsgAddLiquidity) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerror.Wrap(sdkerror.ErrInvalidAddress, msg.Sender.String())
	}
	if !(msg.MaxBaseAmount.IsPositive() && msg.QuoteAmount.IsPositive()) {
		return sdkerror.Wrap(sdkerror.ErrUnknownRequest, "token amount must be positive")
	}
	if !msg.MaxBaseAmount.IsValid() {
		return sdkerror.Wrap(sdkerror.ErrUnknownRequest, "invalid MaxBaseAmount")
	}
	if !msg.QuoteAmount.IsValid() {
		return sdkerror.Wrap(sdkerror.ErrUnknownRequest, "invalid QuoteAmount")
	}
	if msg.QuoteAmount.Denom != common.NativeToken {
		return sdkerror.Wrap(sdkerror.ErrUnknownRequest, "quote token only supports " + common.NativeToken)
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgAddLiquidity) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgAddLiquidity) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// GetSwapTokenPair defines token pair
func (msg MsgAddLiquidity) GetSwapTokenPair() string {
	return msg.MaxBaseAmount.Denom + "_" + msg.QuoteAmount.Denom
}

// MsgRemoveLiquidity burns pool tokens to withdraw okt and Tokens at current ratio.
var _ sdk.Msg = &MsgRemoveLiquidity{}

// NewMsgRemoveLiquidity is a constructor function for MsgAddLiquidity
func NewMsgRemoveLiquidity(liquidity sdk.Dec, minBaseAmount, minQuoteAmount sdk.DecCoin, deadline int64, sender sdk.AccAddress) MsgRemoveLiquidity {
	return MsgRemoveLiquidity{
		Liquidity:      liquidity,
		MinBaseAmount:  minBaseAmount,
		MinQuoteAmount: minQuoteAmount,
		Deadline:       deadline,
		Sender:         sender,
	}
}

// Route should return the name of the module
func (msg MsgRemoveLiquidity) Route() string { return RouterKey }

// Type should return the action
func (msg MsgRemoveLiquidity) Type() string { return "remove_liquidity" }

// ValidateBasic runs stateless checks on the message
func (msg MsgRemoveLiquidity) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerror.Wrap(sdkerror.ErrInvalidAddress, msg.Sender.String())
	}
	if !(msg.Liquidity.IsPositive()) {
		return sdkerror.Wrap(sdkerror.ErrUnknownRequest, "token amount must be positive")
	}
	if !msg.MinBaseAmount.IsValid() {
		return sdkerror.Wrap(sdkerror.ErrUnknownRequest, "invalid MinBaseAmount")
	}
	if !msg.MinQuoteAmount.IsValid() {
		return sdkerror.Wrap(sdkerror.ErrUnknownRequest, "invalid MinQuoteAmount")
	}
	if msg.MinQuoteAmount.Denom != common.NativeToken {
		return sdkerror.Wrap(sdkerror.ErrUnknownRequest, "quote token only supports " + common.NativeToken)
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgRemoveLiquidity) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgRemoveLiquidity) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// GetSwapTokenPair defines token pair
func (msg MsgRemoveLiquidity) GetSwapTokenPair() string {
	return msg.MinBaseAmount.Denom + "_" + msg.MinQuoteAmount.Denom
}

// MsgCreateExchange creates a new exchange with token
var _ sdk.Msg = &MsgCreateExchange{}

// NewMsgCreateExchange create a new exchange with token
func NewMsgCreateExchange(token string, sender sdk.AccAddress) MsgCreateExchange {
	return MsgCreateExchange{
		Token:  token,
		Sender: sender,
	}
}

// Route should return the name of the module
func (msg MsgCreateExchange) Route() string { return RouterKey }

// Type should return the action
func (msg MsgCreateExchange) Type() string { return "create_exchange" }

// ValidateBasic runs stateless checks on the message
func (msg MsgCreateExchange) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerror.Wrap(sdkerror.ErrInvalidAddress, msg.Sender.String())
	}
	if sdk.ValidateDenom(msg.Token) != nil || ValidatePoolTokenName(msg.Token) {
		return sdkerror.Wrap(sdkerror.ErrUnknownRequest, "invalid Token")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgCreateExchange) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgCreateExchange) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// MsgTokenToNativeToken define the message for swap between token and DefaultBondDenom
var _ sdk.Msg = &MsgTokenToNativeToken{}

// NewMsgTokenToNativeToken is a constructor function for MsgTokenOKTSwap
func NewMsgTokenToNativeToken(
	soldTokenAmount, minBoughtTokenAmount sdk.DecCoin, deadline int64, recipient, sender sdk.AccAddress,
) MsgTokenToNativeToken {
	return MsgTokenToNativeToken{
		SoldTokenAmount:      soldTokenAmount,
		MinBoughtTokenAmount: minBoughtTokenAmount,
		Deadline:             deadline,
		Recipient:            recipient,
		Sender:               sender,
	}
}

// Route should return the name of the module
func (msg MsgTokenToNativeToken) Route() string { return RouterKey }

// Type should return the action
func (msg MsgTokenToNativeToken) Type() string { return TypeMsgTokenSwap }

// ValidateBasic runs stateless checks on the message
func (msg MsgTokenToNativeToken) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerror.Wrap(sdkerror.ErrInvalidAddress, msg.Sender.String())
	}

	if msg.Recipient.Empty() {
		return sdkerror.Wrap(sdkerror.ErrInvalidAddress, msg.Recipient.String())
	}

	if msg.SoldTokenAmount.Denom != sdk.DefaultBondDenom && msg.MinBoughtTokenAmount.Denom != sdk.DefaultBondDenom {
		return sdkerror.Wrap(sdkerror.ErrUnknownRequest, fmt.Sprintf("both token to sell and token to buy do not contain %s,"+
			" quote token only supports %s", sdk.DefaultBondDenom, sdk.DefaultBondDenom))
	}
	if !(msg.SoldTokenAmount.IsPositive()) {
		return sdkerror.Wrap(sdkerror.ErrUnknownRequest, "token amount must be positive")
	}
	if !msg.SoldTokenAmount.IsValid() {
		return sdkerror.Wrap(sdkerror.ErrUnknownRequest, "invalid SoldTokenAmount")
	}

	if !msg.MinBoughtTokenAmount.IsValid() {
		return sdkerror.Wrap(sdkerror.ErrUnknownRequest, "invalid MinBoughtTokenAmount")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgTokenToNativeToken) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgTokenToNativeToken) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// GetSwapTokenPair defines token pair
func (msg MsgTokenToNativeToken) GetSwapTokenPair() string {
	if msg.SoldTokenAmount.Denom == sdk.DefaultBondDenom {
		return msg.MinBoughtTokenAmount.Denom + "_" + msg.SoldTokenAmount.Denom
	}
	return msg.SoldTokenAmount.Denom + "_" + msg.MinBoughtTokenAmount.Denom
}
