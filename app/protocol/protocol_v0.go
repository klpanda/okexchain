package protocol

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	govtypes "github.com/okex/okchain/x/gov/types"
	ordertypes "github.com/okex/okchain/x/order/types"
	"os"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	ibckeeper "github.com/cosmos/cosmos-sdk/x/ibc/keeper"
	"github.com/cosmos/cosmos-sdk/x/mint"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	"github.com/okex/okchain/app/utils"
	"github.com/okex/okchain/x/backend"
	"github.com/okex/okchain/x/common/proto"
	"github.com/okex/okchain/x/common/version"
	"github.com/okex/okchain/x/debug"
	"github.com/okex/okchain/x/dex"
	dexClient "github.com/okex/okchain/x/dex/client"
	distr "github.com/okex/okchain/x/distribution"
	"github.com/okex/okchain/x/genutil"
	"github.com/okex/okchain/x/gov"
	"github.com/okex/okchain/x/gov/keeper"
	"github.com/okex/okchain/x/order"
	"github.com/okex/okchain/x/params"
	paramsclient "github.com/okex/okchain/x/params/client"
	"github.com/okex/okchain/x/poolswap"
	"github.com/okex/okchain/x/staking"
	"github.com/okex/okchain/x/stream"
	"github.com/okex/okchain/x/token"
	"github.com/okex/okchain/x/upgrade"
	upgradeClient "github.com/okex/okchain/x/upgrade/client"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
)

var (
	// check the implements of ProtocolV0
	_ Protocol = (*ProtocolV0)(nil)

	// DefaultCLIHome is the directory for okchaincli
	DefaultCLIHome = os.ExpandEnv("$HOME/.okchaincli")
	// DefaultNodeHome is the directory for okchaind
	DefaultNodeHome = os.ExpandEnv("$HOME/.okchaind")

	// ModuleBasics is in charge of setting up basic, non-dependant module elements,
	// such as codec registration and genesis verification
	ModuleBasics = module.NewBasicManager(
		genutil.AppModuleBasic{},
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(
			upgradeClient.ProposalHandler, paramsclient.ProposalHandler,
			dexClient.DelistProposalHandler, distr.ProposalHandler,
		),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},

		// okchain extended
		token.AppModuleBasic{},
		dex.AppModuleBasic{},
		order.AppModuleBasic{},
		backend.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		stream.AppModuleBasic{},
		debug.AppModuleBasic{},
		poolswap.AppModuleBasic{},
	)

	// module account permissions for bankKeeper and supplyKeeper
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName: nil,
		distr.ModuleName:           nil,
		minttypes.ModuleName:       {authtypes.Minter},
		staking.BondedPoolName:     {authtypes.Burner, authtypes.Staking},
		staking.NotBondedPoolName:  {authtypes.Burner, authtypes.Staking},
		gov.ModuleName:             nil,
		token.ModuleName:           {authtypes.Minter, authtypes.Burner},
		order.ModuleName:           nil,
		backend.ModuleName:         nil,
		dex.ModuleName:             nil,
		poolswap.ModuleName:        {authtypes.Minter, authtypes.Burner},
	}
)

// ProtocolV0 is the struct of the original protocol of okchain
type ProtocolV0 struct {
	parent         Parent
	version        uint64
	cdc            *codec.Codec
	appCodec       codec.Marshaler
	logger         log.Logger
	invCheckPeriod uint

	// keys to access the substores
	keys  map[string]*sdk.KVStoreKey
	tkeys map[string]*sdk.TransientStoreKey

	// keepers
	accountKeeper  authkeeper.AccountKeeper
	bankKeeper     bankkeeper.BaseKeeper
	stakingKeeper  staking.Keeper
	slashingKeeper slashingkeeper.Keeper
	mintKeeper     mintkeeper.Keeper
	distrKeeper    distr.Keeper
	govKeeper      gov.Keeper
	crisisKeeper   crisiskeeper.Keeper
	paramsKeeper   params.Keeper
	tokenKeeper    token.Keeper
	dexKeeper      dex.Keeper
	orderKeeper    order.Keeper
	swapKeeper     poolswap.Keeper
	protocolKeeper proto.ProtocolKeeper
	backendKeeper  backend.Keeper
	streamKeeper   stream.Keeper
	upgradeKeeper  upgrade.Keeper
	debugKeeper    debug.Keeper

	stopped     bool
	anteHandler sdk.AnteHandler // ante handler for fee and auth
	router      sdk.Router      // handle any kind of message
	queryRouter sdk.QueryRouter // router for redirecting query calls

	// the module manager
	mm *module.Manager
}

