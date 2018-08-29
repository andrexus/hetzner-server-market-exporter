package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	hetzner "github.com/andrexus/go-hetzner-robot"
	"github.com/andrexus/hetzner-server-market-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	addr            = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	credentialsPath = flag.String("robot-api-credentials", "hetzner-robot-creds.json", "The path to the Hetzner Robot API credentials.")
	refreshInterval = flag.Int("refresh-interval", 600, "Fetch Hetzner Robot API each [interval] seconds.")
)

type apiCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	flag.Parse()
	username, password, err := parseCredentials(*credentialsPath)
	if err != nil {
		log.Fatalf("[ERROR] could not load API credentials: %v", err)
	}
	client := hetzner.NewClient(username, password, nil)
	refreshIntervalSeconds := uint(*refreshInterval)
	if refreshIntervalSeconds < 8 {
		log.Printf("[WARN] potential risk of exceeding API requests limit (500 per hour) if refresh-interval < 8 (seconds). Current: %d seconds", refreshIntervalSeconds)
	}

	r := prometheus.NewRegistry()
	r.MustRegister(collector.NewHetznerRobotCollector(client, refreshIntervalSeconds))

	http.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	log.Printf("[INFO] listening on %s\n", *addr)
	log.Printf("[INFO] metrics available under /metrics\n")
	log.Fatal(http.ListenAndServe(*addr, nil))
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
