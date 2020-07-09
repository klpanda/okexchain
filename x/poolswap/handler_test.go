package poolswap

import (
	"fmt"
	sdkerror "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/okex/okchain/x/poolswap/types"
	token "github.com/okex/okchain/x/token/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestHandleMsgCreateExchange(t *testing.T) {
	mapp, addrKeysSlice := getMockApp(t, 1)
	keeper := mapp.swapKeeper
	mapp.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 2}})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{}).WithBlockHeight(10)
	testToken := initToken(types.TestBasePooledToken)

	mapp.BankKeeper.SetSupply(ctx, banktypes.NewSupply(mapp.TotalCoinsSupply))
	handler := NewHandler(keeper)
	msg := types.NewMsgCreateExchange(testToken.Symbol, addrKeysSlice[0].Address)

	// test case1: token is not exist
	result, err := handler(ctx, &msg)
	require.Nil(t, err)
	require.NotNil(t, result.Log)

	mapp.tokenKeeper.NewToken(ctx, testToken)

	// test case2: success
	result, err = handler(ctx, &msg)
	require.Nil(t, err)
	require.Equal(t, "", result.Log)

	// check account balance
	acc := mapp.AccountKeeper.GetAccount(ctx, addrKeysSlice[0].Address)
	expectCoins := sdk.DecCoins{
		sdk.NewDecCoinFromDec(types.TestQuotePooledToken, sdk.MustNewDecFromStr("100")),
		sdk.NewDecCoinFromDec(types.TestBasePooledToken, sdk.MustNewDecFromStr("100")),
		sdk.NewDecCoinFromDec(types.TestBasePooledToken2, sdk.MustNewDecFromStr("100")),
		sdk.NewDecCoinFromDec(types.TestBasePooledToken3, sdk.MustNewDecFromStr("100")),
	}
	require.EqualValues(t, expectCoins.String(), mapp.BankKeeper.GetAllBalances(ctx, acc.GetAddress()).String())

	expectSwapTokenPair := types.GetTestSwapTokenPair()
	swapTokenPair, err := keeper.GetSwapTokenPair(ctx, types.TestSwapTokenPairName)
	require.Nil(t, err)
	require.EqualValues(t, expectSwapTokenPair, swapTokenPair)

	// test case3: swapTokenPair already exists
	result, err = handler(ctx, &msg)
	require.Nil(t, err)
	require.NotNil(t, result.Log)
}

func initToken(name string) token.Token {
	return token.Token{
		Description:         name,
		Symbol:              name,
		OriginalSymbol:      name,
		WholeName:           name,
		OriginalTotalSupply: sdk.NewDec(0),
		TotalSupply:         sdk.NewDec(0),
		Owner:               authtypes.NewModuleAddress(ModuleName),
		Type:                1,
		Mintable:            true,
	}
}

