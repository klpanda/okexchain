package poolswap

import (
	"fmt"
	sdkerror "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/okex/okchain/x/common"
	"github.com/okex/okchain/x/common/perf"
	"github.com/okex/okchain/x/poolswap/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler creates an sdk.Handler for all the poolswap type messages
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		var handlerFun func() (*sdk.Result, error)
		var name string
		switch msg := msg.(type) {
		case *types.MsgAddLiquidity:
			name = "handleMsgAddLiquidity"
			handlerFun = func() (*sdk.Result, error) {
				return handleMsgAddLiquidity(ctx, k, *msg)
			}
		case *types.MsgRemoveLiquidity:
			name = "handleMsgRemoveLiquidity"
			handlerFun = func() (*sdk.Result, error) {
				return handleMsgRemoveLiquidity(ctx, k, *msg)
			}
		case *types.MsgCreateExchange:
			name = "handleMsgCreateExchange"
			handlerFun = func() (*sdk.Result, error) {
				return handleMsgCreateExchange(ctx, k, *msg)
			}
		case *types.MsgTokenToNativeToken:
			name = "handleMsgTokenToNativeToken"
			handlerFun = func() (*sdk.Result, error) {
				return handleMsgTokenToTokenExchange(ctx, k, *msg)
			}
		default:
			errMsg := fmt.Sprintf("Invalid msg type: %v", msg.Type())
			return nil, sdkerror.Wrap(sdkerror.ErrUnknownRequest, errMsg)
		}
		seq := perf.GetPerf().OnDeliverTxEnter(ctx, types.ModuleName, name)
		defer perf.GetPerf().OnDeliverTxExit(ctx, types.ModuleName, name, seq)
		return handlerFun()
	}
}

func handleMsgTokenToTokenExchange(ctx sdk.Context, k Keeper, msg types.MsgTokenToNativeToken) (*sdk.Result, error) {
	if msg.SoldTokenAmount.Denom != sdk.DefaultBondDenom && msg.MinBoughtTokenAmount.Denom != sdk.DefaultBondDenom {
		return handleMsgTokenToToken(ctx, k, msg)
	}
	return handleMsgTokenToNativeToken(ctx, k, msg)
}

func handleMsgCreateExchange(ctx sdk.Context, k Keeper, msg types.MsgCreateExchange) (*sdk.Result, error) {
	event := sdk.NewEvent(sdk.EventTypeMessage, sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName))
	err := k.IsTokenExist(ctx, msg.Token)
	if err != nil {
		return nil, sdkerror.Wrapf(sdkerror.ErrInternal, err.Error())
	}

	tokenPair := msg.Token + "_" + common.NativeToken

	swapTokenPair, err := k.GetSwapTokenPair(ctx, tokenPair)
	if err == nil {
		return nil, sdkerror.Wrapf(sdkerror.ErrInternal, "Failed: exchange already exists")
	}

	poolName := types.PoolTokenPrefix + msg.Token
	baseToken := sdk.NewDecCoinFromDec(msg.Token, sdk.ZeroDec())
	quoteToken := sdk.NewDecCoinFromDec(common.NativeToken, sdk.ZeroDec())
	poolToken, err := k.GetPoolTokenInfo(ctx, poolName)
	if err == nil {
		return nil, sdkerror.Wrap(sdkerror.ErrInternal, "Failed: pool token already exists")
	}
	k.NewPoolToken(ctx, poolName)
	event = event.AppendAttributes(sdk.NewAttribute("pool-token", poolToken.OriginalSymbol))
	swapTokenPair.BasePooledCoin = baseToken
	swapTokenPair.QuotePooledCoin = quoteToken
	swapTokenPair.PoolTokenName = poolName

	k.SetSwapTokenPair(ctx, tokenPair, swapTokenPair)

	event = event.AppendAttributes(sdk.NewAttribute("token-pair", tokenPair))
	ctx.EventManager().EmitEvent(event)
	return &sdk.Result{Events: ctx.EventManager().Events().ToABCIEvents()}, nil
}

