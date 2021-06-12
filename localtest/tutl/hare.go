package tutl

import (
	"fmt"
	"sort"
	"strings"
	"sudachen.xyz/pkg/localnet/fu"
	"sudachen.xyz/pkg/localtest/tutl/queries"
)

func query(msg string, genesis string, layer int64, opts ...map[string]string) []queries.Row {
	if layer != 0 {
		return queries.GetAllMsgs(genesis, msg,
			fu.MixMaps(
				map[string]string{
					"m.layer_id": fmt.Sprint(layer),
				},
				opts...),
		)
	} else {
		return queries.GetAllMsgs(genesis, msg)
	}
}

func (l *Localnet) ValidateHareForLayer(layer int64) bool {
	defer l.HandlePanic()

	ok := true
	genesis := l.Genesis()

	fu.Verbose("validating hare for layer %v ...",layer)

	ns := map[string][]string{}
	report := func() {
		if len(ns) != 0 {
			for k,v := range ns {
				fu.Error("%v => %v", k, strings.Join(v,"; "))
			}
			ok = false
		}
	}

	for _, r := range query("preround ended", genesis, layer) {
		if strings.Contains(r.Text, "with empty set") {
			ns[r.Name] = append(ns[r.Name],r.Text)
		}
	}
	report()

	ns = map[string][]string{}
	for _, r := range query("status round ended", genesis, layer) {
		if !r.M.IsRsvpReady {
			ns[r.Name] = append(ns[r.Name], "rsvp is not ready")
		}
	}
	report()

	ns = map[string][]string{}
	for _, r := range query("proposal round ended", genesis, layer) {
		if r.M.ProposedSet == "" {
			ns[r.Name] = append(ns[r.Name], "has no proposal")
		}
	}
	report()

	ns = map[string][]string{}
	var eligibilityCount int64
	for _, r := range query("message sent", genesis, layer, map[string]string{"m.msg_type": "Commit"}) {
		eligibilityCount += r.M.EligibilityCount
	}
	if eligibilityCount < l.MinEligibility() {
		fu.Error("layer commits have eligibility %v, expected at least %v", eligibilityCount, l.MinEligibility())
		ok = false
	}

	rs := query("consensus process terminated", genesis, layer)
	hs := map[string][]string{}
	for _, r := range rs {
		x := strings.Split(r.M.CurrentSet, ",")
		sort.Strings(x)
		k := strings.Join(x,",")
		hs[k] = append(hs[k],r.Name)
	}
	if _, exists := hs[""]; exists {
		fu.Error("nil hare result found")
		ok = false
	}
	if len(hs) != 1 {
		fu.Error("inconsistent hare result found")
		for k,v := range hs {
			fu.Error("%v\n  => %v", strings.Join(v,", "), k)
		}
		ok = false
	}

	return ok
}

func (l *Localnet) ValidateHare(last int64) bool{
	defer l.HandlePanic()

	ok := true
	for layer := last; layer >= l.AfterGenesisLayer(); layer-- {
		ok = ok && l.ValidateHareForLayer(layer)
	}
	return ok
}
