package types

import (
	"math/rand"

	. "github.com/tendermint/go-common"
)

//---------------------------------
// utils for generating random txs

func MakeBallot(nCandidates int) Ballot {
	cz := []Candidate{}
	for i := 0; i < nCandidates && len(cz) < maxVotesPerBallot; i++ {
		if r := rand.Intn(2); r > 0 {
			cz = append(cz, Candidate(i))
		}
	}
	return Ballot{Candidates: cz, Source: RandStr(32)}
}

func MakeVoteTx(tally *Tally, pub PubKey, nonce, nCandidates, nballots int) *VoteTx {
	ballots := make([]Ballot, nballots)
	for i := 0; i < nballots; i++ {
		ballots[i] = MakeBallot(nCandidates)
		if tally != nil {
			tally.AddBallot(ballots[i])
		}
	}
	return &VoteTx{
		Ballots: ballots,
		Nonce:   RandBytes(12),
		PubKey:  pub,
	}
}

func MakeAdminTx(signPub, modPub PubKey, accType AccountType, nonce []byte) *AdminTx {
	return &AdminTx{
		PubAccounts: []PubAccount{
			PubAccount{
				PubKey:  modPub,
				Account: &Account{Type: accType},
			},
		},
		Nonce:  nonce,
		PubKey: signPub,
	}
}
