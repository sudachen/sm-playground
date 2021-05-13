// package localnet implements framework to manage local test network
package localnet

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"os"
	"strconv"
	"strings"
	"sudachen.xyz/pkg/localnet/fu"
	"time"
)

const RoleLabel = "spacemesh/role"
const PoetRole = "poet"
const MinerRole = "miner"

const NetworkID = 1
const BootnodesCount = 1

// EnvDockerLabel is the environment variable name
const EnvDockerLabel = "LOCALNET_LABEL"
// DefaultDockerLabel is the default docker label to mak localnet containers
const DefaultDockerLabel = "spacemesh/localnet"

// EnvClientDockerImage is the environment variable name
const EnvClientDockerImage = "LOCALNET_CLIENT_IMAGE"
// DefaultClientDockerImage is the default docker image to run as client
const DefaultClientDockerImage = "local/go-spacemesh:latest"

// EnvClientPoetImage is the environment variable name
const EnvClientPoetImage = "LOCALNET_POET_IMAGE"
// DefaultPoetDockerImage is the default docker image to run as poet server
const DefaultPoetDockerImage = "local/poet:latest"

// EnvNodesCount is the environment variable name
const EnvNodesCount = "LOCALNET_NODES_COUNT"
// DefaultNodesCount is the default clients to start on the local test network
const DefaultNodesCount = 12

// EnvNetworkNamee is the environment variable name
const EnvNetworkName = "LOCALNET_NETWORK_NAME"
// DefaultNetworkName is the default network name
const DefaultNetworkName = "spacemesh/localnet"

// EnvDebug is the environment variable name
const EnvDebug = "LOCALNET_DEBUG"
// DefaultDebug defines default sets of services
var DefaultDebug = []string{"poet", "post", "hare", "block-builder", "atx-builder", "block-oracle", "hare-oracle", "sync", "trtl", "meshDb"}

// EnvClientGrpc is the environment variable name
const EnvClientGrpc = "LOCALNET_CLIENT_GRPC"
// DefaultClientGrpc defines default sets of services
var DefaultClientGrpc = []string{"node","mesh","globalstate","transaction","smesher","gateway"}

// EnvClientJsonPort is the environment variable name
const EnvClientJsonPort = "LOCALNET_CLIENT_JSON_PORT"
// DefaultClientJsonPort is the default json port for client API
const DefaultClientJsonPort = 9000

// EnvClientGrpcPort is the environment variable name
const EnvClientGrpcPort = "LOCALNET_CLIENT_GRPC_PORT"
// DefaultClientGrpcPort is the default grpc port for the client API
const DefaultClientGrpcPort = 9001

// EnvClientP2pPort is the environment variable name
const EnvClientP2pPort = "LOCALNET_CLIENT_P2P_PORT"
// DefaultClientP2pPort is the default p2p port for client API
const DefaultClientP2pPort = 9002

// EnvClientPoetPort is the environment variable name
const EnvClientPoetPort = "LOCALNET_CLIENT_POET_PORT"
// DefaultClientPoetPort is the default poet REST port for client API
const DefaultClientPoetPort = 9003

// EnvNetworkSubnet is the environment variable name
const EnvNetworkSubnet = "LOCALNET_SUBNET_REFIX"
// DefaultNetworkSubnet is the default docker network subnet
const DefaultNetworkSubnet = "192.168.88."

// EnvMasterNodeIP is the environment variable name
const EnvMasterNodeIP = "LOCALNET_GATE_IP"

// DefaultMasterNodeIP defines gate node host IP in network subnet
const DefaultMasterNodeIP = "1"

// EnvPoetNodeIP is the environment variable name
const EnvPoetNodeIP = "LOCALNET_POET_IP"
// DefaultPoetNodeIP defines poet node host IP in network subnet
const DefaultPoetNodeIP = "253"

// EnvForcePull is the environment variable name
const EnvForcePull = "LOCALNET_FORCE_PULL"
// DefaultForcePull defined that localnet must force images pulling
const DefaultForcePull = false

// EnvCoinbase is the environment variable name
const EnvCoinbase = "LOCALNET_COINBASE"
// DefaultCoinbase defines list of coinbase sparated by ':'. It's used by miners.
var DefaultCoinbase = []string{"b8110cfeB1f01011E118BdB93F1Bb14D2052c276"}

const EnvDifficulty = "LOCALNET_DIFFICULTY"
const DefaultDifficulty = 5
const EnvMiningSpace = "LOCALNET_MINING_SPACE"
const DefaultMiningSpace = PostUnitSize*128
const EnvLayersPerEpoch = "LOCALNET_LAYER_PER_EPOCH"
const DefaultLayersPerEpoch = 3
const EnvLayerDuration = "LOCALNET_LAYER_DURATION"
const DefaultLayerDuration = 240 // sec  => 4 min
const EnvHareLimit = "LOCALNET_HARE_LIMIT"
const DefaultHareLimit = 2 // maximum full consensus runs
const EnvCommite = "LOCALNET_COMMITE"
const DefaultCommite = 800
const EnvLeaders = "LOCALNET_LEADERS" // percent of nodes
const DefaultLeaders = 3

const P2pAlfa = 5
const P2pRandCon = 4
const PostUnitSize = 1024

func lookupInt(envVar string, dflt int) int {
	if v, ok := os.LookupEnv(envVar); ok {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return int(i)
		}
	}
	return dflt
}

