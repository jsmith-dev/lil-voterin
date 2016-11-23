package core

import (
	"github.com/tendermint/lil-voterin/app"
)

var voter *app.LilVoterin

func SetLilVoterin(a *app.LilVoterin) {
	voter = a
}
