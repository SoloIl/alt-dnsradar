package main

import (
	"log"
	"net"
	"time"

	"github.com/AdguardTeam/dnsproxy/upstream"
	"github.com/miekg/dns"
)

func diagnoseDNS(domain string) {
	log.Println("DNS diagnostics")
	log.Println("----------------------------")

	local := lookupLocalDNS(domain)
	udp := lookupGoogleUDP(domain)
	doh := lookupGoogleDoH(domain)

	log.Println("")
	log.Println("LOCAL DNS")
	printIPs(local)

	log.Println("")
	log.Println("GOOGLE DNS UDP")
	printIPs(udp)

	log.Println("")
	log.Println("GOOGLE DNS DOH")
	printIPs(doh)

	if !compareIPSets(local, doh) {
		log.Println("")
		log.Println("DNS poisoning detected")
	}

	if !compareIPSets(udp, doh) {
		log.Println("DNS interception detected")
	}

	log.Println("")
	log.Println("Testing DoH IP connectivity")
	log.Println("")

	for _, ip := range doh {
		start := time.Now()

		conn, err := net.DialTimeout("tcp", ip+":443", 2*time.Second)
		if err != nil {
			log.Printf("BLOCKED %s\n", ip)
			continue
		}

		_ = conn.Close()

		log.Printf("OK %s latency=%dms\n", ip, time.Since(start).Milliseconds())
	}

	log.Println("")
	log.Println("-------------------------------------------")
	log.Println("")
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

	return out
}

func lookupGoogleUDP(domain string) []string {
	c := dns.Client{}
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

	return ips
}

func lookupGoogleDoH(domain string) []string {
	u, err := upstream.AddressToUpstream("https://dns.google/dns-query", &upstream.Options{})
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

	return ips
}

func compareIPSets(a, b []string) bool {
	set := map[string]bool{}

	for _, ip := range a {
		set[ip] = true
	}

	for _, ip := range b {
		if set[ip] {
			return true
		}
	}

	return false
}

func printIPs(ips []string) {
	if len(ips) == 0 {
		log.Println("none")
		return
	}

	for _, ip := range ips {
		log.Println(ip)
	}
}
