package config

import "time"

type (
	// Config will holds mapped key value for service configuration
	Config struct {
		AppVersion                    string        `mapstructure:"APP_VERSION"`
		ServerHttpPort                string        `mapstructure:"SERVER_HTTP_PORT"`
		LogLevel                      string        `mapstructure:"LOG_LEVEL"`
		LogFormat                     string        `mapstructure:"LOG_FORMAT"`
		HttpReadTimeout               time.Duration `mapstructure:"HTTP_READ_TIMEOUT"`
		HttpWriteTimeout              time.Duration `mapstructure:"HTTP_WRITE_TIMEOUT"`
		HttpInboundTimeout            time.Duration `mapstructure:"HTTP_INBOUND_TIMEOUT"`
		HTTPTimeout                   time.Duration `mapstructure:"HTTP_TIMEOUT"`
		HTTPDebug                     bool          `mapstructure:"HTTP_DEBUG"`
		HTTPMaxIdleConnections        int           `mapstructure:"HTTP_MAX_IDLE_CONNECTIONS"`
		HTTPMaxIdleConnectionsPerHost int           `mapstructure:"HTTP_MAX_IDLE_CONNECTIONS_PER_HOST"`
		HTTPIdleConnectionTimeout     time.Duration `mapstructure:"HTTP_IDLE_CONNECTION_TIMEOUT"`
	}
)
