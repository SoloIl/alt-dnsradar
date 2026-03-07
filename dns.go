package main

import (
	"fmt"
	"net"
	"slices"
	"strings"
	"time"

	"github.com/AdguardTeam/dnsproxy/upstream"
	"github.com/miekg/dns"
)

func diagnoseDNS(domain string) {
	fmt.Println("DNS diagnostics")
	fmt.Println("----------------------------")

	local := lookupLocalDNS(domain)
	udp := lookupGoogleUDP(domain)
	googleDoH := lookupGoogleDoH(domain)
	cloudflareDoH := lookupCloudflareDoH(domain)

	rows := buildInitialDiagRows(domain, local, udp, googleDoH, cloudflareDoH)
	printInitialDiagTable(rows)

	comparisons := []DNSComparison{
		compareIPSetsDetailed("Local DNS", local, "Google DoH", googleDoH),
		compareIPSetsDetailed("Google UDP", udp, "Google DoH", googleDoH),
		compareIPSetsDetailed("Cloudflare DoH", cloudflareDoH, "Google DoH", googleDoH),
	}

	printDNSSummary(comparisons)

	fmt.Println("")
	fmt.Println("-------------------------------------------")
	fmt.Println("")
}

func lookupLocalDNS(domain string) []string {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return nil
	}

	var out []string
	for _, ip := range ips {
		if ip.To4() != nil {
			out = append(out, ip.String())
		}
	}

	return normalizeSet(out)
}

func lookupGoogleUDP(domain string) []string {
	c := dns.Client{Timeout: requestTimeout()}
	m := dns.Msg{}
	m.SetQuestion(dns.Fqdn(domain), dns.TypeA)

	r, _, err := c.Exchange(&m, "8.8.8.8:53")
	if err != nil {
		return nil
	}

	var ips []string
	for _, ans := range r.Answer {
		if a, ok := ans.(*dns.A); ok {
			ips = append(ips, a.A.String())
		}
	}

	return normalizeSet(ips)
}

func lookupGoogleDoH(domain string) []string {
	return lookupDoH(domain, "https://dns.google/dns-query")
}

func lookupCloudflareDoH(domain string) []string {
	return lookupDoH(domain, "https://cloudflare-dns.com/dns-query")
}

func lookupDoH(domain string, endpoint string) []string {
	u, err := upstream.AddressToUpstream(
		endpoint,
		&upstream.Options{Timeout: requestTimeout()},
	)
	if err != nil {
		return nil
	}
	defer u.Close()

	req := dns.Msg{}
	req.SetQuestion(dns.Fqdn(domain), dns.TypeA)

	resp, err := u.Exchange(&req)
	if err != nil {
		return nil
	}

	var ips []string
	for _, ans := range resp.Answer {
		if a, ok := ans.(*dns.A); ok {
			ips = append(ips, a.A.String())
		}
	}

	return normalizeSet(ips)
}

func compareIPSetsDetailed(leftName string, left []string, rightName string, right []string) DNSComparison {
	left = normalizeSet(left)
	right = normalizeSet(right)

	if len(left) == 0 || len(right) == 0 {
		return DNSComparison{
			LeftName:  leftName,
			RightName: rightName,
			Relation:  RelationUnavailable,
		}
	}

	if sameSet(left, right) {
		return DNSComparison{
			LeftName:  leftName,
			RightName: rightName,
			Relation:  RelationExactMatch,
		}
	}

	if hasOverlap(left, right) {
		return DNSComparison{
			LeftName:  leftName,
			RightName: rightName,
			Relation:  RelationPartialOverlap,
		}
	}

	return DNSComparison{
		LeftName:  leftName,
		RightName: rightName,
		Relation:  RelationNoOverlap,
	}
}

func buildInitialDiagRows(domain string, local []string, udp []string, googleDoH []string, cloudflareDoH []string) []InitialDiagRow {
	sources := []struct {
		name string
		ips  []string
	}{
		{name: "Local DNS", ips: local},
		{name: "Google UDP", ips: udp},
		{name: "Google DoH", ips: googleDoH},
		{name: "Cloudflare DoH", ips: cloudflareDoH},
	}

	refSet := make(map[string]struct{}, len(googleDoH)+len(cloudflareDoH))
	for _, ip := range normalizeSet(googleDoH) {
		refSet[ip] = struct{}{}
	}
	for _, ip := range normalizeSet(cloudflareDoH) {
		refSet[ip] = struct{}{}
	}

	type diagProbeResult struct {
		tcpAlive bool
		latency  time.Duration
		tls      TLSStatus
	}

	probeCache := make(map[string]diagProbeResult)

	var rows []InitialDiagRow

	for _, source := range sources {
		ips := normalizeSet(source.ips)
		if len(ips) == 0 {
			rows = append(rows, InitialDiagRow{
				Source:    source.name,
				IP:        "-",
				TCPAlive:  false,
				TLSStatus: TLSStatusSkip,
				Note:      "no IPv4 answers",
			})
			continue
		}

		for _, ip := range ips {
			row := InitialDiagRow{
				Source:    source.name,
				IP:        ip,
				TLSStatus: TLSStatusSkip,
				Note:      initialDiagNote(source.name, ip, refSet),
			}

			if cached, ok := probeCache[ip]; ok {
				row.TCPAlive = cached.tcpAlive
				row.TCPLatency = cached.latency
				row.TLSStatus = cached.tls
				rows = append(rows, row)
				continue
			}

			probe := diagProbeResult{
				tcpAlive: false,
				latency:  0,
				tls:      TLSStatusSkip,
			}

			latency, err := measureTCPConnect(ip, requestTimeout())
			if err == nil {
				probe.tcpAlive = true
				probe.latency = latency
				probe.tls = testTLSProbe(ip, domain)
			}

			probeCache[ip] = probe
			row.TCPAlive = probe.tcpAlive
			row.TCPLatency = probe.latency
			row.TLSStatus = probe.tls
			rows = append(rows, row)
		}
	}

	return rows
}

