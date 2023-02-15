package pkg

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"main/pkg/config"
	"main/pkg/logger"
	"main/pkg/manager"
	"main/pkg/utils"
)

type App struct {
	Config  *config.Config
	Logger  *zerolog.Logger
	Manager *manager.Manager
}

func NewApp(configPath string) *App {
	appConfig, err := config.GetConfig(configPath)
	if err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Could not load config")
	}

	if err = appConfig.Validate(); err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Provided config is invalid!")
	}

	log := logger.GetLogger(appConfig.LogConfig)
	manager := manager.NewManager(appConfig, log)

	return &App{
		Config:  appConfig,
		Logger:  log,
		Manager: manager,
	}
}

func (a *App) Start() {
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		a.Handler(w, r)
	})

	a.Logger.Info().Str("addr", a.Config.ListenAddress).Msg("Listening")
	err := http.ListenAndServe(a.Config.ListenAddress, nil)
	if err != nil {
		a.Logger.Fatal().Err(err).Msg("Could not start application")
	}
}

func (a *App) Handler(w http.ResponseWriter, r *http.Request) {
	requestStart := time.Now()

	sublogger := a.Logger.With().
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

	denomCoefficientGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_wallets_exporter_denom_coefficient",
			Help: "Denom coefficient info",
		},
		[]string{"chain", "denom", "display_denom"},
	)

	registry := prometheus.NewRegistry()
	registry.MustRegister(successGauge)
	registry.MustRegister(timingsGauge)
	registry.MustRegister(balancesGauge)
	registry.MustRegister(usdBalancesGauge)
	registry.MustRegister(denomCoefficientGauge)

	balances := a.Manager.GetAllBalances()
	for _, balance := range balances {
		successGauge.With(prometheus.Labels{
			"chain":   balance.Chain,
			"address": balance.Wallet.Address,
			"name":    balance.Wallet.Name,
			"group":   balance.Wallet.Group,
		}).Set(utils.BoolToFloat64(balance.Success))

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
			}).Set(utils.StrToFloat64(singleBalance.Amount))
		}
	}

	for _, chain := range a.Config.Chains {
		denomCoefficientGauge.With(prometheus.Labels{
			"chain":         chain.Name,
			"display_denom": chain.Denom,
			"denom":         chain.BaseDenom,
		}).Set(float64(chain.DenomCoefficient))
	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)

	sublogger.Info().
		Str("method", http.MethodGet).
		Str("endpoint", "/metrics").
		Float64("request-time", time.Since(requestStart).Seconds()).
		Msg("Request processed")
}
