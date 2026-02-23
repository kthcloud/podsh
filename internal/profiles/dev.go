package profiles

import (
	"context"
	"log/slog"
	"os"

	"github.com/kthcloud/podsh/internal/gateway"
	ratelimiter "github.com/kthcloud/podsh/internal/ratelimit"
	"github.com/kthcloud/podsh/internal/server"
	"github.com/kthcloud/podsh/internal/sshd"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type DevProfileImpl struct{}

func (DevProfileImpl) Mode() Mode {
	return ModeDev
}

func (DevProfileImpl) Config(ctx context.Context, v *viper.Viper) (*server.Config, error) {
	// FIXME: ensure RBAC is used when actually deployed so we use a restricted client config that only has access to:
	// - List pods in the deploy namespace
	// - Exec pods in the deploy namespace
	cfg, err := clientcmd.BuildConfigFromFlags("", v.GetString("kubeconfig"))
	if err != nil {
		return nil, err
	}

	kc, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	mockSigner, err := sshd.NewMockHostSigner()
	if err != nil {
		return nil, err
	}

	devPublicKey, err := os.ReadFile(v.GetString("dev-public-key-file"))
	if err != nil {
		return nil, err
	}

	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(devPublicKey)
	if err != nil {
		return nil, err
	}

	pubKeyBytes := pubKey.Marshal()

	auth := sshd.NewMapAuthenticator(map[string]*sshd.Identity{
		string(pubKeyBytes): {
			User:      "user@kth.se",
			UserID:    "4efea96b-2d6b-41f6-96a2-656f18d6f8d1",
			PublicKey: pubKeyBytes,
		},
	})

	return &server.Config{
		Ctx: ctx,

		Address: v.GetString("address"),

		SSHDConfig: sshd.Config{
			Ctx:                    ctx,
			Signer:                 mockSigner,
			PublicKeyAuthenticator: auth,
			Limiter:                ratelimiter.New(v.GetFloat64("limit-rate"), v.GetInt("limit-burst"), v.GetDuration("limit-ttl")),
			Hasher:                 ratelimiter.NewHasher([]byte("supersecret")),

			Logger: slog.Default(),
		},
		Handler: gateway.NewHandler(slog.Default(),
			gateway.NewLabelResolver(kc, v.GetString("namespace")),
			gateway.NewK8sExecutor(kc, cfg),
			gateway.NewK8sSFTP(kc, cfg),
		),

		Logger: slog.Default(),
	}, nil
}
