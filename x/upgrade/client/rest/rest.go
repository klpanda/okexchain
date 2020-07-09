package rest

import (
	"github.com/cosmos/cosmos-sdk/client"
	govRest "github.com/okex/okchain/x/gov/client/rest"
)

func ProposalRESTHandler(client.Context) govRest.ProposalRESTHandler {
	return govRest.ProposalRESTHandler{}
}
