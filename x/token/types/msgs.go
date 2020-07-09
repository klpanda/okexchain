// nolint
package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerror "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

const (
	DescLenLimit   = 256
	MultiSendLimit = 1000

	// 90 billion
	TotalSupplyUpperbound = int64(9 * 1e10)
)

var _ sdk.Msg = &MsgTokenIssue{}

func NewMsgTokenIssue(tokenDescription, symbol, originalSymbol, wholeName, totalSupply string, owner sdk.AccAddress, mintable bool) MsgTokenIssue {
	return MsgTokenIssue{
		Description:    tokenDescription,
		Symbol:         symbol,
		OriginalSymbol: originalSymbol,
		WholeName:      wholeName,
		TotalSupply:    totalSupply,
		Owner:          owner,
		Mintable:       mintable,
	}
}

func (msg MsgTokenIssue) Route() string { return RouterKey }

func (msg MsgTokenIssue) Type() string { return "issue" }

func (msg MsgTokenIssue) ValidateBasic() error {
	// check owner
	if msg.Owner.Empty() {
		return sdkerror.Wrapf(sdkerror.ErrInvalidAddress, msg.Owner.String())
	}

	// check original symbol
	if len(msg.OriginalSymbol) == 0 {
		return sdkerror.Wrapf(sdkerror.ErrUnknownRequest, "failed to check issue msg because original symbol cannot be empty")
	}
	if !ValidOriginalSymbol(msg.OriginalSymbol) {
		return sdkerror.Wrapf(sdkerror.ErrUnknownRequest, "failed to check issue msg because invalid original symbol: " + msg.OriginalSymbol)
	}

	// check wholeName
	isValid := wholeNameValid(msg.WholeName)
	if !isValid {
		return sdkerror.Wrapf(sdkerror.ErrUnknownRequest, "failed to check issue msg because invalid wholename")
	}
	// check desc
	if len(msg.Description) > DescLenLimit {
		return sdkerror.Wrapf(sdkerror.ErrUnknownRequest, "failed to check issue msg because invalid desc")
	}
	// check totalSupply
	totalSupply, err := sdk.NewDecFromStr(msg.TotalSupply)
	if err != nil {
		return err
	}
	if totalSupply.GT(sdk.NewDec(TotalSupplyUpperbound)) || totalSupply.LTE(sdk.ZeroDec()) {
		return sdkerror.Wrapf(sdkerror.ErrUnknownRequest, "failed to check issue msg because invalid total supply")
	}
	return nil
}

func (msg MsgTokenIssue) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgTokenIssue) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Owner}
}

var _ sdk.Msg = &MsgTokenBurn{}

func NewMsgTokenBurn(amount sdk.DecCoin, owner sdk.AccAddress) MsgTokenBurn {
	return MsgTokenBurn{
		Amount: amount,
		Owner:  owner,
	}
}

func (msg MsgTokenBurn) Route() string { return RouterKey }

func (msg MsgTokenBurn) Type() string { return "burn" }

func (msg MsgTokenBurn) ValidateBasic() error {
	// check owner
	if msg.Owner.Empty() {
		return sdkerror.Wrapf(sdkerror.ErrInvalidAddress, msg.Owner.String())
	}
	if !msg.Amount.IsValid() {
		return sdkerror.Wrapf(sdkerror.ErrInsufficientFunds, "failed to check burn msg because invalid Coins: " + msg.Amount.String())
	}

	return nil
}

func (msg MsgTokenBurn) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgTokenBurn) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Owner}
}

var _ sdk.Msg = &MsgTokenMint{}

func NewMsgTokenMint(amount sdk.DecCoin, owner sdk.AccAddress) MsgTokenMint {
	return MsgTokenMint{
		Amount: amount,
		Owner:  owner,
	}
}

func (msg MsgTokenMint) Route() string { return RouterKey }

func (msg MsgTokenMint) Type() string { return "mint" }

func (msg MsgTokenMint) ValidateBasic() error {
	if msg.Owner.Empty() {
		return sdkerror.Wrapf(sdkerror.ErrInvalidAddress, msg.Owner.String())
	}

	amount := msg.Amount.Amount
	if amount.GT(sdk.NewDec(TotalSupplyUpperbound)) {
		return sdkerror.Wrapf(sdkerror.ErrUnknownRequest, "failed to check mint msg because invalid amount")
	}
	if !msg.Amount.IsValid() {
		return sdkerror.Wrapf(sdkerror.ErrInsufficientFunds, "failed to check mint msg because invalid Coins: " + msg.Amount.String())
	}
	return nil
}

func (msg MsgTokenMint) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)

	return sdk.MustSortJSON(bz)
}

func (msg MsgTokenMint) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Owner}
}

var _ sdk.Msg = &MsgMultiSend{}

func NewMsgMultiSend(from sdk.AccAddress, transfers []TransferUnit) MsgMultiSend {
	return MsgMultiSend{
		From:      from,
		Transfers: transfers,
	}
}

func (msg MsgMultiSend) Route() string { return RouterKey }

func (msg MsgMultiSend) Type() string { return "multi-send" }

