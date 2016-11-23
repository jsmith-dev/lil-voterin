package types

import (
	"bytes"
	"testing"
)

func TestAccountKey(t *testing.T) {
	_, pub1, _ := NewAccount(AccountTypeVoter)
	s := AccountKeyString(pub1)
	pub2 := BytesToAccountKey([]byte(s))

	if !bytes.Equal(pub1.Bytes(), pub2.Bytes()) {
		t.Fatalf("Pubkeys dont match. %X and %X", pub1, pub2)
	}
}

func TestAccountMarshal(t *testing.T) {
	_, _, acc := NewAccount(AccountTypeVoter)
	for i := 0; i < 10000; i++ {
		b := acc.Marshal()
		acc2 := new(Account)
		err := acc2.Unmarshal(b)
		if err != nil {
			t.Fatal(i, err)
		}
		if acc2.Sequence != acc.Sequence {
			t.Fatalf("sequences dont match. got %v, expected %v", acc2.Sequence, acc.Sequence)
		}
		acc.Sequence += 1
	}
}
