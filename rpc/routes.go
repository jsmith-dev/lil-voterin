package core

import (
	rpc "github.com/tendermint/go-rpc/server"
	"github.com/tendermint/lil-voterin/types"
)

var Routes = map[string]*rpc.RPCFunc{
	"get_tally":    rpc.NewRPCFunc(GetTallyResult, ""),
	"get_account":  rpc.NewRPCFunc(GetAccountResult, "pubkey"),
	"get_accounts": rpc.NewRPCFunc(GetAccountsResult, ""),
}

func GetTallyResult() (LilVoterinResult, error) {
	if r, err := GetTally(); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func GetAccountResult(pubKey types.PubKey) (LilVoterinResult, error) {
	if r, err := GetAccount(pubKey); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}

func GetAccountsResult() (LilVoterinResult, error) {
	if r, err := GetAccounts(); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}
