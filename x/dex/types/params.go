package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/okex/okchain/x/common"
	"github.com/okex/okchain/x/params"
)

var (
	keyDexListFee             = []byte("DexListFee")
	keyTransferOwnershipFee   = []byte("TransferOwnershipFee")
	keyDelistMaxDepositPeriod = []byte("DelistMaxDepositPeriod")
	keyDelistMinDeposit       = []byte("DelistMinDeposit")
	keyDelistVotingPeriod     = []byte("DelistVotingPeriod")
	keyWithdrawPeriod         = []byte("WithdrawPeriod")
)

// Params defines param object
type Params struct {
	ListFee              sdk.DecCoin `json:"list_fee"`
	TransferOwnershipFee sdk.DecCoin `json:"transfer_ownership_fee"`
	//DelistFee            sdk.DecCoins `json:"delist_fee"`

	//  maximum period for okt holders to deposit on a dex delist proposal
	DelistMaxDepositPeriod time.Duration `json:"delist_max_deposit_period"`
	//  minimum deposit for a critical dex delist proposal to enter voting period
	DelistMinDeposit sdk.DecCoins `json:"delist_min_deposit"`
	//  length of the critical voting period for dex delist proposal
	DelistVotingPeriod time.Duration `json:"delist_voting_period"`

	WithdrawPeriod time.Duration `json:"withdraw_period"`
}

func tmpValidate(value interface{}) error {
	return nil
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{keyDexListFee, &p.ListFee, tmpValidate},
		{keyTransferOwnershipFee, &p.TransferOwnershipFee, tmpValidate},
		{keyDelistMaxDepositPeriod, &p.DelistMaxDepositPeriod, tmpValidate},
		{keyDelistMinDeposit, &p.DelistMinDeposit, tmpValidate},
		{keyDelistVotingPeriod, &p.DelistVotingPeriod, tmpValidate},
		{keyWithdrawPeriod, &p.WithdrawPeriod, tmpValidate},
	}
}

// ParamKeyTable for auth module
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns a default set of parameters.
func DefaultParams() *Params {
	defaultListFee := sdk.NewDecCoinFromDec(common.NativeToken, sdk.MustNewDecFromStr(defaultFeeList))
	defaultTransferOwnershipFee := sdk.NewDecCoinFromDec(common.NativeToken, sdk.MustNewDecFromStr(defaultFeeTransferOwnership))
	defaultDelistMinDeposit := sdk.NewDecCoinFromDec(common.NativeToken, sdk.MustNewDecFromStr(defaultDelistMinDeposit))
	return &Params{
		ListFee:                defaultListFee,
		TransferOwnershipFee:   defaultTransferOwnershipFee,
		DelistMaxDepositPeriod: time.Hour * 24,
		DelistMinDeposit:       sdk.DecCoins{defaultDelistMinDeposit},
		DelistVotingPeriod:     time.Hour * 72,
		WithdrawPeriod:         DefaultWithdrawPeriod,
	}
}

// String implements the stringer interface.
func (p Params) String() string {
	return fmt.Sprintf("Params: \nDexListFee:%s\nTransferOwnershipFee:%s\nDelistMaxDepositPeriod:%s\n"+
		"DelistMinDeposit:%s\nDelistVotingPeriod:%s\nWithdrawPeriod:%d\n",
		p.ListFee, p.TransferOwnershipFee, p.DelistMaxDepositPeriod, p.DelistMinDeposit, p.DelistVotingPeriod, p.WithdrawPeriod)
}
