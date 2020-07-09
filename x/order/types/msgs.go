package types

import (
	"fmt"
	sdkerror "github.com/cosmos/cosmos-sdk/types/errors"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"strconv"
)

// nolint
const (
	OrderItemLimit            = 200
	MultiCancelOrderItemLimit = 200
)

// NewMsgNewOrder is a constructor function for MsgNewOrder
func NewMsgNewOrder(sender sdk.AccAddress, product string, side string, price string,
	quantity string) MsgNewOrders {

	return MsgNewOrders{
		Sender: sender,
		OrderItems: []OrderItem{
			{
				Product:  product,
				Side:     side,
				Price:    sdk.MustNewDecFromStr(price),
				Quantity: sdk.MustNewDecFromStr(quantity),
			},
		},
	}
}

// NewMsgCancelOrder is a constructor function for MsgCancelOrder
func NewMsgCancelOrder(sender sdk.AccAddress, orderID string) MsgCancelOrders {
	msgCancelOrder := MsgCancelOrders{
		Sender:   sender,
		OrderIDs: []string{orderID},
	}
	return msgCancelOrder
}

//********************MsgNewOrders*************
// nolint
var _ sdk.Msg = &MsgNewOrders{}

// nolint
func NewOrderItem(product string, side string, price string,
	quantity string) OrderItem {
	return OrderItem{
		Product:  product,
		Side:     side,
		Price:    sdk.MustNewDecFromStr(price),
		Quantity: sdk.MustNewDecFromStr(quantity),
	}
}

// NewMsgNewOrders is a constructor function for MsgNewOrder
func NewMsgNewOrders(sender sdk.AccAddress, orderItems []OrderItem) MsgNewOrders {
	return MsgNewOrders{
		Sender:     sender,
		OrderItems: orderItems,
	}
}

// nolint
func (msg MsgNewOrders) Route() string { return "order" }

// nolint
func (msg MsgNewOrders) Type() string { return "new" }

// ValidateBasic : Implements Msg.
func (msg MsgNewOrders) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerror.Wrap(sdkerror.ErrInvalidAddress, msg.Sender.String())
	}
	if msg.OrderItems == nil || len(msg.OrderItems) == 0 {
		return sdkerror.Wrap(sdkerror.ErrUnknownRequest, "invalid OrderItems")
	}
	if len(msg.OrderItems) > OrderItemLimit {
		return sdkerror.Wrap(sdkerror.ErrUnknownRequest, "Numbers of NewOrderItem should not be more than "+strconv.Itoa(OrderItemLimit))
	}
	for _, item := range msg.OrderItems {
		if len(item.Product) == 0 {
			return sdkerror.Wrap(sdkerror.ErrUnknownRequest, "Product cannot be empty")
		}
		symbols := strings.Split(item.Product, "_")
		if len(symbols) != 2 {
			return sdkerror.Wrap(sdkerror.ErrUnknownRequest, "Product should be in the format of \"base_quote\"")
		}
		if symbols[0] == symbols[1] {
			return sdkerror.Wrap(sdkerror.ErrUnknownRequest, "invalid product")
		}
		if item.Side != BuyOrder && item.Side != SellOrder {
			return sdkerror.Wrap(sdkerror.ErrUnknownRequest,
				fmt.Sprintf("Side is expected to be \"BUY\" or \"SELL\", but got \"%s\"", item.Side))
		}
		if !(item.Price.IsPositive() && item.Quantity.IsPositive()) {
			return sdkerror.Wrap(sdkerror.ErrUnknownRequest, "Price/Quantity must be positive")
		}
	}

	return nil
}

// GetSignBytes : encodes the message for signing
func (msg MsgNewOrders) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners defines whose signature is required
func (msg MsgNewOrders) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// Calculate customize gas
func (msg MsgNewOrders) CalculateGas(gasUnit uint64) uint64 {
	return uint64(len(msg.OrderItems)) * gasUnit
}

// nolint
var _ sdk.Msg = &MsgCancelOrders{}

// NewMsgCancelOrders is a constructor function for MsgCancelOrder
func NewMsgCancelOrders(sender sdk.AccAddress, orderIDItems []string) MsgCancelOrders {
	msgCancelOrder := MsgCancelOrders{
		Sender:   sender,
		OrderIDs: orderIDItems,
	}
	return msgCancelOrder
}

// nolint
func (msg MsgCancelOrders) Route() string { return "order" }

// nolint
func (msg MsgCancelOrders) Type() string { return "cancel" }

// nolint
func (msg MsgCancelOrders) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerror.Wrap(sdkerror.ErrInvalidAddress, msg.Sender.String())
	}
	if msg.OrderIDs == nil || len(msg.OrderIDs) == 0 {
		return sdkerror.Wrap(sdkerror.ErrUnknownRequest, "invalid OrderIDs")
	}
	if len(msg.OrderIDs) > MultiCancelOrderItemLimit {
		return sdkerror.Wrap(sdkerror.ErrUnknownRequest, "Numbers of CancelOrderItem should not be more than " + strconv.Itoa(OrderItemLimit))
	}
	if hasDuplicatedID(msg.OrderIDs) {
		return sdkerror.Wrap(sdkerror.ErrUnknownRequest, "Duplicated order ids detected")
	}
	for _, item := range msg.OrderIDs {
		if item == "" {
			return sdkerror.Wrap(sdkerror.ErrUnauthorized, "orderID cannot be empty")
		}
	}

	return nil
}

func hasDuplicatedID(ids []string) bool {
	idSet := make(map[string]bool)
	for _, item := range ids {
		if !idSet[item] {
			idSet[item] = true
		} else {
			return true
		}
	}
	return false
}

// GetSignBytes encodes the message for signing
func (msg MsgCancelOrders) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners defines whose signature is required
func (msg MsgCancelOrders) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// Calculate customize gas
func (msg MsgCancelOrders) CalculateGas(gasUnit uint64) uint64 {
	return uint64(len(msg.OrderIDs)) * gasUnit
}

// nolint
type OrderResult struct {
	Code    string `json:"code"`    // order return code
	Message string `json:"msg"`     // order return error message
	OrderID string `json:"orderid"` // order return orderid
}
