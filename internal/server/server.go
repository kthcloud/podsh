package server

import (
	"context"
	"log"
	"log/slog"

	"github.com/kthcloud/podsh/internal/gateway"
	"github.com/kthcloud/podsh/internal/sshd"
	"golang.org/x/crypto/ssh"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Server struct {
	sshServer *sshd.Server
}

func New(devPublicKey string, kc *kubernetes.Clientset, rest *rest.Config, namespace string) *Server {
	// FIXME: dont use this in prod!
	signer, err := sshd.NewMockHostSigner()
	if err != nil {
		panic(err)
	}

	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(devPublicKey))
	if err != nil {
		log.Fatal(err)
	}

	pubKeyBytes := pubKey.Marshal()

	// TODO: connect to go-deploy
	auth := sshd.NewMapAuthenticator(map[string]*sshd.Identity{
		string(pubKeyBytes): {
			User:      "user@kth.se",
			UserID:    "4efea96b-2d6b-41f6-96a2-656f18d6f8d1",
			PublicKey: pubKeyBytes,
		},
	})

	s := &Server{
		sshServer: sshd.New(sshd.WithHostSigner(signer), sshd.WithPublicKeyAuth(auth)),
	}

	s.sshServer.HandleSession(gateway.NewHandler(slog.Default(), gateway.NewLabelResolver(kc, namespace), gateway.NewK8sExecutor(kc, rest)))

	return s
}

func (s *Server) Start(ctx context.Context) error {
	if err := s.sshServer.ListenAndServe(ctx, "localhost:2222"); err != nil {
		return err
	}
	return nil
}
