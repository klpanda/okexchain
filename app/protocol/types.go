package protocol

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/baseapp"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/okex/okchain/x/backend"
	distr "github.com/okex/okchain/x/distribution"
	"github.com/okex/okchain/x/staking"
	"github.com/okex/okchain/x/stream"
	"github.com/okex/okchain/x/token"
)

// Protocol shows the expected behavior for any protocol version
type Protocol interface {
	GetVersion() uint64

	// load base installation for each protocol
	LoadContext()
	Init()
	GetCodec() *codec.Codec

	// gracefully stop okchaind
	CheckStopped()

	// setter
	SetLogger(log log.Logger) Protocol
	SetParent(parent Parent) Protocol

	//getter
	GetParent() Parent

	// get specific keeper
	GetBackendKeeper() backend.Keeper
	GetStreamKeeper() stream.Keeper
	GetCrisisKeeper() crisiskeeper.Keeper
	GetStakingKeeper() staking.Keeper
	GetDistrKeeper() distr.Keeper
	GetSlashingKeeper() slashingkeeper.Keeper
	GetTokenKeeper() token.Keeper

	// fit cm36
	GetKVStoreKeysMap() map[string]*sdk.KVStoreKey
	GetTransientStoreKeysMap() map[string]*sdk.TransientStoreKey
	ExportGenesis(ctx sdk.Context) map[string]json.RawMessage
}

// Parent shows the expected behavior of BaseApp(hooks)
type Parent interface {
	DeliverTx(abci.RequestDeliverTx) abci.ResponseDeliverTx
	PushInitChainer(initChainer sdk.InitChainer)
	PushBeginBlocker(beginBlocker sdk.BeginBlocker)
	PushEndBlocker(endBlocker sdk.EndBlocker)
	PushAnteHandler(ah sdk.AnteHandler)
	SetRouter(router sdk.Router, queryRouter sdk.QueryRouter)
	SetParamStore(ps baseapp.ParamStore)
}