type EncodingConfig struct {
	InterfaceRegistry types.InterfaceRegistry
	Marshaler         codec.Marshaler
	TxGenerator       client.TxGenerator
	Amino             *codec.Codec
}

// NewProtocolV0 creates a new instance of NewProtocolV0
func NewProtocolV0(
	parent Parent, version uint64, log log.Logger, invCheckPeriod uint, pk proto.ProtocolKeeper,
) *ProtocolV0 {
	return &ProtocolV0{
		parent:         parent,
		version:        version,
		logger:         log,
		invCheckPeriod: invCheckPeriod,
		protocolKeeper: pk,
		keys:           kvStoreKeysMap,
		tkeys:          transientStoreKeysMap,
		router:         baseapp.NewRouter(),
		queryRouter:    baseapp.NewQueryRouter(),
	}
}

// GetVersion gets the version of this protocol
func (p *ProtocolV0) GetVersion() uint64 {
	return p.version
}

// LoadContext updates the context for the app after the upgrade of protocol
func (p *ProtocolV0) LoadContext() {
	p.logger.Debug("Protocol V0: LoadContext")
	p.setCodec()
	p.produceKeepers()
	p.setManager()
	p.registerRouters()
	p.setAnteHandler()

	p.parent.PushInitChainer(p.InitChainer)
	p.parent.PushBeginBlocker(p.BeginBlocker)
	p.parent.PushEndBlocker(p.EndBlocker)
}

// GetCodec gets tx codec
func (p *ProtocolV0) GetCodec() *codec.Codec {
	if p.cdc == nil {
		panic("Invalid cdc from ProtocolV0")
	}
	return p.cdc
}

// CheckStopped gives a quick check whether okchain needs stopped
func (p *ProtocolV0) CheckStopped() {
	if p.stopped {
		p.logger.Info("OKChain is going to exit")
		server.Stop()
		p.logger.Info("OKChain was stopped")
		select {}
	}
}

// GetBackendKeeper gets backend keeper
func (p *ProtocolV0) GetBackendKeeper() backend.Keeper {
	return p.backendKeeper
}

// GetStreamKeeper gets stream keeper
func (p *ProtocolV0) GetStreamKeeper() stream.Keeper {
	return p.streamKeeper
}

// GetCrisisKeeper gets crisis keeper
func (p *ProtocolV0) GetCrisisKeeper() crisiskeeper.Keeper {
	return p.crisisKeeper
}

// GetStakingKeeper gets staking keeper
func (p *ProtocolV0) GetStakingKeeper() staking.Keeper {
	return p.stakingKeeper
}

// GetDistrKeeper gets distr keeper
func (p *ProtocolV0) GetDistrKeeper() distr.Keeper {
	return p.distrKeeper
}

// GetSlashingKeeper gets slashing keeper
func (p *ProtocolV0) GetSlashingKeeper() slashingkeeper.Keeper {
	return p.slashingKeeper
}

// GetTokenKeeper gets token keeper
func (p *ProtocolV0) GetTokenKeeper() token.Keeper {
	return p.tokenKeeper
}

// GetKVStoreKeysMap gets the map of kv store keys
func (p *ProtocolV0) GetKVStoreKeysMap() map[string]*sdk.KVStoreKey {
	return p.keys
}

// GetTransientStoreKeysMap gets the map of transient store keys
func (p *ProtocolV0) GetTransientStoreKeysMap() map[string]*sdk.TransientStoreKey {
	return p.tkeys
}

// nolint
func (p *ProtocolV0) Init() {}

func (p *ProtocolV0) setCodec() {
	cfg := MakeEncodingConfig()
	p.cdc = cfg.Amino
	p.appCodec = cfg.Marshaler
}

