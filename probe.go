package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"
)

func tlsDiagnosticTimeout() time.Duration {
	return 3 * time.Second
}

func measureTCPConnect(ctx context.Context, ip string, timeout time.Duration) (time.Duration, error) {
	start := time.Now()
	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "tcp", ip+":443")
	if err != nil {
		return 0, err
	}

	_ = conn.Close()
	return time.Since(start), nil
}

func testTLSProbe(ctx context.Context, ip string, domain string) TLSStatus {
	timeout := tlsDiagnosticTimeout()

	dialer := &net.Dialer{Timeout: timeout}
	rawConn, err := dialer.DialContext(ctx, "tcp", ip+":443")
	if err != nil {
		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			return TLSStatusTimeout
		}

		return TLSStatusFail
	}

	conn := tls.Client(rawConn, &tls.Config{
		ServerName:         domain,
		InsecureSkipVerify: true,
	})
	_ = rawConn.SetDeadline(time.Now().Add(timeout))

	handshakeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := conn.HandshakeContext(handshakeCtx); err != nil {
		_ = rawConn.Close()
		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			return TLSStatusTimeout
		}

		return TLSStatusFail
	}

	_ = conn.Close()
	return TLSStatusOK
}

func probeEndpoint(ctx context.Context, ip string, domain string, tcpTimeout time.Duration) ProbeResult {
	result := ProbeResult{
		TCPAlive:   false,
		TCPLatency: 0,
		TLSStatus:  TLSStatusSkip,
	}

	latency := measureMedianLatency(ctx, ip, 3, tcpTimeout)
	if latency == 0 {
		return result
	}

	result.TCPAlive = true
	result.TCPLatency = latency
	result.TLSStatus = testTLSProbe(ctx, ip, domain)
	return result
}

func probeUniqueEndpointsParallel(ctx context.Context, ips []string, domain string, tcpTimeout time.Duration, workers int) map[string]ProbeResult {
	results := make(map[string]ProbeResult)
	if len(ips) == 0 {
		return results
	}

	if workers <= 0 {
		workers = 1
	}

	if workers > len(ips) {
		workers = len(ips)
	}

	type probeJob struct {
		ip string
	}

	type probeOutcome struct {
		ip     string
		result ProbeResult
	}

	jobs := make(chan probeJob)
	outcomes := make(chan probeOutcome)

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case job, ok := <-jobs:
					if !ok {
						return
					}

					outcomes <- probeOutcome{
						ip:     job.ip,
						result: probeEndpoint(ctx, job.ip, domain, tcpTimeout),
					}
				}
			}
		}()
	}

	go func() {
		for _, ip := range ips {
			select {
			case <-ctx.Done():
				close(jobs)
				wg.Wait()
				close(outcomes)
				return
			case jobs <- probeJob{ip: ip}:
			}
		}
		close(jobs)
		wg.Wait()
		close(outcomes)
	}()

	for outcome := range outcomes {
		results[outcome.ip] = outcome.result
	}

	return results
}

func printInitialProbeStatus(uniqueCount int) {
	fmt.Printf("%s\n", msgInitialProbeStatus(uniqueCount))
}
