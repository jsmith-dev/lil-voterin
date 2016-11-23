package state

import (
	"reflect"

	. "github.com/tendermint/go-common"
	"github.com/tendermint/go-wire"
	"github.com/tendermint/lil-voterin/types"
	tmsp "github.com/tendermint/tmsp/types"
)

const maxTxSize = 2048 // 2KB

//----------------------------------------
// execute txs

// unmarshal the tx and execute it
func ExecTxBytes(state *State, txBytes []byte, appendTx bool) (res tmsp.Result) {
	if len(txBytes) > maxTxSize {
		return tmsp.ErrEncodingError.AppendLog("Tx size exceeds maximum")
	}
	// Decode tx
	var tx types.Tx
	var err error
	txI := wire.ReadJSON(struct {
		types.Tx `json:"unwrap"`
	}{tx}, txBytes, &err)
	if err != nil {
		return tmsp.ErrEncodingError.AppendLog("Error decoding tx: " + err.Error())
	}
	tx = txI.(struct {
		types.Tx `json:"unwrap"`
	}).Tx
	// Validate tx
	res = tx.Validate()
	if !res.IsOK() {
		return res
	}

	// Execute tx
	return ExecTx(state, tx, appendTx)
}

// execute the tx
func ExecTx(state *State, tx types.Tx, appendTx bool) (res tmsp.Result) {
	switch tx_ := tx.(type) {
	case *types.VoteTx:
		return ExecVoteTx(state, tx_, appendTx)
	case *types.AdminTx:
		return ExecAdminTx(state, tx_, appendTx)
	case *types.ForkTx:
		return ExecForkTx(state, tx_, appendTx)
	}
	// NOTE: tx should already by decoded properly and be one of the above
	// so this should never happen
	return tmsp.ErrInternalError.AppendLog(Fmt("Unknown types.Tx type %v", reflect.TypeOf(tx)))
}

func ExecVoteTx(state *State, tx *types.VoteTx, appendTx bool) tmsp.Result {
	// load account
	acc, err := state.GetAccount(tx.PubKey)
	if err != nil {
		return tmsp.ErrUnauthorized.AppendLog(Fmt("Error getting account %X: %v", tx.PubKey, err))
	}

	// check account is voter type
	if acc.Type != types.AccountTypeVoter {
		return tmsp.ErrUnauthorized.AppendLog(Fmt("Account %X is type %v, not voter (%v)", tx.PubKey, acc.Type, types.AccountTypeVoter))
	}

	// check tx.Nonce not already used
	if !state.AddNonce(tx.PubKey, tx.Nonce) {
		return tmsp.ErrBadNonce.AppendLog(Fmt("Nonce %X already used", tx.Nonce))
	}

	// add ballots
	tally := state.GetTally()
	for _, ballot := range tx.Ballots {
		// XXX: bad ballots do not cause an error
		// but do not effect the tally
		tally.AddBallot(ballot)
	}
	state.SetTally(tally)

	// increment account sequence number
	acc.Sequence += 1
	// state.SetAccount(tx.PubKey, acc)

	return tmsp.OK
}

func ExecAdminTx(state *State, tx *types.AdminTx, appendTx bool) tmsp.Result {
	// load account
	acc, err := state.GetAccount(tx.PubKey)
	if err != nil {
		return tmsp.ErrUnauthorized.AppendLog(Fmt("Error getting account %X: %v", tx.PubKey, err))
	}

	// check account is admin type
	if acc.Type != types.AccountTypeAdmin {
		return tmsp.ErrUnauthorized.AppendLog(Fmt("Account %X is type %v, not admin (%v)", tx.PubKey, acc.Type, types.AccountTypeAdmin))
	}

	// check tx.Nonce not already used
	if !state.AddNonce(tx.PubKey, tx.Nonce) {
		return tmsp.ErrBadNonce.AppendLog(Fmt("Nonce %X already used", tx.Nonce))
	}

	// update accounts
	for _, pubAcc := range tx.PubAccounts {
		state.SetAccount(pubAcc.PubKey, pubAcc.Account)
	}

	acc.Sequence += 1

	return tmsp.OK
}

func ExecForkTx(state *State, tx *types.ForkTx, appendTx bool) tmsp.Result {
	// load account
	acc, err := state.GetAccount(tx.PubKey)
	if err != nil {
		return tmsp.ErrUnauthorized.AppendLog(Fmt("Error getting account %X: %v", tx.PubKey, err))
	}

	// check account is admin type
	if acc.Type != types.AccountTypeAdmin {
		return tmsp.ErrUnauthorized.AppendLog(Fmt("Account %X is type %v, not admin (%v)", tx.PubKey, acc.Type, types.AccountTypeAdmin))
	}

	// check tx.Nonce not already used
	if !state.AddNonce(tx.PubKey, tx.Nonce) {
		return tmsp.ErrBadNonce.AppendLog(Fmt("Nonce %X already used", tx.Nonce))
	}

	acc.Sequence += 1

	return tmsp.OK
}
