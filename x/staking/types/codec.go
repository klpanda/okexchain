package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
)

// RegisterCodec registers concrete types for codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(&MsgCreateValidator{}, "okchain/staking/MsgCreateValidator", nil)
	cdc.RegisterConcrete(&MsgEditValidator{}, "okchain/staking/MsgEditValidator", nil)
	cdc.RegisterConcrete(&MsgDestroyValidator{}, "okchain/staking/MsgDestroyValidator", nil)
	cdc.RegisterConcrete(&MsgDeposit{}, "okchain/staking/MsgDeposit", nil)
	cdc.RegisterConcrete(&MsgWithdraw{}, "okchain/staking/MsgWithdraw", nil)
	cdc.RegisterConcrete(&MsgAddShares{}, "okchain/staking/MsgAddShares", nil)
	cdc.RegisterConcrete(&MsgRegProxy{}, "okchain/staking/MsgRegProxy", nil)
	cdc.RegisterConcrete(&MsgBindProxy{}, "okchain/staking/MsgBindProxy", nil)
	cdc.RegisterConcrete(&MsgUnbindProxy{}, "okchain/staking/MsgUnbindProxy", nil)
}

var (
	amino = codec.New()

	// ModuleCdc references the global x/staking module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/staking and
	// defined at the application level.
	ModuleCdc = codec.NewHybridCodec(amino, types.NewInterfaceRegistry())
)

func init() {
	RegisterCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
