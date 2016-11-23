package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	. "github.com/tendermint/go-common"
	cfg "github.com/tendermint/go-config"
	dbm "github.com/tendermint/go-db"
	"github.com/tendermint/go-logger"
	"github.com/tendermint/go-p2p"
	rpcserver "github.com/tendermint/go-rpc/server"
	tmspserver "github.com/tendermint/tmsp/server"

	tmcfg "github.com/tendermint/tendermint/config/tendermint"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/proxy"
	rpccore "github.com/tendermint/tendermint/rpc/core"
	tmtypes "github.com/tendermint/tendermint/types"
	"github.com/tendermint/tendermint/version"

	"github.com/tendermint/lil-voterin/app"
	rpc "github.com/tendermint/lil-voterin/rpc"
)

var config cfg.Config

// These end points are available over the internet.
// The rest are only available locally
var safeRoutePoints = []string{
	"status",
	"genesis",
	"block",
	"blockchain",
	"validators",
	"dump_consensus_state",
	"broadcast_tx_sync",
	"num_unconfirmed_txs",
}

func main() {

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println(`Tendermint

Commands:
    node            Run the tendermint node
    version         Show version info
`)
		return
	}

	// Get configuration
	config = tmcfg.GetConfig("")
	parseFlags(config, args[1:]) // Command line overrides

	// set the log level
	logger.SetLogLevel(config.GetString("log_level"))

	switch args[0] {
	case "node":
		if tmspServer != "" {
			RunLilVoterinServer(config, tmspServer)
		} else {
			RunLilVoterinTendermint(config)
		}
	case "version":
		fmt.Println("Tendermint", version.Version)
		fmt.Println("LilVoterin", "0.1.0")
	default:
		Exit(Fmt("Unknown command %v\n", args[0]))
	}
}

func LoadLilVoterin(config cfg.Config) *app.LilVoterin {
	// create db for app
	db := dbm.NewDB("lil-voterin-app", "leveldb", appDataDir)

	voterApp := app.NewLilVoterin(db, nCandidates)
	voterApp.Load(appGenesisFile)

	// set the app rpc
	rpc.SetLilVoterin(voterApp)
	mux := http.NewServeMux()
	rpcserver.RegisterRPCFuncs(mux, rpc.Routes)
	_, err := rpcserver.StartHTTPServer("tcp://0.0.0.0:46680", mux)
	if err != nil {
		Exit(err.Error())
	}
	return voterApp
}

func RunLilVoterinServer(config cfg.Config, tmspServer string) {
	voterApp := LoadLilVoterin(config)

	// Start the tmsp listener
	svr, err := tmspserver.NewServer(tmspAddr, tmspServer, voterApp)
	if err != nil {
		Exit("create listener: " + err.Error())
	}

	// Wait forever
	TrapSignal(func() {
		// Cleanup
		svr.Stop()
	})

}

func RunLilVoterinTendermint(config cfg.Config) {

	voterApp := LoadLilVoterin(config)

	// Create & start tendermint node
	privValidatorFile := config.GetString("priv_validator_file")
	privValidator := tmtypes.LoadOrGenPrivValidator(privValidatorFile)
	n := node.NewNode(config, privValidator, proxy.NewLocalClientCreator(voterApp))

	protocol, address := node.ProtocolAndAddress(config.GetString("node_laddr"))
	l := p2p.NewDefaultListener(protocol, address, config.GetBool("skip_upnp"))
	n.AddListener(l)
	if err := n.Start(); err != nil {
		Exit(Fmt("Failed to start node: %v", err))
	}

	log.Notice("Started node", "nodeInfo", n.NodeInfo())

	// If seedNode is provided by config, dial out.
	if config.GetString("seeds") != "" {
		seeds := strings.Split(config.GetString("seeds"), ",")
		n.DialSeeds(seeds)
	}

	// Run the tendermint RPC server.
	if config.GetString("rpc_laddr") != "" {
		_, err := StartRPC(n, config)
		if err != nil {
			PanicCrisis(err)
		}
	}

	// Sleep forever and then...
	TrapSignal(func() {
		n.Stop()
	})
}

// Only expose safe components to the internet
// Expose the rest to localhost
func StartRPC(n *node.Node, config cfg.Config) ([]net.Listener, error) {
	rpccore.SetConfig(config)

	rpccore.SetEventSwitch(n.EventSwitch())
	rpccore.SetBlockStore(n.BlockStore())
	rpccore.SetConsensusState(n.ConsensusState())
	rpccore.SetMempool(n.MempoolReactor().Mempool)
	rpccore.SetSwitch(n.Switch())
	//rpccore.SetPrivValidator(n.PrivValidator())
	rpccore.SetGenesisDoc(n.GenesisDoc())
	rpccore.SetProxyAppQuery(n.ProxyApp().Query())

	safeRoutes := make(map[string]*rpcserver.RPCFunc)
	for _, k := range safeRoutePoints {
		route, ok := rpccore.Routes[k]
		if !ok {
			PanicSanity(k)
		}
		safeRoutes[k] = route
	}

	var listeners []net.Listener

	listenAddrs := strings.Split(config.GetString("rpc_laddr"), ",")
	listenAddrSafe := listenAddrs[0]

	// the first listener is the public safe rpc
	mux := http.NewServeMux()
	rpcserver.RegisterRPCFuncs(mux, safeRoutes)
	listener, err := rpcserver.StartHTTPServer(listenAddrSafe, mux)
	if err != nil {
		return nil, err
	}
	listeners = append(listeners, listener)

	if len(listenAddrs) > 1 {
		listenAddrUnsafe := listenAddrs[1]
		// expose the full rpc
		mux := http.NewServeMux()
		wm := rpcserver.NewWebsocketManager(rpccore.Routes, n.EventSwitch())
		mux.HandleFunc("/websocket", wm.WebsocketHandler)
		rpcserver.RegisterRPCFuncs(mux, rpccore.Routes)
		listener, err := rpcserver.StartHTTPServer(listenAddrUnsafe, mux)
		if err != nil {
			return nil, err
		}
		listeners = append(listeners, listener)
	}

	return listeners, nil
}
