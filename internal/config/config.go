package config

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/Phillezi/common/utils/or"
	"github.com/spf13/viper"
)

func InitConfig(projectName string, filenames ...string) {
	if len(filenames) > 0 {
		filenames = append(filenames, "config")
	} else {
		filenames = []string{"config"}
	}
	viper.SetConfigName(or.Or(filenames...)) // Name of the config file (without extension)
	viper.SetConfigType("yaml")              // File format (yaml)
	viper.AddConfigPath(GetConfigPath(projectName))
	viper.AddConfigPath(or.Or(func() string {
		wd, _ := os.Getwd()
		return wd
	}(), "."))

	viper.SetEnvPrefix(projectName)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.AutomaticEnv() // Read environment variables

	// Load config file
	if err := viper.ReadInConfig(); err != nil {
		slog.Default().Debug("Config file not found, using defaults or environment variables.")
	} else {
		slog.Default().Debug("Using config file: ", viper.ConfigFileUsed())
	}
}

func getConfigPath(projectName string) (string, error) {
	basePath, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	configPath := path.Join(basePath, projectName)
	fileDescr, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		err := os.MkdirAll(configPath, os.ModePerm)
		if err != nil {
			return "", err
		}
		fileDescr, err = os.Stat(configPath)
		if err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	}
	if !fileDescr.IsDir() {
		return "", fmt.Errorf("default config dir is file")
	}
	return configPath, nil
}

func GetConfigPath(projectName string) string {
	if viper.IsSet("config") {
		cfg := strings.TrimSpace(viper.GetString("config"))
		if cfg != "" {
			if _, err := os.Stat(cfg); err == nil {
				return cfg
			}
		}
	}
	configPath, err := getConfigPath(projectName)
	if err != nil {
		slog.Default().Warn("error getting config path", err)
		slog.Default().Info("defaulting to:", func() string {
			wd, err := os.Getwd()
			if err != nil {
				return "."
			}
			return wd
		}())
		configPath = "."
	}
	return configPath
}
