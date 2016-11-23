package app

import (
	"fmt"
	"testing"
	"time"

	. "github.com/tendermint/go-common"
	dbm "github.com/tendermint/go-db"

	"github.com/tendermint/lil-voterin/types"
	tmsp "github.com/tendermint/tmsp/types"
)

var nTestCandidates = 5

//----------------------------------------------------------------------
// util

func makeTestTx(pub types.PubKey, nonce int) *types.VoteTx {
	b1, b2 := makeTestBallots()
	return &types.VoteTx{
		Ballots: []types.Ballot{b1, b2},
		Nonce:   []byte{byte(nonce)},
		PubKey:  pub,
	}
}

func makeTestBallots() (types.Ballot, types.Ballot) {
	b1 := types.Ballot{[]types.Candidate{0, 2, 3}, RandStr(32)}
	b2 := types.Ballot{[]types.Candidate{0, 1}, RandStr(32)}
	return b1, b2
}

func newLilVoterin(nCandidates int) *LilVoterin {
	db := dbm.NewDB("lil-voterin-app", "memdb", "")
	return NewLilVoterin(db, nCandidates)
}

func expectFail(t *testing.T, r tmsp.Result) {
	if r.Code == 0 {
		panic(Fmt("expected test to fail with bad sig. got code %v, log %s", r.Code, r.Log))
	}
}

func expectPass(t *testing.T, r tmsp.Result) {
	if r.Code != 0 {
		panic(Fmt("expected test to pass. got code %v, log %s", r.Code, r.Log))
	}
}

//----------------------------------------------------------------------
// test signing

func TestSignature(t *testing.T) {
	app := newLilVoterin(nTestCandidates)

	priv, pub, acc := types.NewAccount(types.AccountTypeVoter)
	app.setAccount(pub, acc)
	app.Commit()

	tx := makeTestTx(pub, 1)

	// no sig
	r := app.CheckTx(types.JSONBytes(tx))
	expectFail(t, r)

	// bad sig
	priv2, _, _ := types.NewAccount(types.AccountTypeVoter)
	tx.Sign(priv2)
	r = app.CheckTx(types.JSONBytes(tx))
	expectFail(t, r)

	// good sig
	tx.Sign(priv)
	r = app.CheckTx(types.JSONBytes(tx))
	expectPass(t, r)
}

//----------------------------------------------------------------------
// test nonce behaviour

func TestNonceAppendTx(t *testing.T) {
	app := newLilVoterin(nTestCandidates)

	priv, pub, acc := types.NewAccount(types.AccountTypeVoter)
	app.setAccount(pub, acc)
	app.Commit()

	tx := makeTestTx(pub, 0)
	tx.Sign(priv)

	// good
	r := app.AppendTx(types.JSONBytes(tx))
	expectPass(t, r)

	// invalid nonce
	r = app.AppendTx(types.JSONBytes(tx))
	expectFail(t, r)

	tx = makeTestTx(pub, 1)
	tx.Sign(priv)

	// good
	r = app.AppendTx(types.JSONBytes(tx))
	expectPass(t, r)

	// commit the block
	app.Commit()

	// invalid nonce
	r = app.AppendTx(types.JSONBytes(tx))
	expectFail(t, r)

	tx = makeTestTx(pub, 2)
	tx.Sign(priv)

	// good
	r = app.AppendTx(types.JSONBytes(tx))
	expectPass(t, r)
}

func TestNonceCheckAppend(t *testing.T) {
	app := newLilVoterin(nTestCandidates)

	priv, pub, acc := types.NewAccount(types.AccountTypeVoter)
	app.setAccount(pub, acc)
	app.Commit()

	tx := makeTestTx(pub, 0)
	tx.Sign(priv)

	// good
	r := app.CheckTx(types.JSONBytes(tx))
	expectPass(t, r)

	// good
	r = app.AppendTx(types.JSONBytes(tx))
	expectPass(t, r)

	// invalid nonce
	r = app.CheckTx(types.JSONBytes(tx))
	expectFail(t, r)

	// commit
	app.Commit()

	// invalid nonce
	r = app.CheckTx(types.JSONBytes(tx))
	expectFail(t, r)

	tx = makeTestTx(pub, 1)
	tx.Sign(priv)

	// good
	r = app.CheckTx(types.JSONBytes(tx))
	expectPass(t, r)

	// good
	r = app.AppendTx(types.JSONBytes(tx))
	expectPass(t, r)

	// invalid nonce
	r = app.CheckTx(types.JSONBytes(tx))
	expectFail(t, r)
}

//----------------------------------------------------------------------
// test txs from multiple signers

func TestMultiAccount(t *testing.T) {
	app := newLilVoterin(nTestCandidates)

	priv, pub, acc := types.NewAccount(types.AccountTypeVoter)
	priv2, pub2, acc2 := types.NewAccount(types.AccountTypeVoter)
	app.setAccount(pub, acc)
	app.setAccount(pub2, acc2)
	app.Commit()

	// * first account
	tx := makeTestTx(pub, 0)
	tx.Sign(priv)

	// good
	r := app.AppendTx(types.JSONBytes(tx))
	expectPass(t, r)

	// invalid nonce
	r = app.AppendTx(types.JSONBytes(tx))
	expectFail(t, r)

	// * second account
	tx = makeTestTx(pub2, 0)
	tx.Sign(priv2)

	// good
	r = app.AppendTx(types.JSONBytes(tx))
	expectPass(t, r)

	// invalid nonce
	r = app.AppendTx(types.JSONBytes(tx))
	expectFail(t, r)
}

//----------------------------------------------------------------------
// test admin tx