// produceKeepers initializes all keepers declared in the ProtocolV0 struct
func (p *ProtocolV0) produceKeepers() {
	// get config
	appConfig, err := config.ParseConfig()
	if err != nil {
		p.logger.Error(fmt.Sprintf("the config of OKChain was parsed error : %s", err.Error()))
		panic(err)
	}

	// 1.init params keeper and subspaces
	p.paramsKeeper = params.NewKeeper(
		p.appCodec, p.keys[params.StoreKey], p.tkeys[params.TStoreKey],
	)
	authSubspace := p.paramsKeeper.Subspace(authtypes.ModuleName)
	bankSubspace := p.paramsKeeper.Subspace(banktypes.ModuleName)
	stakingSubspace := p.paramsKeeper.Subspace(staking.DefaultParamspace)
	mintSubspace := p.paramsKeeper.Subspace(minttypes.ModuleName)
	distrSubspace := p.paramsKeeper.Subspace(distr.DefaultParamspace)
	slashingSubspace := p.paramsKeeper.Subspace(slashingtypes.DefaultParamspace)
	govSubspace := p.paramsKeeper.Subspace(gov.DefaultParamspace)
	crisisSubspace := p.paramsKeeper.Subspace(crisistypes.ModuleName)
	tokenSubspace := p.paramsKeeper.Subspace(token.DefaultParamspace)
	orderSubspace := p.paramsKeeper.Subspace(order.DefaultParamspace)
	upgradeSubspace := p.paramsKeeper.Subspace(upgrade.DefaultParamspace)
	dexSubspace := p.paramsKeeper.Subspace(dex.DefaultParamspace)
	swapSubSpace := p.paramsKeeper.Subspace(poolswap.DefaultParamspace)

	// 2.add keepers
	p.accountKeeper = authkeeper.NewAccountKeeper(
		p.appCodec, p.keys[authtypes.StoreKey], authSubspace, authtypes.ProtoBaseAccount, maccPerms)
	p.bankKeeper = bankkeeper.NewBaseKeeper(
		p.appCodec, p.keys[banktypes.StoreKey], p.accountKeeper, bankSubspace, p.moduleAccountAddrs())
	p.paramsKeeper.SetBankKeeper(p.bankKeeper)
	p.parent.SetParamStore(p.paramsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(std.ConsensusParamsKeyTable()))
	stakingKeeper := staking.NewKeeper(p.appCodec, p.keys[staking.StoreKey], p.tkeys[staking.TStoreKey],
		p.accountKeeper, p.bankKeeper, stakingSubspace, staking.DefaultCodespace)

	p.paramsKeeper.SetStakingKeeper(stakingKeeper)
	p.mintKeeper = mintkeeper.NewKeeper(
		p.appCodec, p.keys[minttypes.StoreKey], mintSubspace, stakingKeeper, p.accountKeeper,
		p.bankKeeper, authtypes.FeeCollectorName,
	)

	p.distrKeeper = distr.NewKeeper(p.cdc, p.keys[distr.StoreKey],
		distrSubspace, &stakingKeeper, p.accountKeeper, p.bankKeeper,
		distr.DefaultCodespace, authtypes.FeeCollectorName, p.moduleAccountAddrs(),
	)

	p.slashingKeeper = slashingkeeper.NewKeeper(
		p.appCodec, p.keys[slashingtypes.StoreKey], stakingKeeper, slashingSubspace,
	)

	p.crisisKeeper = crisiskeeper.NewKeeper(crisisSubspace, p.invCheckPeriod, p.bankKeeper, authtypes.FeeCollectorName)

	p.tokenKeeper = token.NewKeeper(
		p.bankKeeper, tokenSubspace, authtypes.FeeCollectorName, p.accountKeeper,
		p.keys[token.StoreKey], p.keys[token.KeyLock],
		p.cdc, appConfig.BackendConfig.EnableBackend)

	p.dexKeeper = dex.NewKeeper(authtypes.FeeCollectorName, p.accountKeeper, dexSubspace, p.tokenKeeper, &stakingKeeper,
		p.bankKeeper, p.keys[dex.StoreKey], p.keys[dex.TokenPairStoreKey], p.cdc)

	p.orderKeeper = order.NewKeeper(
		p.tokenKeeper, p.accountKeeper, p.bankKeeper, p.dexKeeper, orderSubspace, authtypes.FeeCollectorName,
		p.keys[order.OrderStoreKey], p.cdc, appConfig.BackendConfig.EnableBackend, orderMetrics,
	)

	p.swapKeeper = poolswap.NewKeeper(p.bankKeeper, p.tokenKeeper, p.cdc, p.keys[poolswap.StoreKey], swapSubSpace)

	p.streamKeeper = stream.NewKeeper(p.orderKeeper, p.tokenKeeper, &p.dexKeeper, &p.accountKeeper,
		p.cdc, p.logger, appConfig, streamMetrics)

	p.backendKeeper = backend.NewKeeper(p.orderKeeper, p.tokenKeeper, &p.dexKeeper, p.streamKeeper.GetMarketKeeper(),
		p.cdc, p.logger, appConfig.BackendConfig)

	// 3.register the proposal types
	govRouter := gov.NewRouter()
	govRouter.AddRoute(gov.RouterKey, gov.ProposalHandler).
		AddRoute(params.RouterKey, params.NewParamChangeProposalHandler(&p.paramsKeeper)).
		AddRoute(dex.RouterKey, dex.NewProposalHandler(&p.dexKeeper)).
		AddRoute(upgrade.RouterKey, upgrade.NewAppUpgradeProposalHandler(&p.upgradeKeeper)).
		AddRoute(distr.RouterKey, distr.NewCommunityPoolSpendProposalHandler(p.distrKeeper))
	govProposalHandlerRouter := keeper.NewProposalHandlerRouter()
	govProposalHandlerRouter.AddRoute(params.RouterKey, &p.paramsKeeper).
		AddRoute(dex.RouterKey, &p.dexKeeper).
		AddRoute(upgrade.RouterKey, &p.upgradeKeeper)
	p.govKeeper = gov.NewKeeper(
		p.appCodec, p.keys[gov.StoreKey], govSubspace.WithKeyTable(govtypes.ParamKeyTable()),
		p.accountKeeper, p.bankKeeper, stakingKeeper, govRouter,
		govProposalHandlerRouter, authtypes.FeeCollectorName,
	)
	p.paramsKeeper.SetGovKeeper(p.govKeeper)
	p.dexKeeper.SetGovKeeper(p.govKeeper)
	// 4.register the staking hooks
	p.stakingKeeper = *stakingKeeper.SetHooks(
		staking.NewMultiStakingHooks(p.distrKeeper.Hooks(), p.slashingKeeper.Hooks()),
	)
	p.upgradeKeeper = upgrade.NewKeeper(
		p.cdc, p.keys[upgrade.StoreKey], p.protocolKeeper, p.stakingKeeper, p.bankKeeper, upgradeSubspace,
	)
	p.debugKeeper = debug.NewDebugKeeper(p.cdc, p.keys[debug.StoreKey], p.orderKeeper, p.stakingKeeper, authtypes.FeeCollectorName, p.Stop)
}

