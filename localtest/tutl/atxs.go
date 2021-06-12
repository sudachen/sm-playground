package tutl

import "sudachen.xyz/pkg/localnet/fu"

func (l *Localnet) ValidateAtxsForLayer(layer int64) bool {
	ok := true
	rs := query("atx failed contextual validation", l.Genesis(), layer)
	if len(rs) != 0 {
		ok = false
		for _, r := range rs {
			fu.Error("%v failed ATX contextual validation at layer %v", r.Name, layer)
		}
	}
	if layer % int64(l.LayersPerEpoch) == 0 {

	}
	return ok
}

func (l *Localnet) ValidateAtxs(last int64) bool{
	ok := true
	for layer := last; layer >= l.AfterGenesisLayer(); layer-- {
		ok = ok && l.ValidateAtxsForLayer(layer)
	}
	return ok
}
