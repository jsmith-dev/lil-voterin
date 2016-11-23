package types

import (
	"bytes"

	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
)

//------------------------------------------------------
// pubkey is ed25519

type (
	PubKey    crypto.PubKeyEd25519
	Signature crypto.SignatureEd25519
)

// Return the pubkey as a byte array
func (p PubKey) Bytes() []byte {
	return crypto.PubKeyEd25519(p).Bytes()
}

// Verify the pubkey signed the msg
func (p PubKey) VerifyBytes(msg []byte, sig Signature) bool {
	return crypto.PubKeyEd25519(p).VerifyBytes(msg, crypto.SignatureEd25519(sig))
}

//------------------------------------------------------
// database keys for accessing accounts

func AccountKeyBytes(pubKey PubKey) []byte {
	return wire.BinaryBytes(pubKey)
}

func AccountKeyString(pubKey PubKey) string {
	return string(AccountKeyBytes(pubKey))
}

func BytesToAccountKey(pubKeyBytes []byte) PubKey {
	var pubKey PubKey
	err := wire.ReadBinaryBytes(pubKeyBytes, &pubKey)
	if err != nil {
		panic(err)
	}
	return pubKey
}

//---------------------------------------
// Account types

type AccountType uint

const (
	AccountTypeVoter AccountType = 1 + iota
	AccountTypeAdmin

	AccountTypeCorrupt = 100
)

type Account struct {
	Sequence int         `json:"sequence"` // number of transactions committed
	Type     AccountType `json:"type"`     // type for capabilities
}

func (acc *Account) Marshal() []byte {
	return wire.BinaryBytes(acc)
}

func (acc *Account) Unmarshal(b []byte) error {
	r, n, err := bytes.NewBuffer(b), new(int), new(error)
	wire.ReadBinary(acc, r, 0, n, err)
	return *err
}

func NewAccount(typ AccountType) (crypto.PrivKeyEd25519, PubKey, *Account) {
	priv := crypto.GenPrivKeyEd25519()
	return priv,
		PubKey(priv.PubKey().(crypto.PubKeyEd25519)),
		&Account{Type: typ}
}

func PubFromPriv(privKey crypto.PrivKey) PubKey {
	return PubKey(privKey.PubKey().(crypto.PubKeyEd25519))
}

//------------------------------------------

type PubAccount struct {
	PubKey   `json:"pubkey"`
	*Account `json:"account"`
}

type GenesisState struct {
	Accounts []PubAccount `json:"accounts"`
}
