package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	exported "github.com/cosmos/cosmos-sdk/x/auth/types"

	dex "github.com/okex/okchain/x/dex/types"
	"github.com/okex/okchain/x/order/types"
	token "github.com/okex/okchain/x/token/types"
)

// TokenKeeper : expected token keeper
type TokenKeeper interface {
	// Token balance
	GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.DecCoins
	LockCoins(ctx sdk.Context, addr sdk.AccAddress, coins sdk.DecCoins, lockCoinsType int) error
	UnlockCoins(ctx sdk.Context, addr sdk.AccAddress, coins sdk.DecCoins, lockCoinsType int) error
	BalanceAccount(ctx sdk.Context, addr sdk.AccAddress, outputCoins sdk.DecCoins, inputCoins sdk.DecCoins) error
	SendCoinsFromAccountToAccount(ctx sdk.Context, from, to sdk.AccAddress, amt sdk.DecCoins) error
	// Fee detail
	AddFeeDetail(ctx sdk.Context, from string, fee sdk.DecCoins, feeType string)
	GetAllLockedCoins(ctx sdk.Context) (locks []token.AccCoins)
	IterateLockedFees(ctx sdk.Context, cb func(acc sdk.AccAddress, coins sdk.DecCoins) (stop bool))
}

// BankKeeper : expected supply keeper
type AccountKeeper interface {
	GetModuleAccount(ctx sdk.Context, moduleName string) exported.ModuleAccountI
	GetModuleAddress(moduleName string) sdk.AccAddress
}

type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string,
		amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress,
		amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}

// DexKeeper : expected dex keeper
type DexKeeper interface {
	// TokenPair
	GetTokenPair(ctx sdk.Context, product string) *dex.TokenPair
	GetTokenPairs(ctx sdk.Context) []*dex.TokenPair
	SortProducts(ctx sdk.Context, products []string)
	SaveTokenPair(ctx sdk.Context, tokenPair *dex.TokenPair) error
	UpdateTokenPair(ctx sdk.Context, product string, tokenPair *dex.TokenPair)
	CheckTokenPairUnderDexDelist(ctx sdk.Context, product string) (isDelisting bool, err error)
	LockTokenPair(ctx sdk.Context, product string, lock *types.ProductLock)
	UnlockTokenPair(ctx sdk.Context, product string)
	IsTokenPairLocked(ctx sdk.Context, product string) bool
	GetLockedProductsCopy(ctx sdk.Context) *types.ProductLockMap
	IsAnyProductLocked(ctx sdk.Context) bool
}
