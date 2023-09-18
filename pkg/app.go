package pkg

import (
	coingeckoPkg "main/pkg/coingecko"
	"main/pkg/config"
	"main/pkg/logger"
	queriersPkg "main/pkg/queriers"
	"main/pkg/types"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

type App struct {
	Config   *config.Config
	Logger   zerolog.Logger
	Queriers []types.Querier
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
	coingecko := coingeckoPkg.NewCoingecko(appConfig, log)

	queriers := []types.Querier{
		queriersPkg.NewPriceQuerier(appConfig, coingecko),
		queriersPkg.NewBalanceQuerier(appConfig, log),
	}

	return &App{
		Config:   appConfig,
		Logger:   log,
		Queriers: queriers,
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

	registry := prometheus.NewRegistry()

	var wg sync.WaitGroup
	var mutex sync.Mutex

	var queryInfos []types.QueryInfo

	for _, querier := range a.Queriers {
		wg.Add(1)
		go func(querier types.Querier) {
			mutex.Lock()

			metrics, querierQueryInfos := querier.GetMetrics()
			registry.MustRegister(metrics...)
			queryInfos = append(queryInfos, querierQueryInfos...)

			mutex.Unlock()
			wg.Done()
		}(querier)
	}

	wg.Wait()

	queriersQuerier := queriersPkg.NewQueriesQuerier(a.Config, queryInfos)
	metrics, _ := queriersQuerier.GetMetrics()
	registry.MustRegister(metrics...)

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)

	sublogger.Info().
		Str("method", http.MethodGet).
		Str("endpoint", "/metrics").
		Float64("request-time", time.Since(requestStart).Seconds()).
		Msg("Request processed")
}
