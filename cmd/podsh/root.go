package main

import (
	"log"
	"log/slog"
	"os"
	"path"

	viperconf "github.com/Phillezi/common/config/viper"
	"github.com/gliderlabs/ssh"
	"github.com/kthcloud/podsh/internal/defaults"
	"github.com/kthcloud/podsh/internal/server"
	"github.com/kthcloud/podsh/internal/sshd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var rootCmd = cobra.Command{
	Use:     podsh,
	Short:   short,
	Long:    long,
	Version: version,
	Run: func(cmd *cobra.Command, _ []string) {
		banner()

		dat, err := os.ReadFile(viper.GetString("dev-public-key-file"))
		if err != nil {
			log.Fatal(err)
		}

		// FIXME: ensure RBAC is used when actually deployed so we use a restricted client config that only has access to:
		// - List pods in the deploy namespace
		// - Exec pods in the deploy namespace
		cfg, err := clientcmd.BuildConfigFromFlags("", viper.GetString("kubeconfig"))
		if err != nil {
			log.Fatal(err)
		}

		kc, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			log.Fatal(err)
		}

		// TODO: set cfg
		s := server.New()

		slog.SetLogLoggerLevel(slog.LevelDebug)

		if err := s.Start(cmd.Context()); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	cobra.OnInitialize(func() { viperconf.InitConfig(podsh) })

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	rootCmd.Flags().String("address", defaults.DefaultBindAddress, "The server address")
	viper.BindPFlag("address", rootCmd.Flags().Lookup("address"))

	rootCmd.Flags().StringP("namespace", "n", defaults.DefaultNamespace, "The namespace that should be accessible")
	viper.BindPFlag("namespace", rootCmd.Flags().Lookup("namespace"))

	rootCmd.Flags().String("kubeconfig", path.Join(home, ".kube", "config"), "The kubeconfig that should be used")
	viper.BindPFlag("kubeconfig", rootCmd.Flags().Lookup("kubeconfig"))

	rootCmd.Flags().String("dev-public-key-file", path.Join(home, ".ssh", "id_ed25519.pub"), "The ")
	viper.BindPFlag("dev-public-key-file", rootCmd.Flags().Lookup("dev-public-key-file"))

	rootCmd.Flags().Float32("limit-rate", defaults.DefaultLimitRate, "The ratelimit rate to use")
	viper.BindPFlag("limit-rate", rootCmd.Flags().Lookup("limit-rate"))

	rootCmd.Flags().Int("limit-burst", defaults.DefaultLimitBurst, "The ratelimit burst to use")
	viper.BindPFlag("limit-burst", rootCmd.Flags().Lookup("limit-burst"))

	rootCmd.Flags().Duration("limit-ttl", defaults.DefaultLimitTTL, "The ratelimit ttl to use")
	viper.BindPFlag("limit-ttl", rootCmd.Flags().Lookup("limit-ttl"))
}

func mkDevAutAuthh(devPublicKey string) sshd.PublicKeyAuthenticator {
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(devPublicKey))
	if err != nil {
		log.Fatal(err)
	}

	pubKeyBytes := pubKey.Marshal()

	// TODO: connect to go-deploy
	return sshd.NewMapAuthenticator(map[string]*sshd.Identity{
		string(pubKeyBytes): {
			User:      "user@kth.se",
			UserID:    "4efea96b-2d6b-41f6-96a2-656f18d6f8d1",
			PublicKey: pubKeyBytes,
		},
	})
}
