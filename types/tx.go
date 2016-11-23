package types

import (
	. "github.com/tendermint/go-common"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
	tmsp "github.com/tendermint/tmsp/types"
)

// XXX: SignBytes must be unique for each tx type
// or adapted to include the type-byte

//---------------------------------------
// tx types

const maxTxNonceSize = 32

const (
	txTypeVote = 1 + iota
	txTypeAdmin
	txTypeFork
)

type Tx interface {
	SignBytes() []byte
	Validate() tmsp.Result

	// for testing
	Sign(crypto.PrivKey)
}

var _ = wire.RegisterInterface(
	struct{ Tx }{},
	wire.ConcreteType{&VoteTx{}, txTypeVote},
	wire.ConcreteType{&AdminTx{}, txTypeAdmin},
	wire.ConcreteType{&ForkTx{}, txTypeFork},
)

func JSONBytes(tx Tx) []byte {
	return wire.JSONBytes(struct {
		Tx `json:"unwrap"`
	}{tx})
}

//---------------------------------------
// Vote Tx

type VoteTx struct {
	Ballots []Ballot `json:"ballots"`

	Nonce     []byte    `json:"nonce"`
	PubKey    PubKey    `json:"pubkey,omitempty"` // TODO: replace with AccountIndex
	Signature Signature `json:"signature,omitempty"`
}

// Sign bytes is canonical json encoded tx without the signature
func (tx *VoteTx) SignBytes() []byte {
	return wire.JSONBytes(struct {
		Ballots []Ballot `json:"ballots"`
		Nonce   []byte   `json:"nonce"`
		Pubkey  PubKey   `json:"pubkey"`
	}{
		tx.Ballots,
		tx.Nonce,
		tx.PubKey,
	})
}

func (tx *VoteTx) Validate() tmsp.Result {
	// NOTE
	// pubkey length is enforced by type;
	// tx byte length is enforced by maxTxSize;
	// ballot is checked later in AddBallot

	if len(tx.Nonce) > maxTxNonceSize {
		return tmsp.ErrBadNonce.AppendLog(Fmt("Nonce too big (%d). Max is %d", len(tx.Nonce), maxTxNonceSize))
	}

	// verify sig
	if !tx.PubKey.VerifyBytes(tx.SignBytes(), tx.Signature) {
		return tmsp.ErrUnauthorized.AppendLog("Invalid signature")
	}
	return tmsp.OK
}

// Sign transaction. For testing
func (tx *VoteTx) Sign(priv crypto.PrivKey) {
	tx.Signature = Signature(priv.Sign(tx.SignBytes()).(crypto.SignatureEd25519))
}

//---------------------------------------
// Admin Tx

type AdminTx struct {
	PubAccounts []PubAccount `json:"pub_accounts"`

	Nonce     []byte    `json:"nonce"`
	PubKey    PubKey    `json:"pubkey,omitempty"` // TODO: replace with AccountIndex
	Signature Signature `json:"signature,omitempty"`
}

func (tx *AdminTx) SignBytes() []byte {
	return wire.JSONBytes(struct {
		Nonce       []byte       `json:"nonce"`
		PubAccounts []PubAccount `json:"pub_accounts"`
		Pubkey      PubKey       `json:"pubkey"`
	}{
		tx.Nonce,
		tx.PubAccounts,
		tx.PubKey,
	})
}

func (tx *AdminTx) Validate() tmsp.Result {
	// NOTE
	// pubkey length is enforced by type;
	// tx byte length is enforced by maxTxSize;

	if len(tx.Nonce) > maxTxNonceSize {
		return tmsp.ErrBadNonce.AppendLog(Fmt("Nonce too big (%d). Max is %d", len(tx.Nonce), maxTxNonceSize))
	}

	// verify sig
	if !tx.PubKey.VerifyBytes(tx.SignBytes(), tx.Signature) {
		return tmsp.ErrUnauthorized.AppendLog("Invalid signature")
	}
	return tmsp.OK
}

// Sign transaction. For testing
func (tx *AdminTx) Sign(priv crypto.PrivKey) {
	tx.Signature = Signature(priv.Sign(tx.SignBytes()).(crypto.SignatureEd25519))
}

//---------------------------------------
// Fork Tx

type ForkTx struct {
	Name      string    `json:"name"`
	Nonce     []byte    `json:"nonce"`
	PubKey    PubKey    `json:"pubkey,omitempty"` // TODO: replace with AccountIndex
	Signature Signature `json:"signature,omitempty"`
}

func (tx *ForkTx) SignBytes() []byte {
	return wire.JSONBytes(struct {
		Name   string `json:"name"`
		Nonce  []byte `json:"nonce"`
		Pubkey PubKey `json:"pubkey"`
	}{
		tx.Name,
		tx.Nonce,
		tx.PubKey,
	})
}

func (tx *ForkTx) Validate() tmsp.Result {
	// NOTE
	// pubkey length is enforced by type;
	// tx byte length is enforced by maxTxSize;

	if len(tx.Nonce) > maxTxNonceSize {
		return tmsp.ErrBadNonce.AppendLog(Fmt("Nonce too big (%d). Max is %d", len(tx.Nonce), maxTxNonceSize))
	}

	// verify sig
	if !tx.PubKey.VerifyBytes(tx.SignBytes(), tx.Signature) {
		return tmsp.ErrUnauthorized.AppendLog("Invalid signature")
	}
	return tmsp.OK
}

// Sign transaction. For testing
func (tx *ForkTx) Sign(priv crypto.PrivKey) {
	tx.Signature = Signature(priv.Sign(tx.SignBytes()).(crypto.SignatureEd25519))
}
