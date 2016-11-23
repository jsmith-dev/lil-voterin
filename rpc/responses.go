package core

import (
	"github.com/tendermint/go-rpc/types"
	"github.com/tendermint/go-wire"
	"github.com/tendermint/lil-voterin/types"
)

type ResultGetTally struct {
	Tally *types.Tally `json:"tally"`
}

type ResultGetAccount struct {
	Account types.Account `json:"account"`
}

type ResultGetAccounts struct {
	NumAccounts int                 `json:"num_accounts"`
	Accounts    []*types.PubAccount `json:"accounts"`
}

//----------------------------------------
// response & result types

const (
	ResultTypeGetTally = byte(0x01)

	ResultTypeGetAccount  = byte(0x10)
	ResultTypeGetAccounts = byte(0x11)
)

type LilVoterinResult interface {
	rpctypes.Result
}

// for wire.readReflect
var _ = wire.RegisterInterface(
	struct{ LilVoterinResult }{},
	wire.ConcreteType{&ResultGetTally{}, ResultTypeGetTally},
	wire.ConcreteType{&ResultGetAccount{}, ResultTypeGetAccount},
	wire.ConcreteType{&ResultGetAccounts{}, ResultTypeGetAccounts},
)
