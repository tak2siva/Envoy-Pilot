package metrics

import (
	"Envoy-Pilot/cmd/server/model"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// metrics
var METRICS_ACTIVE_CONNECTIONS = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "xds_active_connections",
	Help: "Xds Active Connections",
},
	[]string{
		"cluster",
	})

var METRICS_ACTIVE_SUBSCRIBERS = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "xds_active_subscribers",
	Help: "Xds Active Subscribers By Topic",
},
	[]string{
		"cluster",
		"type",
	})

func IncActiveConnections(en *model.EnvoySubscriber) {
	METRICS_ACTIVE_CONNECTIONS.With(prometheus.Labels{
		"cluster": en.Cluster,
	}).Inc()
}

func DecActiveConnections(en *model.EnvoySubscriber) {
	METRICS_ACTIVE_CONNECTIONS.With(prometheus.Labels{
		"cluster": en.Cluster,
	}).Dec()
}

func IncActiveSubscribers(en *model.EnvoySubscriber, topic string) {
	METRICS_ACTIVE_SUBSCRIBERS.With(prometheus.Labels{
		"cluster": en.Cluster,
		"type":    topic,
	}).Inc()
}

func DecActiveSubscribers(en *model.EnvoySubscriber) {
	if en.IsADS() {
		for topic := range en.AdsList {
			METRICS_ACTIVE_SUBSCRIBERS.With(prometheus.Labels{
				"cluster": en.Cluster,
				"type":    topic,
			}).Dec()
		}
	} else {
		METRICS_ACTIVE_SUBSCRIBERS.With(prometheus.Labels{
			"cluster": en.Cluster,
			"type":    en.SubscribedTo,
		}).Dec()
	}
}

func StartMetricsServer() {
	prometheus.MustRegister(METRICS_ACTIVE_CONNECTIONS)
	prometheus.MustRegister(METRICS_ACTIVE_SUBSCRIBERS)
	http.Handle("/metrics", promhttp.Handler())

	log.Println("Starting metrics server on :8081..")
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
