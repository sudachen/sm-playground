package localnet

import (
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
)

func (l *Localnet) createNetwork() (err error) {
	networks, err := l.docker.NetworkList(l.ctx, types.NetworkListOptions{
		Filters:filters.NewArgs(
			filters.Arg("name",l.NetworkName)),
	})

	if err != nil {
		return
	}

	if len(networks) == 0 {
		// create network
		_, err = l.docker.NetworkCreate(l.ctx, l.NetworkName, types.NetworkCreate{
			Driver:     "bridge",
			EnableIPv6: false,
			IPAM: &network.IPAM{
				Config: []network.IPAMConfig{{
					Subnet: l.Subnet(),
					Gateway: l.Gateway(),
				}},
			},
		})
	}
	return
}

func (l *Localnet) destroyNetwork() (err error) {
	networks, err := l.docker.NetworkList(l.ctx, types.NetworkListOptions{
		Filters:filters.NewArgs(
			filters.Arg("name",l.NetworkName)),
	})
	for _, n := range networks {
		fmt.Println(n)
		fmt.Printf("removing network %v\n", n.Name)
		err = l.docker.NetworkRemove(l.ctx,n.ID)
		if err != nil {
			fmt.Println("\t can't remove network: %v",err.Error())
			break
		}
	}
	return
}

