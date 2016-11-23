package types

import (
	"testing"

	. "github.com/tendermint/go-common"
)

func MakeTestBallots() (Ballot, Ballot) {
	b1 := Ballot{[]Candidate{0, 2, 3}, RandStr(32)}
	b2 := Ballot{[]Candidate{0, 1}, RandStr(32)}
	return b1, b2
}

func TestTally(t *testing.T) {
	N := 5
	tally := NewTally(N)
	b1, b2 := MakeTestBallots()

	if err := tally.AddBallot(b1); err != nil {
		t.Fatal(err)
	}
	if err := tally.AddBallot(b2); err != nil {
		t.Fatal(err)
	}

	checkExpected(t, tally, 0, 2)
	checkExpected(t, tally, 1, 1)
	checkExpected(t, tally, 2, 1)
	checkExpected(t, tally, 3, 1)
	checkExpected(t, tally, 4, 0)
}

func checkExpected(t *testing.T, tally *Tally, index, expected int64) {
	if tally.Counts[index] != expected {
		t.Fatalf("Got %d, expected %d", tally.Counts[index], expected)
	}
}

func TestTallyMarshal(t *testing.T) {
	n := 15
	tally1 := NewTally(n)
	ballot1, ballot2 := MakeTestBallots()
	for i := 0; i < 10000; i++ {
		switch i % 3 {
		case 0:
			if err := tally1.AddBallot(ballot1); err != nil {
				t.Fatal(err)
			}
		case 1:
			if err := tally1.AddBallot(ballot2); err != nil {
				t.Fatal(err)
			}
		case 2:
			if err := tally1.AddBallot(ballot1); err != nil {
				t.Fatal(err)
			}
			if err := tally1.AddBallot(ballot2); err != nil {
				t.Fatal(err)
			}
		}
		b := tally1.Marshal()
		tally2 := NewTally(n)
		err := tally2.Unmarshal(b)
		if err != nil {
			t.Fatal(i, err)
		}
		for i, c := range tally1.Counts {
			if c != tally2.Counts[i] {
				t.Fatalf("Counts dont match for candidate %d. got %d, expected %v", i, tally2.Counts[i], c)
			}
		}
	}
}
