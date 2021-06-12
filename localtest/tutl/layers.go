package tutl

import (
	"strings"
	"sudachen.xyz/pkg/localnet/fu"
)

func (l *Localnet) ValidateOneLayer(layer int64) bool {
	ok := true
	genesis := l.Genesis()

	rs := query("new layer hash", genesis, layer)
	hs := map[string][]string{}
	for _, r := range rs {
		hs[r.M.LayerHash] = append(hs[r.M.LayerHash],r.Name)
	}
	if _, exists := hs[""]; exists {
		fu.Error("nil layer hash found for layer %v", layer)
		ok = false
	}
	if len(hs) != 1 {
		fu.Error("inconsistent layer hash found, there are %d hashes for layer %v", len(hs), layer)
		for k,v := range hs {
			fu.Error("%v\n  => %v", strings.Join(v,", "), k)
		}
		ok = false
	}

	return ok
}

func (l *Localnet) ValidateLayers(last int64) bool{
	ok := true
	for layer := last; layer >= l.AfterGenesisLayer(); layer-- {
		ok = ok && l.ValidateOneLayer(layer)
	}
	return ok
}
