# Lil-Voterin

An app for casting and tallying ballots.

## Install

Install Go and set GOPATH. Then,

```
go get github.com/Mastermind/glide
go get github.com/tendermint/lil-voterin
cd $GOPATH/src/github.com/tendermint/lil-voterin
glide install
go install ./cmd/lil-voterin
```

## Run

Run it using the default keys with 

```
TMROOT=data/tendermint/ lil-voterin node --app_genesis data/lil-voterin/genesis.json 
```


## Vote

See `types/tx.go` for details on formatting. 

TODO: client tool

