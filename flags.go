package main

import (
	"flag"
	"fmt"
	"os"
)

func initFlags() {
	flagSettings.LogName = flag.String(
		"l",
		thisFilename()+".log",
		"log file name",
	)

	flagSettings.Resolver = flag.String(
		"resolver",
		"https://dns.google/dns-query",
		"DoH resolver with ECS support",
	)

	flagSettings.RequestsTimeout = flag.Int(
		"timeout",
		3,
		"network timeout in seconds",
	)

	flagSettings.ShowAll = flag.Bool(
		"all",
		false,
		"show all discovered IPs",
	)

	flagSettings.Verbose = flag.Bool(
		"v",
		false,
		"verbose output",
	)

	flagSettings.QuietMode = flag.Bool(
		"q",
		false,
		"quiet console",
	)

	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage:")
		fmt.Println("  dnsradar <domain> [flags]")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  dnsradar instagram.com")
		fmt.Println("  dnsradar youtube.com --all")
		os.Exit(1)
	}

	domain := flag.Arg(0)
	flagSettings.URL = &domain
}
