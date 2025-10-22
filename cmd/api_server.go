package cmd

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Depado/ginprom"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/rm-hull/godx"
	"github.com/rm-hull/placenames-api/internal"
	"github.com/rm-hull/placenames-api/internal/routes"
	healthcheck "github.com/tavsec/gin-healthcheck"
	"github.com/tavsec/gin-healthcheck/checks"
	hc_config "github.com/tavsec/gin-healthcheck/config"
	cachecontrol "go.eigsys.de/gin-cachecontrol/v2"
)

func ApiServer(filePath string, port int, debug bool) error {

	godx.GitVersion()
	godx.EnvironmentVars()
	godx.UserInfo()

	trie, err := internal.LoadData(filePath)
	if err != nil {
		return fmt.Errorf("error loading data: %w", err)
	}

	r := gin.New()

	prometheus := ginprom.New(
		ginprom.Engine(r),
		ginprom.Path("/metrics"),
		ginprom.Ignore("/healthz"),
	)

	r.Use(
		gin.Recovery(),
		gin.LoggerWithWriter(gin.DefaultWriter, "/healthz", "/metrics"),
		prometheus.Instrument(),
		cachecontrol.New(cachecontrol.Config{
			MaxAge:    cachecontrol.Duration(28 * 24 * time.Hour),
			Immutable: true,
			Public:    true,
		}),
		cors.Default(),
	)

	if debug {
		log.Println("WARNING: pprof endpoints are enabled and exposed. Do not run with this flag in production.")
		pprof.Register(r)
	}

	if err = healthcheck.New(r, hc_config.DefaultConfig(), []checks.Check{}); err != nil {
		return fmt.Errorf("failed to initialize healthcheck: %w", err)
	}

	v1 := r.Group("/v1")
	v1.GET("/place-names/prefix/:query", routes.Prefix(trie))

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting HTTP API Server on port %d...", port)
	if err := r.Run(addr); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start HTTP API Server on port %d: %w", port, err)
	}
	return nil
}
