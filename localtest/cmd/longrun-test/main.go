package main

import (
	"sudachen.xyz/pkg/localnet/fu"
	"sudachen.xyz/pkg/localtest/tutl"
	"sudachen.xyz/pkg/localtest/tutl/actions"
	"time"
)

func main() {
	fu.VerboseOpt = true
	l := tutl.NewLocalNet()
	l.CpuPerNode = 2
	l.MemoryLimit = 1024*3 //Mb
	l.Count = 20
	l.LayerSize = l.Count
	l.LayerDuration = 15
	l.LayersPerEpoch = 3
	l.ReportPerf = false
	l.Debug = []string{}
	//l.Massif = []int{2,10}
	if ok, _ := l.AlreadyStarted(); ok {
		l.Stop()
	}
	l.Start()
	defer l.Stop()

	l.WaitForStart(60*time.Second)
	fu.Verbose("waiting for genesis %v", l.Genesis())
	l.WaitForGenesis()

	l.DoNotPanic = true
	for seed := 0; ; seed++ {
		actions.Transactions(l,seed)
		lastLayer := l.GetLastLayer()
		if !l.ValidateHare(lastLayer) {
			fu.Error("failed to validate hare")
		}
		if !l.ValidateAtxs(lastLayer) {
			fu.Error("failed to validate atxs")
		}
		if !l.ValidateLayers(lastLayer) {
			fu.Error("failed to validate layers")
		}
	}
}

