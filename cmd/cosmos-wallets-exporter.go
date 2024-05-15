package main

import (
	"main/pkg"
	configPkg "main/pkg/config"
	"main/pkg/logger"

	"github.com/spf13/cobra"
)

var (
	version = "unknown"
)

func ExecuteMain(configPath string) {
	app := pkg.NewApp(configPath)
	app.Start()
}

func ExecuteValidateConfig(configPath string) {
	config, err := configPkg.GetConfig(configPath)
	if err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Could not load config!")
	}

	if err := config.Validate(); err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Provided config is invalid!")
	}

	logger.GetDefaultLogger().Info().Msg("Provided config is valid.")
}

func main() {
	var ConfigPath string

	rootCmd := &cobra.Command{
		Use:     "cosmos-wallets-exporter --config [config path]",
		Long:    "A Prometheus exporter that returns wallets balances on cosmos-sdk chains.",
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			ExecuteMain(ConfigPath)
		},
	}

	validateConfigCmd := &cobra.Command{
		Use:     "validate-config --config [config path]",
		Long:    "Validate config.",
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			ExecuteValidateConfig(ConfigPath)
		},
	}

	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "", "Config file path")
	if err := rootCmd.MarkPersistentFlagRequired("config"); err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Could not set flag as required")
	}

	validateConfigCmd.PersistentFlags().StringVar(&ConfigPath, "config", "", "Config file path")
	if err := validateConfigCmd.MarkPersistentFlagRequired("config"); err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Could not set flag as required")
	}

	rootCmd.AddCommand(validateConfigCmd)

	if err := rootCmd.Execute(); err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Could not start application")
	}
}
