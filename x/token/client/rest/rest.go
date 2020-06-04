package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/okex/okchain/x/common"
	govRest "github.com/okex/okchain/x/gov/client/rest"
	"github.com/okex/okchain/x/token/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"
)

// RegisterRoutes, a central function to define routes
// which is called by the rest module in main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, storeName string) {
	r.HandleFunc(fmt.Sprintf("/token/{symbol}"), tokenHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/tokens"), tokensHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/currency/describe"), currencyDescribeHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/accounts/{address}"), spotAccountsHandler(cliCtx, storeName)).Methods("GET")
}

func tokenHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tokenName := vars["symbol"]

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/info/%s", storeName, tokenName), nil)
		if err != nil {
			common.HandleErrorMsg(w, cliCtx, err.Error())
			return
		}
		result := common.GetBaseResponse("hello")
		result2, err2 := json.Marshal(result)
		if err2 != nil {
			common.HandleErrorMsg(w, cliCtx, err2.Error())
			return
		}
		result2 = []byte(strings.Replace(string(result2), "\"hello\"", string(res), 1))
		rest.PostProcessResponse(w, cliCtx, result2)
	}
}

func tokensHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/tokens", storeName), nil)
		if err != nil {
			common.HandleErrorMsg(w, cliCtx, err.Error())
			return
		}

		result := common.GetBaseResponse("hello")
		result2, err2 := json.Marshal(result)
		if err2 != nil {
			common.HandleErrorMsg(w, cliCtx, err2.Error())
			return
		}
		result2 = []byte(strings.Replace(string(result2), "\"hello\"", string(res), 1))
		rest.PostProcessResponse(w, cliCtx, result2)
	}
}

func currencyDescribeHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/currency/describe", storeName), nil)
		if err != nil {
			common.HandleErrorMsg(w, cliCtx, err.Error())
			return
		}

		result := common.GetBaseResponse("hello")
		result2, err2 := json.Marshal(result)
		if err2 != nil {
			common.HandleErrorMsg(w, cliCtx, err2.Error())
			return
		}
		result2 = []byte(strings.Replace(string(result2), "\"hello\"", string(res), 1))
		rest.PostProcessResponse(w, cliCtx, result2)
	}
}

func spotAccountsHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		address := vars["address"]

		symbol := r.URL.Query().Get("symbol")
		show := r.URL.Query().Get("show")

		if show == "" {
			show = "partial"
		}
		if show != "partial" && show != "all" {
			result := common.GetErrorResponseJSON(1, "", "param show not valid")
			rest.PostProcessResponse(w, cliCtx, result)
			return
		}

		accountParam := types.AccountParam{
			Symbol: symbol,
			Show:   show,
			//QueryPage: token.QueryPage{
			//	Page:    pageInt,
			//	PerPage: perPageInt,
			//},
		}

		bz, err := cliCtx.Codec.MarshalJSON(accountParam)
		if err != nil {
			common.HandleErrorMsg(w, cliCtx, err.Error())
			return
		}
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/accounts/%s", storeName, address), bz)
		if err != nil {
			common.HandleErrorMsg(w, cliCtx, err.Error())
			return
		}

		result := common.GetBaseResponse("hello")
		result2, err2 := json.Marshal(result)
		if err2 != nil {
			common.HandleErrorMsg(w, cliCtx, err2.Error())
			return
		}
		result2 = []byte(strings.Replace(string(result2), "\"hello\"", string(res), 1))
		rest.PostProcessResponse(w, cliCtx, result2)
	}
}

// ProposalRESTHandler defines token proposal handler
func ProposalRESTHandler(context.CLIContext) govRest.ProposalRESTHandler {
	return govRest.ProposalRESTHandler{}
}