func (msg MsgMultiSend) ValidateBasic() error {
	if msg.From.Empty() {
		return sdkerror.Wrapf(sdkerror.ErrInvalidAddress, msg.From.String())
	}

	// check transfers
	if len(msg.Transfers) > MultiSendLimit {
		return sdkerror.Wrapf(sdkerror.ErrUnknownRequest, "failed to check multisend msg because restrictions on the number of transfers")
	}
	for _, transfer := range msg.Transfers {
		if !transfer.Coins.IsAllPositive() || !transfer.Coins.IsValid() {
			return sdkerror.Wrapf(sdkerror.ErrInvalidCoins, "failed to check multisend msg because send amount must be positive")
		}

		if transfer.To.Empty() {
			return sdkerror.Wrapf(sdkerror.ErrInvalidAddress, "failed to check multisend msg because address is empty, not valid")
		}
	}
	return nil
}

func (msg MsgMultiSend) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgMultiSend) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.From}
}

var _ sdk.Msg = &MsgSend{}

func NewMsgTokenSend(from, to sdk.AccAddress, coins sdk.DecCoins) MsgSend {
	return MsgSend{
		FromAddress: from,
		ToAddress:   to,
		Amount:      coins,
	}
}

func (msg MsgSend) Route() string { return RouterKey }

func (msg MsgSend) Type() string { return "send" }

func (msg MsgSend) ValidateBasic() error {
	if msg.FromAddress.Empty() {
		return sdkerror.Wrapf(sdkerror.ErrInvalidAddress, "failed to check send msg because miss sender address")
	}
	if msg.ToAddress.Empty() {
		return sdkerror.Wrapf(sdkerror.ErrInvalidAddress, "failed to check send msg because miss recipient address")
	}
	if !msg.Amount.IsValid() {
		return sdkerror.Wrapf(sdkerror.ErrInvalidCoins, "failed to check send msg because send amount is invalid: " + msg.Amount.String())
	}
	if !msg.Amount.IsAllPositive() {
		return sdkerror.Wrapf(sdkerror.ErrInsufficientFunds, "failed to check send msg because send amount must be positive")
	}
	return nil
}

func (msg MsgSend) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgSend) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.FromAddress}
}

// MsgTransferOwnership - high level transaction of the coin module
var _ sdk.Msg = &MsgTransferOwnership{}

func NewMsgTransferOwnership(from, to sdk.AccAddress, symbol string) MsgTransferOwnership {
	return MsgTransferOwnership{
		FromAddress: from,
		ToAddress:   to,
		Symbol:      symbol,
	}
}

func (msg MsgTransferOwnership) Route() string { return RouterKey }

func (msg MsgTransferOwnership) Type() string { return "transfer" }

func (msg MsgTransferOwnership) ValidateBasic() error {
	if msg.FromAddress.Empty() {
		return sdkerror.Wrapf(sdkerror.ErrInvalidAddress, "failed to check transferownership msg because miss sender address")
	}
	if msg.ToAddress.Empty() {
		return sdkerror.Wrapf(sdkerror.ErrInvalidAddress, "failed to check transferownership msg because miss recipient address")
	}
	if len(msg.Symbol) == 0 {
		return sdkerror.Wrapf(sdkerror.ErrUnknownRequest, "failed to check transferownership msg because symbol cannot be empty")
	}

	if sdk.ValidateDenom(msg.Symbol) != nil {
		return sdkerror.Wrapf(sdkerror.ErrUnknownRequest, "failed to check transferownership msg because invalid token symbol: " + msg.Symbol)
	}

	if !msg.checkMultiSign() {
		return sdkerror.Wrapf(sdkerror.ErrUnauthorized, "failed to check transferownership msg because invalid multi signature")
	}
	return nil
}

func (msg MsgTransferOwnership) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

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

var _ sdk.Msg = &MsgTokenModify{}

func NewMsgTokenModify(symbol, desc, wholeName string, isDescEdit, isWholeNameEdit bool, owner sdk.AccAddress) MsgTokenModify {
	return MsgTokenModify{
		Symbol:                symbol,
		IsDescriptionModified: isDescEdit,
		Description:           desc,
		IsWholeNameModified:   isWholeNameEdit,
		WholeName:             wholeName,
		Owner:                 owner,
	}
}

func (msg MsgTokenModify) Route() string { return RouterKey }

func (msg MsgTokenModify) Type() string { return "edit" }

func (msg MsgTokenModify) ValidateBasic() error {
	// check owner
	if msg.Owner.Empty() {
		return sdkerror.Wrapf(sdkerror.ErrInvalidAddress, msg.Owner.String())
	}
	// check symbol
	if len(msg.Symbol) == 0 {
		return sdkerror.Wrapf(sdkerror.ErrUnknownRequest, "failed to check modify msg because symbol cannot be empty")
	}
	if sdk.ValidateDenom(msg.Symbol) != nil {
		return sdkerror.Wrapf(sdkerror.ErrUnknownRequest, "failed to check modify msg because invalid token symbol: " + msg.Symbol)
	}
	// check wholeName
	if msg.IsWholeNameModified {
		isValid := wholeNameValid(msg.WholeName)
		if !isValid {
			return sdkerror.Wrapf(sdkerror.ErrUnknownRequest, "failed to check modify msg because invalid wholename")
		}
	}
	// check desc
	if msg.IsDescriptionModified {
		if len(msg.Description) > DescLenLimit {
			return sdkerror.Wrapf(sdkerror.ErrUnknownRequest, "failed to check modify msg because invalid desc")
		}
	}
	return nil
}

func (msg MsgTokenModify) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgTokenModify) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Owner}
}
