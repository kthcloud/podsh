package main

import (
	"context"
	"errors"
	"log"
	"os"
	"path"

	"github.com/kthcloud/podsh/internal/config"
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

		if err := s.Start(cmd.Context()); err != nil && !errors.Is(err, context.Canceled) {
			return err
		}

		return nil
	},
}

func init() {
	cobra.OnInitialize(func() { config.InitConfig(podsh) })

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	profileFlag, _ := profiles.NewProfileFlag(profiles.ProfileKeyDev)

	rootCmd.Flags().Var(profileFlag, "profile", "The profile")
	viper.BindPFlag("profile", rootCmd.Flags().Lookup("profile"))

	rootCmd.Flags().String("config", "", "Where to look for the config.yaml")
	viper.BindPFlag("config", rootCmd.Flags().Lookup("config"))

	rootCmd.Flags().String("log-level", "info", "The log level to use")
	viper.BindPFlag("log.level", rootCmd.Flags().Lookup("log-level"))

	rootCmd.Flags().String("address", defaults.DefaultBindAddress, "The server address")
	viper.BindPFlag("address", rootCmd.Flags().Lookup("address"))

	rootCmd.Flags().StringP("namespace", "n", defaults.DefaultNamespace, "The namespace that should be accessible")
	viper.BindPFlag("namespace", rootCmd.Flags().Lookup("namespace"))

	rootCmd.Flags().String("kubeconfig", path.Join(home, ".kube", "config"), "The kubeconfig that should be used")
	viper.BindPFlag("kubeconfig", rootCmd.Flags().Lookup("kubeconfig"))

	rootCmd.Flags().String("dev-public-key-file", path.Join(home, ".ssh", "id_ed25519.pub"), "The ")
	viper.BindPFlag("dev.publickeyfile", rootCmd.Flags().Lookup("dev-public-key-file"))

	rootCmd.Flags().Float64("limit-rate", defaults.DefaultLimitRate, "The ratelimit rate to use")
	viper.BindPFlag("limit.rate", rootCmd.Flags().Lookup("limit-rate"))

	rootCmd.Flags().Int("limit-burst", defaults.DefaultLimitBurst, "The ratelimit burst to use")
	viper.BindPFlag("limit.burst", rootCmd.Flags().Lookup("limit-burst"))

	rootCmd.Flags().Duration("limit-ttl", defaults.DefaultLimitTTL, "The ratelimit ttl to use")
	viper.BindPFlag("limit.ttl", rootCmd.Flags().Lookup("limit-ttl"))

	rootCmd.Flags().String("metrics-address", defaults.DefaultMetricsAddr, "The address the metrics server should use")
	viper.BindPFlag("metrics.address", rootCmd.Flags().Lookup("metrics-address"))

	rootCmd.Flags().String("ssh-host-signer-path", defaults.DefaultHostSignerPath, "The path where the hosts signing key is")
	viper.BindPFlag("ssh.hostsignerpath", rootCmd.Flags().Lookup("ssh-host-signer-path"))

	rootCmd.Flags().String("redis-address", defaults.DefaultRedisAddress, "The address to redis")
	viper.BindPFlag("redis.address", rootCmd.Flags().Lookup("redis-address"))

	rootCmd.Flags().Int("redis-db", defaults.DefaultRedisDB, "The db to use in redis")
	viper.BindPFlag("redis.db", rootCmd.Flags().Lookup("redis-db"))

	rootCmd.Flags().String("redis-password", defaults.DefaultRedisPassword, "The password to use for redis")
	viper.BindPFlag("redis.password", rootCmd.Flags().Lookup("redis-password"))
}
