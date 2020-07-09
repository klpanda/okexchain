package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/okex/okchain/x/staking/exported"
)

var _ exported.DelegatorI = &Delegator{}

// NewDelegator creates a new Delegator object
func NewDelegator(delAddr sdk.AccAddress) Delegator {
	return Delegator{
		delAddr,
		nil,
		sdk.ZeroDec(),
		sdk.ZeroDec(),
		false,
		sdk.ZeroDec(),
		nil,
	}
}

// GetShareAddedValidatorAddresses gets validator address that the delegator added shares to for other module
func (d Delegator) GetShareAddedValidatorAddresses() []sdk.ValAddress {
	return d.ValidatorAddresses
}

// GetLastAddedShares gets the last shares added to validators of a delegator for other module
func (d Delegator) GetLastAddedShares() sdk.Dec {
	return d.Shares
}

// RegProxy registers or deregisters the identity of proxy
func (d *Delegator) RegProxy(reg bool) {
	d.IsProxy = reg
	if reg {
		d.UnbindProxy()
	}
}

// BindProxy sets relationship between a delegator and proxy
func (d *Delegator) BindProxy(proxyAddr sdk.AccAddress) {
	d.ProxyAddress = proxyAddr
	d.IsProxy = false
}

// UnbindProxy clears the proxy address on a delegator
func (d *Delegator) UnbindProxy() {
	d.ProxyAddress = nil
}

// HasProxy tells whether the delegator has bound a proxy
func (d Delegator) HasProxy() bool {
	return d.ProxyAddress != nil
}

// MustUnMarshalDelegator must return a delegator entity by unmarshalling
func MustUnMarshalDelegator(cdc codec.Marshaler, value []byte) (delegator Delegator) {
	cdc.MustUnmarshalBinaryBare(value, &delegator)
	return
}
