package main

import (
	"flag"
	"fmt"
	"prof"
	"sync"
)
import "time"

var nsec = flag.Int("nsec", 2, "Time to run in seconds")
var cases = flag.Int("cases", 0, "Number of case statements (0=no select loop)")

func main() {
	flag.Parse()
	var wg sync.WaitGroup
	wg.Add(1)
	m := make(map[int]int)
	for i := 0; i < 100000; i++ {
		m[i] = i
	}
	n := 0

	p := prof.StartProfile()
	start := time.Now()
	go func() {
		switch *cases {
		case -1:
			duration := time.Now().Add(time.Duration(*nsec) * time.Second)
			for duration.After(time.Now()) {
				_ = m[n%100000]
				n = n + 1
			}
			wg.Done()
			return
		case 0:
			duration := time.Now().Add(time.Duration(*nsec) * time.Second)
			for {
				select {
				default:
					_ = m[n%100000]
					n = n + 1
					if time.Now().After(duration) {
						wg.Done()
						return
					}
				}
			}

		case 1:
			done := time.NewTicker(time.Duration(*nsec) * time.Second).C
			for {
				select {
				case <-done:
					wg.Done()
					return
				default:
					_ = m[n%100000]
					n = n + 1
				}
			}
		case 2:
			done := time.NewTicker(time.Duration(*nsec) * time.Second).C
			tm := time.NewTicker(time.Duration(*nsec) * time.Millisecond * 100).C
			for {
				select {
				case <-done:
					wg.Done()
					return
				case <-tm:
					_ = m[n%100000]
					n = n + 1
				default:
					_ = m[n%100000]
					n = n + 1
				}
			}
		case 3:
			done := time.NewTicker(time.Duration(*nsec) * time.Second).C
			tm := time.NewTicker(time.Duration(*nsec) * time.Millisecond * 100).C
			tm2 := time.NewTicker(time.Duration(*nsec) * time.Millisecond * 101).C
			for {
				select {
				case <-done:
					wg.Done()
					return
				case <-tm:
					_ = m[n%100000]
					n = n + 1
				case <-tm2:
					_ = m[n%100000]
					n = n + 1
				default:
					_ = m[n%100000]
					n = n + 1
				}
			}
		}
	}()
	wg.Wait()
	end := time.Since(start)
	p.Stop()
	fmt.Printf("cases: %v \tnitr: %v \ttime: %v \trate: %v\n", *cases, n, end, float64(n)/end.Seconds())
}
