package config

import "github.com/guregu/null/v5"

type TracingConfig struct {
	Enabled                   null.Bool `default:"false"                     toml:"enabled"`
	OpenTelemetryHTTPHost     string    `toml:"open-telemetry-http-host"`
	OpenTelemetryHTTPInsecure null.Bool `default:"true"                      toml:"open-telemetry-http-insecure"`
	OpenTelemetryHTTPUser     string    `toml:"open-telemetry-http-user"`
	OpenTelemetryHTTPPassword string    `toml:"open-telemetry-http-password"`
}
