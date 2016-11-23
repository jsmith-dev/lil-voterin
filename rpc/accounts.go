package core

import (
	"github.com/tendermint/lil-voterin/types"
)

func GetAccount(pubKey types.PubKey) (*ResultGetAccount, error) {
	acc, err := voter.GetAccount(pubKey)
	if err != nil {
		return nil, err
	}
	return &ResultGetAccount{*acc}, nil
}

func GetAccounts() (*ResultGetAccounts, error) {
	accs, err := voter.GetAccounts()
	return &ResultGetAccounts{
		NumAccounts: len(accs),
		Accounts:    accs,
	}, err
}
