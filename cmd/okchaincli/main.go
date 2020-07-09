package main

import (
	"fmt"
	"os"
	"path"

	"github.com/okex/okchain/app"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/okex/okchain/app/protocol"
	debugcli "github.com/okex/okchain/x/debug/client/cli"
	tokencli "github.com/okex/okchain/x/token/client/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"
)

func main() {
	// configure cobra to sort commands
	cobra.EnableCommandSorting = false

	// instantiate the codec for the command line application
	cdcCfg := app.MakeEncodingConfig()

	// read in the configuration file for the sdk
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(sdk.Bech32PrefixAccAddr, sdk.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(sdk.Bech32PrefixValAddr, sdk.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(sdk.Bech32PrefixConsAddr, sdk.Bech32PrefixConsPub)
	config.Seal()

	// TODO: setup keybase, viper object, etc. to be passed into
	// the below functions and eliminate global vars, like we do with the cdc

	rootCmd := &cobra.Command{
		Use:   "okchaincli",
		Short: "Command line interface for interacting with okchaind",
	}

	// add --chain-id to persistent flags and mark it required
	rootCmd.PersistentFlags().String(flags.FlagChainID, "", "Chain ID of tendermint node")
	rootCmd.PersistentFlags().String(keys.FlagKeyPass, keys.DefaultKeyPass, "Pass word of sender")

	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		return initConfig(rootCmd)
	}

	// construct root command
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		queryCmd(cdcCfg),
		txCmd(cdcCfg),
		debugcli.GetDebugCmd(cdcCfg),
		flags.LineBreak,
		flags.LineBreak,
		keys.Commands(),
		flags.LineBreak,
		version.Cmd,
		flags.NewCompletionCmd(rootCmd, true),
	)

	// add flags and prefix all env exposed with OKCHAIN
	executor := cli.PrepareMainCmd(rootCmd, "OKCHAIN", app.DefaultCLIHome)

	if err := executor.Execute(); err != nil {
		fmt.Printf("Failed executing CLI command: %s, exiting...\n", err)
		os.Exit(1)
	}
}

func queryCmd(config protocol.EncodingConfig) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:     "query",
		Aliases: []string{"q"},
		Short:   "Querying subcommands",
	}

	cdc := config.Amino
	queryCmd.AddCommand(
		authcmd.GetAccountCmd(cdc),
		flags.LineBreak,
		rpc.ValidatorCommand(cdc),
		rpc.BlockCommand(),
		authcmd.QueryTxsByEventsCmd(cdc),
		authcmd.QueryTxCmd(cdc),
		flags.LineBreak,
	)

	// add modules' query commands
	clientCtx := client.Context{}
	clientCtx = clientCtx.
		WithJSONMarshaler(config.Marshaler).
		WithCodec(cdc)

	// add modules' query commands
	app.ModuleBasics.AddQueryCommands(queryCmd, clientCtx)

	return queryCmd
}

func txCmd(config protocol.EncodingConfig) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "tx",
		Short: "Transactions subcommands",
	}

	cdc := config.Amino
	clientCtx := client.Context{}
	clientCtx = clientCtx.
		WithJSONMarshaler(config.Marshaler).
		WithTxGenerator(config.TxGenerator).
		WithAccountRetriever(types.NewAccountRetriever(config.Marshaler)).
		WithCodec(cdc)

	txCmd.AddCommand(
		tokencli.SendTxCmd(cdc),
		flags.LineBreak,
		authcmd.GetSignCommand(clientCtx),
		authcmd.GetMultiSignCommand(clientCtx),
		flags.LineBreak,
		authcmd.GetBroadcastCommand(clientCtx),
		authcmd.GetEncodeCommand(clientCtx),
		flags.LineBreak,
	)

	// add modules' tx commands
	app.ModuleBasics.AddTxCommands(txCmd, clientCtx)

	// remove auth and bank commands as they're mounted under the root tx command
	var cmdsToRemove []*cobra.Command

	for _, cmd := range txCmd.Commands() {
		if cmd.Use == authtypes.ModuleName || cmd.Use == banktypes.ModuleName {
			cmdsToRemove = append(cmdsToRemove, cmd)
		}
	}

	txCmd.RemoveCommand(cmdsToRemove...)

	return txCmd
}

func initConfig(cmd *cobra.Command) error {
	home, err := cmd.PersistentFlags().GetString(cli.HomeFlag)
	if err != nil {
		return err
	}

	cfgFile := path.Join(home, "config", "config.toml")
	if _, err := os.Stat(cfgFile); err == nil {
		viper.SetConfigFile(cfgFile)

		if err := viper.ReadInConfig(); err != nil {
			return err
		}
	}
	if err := viper.BindPFlag(flags.FlagChainID, cmd.PersistentFlags().Lookup(flags.FlagChainID)); err != nil {
		return err
	}
	if err := viper.BindPFlag(cli.EncodingFlag, cmd.PersistentFlags().Lookup(cli.EncodingFlag)); err != nil {
		return err
	}
	return viper.BindPFlag(cli.OutputFlag, cmd.PersistentFlags().Lookup(cli.OutputFlag))
}
