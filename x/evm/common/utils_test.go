package common

import (
	"bytes"
	"testing"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestCreateAddress2(t *testing.T) {
	type testcase struct {
		origin   string
		salt     string
		code     string
		expected string
	}

	for i, tt := range []testcase{
		{
			origin:   "okchain16msmkmaqjhcwxcs9ffwd0eprrawq6txe7y5h53",
			salt:     "0x0000000000000000000000000000000000000000",
			code:     "0x00",
			expected: "okchain1uv6zeqthgwn3qts0tvn958ke5karsxt2zpm8dr",
		},
		{
			origin:   "okchain1g4lvreq7c20sq7p6nphsp9qw29a2a3q40favg2",
			salt:     "0x0000000000000000000000000000000000000000",
			code:     "0x00",
			expected: "okchain1u9mvmvgsu0k25a4957ng4c7dpqtv5g8tqx06fg",
		},
		{
			origin:   "okchain1kdsvks20n9dk05l8qv77qwjgnxphn96k4jhu67",
			salt:     "0xfeed000000000000000000000000000000000000",
			code:     "0x00",
			expected: "okchain1x8ujmx4uh0ujwsalgrqy89ypp4pvslrtclmg8x",
		},
		{
			origin:   "okchain16msmkmaqjhcwxcs9ffwd0eprrawq6txe7y5h53",
			salt:     "0x0000000000000000000000000000000000000000",
			code:     "0xdeadbeef",
			expected: "okchain1xw0uetrfp45rdpml4ypypc8g5s9ygg972xq2hw",
		},
		{
			origin:   "okchain16msmkmaqjhcwxcs9ffwd0eprrawq6txe7y5h53",
			salt:     "0xcafebabe",
			code:     "0xdeadbeef",
			expected: "okchain1rjdyk4weuunvas60a854tmm9hmrmn85cy3a3q0",
		},
		{
			origin:   "okchain16msmkmaqjhcwxcs9ffwd0eprrawq6txe7y5h53",
			salt:     "0xcafebabe",
			code:     "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			expected: "okchain1rg0xh4yk25fldnntnr66xwtpf8p8l8sagsuphn",
		},
		{
			origin:   "okchain16msmkmaqjhcwxcs9ffwd0eprrawq6txe7y5h53",
			salt:     "0x0000000000000000000000000000000000000000",
			code:     "0x",
			expected: "okchain1pk7p843du7ds4jdmjk7fq2caamteseencjyk00",
		},
	} {

		origin, _ := sdk.AccAddressFromBech32(tt.origin)
		salt := sdk.BytesToHash(FromHex(tt.salt))
		codeHash := crypto.Sha256(FromHex(tt.code))
		address := CreateAddress2(origin, salt, codeHash)

		expected, _ := sdk.AccAddressFromBech32(tt.expected)
		if !bytes.Equal(expected.Bytes(), address.Bytes()) {
			t.Errorf("test %d: expected %s, got %s", i, expected.String(), address.String())
		}

	}
}
