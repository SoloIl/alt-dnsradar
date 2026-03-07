package main

import "strings"

func clusterIPs(ips []string) map[string][]string {
	clusters := make(map[string][]string)

	for _, ip := range ips {
		parts := strings.Split(ip, ".")

		if len(parts) != 4 {
			continue
		}

		pop := parts[0] + "." + parts[1] + "." + parts[2] + ".0/24"
		clusters[pop] = append(clusters[pop], ip)
	}

	return clusters
}

func extractPOPIPs(clusters map[string][]string) []string {
	var ips []string

	for _, v := range clusters {
		ips = append(ips, v[0])
	}

	return ips
}
