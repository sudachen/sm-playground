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
)

const RoleLabel = "spacemesh/role"
const PoetRole = "poet"
const MinerRole = "miner"

const NetworkID = 1

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
const DefaultNodesCount = 8

// EnvNetworkNamee is the environment variable name
const EnvNetworkName = "LOCALNET_NETWORK_NAME"
// DefaultNetworkName is the default network name
const DefaultNetworkName = "spacemesh/localnet"

// EnvDebug is the environment variable name
const EnvDebug = "LOCALNET_DEBUG"
// DefaultDebug defines default sets of services
var DefaultDebug = []string{"poet", "post", "hare", "blockBuilder", "atxBuilder"}

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
const DefaultMasterNodeIP = "100"

// EnvPoetNodeIP is the environment variable name
const EnvPoetNodeIP = "LOCALNET_POET_IP"
// DefaultPoetNodeIP defines poet node host IP in network subnet
const DefaultPoetNodeIP = "101"

// EnvForcePull is the environment variable name
const EnvForcePull = "LOCALNET_FORCE_PULL"
// DefaultForcePull defined that localnet must force images pulling
const DefaultForcePull = false

// EnvCoinbase is the environment variable name
const EnvCoinbase = "LOCALNET_COINBASE"
// DefaultCoinbase defines list of coinbase sparated by ':'. It's used by miners.
var DefaultCoinbase = []string{"b8110cfeB1f01011E118BdB93F1Bb14D2052c276"}

const EnvComplexity = "LOCALNET_COMPLEXITY"
const DefaultComplexity = 8
const EnvMiningSpace = "LOCALNET_MINING_SPACE"
const DefaultMiningSpace = 1024 //*1024
const EnvLayersPerEpoch = "LOCALNET_LAYER_PER_EPOCH"
const DefaultLayersPerEpoch = 2
const EnvLayerDuration = "LOCALNET_LAYER_DURATION"
const DefaultLayerDuration = 50
const EnvHareDuration = "LOCALNET_HARE_DURATION"
const DefaultHareDuration = 20
const EnvHareLimit = "LOCALNET_HARE_LIMIT"
const DefaultHareLimit = 12
const EnvCommite = "LOCALNET_COMMITE"
const DefaultCommite = 800
const EnvLeaders = "LOCALNET_LEADERS" // percent of nodes
const DefaultLeaders = 10

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

	Count 		int

	Json, Grpc, P2p, Poet int
	Services []string

	MiningSpace    int
	LayersPerEpoch int
	LayerDuration  int
	HareDuration   int
	HareLimit      int
	Complexity     int
	Commite        int
	Leaders        int

	MinerLabels map[string]string
	PoetLabels map[string]string
	NetworkLabels map[string]string

	DockerLabel string

	NetworkName string
	Subnet string
	Gateway string
	PoetIP  string
	MasterIP  string

	docker *client.Client
	ctx    context.Context

	Coinbase []string
	ClientCmd []string

	Debug []string

	P2pRandcon int
	P2pAlpha int

	bootstrapId   string
	genesis  string
}

func New() (l *Localnet) {
	l = &Localnet{}

	l.MinerImage = lookupString(EnvClientDockerImage,DefaultClientDockerImage)
	l.PoetImage = lookupString(EnvClientPoetImage,DefaultPoetDockerImage)
	l.PullImages = lookupBool(EnvForcePull,DefaultForcePull)

	l.Count = lookupInt(EnvNodesCount,DefaultNodesCount)

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

	l.Complexity = lookupInt(EnvComplexity,DefaultComplexity)

	l.NetworkLabels = map[string]string{
		l.DockerLabel: "true",
	}

	l.NetworkName = lookupString(EnvNetworkName,DefaultNetworkName)
	subnetPrefix := lookupString(EnvNetworkSubnet,DefaultNetworkSubnet)
	l.Subnet = subnetPrefix + "0/24"
	l.Gateway = subnetPrefix + "1"
	l.PoetIP = subnetPrefix + lookupString(EnvPoetNodeIP,DefaultPoetNodeIP)
	l.MasterIP = subnetPrefix + lookupString(EnvMasterNodeIP,DefaultMasterNodeIP)

	l.Coinbase = lookupStringArray(EnvCoinbase,DefaultCoinbase)
	l.Debug = lookupStringArray(EnvDebug,DefaultDebug)

	l.P2pRandcon = fu.Maxi(3,l.Count/10)
	l.P2pAlpha = 3

	l.MiningSpace = lookupInt(EnvMiningSpace, DefaultMiningSpace)
	l.LayersPerEpoch = lookupInt(EnvLayersPerEpoch, DefaultLayersPerEpoch)
	l.LayerDuration = lookupInt(EnvLayerDuration, DefaultLayerDuration)
	l.HareDuration = lookupInt(EnvHareDuration, DefaultHareDuration)
	l.HareLimit = lookupInt(EnvHareLimit, DefaultHareLimit)
	l.Commite = lookupInt(EnvCommite, DefaultCommite)
	l.Leaders = lookupInt(EnvLeaders, DefaultLeaders) * l.Commite / 100
	return
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

