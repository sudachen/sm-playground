package localnet

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"io"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"
	"sudachen.xyz/pkg/errstr"
	"sudachen.xyz/pkg/localnet/client"
	"sudachen.xyz/pkg/localnet/fu"
	"sudachen.xyz/pkg/localnet/japi"
	"time"
)

func (l * Localnet) MinEligibility() int64 {
	// for small nodes count we have real commite variance about 20%
	return int64(float64(l.Commite) * 0.4)
}

func (l *Localnet) writeMinerConfig() (err error) {
	cfg := make(map[string]interface{})
	tick := fu.Maxi(l.LayerDuration/7,2)

	main := make(map[string]interface{})
	main["test-mode"] = true
	main["layers-per-epoch"] = l.LayersPerEpoch
	main["layer-duration-sec"] = l.LayerDuration
	main["layer-average-size"] = fu.Fnzi(l.LayerSize,l.Count*8/10)
	main["coinbase"] = strings.Join(l.Coinbase,",")
	main["genesis-time"] = l.genesis.Format(time.RFC3339)
	main["genesis-total-weight"] = l.Count*l.MiningSpace
	main["space-to-commit"] = l.MiningSpace
	main["hdist"] = fu.Mini(fu.Maxi(l.LayersPerEpoch/2,10),l.LayersPerEpoch*2-1)
	main["poet-server"] = fmt.Sprintf("%s:%d",l.PoetIP(),l.Poet)
	main["start-mining"] = true
	main["sync-interval"] = fu.Maxi(tick*2/3,2)
	main["sync-validation-delta"] = tick*2

	cfg["main"] = main

	hare := make(map[string]interface{})
	hare["hare-committee-size"] = l.Commite
	hare["hare-max-adversaries"] = l.MinEligibility()
	hare["hare-exp-leaders"] = l.Leaders()
	hare["hare-limit-concurrent"] = 4
	hare["hare-limit-iterations"] = l.HareLimit
	hareDuration := tick
	hare["hare-wakeup-delta"] = tick*2
	hare["hare-round-duration-sec"] = hareDuration
	cfg["hare"] = hare

	elig := make(map[string]interface{})
	// hare must terminate or fail before safe layer be used.
	elig["eligibility-confidence-param"] = fu.Mini(l.LayersPerEpoch*2-1,10/*m*/*60/l.LayerDuration)
	cfg["hare-eligibility"] = elig

	swarm := make(map[string]interface{})
	swarm["randcon"] = P2pRandCon
	swarm["alpha"] = P2pAlfa
	swarm["bucketsize"] = P2pRandCon*2

	p2p := make(map[string]interface{})
	p2p["swarm"] = swarm
	p2p["network-id"] = NetworkID
	p2p["acquire-port"] = false
	p2p["tcp-port"] = l.P2p
	cfg["p2p"] = p2p

	api := make(map[string]interface{})
	api["grpc"] = strings.Join(l.Services,",")
	api["json-server"] = true
	api["grpc-port"] = l.Grpc
	api["json-port"] = l.Json
	cfg["api"] = api

	post := make(map[string]interface{})
	post["post-space"] = PostUnitSize
	post["post-difficulty"] = l.Difficulty
	cfg["post"] = post

	logx := make(map[string]interface{})
	for _, x := range l.Debug {
		logx[x] = "debug"
	}
	cfg["logging"] = logx

	b := bytes.Buffer{}
	if err = json.NewEncoder(&b).Encode(&cfg); err != nil {
		return
	}
	config := "/tmp/spacemesh-miner-config.json"
	err = ioutil.WriteFile(config,b.Bytes(),0644)
	return
}

func (l *Localnet) startMiner(i int) (err error) {
	n := fmt.Sprintf("spacemesh_miner_%d",i)
	fu.Verbose("Starting miner %s ...", n)

	var cmd []string

	for _, k := range l.Massif {
		if k == i {
			cmd = []string{"/usr/bin/valgrind", "-v", "--tool=massif", "--stacks=yes", "--detailed-freq=10000", fmt.Sprintf("--massif-out-file=/massif/%s.out."+l.Genesis(), n)}
			//cmd = []string{"/usr/bin/valgrind", "-v", "--tool=massif", "--stacks=yes", "--time-unit=ms", "--max-snapshots=1000", "--detailed-freq=1000", fmt.Sprintf("--massif-out-file=/massif/%s.out."+l.Genesis(), n)}
			break
		}
	}

	cmd = append(cmd,"/bin/go-spacemesh")
	cmd = append(cmd,l.ClientCmd...)
	cmd = append(cmd,"--config=/config.json")

	bnCount := fu.Mini(BootnodesCount, l.Count/2)

	if i != 1 {
		var (
			id string
			nodes []string
		)
		for j := 1; j < i && j < bnCount; j++ {
			if id, err = l.getIdFor(fmt.Sprintf("spacemesh_miner_%d",j)); err != nil {
				fu.Error(err.Error())
				return
			}
			nodes = append(nodes,fmt.Sprintf("spacemesh://%s@%s%d:%d", id, l.SubnetPrefix, j, l.P2p))
		}
		cmd = append(cmd,"--bootstrap","--bootnodes="+strings.Join(nodes,","))
	}

	if l.ReportPerf {
		cmd = append(cmd,
			fmt.Sprintf("--profiler-name=%v",n),
			fmt.Sprintf("--profiler-url=http://%s:4040",l.Gateway()),
			)
	}

	netcfg := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			l.NetworkName : {
				IPAMConfig: &network.EndpointIPAMConfig{
					IPv4Address: l.SubnetPrefix + fmt.Sprint(i),
				},
			},
		},
	}

	ports := nat.PortSet{}
	k, _ := nat.NewPort("tcp", fmt.Sprint(l.P2p))
	ports[k] = struct{}{}
	k, _ = nat.NewPort("udp", fmt.Sprint(l.P2p))
	ports[k] = struct{}{}
	k, _ = nat.NewPort("tcp", fmt.Sprint(l.Json))
	ports[k] = struct{}{}
	k, _ = nat.NewPort("tcp", fmt.Sprint(l.Grpc))
	ports[k] = struct{}{}
	k, _ = nat.NewPort("tcp", fmt.Sprint("6060"))
	ports[k] = struct{}{}

	env := []string{
		"GODEBUG=gctrace=1,scavtrace=1,gcpacertrace=1,madvdontneed=1,schedtrace=60000",
		"GENESIS="+l.Genesis(),
	}

	if l.CpuPerNode > 0 {
		env = append(env,fmt.Sprintf("GOMAXPROCS=%d",l.CpuPerNode))
	}

	body, err := l.docker.ContainerCreate(l.ctx,
		&container.Config{
			Image: l.MinerImage,
			Labels: mixLabels(l.MinerLabels, map[string]string{
				"number" : fmt.Sprint(i),
			}),
			Entrypoint: cmd,
			ExposedPorts: ports,
			Env: env,
		},
		&container.HostConfig{
			Binds: []string{
				"/tmp/spacemesh-miner-config.json:/config.json:ro",
				"/tmp/spacemesh-massif:/massif",
			},
			Resources: container.Resources{Memory: int64(l.MemoryLimit)*1024*1024 },
			//AutoRemove: true,
		},
		netcfg,
		nil,
		n)
	if err != nil {
		return
	}
	err = l.docker.ContainerStart(l.ctx, body.ID, types.ContainerStartOptions{})
	if err != nil { return }
	err = l.waitFor(n)
	if err != nil { return }
	return
}

