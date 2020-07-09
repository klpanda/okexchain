package keeper

import (
	sdkerror "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/okex/okchain/x/common"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/okex/okchain/x/poolswap/types"
)

// NewQuerier creates a new querier for swap clients.
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QuerySwapTokenPair:
			return querySwapTokenPair(ctx, path[1:], req, k)

		default:
			return nil, sdkerror.Wrap(sdkerror.ErrUnknownRequest, "unknown swap query endpoint")
		}
	}
}

// nolint
func querySwapTokenPair(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte,
	err error) {
	tokenPairName := path[0] + "_" + common.NativeToken
	tokenPair, error := keeper.GetSwapTokenPair(ctx, tokenPairName)
	if error != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrUnknownRequest, error.Error())
	}
	bz := keeper.cdc.MustMarshalJSON(tokenPair)
	return bz, nil
}
