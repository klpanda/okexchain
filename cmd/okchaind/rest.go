package main

import (
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/server/api"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankrest "github.com/cosmos/cosmos-sdk/x/bank/client/rest"
	backendrest "github.com/okex/okchain/x/backend/client/rest"
	dexrest "github.com/okex/okchain/x/dex/client/rest"
	dist "github.com/okex/okchain/x/distribution"
	distrest "github.com/okex/okchain/x/distribution/client/rest"
	orderrest "github.com/okex/okchain/x/order/client/rest"
	stakingrest "github.com/okex/okchain/x/staking/client/rest"
	"github.com/okex/okchain/x/token"
	tokensrest "github.com/okex/okchain/x/token/client/rest"
)

// registerRoutes registers the routes from the different modules for the LCD.
// NOTE: details on the routes added for each module are in the module documentation
// NOTE: If making updates here you also need to update the test helper in client/lcd/test_helper.go
func registerRoutes(rs *api.Server) {
	registerRoutesV1(rs)
	registerRoutesV2(rs)
}

func registerRoutesV1(rs *api.Server) {
	v1Router := rs.Router.PathPrefix("/okchain/v1").Name("v1").Subrouter()
	rpc.RegisterRoutes(rs.ClientCtx, v1Router)
	authrest.RegisterRoutes(rs.ClientCtx, v1Router, authtypes.StoreKey)
	bankrest.RegisterRoutes(rs.ClientCtx, v1Router)
	stakingrest.RegisterRoutes(rs.ClientCtx, v1Router)
	distrest.RegisterRoutes(rs.ClientCtx, v1Router, dist.StoreKey)

	orderrest.RegisterRoutes(rs.ClientCtx, v1Router)
	tokensrest.RegisterRoutes(rs.ClientCtx, v1Router, token.ModuleName)
	backendrest.RegisterRoutes(rs.ClientCtx, v1Router)
	dexrest.RegisterRoutes(rs.ClientCtx, v1Router)
}

func registerRoutesV2(rs *api.Server) {
	v2Router := rs.Router.PathPrefix("/okchain/v2").Name("v1").Subrouter()
	rpc.RegisterRoutes(rs.ClientCtx, v2Router)
	authrest.RegisterRoutes(rs.ClientCtx, v2Router, authtypes.StoreKey)
	bankrest.RegisterRoutes(rs.ClientCtx, v2Router)
	stakingrest.RegisterRoutes(rs.ClientCtx, v2Router)
	distrest.RegisterRoutes(rs.ClientCtx, v2Router, dist.StoreKey)

	orderrest.RegisterRoutesV2(rs.ClientCtx, v2Router)
	tokensrest.RegisterRoutesV2(rs.ClientCtx, v2Router, token.ModuleName)
	backendrest.RegisterRoutesV2(rs.ClientCtx, v2Router)
}
