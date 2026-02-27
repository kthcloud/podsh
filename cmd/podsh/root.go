package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"os"
	"path"

	viperconf "github.com/Phillezi/common/config/viper"
	"github.com/kthcloud/podsh/internal/defaults"
	"github.com/kthcloud/podsh/internal/profiles"
	"github.com/kthcloud/podsh/internal/server"
	"github.com/kthcloud/podsh/pkg/notice"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = cobra.Command{
	Use:     podsh,
	Short:   short,
	Long:    long,
	Version: version,
	RunE: func(cmd *cobra.Command, _ []string) error {
		banner()

		prof, err := profiles.Get(profiles.ProfileKey(viper.GetString("profile")))
		if err != nil {
			return err
		}

		if prof.Mode() == profiles.ModeDev {
			notice.Warn("Unsafe config", `Using the development profile, DONT USE THIS IN PRODUCTION!`)
		}

		cfg, err := prof.Config(cmd.Context(), viper.GetViper())
		if err != nil {
			return err
		}

		s := server.New(server.WithConfig(*cfg))
		if err := s.Validate(); err != nil {
			return err
		}

		slog.SetLogLoggerLevel(slog.LevelDebug)

		if err := s.Start(cmd.Context()); err != nil && !errors.Is(err, context.Canceled) {
			return err
		}

		return nil
	},
}

func init() {
	cobra.OnInitialize(func() { viperconf.InitConfig(podsh) })

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	profileFlag, _ := profiles.NewProfileFlag(profiles.ProfileKeyDev)

	rootCmd.Flags().Var(profileFlag, "profile", "The profile")
	viper.BindPFlag("profile", rootCmd.Flags().Lookup("profile"))

	rootCmd.Flags().String("address", defaults.DefaultBindAddress, "The server address")
	viper.BindPFlag("address", rootCmd.Flags().Lookup("address"))

	rootCmd.Flags().StringP("namespace", "n", defaults.DefaultNamespace, "The namespace that should be accessible")
	viper.BindPFlag("namespace", rootCmd.Flags().Lookup("namespace"))

	rootCmd.Flags().String("kubeconfig", path.Join(home, ".kube", "config"), "The kubeconfig that should be used")
	viper.BindPFlag("kubeconfig", rootCmd.Flags().Lookup("kubeconfig"))

	rootCmd.Flags().String("dev-public-key-file", path.Join(home, ".ssh", "id_ed25519.pub"), "The ")
	viper.BindPFlag("dev-public-key-file", rootCmd.Flags().Lookup("dev-public-key-file"))

	rootCmd.Flags().Float64("limit-rate", defaults.DefaultLimitRate, "The ratelimit rate to use")
	viper.BindPFlag("limit-rate", rootCmd.Flags().Lookup("limit-rate"))

	rootCmd.Flags().Int("limit-burst", defaults.DefaultLimitBurst, "The ratelimit burst to use")
	viper.BindPFlag("limit-burst", rootCmd.Flags().Lookup("limit-burst"))

	rootCmd.Flags().Duration("limit-ttl", defaults.DefaultLimitTTL, "The ratelimit ttl to use")
	viper.BindPFlag("limit-ttl", rootCmd.Flags().Lookup("limit-ttl"))

	rootCmd.Flags().String("metrics-address", defaults.DefaultMetricsAddr, "The address the metrics server should use")
	viper.BindPFlag("metrics-address", rootCmd.Flags().Lookup("metrics-address"))
}
