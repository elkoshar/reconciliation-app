package server

import (
	httpapi "github.com/elkoshar/reconciliation-app/api/http"
	config "github.com/elkoshar/reconciliation-app/configs"
)

// Init to initiate all DI for service handler implementation
func InitHttp(config *config.Config) error {

	httpserver := httpapi.Server{
		Cfg: config,
	}

	return runHTTPServer(httpserver, config.ServerHttpPort)
}
