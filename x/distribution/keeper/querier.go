package keeper

import (
	"fmt"
	sdkerror "github.com/cosmos/cosmos-sdk/types/errors"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/okex/okchain/x/distribution/types"
)

// NewQuerier creates a querier for distribution REST endpoints
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryParams:
			return queryParams(ctx, path[1:], req, k)

		case types.QueryValidatorCommission:
			return queryValidatorCommission(ctx, path[1:], req, k)

		case types.QueryWithdrawAddr:
			return queryDelegatorWithdrawAddress(ctx, path[1:], req, k)

		case types.QueryCommunityPool:
			return queryCommunityPool(ctx, path[1:], req, k)

		default:
			return nil, sdkerror.Wrap(sdkerror.ErrUnknownRequest, "unknown distr query endpoint")
		}
	}
}

func queryParams(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, error) {
	switch path[0] {
	case types.ParamCommunityTax:
		bz, err := codec.MarshalJSONIndent(k.cdc, k.GetCommunityTax(ctx))
		if err != nil {
			return nil, sdkerror.Wrap(sdkerror.ErrInternal, "could not marshal result to JSON" + err.Error())
		}
		return bz, nil
	case types.ParamWithdrawAddrEnabled:
		bz, err := codec.MarshalJSONIndent(k.cdc, k.GetWithdrawAddrEnabled(ctx))
		if err != nil {
			return nil, sdkerror.Wrap(sdkerror.ErrInternal, "could not marshal result to JSON" + err.Error())
		}
		return bz, nil
	default:
		return nil, sdkerror.Wrap(sdkerror.ErrUnknownRequest, fmt.Sprintf("%s is not a valid query request path", req.Path))
	}
}

func queryValidatorCommission(ctx sdk.Context, _ []string, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryValidatorCommissionParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrUnknownRequest, "incorrectly formatted request data" + err.Error())
	}
	commission := k.GetValidatorAccumulatedCommission(ctx, params.ValidatorAddress)
	bz, err := codec.MarshalJSONIndent(k.cdc, commission)
	if err != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrInternal, "could not marshal result to JSON" + err.Error())
	}
	return bz, nil
}

func queryDelegatorWithdrawAddress(ctx sdk.Context, _ []string, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryDelegatorWithdrawAddrParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrUnknownRequest, "incorrectly formatted request data" + err.Error())
	}

	// cache-wrap context as to not persist state changes during querying
	ctx, _ = ctx.CacheContext()
	withdrawAddr := k.GetDelegatorWithdrawAddr(ctx, params.DelegatorAddress)

	bz, err := codec.MarshalJSONIndent(k.cdc, withdrawAddr)
	if err != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrInternal, "could not marshal result to JSON" + err.Error())
	}

	return bz, nil
}

func queryCommunityPool(ctx sdk.Context, _ []string, req abci.RequestQuery, k Keeper) ([]byte, error) {
	bz, err := k.cdc.MarshalJSON(k.GetFeePoolCommunityCoins(ctx))
	if err != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrInternal, "could not marshal result to JSON" + err.Error())
	}
	return bz, nil
}
