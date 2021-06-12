package tutl

import (
	"encoding/binary"
	"fmt"
	"strings"
	"sudachen.xyz/pkg/errstr"
	"sudachen.xyz/pkg/localnet"
	"sudachen.xyz/pkg/localnet/client"
	"sudachen.xyz/pkg/localnet/fu"
	"sudachen.xyz/pkg/localtest/tutl/queries"
	"time"
)

type Localnet struct {
	*localnet.Localnet
	DoNotPanic bool
	except map[int]bool
}

func NewLocalNet() *Localnet {
	l := &Localnet{ localnet.New(), false, map[int]bool{} }
	for _, i := range l.Massif {
		l.except[i] =  true
	}
	return l
}

func (l *Localnet) HandlePanic() {
	if l.DoNotPanic {
		if e := recover(); e != nil {
			if x, ok := e.(errstr.ErrorStr); ok {
				fu.Error(x.String())
				return
			} else {
				panic(e)
			}
		}
	}
}

func (l *Localnet) WaitForStart(timeout time.Duration) {
	t1 := time.NewTicker(time.Second)
	defer t1.Stop()
	t2 := time.NewTimer(timeout)
	defer t2.Stop()
	for len(queries.GetAppStartedMsgs(l.Genesis())) != l.Count {
		select {
		case <- t1.C:
		case <- t2.C:
			panic(errstr.Errorf("out of time on WaitForStart"))
		}
	}
}

func (l *Localnet) WaitForLayer(layer int64, timeout time.Duration, count ...int) (r []queries.Row){
	t1 := time.NewTicker(time.Second)
	defer t1.Stop()
	t2 := time.NewTimer(timeout)
	defer t2.Stop()

	for {
		r = queries.GetTickMsgs(l.Genesis(), layer);
		if len(r) == fu.Fnzi(fu.Fnzi(count...),l.Count) {
			return
		}
		select {
		case <- t1.C:
		case <- t2.C:
			panic(errstr.Errorf("out of time on WaitForLayer"))
		}
	}
}

func (l *Localnet) GetLastLayer() int64 {
	defer l.HandlePanic()
	return queries.GetLastFinishedLayer(l.Genesis(),l.Count)
}

func (l *Localnet) WaitForGenesis() []queries.Row {
	timeout := l.GenesisTime().Sub(time.Now()) +
		time.Duration(l.AfterGenesisLayer() + 1) * time.Duration(l.LayerDuration) * time.Second
	return l.WaitForLayer(l.AfterGenesisLayer(), timeout)
}

func (l *Localnet) AfterGenesisLayer() int64 {
	return int64(l.LayersPerEpoch*2)
}

type Backend struct {
	*client.Backend
	localnet.MinerInfo
}

func (l *Localnet) Connect(nfo localnet.MinerInfo) *Backend {
	 bx, err := client.OpenConnection(fmt.Sprintf("%s:%d",nfo.Ip,l.Grpc))
	 if err != nil {
	 	panic(err)
	 }
	 return &Backend{bx, nfo}
}

type Backends []*Backend

func (l *Localnet) List() []localnet.MinerInfo {
	lst, err := l.Localnet.ListMiners()
	if err != nil {
		panic(err)
	}
	return lst
}

func (l *Localnet) Backends() (bk Backends) {
	for _, n := range l.List() {
		if !l.except[n.Number] {
			bk = append(bk,l.Connect(n))
		}
	}
	return
}

func (bk Backends) Select(a Account) *Backend {
	pk := a.Pk()
	i := int(binary.LittleEndian.Uint64(pk)%uint64(len(bk)))
	return bk[i]
}

func (bk Backends) Balance(sk Account) uint64 {
	b := bk.Select(sk)
	ac, err := b.AccountState(sk.Address())
	if err != nil {
		panic(err)
	}
	return ac.StateCurrent.Balance.Value
}

func (bk Backends) Nonce(sk Account) uint64 {
	b := bk.Select(sk)
	ac, err := b.AccountState(sk.Address())
	if err != nil {
		panic(err)
	}
	return ac.StateCurrent.Counter
}

func (b Backend) SendCoins(from, to Account, amount, fee uint64, tracker *AccountTracker) []byte {
	nonce := tracker.Nonce(from)
	fu.Verbose("transfer[%s] %s -> %s, %v, fee %v, nonce %v", b.Name, from.Address().Hex(), to.Address().Hex(), amount, fee, nonce)
	for {
		st, err := b.Transfer(to.Address(), nonce, amount, fee, 1, from.Sk())
		if err != nil {
			if strings.Contains(err.Error(), "try again later") {
				fu.Verbose("node %s is not synced yet: %v", b.Name, err.Error())
				time.Sleep(2*time.Second)
				continue
			}
			panic(errstr.Wrapf(0, err, "failed to transfer[%s] %s -> %s: %v", b.Name, from.Address().Hex(), to.Address().Hex(), err.Error()))
		}
		tracker.Transfer(from, to, amount, fee)
		return st.Id.Id
	}
}

func (bk Backends) SendCoins(from, to Account, amount, fee uint64, tracker *AccountTracker) []byte {
	return bk.Select(to).SendCoins(from,to,amount,fee,tracker)
}

func (b Backend) WaitFor(txid []byte, strict ...int) {
	s := fu.Fnzi(strict...)
loop:
	for {
		r, _, err := b.TransactionState(txid, true)
		if err != nil {
			fu.Error("TX => %s", b.Name, err.Error())
		} else {
			switch r.State.String() {
			case "TRANSACTION_STATE_PROCESSED":
				//fu.Verbose("TX => %s: %v", b.Name, r.State.String())
				break loop
			case "TRANSACTION_STATE_MESH":
				if s < 2 {
					//fu.Verbose("TX => %s: %v", b.Name, r.State.String())
					break loop
				}
			case "TRANSACTION_STATE_MEMPOOL":
				if s == 0 {
					//fu.Verbose("TX => %s: %v", b.Name, r.State.String())
					break loop
				}
			case "TRANSACTION_STATE_UNSPECIFIED":
				// nothing
			default:
				fu.Error("TX => %s: %v", b.Name, r.State.String())
				return
			}
		}
		time.Sleep(time.Second/2)
	}
}

func (bk Backends) WaitFor(txid []byte, strict ...int) {
	s := fu.Fnzi(strict...)
	ok := make([]bool, len(bk))
	for repeat := true; repeat; time.Sleep(time.Second/2) {
		repeat = false
		for i, b := range bk {
			if !ok[i] {
				r, _, err := b.TransactionState(txid, true)
				if err != nil {
					fu.Error("TX => %s", b.Name, err.Error())
					repeat = true
				} else {
					switch r.State.String() {
					case "TRANSACTION_STATE_PROCESSED":
						//fu.Verbose("TX => %s: %v", b.Name, r.State.String())
						ok[i] = true
					case "TRANSACTION_STATE_MESH":
						if s < 2 {
							//fu.Verbose("TX => %s: %v", b.Name, r.State.String())
							ok[i] = true
 						} else {
 							repeat = true
						}
					case "TRANSACTION_STATE_MEMPOOL":
						if s == 0 {
							//fu.Verbose("TX => %s: %v", b.Name, r.State.String())
							ok[i] = true
						} else {
							repeat = true
						}
					case "TRANSACTION_STATE_UNSPECIFIED":
						repeat = true
					default:
						fu.Error("TX => %s: %v", b.Name, r.State.String())
						return
					}
				}
			}
		}
	}
}

