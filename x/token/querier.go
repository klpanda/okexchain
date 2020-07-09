package token

import (
	"fmt"
	sdkerror "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/okex/okchain/x/token/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryInfo:
			return queryInfo(ctx, path[1:], req, keeper)
		case types.QueryTokens:
			return queryTokens(ctx, path[1:], req, keeper)
		case types.QueryParameters:
			return queryParameters(ctx, keeper)
		case types.QueryCurrency:
			return queryCurrency(ctx, path[1:], req, keeper)
		case types.QueryAccount:
			return queryAccount(ctx, path[1:], req, keeper)
		case types.QueryKeysNum:
			return queryKeysNum(ctx, keeper)
		case types.QueryAccountV2:
			return queryAccountV2(ctx, path[1:], req, keeper)
		case types.QueryTokensV2:
			return queryTokensV2(ctx, path[1:], req, keeper)
		case types.QueryTokenV2:
			return queryTokenV2(ctx, path[1:], req, keeper)
		default:
			return nil, sdkerror.Wrap(sdkerror.ErrUnknownRequest, "unknown token query endpoint")
		}
	}
}

// nolint: unparam
func queryInfo(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	name := path[0]

	token := keeper.GetTokenInfo(ctx, name)

	if token.Symbol == "" {
		return nil, sdkerror.Wrap(sdkerror.ErrInvalidCoins, "unknown token")
	}

	bz, err := codec.MarshalJSONIndent(keeper.cdc, token)
	if err != nil {
		panic("could not marshal result to JSON")
	}
	return bz, nil
}

func queryTokens(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	var tokens []types.Token
	if len(path) > 0 && path[0] != "" {
		ownerAddr, err := sdk.AccAddressFromBech32(path[0])
		if err != nil {
			return nil, sdkerror.Wrap(sdkerror.ErrInvalidAddress, fmt.Sprintf("invalid address：%s", path[0]))
		}
		tokens = keeper.GetUserTokensInfo(ctx, ownerAddr)
	} else {
		tokens = keeper.GetTokensInfo(ctx)
	}

	bz, err := codec.MarshalJSONIndent(keeper.cdc, tokens)
	if err != nil {
		panic("could not marshal result to JSON")
	}
	return bz, nil
}

func queryCurrency(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	tokens := keeper.GetCurrenciesInfo(ctx)

	bz, err := codec.MarshalJSONIndent(keeper.cdc, tokens)
	if err != nil {
		panic("could not marshal result to JSON")
	}
	return bz, nil
}

func queryAccount(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	addr, err := sdk.AccAddressFromBech32(path[0])
	if err != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrInvalidAddress, fmt.Sprintf("invalid address：%s", path[0]))
	}

	//var queryPage QueryPage
	var accountParam types.AccountParam
	//var symbol string
	err = keeper.cdc.UnmarshalJSON(req.Data, &accountParam)
	if err != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrUnknownRequest, err.Error())
	}

	coinsInfo := keeper.GetCoinsInfo(ctx, addr)
	coinsInfoChoosen := make([]types.CoinInfo, 0)
	if accountParam.Symbol == "" {
		coinsInfoChoosen = coinsInfo

		// show all or partial
		if accountParam.Show == "all" {
			tokens := keeper.GetTokensInfo(ctx)

			for _, token := range tokens {
				found := false
				for _, coinInfo := range coinsInfo {
					if coinInfo.Symbol == token.Symbol {
						found = true
						break
					}
				}
				// not found
				if !found {
					ci := types.NewCoinInfo(token.Symbol, "0", "0")
					coinsInfoChoosen = append(coinsInfoChoosen, *ci)
				}
			}
		}
		// page and pageSize
		//coinsInfoChoosen = coinsInfoChoosen[Min((accountParam.Page-1)*accountParam.PerPage, len(coinsInfoChoosen)):Min(accountParam.Page*accountParam.PerPage, len(coinsInfoChoosen))]
	} else {
		for _, coinInfo := range coinsInfo {
			if coinInfo.Symbol == accountParam.Symbol {
				coinsInfoChoosen = append(coinsInfoChoosen, coinInfo)
			}
		}
	}
	accountResponse := types.NewAccountResponse(path[0])
	accountResponse.Currencies = coinsInfoChoosen

	bz, err := codec.MarshalJSONIndent(keeper.cdc, accountResponse)
	if err != nil {
		panic("could not marshal result to JSON")
	}
	return bz, nil
}

func queryParameters(ctx sdk.Context, keeper Keeper) ([]byte, error) {
	params := keeper.GetParams(ctx)
	res, err := codec.MarshalJSONIndent(keeper.cdc, params)
	if err != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrInternal, "could not marshal result to JSON" + err.Error())
	}
	return res, nil
}

func queryKeysNum(ctx sdk.Context, keeper Keeper) ([]byte, error) {
	tokenStoreKeyNum, lockStoreKeyNum := keeper.getNumKeys(ctx)
	res, err := codec.MarshalJSONIndent(keeper.cdc,
		map[string]int64{"token": tokenStoreKeyNum,
			"lock": lockStoreKeyNum})
	if err != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrInternal, "could not marshal result to JSON" + err.Error())
	}
	return res, nil
}
