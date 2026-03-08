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
			_, _ = fmt.Fprint(os.Stderr, msgErrorClosingLog(closeErr))
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

	fmt.Printf("%s\n\n", msgProgramBanner(PROGRAMNAME, VERSION))
	fmt.Printf("%s\n\n", msgProcessingURL(*flagSettings.URL))
	tracef("%s v%s\n", PROGRAMNAME, VERSION)
	tracef("Processing URL '%s'\n", *flagSettings.URL)

	diagnoseDNS(*flagSettings.URL)

	resolver, err := createResolver()
	check(err)
	defer resolver.Close()

	subnets := generateSubnets()

	fmt.Printf("%s\n", msgStartingECSScan())
	fmt.Printf("%s\n\n", msgTotalECSSubnets(len(subnets)))
	tracef("Starting ECS scan\n")
	tracef("Total ECS subnets: %d\n", len(subnets))

	rawReplies := sendRequests(ctx, subnets, resolver, len(subnets))
	unique := removeDuplicates(rawReplies)

	fmt.Printf("\n%s\n", msgSuccessfulDNSReplies(successfulDNS))
	fmt.Printf("%s\n\n", msgUniqueIPDiscovered(len(unique)))
	tracef("DNS successful replies: %d\n", successfulDNS)
	tracef("Unique IP discovered: %d\n", len(unique))

	if *flagSettings.Verbose {
		clusters := clusterIPs(unique)
		fmt.Printf("%s\n\n", msgPOPClustersDiscovered(len(clusters)))
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
