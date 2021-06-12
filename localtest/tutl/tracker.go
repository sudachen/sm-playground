package tutl

import (
	"sudachen.xyz/pkg/errstr"
	"sudachen.xyz/pkg/localnet/fu"
	"sync"
)

type AccountTracker struct {
	credit map[Account]uint64
	debit  map[Account]uint64
	nonce  map[Account]uint64
	mu sync.Mutex
	DoNotPanic bool
}

func NewAccountTracker() *AccountTracker {
	return &AccountTracker{
		make(map[Account]uint64),
		make(map[Account]uint64),
		make(map[Account]uint64),
		sync.Mutex{},
		false,
	}
}

func (at *AccountTracker) Debit(pk Account, amount, nonce uint64) {
	at.mu.Lock()
	at.debit[pk] = amount
	at.nonce[pk] = nonce
	at.mu.Unlock()
}

func (at *AccountTracker) Transfer(from, to Account, amount uint64, fee uint64) {
	at.mu.Lock()
	at.debit[to] = at.debit[to] + amount
	at.credit[from] = at.credit[from] + amount + fee
	at.nonce[from] = at.nonce[from] + 1
	at.mu.Unlock()
}

func (at *AccountTracker) Balance(a Account) uint64 {
	at.mu.Lock()
	debit := at.debit[a]
	credit := at.credit[a]
	if debit < credit { panic(errstr.Errorf("negative balance for account %s", a.Address().Hex())) }
	at.mu.Unlock()
	return debit-credit
}

func (at *AccountTracker) Nonce(a Account) (nonce uint64) {
	at.mu.Lock()
	nonce = at.nonce[a]
	at.mu.Unlock()
	return
}

func (at *AccountTracker) Validate(bk Backends) {
	at.mu.Lock()
	defer at.mu.Unlock()
	if at.DoNotPanic {
		defer func() {
			if e, ok := recover().(errstr.ErrorStr); ok {
				fu.Error(e.Error())
			}
		}()
	}
	accs := map[Account]uint64{}
	for a := range at.debit {
		accs[a] = 0
	}
	for a, nonce := range at.nonce {
		accs[a] = nonce
	}
	for a, nonce := range accs {
		balance := at.debit[a] - at.credit[a]
		meshNonce := bk.Nonce(a)
		meshBalance := bk.Balance(a)
		fu.Verbose("%v => balance %v ?= %v, nonce %v ?= %v", a.Address().Hex(), balance, meshBalance, nonce, meshNonce)
		if meshNonce != nonce {
			panic(errstr.Errorf("account %v has invalied nonce %v, expected %v",a.Address().Hex(), meshNonce, nonce ))
		}
		if meshBalance !=  balance {
			panic(errstr.Errorf("account %v has invalied balance %v, expected %v",a.Address().Hex(), meshBalance, balance ))
		}
	}
}
