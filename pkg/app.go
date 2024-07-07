package pkg

import (
	"context"
	coingeckoPkg "main/pkg/coingecko"
	"main/pkg/config"
	"main/pkg/logger"
	queriersPkg "main/pkg/queriers"
	"main/pkg/tracing"
	"main/pkg/types"
	"net/http"
	"sync"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

type App struct {
	Config   *config.Config
	Logger   zerolog.Logger
	Queriers []types.Querier

	Tracer trace.Tracer
}

func NewApp(configPath string, version string) *App {
	appConfig, err := config.GetConfig(configPath)
	if err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Could not load config")
	}

	if err = appConfig.Validate(); err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Provided config is invalid!")
	}

	tracer := tracing.InitTracer(appConfig.TracingConfig, version)
	log := logger.GetLogger(appConfig.LogConfig)
	coingecko := coingeckoPkg.NewCoingecko(appConfig, log, tracer)

	queriers := []types.Querier{
		queriersPkg.NewPriceQuerier(appConfig, coingecko, tracer),
		queriersPkg.NewBalanceQuerier(appConfig, log, tracer),
		queriersPkg.NewUptimeQuerier(tracer),
	}

	return &App{
		Config:   appConfig,
		Logger:   log,
		Queriers: queriers,
		Tracer:   tracer,
	}
}

func (a *App) Start() {
	otelHandler := otelhttp.NewHandler(http.HandlerFunc(a.Handler), "prometheus")
	http.Handle("/metrics", otelHandler)

	a.Logger.Info().Str("addr", a.Config.ListenAddress).Msg("Listening")
	err := http.ListenAndServe(a.Config.ListenAddress, nil)
	if err != nil {
		a.Logger.Fatal().Err(err).Msg("Could not start application")
	}
}

func (a *App) Handler(w http.ResponseWriter, r *http.Request) {
	requestStart := time.Now()
	requestID := uuid.New().String()

	sublogger := a.Logger.With().
		Str("request-id", requestID).
		Logger()

	span := trace.SpanFromContext(r.Context())
	span.SetAttributes(attribute.String("request-id", requestID))
	rootSpanCtx := r.Context()

	defer span.End()

	registry := prometheus.NewRegistry()

	var wg sync.WaitGroup
	var mutex sync.Mutex

	var queryInfos []types.QueryInfo

	for _, querier := range a.Queriers {
		wg.Add(1)
		go func(querier types.Querier, ctx context.Context) {
			metrics, querierQueryInfos := querier.GetMetrics(ctx)

			mutex.Lock()
			registry.MustRegister(metrics...)
			queryInfos = append(queryInfos, querierQueryInfos...)
			mutex.Unlock()
			wg.Done()
		}(querier, rootSpanCtx)
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
