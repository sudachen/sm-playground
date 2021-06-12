package localnet

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"io"
	"sudachen.xyz/pkg/errstr"
	"sudachen.xyz/pkg/localnet/stdio"
	"time"
)

func (l *Localnet) ConnectDocker() (err error) {
	if l.docker == nil {
		l.ctx = context.Background()
		l.docker, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	}
	return
}

// Start local network
func (l *Localnet) Start() (err error) {
	if err = l.ConnectDocker(); err != nil {
		return
	}
	if ok, err := l.AlreadyStarted(); ok {
		return errstr.New("already AlreadyStarted")
	} else if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			//l.terminate()
		}
	}()

	if l.PullImages || !l.exists(l.MinerImage) {
		if _, err = l.docker.ImagePull(l.ctx, l.MinerImage, types.ImagePullOptions{}); err != nil {
			return err
		}
	}

	if l.PullImages || !l.exists(l.PoetImage) {
		if _, err = l.docker.ImagePull(l.ctx, l.PoetImage, types.ImagePullOptions{}); err != nil {
			return err
		}
	}

	if err = l.createNetwork(); err != nil {
		return
	}
	if err = l.stopAll(); err != nil {
		return
	}

	l.genesis = time.Now().Add(10 * time.Second)
	l.MinerLabels["genesis"] = l.Genesis()
	l.PoetLabels["genesis"] = l.Genesis()
	if err = l.writeMinerConfig(); err != nil {
		return
	}
	if err = l.startMiner(1); err != nil {
		return
	}
	if err = l.waitForMinerGrpc(l.SubnetPrefix+"1"); err != nil { return }
	if err = l.startPoet(l.genesis); err != nil {
		return
	}
	if err = l.activatePoet(); err != nil {
		return
	}
	for i := 2; i <= l.Count; i++ {
		if err = l.startMiner(i); err != nil {
			return
		}
	}
	return
}

func (l *Localnet) exists(image string) bool {
	imgs, err := l.docker.ImageList(l.ctx,types.ImageListOptions{
		Filters: filters.NewArgs(filters.Arg("reference",image)),
	})
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return len(imgs) != 0
}

func (l *Localnet) waitFor(conname string) (err error) {
loop:
	for {
		var containers []types.Container
		containers, err = l.docker.ContainerList(l.ctx, types.ContainerListOptions{
			Filters: filters.NewArgs(
				filters.Arg("name", conname)),
		})

		if err != nil {
			return
		}

		switch containers[0].State {
		default:
			var rdr io.ReadCloser
			rdr, err = l.docker.ContainerLogs(l.ctx, conname, types.ContainerLogsOptions{
				ShowStderr: true,
				ShowStdout: true,
				Tail: "10",
			})
			if err == nil {
				_, stdout, _ := stdio.StdIo()
				io.Copy(stdout,rdr)
				rdr.Close()
			}
			return errstr.Errorf("%d fialed to start: %s",conname, containers[0].State)
		case "running":
			break loop
		}
	}

	return
}

func mixLabels(a, b map[string]string) map[string]string {
	r := make(map[string]string,len(a)+len(b))
	for k,v := range a {
		r[k] = v
	}
	for k,v := range b {
		r[k] = v
	}
	return r
}
