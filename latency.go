package main

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

func validateIPs(ctx context.Context, ips []string) []IPResult {
	if len(ips) == 0 {
		return nil
	}

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

	go func() {
		<-ctx.Done()
		bar.Abort(true)
	}()

	for w := 0; w < workers; w++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case ip, ok := <-jobs:
					if !ok {
						return
					}

					lat := measureMedianLatency(ctx, ip, probes, timeout)

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
			}
		}()
	}

	go func() {
		defer close(jobs)

		for _, ip := range ips {
			select {
			case <-ctx.Done():
				return
			case jobs <- ip:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var out []IPResult
	for result := range results {
		out = append(out, result)
	}

	if ctx.Err() != nil {
		bar.Abort(true)
	}
	p.Wait()

	return out
}

func measureMedianLatency(ctx context.Context, ip string, probes int, timeout time.Duration) time.Duration {
	var samples []time.Duration

	for i := 0; i < probes; i++ {
		select {
		case <-ctx.Done():
			return 0
		default:
		}

		latency, err := measureTCPConnect(ctx, ip, timeout)
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
