package main

import (
	"context"
	"fmt"
	"net"
	"slices"
	"strings"
	"time"

	"github.com/AdguardTeam/dnsproxy/upstream"
	"github.com/miekg/dns"
)

const dnsDiagnosticAttempts = 2

func diagnoseDNS(ctx context.Context, domain string) {
	fmt.Println(msgDNSDiagnosticsTitle())
	fmt.Println("----------------------------")

	local := lookupLocalDNS(domain)
	udp := lookupGoogleUDP(domain)
	googleDoH := lookupGoogleDoH(domain)
	cloudflareDoH := lookupCloudflareDoH(domain)

	if ctx.Err() != nil {
		return
	}

	rows := buildInitialDiagRows(ctx, domain, local, udp, googleDoH, cloudflareDoH)
	if ctx.Err() != nil {
		return
	}
	printInitialDiagTable(rows)

	comparisons := []DNSComparison{
		compareIPSetsDetailed("Local DNS", local, "Google DoH", googleDoH),
		compareIPSetsDetailed("Google UDP", udp, "Google DoH", googleDoH),
		compareIPSetsDetailed("Cloudflare DoH", cloudflareDoH, "Google DoH", googleDoH),
	}

	if ctx.Err() != nil {
		return
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
	return lookupWithRetry(func() []string {
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
	})
}

func lookupGoogleDoH(domain string) []string {
	return lookupDoH(domain, "https://dns.google/dns-query")
}

func lookupCloudflareDoH(domain string) []string {
	return lookupDoH(domain, "https://cloudflare-dns.com/dns-query")
}

func lookupDoH(domain string, endpoint string) []string {
	return lookupWithRetry(func() []string {
		u, err := upstream.AddressToUpstream(
			endpoint,
			&upstream.Options{
				Timeout: requestTimeout(),
				Logger:  quietUpstreamLogger,
			},
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
	})
}

func lookupWithRetry(fn func() []string) []string {
	for attempt := 0; attempt < dnsDiagnosticAttempts; attempt++ {
		ips := normalizeSet(fn())
		if len(ips) > 0 {
			return ips
		}
	}

	return nil
}

func compareIPSetsDetailed(leftName string, left []string, rightName string, right []string) DNSComparison {
	left = normalizeSet(left)
	right = normalizeSet(right)

	if len(left) == 0 || len(right) == 0 {
		return DNSComparison{
			LeftName:   leftName,
			RightName:  rightName,
			LeftCount:  len(left),
			RightCount: len(right),
			Relation:   RelationUnavailable,
		}
	}

	if sameSet(left, right) {
		return DNSComparison{
			LeftName:   leftName,
			RightName:  rightName,
			LeftCount:  len(left),
			RightCount: len(right),
			Relation:   RelationExactMatch,
		}
	}

	if hasOverlap(left, right) {
		return DNSComparison{
			LeftName:   leftName,
			RightName:  rightName,
			LeftCount:  len(left),
			RightCount: len(right),
			Relation:   RelationPartialOverlap,
		}
	}

	return DNSComparison{
		LeftName:   leftName,
		RightName:  rightName,
		LeftCount:  len(left),
		RightCount: len(right),
		Relation:   RelationNoOverlap,
	}
}

func buildInitialDiagRows(ctx context.Context, domain string, local []string, udp []string, googleDoH []string, cloudflareDoH []string) []InitialDiagRow {
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

	uniqueIPs := make(map[string]struct{})
	for _, source := range sources {
		for _, ip := range normalizeSet(source.ips) {
			uniqueIPs[ip] = struct{}{}
		}
	}

	ipsToProbe := make([]string, 0, len(uniqueIPs))
	for ip := range uniqueIPs {
		ipsToProbe = append(ipsToProbe, ip)
	}
	slices.Sort(ipsToProbe)

	printInitialProbeStatus(len(ipsToProbe))
	probeCache := probeUniqueEndpointsParallel(ctx, ipsToProbe, domain, requestTimeout(), 4)

	var rows []InitialDiagRow

	for _, source := range sources {
		ips := normalizeSet(source.ips)
		if len(ips) == 0 {
			rows = append(rows, InitialDiagRow{
				Source:    source.name,
				IP:        "-",
				TCPAlive:  false,
				TLSStatus: TLSStatusSkip,
				Note:      msgNoIPv4Answers(),
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
				row.TCPAlive = cached.TCPAlive
				row.TCPLatency = cached.TCPLatency
				row.TLSStatus = cached.TLSStatus
			}
			rows = append(rows, row)
		}
	}

	return rows
}

func printInitialDiagTable(rows []InitialDiagRow) {
	fmt.Println("")
	fmt.Println(msgInitialEndpointDiagnosticsTitle())
	fmt.Println("")
	fmt.Printf("%-17s %-16s %-9s %-8s %s\n", "SOURCE", "IP", "TCP", "TLS", "NOTE")
	fmt.Println(strings.Repeat("-", 80))

	for _, row := range rows {
		fmt.Printf(
			"%-17s %-16s %s %s %s\n",
			row.Source,
			row.IP,
			colorizeTCPLabel(initialTCPLabel(row)),
			colorizeTLSStatus(row.TLSStatus),
			row.Note,
		)
	}
}

func printDNSSummary(comparisons []DNSComparison) {
	fmt.Println("")
	fmt.Println(msgDNSSummaryTitle())

	for _, comparison := range comparisons {
		fmt.Printf("- %s\n", dnsSummaryLine(comparison))
	}
}

func dnsSummaryLine(comparison DNSComparison) string {
	switch comparison.LeftName {
	case "Google UDP":
		switch comparison.Relation {
		case RelationUnavailable:
			return msgGoogleUDPAndGoogleDoHUnavailable()
		case RelationExactMatch:
			return msgGoogleUDPConsistent()
		case RelationPartialOverlap:
			return msgGoogleUDPPartialDiffers()
		case RelationNoOverlap:
			if isMultiEndpointMismatch(comparison) {
				return msgGoogleUDPMultiSetMismatch()
			}
			return msgGoogleUDPStrongMismatch()
		}
	case "Local DNS":
		switch comparison.Relation {
		case RelationUnavailable:
			return msgLocalDNSUnavailable()
		case RelationExactMatch:
			return msgLocalDNSConsistent()
		case RelationPartialOverlap:
			return msgLocalDNSPartialDiffers()
		case RelationNoOverlap:
			if isMultiEndpointMismatch(comparison) {
				return msgLocalDNSMultiSetMismatch()
			}
			return msgLocalDNSStrongMismatch()
		}
	case "Cloudflare DoH":
		switch comparison.Relation {
		case RelationUnavailable:
			return msgCloudflareDoHUnavailable()
		case RelationExactMatch:
			return msgCloudflareDoHConsistent()
		case RelationPartialOverlap:
			return msgCloudflareDoHPartialDiffers()
		case RelationNoOverlap:
			if isMultiEndpointMismatch(comparison) {
				return msgCloudflareDoHMultiSetMismatch()
			}
			return msgCloudflareDoHStrongMismatch()
		}
	}

	return comparison.LeftName + " comparison unavailable"
}

func isMultiEndpointMismatch(comparison DNSComparison) bool {
	return comparison.LeftCount > 1 || comparison.RightCount > 1
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
		return msgReferenceUnavailable()
	}

	switch source {
	case "Google DoH", "Cloudflare DoH":
		return msgReference()
	case "Local DNS", "Google UDP":
		if _, ok := refSet[ip]; ok {
			return msgSharedWithDoHReference()
		}
		return msgNotInDoHReference()
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
