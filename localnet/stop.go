package localnet

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"sudachen.xyz/pkg/errstr"
)

// Stop local network gracefully
func (l *Localnet) Stop() (err error) {
	if l.docker == nil {
		l.ctx = context.Background()
		if l.docker, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation()); err != nil {
			return
		}
	}
	if err = l.stopAll(); err != nil {
		return
	}
	return l.destroyNetwork()
}

// Terminate local network
func (l *Localnet) Terminate() (err error) {
	if l.docker == nil {
		l.ctx = context.Background()
		if l.docker, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation()); err != nil {
			return
		}
	}
	l.terminate()
	return
}

func (l *Localnet) stopAll() (err error) {
	containers, err := l.docker.ContainerList(l.ctx,
		types.ContainerListOptions{
			All: true,
			Filters: filters.NewArgs(
				filters.Arg("label",l.DockerLabel)),
		})
	if err == nil {
		for _, c := range containers {
			fmt.Printf("stopping %v\n", c.Names)
			_ = l.docker.ContainerStop(l.ctx,c.ID,nil)
			err = l.docker.ContainerRemove(l.ctx,c.ID, types.ContainerRemoveOptions{Force: true})
			if err != nil {
				return errstr.Wrapf(0, err,"can't remove container: %v",err.Error())
			}
		}
	}
	return
}

func (l *Localnet) terminate() {
	fmt.Println("terminating")
	if l.docker != nil {
		containers, err := l.docker.ContainerList(l.ctx,
			types.ContainerListOptions{
				All: true,
				Filters: filters.NewArgs(
					filters.Arg("label",l.DockerLabel)),
			})
		if err == nil {
			for _, c := range containers {
				fmt.Printf("killing %v\n", c.Names)
				_ = l.docker.ContainerKill(l.ctx,c.ID,"SIGKILL")
				err = l.docker.ContainerRemove(l.ctx,c.ID, types.ContainerRemoveOptions{Force: true})
				if err != nil {
					fmt.Printf("\tcan't remove container: %v\n",err.Error())
				}
			}
		}
		networks, err := l.docker.NetworkList(l.ctx, types.NetworkListOptions{
			Filters:filters.NewArgs(
				filters.Arg("name",l.NetworkName)),
		})
		for _, n := range networks {
			fmt.Println(n)
			//fmt.Printf("removing network %v\n", n.Name)
			//err = l.docker.NetworkRemove(l.ctx,n.ID)
			//if err != nil {
			//	fmt.Printf("\tcan't remove network: %v\n",err.Error())
			//}
		}
	}
}
