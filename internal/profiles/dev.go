package profiles

import (
	"context"
	"log/slog"
	"os"

	"github.com/kthcloud/podsh/internal/k8s"
	"github.com/kthcloud/podsh/internal/k8s/validate"
	register "github.com/kthcloud/podsh/internal/metrics"
	ratelimiter "github.com/kthcloud/podsh/internal/ratelimit"
	"github.com/kthcloud/podsh/internal/server"
	"github.com/kthcloud/podsh/internal/sshd"
	"github.com/kthcloud/podsh/pkg/metrics"
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
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: MustParseLevel(v.GetString("log.level")),
	}))
	slog.SetDefault(logger)

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

	devPublicKey, err := os.ReadFile(v.GetString("dev.publickeyfile"))
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

	metrics := metrics.NewPrometheus()
	register.RegisterSSHdMetrics(metrics)
	register.RegisterK8sMetrics(metrics)

	return &server.Config{
		Ctx: ctx,

		Address:        v.GetString("address"),
		MetricsAddress: v.GetString("metrics.address"),
		Metrics:        metrics,

		SSHDConfig: sshd.Config{
			Ctx:                    ctx,
			Signer:                 mockSigner,
			PublicKeyAuthenticator: auth,
			Limiter:                ratelimiter.New(v.GetFloat64("limit.rate"), v.GetInt("limit.burst"), v.GetDuration("limit.ttl")),
			Hasher:                 ratelimiter.NewHasher([]byte("supersecret")),
			Metrics:                metrics,

			Logger: slog.Default(),

			Handler2: k8s.New(kc, cfg, k8s.NewLabelResolver(kc, v.GetString("namespace")), metrics),
		},

		Logger: slog.Default(),
	}, nil
}