// moduleAccountAddrs returns all the module account addresses
func (p *ProtocolV0) moduleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[p.accountKeeper.GetModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// setManager sets module.Manager in protocolV0
func (p *ProtocolV0) setManager() {

	p.mm = module.NewManager(
		genutil.NewAppModule(p.accountKeeper, p.stakingKeeper, p.parent.DeliverTx),
		auth.NewAppModule(p.appCodec, p.accountKeeper),
		bank.NewAppModule(p.appCodec, p.bankKeeper, p.accountKeeper),
		crisis.NewAppModule(&p.crisisKeeper),
		params.NewAppModule(p.paramsKeeper),
		mint.NewAppModule(p.appCodec, p.mintKeeper, p.accountKeeper),
		slashing.NewAppModule(p.appCodec, p.slashingKeeper, p.accountKeeper, p.bankKeeper, p.stakingKeeper),
		staking.NewAppModule(p.stakingKeeper, p.accountKeeper, p.bankKeeper),
		distr.NewAppModule(p.distrKeeper, p.bankKeeper),
		gov.NewAppModule(version.ProtocolVersionV0, p.govKeeper, p.accountKeeper, p.bankKeeper),
		order.NewAppModule(version.ProtocolVersionV0, p.orderKeeper),
		token.NewAppModule(version.ProtocolVersionV0, p.tokenKeeper),
		poolswap.NewAppModule(p.swapKeeper),

		// TODO
		dex.NewAppModule(version.ProtocolVersionV0, p.dexKeeper, p.accountKeeper, p.bankKeeper),
		backend.NewAppModule(p.backendKeeper),
		stream.NewAppModule(p.streamKeeper),
		upgrade.NewAppModule(p.upgradeKeeper),

		debug.NewAppModule(p.debugKeeper),
	)

	// ORDER SETTING
	p.mm.SetOrderBeginBlockers(
		stream.ModuleName,
		order.ModuleName,
		token.ModuleName,
		dex.ModuleName,
		minttypes.ModuleName,
		distr.ModuleName,
		slashingtypes.ModuleName,
		staking.ModuleName,
	)

	p.mm.SetOrderEndBlockers(
		crisistypes.ModuleName,
		gov.ModuleName,
		dex.ModuleName,
		order.ModuleName,
		staking.ModuleName,
		backend.ModuleName,
		stream.ModuleName,
		upgrade.ModuleName,
	)

	p.mm.SetOrderInitGenesis(
		distr.ModuleName,
		staking.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		slashingtypes.ModuleName,
		gov.ModuleName,
		minttypes.ModuleName,
		token.ModuleName,
		dex.ModuleName,
		order.ModuleName,
		poolswap.ModuleName,
		upgrade.ModuleName,
		crisistypes.ModuleName,
		genutil.ModuleName,
		params.ModuleName,
	)
}

// registerRouters registers Routers by Manager
func (p *ProtocolV0) registerRouters() {
	p.mm.RegisterInvariants(&p.crisisKeeper)
	p.mm.RegisterRoutes(p.router, p.queryRouter)
	p.parent.SetRouter(p.router, p.queryRouter)
}

// setAnteHandler sets ante handler
func (p *ProtocolV0) setAnteHandler() {
	p.anteHandler = ante.NewAnteHandler(
		p.accountKeeper,
		p.bankKeeper,
		ibckeeper.Keeper{},
		ante.DefaultSigVerificationGasConsumer,
		authtypes.LegacyAminoJSONHandler{},
	)
	ante.IsSysFreeHandler = isSystemFreeHook
	ante.ValMsgHandler = validateMsgHook(p.orderKeeper)
	p.parent.PushAnteHandler(p.anteHandler)
}

// InitChainer initializes application state at genesis as a hook
func (p *ProtocolV0) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState simapp.GenesisState
	p.cdc.MustUnmarshalJSON(req.AppStateBytes, &genesisState)

	//var accGenesisState authtypes.GenesisState
	//p.cdc.MustUnmarshalJSON(genesisState[authtypes.ModuleName], &accGenesisState)
	//
	//var acc authtypes.AccountI
	//if len(accGenesisState.Accounts) > 0 {
	//	acc = accGenesisState.Accounts[0]
	//}
	//
	//if err := token.IssueOKT(ctx, p.tokenKeeper, genesisState[token.ModuleName], acc); err != nil {
	//	panic(err)
	//}
	return p.mm.InitGenesis(ctx, p.appCodec, genesisState)

}

