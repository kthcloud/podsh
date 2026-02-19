package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"path"

	"github.com/Phillezi/common/utils/or"
	"github.com/Phillezi/interrupt/pkg/interrupt"
	"github.com/Phillezi/interrupt/pkg/manager"
	"github.com/kthcloud/podsh/internal/server"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const namespace = "deploy"

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	dat, err := os.ReadFile(or.Or(os.Getenv("PODSH_DEV_PUBLIC_KEY_FILE"), path.Join(home, ".ssh/id_ed25519.pub")))
	if err != nil {
		log.Fatal(err)
	}

	// FIXME: ensure RBAC is used when actually deployed so we use a restricted client config that only has access to:
	// - List pods in the deploy namespace
	// - Exec pods in the deploy namespace
	cfg, err := clientcmd.BuildConfigFromFlags("", or.Or(os.Getenv("KUBECONFIG"), path.Join(home, ".kube/config")))
	if err != nil {
		log.Fatal(err)
	}

	kc, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	s := server.New(string(dat), kc, cfg, namespace)

	slog.SetLogLoggerLevel(slog.LevelDebug)

	interrupt.Main(func(m manager.ManagedManager, cancel context.CancelFunc) {
		if err := s.Start(m.Context()); err != nil {
			cancel()
			log.Fatal(err)
		}
	}, interrupt.WithManagerOpts(manager.WithPrompt(true)))
}