func makeTestAdminTx(pub types.PubKey, nonce int) *types.AdminTx {
	_, newPub, _ := types.NewAccount(types.AccountTypeVoter)
	return types.MakeAdminTx(pub, newPub, types.AccountTypeVoter, []byte{byte(nonce)})
}

func TestAccountType(t *testing.T) {
	app := newLilVoterin(nTestCandidates)

	priv1, pub1, acc1 := types.NewAccount(types.AccountTypeVoter)
	priv2, pub2, acc2 := types.NewAccount(types.AccountTypeAdmin)
	priv3, pub3, acc3 := types.NewAccount(types.AccountTypeCorrupt)
	app.setAccount(pub1, acc1)
	app.setAccount(pub2, acc2)
	app.setAccount(pub3, acc3)
	app.Commit()

	var tx types.Tx
	var r tmsp.Result
	nonce := 0

	// first account cant make AdminTx
	tx = makeTestAdminTx(pub1, nonce)
	tx.Sign(priv1)
	r = app.AppendTx(types.JSONBytes(tx))
	expectFail(t, r)

	// second account cant make VoteTx
	tx = makeTestTx(pub2, nonce)
	tx.Sign(priv2)
	r = app.AppendTx(types.JSONBytes(tx))
	expectFail(t, r)

	// third account cant do either
	tx = makeTestAdminTx(pub3, nonce)
	tx.Sign(priv3)
	r = app.AppendTx(types.JSONBytes(tx))
	expectFail(t, r)

	tx = makeTestTx(pub3, nonce)
	tx.Sign(priv3)
	r = app.AppendTx(types.JSONBytes(tx))
	expectFail(t, r)
}

func TestAdminTx(t *testing.T) {
	app := newLilVoterin(nTestCandidates)

	// voter1 (secret, pub, account)
	v1s, v1p, v1a := types.NewAccount(types.AccountTypeVoter)
	// admin1 (secret, pub, account)
	a1s, a1p, a1a := types.NewAccount(types.AccountTypeAdmin)
	app.setAccount(v1p, v1a)
	app.setAccount(a1p, a1a)
	app.Commit()

	// admin2. added by admin1
	a2s, a2p, _ := types.NewAccount(types.AccountTypeAdmin)

	// voter2. added by admin1
	v2s, v2p, _ := types.NewAccount(types.AccountTypeVoter)
	// voter3. added by admin2
	v3s, v3p, _ := types.NewAccount(types.AccountTypeVoter)

	var tx types.Tx
	var r tmsp.Result

	// add a new voter and ensure it can vote
	{
		tx = types.MakeAdminTx(a1p, v2p, types.AccountTypeVoter, []byte{0})
		tx.Sign(a1s)
		r = app.AppendTx(types.JSONBytes(tx))
		expectPass(t, r)

		tx = makeTestTx(v2p, 0)
		tx.Sign(v2s)
		r = app.AppendTx(types.JSONBytes(tx))
		expectPass(t, r)
	}

	// remove an old voter and ensure it cant vote
	{
		tx = types.MakeAdminTx(a1p, v1p, types.AccountTypeCorrupt, []byte{1})
		tx.Sign(a1s)
		r = app.AppendTx(types.JSONBytes(tx))
		expectPass(t, r)

		tx = makeTestTx(v1p, 0)
		tx.Sign(v1s)
		r = app.AppendTx(types.JSONBytes(tx))
		expectFail(t, r)
	}

	// add a new admin and ensure it can add an user who can vote
	{
		tx = types.MakeAdminTx(a1p, a2p, types.AccountTypeAdmin, []byte{2})
		tx.Sign(a1s)
		r = app.AppendTx(types.JSONBytes(tx))
		expectPass(t, r)

		tx = types.MakeAdminTx(a2p, v3p, types.AccountTypeVoter, []byte{0})
		tx.Sign(a2s)
		r = app.AppendTx(types.JSONBytes(tx))
		expectPass(t, r)

		tx = makeTestTx(v3p, 0)
		tx.Sign(v3s)
		r = app.AppendTx(types.JSONBytes(tx))
		expectPass(t, r)
	}
}

//----------------------------------------------------------------------
// test throughput

func TestThroughput(t *testing.T) {
	app := newLilVoterin(nTestCandidates)

	priv1, pub1, acc1 := types.NewAccount(types.AccountTypeVoter)
	priv2, pub2, acc2 := types.NewAccount(types.AccountTypeVoter)
	app.setAccount(pub1, acc1)
	app.setAccount(pub2, acc2)
	app.Commit()

	ntxs := 1000
	nballots := 10

	tally := types.NewTally(nTestCandidates)
	start := time.Now()
	for i := 0; i < ntxs; i++ {
		var tx *types.VoteTx
		// split txs over two accounts
		if i%2 == 0 {
			tx = types.MakeVoteTx(tally, pub1, i, nTestCandidates, nballots)
			tx.Sign(priv1)
		} else {
			tx = types.MakeVoteTx(tally, pub2, i, nTestCandidates, nballots)
			tx.Sign(priv2)
		}
		r := app.AppendTx(types.JSONBytes(tx))
		if r.Code != 0 {
			t.Fatalf("Transaction failed %v: %v", r, string(types.JSONBytes(tx)))
		}
	}

	fmt.Printf("took %v. %v ballots/sec\n", time.Since(start), float64(ntxs*nballots)/float64(time.Since(start).Seconds()))

	app.Commit()

	appTally := app.GetTally()
	for i, c := range tally.Counts {
		if c != appTally.Counts[i] {
			t.Fatalf("tallys don't match for index %d. got %d, expected %d", i, appTally.Counts[i], c)
		}
	}
}