func printInitialDiagTable(rows []InitialDiagRow) {
	fmt.Println("")
	fmt.Println("Initial endpoint diagnostics")
	fmt.Println("")
	fmt.Printf("%-17s %-16s %-8s %-8s %s\n", "SOURCE", "IP", "TCP", "TLS", "NOTE")
	fmt.Println(strings.Repeat("-", 80))

	for _, row := range rows {
		fmt.Printf(
			"%-17s %-16s %-8s %-8s %s\n",
			row.Source,
			row.IP,
			initialTCPLabel(row),
			row.TLSStatus,
			row.Note,
		)
	}
}

func printDNSSummary(comparisons []DNSComparison) {
	fmt.Println("")
	fmt.Println("DNS diagnostic summary")

	for _, comparison := range comparisons {
		fmt.Printf("- %s\n", dnsSummaryLine(comparison))
	}
}

func dnsSummaryLine(comparison DNSComparison) string {
	switch comparison.LeftName {
	case "Google UDP":
		switch comparison.Relation {
		case RelationUnavailable:
			return "Google UDP and Google DoH comparison unavailable"
		case RelationExactMatch:
			return "Google UDP is consistent with Google DoH"
		case RelationPartialOverlap:
			return "Google UDP partially differs from Google DoH; possible cache or CDN variance"
		case RelationNoOverlap:
			return "Strong mismatch between Google UDP and Google DoH; possible DNS interception"
		}
	case "Local DNS":
		switch comparison.Relation {
		case RelationUnavailable:
			return "Local DNS comparison with Google DoH unavailable"
		case RelationExactMatch:
			return "Local DNS is consistent with Google DoH"
		case RelationPartialOverlap:
			return "Local DNS partially differs from Google DoH"
		case RelationNoOverlap:
			return "Local DNS strongly differs from Google DoH"
		}
	case "Cloudflare DoH":
		switch comparison.Relation {
		case RelationUnavailable:
			return "Cloudflare DoH and Google DoH comparison unavailable"
		case RelationExactMatch:
			return "Cloudflare DoH is consistent with Google DoH"
		case RelationPartialOverlap:
			return "Cloudflare DoH partially differs from Google DoH; possible CDN or cache variance"
		case RelationNoOverlap:
			return "Strong mismatch between Cloudflare DoH and Google DoH; reference confidence is lower"
		}
	}

	return comparison.LeftName + " comparison unavailable"
}

func normalizeSet(in []string) []string {
	if len(in) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(in))
	var out []string

	for _, ip := range in {
		if _, ok := seen[ip]; ok {
			continue
		}

		seen[ip] = struct{}{}
		out = append(out, ip)
	}

	slices.Sort(out)
	return out
}

func sameSet(a []string, b []string) bool {
	a = normalizeSet(a)
	b = normalizeSet(b)

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func hasOverlap(a []string, b []string) bool {
	set := make(map[string]struct{}, len(a))
	for _, ip := range normalizeSet(a) {
		set[ip] = struct{}{}
	}

	for _, ip := range normalizeSet(b) {
		if _, ok := set[ip]; ok {
			return true
		}
	}

	return false
}

func initialDiagNote(source string, ip string, refSet map[string]struct{}) string {
	if source != "Google DoH" && source != "Cloudflare DoH" && len(refSet) == 0 {
		return "reference unavailable"
	}

	switch source {
	case "Google DoH", "Cloudflare DoH":
		return "reference"
	case "Local DNS", "Google UDP":
		if _, ok := refSet[ip]; ok {
			return "shared with DoH reference"
		}
		return "not in DoH reference"
	default:
		return "-"
	}
}

func initialTCPLabel(row InitialDiagRow) string {
	if row.IP == "-" {
		return "-"
	}

	if !row.TCPAlive {
		return "BLOCKED"
	}

	return row.TCPLatency.Round(time.Millisecond).String()
}
