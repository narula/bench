package main

import (
	"flag"
	"fmt"
	"log"
	"net/rpc"
	"runtime"
	"runtime/pprof"
	"time"
	"os"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var nsec = flag.Int("nsec", 2, "Number of seconds; will wait to finish all requests before completing.")
var maxOutstanding = flag.Int("nmax", 5000, "Max outstanding requests.")
var cacheServer = flag.String("cs", "localhost:8000", "Caching server host and port.")

var nops int32
var sem chan int

func read(ck *rpc.Client, reqs chan int) {
	for req := range reqs {
		<-sem
		go func(req int) {
			ck.Call("Server.Nothing", &struct{}{}, &struct{}{})
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
	for i:=0; i < *maxOutstanding; i++ {
		sem <- 1
	}

	ck, err := rpc.Dial("tcp", *cacheServer)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	reqs := make(chan int)
	done := time.NewTimer(time.Duration(*nsec)*time.Second).C
	start := time.Now()

	go read(ck, reqs)
outer:
	for {
		select {
		case <- done:
			break outer
		default:
			reqs <- 1
			nops++
		}
	}
	close(reqs)
	for i:=0; i < *maxOutstanding; i++ {
		<-sem
	}
	x := time.Since(start)
	fmt.Printf("nops: %v, %v/sec actual time: %v.\n", nops, float64(nops)/x.Seconds(), x)

	if *cpuprofile != "" {
		pprof.StopCPUProfile()
		fmt.Println("Stopping...")
	}
}
