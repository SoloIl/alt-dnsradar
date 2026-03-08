package main

import (
	"fmt"
	"sort"
)

func sortResults(results []IPResult) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].TCPLatency < results[j].TCPLatency
	})
}

func printAll(ips []string) {
	fmt.Println("")
	fmt.Println(msgAllDiscoveredIPs())
	fmt.Println("")

	for _, ip := range ips {
		fmt.Println(ip)
	}
}
