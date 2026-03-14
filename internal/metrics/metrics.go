package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	CacheHits = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cache_hits_total",
		Help: "total number of cache hits",
	})

	CacheMisses = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cache_misses_total",
		Help: "total number of cache misses",
	})

	CacheEvictions = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cache_evictions_total",
		Help: "total number of cache evictions",
	})

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_duration_seconds",
			Help:    "gRPC request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	CacheSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cache_size_current",
		Help: "current number of items in the cache",
	})
)

func Init() {
	prometheus.MustRegister(CacheHits)
	prometheus.MustRegister(CacheMisses)
	prometheus.MustRegister(CacheEvictions)
	prometheus.MustRegister(RequestDuration)
	prometheus.MustRegister(CacheSize)
}

func StartServer(address string) {
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(address, nil)
}
