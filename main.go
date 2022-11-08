package main

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func Execute(configPath string) {
	config, err := GetConfig(configPath)
	if err != nil {
		GetDefaultLogger().Fatal().Err(err).Msg("Could not load config")
	}

	if err = config.Validate(); err != nil {
		GetDefaultLogger().Fatal().Err(err).Msg("Provided config is invalid!")
	}

	log := GetLogger(config.LogConfig)
	manager := NewManager(*config, log)

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		Handler(w, r, manager, log)
	})

	log.Info().Str("addr", config.ListenAddress).Msg("Listening")
	err = http.ListenAndServe(config.ListenAddress, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not start application")
	}
}

func Handler(w http.ResponseWriter, r *http.Request, manager *Manager, log *zerolog.Logger) {
	requestStart := time.Now()

	sublogger := log.With().
		Str("request-id", uuid.New().String()).
		Logger()

	successGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_wallets_exporter_success",
			Help: "Whether a scrape was successful",
		},
		[]string{"chain", "address", "name", "group"},
	)

	timingsGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_wallets_exporter_timings",
			Help: "External LCD query timing",
		},
		[]string{"chain", "address", "name", "group"},
	)

	balancesGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_wallets_exporter_balance",
			Help: "A wallet balance (in tokens)",
		},
		[]string{"chain", "address", "name", "group", "denom"},
	)

	usdBalancesGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_wallets_exporter_balance_usd",
			Help: "A wallet balance (in USD)",
		},
		[]string{"chain", "address", "name", "group"},
	)

	registry := prometheus.NewRegistry()
	registry.MustRegister(successGauge)
	registry.MustRegister(timingsGauge)
	registry.MustRegister(balancesGauge)
	registry.MustRegister(usdBalancesGauge)

	balances := manager.GetAllBalances()
	for _, balance := range balances {
		successGauge.With(prometheus.Labels{
			"chain":   balance.Chain,
			"address": balance.Wallet.Address,
			"name":    balance.Wallet.Name,
			"group":   balance.Wallet.Group,
		}).Set(BoolToFloat64(balance.Success))

		timingsGauge.With(prometheus.Labels{
			"chain":   balance.Chain,
			"address": balance.Wallet.Address,
			"name":    balance.Wallet.Name,
			"group":   balance.Wallet.Group,
		}).Set(balance.Duration.Seconds())

		if !balance.Success {
			continue
		}

		if balance.UsdPrice != 0 {
			usdBalancesGauge.With(prometheus.Labels{
				"chain":   balance.Chain,
				"address": balance.Wallet.Address,
				"name":    balance.Wallet.Name,
				"group":   balance.Wallet.Group,
			}).Set(balance.UsdPrice)
		}

		for _, singleBalance := range balance.Balances {
			balancesGauge.With(prometheus.Labels{
				"chain":   balance.Chain,
				"address": balance.Wallet.Address,
				"name":    balance.Wallet.Name,
				"group":   balance.Wallet.Group,
				"denom":   singleBalance.Denom,
			}).Set(StrToFloat64(singleBalance.Amount))
		}
	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)

	sublogger.Info().
		Str("method", http.MethodGet).
		Str("endpoint", "/metrics").
		Float64("request-time", time.Since(requestStart).Seconds()).
		Msg("Request processed")
}

func main() {
	var ConfigPath string

	rootCmd := &cobra.Command{
		Use:  "cosmos-wallets-exporter",
		Long: "Checks the specific wallets on different chains for proposal votes.",
		Run: func(cmd *cobra.Command, args []string) {
			Execute(ConfigPath)
		},
	}

	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "", "Config file path")
	if err := rootCmd.MarkPersistentFlagRequired("config"); err != nil {
		GetDefaultLogger().Fatal().Err(err).Msg("Could not set flag as required")
	}

	if err := rootCmd.Execute(); err != nil {
		GetDefaultLogger().Fatal().Err(err).Msg("Could not start application")
	}
}
