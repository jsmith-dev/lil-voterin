package app

import (
	"fmt"
	"sync"

	sm "github.com/tendermint/lil-voterin/state"
	"github.com/tendermint/lil-voterin/types"

	. "github.com/tendermint/go-common"
	dbm "github.com/tendermint/go-db"
	"github.com/tendermint/go-wire"
	tmsp "github.com/tendermint/tmsp/types"
)

const version = "0.1"

type LilVoterin struct {
	mtx          sync.Mutex
	state        *sm.State // consensus
	blockState   *sm.State // mid-block state
	mempoolState *sm.State // mempool
}

func NewLilVoterin(db dbm.DB, nCandidates int) *LilVoterin {
	state := sm.NewState(db, nCandidates)
	return &LilVoterin{
		state:        state,
		blockState:   state.Copy(),
		mempoolState: state.Copy(),
	}
}

func (app *LilVoterin) Load(genesisFile string) {
	// TODO: better error check
	if err := app.blockState.Load(); err != nil {
		fmt.Println("Failed to load state", err)
		fmt.Println("trying genesis")

		// Load GenesisState for app
		jsonBytes, err := ReadFile(genesisFile)
		if err != nil {
			Exit("read genesis: " + err.Error())
		}
		fmt.Println(string(jsonBytes))
		genesisState := types.GenesisState{}
		wire.ReadJSONPtr(&genesisState, jsonBytes, &err)
		if err != nil {
			Exit("parsing genesis JSON: " + err.Error())
		}
		fmt.Println("Gen:", genesisState)
		for _, account := range genesisState.Accounts {
			if err := app.setAccount(account.PubKey, account.Account); err != nil {
				Exit("loading genesis accounts: " + err.Error())
			}
		}
	}
	app.Commit()
}

// TMSP::Info
func (app *LilVoterin) Info() string {
	return "LilVoterin v" + version
}

// TMSP::SetOption
func (app *LilVoterin) SetOption(key string, value string) (log string) {
	return ""
}

// TMSP::AppendTx
func (app *LilVoterin) AppendTx(txBytes []byte) (res tmsp.Result) {
	return sm.ExecTxBytes(app.blockState, txBytes, true)
}

// TMSP::CheckTx
func (app *LilVoterin) CheckTx(txBytes []byte) (res tmsp.Result) {
	return sm.ExecTxBytes(app.mempoolState, txBytes, false)
}

// TMSP::Query
// TODO
func (app *LilVoterin) Query(query []byte) (res tmsp.Result) {
	return tmsp.NewResultOK(nil, "")
	/*
		value, err := app.db.GetSync(query)
		if err != nil {
			panic("Error making query: " + err.Error())
		}
		return tmsp.NewResultOK(value, "Success")*/
}

// TMSP::Commit
func (app *LilVoterin) Commit() (res tmsp.Result) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	// commit the state to disk
	hash, err := app.blockState.Save()
	if err != nil {
		// XXX
		return tmsp.ErrInternalError.AppendLog(err.Error())
	}

	app.state = app.blockState.Copy()

	// reset the mempool and block state
	app.mempoolState = app.state.Copy()
	app.blockState = app.state.Copy()

	return tmsp.NewResultOK(hash, "")
}

//--------------------------------

func (app *LilVoterin) GetTally() *types.Tally {
	app.mtx.Lock()
	defer app.mtx.Unlock()
	return app.state.GetTally()
}

func (app *LilVoterin) GetAccount(pubKey types.PubKey) (*types.Account, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()
	return app.state.GetAccount(pubKey)
}

// For testing - acts on the blockState and must call Commit() to take effect
func (app *LilVoterin) setAccount(pubKey types.PubKey, acc *types.Account) error {
	return app.blockState.SetAccount(pubKey, acc)
}

func (app *LilVoterin) GetAccounts() ([]*types.PubAccount, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()
	return app.state.GetAccounts()
}
