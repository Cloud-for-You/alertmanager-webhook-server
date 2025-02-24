package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	receiver "github.com/cloud-for-you/alertmanager-webhook-server/pkg/receivers/kafka"
)

var (
	debug bool

	receivedMessages = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "received_messages_total",
			Help: "Celkový počet přijatých zpráv",
		},
	)

	sentMessages = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sent_messages_total",
			Help: "Celkový počet odeslaných zpráv",
		},
		[]string{"status"},
	)

	messageSendDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "message_send_duration_seconds",
			Help:    "Doba trvání odesílání zpráv v sekundách",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"status"},
	)
)

func init() {
	debug = os.Getenv("DEBUG") == "true"

	// Registrace metrik
	prometheus.MustRegister(receivedMessages)
	prometheus.MustRegister(sentMessages)
	prometheus.MustRegister(messageSendDuration)
}

func debugLog(data []byte) {
	if debug {
		log.Printf("Přijatý webhook: %s", string(data))
	}
}

func alertHandler(w http.ResponseWriter, r *http.Request) {
    // Parsování JSON alertu do struktury Alertmanageru
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Chyba při čtení requestu", http.StatusBadRequest)
        return
    }
    defer r.Body.Close() // Ensure the body is closed after reading

    debugLog(body)
    receivedMessages.Inc()

    var alertData template.Data

    err = json.Unmarshal(body, &alertData)
    if err != nil {
        http.Error(w, "Chyba při parsování JSON", http.StatusBadRequest)
        return
    }

    // Odeslání do Kafka topicu
    for _, alert := range alertData.Alerts {
        message := fmt.Sprintf("Alert: %s, Status: %s", alert.Labels["alertname"], alert.Status)
        start := time.Now()
        err = receiver.SendMessage(r.Context(), message)
        duration := time.Since(start).Seconds()
        if err != nil {
            sentMessages.WithLabelValues("error").Inc()
            messageSendDuration.WithLabelValues("error").Observe(duration)
            http.Error(w, "Chyba při odesílání do Kafka", http.StatusInternalServerError)
            return
        }
        sentMessages.WithLabelValues("success").Inc()
        messageSendDuration.WithLabelValues("success").Observe(duration)
    }

    w.WriteHeader(http.StatusOK)
}

func main() {
	config := receiver.KafkaConfig{
		ClientCertPath: os.Getenv("KAFKA_CLIENT_CERT"),
		ClientKeyPath:  os.Getenv("KAFKA_CLIENT_KEY"),
		CACertPath:     os.Getenv("KAFKA_CA_CERT"),
		KafkaBrokerURL: os.Getenv("KAFKA_BROKER_URL"),
		KafkaTopic:     os.Getenv("KAFKA_TOPIC"),
	}

	receiver.InitKafka(config)

	http.HandleFunc("/", alertHandler)
	http.Handle("/metrics", promhttp.Handler())
	log.Println("Start webserver on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}