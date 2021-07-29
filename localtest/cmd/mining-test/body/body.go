package body

import (
	"sudachen.xyz/pkg/errstr"
	"sudachen.xyz/pkg/localnet"
	"sudachen.xyz/pkg/localnet/fu"
	"sudachen.xyz/pkg/localtest/tutl"
	"sudachen.xyz/pkg/localtest/tutl/actions"
	"time"
)

func DoIt(rev localnet.Revision) {
	fu.VerboseOpt = true
	l := tutl.NewLocalNet()
	l.Count = 12
	l.LayerDuration = 30
	l.LayersPerEpoch = 3
	l.ReportPerf = false
	//l.Massif = []int{2,5}
	l.CpuPerNode = 4
	l.Rev = rev
	l.P2pRandCon = 8

	if ok, _ := l.AlreadyStarted(); ok {
		l.Stop()
	}
	l.Start()
	defer l.Stop()

	l.WaitForStart(time.Duration(fu.Maxi(l.LayerDuration*2,l.Count*5))*time.Second)
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


