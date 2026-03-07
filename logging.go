package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

func createLog(logfile string, silent bool) error {
	errWriter := io.Writer(os.Stderr)
	runLog = io.Discard

	if logfile == "" {
		log.SetOutput(errWriter)
		log.SetFlags(0)
		log.SetPrefix("")
		return nil
	}

	f, err := os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	logFile = f
	runLog = f

	if silent {
		errWriter = f
	} else {
		errWriter = io.MultiWriter(os.Stderr, f)
	}

	log.SetOutput(errWriter)
	log.SetFlags(0)
	log.SetPrefix("")

	return nil
}

func closeLog() error {
	if logFile == nil {
		return nil
	}

	err := logFile.Close()
	logFile = nil
	runLog = io.Discard
	return err
}

func tracef(format string, args ...any) {
	if runLog == io.Discard {
		return
	}

	_, _ = fmt.Fprintf(runLog, format, args...)
}
