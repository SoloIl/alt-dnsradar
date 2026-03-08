package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"
)

func tlsDiagnosticTimeout() time.Duration {
	return 3 * time.Second
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

func testTLSProbe(ip string, domain string) TLSStatus {
	timeout := tlsDiagnosticTimeout()

	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: timeout},
		"tcp",
		ip+":443",
		&tls.Config{
			ServerName:         domain,
			InsecureSkipVerify: true,
		},
	)

	if err != nil {
		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			return TLSStatusTimeout
		}

		return TLSStatusFail
	}

	_ = conn.Close()
	return TLSStatusOK
}

func probeEndpoint(ip string, domain string, tcpTimeout time.Duration) ProbeResult {
	result := ProbeResult{
		TCPAlive:   false,
		TCPLatency: 0,
		TLSStatus:  TLSStatusSkip,
	}

	latency := measureMedianLatency(ip, 3, tcpTimeout)
	if latency == 0 {
		return result
	}

	result.TCPAlive = true
	result.TCPLatency = latency
	result.TLSStatus = testTLSProbe(ip, domain)
	return result
}

func probeUniqueEndpointsParallel(ips []string, domain string, tcpTimeout time.Duration, workers int) map[string]ProbeResult {
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

			for job := range jobs {
				outcomes <- probeOutcome{
					ip:     job.ip,
					result: probeEndpoint(job.ip, domain, tcpTimeout),
				}
			}
		}()
	}

	go func() {
		for _, ip := range ips {
			jobs <- probeJob{ip: ip}
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
	fmt.Printf(
		"Running initial TCP/TLS diagnostics for %d unique endpoint(s)...\n",
		uniqueCount,
	)
}
