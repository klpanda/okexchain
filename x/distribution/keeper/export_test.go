package keeper

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeeper(t *testing.T) {
	ctx, _, k, _, _ := CreateTestInputDefault(t, false, 1000)
	require.NotNil(t, ctx)
	require.Equal(t, authtypes.FeeCollectorName, k.GetFeeCollectorName())
}
