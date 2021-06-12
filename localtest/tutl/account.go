package tutl

import (
	"github.com/spacemeshos/ed25519"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/common/util"
)

type Account string

func GenSk() Account {
	_, sk, err := ed25519.GenerateKey(nil)
	//fu.Verbose("new sk %v, pk %v", util.Bytes2Hex(sk), Account(util.Bytes2Hex(sk)).Pk())
	if err != nil {
		panic(err)
	}
	return Account(util.Bytes2Hex(sk))
}

func (a Account) Address() types.Address {
	return types.HexToAddress(string(a[64:]))
}

func (a Account) Key() string {
	return string(a[64:64+40])
}

func (a Account) Pk() []byte {
	return util.Hex2Bytes(string(a[64:]))
}

func (a Account) Sk() []byte {
	return util.Hex2Bytes(string(a))
}
