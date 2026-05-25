package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	queriesIngestedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "search_queries_ingested_total",
		Help: "Total number of ingested search queries",
	})
	queryIngestErrorsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "search_queries_ingest_errors_total",
		Help: "Total number of errors while ingesting search queries",
	})
	topRequestsDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "top_requests_duration_seconds",
		Help:    "Duration of top requests handler",
		Buckets: prometheus.DefBuckets,
	})
	topRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "top_requests_http_total",
		Help: "Total number of top requests HTTP calls by status code",
	}, []string{"status"})
)

func init() {
	prometheus.MustRegister(queriesIngestedTotal, queryIngestErrorsTotal, topRequestsDuration, topRequestsTotal)
}

func IncQueryIngested() {
	queriesIngestedTotal.Inc()
}

func IncQueryIngestError() {
	queryIngestErrorsTotal.Inc()
}

func ObserveTopRequests(duration time.Duration, status int) {
	topRequestsDuration.Observe(duration.Seconds())
	topRequestsTotal.WithLabelValues(strconv.Itoa(status)).Inc()
}
