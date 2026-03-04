package profiles

import (
	"context"
	"log/slog"
	"os"

	"github.com/go-redis/redis"
	"github.com/kthcloud/podsh/internal/auth"
	"github.com/kthcloud/podsh/internal/k8s"
	"github.com/kthcloud/podsh/internal/k8s/validate"
	ratelimiter "github.com/kthcloud/podsh/internal/ratelimit"
	"github.com/kthcloud/podsh/internal/server"
	"github.com/kthcloud/podsh/internal/sshd"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type ProdProfileImpl struct{}

func (ProdProfileImpl) Mode() Mode {
	return ModeProd
}

func (ProdProfileImpl) Config(ctx context.Context, v *viper.Viper) (*server.Config, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	var (
		cfg *rest.Config
		err error
	)

	kubeconfig := v.GetString("kubeconfig")

	if kubeconfig != "" {
		slog.Info("using kubeconfig", "path", kubeconfig)
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			slog.Error("could not build k8s config from kubeconfig", "error", err)
			return nil, err
		}
	} else {
		slog.Info("no kubeconfig provided, trying in-cluster config")
		cfg, err = rest.InClusterConfig()
		if err != nil {
			slog.Error("could not load in-cluster config", "error", err)
			return nil, err
		}
	}

	kc, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		slog.Default().Error("could not get k8s clientset", "error", err)
		return nil, err
	}

	hostSigner, err := sshd.LoadHostSigner(v.GetString("ssh-host-signer-path"))
	if err != nil {
		slog.Default().Error("could not load host signer key", "error", err)
		return nil, err
	}

	client := redis.NewClient(&redis.Options{
		Addr:     v.GetString("redis-address"),
		Password: v.GetString("redis-password"),
		DB:       v.GetInt("redis-db"),
	}).WithContext(ctx)

	if err := client.Ping().Err(); err != nil {
		slog.Default().Error("could not ping redis", "error", err)
		return nil, err
	}

	auth := auth.NewRedisPublicKeyAuthenticator(client)

	if err := validate.ValidatePermissions(ctx, v.GetString("namespace"), kc, cfg); err != nil {
		return nil, err
	}

	return &server.Config{
		Ctx: ctx,

		Address: v.GetString("address"),

		SSHDConfig: sshd.Config{
			Ctx:                    ctx,
			Signer:                 hostSigner,
			PublicKeyAuthenticator: auth,
			Limiter:                ratelimiter.New(v.GetFloat64("limit-rate"), v.GetInt("limit-burst"), v.GetDuration("limit-ttl")),
			Hasher:                 ratelimiter.NewHasher([]byte("supersecret")),

			Logger: slog.Default(),

			Handler2: k8s.New(kc, cfg, k8s.NewLabelResolver(kc, v.GetString("namespace"))),
		},

		Logger: slog.Default(),
	}, nil
}
