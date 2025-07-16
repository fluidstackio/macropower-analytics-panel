package main

import (
	"net/http"
	"os"
	"time"

	"github.com/MacroPower/macropower-analytics-panel/server/cacher"
	"github.com/MacroPower/macropower-analytics-panel/server/collector"
	"github.com/MacroPower/macropower-analytics-panel/server/payload"
	"github.com/alecthomas/kong"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
)

var (
	cli struct {
		HTTPAddress        string        `help:"Address to listen on for payloads and metrics." env:"HTTP_ADDRESS" default:":8080"`
		SessionTimeout     time.Duration `help:"The maximum duration that may be added between heartbeats. 0 = auto." type:"time.Duration" env:"SESSION_TIMEOUT" default:"0"`
		MaxCacheSize       int           `help:"The maximum number of sessions to store in the cache before resetting. 0 = unlimited." env:"MAX_CACHE_SIZE" default:"100000"`
		LogFormat          string        `help:"One of: [logfmt, json]." env:"LOG_FORMAT" enum:"logfmt,json" default:"logfmt"`
		LogRaw             bool          `help:"Outputs raw payloads as they are received." env:"LOG_RAW"`
		DisableUserMetrics bool          `help:"Disables user labels in metrics." env:"DISABLE_USER_METRICS"`
		DisableSessionLog  bool          `help:"Disables logging sessions to the console." env:"DISABLE_SESSION_LOG"`
		DisableVariableLog bool          `help:"Disables logging variables to the console." env:"DISABLE_VARIABLE_LOG"`
		GrafanaAuthURL     string        `help:"The URL of the Grafana API server for authentication (e.g., http://grafana:3000/api/auth/keys). If empty, authentication is disabled." env:"GRAFANA_AUTH_URL"`
	}
)

func main() {
	ctx := kong.Parse(
		&cli,
		kong.Name("macropower_analytics_panel_server"),
		kong.Description("A receiver for the macropower-analytics-panel Grafana plugin."),
	)

	logWriter := log.NewSyncWriter(os.Stdout)
	logger := func() log.Logger {
		if cli.LogFormat == "json" || cli.LogRaw {
			return log.NewJSONLogger(logWriter)
		}

		return log.NewLogfmtLogger(logWriter)
	}()

	level.Info(logger).Log(
		"msg", "Starting server for macropower-analytics-panel",
		"version", version.Version,
		"branch", version.Branch,
		"revision", version.Revision,
	)
	level.Info(logger).Log(
		"msg", "Build context",
		"go", version.GoVersion,
		"user", version.BuildUser,
		"date", version.BuildDate,
	)

	cache := cacher.NewCache()
	if cli.MaxCacheSize != 0 {
		go cacher.StartFlusher(cache, cli.MaxCacheSize, logger)
	}

	mux := http.NewServeMux()

	handler := payload.NewHandler(cache, 10, !cli.DisableSessionLog, !cli.DisableVariableLog, cli.LogRaw, cli.GrafanaAuthURL, logger)
	mux.Handle("/write", handler)

	exporter := version.NewCollector("grafana_analytics")
	metricExporter := collector.NewExporter(cache, cli.SessionTimeout, !cli.DisableUserMetrics, logger)
	prometheus.MustRegister(exporter, metricExporter)
	mux.Handle("/metrics", promhttp.Handler())

	err := http.ListenAndServe(cli.HTTPAddress, mux)
	ctx.FatalIfErrorf(err)
}
