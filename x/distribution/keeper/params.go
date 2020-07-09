package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/okex/okchain/x/params"
)

func tmpValidate(value interface{}) error {
	return nil
}

// ParamKeyTable is the type declaration for parameters
func ParamKeyTable() params.KeyTable {
	pairs := params.ParamSetPairs{
		{ParamStoreKeyCommunityTax, sdk.Dec{}, tmpValidate},
		{ParamStoreKeyWithdrawAddrEnabled, true, tmpValidate},
	}
	return params.NewKeyTable(pairs...)
}

// GetCommunityTax returns the current CommunityTax rate from the global param store
// nolint: errcheck
func (k Keeper) GetCommunityTax(ctx sdk.Context) sdk.Dec {
	var percent sdk.Dec
	k.paramSpace.Get(ctx, ParamStoreKeyCommunityTax, &percent)
	return percent
}

// SetCommunityTax sets the value of community tax
// nolint: errcheck
func (k Keeper) SetCommunityTax(ctx sdk.Context, percent sdk.Dec) {
	k.paramSpace.Set(ctx, ParamStoreKeyCommunityTax, &percent)
}

// GetWithdrawAddrEnabled returns the current WithdrawAddrEnabled
// nolint: errcheck
func (k Keeper) GetWithdrawAddrEnabled(ctx sdk.Context) bool {
	var enabled bool
	k.paramSpace.Get(ctx, ParamStoreKeyWithdrawAddrEnabled, &enabled)
	return enabled
}

// SetWithdrawAddrEnabled sets the value of enabled
// nolint: errcheck
func (k Keeper) SetWithdrawAddrEnabled(ctx sdk.Context, enabled bool) {
	k.paramSpace.Set(ctx, ParamStoreKeyWithdrawAddrEnabled, &enabled)
}
