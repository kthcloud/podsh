package profiles

import (
	"context"
	"log/slog"
	"os"

	"github.com/kthcloud/podsh/internal/k8s"
	"github.com/kthcloud/podsh/internal/k8s/validate"
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

	if err := validate.ValidatePermissions(ctx, v.GetString("namespace"), kc, cfg); err != nil {
		return nil, err
	}

	return &server.Config{
		Ctx: ctx,

		Address:        v.GetString("address"),
		MetricsAddress: v.GetString("metrics-address"),

		SSHDConfig: sshd.Config{
			Ctx:                    ctx,
			Signer:                 mockSigner,
			PublicKeyAuthenticator: auth,
			Limiter:                ratelimiter.New(v.GetFloat64("limit-rate"), v.GetInt("limit-burst"), v.GetDuration("limit-ttl")),
			Hasher:                 ratelimiter.NewHasher([]byte("supersecret")),

			Logger: slog.Default(),

			Handler2: k8s.New(kc, cfg, k8s.NewLabelResolver(kc, v.GetString("namespace"))),
		},

		Logger: slog.Default(),
	}, nil
}
