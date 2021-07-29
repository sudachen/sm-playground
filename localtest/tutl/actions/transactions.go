package actions

import (
	"math/rand"
	"sudachen.xyz/pkg/localnet/fu"
	"sudachen.xyz/pkg/localtest/tutl"
	"sync"
	"time"
)

var TapSk = tutl.Account("81c90dd832e18d1cf9758254327cb3135961af6688ac9c2a8c5d71f73acc5ce57be017a967db77fd10ac7c891b3d6d946dea7e3e14756e2f0f9e09b9663f0d9c")
const Amount1 = 10000
const Amount2 = 500
const Amount3 = 1
const Fee = 1
const NewAccountsCount = 10

func Transactions(l *tutl.Localnet, seed int) {

	bk := l.Backends()
	tracker := tutl.NewAccountTracker()
	tracker.DoNotPanic = l.DoNotPanic
	tracker.Debit(TapSk,bk.Balance(TapSk),bk.Nonce(TapSk))

	fu.Verbose("sequential transfer from tap to new accounts")

	wg := sync.WaitGroup{}
	acc := make([]tutl.Account, NewAccountsCount)
	xid := make([][]byte, NewAccountsCount)
	for i := range acc {
		acc[i] = tutl.GenSk()
		xid[i] = bk.SendCoins(TapSk, acc[i], Amount1, Fee, tracker)
		bk.WaitFor(xid[i])
	}

	for _, x := range xid {
		fu.Verbose("wait for transaction %x", x)
		bk.WaitFor(x, 2)
	}

	fu.Verbose("validating...")
	tracker.Validate(bk)

	fu.Verbose("parallel transfer from created accounts to new accounts")

	wg = sync.WaitGroup{}
	acc2 := make([]tutl.Account, NewAccountsCount)
	for i := range acc {
		acc2[i] = tutl.GenSk()
		wg.Add(1)
		go func(a,b tutl.Account) {
			defer l.HandlePanic()
			e := bk.Select(acc[i])
			xid := e.SendCoins(a, b, Amount2, Fee, tracker)
			bk.WaitFor(xid, 2)
			wg.Done()
		}(acc[i],acc2[i])
	}

	wg.Wait()
	fu.Verbose("validating...")
	tracker.Validate(bk)

	fu.Verbose("random transfer from created accounts to new or existing accounts")

	acc2 = append(acc2,acc...)
	wg = sync.WaitGroup{}
	for i := range acc2 {
		wg.Add(1)
		go func(i int) {
			defer l.HandlePanic()
			var xid []byte
			r := rand.New(rand.NewSource(int64(i+seed)))
			for _ = range acc2 {
				j := r.Intn(len(acc2))
				a := acc2[j]
				if j == i {
					a = tutl.GenSk()
				}
				b := bk.Select(a)
				if len(xid) != 0 {
					b.WaitFor(xid, 1)
				}
				xid = b.SendCoins(acc2[i], a, Amount3, Fee, tracker)
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
	time.Sleep(2*time.Duration(l.LayerDuration)*time.Second)
	fu.Verbose("validating...")
	tracker.Validate(bk)

	// final validation
}

