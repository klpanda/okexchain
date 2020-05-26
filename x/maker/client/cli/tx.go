package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/okex/okchain/x/maker/types"
	"github.com/spf13/cobra"
)

// GetTxCmd returns the transaction commands for this module
// TODO
func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Maker transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
}