func handleMsgAddLiquidity(ctx sdk.Context, k Keeper, msg types.MsgAddLiquidity) (*sdk.Result, error) {
	event := sdk.NewEvent(sdk.EventTypeMessage, sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName))

	if msg.Deadline < ctx.BlockTime().Unix() {
		return nil, sdkerror.Wrap(sdkerror.ErrInternal, "Failed: block time exceeded deadline")
	}
	swapTokenPair, err := k.GetSwapTokenPair(ctx, msg.GetSwapTokenPair())
	if err != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrInternal, err.Error())
	}
	baseTokens := sdk.NewDecCoinFromDec(msg.MaxBaseAmount.Denom, sdk.ZeroDec())
	var liquidity sdk.Dec
	poolToken, err := k.GetPoolTokenInfo(ctx, swapTokenPair.PoolTokenName)
	if err != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrInternal,
			fmt.Sprintf("failed to get pool token %s : %s", swapTokenPair.PoolTokenName, err.Error()))
	}
	if swapTokenPair.QuotePooledCoin.Amount.IsZero() && swapTokenPair.BasePooledCoin.Amount.IsZero() {
		baseTokens.Amount = msg.MaxBaseAmount.Amount
		liquidity = sdk.NewDec(1)
	} else if swapTokenPair.BasePooledCoin.IsPositive() && swapTokenPair.QuotePooledCoin.IsPositive() {
		baseTokens.Amount = msg.QuoteAmount.Amount.Mul(swapTokenPair.BasePooledCoin.Amount).Quo(swapTokenPair.QuotePooledCoin.Amount)
		if poolToken.TotalSupply.IsZero() {
			return nil, sdkerror.Wrap(sdkerror.ErrInternal,
				fmt.Sprintf("unexpected totalSupply in pool token %s", poolToken.String()))
		}
		liquidity = msg.QuoteAmount.Amount.Quo(swapTokenPair.QuotePooledCoin.Amount).Mul(poolToken.TotalSupply)
	} else {
		return nil, sdkerror.Wrap(sdkerror.ErrInternal,
			fmt.Sprintf("invalid token pair %s", swapTokenPair.String()))
	}
	if baseTokens.Amount.GT(msg.MaxBaseAmount.Amount) {
		return nil, sdkerror.Wrap(sdkerror.ErrInternal,
			"The required base token amount are greater than MaxBaseAmount")
	}
	if liquidity.LT(msg.MinLiquidity) {
		return nil, sdkerror.Wrap(sdkerror.ErrInternal,
			"The available liquidity is less than MinLiquidity")
	}

	// transfer coins
	coins := sdk.DecCoins{
		msg.QuoteAmount,
		baseTokens,
	}
	coins = coins.Sort()
	err = k.SendCoinsToPool(ctx, coins, msg.Sender)
	if err != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrInsufficientFunds,
			fmt.Sprintf("insufficient coins %s", err.Error()))
	}
	// update swapTokenPair
	swapTokenPair.QuotePooledCoin = swapTokenPair.QuotePooledCoin.Add(msg.QuoteAmount)
	swapTokenPair.BasePooledCoin = swapTokenPair.BasePooledCoin.Add(baseTokens)
	k.SetSwapTokenPair(ctx, msg.GetSwapTokenPair(), swapTokenPair)

	// update poolToken
	poolCoins := sdk.NewDecCoinFromDec(poolToken.Symbol, liquidity)
	err = k.MintPoolCoinsToUser(ctx, sdk.DecCoins{poolCoins}, msg.Sender)
	if err != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrInternal,
			"failed to mint pool token")
	}

	event.AppendAttributes(sdk.NewAttribute("liquidity", liquidity.String()))
	event.AppendAttributes(sdk.NewAttribute("baseAmount", baseTokens.String()))
	ctx.EventManager().EmitEvent(event)
	return &sdk.Result{Events: ctx.EventManager().Events().ToABCIEvents()}, nil
}

