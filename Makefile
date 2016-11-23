.PHONY: all test get_deps

all: test install

install: get_deps
	go install github.com/tendermint/lil-voterin/cmd/...

test:
	go test github.com/tendermint/lil-voterin/...

get_deps:
	go get -d github.com/tendermint/lil-voterin/...
