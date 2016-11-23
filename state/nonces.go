package state

import (
	dbm "github.com/tendermint/go-db"
	"github.com/tendermint/lil-voterin/types"
)

func NonceKey(pubKey types.PubKey, nonce []byte) string {
	return string(pubKey.Bytes()) + string(nonce)
}

// Cached nonces backed by db
// Suitable for blocks or mempool
type Nonces struct {
	cache map[string]struct{}
	db    dbm.DB
}

func NewNonces(db dbm.DB) *Nonces {
	return &Nonces{
		cache: make(map[string]struct{}),
		db:    db,
	}
}

func (n *Nonces) Copy() *Nonces {
	return NewNonces(n.db)
}

// If nonce already exists for pubkey, return false.
// Else, add the nonce and return true
func (n *Nonces) AddNonce(pubKey types.PubKey, nonce []byte) bool {
	nonceKey := NonceKey(pubKey, nonce)
	// check if nonce is in cache
	_, ok := n.cache[nonceKey]
	if ok {
		return false
	}
	// check if nonce is in db
	b := n.db.Get([]byte(nonceKey))
	if len(b) > 0 {
		return false
	}
	// cache nonce
	n.cache[nonceKey] = struct{}{}
	return true
}

func (n *Nonces) Save() {
	b := []byte{1}
	for nonceKey, _ := range n.cache {
		n.db.Set([]byte(nonceKey), b)
	}

	// clear the cache
	n.cache = make(map[string]struct{})
}
