package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/Phillezi/interrupt/pkg/interrupt"
	"github.com/Phillezi/interrupt/pkg/manager"
	"github.com/pkg/sftp"
)

// Helper binary for sftp support, gets spawned as a ephemeral container
func main() {
	interrupt.Main(func(m manager.ManagedManager, cancel context.CancelFunc) {
		// TODO: respect context
		srv, err := sftp.NewServer(
			struct {
				io.Reader
				io.WriteCloser
			}{
				os.Stdin,
				os.Stdout,
			},
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating sftp server: %s\n", err.Error())
			cancel()
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Starting SFTP server [IN=STDIN,OUT=STDOUT]\n")
		srv.Serve()
	})
}