func (l *Localnet) getIdFor(name string) (id string, err error) {
	if v, ok := l.ids[name]; ok {
		return v, nil
	}
	for j := 0; j < 10; j++ {
		var rdc io.ReadCloser
		rdc, _, err = l.docker.CopyFromContainer(l.ctx, name, fmt.Sprintf("/root/spacemesh/%d/p2p/nodes", NetworkID))
		if err != nil {
			if !strings.Contains(err.Error(),"No such") {
				fu.Verbose("id copy failed: %v",err.Error())
			}
			time.Sleep(3*time.Second)
			continue
		}
		defer rdc.Close()
		tr := tar.NewReader(rdc)
		for {
			if h, e := tr.Next(); e != io.EOF {
				if e != nil {
					return "", err
				}
				if strings.HasSuffix(h.Name, "id.json") {
					v := map[string]interface{}{}
					if err = json.NewDecoder(tr).Decode(&v); err != nil {
						return
					}
					if i, ok := v["pubKey"]; ok {
						l.ids[name] = i.(string)
						return i.(string), nil
					}
					break
				}
			}
		}
	}
	return "", errstr.New("node did not create id.json file")
}

func (l *Localnet) waitForMinerJson(ip string) (err error) {
	for i:=0 ;; i++ {
		c := japi.Remote{
			Endpoint: fmt.Sprintf("%s:%d",ip,l.Json),
			Verbose: fu.Verbose }.New()
		_, err := c.Echo("hello")
		if err == nil {
			break
		}
		if i == 20 {
			return errstr.Errorf("miner is not AlreadyStarted in required time: %s", err.Error())
		}
		fu.Verbose("Waiting for active json port on %s",ip)
		time.Sleep(5*time.Second)
	}
	return
}

func (l *Localnet) waitForMinerGrpc(ip string) (err error) {
	for i:=0 ;; i++ {
		var b *client.Backend
		if  b, err = client.OpenConnection(fmt.Sprintf("%s:%d",ip,l.Grpc)); err == nil {
			defer b.Close()
			if b.Echo() == nil { return }
		}
		if i == 20 {
			return errstr.Errorf("miner is not AlreadyStarted in required time: %s", err.Error())
		}
		fu.Verbose("Waiting for active grpc port on %s", ip)
		time.Sleep(5*time.Second)
	}
}

type MinerInfo struct {
	NodeID string
	Name string
	Genesis string
	Ip string
	Number int
}

func (l *Localnet) ListMiners() (lst []MinerInfo, err error) {
	if err = l.ConnectDocker(); err != nil {
		return nil, err
	}
	containers, err := l.docker.ContainerList(l.ctx,
		types.ContainerListOptions{
			All: true,
			Filters: filters.NewArgs(
				filters.Arg("label",l.DockerLabel)),
		})
	if err != nil {
		return nil, err
	}
	cc := map[int]types.Container{}
	ccn := []int{}
	for _,c := range containers {
		if c.Labels[RoleLabel] != MinerRole { continue }
		n, _ := strconv.Atoi(c.Labels["number"])
		cc[n] = c
		ccn = append(ccn,n)
	}
	sort.Ints(ccn)
	for _, i := range ccn {
		c := cc[i]
		if c.Labels[RoleLabel] != MinerRole { continue }
		n := c.Names[0]
		if strings.HasPrefix(n,"/") {
			n = n[1:]
		}
		if err = l.waitForMinerGrpc(c.NetworkSettings.Networks[l.NetworkName].IPAddress); err != nil {
			fu.Error("miner %v is not ready: %s", n, err.Error())
			continue
		}
		m := MinerInfo{Name: n, Number: i}
		m.NodeID, err = l.getIdFor(n)
		if err != nil {
			fu.Error("failed to get miner info for %v", n)
			continue
		}
		if v, ok := c.Labels["genesis"]; ok {
			m.Genesis = v
		}
		m.Ip = c.NetworkSettings.Networks[l.NetworkName].IPAddress
		lst = append(lst,m)
	}
	return
}
