package server

import (
	httpapi "github.com/elkoshar/reconciliation-app/api/http"
	config "github.com/elkoshar/reconciliation-app/configs"
	"github.com/elkoshar/reconciliation-app/service/reconciliation"
)

// Init to initiate all DI for service handler implementation
func InitHttp(config *config.Config) error {

	reconService := reconciliation.NewReconciliationService()
	httpserver := httpapi.Server{
		Cfg:   config,
		Recon: reconService,
	}

	return runHTTPServer(httpserver, config.ServerHttpPort)
}