func handleMsgRemoveLiquidity(ctx sdk.Context, k Keeper, msg types.MsgRemoveLiquidity) (*sdk.Result, error) {
	event := sdk.NewEvent(sdk.EventTypeMessage, sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName))

	if msg.Deadline < ctx.BlockTime().Unix() {
		return nil, sdkerror.Wrap(sdkerror.ErrInternal, "Failed: block time exceeded deadline")
	}
	swapTokenPair, err := k.GetSwapTokenPair(ctx, msg.GetSwapTokenPair())
	if err != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrInternal, err.Error())
	}

	liquidity := msg.Liquidity
	poolTokenAmount, err := k.GetPoolTokenAmount(ctx, swapTokenPair.PoolTokenName)
	if err != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrInternal,
			fmt.Sprintf("failed to get pool token %s : %s", swapTokenPair.PoolTokenName, err.Error()))
	}
	if poolTokenAmount.LT(liquidity) {
		return nil, sdkerror.Wrap(sdkerror.ErrInsufficientFunds, "insufficient pool token")
	}

	baseDec := swapTokenPair.BasePooledCoin.Amount.Mul(liquidity).Quo(poolTokenAmount)
	quoteDec := swapTokenPair.QuotePooledCoin.Amount.Mul(liquidity).Quo(poolTokenAmount)
	baseAmount := sdk.NewDecCoinFromDec(swapTokenPair.BasePooledCoin.Denom, baseDec)
	quoteAmount := sdk.NewDecCoinFromDec(swapTokenPair.QuotePooledCoin.Denom, quoteDec)

	if baseAmount.IsLT(msg.MinBaseAmount) {
		return nil, sdkerror.Wrap(sdkerror.ErrInternal,
			"Failed: available base amount are less than least base amount")
	}
	if quoteAmount.IsLT(msg.MinQuoteAmount) {
		return nil, sdkerror.Wrap(sdkerror.ErrInternal,
			"Failed: available quote amount are less than least quote amount")
	}

	// transfer coins
	coins := sdk.DecCoins{
		baseAmount,
		quoteAmount,
	}
	coins = coins.Sort()
	err = k.SendCoinsFromPoolToAccount(ctx, coins, msg.Sender)
	if err != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrInsufficientFunds, "insufficient coins")
	}
	// update swapTokenPair
	swapTokenPair.QuotePooledCoin = swapTokenPair.QuotePooledCoin.Sub(quoteAmount)
	swapTokenPair.BasePooledCoin = swapTokenPair.BasePooledCoin.Sub(baseAmount)
	k.SetSwapTokenPair(ctx, msg.GetSwapTokenPair(), swapTokenPair)

	// update poolToken
	poolCoins := sdk.NewDecCoinFromDec(swapTokenPair.PoolTokenName, liquidity)
	err = k.BurnPoolCoinsFromUser(ctx, sdk.DecCoins{poolCoins}, msg.Sender)
	if err != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrInternal, "failed to burn pool token")
	}

	event.AppendAttributes(sdk.NewAttribute("quoteAmount", quoteAmount.String()))
	event.AppendAttributes(sdk.NewAttribute("baseAmount", baseAmount.String()))
	ctx.EventManager().EmitEvent(event)
	return &sdk.Result{Events: ctx.EventManager().Events().ToABCIEvents()}, nil
}

func handleMsgTokenToNativeToken(ctx sdk.Context, k Keeper, msg types.MsgTokenToNativeToken) (*sdk.Result, error) {
	event := sdk.NewEvent(sdk.EventTypeMessage, sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName))

	if err := common.HasSufficientCoins(msg.Sender, k.GetTokenKeeper().GetCoins(ctx, msg.Sender),
		sdk.DecCoins{msg.SoldTokenAmount}); err != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrInsufficientFunds, err.Error())
	}
	if msg.Deadline < ctx.BlockTime().Unix() {
		return nil, sdkerror.Wrap(sdkerror.ErrInternal, "Failed: block time exceeded deadline")
	}
	swapTokenPair, err := k.GetSwapTokenPair(ctx, msg.GetSwapTokenPair())
	if err != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrInternal, err.Error())
	}
	params := k.GetParams(ctx)
	tokenBuy := calculateTokenToBuy(swapTokenPair, msg, params)
	if tokenBuy.Amount.LT(msg.MinBoughtTokenAmount.Amount) {
		return nil, sdkerror.Wrap(sdkerror.ErrInternal,
			fmt.Sprintf("Failed: expected minimum token to buy is %s but got %s", msg.MinBoughtTokenAmount, tokenBuy))
	}

	err = swapTokenNativeToken(ctx, k, swapTokenPair, tokenBuy, msg)
	if err != nil {
		return nil, err
	}
	event.AppendAttributes(sdk.NewAttribute("bought_token_amount", tokenBuy.String()))
	event.AppendAttributes(sdk.NewAttribute("recipient", msg.Recipient.String()))
	ctx.EventManager().EmitEvent(event)
	return &sdk.Result{Events: ctx.EventManager().Events().ToABCIEvents()}, nil
}

