package localnet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"io/ioutil"
	"net/http"
	"sudachen.xyz/pkg/errstr"
	"sudachen.xyz/pkg/localnet/fu"
)

func (l *Localnet) startPoet(delay int) (err error) {

	b := bytes.Buffer{}
	b.WriteString("jsonlog = 1\n")
	b.WriteString(fmt.Sprintf("restlisten = %s:%d\n",l.PoetIP,l.Poet))
	b.WriteString("[Service]\n")
	b.WriteString(fmt.Sprintf("n = %d\n", l.Complexity))
	b.WriteString(fmt.Sprintf("duration = %ds\n",l.LayerDuration*l.LayersPerEpoch))
	b.WriteString(fmt.Sprintf("initialduration = %ds\n",l.LayerDuration*l.LayersPerEpoch/2+delay))
	err = ioutil.WriteFile("/tmp/spacemesh-poet-config.conf",b.Bytes(),0644)

	ports := nat.PortSet{}
	k, _ := nat.NewPort("tcp", fmt.Sprint(l.Poet))
	ports[k] = struct{}{}

	body, err := l.docker.ContainerCreate(l.ctx,
		&container.Config{
			Image: l.PoetImage,
			Cmd: []string{"--configfile=/config.conf"},
			Labels: l.PoetLabels,
			ExposedPorts: ports,
		},
		&container.HostConfig{
			Binds: []string{
				"/tmp/spacemesh-poet-config.conf:/config.conf:ro",
			},
			//AutoRemove: true,
		},
		&network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				l.NetworkName : {
					Aliases: []string{"poet"},
					IPAMConfig: &network.EndpointIPAMConfig{
						IPv4Address: l.PoetIP,
					},
				},
			},
		},
		nil,
		"spacemesh_poet")
	if err != nil {
		return
	}
	err = l.docker.ContainerStart(l.ctx, body.ID, types.ContainerStartOptions{})
	if err != nil { return }


	return l.waitFor("spacemesh_poet")
}

func (l *Localnet) activatePoet() (err error) {
	fu.Verbose("Activating poet server")
	ht := &http.Client{}
	bs, err := json.Marshal(struct{GatewayAddresses []string `json:"gatewayAddresses"`}{
		[]string{fmt.Sprintf("%s:%d",l.MasterIP,l.Grpc)},
	})
	if err != nil { return }
	req, err := http.NewRequest("POST",
		fmt.Sprintf("http://%s:%d/v1/start",l.PoetIP,l.Poet),
		bytes.NewBuffer(bs))
	if err != nil { return }
	req.Header.Set("Content-Type", "application/json")
	res, err := ht.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	var r map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&r)
	if err != nil { return }
	if v, ok := r["error"]; ok {
		err = errstr.New(v.(string))
	}
	return
}

