package main

import (
	"fmt"
	"log/slog"
	"os"

	config "github.com/elkoshar/reconciliation-app/configs"
	"github.com/elkoshar/reconciliation-app/pkg/helpers"
	"github.com/elkoshar/reconciliation-app/pkg/logger"
	"github.com/elkoshar/reconciliation-app/pkg/panics"
	"github.com/elkoshar/reconciliation-app/server"
)

// @title RECONCILIATION APP API
// @version 0.1
// @description This service is to handle Reconciliation App. For more detail, please visit https://github.com/elkoshar/reconciliation-app
// @contact.name Elko Sharhadi Eppasa
// @contact.url https://github.com/elkoshar/reconciliation-app
// @contact.email elko.s.eppasa@gmail.com
// @BasePath /reconciliation/
func main() {

	var (
		cfg *config.Config
	)

	// init config
	err := config.Init(
		config.WithConfigFile("config"),
		config.WithConfigType("env"),
	)
	if err != nil {
		slog.Warn(fmt.Sprintf("failed to initialize config: %v", err))
		os.Exit(1)
	}
	cfg = config.Get()

	//init logging
	logger.InitLogger(cfg)

	// init send to Slack when panics
	panics.SetOptions(&panics.Options{
		Env: helpers.GetEnvString(),
	})

	fmt.Println("Starting Reconciliation App Service...")
	fmt.Printf("Environment: %s\n", helpers.GetEnvString())
	fmt.Printf("HTTP Server Port: %d\n", cfg.ServerHttpPort)
	fmt.Printf("Config: %v", cfg)

	// init all DI for service handler implementation
	if err := server.InitHttp(cfg); err != nil {
		fmt.Printf("Error starting HTTP server: %v\n", err)
		slog.Error(err.Error())
		os.Exit(1)
	}

}
