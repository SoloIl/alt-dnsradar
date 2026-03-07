package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

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

	var data IPInfo
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
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

	for i := 0; i < top && i < len(results); i++ {
		r := results[i]

		info, err := lookupGeo(r.IP)
		if err != nil {
			geo = append(geo, GeoResult{
				IP:      r.IP,
				Latency: r.TCPLatency.Milliseconds(),
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
			City:    valueOrDash(info.City),
			Country: valueOrDash(info.Country),
			ASN:     asn,
			CDN:     cdn,
		})
	}

	return geo
}

func printGeoTable(results []IPResult, top int) {
	if len(results) == 0 {
		fmt.Println("\nNo reachable edges found")
		return
	}

	geo := buildGeoResults(results, top)

	fmt.Println("\nTop fastest edges\n")
	fmt.Printf("%-16s %-6s %-14s %-8s %-7s %-10s\n",
		"IP", "TCP", "CDN", "ASN", "COUNTRY", "CITY")
	fmt.Println(strings.Repeat("-", 70))

	for _, g := range geo {
		fmt.Printf("%-16s %-6s %-14s %-8s %-7s %-10s\n",
			g.IP,
			fmt.Sprintf("%dms", g.Latency),
			g.CDN,
			g.ASN,
			g.Country,
			g.City,
		)
	}
}

func valueOrDash(v string) string {
	if strings.TrimSpace(v) == "" {
		return "-"
	}
	return v
}
