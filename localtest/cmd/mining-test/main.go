package main

import (
	"sudachen.xyz/pkg/errstr"
	"sudachen.xyz/pkg/localnet/fu"
	"sudachen.xyz/pkg/localtest/tutl"
	"sudachen.xyz/pkg/localtest/tutl/actions"
	"time"
)

func main() {
	fu.VerboseOpt = true
	l := tutl.NewLocalNet()
	l.Count = 12
	l.LayerDuration = 15
	l.LayersPerEpoch = 3
	l.ReportPerf = false
	l.Massif = []int{2,5}
	l.CpuPerNode = 2

	if ok, _ := l.AlreadyStarted(); ok {
		l.Stop()
	}
	l.Start()
	defer l.Stop()

	l.WaitForStart(60*time.Second)
	fu.Verbose("waiting for genesis %v", l.Genesis())
	l.WaitForGenesis()

	actions.Transactions(l,0)

	lastLayer := l.GetLastLayer()
	if !l.ValidateHare(lastLayer) {
		panic(errstr.Errorf("failed to validate hare"))
	}
	if !l.ValidateAtxs(lastLayer) {
		panic(errstr.Errorf("failed to validate atxs"))
	}
	if !l.ValidateLayers(lastLayer) {
		panic(errstr.Errorf("failed to validate layers"))
	}
}

