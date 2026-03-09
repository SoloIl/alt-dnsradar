package main

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/AdguardTeam/dnsproxy/upstream"
	"github.com/miekg/dns"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

func generateSubnets() []string {
	var subnets []string

	for i := 0; i <= 255; i += 10 {
		for j := 0; j <= 255; j += 10 {
			ip := net.ParseIP(fmt.Sprintf("%d.%d.0.0", i, j))

			if isPublicIPv4(ip) {
				subnets = append(subnets, fmt.Sprintf("%d.%d.0.0/16", i, j))
			}
		}
	}

	return subnets
}

func createResolver() (upstream.Upstream, error) {
	o := &upstream.Options{
		Timeout: requestTimeout(),
	}

	u, err := upstream.AddressToUpstream(*flagSettings.Resolver, o)
	if err != nil {
		return nil, fmt.Errorf("create resolver %q: %w", *flagSettings.Resolver, err)
	}

	return u, nil
}

func sendRequests(ctx context.Context, subnets []string, u upstream.Upstream, total int) []string {
	var wg sync.WaitGroup
	threads := 20
	semaphore := make(chan struct{}, threads)

	var mu sync.Mutex
	var rawReplies []string

	p := mpb.New(mpb.WithWidth(50))
	bar := p.New(int64(total),
		mpb.BarStyle().Rbound("|"),
		mpb.PrependDecorators(
			decor.Name("ECS scan "),
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

	for _, subnet := range subnets {
		select {
		case <-ctx.Done():
			wg.Wait()
			bar.Abort(true)
			p.Wait()
			return rawReplies
		default:
		}

		wg.Add(1)
		semaphore <- struct{}{}

		go func(subnet string) {
			defer func() {
				<-semaphore
				wg.Done()
			}()

			reply := dnsLookup(subnet, u)

			mu.Lock()
			rawReplies = append(rawReplies, reply...)
			mu.Unlock()

			bar.Increment()
		}(subnet)
	}

	wg.Wait()
	if ctx.Err() != nil {
		bar.Abort(true)
	}
	p.Wait()

	return rawReplies
}

func dnsLookup(subnet string, u upstream.Upstream) []string {
	var rawReply []string

	q := dns.Question{
		Name:   dns.Fqdn(*flagSettings.URL),
		Qclass: dns.ClassINET,
		Qtype:  dns.TypeA,
	}

	req := &dns.Msg{}
	req.Id = dns.Id()
	req.RecursionDesired = true
	req.Question = []dns.Question{q}

	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil
	}

	ones, _ := ipNet.Mask.Size()

	ecs := &dns.EDNS0_SUBNET{
		Code:          dns.EDNS0SUBNET,
		Family:        1,
		SourceNetmask: uint8(ones),
		Address:       ipNet.IP,
	}

	req.SetEdns0(dns.DefaultMsgSize, false)

	opt := req.IsEdns0()
	opt.Option = append(opt.Option, ecs)

	reply, err := u.Exchange(req)
	if err != nil {
		return nil
	}

	dnsMutex.Lock()
	successfulDNS++
	dnsMutex.Unlock()

	for _, ans := range reply.Answer {
		if a, ok := ans.(*dns.A); ok {
			rawReply = append(rawReply, a.A.String())
		}
	}

	return rawReply
}
