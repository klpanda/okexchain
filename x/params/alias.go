package params

import (
	sdkparamstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	sdkparamsproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/okex/okchain/x/params/types"
)

// const
const (
	ModuleName        = sdkparamstypes.ModuleName
	DefaultParamspace = sdkparamstypes.ModuleName
	StoreKey          = sdkparamstypes.StoreKey
	TStoreKey         = sdkparamstypes.TStoreKey
	RouterKey         = sdkparamsproposal.RouterKey
)

type (
	// KeyTable is the type alias of the one in cmsdk
	KeyTable = sdkparamstypes.KeyTable
	// ParamSetPairs is the type alias of the one in cmsdk
	ParamSetPairs = sdkparamstypes.ParamSetPairs
	// Subspace is the type alias of the one in cmsdk
	Subspace = sdkparamstypes.Subspace
	// ParamSet is the type alias of the one in cmsdk
	ParamSet = sdkparamstypes.ParamSet
	// ParamChange is the type alias of the one in cmsdk
	ParamChange = sdkparamsproposal.ParamChange
	// ParameterChangeProposal is alias of ParameterChangeProposal in types
	ParameterChangeProposal = types.ParameterChangeProposal
	ParamSetPair           = sdkparamstypes.ParamSetPair
)

var (
	// nolint
	NewKeyTable    = sdkparamstypes.NewKeyTable
	NewParamChange = sdkparamsproposal.NewParamChange
	DefaultParams  = types.DefaultParams
)
