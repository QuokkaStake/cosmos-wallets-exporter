package pkg

import (
	"context"
	coingeckoPkg "main/pkg/coingecko"
	"main/pkg/config"
	"main/pkg/fs"
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
	Server   *http.Server
	Tracer   trace.Tracer
}

func NewApp(filesystem fs.FS, configPath string, version string) *App {
	appConfig, err := config.GetConfig(configPath, filesystem)
	if err != nil {
		logger.GetDefaultLogger().Panic().Err(err).Msg("Could not load config")
	}

	if err = appConfig.Validate(); err != nil {
		logger.GetDefaultLogger().Panic().Err(err).Msg("Provided config is invalid!")
	}

	tracer := tracing.InitTracer(appConfig.TracingConfig, version)
	log := logger.GetLogger(appConfig.LogConfig)
	coingecko := coingeckoPkg.NewCoingecko(appConfig, log, tracer)

	queriers := []types.Querier{
		queriersPkg.NewPriceQuerier(appConfig, coingecko, tracer),
		queriersPkg.NewBalanceQuerier(appConfig, log, tracer),
		queriersPkg.NewUptimeQuerier(tracer),
	}

	server := &http.Server{Addr: appConfig.ListenAddress, Handler: nil}

	return &App{
		Config:   appConfig,
		Logger:   log,
		Queriers: queriers,
		Tracer:   tracer,
		Server:   server,
	}
}

func (a *App) Start() {
	otelHandler := otelhttp.NewHandler(http.HandlerFunc(a.Handler), "prometheus")
	handler := http.NewServeMux()
	handler.Handle("/metrics", otelHandler)
	handler.HandleFunc("/healthcheck", a.Healthcheck)
	a.Server.Handler = handler

	a.Logger.Info().Str("addr", a.Config.ListenAddress).Msg("Listening")

	err := a.Server.ListenAndServe()
	if err != nil {
		a.Logger.Panic().Err(err).Msg("Could not start application")
	}
}

func (a *App) Stop() {
	a.Logger.Info().Str("addr", a.Config.ListenAddress).Msg("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = a.Server.Shutdown(ctx)
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

func (a *App) Healthcheck(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("ok"))
}
