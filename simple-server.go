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
	"syscall"
	_ "net/http/pprof"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var memprofile = flag.String("memprofile", "", "write mem profile to file")
var lockprofile = flag.String("lockprofile", "", "write lock profile to file")
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
	flag.Parse()
	runtime.GOMAXPROCS(*nprocs)

    if *cpuprofile != "" {
        f, err := os.Create(*cpuprofile)
        if err != nil {
            log.Fatal(err)
        }
        pprof.StartCPUProfile(f)
    }

	if *lockprofile != "" {
		prof, err := os.Create(*lockprofile)
		if err != nil {
			log.Fatal(err)
		}
		runtime.SetBlockProfileRate(1)
		defer func() {
			pprof.Lookup("block").WriteTo(prof, 0)
			prof.Close()
		}()
	}
	s := NewServer(*port)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGQUIT)
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
	x := <-interrupt
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
	if x == syscall.SIGQUIT {
		pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
	}
	fmt.Println("Caught signal")
	os.Exit(0)
}
