package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/ugorji/go/codec"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var nsec = flag.Int("nsec", 2, "Number of seconds; will wait to finish all requests before completing.")
var maxOutstanding = flag.Int("nmax", 5000, "Max outstanding requests.")
var cacheServer = flag.String("cs", "localhost:8000", "Caching server host and port.")
var nc = flag.Int("nc", 10, "Number of rpc.Clients")
var use_codec = flag.Bool("codec", false, "Use ugorji's codec and msgpack")
var exp = flag.String("exp", "nothing", "Experiment; echo or nothing")

var nops int32
var sem chan int

// create and configure Handle
var (
	bh codec.BincHandle
	mh codec.MsgpackHandle
)

func read(clients []*rpc.Client, reqs chan int) {
	for req := range reqs {
		<-sem
		go func(req int) {
			if *exp == "echo" {
				var x string
				err := clients[req%*nc].Call("Simple.Echo", "hi", &x)
				if err != nil || x != "hi" {
					panic(err)
				}
			} else {
				err := clients[req%*nc].Call("Simple.Nothing", &struct{}{}, &struct{}{})
				if err != nil {
					panic(err)
				}
			}
			sem <- 1
		}(req)
	}
}

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Starting...")
		pprof.StartCPUProfile(f)
	}

	nops = 0
	sem = make(chan int, *maxOutstanding)
	for i := 0; i < *maxOutstanding; i++ {
		sem <- 1
	}

	clients := make([]*rpc.Client, *nc)
	for i := 0; i < *nc; i++ {
		if *use_codec {
			conn, err := net.Dial("tcp", *cacheServer)
			if err != nil {
				log.Fatalf("error: %v\n", err)
			}
			rpcCodec := codec.MsgpackSpecRpc.ClientCodec(conn, &mh)
			clients[i] = rpc.NewClientWithCodec(rpcCodec)
		} else {
			var e error
			clients[i], e = rpc.Dial("tcp", *cacheServer)
			if e != nil {
				log.Fatalf("error: %v\n", e)
			}
		}
	}

	reqs := make(chan int)
	done := time.NewTimer(time.Duration(*nsec) * time.Second).C
	start := time.Now()

	go read(clients, reqs)
outer:
	for {
		select {
		case <-done:
			break outer
		default:
			reqs <- int(nops) // Don't care just trying to spread out clients
			nops++
		}
	}
	close(reqs)
	for i := 0; i < *maxOutstanding; i++ {
		<-sem
	}
	x := time.Since(start)
	fmt.Printf("nops: %v, %v/sec actual time: %v.\n", nops, float64(nops)/x.Seconds(), x)

	if *cpuprofile != "" {
		pprof.StopCPUProfile()
		fmt.Println("Stopping...")
	}
}
