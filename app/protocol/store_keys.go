package protocol

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	upgradetypes "github.com/okex/okchain/x/upgrade/types"

	"github.com/okex/okchain/x/debug"
	"github.com/okex/okchain/x/dex"
	distr "github.com/okex/okchain/x/distribution"
	"github.com/okex/okchain/x/gov"
	"github.com/okex/okchain/x/order"
	"github.com/okex/okchain/x/params"
	"github.com/okex/okchain/x/poolswap"
	"github.com/okex/okchain/x/staking"
	"github.com/okex/okchain/x/token"
	"github.com/okex/okchain/x/upgrade"
)

// store keys used in all modules
var (
	kvStoreKeysMap = sdk.NewKVStoreKeys(
		authtypes.StoreKey,
		staking.StoreKey,
		banktypes.StoreKey,
		minttypes.StoreKey,
		slashingtypes.StoreKey,
		distr.StoreKey,
		gov.StoreKey,
		params.StoreKey,
		token.StoreKey, token.KeyMint, token.KeyLock,
		order.OrderStoreKey,
		upgrade.StoreKey,
		dex.StoreKey, dex.TokenPairStoreKey,
		debug.StoreKey,
		poolswap.StoreKey,
	)

	transientStoreKeysMap = sdk.NewTransientStoreKeys(staking.TStoreKey, params.TStoreKey)
)

// GetKVStoreKeysMap gets the map of all kv store keys
func GetKVStoreKeysMap() map[string]*sdk.KVStoreKey {
	return kvStoreKeysMap
}

// GetTransientStoreKeysMap gets the map of all transient store keys
func GetTransientStoreKeysMap() map[string]*sdk.TransientStoreKey {
	return transientStoreKeysMap
}

func GetMainStoreKey() *sdk.KVStoreKey {
	return kvStoreKeysMap[upgradetypes.StoreKey]
}
