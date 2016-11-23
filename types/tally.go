package types

import (
	"bytes"
	"fmt"

	"github.com/tendermint/go-wire"
)

const maxVotesPerBallot = 5

//------------------------------------------
// database key for accessing tally

// NOTE: must not be 32 bytes
var (
	TallyKeyString = "TALLYKEY"
	TallyKeyBytes  = []byte(TallyKeyString)
)

//------------------------------------------
// Candidate is an integer - 0-based

type Candidate int

//------------------------------------------
// ballot is a list of candidates voted for.

type Ballot struct {
	Candidates []Candidate `json:"c"`
	Source     string      `json:"s"`
}

//------------------------------------------
// tally is a score for each candidate.

type Tally struct {
	Counts []int64
}

func NewTally(n int) *Tally {
	return &Tally{
		Counts: make([]int64, n),
	}
}

// Add 1 to the tally for each unique index in the ballot
// Returns an error if any element in a ballot is duplicated or greater than len(t)
func (t *Tally) AddBallot(ballot Ballot) error {
	if len(ballot.Candidates) > maxVotesPerBallot {
		return fmt.Errorf("Too many candidates per ballot (%d). Max is %d", len(ballot.Candidates), maxVotesPerBallot)
	}

	l := len(t.Counts)
	diff := make([]int64, l) // better to allocate once and zero?
	for _, v := range ballot.Candidates {
		// -1 is ignored the other votes in the ballot still count
		// but this means we could have eg `[0,5,-1,10,-1]`
		if int(v) == -1 {
			continue
		}

		// check bounds and for duplicates
		if int(v) < 0 {
			return fmt.Errorf("Candidate cannot be negative")
		}
		if int(v) >= l {
			return fmt.Errorf("Vote for candidate %d exceeds number of candidates %d", v, l)
		}
		if diff[v] > 0 {
			return fmt.Errorf("Duplicate candidate %d", v)
		}
		// a vote for candidate v!
		diff[v] += 1
	}
	for i, v := range diff {
		t.Counts[i] += v // TODO: overflow
	}
	return nil
}

func (t *Tally) Copy() *Tally {
	t2 := NewTally(t.N())
	copy(t2.Counts, t.Counts)
	return t2
}

func (t *Tally) N() int {
	return len(t.Counts)
}

func (t *Tally) Marshal() []byte {
	return wire.BinaryBytes(t)
}

func (t *Tally) Unmarshal(b []byte) error {
	r, n, err := bytes.NewBuffer(b), new(int), new(error)
	wire.ReadBinary(t, r, 0, n, err)
	return *err
}
