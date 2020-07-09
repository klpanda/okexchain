package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerror "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/okex/okchain/x/upgrade/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// const
const (
	QueryUpgradeConfig        = "config"
	QueryUpgradeVersion       = "version"
	QueryUpgradeFailedVersion = "failed_version"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case QueryUpgradeConfig:
			return queryUpgradeConfig(ctx, req, keeper)
		case QueryUpgradeVersion:
			return queryUpgradeVersion(ctx, req, keeper)
		case QueryUpgradeFailedVersion:
			return queryUpgradeLastFailedVersion(ctx, req, keeper)
		default:
			return nil, sdkerror.Wrap(sdkerror.ErrUnknownRequest, "unknown query endpoint")
		}
	}
}

func queryUpgradeConfig(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	appUpgradeConfig, found := keeper.GetAppUpgradeConfig(ctx)

	if !found {
		return nil, types.NewError(types.DefaultCodespace, types.CodeNoUpgradeConfig, "app upgrade config not found")
	}

	bz, err := codec.MarshalJSONIndent(keeper.cdc, appUpgradeConfig)
	if err != nil {
		panic("could not marshal result to JSON")
	}
	return bz, nil
}

func queryUpgradeVersion(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	version := keeper.protocolKeeper.GetCurrentVersion(ctx)

	bz, err := codec.MarshalJSONIndent(keeper.cdc, types.NewQueryVersion(version))
	if err != nil {
		panic("could not marshal result to JSON")
	}
	return bz, nil
}

func queryUpgradeLastFailedVersion(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	version := keeper.protocolKeeper.GetLastFailedVersion(ctx)

	bz, err := codec.MarshalJSONIndent(keeper.cdc, types.NewQueryVersion(version))
	if err != nil {
		panic("could not marshal result to JSON")
	}
	return bz, nil
}
