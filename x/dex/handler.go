package dex

import (
	"fmt"
	sdkerror "github.com/cosmos/cosmos-sdk/types/errors"
	"strconv"

	"github.com/okex/okchain/x/dex/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/okex/okchain/x/common/perf"
	"github.com/tendermint/tendermint/libs/log"
)

// NewHandler handles all "dex" type messages.
func NewHandler(k IKeeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		logger := ctx.Logger().With("module", ModuleName)

		var handlerFun func() (*sdk.Result, error)
		var name string
		switch msg := msg.(type) {
		case *MsgList:
			name = "handleMsgList"
			handlerFun = func() (*sdk.Result, error) {
				return handleMsgList(ctx, k, msg, logger)
			}
		case *MsgDeposit:
			name = "handleMsgDeposit"
			handlerFun = func() (*sdk.Result, error) {
				return handleMsgDeposit(ctx, k, msg, logger)
			}
		case *MsgWithdraw:
			name = "handleMsgWithDraw"
			handlerFun = func() (*sdk.Result, error) {
				return handleMsgWithDraw(ctx, k, msg, logger)
			}
		case *MsgTransferOwnership:
			name = "handleMsgTransferOwnership"
			handlerFun = func() (*sdk.Result, error) {
				return handleMsgTransferOwnership(ctx, k, msg, logger)
			}
		default:
			errMsg := fmt.Sprintf("unrecognized dex message type: %T", msg)
			return nil, sdkerror.Wrapf(sdkerror.ErrUnknownRequest, errMsg)
		}

		seq := perf.GetPerf().OnDeliverTxEnter(ctx, ModuleName, name)
		defer perf.GetPerf().OnDeliverTxExit(ctx, ModuleName, name, seq)
		return handlerFun()
	}
}

func handleMsgList(ctx sdk.Context, keeper IKeeper, msg *MsgList, logger log.Logger) (*sdk.Result, error) {

	if !keeper.GetTokenKeeper().TokenExist(ctx, msg.ListAsset) ||
		!keeper.GetTokenKeeper().TokenExist(ctx, msg.QuoteAsset) {
		return nil, sdkerror.Wrap(sdkerror.ErrInvalidCoins,
			fmt.Sprintf("%s or %s is not valid", msg.ListAsset, msg.QuoteAsset))
	}

	tokenPair := &TokenPair{
		BaseAssetSymbol:  msg.ListAsset,
		QuoteAssetSymbol: msg.QuoteAsset,
		InitPrice:        msg.InitPrice,
		MaxPriceDigit:    int64(DefaultMaxPriceDigitSize),
		MaxQuantityDigit: int64(DefaultMaxQuantityDigitSize),
		MinQuantity:      sdk.MustNewDecFromStr("0.00000001"),
		Owner:            msg.Owner,
		Delisting:        false,
		Deposits:         DefaultTokenPairDeposit,
		BlockHeight:      ctx.BlockHeight(),
	}

	// check whether a specific token pair exists with the symbols of base asset and quote asset
	// Note: aaa_bbb and bbb_aaa are actually one token pair
	if keeper.GetTokenPair(ctx, fmt.Sprintf("%s_%s", tokenPair.BaseAssetSymbol, tokenPair.QuoteAssetSymbol)) != nil ||
		keeper.GetTokenPair(ctx, fmt.Sprintf("%s_%s", tokenPair.QuoteAssetSymbol, tokenPair.BaseAssetSymbol)) != nil {
		return nil, types.ErrTokenPairExisted(tokenPair.BaseAssetSymbol, tokenPair.QuoteAssetSymbol)
	}

	// deduction fee
	feeCoins := keeper.GetParams(ctx).ListFee.ToCoins()
	err := keeper.GetBankKeeper().SendCoinsFromAccountToModule(ctx, msg.Owner, keeper.GetFeeCollector(), feeCoins)
	if err != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrInsufficientFunds, fmt.Sprintf("insufficient fee coins(need %s)",
			feeCoins.String()))
	}

	err2 := keeper.SaveTokenPair(ctx, tokenPair)
	if err2 != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrInternal, fmt.Sprintf("failed to SaveTokenPair: %s", err2.Error()))
	}

	logger.Debug(fmt.Sprintf("successfully handleMsgList: "+
		"BlockHeight: %d, Msg: %+v", ctx.BlockHeight(), msg))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute("list-asset", tokenPair.BaseAssetSymbol),
			sdk.NewAttribute("quote-asset", tokenPair.QuoteAssetSymbol),
			sdk.NewAttribute("init-price", tokenPair.InitPrice.String()),
			sdk.NewAttribute("max-price-digit", strconv.FormatInt(tokenPair.MaxPriceDigit, 10)),
			sdk.NewAttribute("max-size-digit", strconv.FormatInt(tokenPair.MaxQuantityDigit, 10)),
			sdk.NewAttribute("min-trade-size", tokenPair.MinQuantity.String()),
			sdk.NewAttribute("delisting", fmt.Sprintf("%t", tokenPair.Delisting)),
			sdk.NewAttribute(sdk.AttributeKeyFee, feeCoins.String()),
		),
	)

	return &sdk.Result{Events: ctx.EventManager().Events().ToABCIEvents()}, nil
}

func handleMsgDeposit(ctx sdk.Context, keeper IKeeper, msg *MsgDeposit, logger log.Logger) (*sdk.Result, error) {
	if err := keeper.Deposit(ctx, msg.Product, msg.Depositor, msg.Amount); err != nil {
		return nil, err
	}

	logger.Debug(fmt.Sprintf("successfully handleMsgDeposit: "+
		"BlockHeight: %d, Msg: %+v", ctx.BlockHeight(), msg))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
		),
	)

	return &sdk.Result{Events: ctx.EventManager().Events().ToABCIEvents()}, nil

}

func handleMsgWithDraw(ctx sdk.Context, keeper IKeeper, msg *MsgWithdraw, logger log.Logger) (*sdk.Result, error) {
	if err := keeper.Withdraw(ctx, msg.Product, msg.Depositor, msg.Amount); err != nil {
		return nil, err
	}

	logger.Debug(fmt.Sprintf("successfully handleMsgWithDraw: "+
		"BlockHeight: %d, Msg: %+v", ctx.BlockHeight(), msg))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
		),
	)

	return &sdk.Result{Events: ctx.EventManager().Events().ToABCIEvents()}, nil
}

func handleMsgTransferOwnership(ctx sdk.Context, keeper IKeeper, msg *MsgTransferOwnership,
	logger log.Logger) (*sdk.Result, error) {
	if err := keeper.TransferOwnership(ctx, msg.Product, msg.FromAddress, msg.ToAddress); err != nil {
		return nil, err
	}

	// deduction fee
	feeCoins := keeper.GetParams(ctx).TransferOwnershipFee.ToCoins()
	err := keeper.GetBankKeeper().SendCoinsFromAccountToModule(ctx, msg.FromAddress, keeper.GetFeeCollector(), feeCoins)
	if err != nil {
		return nil, sdkerror.Wrap(sdkerror.ErrInsufficientFunds, fmt.Sprintf("insufficient fee coins(need %s)",
			feeCoins.String()))
	}

	logger.Debug(fmt.Sprintf("successfully handleMsgTransferOwnership: "+
		"BlockHeight: %d, Msg: %+v", ctx.BlockHeight(), msg))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyFee, feeCoins.String()),
		),
	)
	return &sdk.Result{Events: ctx.EventManager().Events().ToABCIEvents()}, nil
}
