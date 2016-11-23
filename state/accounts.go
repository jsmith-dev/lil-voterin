package state

import (
	"fmt"
	"sort"

	"github.com/tendermint/go-merkle"
	"github.com/tendermint/lil-voterin/types"
)

// Cached accounts backed by merkle tree.
// Suitable for blocks and mempool
type Accounts struct {
	cache map[string]*types.Account
	tree  merkle.Tree
}

func NewAccounts(tree merkle.Tree) *Accounts {
	return &Accounts{
		cache: make(map[string]*types.Account),
		tree:  tree,
	}
}

func (accounts *Accounts) Copy() *Accounts {
	return &Accounts{
		cache: make(map[string]*types.Account),
		tree:  accounts.tree.Copy(),
	}
}

func (accounts *Accounts) GetAccount(pubKey types.PubKey) (*types.Account, error) {
	var err error
	acc, ok := accounts.cache[types.AccountKeyString(pubKey)]
	if !ok {
		acc, err = accounts.getAccount(pubKey)
		if err != nil {
			return nil, err
		}
		accounts.cache[types.AccountKeyString(pubKey)] = acc
	}
	if acc == nil {
		return nil, fmt.Errorf("Account not found for pubkey %X", pubKey)
	}
	return acc, nil
}

func (accounts *Accounts) getAccount(pubKey types.PubKey) (*types.Account, error) {
	_, accBytes, exists := accounts.tree.Get(types.AccountKeyBytes(pubKey))
	if !exists {
		return nil, fmt.Errorf("key not found in tree")
	}
	if len(accBytes) == 0 {
		return nil, fmt.Errorf("accBytes is empty")
	}
	account := new(types.Account)
	err := account.Unmarshal(accBytes)
	return account, err
}

func (accounts *Accounts) setAccount(pubKey types.PubKey, acc *types.Account) {
	pubKeyBytes := types.AccountKeyBytes(pubKey)
	accBytes := acc.Marshal()
	accounts.tree.Set(pubKeyBytes, accBytes)
}

func (accounts *Accounts) SetAccount(pubKey types.PubKey, acc *types.Account) error {
	accounts.cache[types.AccountKeyString(pubKey)] = acc
	return nil
}

// sync cache to merkle tree
func (accounts *Accounts) Sync() {
	keys := []string{}
	for pubKey, _ := range accounts.cache {
		keys = append(keys, pubKey)
	}
	sort.Strings(keys)
	for _, k := range keys {
		accounts.setAccount(types.BytesToAccountKey([]byte(k)), accounts.cache[k])
	}
}

// persist merkle tree
func (accounts *Accounts) Save() []byte {
	accounts.Sync()
	return accounts.tree.Save()
}
