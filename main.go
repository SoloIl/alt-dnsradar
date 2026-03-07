package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	initFlags()

	err := createLog(*flagSettings.LogName, *flagSettings.QuietMode)
	check(err)
	defer func() {
		closeErr := closeLog()
		if closeErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error closing log: %v\n", closeErr)
		}
	}()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	var cleanErr error
	*flagSettings.URL, cleanErr = cleanURL(*flagSettings.URL)
	check(cleanErr)

	fmt.Printf("%s v%s\n\n", PROGRAMNAME, VERSION)
	fmt.Printf("Processing URL '%s'\n\n", *flagSettings.URL)
	tracef("%s v%s\n", PROGRAMNAME, VERSION)
	tracef("Processing URL '%s'\n", *flagSettings.URL)

	diagnoseDNS(*flagSettings.URL)

	resolver, err := createResolver()
	check(err)
	defer resolver.Close()

	subnets := generateSubnets()

	fmt.Printf("Starting ECS scan\n")
	fmt.Printf("Total ECS subnets: %d\n\n", len(subnets))
	tracef("Starting ECS scan\n")
	tracef("Total ECS subnets: %d\n", len(subnets))

	rawReplies := sendRequests(ctx, subnets, resolver, len(subnets))
	unique := removeDuplicates(rawReplies)

	fmt.Printf("\nDNS successful replies: %d\n", successfulDNS)
	fmt.Printf("Unique IP discovered: %d\n\n", len(unique))
	tracef("DNS successful replies: %d\n", successfulDNS)
	tracef("Unique IP discovered: %d\n", len(unique))

	if *flagSettings.Verbose {
		clusters := clusterIPs(unique)
		fmt.Printf("POP clusters discovered: %d\n\n", len(clusters))
		tracef("POP clusters discovered: %d\n", len(clusters))
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
