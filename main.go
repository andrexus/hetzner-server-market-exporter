package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	hetzner "github.com/andrexus/go-hetzner-robot"
	"github.com/andrexus/hetzner-server-market-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	addr            = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	credentialsPath = flag.String("robot-api-credentials", "hetzner-robot-creds.json", "The path to the Hetzner Robot API credentials.")
	refreshInterval = flag.Int("refresh-interval", 600, "Fetch Hetzner Robot API each [interval] seconds.")
	logFormat       = flag.String("log-format", "text", "Log format [text|json].")
	logLevel        = flag.String("log-level", "info", "Log level.")
)

const exporterName = "hetzner-server-market-exporter"

type apiCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	flag.Parse()
	logger := configureLogging(*logFormat, *logLevel)
	username, password, err := parseCredentials(*credentialsPath)
	if err != nil {
		logger.WithField("error", err).Fatal("could not load API credentials")
	}
	client := hetzner.NewClient(username, password, nil)
	refreshIntervalSeconds := uint(*refreshInterval)
	if refreshIntervalSeconds < 8 {
		logger.WithField("current_refresh_interval_seconds", refreshIntervalSeconds).Warn("potential risk of exceeding API requests limit (500 per hour) if refresh-interval < 8 (seconds)")
	}

	r := prometheus.NewRegistry()
	r.MustRegister(collector.NewHetznerRobotCollector(client, refreshIntervalSeconds, logger))

	http.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	logger.WithField("addr", *addr).Info("server started")
	logger.Info("metrics available under /metrics")
	logger.Fatal(http.ListenAndServe(*addr, nil))
}

func configureLogging(format, level string) *log.Entry {
	textFormatter := &log.TextFormatter{
		FullTimestamp:          true,
		ForceColors:            true,
		DisableLevelTruncation: true,
	}
	log.SetFormatter(textFormatter)
	if format == "json" {
		log.SetFormatter(&log.JSONFormatter{})
	}
	if level, err := log.ParseLevel(strings.ToUpper(level)); err == nil {
		log.SetLevel(level)
	}
	return log.WithFields(log.Fields{"exporter": exporterName})
}

func parseCredentials(path string) (string, string, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer jsonFile.Close()
	bytes, _ := ioutil.ReadAll(jsonFile)

	var c apiCredentials
	if err := json.Unmarshal(bytes, &c); err != nil {
		return "", "", err
	}
	return c.Username, c.Password, nil
}
