package main

import (
	"context"
	"log"

	"github.com/Phillezi/interrupt/pkg/interrupt"
	"github.com/Phillezi/interrupt/pkg/manager"
)

func main() {
	interrupt.Main(func(m manager.ManagedManager, cancel context.CancelFunc) {
		if err := rootCmd.ExecuteContext(m.Context()); err != nil {
			cancel()
			log.Fatal(err)
		}
	}, interrupt.WithManagerOpts(manager.WithPrompt(true)))
}
