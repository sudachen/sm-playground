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

	main := make(map[string]interface{})
	main["test-mode"] = true
	main["layers-per-epoch"] = l.LayersPerEpoch
	main["layer-duration-sec"] = l.LayerDuration
	main["coinbase"] = strings.Join(l.Coinbase,",")
	main["genesis-time"] = l.genesis
	main["genesis-total-weight"] = l.Count*l.MiningSpace
	main["hdist"] = fu.Mini(l.LayersPerEpoch,100)
	main["poet-server"] = fmt.Sprintf("%s:%d",l.PoetIP,l.Poet)
	main["start-mining"] = true
	cfg["main"] = main

	hare := make(map[string]interface{})
	hare["hare-committee-size"] = l.Commite
	hare["hare-max-adversaries"] = l.Commite/2-1
	hare["hare-exp-leaders"] = l.Leaders
	hare["hare-limit-concurrent"] = fu.Maxi(3,l.LayerDuration/l.HareDuration)
	hare["hare-limit-iterations"] = l.HareLimit
	hare["hare-wakeup-delta"] = fu.Maxi(l.HareDuration/3,5)
	hare["hare-round-duration-sec"] = l.HareDuration
	cfg["hare"] = hare

	elig := make(map[string]interface{})
	elig["eligibility-confidence-param"] = uint64(fu.Maxi(2,l.LayersPerEpoch/2))
	cfg["hare-eligibility"] = elig

	swarm := make(map[string]interface{})
	swarm["randcon"] = l.P2pRandcon
	swarm["alpha"] = l.P2pAlpha
	swarm["bucketsize"] = l.P2pRandcon

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
	post["post-space"] = l.MiningSpace
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
	fu.Verbose("Starting miner %d ...", i)

	cmd := append(l.ClientCmd, "--config=/config.json")

	if i != 0 {
		cmd = append(cmd,"--bootstrap",
			"--bootnodes="+fmt.Sprintf("spacemesh://%s@%s:%d", l.bootstrapId, l.MasterIP, l.P2p),
		)
	}

	var netcfg *network.NetworkingConfig
	if i == 0 {
		netcfg = &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				l.NetworkName : {
					Aliases: []string{"miner_0"},
					IPAMConfig: &network.EndpointIPAMConfig{
						IPv4Address: l.MasterIP,
					},
				},
			},
		}
	} else {
		netcfg = &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				l.NetworkName : {
					Aliases: []string{fmt.Sprintf("miner_%d",i)},
				},
			},
		}
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
		fmt.Sprintf("spacemesh_miner_%d",i))
	if err != nil {
		return
	}
	err = l.docker.ContainerStart(l.ctx, body.ID, types.ContainerStartOptions{})
	if err != nil { return }
	err = l.waitFor(fmt.Sprintf("spacemesh_miner_%d",i))
	if err != nil { return }

	if i == 0 {
		for j := 0; j < 10; j++ {
			var rdc io.ReadCloser
			rdc, _, err = l.docker.CopyFromContainer(l.ctx, "spacemesh_miner_0", fmt.Sprintf("/root/spacemesh/%d/p2p/nodes", NetworkID))
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
				h, e := tr.Next()
				if e != nil {
					return e
				}
				if e != io.EOF && strings.HasSuffix(h.Name, "id.json") {
					v := map[string]interface{}{}
					if err = json.NewDecoder(tr).Decode(&v); err != nil {
						return
					}
					if i, ok := v["pubKey"]; ok {
						l.bootstrapId = i.(string)
					}
					fu.Verbose("bootstap node ID: %v", l.bootstrapId)
					return
				}
			}
		}

		return errstr.New("failed to get bootstrap id")
	}

	return
}

func (l *Localnet) waitForMinerJson() (err error) {
	fu.Verbose("Waiting for bootstrap node json port")
	for i:=0 ;; i++ {
		c := japi.Remote{
			Endpoint: fmt.Sprintf("%s:%d",l.MasterIP,l.Json),
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