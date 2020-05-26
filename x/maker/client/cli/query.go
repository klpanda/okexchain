package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/okex/okchain/x/maker/types"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for maker module
// TODO
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the maker module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
}
