package types

import (
	"github.com/tendermint/go-merkle"
)

//-------------------
// implements app.DB

type DB struct {
	merkle.Tree
}

func (db DB) GetSync(key []byte) ([]byte, error) {
	_, value, _ := db.Get(key)
	return value, nil
}

func (db DB) SetSync(key []byte, value []byte) error {
	db.Set(key, value)
	return nil
}

func (db DB) RemSync(key []byte) error {
	db.Remove(key)
	return nil
}

func (db DB) CommitSync() ([]byte, string, error) {
	hash := db.Save()
	return hash, "Success", nil
}
