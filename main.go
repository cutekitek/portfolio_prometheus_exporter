package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/cutekitek/portfolio_prometheus_exporter/config"
	"github.com/cutekitek/portfolio_prometheus_exporter/exporters/tinkoff"
	"github.com/cutekitek/portfolio_prometheus_exporter/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	f = flag.String("config", "config.yml", "YAML config file name")
)

func main() {
	flag.Parse()
	cfg, err := config.ConfigFromYaml(*f)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	reg := prometheus.NewRegistry()
	ctx := context.Background()

	tinkoffCfg := cfg.Exchanges["tinkoff"]
	tinkoffManager := metrics.NewMetricsManager(reg, "tinkoff")

	exporter, err := tinkoff.NewTinkoffExporter(ctx, tinkoffCfg, tinkoffManager)
	if err != nil {
		panic(err)
	}

	go exporter.Scraper(ctx)

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	slog.Info("starting exporter", "address", cfg.ListenAddress())
	http.ListenAndServe(cfg.ListenAddress(), nil)
}