func handleMsgTokenToToken(ctx sdk.Context, k Keeper, msg types.MsgTokenToNativeToken) (*sdk.Result, error) {
	event := sdk.NewEvent(sdk.EventTypeMessage, sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName))

	if msg.Deadline < ctx.BlockTime().Unix() {
		return nil, sdkerror.Wrap(sdkerror.ErrInternal, "Failed: block time exceeded deadline")
	}
	if err := common.HasSufficientCoins(msg.Sender, k.GetTokenKeeper().GetCoins(ctx, msg.Sender),
		sdk.DecCoins{msg.SoldTokenAmount}); err != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrInsufficientFunds, err.Error())
	}
	tokenPairOne := msg.SoldTokenAmount.Denom + "_" + sdk.DefaultBondDenom
	swapTokenPairOne, err := k.GetSwapTokenPair(ctx, tokenPairOne)
	if err != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrInternal, err.Error())
	}
	tokenPairTwo := msg.MinBoughtTokenAmount.Denom + "_" + sdk.DefaultBondDenom
	swapTokenPairTwo, err := k.GetSwapTokenPair(ctx, tokenPairTwo)
	if err != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrInternal, err.Error())
	}

	nativeAmount := sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.MustNewDecFromStr("0"))
	params := k.GetParams(ctx)
	msgOne := msg
	msgOne.MinBoughtTokenAmount = nativeAmount
	tokenNative := calculateTokenToBuy(swapTokenPairOne, msgOne, params)

	msgTwo := msg
	msgTwo.SoldTokenAmount = tokenNative
	tokenBuy := calculateTokenToBuy(swapTokenPairOne, msgTwo, params)

	if tokenBuy.Amount.LT(msg.MinBoughtTokenAmount.Amount) {
		return nil, sdkerror.Wrap(sdkerror.ErrInternal,
			fmt.Sprintf("Failed: expected minimum token to buy is %s but got %s", msg.MinBoughtTokenAmount, tokenBuy))
	}

	err = swapTokenNativeToken(ctx, k, swapTokenPairOne, tokenNative, msgOne)
	if err != nil {
		return nil, err
	}
	//TODO if fail,revert last swap
	err = swapTokenNativeToken(ctx, k, swapTokenPairTwo, tokenBuy, msgTwo)
	if err != nil {
		return nil, err
	}

	event.AppendAttributes(sdk.NewAttribute("bought_token_amount", tokenBuy.String()))
	event.AppendAttributes(sdk.NewAttribute("recipient", msg.Recipient.String()))
	ctx.EventManager().EmitEvent(event)
	return &sdk.Result{Events: ctx.EventManager().Events().ToABCIEvents()}, nil
}

//calculate the amount to buy
func calculateTokenToBuy(swapTokenPair SwapTokenPair, msg types.MsgTokenToNativeToken, params types.Params) sdk.DecCoin {
	var inputReserve, outputReserve sdk.Dec
	if msg.SoldTokenAmount.Denom == sdk.DefaultBondDenom {
		inputReserve = swapTokenPair.QuotePooledCoin.Amount
		outputReserve = swapTokenPair.BasePooledCoin.Amount
	} else {
		inputReserve = swapTokenPair.BasePooledCoin.Amount
		outputReserve = swapTokenPair.QuotePooledCoin.Amount
	}
	tokenBuyAmt := getInputPrice(msg.SoldTokenAmount.Amount, inputReserve, outputReserve, params.FeeRate)
	tokenBuy := sdk.NewDecCoinFromDec(msg.MinBoughtTokenAmount.Denom, tokenBuyAmt)

	return tokenBuy
}

func swapTokenNativeToken(
	ctx sdk.Context, k Keeper, swapTokenPair SwapTokenPair, tokenBuy sdk.DecCoin,
	msg types.MsgTokenToNativeToken,
) error {
	// transfer coins
	err := k.SendCoinsToPool(ctx, sdk.DecCoins{msg.SoldTokenAmount}, msg.Sender)
	if err != nil {
		return sdkerror.Wrap(sdkerror.ErrInsufficientFunds, "insufficient Coins")
	}

	err = k.SendCoinsFromPoolToAccount(ctx, sdk.DecCoins{tokenBuy}, msg.Recipient)
	if err != nil {
		return sdkerror.Wrap(sdkerror.ErrInsufficientFunds, "insufficient Coins")
	}

	// update swapTokenPair
	if msg.SoldTokenAmount.Denom == sdk.DefaultBondDenom {
		swapTokenPair.QuotePooledCoin = swapTokenPair.QuotePooledCoin.Add(msg.SoldTokenAmount)
		swapTokenPair.BasePooledCoin = swapTokenPair.BasePooledCoin.Sub(tokenBuy)
	} else {
		swapTokenPair.QuotePooledCoin = swapTokenPair.QuotePooledCoin.Sub(tokenBuy)
		swapTokenPair.BasePooledCoin = swapTokenPair.BasePooledCoin.Add(msg.SoldTokenAmount)
	}
	k.SetSwapTokenPair(ctx, msg.GetSwapTokenPair(), swapTokenPair)
	return nil
}

func getInputPrice(inputAmount, inputReserve, outputReserve, feeRate sdk.Dec) sdk.Dec {
	inputAmountWithFee := inputAmount.Mul(sdk.OneDec().Sub(feeRate))
	numerator := inputAmountWithFee.Mul(outputReserve)
	denominator := inputReserve.Add(inputAmountWithFee)
	return numerator.Quo(denominator)
}