func lookupString(envVar string, dflt string) string {
	if v, ok := os.LookupEnv(envVar); ok {
		return v
	}
	return dflt
}

func lookupBool(envVar string, dflt bool) bool {
	if v, ok := os.LookupEnv(envVar); ok {
		switch strings.ToLower(v) {
		case "false","no","0","disable","":
			return false
		case "true","yes","1","enable":
			return true
		}
	}
	return dflt
}

func lookupStringArray(envVar string, dflt []string) []string {
	if v, ok := os.LookupEnv(envVar); ok {
		return strings.Split(v,",")
	}
	return dflt
}

type Localnet struct {
	MinerImage string
	PoetImage  string
	PullImages bool

	Count int

	Json, Grpc, P2p, Poet int
	Services              []string

	MiningSpace    int
	LayersPerEpoch int
	LayerDuration  int
	HareLimit      int
	Difficulty     int
	Commite        int
	ExpLeaders     int

	MinerLabels   map[string]string
	PoetLabels    map[string]string
	NetworkLabels map[string]string

	DockerLabel string

	NetworkName  string

	SubnetPrefix string
	PoetIPsfx   string
	MasterIPsfx string

	docker *client.Client
	ctx    context.Context

	Coinbase  []string
	ClientCmd []string

	Debug []string

	genesis     time.Time
	ids map[int]string

}

func New() (l *Localnet) {
	l = &Localnet{ids:map[int]string{}}

	l.MinerImage = lookupString(EnvClientDockerImage,DefaultClientDockerImage)
	l.PoetImage = lookupString(EnvClientPoetImage,DefaultPoetDockerImage)
	l.PullImages = lookupBool(EnvForcePull,DefaultForcePull)

	l.Count = fu.Maxi(3, lookupInt(EnvNodesCount,DefaultNodesCount) )

	l.Services = lookupStringArray(EnvClientGrpc,DefaultClientGrpc)

	l.Json = lookupInt(EnvClientJsonPort,DefaultClientJsonPort)
	l.Grpc = lookupInt(EnvClientGrpcPort,DefaultClientGrpcPort)
	l.P2p = lookupInt(EnvClientP2pPort,DefaultClientP2pPort)
	l.Poet = lookupInt(EnvClientPoetPort,DefaultClientPoetPort)

	l.DockerLabel = lookupString(EnvDockerLabel,DefaultDockerLabel)

	l.MinerLabels = map[string]string{
		l.DockerLabel: "true",
		RoleLabel: MinerRole, // for internal usage
		"kind": "spacemesh",  // for external analytics/logging
	}

	l.PoetLabels = map[string]string{
		l.DockerLabel: "true",
		RoleLabel: PoetRole, // for internal usage
		"kind": "spacemesh", // for external analytics/logging
	}

	l.Difficulty = lookupInt(EnvDifficulty, DefaultDifficulty)

	l.NetworkLabels = map[string]string{
		l.DockerLabel: "true",
	}

	l.NetworkName = lookupString(EnvNetworkName,DefaultNetworkName)
	l.SubnetPrefix = lookupString(EnvNetworkSubnet,DefaultNetworkSubnet)
	l.PoetIPsfx = lookupString(EnvPoetNodeIP,DefaultPoetNodeIP)
	l.MasterIPsfx = lookupString(EnvMasterNodeIP,DefaultMasterNodeIP)

	l.Coinbase = lookupStringArray(EnvCoinbase,DefaultCoinbase)
	l.Debug = lookupStringArray(EnvDebug,DefaultDebug)

	l.ExpLeaders = lookupInt(EnvLeaders, DefaultLeaders)
	l.MiningSpace = fu.Maxi((lookupInt(EnvMiningSpace, DefaultMiningSpace)+ (PostUnitSize-1))/PostUnitSize*PostUnitSize, PostUnitSize)
	l.LayersPerEpoch = lookupInt(EnvLayersPerEpoch, DefaultLayersPerEpoch)
	l.LayerDuration = lookupInt(EnvLayerDuration, DefaultLayerDuration)
	l.HareLimit = lookupInt(EnvHareLimit, DefaultHareLimit)
	l.Commite = lookupInt(EnvCommite, DefaultCommite)

	return
}

func (l *Localnet) PoetIP() string {
	return l.SubnetPrefix + l.PoetIPsfx
}

func (l *Localnet) MasterIP() string {
	return l.SubnetPrefix + l.MasterIPsfx
}

func (l *Localnet) Subnet() string {
	return l.SubnetPrefix + "0/24"
}

func (l *Localnet) Gateway() string {
	return l.SubnetPrefix + "254"
}

func (l *Localnet) Leaders() int {
	return  l.ExpLeaders * l.Commite / l.Count
}

func (l *Localnet) started() (ok bool,err error) {
	containers, err := l.docker.ContainerList(l.ctx,
		types.ContainerListOptions{
			All: true,
			Filters: filters.NewArgs(
				filters.Arg("label",l.DockerLabel+"=true")),
		})
	if err != nil {
		return
	}
	var poet, miner bool
	for _, c := range containers {
		if r, ok := c.Labels[RoleLabel]; ok {
			switch r {
			case PoetRole: poet = true
			case MinerRole: miner = true
			}
		}
	}
	return poet && miner, nil
}

func (l *Localnet) filters(labels map[string]string) filters.Args {
	args := filters.NewArgs()
	for a,v := range labels {
		args.Add(a,v)
	}
	return args
}

