package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ensure Msg interface compliance at compile time
var (
	_ sdk.Msg = (*MsgAddShares)(nil)
	_ sdk.Msg = (*MsgDestroyValidator)(nil)
	_ sdk.Msg = (*MsgUnbindProxy)(nil)
	_ sdk.Msg = (*MsgRegProxy)(nil)
	_ sdk.Msg = (*MsgBindProxy)(nil)
	_ sdk.Msg = (*MsgDeposit)(nil)
	_ sdk.Msg = (*MsgWithdraw)(nil)
)

// NewMsgDestroyValidator creates a msg of destroy-validator
func NewMsgDestroyValidator(delAddr sdk.AccAddress) MsgDestroyValidator {
	return MsgDestroyValidator{
		DelAddr: delAddr,
	}
}

// nolint
func (MsgDestroyValidator) Route() string { return RouterKey }
func (MsgDestroyValidator) Type() string  { return "destroy_validator" }
func (msg MsgDestroyValidator) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.DelAddr}
}

// ValidateBasic gives a quick validity check
func (msg MsgDestroyValidator) ValidateBasic() error {
	if msg.DelAddr.Empty() {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}

	return nil
}

// GetSignBytes returns the message bytes to sign over
func (msg MsgDestroyValidator) GetSignBytes() []byte {
	bytes := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bytes)
}

// NewMsgUnbindProxy creates a msg of unbinding proxy
func NewMsgUnbindProxy(delAddr sdk.AccAddress) MsgUnbindProxy {
	return MsgUnbindProxy{
		DelAddr: delAddr,
	}
}

// nolint
func (MsgUnbindProxy) Route() string { return RouterKey }
func (MsgUnbindProxy) Type() string  { return "unbind_proxy" }
func (msg MsgUnbindProxy) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.DelAddr}
}

// ValidateBasic gives a quick validity check
func (msg MsgUnbindProxy) ValidateBasic() error {
	if msg.DelAddr.Empty() {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	return nil
}

// GetSignBytes returns the message bytes to sign over
func (msg MsgUnbindProxy) GetSignBytes() []byte {
	bytes := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bytes)
}

// NewMsgRegProxy creates a msg of registering proxy
func NewMsgRegProxy(proxyAddress sdk.AccAddress, reg bool) MsgRegProxy {
	return MsgRegProxy{
		ProxyAddress: proxyAddress,
		Reg:          reg,
	}
}

// nolint
func (MsgRegProxy) Route() string { return RouterKey }
func (MsgRegProxy) Type() string  { return "reg_or_unreg_proxy" }
func (msg MsgRegProxy) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.ProxyAddress}
}

// ValidateBasic gives a quick validity check
func (msg MsgRegProxy) ValidateBasic() error {
	if msg.ProxyAddress.Empty() {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	return nil
}

// GetSignBytes returns the message bytes to sign over
func (msg MsgRegProxy) GetSignBytes() []byte {
	bytes := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bytes)
}

// NewMsgBindProxy creates a msg of binding proxy
func NewMsgBindProxy(delAddr sdk.AccAddress, ProxyDelAddr sdk.AccAddress) MsgBindProxy {
	return MsgBindProxy{
		DelAddr:      delAddr,
		ProxyAddress: ProxyDelAddr,
	}
}

// nolint
func (MsgBindProxy) Route() string { return RouterKey }
func (MsgBindProxy) Type() string  { return "bind_proxy" }
func (msg MsgBindProxy) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.DelAddr}
}

// ValidateBasic gives a quick validity check
func (msg MsgBindProxy) ValidateBasic() error {
	if msg.DelAddr.Empty() || msg.ProxyAddress.Empty() {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}

	if msg.DelAddr.Equals(msg.ProxyAddress) {
		return ErrWrongOperationAddr(DefaultCodespace,
			fmt.Sprintf("ProxyAddress: %s eqauls to DelegatorAddress: %s",
				msg.ProxyAddress.String(), msg.DelAddr.String()))
	}

	return nil
}

// GetSignBytes returns the message bytes to sign over
func (msg MsgBindProxy) GetSignBytes() []byte {
	bytes := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bytes)
}

// NewMsgAddShares creates a msg of adding shares to vals
func NewMsgAddShares(delAddr sdk.AccAddress, valAddrs []sdk.ValAddress) MsgAddShares {
	return MsgAddShares{
		DelAddr:  delAddr,
		ValAddrs: valAddrs,
	}
}

// nolint
func (MsgAddShares) Route() string { return RouterKey }
func (MsgAddShares) Type() string  { return "add_shares_to_validators" }
func (msg MsgAddShares) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.DelAddr}
}

// ValidateBasic gives a quick validity check
func (msg MsgAddShares) ValidateBasic() error {
	if msg.DelAddr.Empty() {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}

	if msg.ValAddrs == nil || len(msg.ValAddrs) == 0 {
		return ErrWrongOperationAddr(DefaultCodespace, "ValAddrs is empty")
	}

	if isValsDuplicate(msg.ValAddrs) {
		return ErrTargetValsDuplicate(DefaultCodespace)
	}

	return nil
}

// GetSignBytes returns the message bytes to sign over
func (msg MsgAddShares) GetSignBytes() []byte {
	bytes := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bytes)
}

func isValsDuplicate(valAddrs []sdk.ValAddress) bool {
	lenAddrs := len(valAddrs)
	filter := make(map[string]struct{}, lenAddrs)
	for i := 0; i < lenAddrs; i++ {
		key := valAddrs[i].String()
		if _, ok := filter[key]; ok {
			return true
		}
		filter[key] = struct{}{}
	}

	return false
}

// NewMsgDeposit creates a new instance of MsgDeposit
func NewMsgDeposit(delAddr sdk.AccAddress, amount sdk.DecCoin) MsgDeposit {
	return MsgDeposit{
		DelegatorAddress: delAddr,
		Amount:           amount,
	}
}

// nolint
func (msg MsgDeposit) Route() string { return RouterKey }
func (msg MsgDeposit) Type() string  { return "deposit" }
func (msg MsgDeposit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.DelegatorAddress}
}

// ValidateBasic gives a quick validity check
func (msg MsgDeposit) ValidateBasic() error {
	if msg.DelegatorAddress.Empty() {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	if msg.Amount.Amount.LTE(sdk.ZeroDec()) || !msg.Amount.IsValid() {
		return ErrBadDelegationAmount(DefaultCodespace)
	}
	return nil
}

// GetSignBytes returns the message bytes to sign over
func (msg MsgDeposit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// NewMsgWithdraw creates a new instance of MsgWithdraw
func NewMsgWithdraw(delAddr sdk.AccAddress, amount sdk.DecCoin) MsgWithdraw {
	return MsgWithdraw{
		DelegatorAddress: delAddr,
		Amount:           amount,
	}
}

// nolint
func (msg MsgWithdraw) Route() string { return RouterKey }
func (msg MsgWithdraw) Type() string  { return "withdraw" }
func (msg MsgWithdraw) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.DelegatorAddress}
}

// ValidateBasic gives a quick validity check
func (msg MsgWithdraw) ValidateBasic() error {
	if msg.DelegatorAddress.Empty() {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	if !msg.Amount.IsValid() {
		return ErrBadUnDelegationAmount(DefaultCodespace)
	}
	return nil
}

// GetSignBytes returns the message bytes to sign over
func (msg MsgWithdraw) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}
