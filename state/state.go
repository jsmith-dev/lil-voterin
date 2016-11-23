package state

import (
	"fmt"

	dbm "github.com/tendermint/go-db"
	"github.com/tendermint/go-merkle"

	"github.com/tendermint/lil-voterin/types"
)

var StateKey = []byte("STATE")

// State manages accounts, their nonces, and the tally.
// It is suitable for blocks and mempool
// CONTRACT: State should be quick to copy.
// See CacheWrap().
// Not thread-safe
type State struct {
	chainID string

	tally    *types.Tally
	accounts *Accounts
	nonces   *Nonces

	db dbm.DB
}

func (s *State) Copy() *State {
	return &State{
		chainID:  s.chainID,
		tally:    s.tally.Copy(),
		accounts: s.accounts.Copy(),
		nonces:   s.nonces.Copy(),
		db:       s.db,
	}
}

func NewState(db dbm.DB, nCandidates int) *State {
	return &State{
		chainID:  "", // TODO
		tally:    types.NewTally(nCandidates),
		accounts: NewAccounts(merkle.NewIAVLTree(100, db)),
		nonces:   NewNonces(db),
		db:       db,
	}
}

func (s *State) GetChainID() string {
	return s.chainID
}

func (s *State) GetTally() *types.Tally {
	return s.tally
}

func (s *State) SetTally(tally *types.Tally) {
	s.tally = tally
}

func (s *State) GetAccount(pubKey types.PubKey) (*types.Account, error) {
	return s.accounts.GetAccount(pubKey)
}

func (s *State) SetAccount(pubKey types.PubKey, account *types.Account) error {
	return s.accounts.SetAccount(pubKey, account)
}

func (s *State) AddNonce(pubKey types.PubKey, nonce []byte) bool {
	return s.nonces.AddNonce(pubKey, nonce)
}

// Sync the state caches to their dbs and save the merkleized state
func (s *State) Save() ([]byte, error) {
	// sync nonces to disk
	s.nonces.Save()

	// write the merkle tree updates to disk
	rootHash := s.saveAccountsAndTally()

	// save  the rootHash
	s.db.Set(StateKey, rootHash)

	return rootHash, nil
}

func (s *State) saveAccountsAndTally() []byte {
	// add the tally to the merkle tree
	s.accounts.tree.Set(types.TallyKeyBytes, s.tally.Marshal())

	// save tally to db
	s.db.Set(types.TallyKeyBytes, s.tally.Marshal())

	// sync the accounts to the tree and save the tree
	return s.accounts.Save()
}

func (s *State) Load() error {
	// get the root hash
	rootHash := s.db.Get(StateKey)

	// load the merkle tree
	s.accounts.tree.Load(rootHash)

	tallyBytes := s.db.Get(types.TallyKeyBytes)
	if len(tallyBytes) == 0 {
		return fmt.Errorf("New Tally not found in DB")
	}
	err := s.tally.Unmarshal(tallyBytes)
	if err != nil {
		return err
	}

	// grab the tally bytes and unmarshal
	_, tallyBytes, exists := s.accounts.tree.Get(types.TallyKeyBytes)
	if !exists {
		return fmt.Errorf("Tally not found in DB")
	}
	return s.tally.Unmarshal(tallyBytes)
}

//------------------------------------------------------------------------

func (s *State) GetAccounts() ([]*types.PubAccount, error) {
	var accs []*types.PubAccount
	var iterErr error
	stopped := s.accounts.tree.Iterate(func(key []byte, value []byte) (stop bool) {
		// this ignores the tally key
		if len(key) != 32 {
			return false
		}
		var pubKey [32]byte

		copy(pubKey[:], key)
		acc := new(types.Account)
		err := acc.Unmarshal(value)
		if err != nil {
			iterErr = err
			return true
		}
		accs = append(accs, &types.PubAccount{
			PubKey:  types.PubKey(pubKey),
			Account: acc,
		})
		return false
	})
	if stopped {
		return nil, iterErr
	}
	return accs, nil
}
