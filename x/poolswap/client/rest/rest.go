package rest

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/gorilla/mux"
)

// RegisterRoutes registers poolswap-related REST handlers to a router
func RegisterRoutes(cliCtx client.Context, r *mux.Router) {
	registerTxRoutes(cliCtx, r)
}
