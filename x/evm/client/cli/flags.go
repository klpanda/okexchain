package cli

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	flagCodeFile     = "code_file"
	flagAmount       = "amount"
	flagArgs         = "args"
	flagMethod       = "method"
	flagContractAddr = "contract_addr"
	flagAbiFile      = "abi_file"
	flagShowCode     = "show_code"
	flagAll          = "all"
)

const (
	DefaultAmount = "0" + sdk.DefaultBondDenom
)
