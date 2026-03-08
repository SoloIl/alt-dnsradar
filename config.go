package main

import (
	"io"
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

type SetRelation int

const (
	RelationUnavailable SetRelation = iota
	RelationExactMatch
	RelationPartialOverlap
	RelationNoOverlap
)

type TLSStatus string

const (
	TLSStatusOK      TLSStatus = "OK"
	TLSStatusFail    TLSStatus = "FAIL"
	TLSStatusTimeout TLSStatus = "TIMEOUT"
	TLSStatusSkip    TLSStatus = "-"
)

type DNSComparison struct {
	LeftName   string
	RightName  string
	LeftCount  int
	RightCount int
	Relation   SetRelation
}

type InitialDiagRow struct {
	Source     string
	IP         string
	TCPAlive   bool
	TCPLatency time.Duration
	TLSStatus  TLSStatus
	Note       string
}

type ProbeResult struct {
	TCPAlive   bool
	TCPLatency time.Duration
	TLSStatus  TLSStatus
}

var (
	flagSettings Settings

	successfulDNS int
	dnsMutex      sync.Mutex
	logFile       io.Closer
	runLog        io.Writer = io.Discard
)
