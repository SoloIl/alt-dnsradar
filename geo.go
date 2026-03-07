package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

var ErrGeoRateLimited = errors.New("geo rate limit exceeded")

type IPInfo struct {
	IP      string `json:"ip"`
	City    string `json:"city"`
	Region  string `json:"region"`
	Country string `json:"country"`
	Org     string `json:"org"`
}

type GeoResult struct {
	IP      string
	Latency int64
	TLS     TLSStatus
	City    string
	Country string
	ASN     string
	CDN     string
}

func lookupGeo(ip string) (*IPInfo, error) {
	url := fmt.Sprintf("https://ipinfo.io/%s/json", ip)

	client := http.Client{
		Timeout: 3 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusTooManyRequests:
		return nil, ErrGeoRateLimited
	default:
		return nil, fmt.Errorf("geo lookup for %s returned status %d", ip, resp.StatusCode)
	}

	var data IPInfo
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("decode geo response for %s: %w", ip, err)
	}

	return &data, nil
}

func detectCDN(org string) string {
	switch {
	case strings.Contains(org, "Facebook"):
		return "Meta"
	case strings.Contains(org, "Google"):
		return "Google"
	case strings.Contains(org, "Cloudflare"):
		return "Cloudflare"
	case strings.Contains(org, "Amazon"):
		return "Amazon"
	case strings.Contains(org, "Microsoft"):
		return "Microsoft"
	default:
		return "-"
	}
}

func parseASN(org string) string {
	fields := strings.Fields(org)

	if len(fields) > 0 && strings.HasPrefix(fields[0], "AS") {
		return fields[0]
	}

	return "-"
}

func buildGeoResults(results []IPResult, top int) []GeoResult {
	var geo []GeoResult
	rateLimitWarned := false

	for i := 0; i < top && i < len(results); i++ {
		r := results[i]

		info, err := lookupGeo(r.IP)
		if err != nil {
			if errors.Is(err, ErrGeoRateLimited) && !rateLimitWarned {
				log.Println("Geo lookup rate limit reached; location fields may be incomplete")
				rateLimitWarned = true
			}

			geo = append(geo, GeoResult{
				IP:      r.IP,
				Latency: r.TCPLatency.Milliseconds(),
				TLS:     TLSStatusSkip,
				City:    "-",
				Country: "-",
				ASN:     "-",
				CDN:     "-",
			})
			continue
		}

		asn := parseASN(info.Org)
		cdn := detectCDN(info.Org)

		geo = append(geo, GeoResult{
			IP:      r.IP,
			Latency: r.TCPLatency.Milliseconds(),
			TLS:     TLSStatusSkip,
			City:    valueOrDash(info.City),
			Country: valueOrDash(info.Country),
			ASN:     asn,
			CDN:     cdn,
		})
	}

	populateTLSStatuses(geo)

	return geo
}

func populateTLSStatuses(results []GeoResult) {
	if len(results) == 0 {
		return
	}

	workers := len(results)
	semaphore := make(chan struct{}, workers)

	var wg sync.WaitGroup

	for i := range results {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(idx int) {
			defer func() {
				<-semaphore
				wg.Done()
			}()

			results[idx].TLS = testTLSProbe(results[idx].IP, *flagSettings.URL)
		}(i)
	}

	wg.Wait()
}

func printGeoTable(results []IPResult, top int) {
	if len(results) == 0 {
		fmt.Println("\nNo reachable edges found")
		return
	}

	fmt.Printf(
		"\nPreparing top endpoint table for %s (geo lookup + TLS diagnostics)...\n",
		*flagSettings.URL,
	)

	geo := buildGeoResults(results, top)

	fmt.Printf("\nTop fastest endpoints for %s\n\n", *flagSettings.URL)
	fmt.Printf("%-16s %-6s %-8s %-14s %-8s %s\n",
		"IP", "TCP", "TLS", "CDN", "ASN", "LOCATION")
	fmt.Println(strings.Repeat("-", 80))

	for _, g := range geo {
		location := fmt.Sprintf("%-4s %s", g.Country, g.City)
		fmt.Printf("%-16s %-6s %-8s %-14s %-8s %s\n",
			g.IP,
			fmt.Sprintf("%dms", g.Latency),
			g.TLS,
			g.CDN,
			g.ASN,
			location,
		)
	}
}

func valueOrDash(v string) string {
	if strings.TrimSpace(v) == "" {
		return "-"
	}
	return v
}