func TestHandleMsgAddLiquidity(t *testing.T) {
	mapp, addrKeysSlice := getMockAppWithBalance(t, 1, 100000)
	keeper := mapp.swapKeeper
	mapp.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 2}})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{}).WithBlockHeight(10).WithBlockTime(time.Now())
	testToken := initToken(types.TestBasePooledToken)

	mapp.BankKeeper.SetSupply(ctx, banktypes.NewSupply(mapp.TotalCoinsSupply))
	handler := NewHandler(keeper)
	msg := types.NewMsgCreateExchange(testToken.Symbol, addrKeysSlice[0].Address)
	mapp.tokenKeeper.NewToken(ctx, testToken)

	result, err := handler(ctx, &msg)
	require.Nil(t, err)
	require.Equal(t, "", result.Log)

	minLiquidity := sdk.NewDec(1)
	maxBaseAmount := sdk.NewDecCoinFromDec(types.TestBasePooledToken, sdk.NewDec(10000))
	quoteAmount := sdk.NewDecCoinFromDec(types.TestQuotePooledToken, sdk.NewDec(10000))
	nonExistMaxBaseAmount := sdk.NewDecCoinFromDec("abc", sdk.NewDec(10000))
	invalidMinLiquidity := sdk.NewDec(1000)
	invalidMaxBaseAmount := sdk.NewDecCoinFromDec(types.TestBasePooledToken, sdk.NewDec(1))
	insufficientMaxBaseAmount := sdk.NewDecCoinFromDec(types.TestBasePooledToken, sdk.NewDec(1000000))
	insufficientQuoteAmount := sdk.NewDecCoinFromDec(types.TestQuotePooledToken, sdk.NewDec(1000000))
	deadLine := time.Now().Unix()
	addr := addrKeysSlice[0].Address

	tests := []struct {
		testCase         string
		minLiquidity     sdk.Dec
		maxBaseAmount    sdk.DecCoin
		quoteAmount      sdk.DecCoin
		deadLine         int64
		addr             sdk.AccAddress
		exceptResultCode string
	}{
		{"success", minLiquidity, maxBaseAmount, quoteAmount, deadLine, addr, ""},
		{"blockTime exceeded deadline", minLiquidity, maxBaseAmount, quoteAmount, 0, addr, sdkerror.ErrInternal.Error()},
		{"unknown swapTokenPair", minLiquidity, nonExistMaxBaseAmount, quoteAmount, deadLine, addr, sdkerror.ErrInternal.Error()},
		{"The required baseTokens are greater than MaxBaseAmount", minLiquidity, invalidMaxBaseAmount, quoteAmount, deadLine, addr, sdkerror.ErrInternal.Error()},
		{"The available liquidity is less than MinLiquidity", invalidMinLiquidity, maxBaseAmount, quoteAmount, deadLine, addr, sdkerror.ErrInternal.Error()},
		{"insufficient Coins", minLiquidity, insufficientMaxBaseAmount, insufficientQuoteAmount, deadLine, addr, sdkerror.ErrInsufficientFunds.Error()},
	}

	for _, testCase := range tests {
		addLiquidityMsg := types.NewMsgAddLiquidity(testCase.minLiquidity, testCase.maxBaseAmount, testCase.quoteAmount, testCase.deadLine, testCase.addr)
		result, err = handler(ctx, &addLiquidityMsg)
		if err != nil {
			require.Contains(t, err.Error(), testCase.exceptResultCode)
		}
	}
}

func TestHandleMsgRemoveLiquidity(t *testing.T) {
	mapp, addrKeysSlice := getMockAppWithBalance(t, 1, 100000)
	keeper := mapp.swapKeeper
	mapp.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 2}})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{}).WithBlockHeight(10).WithBlockTime(time.Now())
	testToken := initToken(types.TestBasePooledToken)

	mapp.BankKeeper.SetSupply(ctx, banktypes.NewSupply(mapp.TotalCoinsSupply))
	handler := NewHandler(keeper)
	msg := types.NewMsgCreateExchange(testToken.Symbol, addrKeysSlice[0].Address)
	mapp.tokenKeeper.NewToken(ctx, testToken)

	result, err := handler(ctx, &msg)
	require.Nil(t, err)
	require.Equal(t, "", result.Log)

	minLiquidity := sdk.NewDec(1)
	maxBaseAmount := sdk.NewDecCoinFromDec(types.TestBasePooledToken, sdk.NewDec(10000))
	quoteAmount := sdk.NewDecCoinFromDec(types.TestQuotePooledToken, sdk.NewDec(10000))
	deadLine := time.Now().Unix()
	addr := addrKeysSlice[0].Address

	addLiquidityMsg := types.NewMsgAddLiquidity(minLiquidity, maxBaseAmount, quoteAmount, deadLine, addr)
	result, err = handler(ctx, &addLiquidityMsg)
	require.Nil(t, err)
	require.Equal(t, "", result.Log)

	liquidity, err := sdk.NewDecFromStr("0.01")
	require.Nil(t, err)
	minBaseAmount := sdk.NewDecCoinFromDec(types.TestBasePooledToken, sdk.NewDec(1))
	minQuoteAmount := sdk.NewDecCoinFromDec(types.TestQuotePooledToken, sdk.NewDec(1))
	nonExistMinBaseAmount := sdk.NewDecCoinFromDec("abc", sdk.NewDec(10000))
	invalidMinBaseAmount := sdk.NewDecCoinFromDec(types.TestBasePooledToken, sdk.NewDec(1000000))
	invalidMinQuoteAmount := sdk.NewDecCoinFromDec(types.TestQuotePooledToken, sdk.NewDec(1000000))
	invalidLiquidity := sdk.NewDec(1)

	tests := []struct {
		testCase         string
		liquidity        sdk.Dec
		minBaseAmount    sdk.DecCoin
		minQuoteAmount   sdk.DecCoin
		deadLine         int64
		addr             sdk.AccAddress
		exceptResultCode string
	}{
		{"success", liquidity, minBaseAmount, minQuoteAmount, deadLine, addr, ""},
		{"blockTime exceeded deadline", liquidity, minBaseAmount, minQuoteAmount, 0, addr, sdkerror.ErrInternal.Error()},
		{"unknown swapTokenPair", liquidity, nonExistMinBaseAmount, minQuoteAmount, deadLine, addr, sdkerror.ErrInternal.Error()},
		{"The available baseAmount are less than MinBaseAmount", liquidity, invalidMinBaseAmount, minQuoteAmount, deadLine, addr, sdkerror.ErrInternal.Error()},
		{"The available quoteAmount are less than MinQuoteAmount", liquidity, minBaseAmount, invalidMinQuoteAmount, deadLine, addr, sdkerror.ErrInternal.Error()},
		{"insufficient poolToken", invalidLiquidity, minBaseAmount, minQuoteAmount, deadLine, addr, sdkerror.ErrInsufficientFunds.Error()},
	}

	for _, testCase := range tests {
		addLiquidityMsg := types.NewMsgRemoveLiquidity(testCase.liquidity, testCase.minBaseAmount, testCase.minQuoteAmount, testCase.deadLine, testCase.addr)
		result, err = handler(ctx, &addLiquidityMsg)
		if err != nil {
			require.Contains(t, err.Error(), testCase.exceptResultCode)
		}
	}
}

