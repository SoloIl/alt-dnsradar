package main

import (
	"net"
	"sort"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

func validateIPs(ips []string) []IPResult {
	timeout := requestTimeout()
	workers := 20
	probes := 3

	jobs := make(chan string)
	results := make(chan IPResult)

	var wg sync.WaitGroup

	p := mpb.New(mpb.WithWidth(50))
	bar := p.New(int64(len(ips)),
		mpb.BarStyle().Rbound("|"),
		mpb.PrependDecorators(
			decor.Name("TCP latency "),
			decor.CountersNoUnit("%d/%d"),
		),
		mpb.AppendDecorators(
			decor.Percentage(),
		),
	)

	for w := 0; w < workers; w++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for ip := range jobs {
				lat := measureMedianLatency(ip, probes, timeout)

				if lat == 0 {
					results <- IPResult{
						IP:    ip,
						Alive: false,
					}
				} else {
					results <- IPResult{
						IP:         ip,
						TCPLatency: lat,
						Alive:      true,
					}
				}

				bar.Increment()
			}
		}()
	}

	go func() {
		for _, ip := range ips {
			jobs <- ip
		}
		close(jobs)
	}()

	var out []IPResult
	for i := 0; i < len(ips); i++ {
		out = append(out, <-results)
	}

	wg.Wait()
	p.Wait()

	return out
}

func measureTCPConnect(ip string, timeout time.Duration) (time.Duration, error) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", ip+":443", timeout)
	if err != nil {
		return 0, err
	}

	_ = conn.Close()
	return time.Since(start), nil
}

func measureMedianLatency(ip string, probes int, timeout time.Duration) time.Duration {
	var samples []time.Duration

	for i := 0; i < probes; i++ {
		latency, err := measureTCPConnect(ip, timeout)
		if err != nil {
			continue
		}

		samples = append(samples, latency)
	}

	if len(samples) == 0 {
		return 0
	}

	return medianDuration(samples)
}

func medianDuration(d []time.Duration) time.Duration {
	cp := make([]time.Duration, len(d))
	copy(cp, d)

	sort.Slice(cp, func(i, j int) bool { return cp[i] < cp[j] })

	n := len(cp)
	mid := n / 2

	if n%2 == 1 {
		return cp[mid]
	}

	return (cp[mid-1] + cp[mid]) / 2
}
