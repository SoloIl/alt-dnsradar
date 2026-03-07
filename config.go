package main

import (
	"sync"
	"time"
)

const (
	PROGRAMNAME = "Alt DNSRadar"
	VERSION     = "0.12"
)

type Settings struct {
	LogName         *string
	URL             *string
	Resolver        *string
	RequestsTimeout *int
	ShowAll         *bool
	Verbose         *bool
	QuietMode       *bool
}

type IPResult struct {
	IP         string
	TCPLatency time.Duration
	Alive      bool
}

var (
	flagSettings Settings

	successfulDNS int
	dnsMutex      sync.Mutex
)
