package evm

import (
	"fmt"
	"math/big"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/okex/okchain/x/evm/types"
)

// StateTransition defines data to transitionDB in vm
type StateTransition struct {
	Sender    sdk.AccAddress
	Recipient sdk.AccAddress
	Amount    sdk.Int
	Payload   []byte
	StateDB   *types.CommitStateDB
}

func (st StateTransition) CanTransfer(acc sdk.AccAddress, amount *big.Int) bool {
	return st.StateDB.GetBalance(acc).Cmp(amount) >= 0
}

func (st StateTransition) Transfer(from, to sdk.AccAddress, amount *big.Int) {
	st.StateDB.SubBalance(from, amount)
	st.StateDB.AddBalance(to, amount)
}

func (st StateTransition) GetHashFn(header abci.Header) func() sdk.Hash {
	return func() sdk.Hash {
		var res = sdk.Hash{}
		blockID := header.GetLastBlockId()
		res.SetBytes(blockID.GetHash())
		return res
	}
}

func (st StateTransition) TransitionCSDB(ctx sdk.Context, k Keeper) (*big.Int, *sdk.Result, sdk.Error) {
	ctx = ctx.WithLogger(ctx.Logger().With("module", fmt.Sprintf("modules/%s", types.ModuleName)))

	evmCtx := Context{
		CanTransfer: st.CanTransfer,
		Transfer:    st.Transfer,
		GetHash:     st.GetHashFn(ctx.BlockHeader()),
		Origin:      st.Sender,
		CoinBase:    ctx.BlockHeader().ProposerAddress,
		Time:        sdk.NewInt(ctx.BlockHeader().Time.Unix()).BigInt(),
		BlockNumber: sdk.NewInt(ctx.BlockHeader().Height).BigInt(),
	}

	gasLimitForVm := uint64(DefaultVmGasLimit)
	if !ctx.Simulate {
		gasLimitForVm = ctx.GasMeter().Limit() - ctx.GasMeter().GasConsumed()
	}
	evmCtx.GasLimit = gasLimitForVm

	curGasMeter := ctx.GasMeter()
	gasMeterForEvm := sdk.NewInfiniteGasMeter()

	vmParams := k.GetParams(ctx) // will consume gas
	st.StateDB.UpdateAccounts()  // will consume gas

	cfg := Config{OpConstGasConfig: &vmParams.VMOpGasParams, CommonGasConfig: &vmParams.VMCommonGasParams}
	evm := NewEVM(evmCtx, st.StateDB.WithContext(ctx.WithGasMeter(gasMeterForEvm)), cfg)

	var (
		ret         []byte
		leftOverGas uint64
		addr        sdk.AccAddress
		vmerr       sdk.Error
	)

	if st.Recipient.Empty() {
		ret, addr, leftOverGas, vmerr = evm.Create(st.Sender, st.Payload, gasLimitForVm, st.Amount.BigInt())
		ctx.Logger().Info(fmt.Sprintf("create contract, consumed gas = %v, leftOverGas = %v, vm err = %v ", gasLimitForVm-leftOverGas, leftOverGas, vmerr))
	} else {
		ret, leftOverGas, vmerr = evm.Call(st.Sender, st.Recipient, st.Payload, gasLimitForVm, st.Amount.BigInt())
		if vmerr == ErrExecutionReverted() {
			reason := "null"
			if len(ret) > 4 {
				reason = string(ret[4:])
			}
			ctx.Logger().Info(fmt.Sprintf("VM revert error, reason provided by the contract: %s", reason))
		}

		ctx.Logger().Info(fmt.Sprintf("call contract, ret = %x, consumed gas = %v, leftOverGas = %v, vm err = %v", ret, gasLimitForVm-leftOverGas, leftOverGas, vmerr))
	}

	vmGasUsed := gasLimitForVm - leftOverGas

	if vmerr != nil {
		return nil, &sdk.Result{Data: ret, GasUsed: curGasMeter.GasConsumed() + vmGasUsed}, vmerr
	}

	st.StateDB.Finalise(true)

	// comsume vm gas
	ctx.WithGasMeter(curGasMeter).GasMeter().ConsumeGas(vmGasUsed, "VM execution consumption")

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeNewContract,
			sdk.NewAttribute(types.AttributeKeyAddress, addr.String()),
		),
	})

	return nil, &sdk.Result{Data: ret, GasUsed: ctx.GasMeter().GasConsumed()}, nil
}

func DoStateTransition(ctx sdk.Context, msg types.MsgContract, k Keeper, readonly bool) (*big.Int, *sdk.Result, sdk.Error) {
	st := StateTransition{
		Sender:    msg.From,
		Recipient: msg.To,
		Payload:   msg.Payload,
		Amount:    msg.Amount.Amount.TruncateInt(),
		StateDB:   k.StateDB.WithContext(ctx).WithTxHash(tmhash.Sum(ctx.TxBytes())),
	}

	if readonly {
		ctx.Simulate = true
	}

	if ctx.Simulate == false && ctx.GasMeter().Limit() == 0 {
		return nil, &sdk.Result{Data: nil}, ErrWrongCtx()
	}

	if ctx.Simulate {
		st.StateDB = types.NewStateDB(k.StateDB).WithContext(ctx)
	}

	return st.TransitionCSDB(ctx, k)
}
