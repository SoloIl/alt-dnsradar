package main

import (
	"io"
	"log"
	"os"
)

func createLog(logfile string, silent bool) error {
	if _, err := os.Stat(logfile); err == nil {
		_ = os.Remove(logfile)
	}

	f, err := os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	if silent {
		log.SetOutput(f)
	} else {
		mw := io.MultiWriter(os.Stdout, f)
		log.SetOutput(mw)
	}

	log.SetFlags(0)
	log.SetPrefix("")

	return nil
}
