package main

import (
	"os"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net"
	"net/rpc"
	"runtime"
	"runtime/pprof"
	"os/signal"
	_ "net/http/pprof"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var memprofile = flag.String("memprofile", "", "write mem profile to file")
var port = flag.Int("port", 8000, "port")
var nprocs = flag.Int("nprocs", 2, "GOMAXPROCS default 2")

type Simple struct {
	port int
	listener net.Listener
}

func (s *Simple) Nothing(req, rep *struct{}) error {
	return nil
}

func NewServer(port int) *Simple {
	c := &Simple{port, nil}
	go c.run()
	return c
}

func (c *Simple) waitForConnections(rpcs *rpc.Server) {
	for {
		conn, err := c.listener.Accept()
		if err == nil {
			go rpcs.ServeConn(conn)
		} else {
			// handle error
			//fmt.Println("ERROR: ", err)
		}
	}
}

func (c *Simple) run() {
	rpcs := rpc.NewServer()
	rpcs.Register(c)

	var err error
	addr := fmt.Sprintf("localhost:%d", c.port)
	c.listener, err = net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Listen error: %v\n", err)
	}
	go c.waitForConnections(rpcs)
}

func main() {
	runtime.GOMAXPROCS(*nprocs)
	flag.Parse()

    if *cpuprofile != "" {
        f, err := os.Create(*cpuprofile)
        if err != nil {
            log.Fatal(err)
        }
        pprof.StartCPUProfile(f)
    }

	s := NewServer(*port)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt)
	go catchKill(interrupt)

	fmt.Println("Started server")
	_ = s

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", *port+1000))
	if err != nil {
		log.Fatal("listen error:", err)
	}
	http.Serve(l, nil)
}

// Dump profiling information and stats before exiting.
func catchKill(interrupt chan os.Signal) {
	<-interrupt
	if *cpuprofile != "" {
		pprof.StopCPUProfile()
	}
	if *memprofile != "" {
        f, err := os.Create(*memprofile)
        if err != nil {
            log.Fatal(err)
        }
        pprof.WriteHeapProfile(f)
    }
	fmt.Println("Caught signal")
	os.Exit(0)
}
