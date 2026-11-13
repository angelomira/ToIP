package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	url := flag.String("url", "http://localhost:8080/work", "target url")
	conc := flag.Int("c", 8, "concurrency")
	dur := flag.Duration("d", 30*time.Second, "duration")
	flag.Parse()

	client := &http.Client{Timeout: 60 * time.Second}
	var ok, fail int64
	deadline := time.Now().Add(*dur)

	var wg sync.WaitGroup
	start := time.Now()
	for i := 0; i < *conc; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for time.Now().Before(deadline) {
				resp, err := client.Get(*url)
				if err != nil {
					atomic.AddInt64(&fail, 1)
					continue
				}
				_, _ = io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					atomic.AddInt64(&ok, 1)
				} else {
					atomic.AddInt64(&fail, 1)
				}
			}
		}()
	}
	wg.Wait()

	el := time.Since(start).Seconds()
	fmt.Printf("requests ok=%d fail=%d in %.1fs => RPS=%.2f\n",
		ok, fail, el, float64(ok)/el)
}