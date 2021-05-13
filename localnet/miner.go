package localnet

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"io"
	"io/ioutil"
	"strings"
	"sudachen.xyz/pkg/errstr"
	"sudachen.xyz/pkg/localnet/fu"
	"sudachen.xyz/pkg/localnet/japi"
	"time"
)

func (l *Localnet) writeMinerConfig() (err error) {
	cfg := make(map[string]interface{})
	tick := fu.Maxi(l.LayerDuration/7,2)

	main := make(map[string]interface{})
	main["test-mode"] = true
	main["layers-per-epoch"] = l.LayersPerEpoch
	main["layer-duration-sec"] = l.LayerDuration
	main["layer-average-size"] = l.Count/2
	main["coinbase"] = strings.Join(l.Coinbase,",")
	main["genesis-time"] = l.genesis.Add(time.Duration(5)*time.Second).Format(time.RFC3339)
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
	hare["hare-max-adversaries"] = int(float64(l.Commite) * 0.4) // for small nodes count we have real commite variance about 20%
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
	fu.Verbose("Starting miner %d ...", i+1)

	cmd := append(l.ClientCmd, "--config=/config.json")

	bnCount := fu.Mini(BootnodesCount, l.Count/2)

	if i != 0 {
		var (
			id string
			nodes []string
		)
		for j := 0; j < i && j < bnCount; j++ {
			if id, err = l.getId(j); err != nil {
				return
			}
			nodes = append(nodes,fmt.Sprintf("spacemesh://%s@%s%d:%d", id, l.SubnetPrefix, j+1, l.P2p))
		}
		cmd = append(cmd,"--bootstrap","--bootnodes="+strings.Join(nodes,","))
	}

	netcfg := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			l.NetworkName : {
				//Aliases: []string{fmt.Sprintf("miner_%d",i+1)},
				IPAMConfig: &network.EndpointIPAMConfig{
					IPv4Address: l.SubnetPrefix + fmt.Sprint(i+1),
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

	body, err := l.docker.ContainerCreate(l.ctx,
		&container.Config{
			Image: l.MinerImage,
			Cmd: cmd,
			Labels: l.MinerLabels,
			Entrypoint: []string{"/bin/go-spacemesh"},
			ExposedPorts: ports,
		},
		&container.HostConfig{
			Binds: []string{
				"/tmp/spacemesh-miner-config.json:/config.json:ro",
			},
			//AutoRemove: true,
		},
		netcfg,
		nil,
		fmt.Sprintf("spacemesh_miner_%d",i+1))
	if err != nil {
		return
	}
	err = l.docker.ContainerStart(l.ctx, body.ID, types.ContainerStartOptions{})
	if err != nil { return }
	err = l.waitFor(fmt.Sprintf("spacemesh_miner_%d",i+1))
	if err != nil { return }
	return
}

func (l *Localnet) getId(n int) (id string, err error) {
	if v, ok := l.ids[n]; ok {
		return v, nil
	}
	for j := 0; j < 10; j++ {
		var rdc io.ReadCloser
		rdc, _, err = l.docker.CopyFromContainer(l.ctx, fmt.Sprintf("spacemesh_miner_%d",n+1), fmt.Sprintf("/root/spacemesh/%d/p2p/nodes", NetworkID))
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
						l.ids[n] = i.(string)
						return l.ids[n], nil
					}
					break
				}
			}
		}
	}
	return "", errstr.New("node did not create id.json file")
}

func (l *Localnet) waitForMinerJson() (err error) {
	fu.Verbose("Waiting for bootstrap node json port")
	for i:=0 ;; i++ {
		c := japi.Remote{
			Endpoint: fmt.Sprintf("%s:%d",l.SubnetPrefix+"1",l.Json),
			Verbose: fu.Verbose }.New()
		_, err := c.Echo("hello")
		if err == nil {
			break
		}
		if i == 20 {
			return errstr.Errorf("miner is not started in required time: %s", err.Error())
		}
		time.Sleep(5*time.Second)
	}
	return
}
