package types

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewUndelegationInfo creates a new delegation object
func NewUndelegationInfo(delegatorAddr sdk.AccAddress, sharesQuantity Shares, completionTime time.Time) UndelegationInfo {
	return UndelegationInfo{
		DelegatorAddress: delegatorAddr,
		Quantity:         sharesQuantity,
		CompletionTime:   completionTime,
	}
}

// MustUnMarshalUndelegationInfo must return the UndelegationInfo object by unmarshaling
func MustUnMarshalUndelegationInfo(cdc codec.Marshaler, value []byte) UndelegationInfo {
	undelegationInfo, err := UnmarshalUndelegationInfo(cdc, value)
	if err != nil {
		panic(err)
	}
	return undelegationInfo
}

// UnmarshalUndelegationInfo returns the UndelegationInfo object by unmarshaling
func UnmarshalUndelegationInfo(cdc codec.Marshaler, value []byte) (undelegationInfo UndelegationInfo, err error) {
	err = cdc.UnmarshalBinaryBare(value, &undelegationInfo)
	return undelegationInfo, err
}

// String returns a human readable string representation of UndelegationInfo
func (ud UndelegationInfo) String() string {
	return fmt.Sprintf(`UnDelegation:
  Delegator: %s
  Quantity:    %s
  CompletionTime:    %s`,
		ud.DelegatorAddress, ud.Quantity, ud.CompletionTime.Format(time.RFC3339))
}

// DefaultUndelegation returns default entity for UndelegationInfo
func DefaultUndelegation() UndelegationInfo {
	return UndelegationInfo{
		nil, sdk.ZeroDec(), time.Unix(0, 0).UTC(),
	}
}
