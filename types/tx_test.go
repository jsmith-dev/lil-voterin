package types

import (
	"fmt"
	"testing"

	"github.com/tendermint/go-wire"
)

func TestVoteTxSign(t *testing.T) {
	priv, pub, _ := NewAccount(AccountTypeVoter)
	b1, b2 := MakeTestBallots()
	tx := &VoteTx{
		Ballots: []Ballot{b1, b2},
		Nonce:   []byte{1, 2, 3},
		PubKey:  pub,
	}
	fmt.Printf("%X\n", priv.Bytes())
	fmt.Println(string(tx.SignBytes()))
	tx.Sign(priv)
	fmt.Println(string(wire.JSONBytes(struct {
		Tx `json:"unwrap"`
	}{tx})))

	r := tx.Validate()
	if !r.IsOK() {
		t.Fatal(r)
	}
}

func TestAdminTxSign(t *testing.T) {
	priv, pub, _ := NewAccount(AccountTypeVoter)
	_, pub1, _ := NewAccount(AccountTypeVoter)
	_, pub2, _ := NewAccount(AccountTypeVoter)
	tx := &AdminTx{
		PubAccounts: []PubAccount{
			PubAccount{
				PubKey:  pub1,
				Account: &Account{Type: AccountTypeVoter},
			},
			PubAccount{
				PubKey:  pub2,
				Account: &Account{Type: AccountTypeVoter},
			},
		},
		Nonce:  []byte{1, 2, 3},
		PubKey: pub,
	}
	fmt.Printf("%X\n", priv.Bytes())
	fmt.Println(string(tx.SignBytes()))
	tx.Sign(priv)
	fmt.Println(string(wire.JSONBytes(struct {
		Tx `json:"unwrap"`
	}{tx})))

	r := tx.Validate()
	if !r.IsOK() {
		t.Fatal(r)
	}
}
