package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerror "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/okex/okchain/x/common/proto"
	"github.com/okex/okchain/x/upgrade/types"
)

// only for unit test

// SetAppUpgradeConfig sets app upgrade config for test
// deprecated
func (k Keeper) SetAppUpgradeConfig(ctx sdk.Context, proposalID, version, upgradeHeight uint64, software string,
) error {
	if _, found := k.GetAppUpgradeConfig(ctx); found {
		return sdkerror.Wrap(sdkerror.ErrInternal, "failed. an app upgrade config has existed, only one entry is permitted")
	}

	appUpgradeConfig := proto.NewAppUpgradeConfig(
		proposalID,
		proto.NewProtocolDefinition(version, software, upgradeHeight, sdk.NewDecWithPrec(7, 1)),
	)
	k.protocolKeeper.SetUpgradeConfig(ctx, appUpgradeConfig)
	return nil
}

// deprecated
func (k Keeper) getVersionInfoSuccessResult(ctx sdk.Context, version uint64) (proposalID uint64) {
	kvStore := ctx.KVStore(k.storeKey)
	bytes := kvStore.Get(types.GetSuccessVersionKey(version))
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bytes, &proposalID)
	return
}

// deprecated
func (k Keeper) getVersionInfoFailResult(ctx sdk.Context, version uint64, proposalID uint64) (proposalIDRet uint64) {
	kvStore := ctx.KVStore(k.storeKey)
	bytes := kvStore.Get(types.GetFailedVersionKey(version, proposalID))
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bytes, &proposalIDRet)
	return
}
