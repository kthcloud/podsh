package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {
	ch := make(chan os.Signal, 1)
	defer close(ch)
	signal.Notify(ch, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
	dur := parseArgs()
	if dur != nil {
		select {
		case <-time.After(*dur):
		case <-ch:
		}
	} else {
		<-ch
	}
}

func parseArgs() *time.Duration {
	for _, arg := range os.Args[1:] {
		sanitized := strings.TrimSpace(arg)
		if sanitized == "" {
			continue
		}

		if strings.EqualFold(sanitized, "infinity") {
			return nil
		}

		// Try Go duration parsing first
		if t, err := time.ParseDuration(sanitized); err == nil {
			return &t
		}

		// If not a duration, see if its a plain number => default to seconds
		if sec, err := strconv.Atoi(sanitized); err == nil {
			d := time.Duration(sec) * time.Second
			return &d
		}

		fmt.Fprintf(os.Stderr, "Failed to parse duration '%s'\n", sanitized)
		os.Exit(1)
	}

	return nil
}
