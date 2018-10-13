package collector

import (
	"context"

	"strconv"
	"strings"
	"sync"
	"time"

	hetzner "github.com/andrexus/go-hetzner-robot"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type hetznerRobotCollector struct {
	client *hetzner.Client

	prices               *prometheus.GaugeVec
	servers              map[int]hetzner.Product
	deletedServers       map[int]hetzner.Product
	deletedServersNotify chan int
	mutex                sync.Mutex
	logger               *log.Entry
}

const (
	metricNamespace   = "hetzner"
	metricSubsystem   = "server_market"
	metricName        = "price"
	metricDescription = "Monthly price in euros"
	exporterName      = "hetzner-server-market-exporter"
)

var serverDefaultLabels = []string{"id", "name", "description", "traffic", "dist", "arch", "lang", "cpu", "cpu_benchmark", "memory_size", "hdd_size", "hdd_text", "hdd_count", "datacenter", "network_speed", "fixed_price"}

//NewHetznerRobotCollector returns new instance of hetznerRobotCollector
func NewHetznerRobotCollector(client *hetzner.Client, refreshIntervalSeconds uint, logger *log.Entry) prometheus.Collector {
	pricesVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricNamespace,
		Subsystem: metricSubsystem,
		Name:      metricName,
		Help:      metricDescription,
	}, serverDefaultLabels)

	collector := &hetznerRobotCollector{
		client:               client,
		prices:               pricesVec,
		servers:              make(map[int]hetzner.Product, 0),
		deletedServers:       make(map[int]hetzner.Product, 0),
		deletedServersNotify: make(chan int),
		logger:               logger,
	}
	go collector.updateServersMap(time.Now())
	ticker := time.NewTicker(time.Duration(refreshIntervalSeconds) * time.Second)
	logger.WithField("refresh_interval", refreshIntervalSeconds).Debug("fetching Hetzner Robot API")
	go func(c *hetznerRobotCollector) {
		go c.collectDeletedServers()
		for t := range ticker.C {
			err := c.updateServersMap(t)
			if err != nil {
				logger.WithField("error", err).Error("could not fetch server market products")
			}
		}
	}(collector)
	return collector
}

func (c *hetznerRobotCollector) Describe(ch chan<- *prometheus.Desc) {
	c.prices.Describe(ch)
}

func (c *hetznerRobotCollector) Collect(ch chan<- prometheus.Metric) {
	// Delete metric for servers that don't exist anymore
	for _, server := range c.deletedServers {
		c.prices.DeleteLabelValues(extractServerLabels(&server)...)
		c.deletedServersNotify <- server.ID
	}
	for _, server := range c.servers {
		gauge := c.prices.WithLabelValues(extractServerLabels(&server)...)
		price, err := strconv.ParseFloat(server.PriceVat, 32)
		if err != nil {
			c.logger.WithFields(log.Fields{
				"price": server.PriceVat,
				"error": err,
			}).Warn("could not convert price string to float")
			continue
		}
		gauge.Set(price)
		gauge.Collect(ch)
	}
}

func (c *hetznerRobotCollector) updateServersMap(t time.Time) error {
	c.logger.WithField("fetch_time", t.Format(time.RFC3339)).Debug("fetching server market products")
	products, err := c.fetchServerMarketProducts()
	if err != nil {
		return err
	}
	productIDs := make([]int, len(products))
	for _, product := range products {
		productIDs = append(productIDs, product.ID)
	}
	c.logger.WithField("total", len(products)).Debug("products fetched")
	for _, product := range products {
		c.mutex.Lock()
		if _, ok := c.servers[product.ID]; !ok {
			c.servers[product.ID] = product
		}
		c.mutex.Unlock()
	}
	for id, server := range c.servers {
		if !contains(productIDs, id) {
			c.logger.WithField("id", id).Debug("server was deleted")
			c.deletedServers[id] = server
		}
	}
	return nil
}

func (c *hetznerRobotCollector) collectDeletedServers() {
	log.Debug("start collecting deleted servers")
	for {
		select {
		case id := <-c.deletedServersNotify:
			c.logger.WithField("id", id).Debug("collecting deleted server")
			delete(c.deletedServers, id)
		}
	}
}

func (c *hetznerRobotCollector) fetchServerMarketProducts() ([]hetzner.Product, error) {
	//opts := &hetzner.ProductSearchRequest{MinMemorySize: "32", MaxPrice: "30", Search: "ssd"}
	opts := &hetzner.ProductSearchRequest{}
	servers, _, err := c.client.Order.ListServerMarketProducts(context.Background(), opts)
	return servers, err
}

func extractServerLabels(server *hetzner.Product) []string {
	id := strconv.Itoa(server.ID)
	name := server.Name
	description := strings.Join(server.Description, "; ")
	traffic := server.Traffic
	dist := strings.Join(server.Dist, "; ")
	archStr := make([]string, 0)
	for _, arch := range server.Arch {
		archStr = append(archStr, strconv.Itoa(arch))
	}
	arch := strings.Join(archStr, "; ")
	lang := strings.Join(server.Lang, "; ")
	cpu := server.CPU
	cpuBenchmark := strconv.Itoa(server.CPUBenchmark)
	memorySize := strconv.Itoa(server.MemorySize)
	hddSize := strconv.Itoa(server.HddSize)
	hddText := server.HddText
	hddCount := strconv.Itoa(server.HddCount)
	datacenter := server.Datacenter
	networkSpeed := server.NetworkSpeed
	fixedPrice := strconv.FormatBool(server.FixedPrice)

	return []string{id, name, description, traffic, dist, arch, lang, cpu, cpuBenchmark, memorySize, hddSize, hddText, hddCount, datacenter, networkSpeed, fixedPrice}
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
