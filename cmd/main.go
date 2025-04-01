package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/cloud-for-you/alertmanager-webhook-server/internal/logger"
	"github.com/cloud-for-you/alertmanager-webhook-server/pkg/receivers"
	"github.com/cloud-for-you/alertmanager-webhook-server/pkg/receivers/kafka"
	"github.com/cloud-for-you/alertmanager-webhook-server/pkg/receivers/msteams"
	stdout "github.com/cloud-for-you/alertmanager-webhook-server/pkg/receivers/stdout"
)

var (
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
	logger.InitLogger()
	defer logger.SyncLogger()

	// Registrace metrik
	prometheus.MustRegister(receivedMessages)
	prometheus.MustRegister(sentMessages)
	prometheus.MustRegister(messageSendDuration)
}

func main() {
	// Spuštění webserveru
	http.HandleFunc("/", receiverHandler)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	http.Handle("/metrics", promhttp.Handler())
	logger.Log.Info("Server is running on address :8080")
	logger.Log.Fatal(http.ListenAndServe(":8080", nil))
}

func receiverHandler(w http.ResponseWriter, r *http.Request) {
    // Parsování JSON alertu do struktury Alertmanageru
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Chyba při čtení requestu", http.StatusBadRequest)
        return
    }
    defer r.Body.Close() // Ensure the body is closed after reading

    receivedMessages.Inc()

    var alertData template.Data

    err = json.Unmarshal(body, &alertData)
    if err != nil {
			logger.Log.Error("Chyba při parsování JSON")
      http.Error(w, "Chyba při parsování JSON", http.StatusBadRequest)
      return
    }

    // Odeslání dat do receiveru podle hodnoty v os.Env RECEIVER_TYPE
		receiverType := os.Getenv("RECEIVER_TYPE")
		var receiver receivers.Receiver
		switch receiverType {
    case "stdout":
        receiver = stdout.Client()
		case "kafka":
				receiver = kafka.Client()
		case "msteams":
		    receiver = msteams.Client()
    default:
	      receiver = stdout.Client()
    }

		mainReceiver := receivers.NewMainReceiver(receiver)
		// Simulace příjmu dat
		err = mainReceiver.Receive(body)
    if err != nil {
        fmt.Println("Error:", err)
    }

    w.WriteHeader(http.StatusOK)
		
}
