package rest

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"net/http"

	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"
	"github.com/okex/okchain/x/common"
	"github.com/okex/okchain/x/poolswap/types"
)

func registerTxRoutes(cliCtx client.Context, r *mux.Router) {
	r.HandleFunc("/poolswap/exchange", swapExchangeHandler(cliCtx)).Methods("GET")
}

func swapExchangeHandler(cliCtx client.Context) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tokenName := vars["token"]

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/poolswap/swapTokenPair/%s", tokenName), nil)
		if err != nil {
			common.HandleErrorMsg(w, cliCtx, err.Error())
			return
		}

		exchange := types.SwapTokenPair{}
		cliCtx.Codec.MustUnmarshalJSON(res, exchange)
		response := common.GetBaseResponse(exchange)
		resBytes, err := json.Marshal(response)
		if err != nil {
			common.HandleErrorMsg(w, cliCtx, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, resBytes)
	}
}