// BeginBlocker set function to BaseApp as a hook
func (p *ProtocolV0) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return p.mm.BeginBlock(ctx, req)
}

// EndBlocker sets function to BaseApp as a hook
func (p *ProtocolV0) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return p.mm.EndBlock(ctx, req)
}

// Stop makes okchain exit gracefully
func (p *ProtocolV0) Stop() {
	p.logger.Info(fmt.Sprintf("[%s]%s", utils.GoID, "OKChain stops notification."))
	p.stopped = true
}

func MakeEncodingConfig() EncodingConfig {
	cdc := codec.New()
	interfaceRegistry := types.NewInterfaceRegistry()
	marshaler := codec.NewHybridCodec(cdc, interfaceRegistry)

	encodingConfig := EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxGenerator:       authtypes.StdTxGenerator{Cdc: cdc},
		Amino:             cdc,
	}
	std.RegisterCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	ModuleBasics.RegisterCodec(encodingConfig.Amino)
	ModuleBasics.RegisterInterfaceModules(encodingConfig.InterfaceRegistry)
	return encodingConfig
}

func validateMsgHook(orderKeeper order.Keeper) ante.ValidateMsgHandler {
	return func(newCtx sdk.Context, msgs []sdk.Msg) error {
		for _, msg := range msgs {
			if msg != nil {
				switch assertedMsg := msg.(type) {
				case *ordertypes.MsgNewOrders:
					return order.ValidateMsgNewOrders(newCtx, orderKeeper, assertedMsg)
				case *ordertypes.MsgCancelOrders:
					return order.ValidateMsgCancelOrders(newCtx, orderKeeper, assertedMsg)
				}
			}
		}
		return nil
	}
}

func isSystemFreeHook(ctx sdk.Context, msgs []sdk.Msg) bool {
	if ctx.BlockHeight() < 1 {
		return true
	}

	return false
}

// ExportGenesis exports the genesis state for whole protocol
func (p *ProtocolV0) ExportGenesis(ctx sdk.Context) map[string]json.RawMessage {
	return p.mm.ExportGenesis(ctx, p.appCodec)
}

// SetLogger sets logger
func (p *ProtocolV0) SetLogger(log log.Logger) Protocol {
	p.logger = log
	return p
}

// SetParent sets parent implement
func (p *ProtocolV0) SetParent(parent Parent) Protocol {
	p.parent = parent
	return p
}

// GetParent gets parent implement
func (p *ProtocolV0) GetParent() Parent {
	if p.parent == nil {
		panic("parent is nil in protocol")
	}
	return p.parent
}
