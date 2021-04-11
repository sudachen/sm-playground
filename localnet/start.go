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

// Start local network
func (l *Localnet) Start() (err error) {

	l.ctx = context.Background()
	if l.docker, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation()); err != nil {
		return
	}

	if ok, err := l.started(); ok {
		return errstr.New("already started")
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
	delay := 10+l.HareDuration
	l.genesis = time.Now().Add(time.Duration(delay)*time.Second).Format(time.RFC3339)
	if err = l.startPoet(delay); err != nil {
		return
	}
	if err = l.writeMinerConfig(); err != nil {
		return
	}
	if err = l.startMiner(0); err != nil {
		return
	}
	if err = l.waitForMinerJson(); err != nil { return }
	if err = l.activatePoet(); err != nil {
		return
	}
	for i := 1; i < l.Count; i++ {
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



