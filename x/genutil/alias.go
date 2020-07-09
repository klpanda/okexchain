package genutil

import (
	"github.com/okex/okchain/x/genutil/types"

	sdkgenutil "github.com/cosmos/cosmos-sdk/x/genutil"
	sdkgenutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

// const
const (
	ModuleName = types.ModuleName
)

type (
	// GenesisState is the type alias of the one in cmsdk
	GenesisState = sdkgenutiltypes.GenesisState
	// InitConfig is the type alias of the one in cmsdk
	InitConfig = sdkgenutiltypes.InitConfig
	// GenesisAccountsIterator is the type alias of the one in cmsdk
	GenesisAccountsIterator = sdkgenutiltypes.GenesisAccountsIterator
	GenesisBalancesIterator = sdkgenutiltypes.GenesisBalancesIterator
)

var (
	// nolint
	ModuleCdc                    = types.ModuleCdc
	GenesisStateFromGenFile      = sdkgenutiltypes.GenesisStateFromGenFile
	NewGenesisState              = sdkgenutiltypes.NewGenesisState
	SetGenesisStateInAppState    = sdkgenutiltypes.SetGenesisStateInAppState
	InitializeNodeValidatorFiles = sdkgenutil.InitializeNodeValidatorFiles
	ExportGenesisFileWithTime    = sdkgenutil.ExportGenesisFileWithTime
	NewInitConfig                = sdkgenutiltypes.NewInitConfig
	ValidateGenesis              = sdkgenutiltypes.ValidateGenesis
	GenesisStateFromGenDoc       = sdkgenutiltypes.GenesisStateFromGenDoc
	SetGenTxsInAppGenesisState   = sdkgenutil.SetGenTxsInAppGenesisState
	ExportGenesisFile            = sdkgenutil.ExportGenesisFile
)
