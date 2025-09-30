package main


import (
    "flag"
    "fmt"
    "log"
    "net/http"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"

    "rds_exporter/collector"
)

func main() {
    // Command-line flag for custom port
    port := flag.Int("port", 9195, "Port to expose metrics on")
    flag.Parse()

    collector := collector.NewRDSSessionCollector()
    prometheus.MustRegister(collector)

    http.Handle("/metrics", promhttp.Handler())
    addr := fmt.Sprintf(":%d", *port)
    fmt.Printf("Starting RDS Prometheus exporter on %s\n", addr)
    log.Fatal(http.ListenAndServe(addr, nil))
}
