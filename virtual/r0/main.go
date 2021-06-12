package main

import (
	"fmt"
	"github.com/spacemeshos/go-spacemesh/cmd"
	"github.com/spacemeshos/go-spacemesh/cmd/node"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/debug"
	"time"
)

var (
	version string
	commit  string
	branch  string
)

func main() { // run the app
	const Mb uint64 = 1024*1024
	const Ballast = 2*1024*1024*1024
	debug.SetGCPercent(10) // start GC when 200Mb+ is allocated
	ballast := make([]byte,Ballast)
	cmd.Version = version
	cmd.Commit = commit
	cmd.Branch = branch

	go func() {
		tk := time.NewTicker(30*time.Second)
		for {
			select {
			case <- tk.C:
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				//var MiB uint64 = 1000*1000
				fmt.Fprintf(os.Stderr, "MemStats: Total %v Mb, Active %v Mb, Heap %v/%v Mb, Stack %v/%v Mb, Span %v/%v Mb, Cache %v/%v Mb, Buck %v Mb, GC %v Mb, Other %v Mb, Ballast %v Mb, Wtf? %v Mb\n",
					m.Sys/Mb,
					(m.Sys-Ballast)/Mb,
					(m.HeapInuse-Ballast)/Mb, (m.HeapSys-Ballast)/Mb,
					m.StackInuse/Mb, m.StackSys/Mb,
					m.MSpanInuse/Mb, m.MSpanSys/Mb,
					m.MCacheInuse/Mb, m.MCacheSys/Mb,
					m.BuckHashSys/Mb,
					m.GCSys/Mb,
					m.OtherSys/Mb,
					Ballast/Mb,
					(m.Sys-(m.OtherSys+m.GCSys+m.BuckHashSys+m.MCacheSys+m.MSpanSys+m.StackSys+m.HeapSys))/Mb)
				fmt.Fprintf(os.Stderr, "Current goroutines number is %d\n", runtime.NumGoroutine())
			}
		}
	}()

	go http.ListenAndServe("0.0.0.0:6060", nil)

	if err := node.Cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	runtime.KeepAlive(ballast)
}


