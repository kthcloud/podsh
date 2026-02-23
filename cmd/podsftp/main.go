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
			// This doesnt seem to work, TODO: we should be able to make it work somehow
			sftp.WithServerWorkingDirectory("/proc/1/root/"),
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
