package keeper

import (
	"fmt"
	sdkerror "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/okex/okchain/x/poolswap/types"
	token "github.com/okex/okchain/x/token/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestKeeper_IsTokenExistTable(t *testing.T) {
	mapp, _ := GetTestInput(t, 1)
	keeper := mapp.swapKeeper
	mapp.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: 2}})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{}).WithBlockHeight(10)
	mapp.BankKeeper.SetSupply(ctx, banktypes.NewSupply(mapp.TotalCoinsSupply))

	tests := []struct {
		testCase         string
		tokennames       []string
		tokentypes       []int
		tokenname        string
		exceptResultCode string
	}{
		{"token is not exist", []string{"toa", "tob"}, []int{1, 1}, "nota", sdkerror.ErrInternal.Error()},
		{"token is not exist", nil, nil, "nota", sdkerror.ErrInternal.Error()},
		{"token is exist", []string{"boa", "bob"}, []int{1, 1}, "boa", ""},
		{"token is pool token", []string{"tkoa", "tkob"}, []int{1, 2}, "tkob", sdkerror.ErrInvalidCoins.Error()},
	}

	for _, testCase := range tests {
		fmt.Println(testCase.testCase)
		genToken(mapp, ctx, testCase.tokennames, testCase.tokentypes)
		err := keeper.IsTokenExist(ctx, testCase.tokenname)
		if nil != err {
			require.Contains(t, err.Error(), testCase.exceptResultCode)
		}
	}

}

func genToken(mapp *TestInput, ctx sdk.Context, tokennames []string, tokentypes []int) {
	for i, t := range tokennames {
		tok := token.Token{
			Description:         t,
			Symbol:              t,
			OriginalSymbol:      t,
			WholeName:           t,
			OriginalTotalSupply: sdk.NewDec(0),
			TotalSupply:         sdk.NewDec(0),
			Owner:               authtypes.NewModuleAddress(types.ModuleName),
			Mintable:            true,
			Type:                tokentypes[i],
		}
		mapp.tokenKeeper.NewToken(ctx, tok)
	}
}
