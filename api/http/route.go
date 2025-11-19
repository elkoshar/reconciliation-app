package http

import (
	"encoding/json"

	"net/http"

	"github.com/elkoshar/reconciliation-app/api"
	config "github.com/elkoshar/reconciliation-app/configs"
	"github.com/elkoshar/reconciliation-app/pkg/helpers"
	"github.com/elkoshar/reconciliation-app/pkg/logger"
	"github.com/elkoshar/reconciliation-app/pkg/panics"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

func root(w http.ResponseWriter, r *http.Request) {
	app := map[string]interface{}{
		"name":        "reconciliation-app",
		"description": "reconciliation-app",
	}

	data, _ := json.Marshal(app)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func handler(cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Heartbeat("/ping"))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(panics.HTTPRecoveryMiddleware)
	r.Use(middleware.Timeout(cfg.HttpInboundTimeout))

	//skip middleware group
	r.Get("/", root)
	if helpers.GetEnvString() != helpers.EnvProduction {
		r.Get("/swagger.json", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "./docs/swagger.json")
		}))

		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("/swagger.json"),
		))

	}

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequestLogger(&logger.CustomLogFormatter{Logger: logger.NewSlogWrapper(cfg)}))

		cors := cors.New(cors.Options{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
			AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
		})
		r.Use(cors.Handler)

		// Test Panics to Slack function
		r.Handle("/panics", panics.CaptureHandler(func(w http.ResponseWriter, r *http.Request) {
			panic("Panics from /test/panics endpoint")
		}))

		r.With(api.InterceptorRequest()).Route("/reconciliation-app", func(r chi.Router) {
			r.Use(api.NewMetricMiddleware())
			// reconciliation group
			r.Route("/reconciliation", func(r chi.Router) {
			})

		})
	})

	return r
}