func TestHandleMsgTokenToTokenExchange(t *testing.T) {
	mapp, addrKeysSlice := getMockAppWithBalance(t, 1, 100000)
	keeper := mapp.swapKeeper
	mapp.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 2}})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{}).WithBlockHeight(10).WithBlockTime(time.Now())
	testToken := initToken(types.TestBasePooledToken)
	secondTestTokenName := types.TestBasePooledToken2
	secondTestToken := initToken(secondTestTokenName)
	mapp.swapKeeper.SetParams(ctx, types.DefaultParams())

	mapp.BankKeeper.SetSupply(ctx, banktypes.NewSupply(mapp.TotalCoinsSupply))
	handler := NewHandler(keeper)
	msgCreateExchange := types.NewMsgCreateExchange(testToken.Symbol, addrKeysSlice[0].Address)
	msgCreateExchange2 := types.NewMsgCreateExchange(secondTestToken.Symbol, addrKeysSlice[0].Address)
	mapp.tokenKeeper.NewToken(ctx, testToken)
	mapp.tokenKeeper.NewToken(ctx, secondTestToken)

	result, err := handler(ctx, &msgCreateExchange)
	require.Nil(t, err)
	require.Equal(t, "", result.Log)
	result, err = handler(ctx, &msgCreateExchange2)
	require.Nil(t, err)
	require.Equal(t, "", result.Log)

	minLiquidity := sdk.NewDec(1)
	maxBaseAmount := sdk.NewDecCoinFromDec(types.TestBasePooledToken, sdk.NewDec(10000))
	maxBaseAmount2 := sdk.NewDecCoinFromDec(secondTestTokenName, sdk.NewDec(10000))
	quoteAmount := sdk.NewDecCoinFromDec(types.TestQuotePooledToken, sdk.NewDec(10000))
	deadLine := time.Now().Unix()
	addr := addrKeysSlice[0].Address

	addLiquidityMsg := types.NewMsgAddLiquidity(minLiquidity, maxBaseAmount, quoteAmount, deadLine, addr)
	result, err = handler(ctx, &addLiquidityMsg)
	require.Nil(t, err)
	require.Equal(t, "", result.Log)
	addLiquidityMsg2 := types.NewMsgAddLiquidity(minLiquidity, maxBaseAmount2, quoteAmount, deadLine, addr)
	result, err = handler(ctx, &addLiquidityMsg)
	require.Nil(t, err)
	require.Equal(t, "", result.Log)
	result, err = handler(ctx, &addLiquidityMsg2)
	require.Nil(t, err)
	require.Equal(t, "", result.Log)

	minBoughtTokenAmount := sdk.NewDecCoinFromDec(types.TestBasePooledToken, sdk.NewDec(1))
	deadLine = time.Now().Unix()
	soldTokenAmount := sdk.NewDecCoinFromDec(types.TestQuotePooledToken, sdk.NewDec(2))
	insufficientSoldTokenAmount := sdk.NewDecCoinFromDec(types.TestQuotePooledToken, sdk.NewDec(100000000))
	unkownBoughtTokenAmount := sdk.NewDecCoinFromDec(types.TestBasePooledToken3, sdk.NewDec(1))
	invalidMinBoughtTokenAmount := sdk.NewDecCoinFromDec(types.TestBasePooledToken, sdk.NewDec(100000))

	minBoughtTokenAmount2 := sdk.NewDecCoinFromDec(secondTestTokenName, sdk.NewDec(1))
	unkownBountTokenAmount2 := sdk.NewDecCoinFromDec(types.TestBasePooledToken3, sdk.NewDec(1))
	soldTokenAmount2 := sdk.NewDecCoinFromDec(types.TestBasePooledToken, sdk.NewDec(2))
	unkownSoldTokenAmount2 := sdk.NewDecCoinFromDec(types.TestBasePooledToken3, sdk.NewDec(1))
	insufficientSoldTokenAmount2 := sdk.NewDecCoinFromDec(types.TestBasePooledToken, sdk.NewDec(10000000))
	invalidMinBoughtTokenAmount2 := sdk.NewDecCoinFromDec(secondTestTokenName, sdk.NewDec(100000))

	tests := []struct {
		testCase             string
		minBoughtTokenAmount sdk.DecCoin
		soldTokenAmount      sdk.DecCoin
		deadLine             int64
		recipient            sdk.AccAddress
		addr                 sdk.AccAddress
		exceptResultCode     string
	}{
		{"(tokenToNativeToken) success", minBoughtTokenAmount, soldTokenAmount, deadLine, addr, addr, ""},
		{"(tokenToToken) success", minBoughtTokenAmount2, soldTokenAmount2, deadLine, addr, addr, ""},
		{"(tokenToNativeToken) blockTime exceeded deadline", minBoughtTokenAmount, soldTokenAmount, 0, addr, addr, sdkerror.ErrInternal.Error()},
		{"(tokenToToken) blockTime exceeded deadline", minBoughtTokenAmount2, soldTokenAmount2, 0, addr, addr, sdkerror.ErrInternal.Error()},
		{"(tokenToNativeToken) insufficient SoldTokenAmount", minBoughtTokenAmount, insufficientSoldTokenAmount, deadLine, addr, addr, sdkerror.ErrInsufficientFunds.Error()},
		{"(tokenToToken) insufficient SoldTokenAmount", minBoughtTokenAmount2, insufficientSoldTokenAmount2, deadLine, addr, addr, sdkerror.ErrInsufficientFunds.Error()},
		{"(tokenToNativeToken) unknown swapTokenPair", unkownBoughtTokenAmount, soldTokenAmount, deadLine, addr, addr, sdkerror.ErrInternal.Error()},
		{"(tokenToToken) unknown swapTokenPair", unkownBountTokenAmount2, soldTokenAmount2, deadLine, addr, addr, sdkerror.ErrInternal.Error()},
		{"(tokenToToken) unknown swapTokenPair2", minBoughtTokenAmount2, unkownSoldTokenAmount2, deadLine, addr, addr, sdkerror.ErrInternal.Error()},
		{"(tokenToNativeToken) The available BoughtTokenAmount are less than minBoughtTokenAmount", invalidMinBoughtTokenAmount, soldTokenAmount, deadLine, addr, addr, sdkerror.ErrInternal.Error()},
		{"(tokenToToken) The available BoughtTokenAmount are less than minBoughtTokenAmount", invalidMinBoughtTokenAmount2, soldTokenAmount2, deadLine, addr, addr, sdkerror.ErrInternal.Error()},
	}

	for _, testCase := range tests {
		fmt.Println(testCase.testCase)
		addLiquidityMsg := types.NewMsgTokenToNativeToken(testCase.soldTokenAmount, testCase.minBoughtTokenAmount, testCase.deadLine, testCase.recipient, testCase.addr)
		result, err = handler(ctx, &addLiquidityMsg)
		if err != nil {
			require.Contains(t, err.Error(), testCase.exceptResultCode)
		}
	}
}
