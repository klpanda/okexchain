package token

import (
	"fmt"
	sdkerror "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/okex/okchain/x/common"
	"github.com/okex/okchain/x/token/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func queryAccountV2(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	addr, err := sdk.AccAddressFromBech32(path[0])
	if err != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrInvalidAddress, fmt.Sprintf("invalid address：%s", path[0]))
	}

	//var queryPage QueryPage
	var accountParam types.AccountParamV2
	//var symbol string
	err = types.ModuleCdc.UnmarshalJSON(req.Data, &accountParam)
	if err != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrUnknownRequest, err.Error())
	}

	coinsInfo := keeper.GetCoinsInfo(ctx, addr)
	coinsInfoChosen := make([]CoinInfo, 0)
	if accountParam.Currency == "" {
		coinsInfoChosen = coinsInfo

		// hide_zero yes or no
		if accountParam.HideZero == "no" {
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
					coinsInfoChosen = append(coinsInfoChosen, *ci)
				}
			}
		}
	} else {
		for _, coinInfo := range coinsInfo {
			if coinInfo.Symbol == accountParam.Currency {
				coinsInfoChosen = append(coinsInfoChosen, coinInfo)
			}
		}
	}

	res, err := common.JSONMarshalV2(coinsInfoChosen)
	if err != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrInternal, err.Error())
	}
	return res, nil
}

func queryTokensV2(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	tokens := keeper.GetTokensInfo(ctx)

	res, err := common.JSONMarshalV2(tokens)
	if err != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrInternal, err.Error())
	}
	return res, nil
}

func queryTokenV2(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	name := path[0]

	token := keeper.GetTokenInfo(ctx, name)

	if token.Symbol == "" {
		return nil, sdkerror.Wrap(sdkerror.ErrInvalidCoins, "unknown token")
	}

	res, err := common.JSONMarshalV2(token)
	if err != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrInternal, err.Error())
	}
	return res, nil
}
