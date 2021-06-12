package client

import (
	"bytes"
	"sudachen.xyz/pkg/localnet/fu"
	"time"

	xdr "github.com/davecgh/go-xdr/xdr2"
	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/ed25519"
	"github.com/spacemeshos/go-spacemesh/common/types"
)

// Backend wallet holder
type Backend struct {
	*gRPCClient      // Embedded interface
}

func friendlyTime(nastyString string) string {
	t, err := time.Parse("2006-01-02T15-04-05.000Z", nastyString)
	if err != nil {
		return nastyString
	}
	return t.Format("Jan 02 2006 03:04 PM")
}


// OpenConnection opens a connection but not the wallet
func OpenConnection(grpcServer string) (wbx *Backend, err error) {
	wbe := &Backend{}
	wbe.gRPCClient = newGRPCClient(grpcServer, false)
	if err = wbe.gRPCClient.Connect(); err != nil {
		fu.Error("failed to connect to the grpc server: %s", err)
		return
	}
	return wbe, nil
}

func interfaceToBytes(i interface{}) ([]byte, error) {
	var w bytes.Buffer
	if _, err := xdr.Marshal(&w, &i); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// Transfer creates a sign coin transaction and submits it
func (w *Backend) Transfer(recipient types.Address, nonce, amount, gasPrice, gasLimit uint64, key ed25519.PrivateKey) (*pb.TransactionState, error) {
	tx := SerializableSignedTransaction{}
	tx.AccountNonce = nonce
	tx.Amount = amount
	tx.Recipient = recipient
	tx.GasLimit = gasLimit
	tx.Price = gasPrice

	buf, _ := interfaceToBytes(&tx.InnerSerializableSignedTransaction)
	copy(tx.Signature[:], ed25519.Sign2(key, buf))
	b, err := interfaceToBytes(&tx)
	if err != nil {
		return nil, err
	}
	return w.SubmitCoinTransaction(b)
}

type InnerSerializableSignedTransaction struct {
	AccountNonce uint64
	Recipient    types.Address
	GasLimit     uint64
	Price        uint64
	Amount       uint64
}

type SerializableSignedTransaction struct {
	InnerSerializableSignedTransaction
	Signature [64]byte
}
