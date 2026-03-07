package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	initFlags()

	err := createLog(*flagSettings.LogName, *flagSettings.QuietMode)
	check(err)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	var cleanErr error
	*flagSettings.URL, cleanErr = cleanURL(*flagSettings.URL)
	check(cleanErr)

	log.Printf("%s v%s\n\n", PROGRAMNAME, VERSION)
	log.Printf("Processing URL '%s'\n\n", *flagSettings.URL)

	diagnoseDNS(*flagSettings.URL)

	resolver := createResolver()
	defer resolver.Close()

	subnets := generateSubnets()

	log.Printf("Starting ECS scan\n")
	log.Printf("Total ECS subnets: %d\n\n", len(subnets))

	rawReplies := sendRequests(ctx, subnets, resolver, len(subnets))
	unique := removeDuplicates(rawReplies)

	log.Printf("\nDNS successful replies: %d\n", successfulDNS)
	log.Printf("Unique IP discovered: %d\n\n", len(unique))

	if *flagSettings.Verbose {
		clusters := clusterIPs(unique)
		log.Printf("POP clusters discovered: %d\n\n", len(clusters))
	}

	results := validateIPs(unique)

	var alive []IPResult
	for _, r := range results {
		if r.Alive {
			alive = append(alive, r)
		}
	}

	sortResults(alive)
	printGeoTable(alive, 5)

	if *flagSettings.ShowAll {
		printAll(unique)
	}
}
