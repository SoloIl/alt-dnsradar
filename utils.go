package main

import (
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func requestTimeout() time.Duration {
	return time.Duration(*flagSettings.RequestsTimeout) * time.Second
}

func removeDuplicates(slice []string) []string {
	seen := make(map[string]struct{})
	var result []string

	for _, v := range slice {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}

	return result
}

func cleanURL(rawURL string) (string, error) {
	if !strings.Contains(rawURL, "://") {
		rawURL = "http://" + rawURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	return parsed.Hostname(), nil
}

func thisFilename() string {
	f := os.Args[0]
	return strings.TrimSuffix(filepath.Base(f), filepath.Ext(f))
}

func check(err error) {
	if err != nil {
		log.Print(msgErrorPrefix(err))
		os.Exit(1)
	}
}
