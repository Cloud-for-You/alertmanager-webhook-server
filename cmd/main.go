package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	receiver "github.com/cloud-for-you/alertmanager-webhook-server/pkg/receivers/kafka"
	"github.com/prometheus/alertmanager/template"
	"github.com/segmentio/kafka-go"
)

var debug bool

func init() {
    debug = os.Getenv("DEBUG") == "true"
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

    var alertData template.Data

    err = json.Unmarshal(body, &alertData)
    if err != nil {
        http.Error(w, "Chyba při parsování JSON", http.StatusBadRequest)
        return
    }

    // Odeslání do Kafka topicu
    writer := &kafka.Writer{
        Addr:     kafka.TCP(os.Getenv("KAFKA_BROKER_URL")),
        Topic:    "alerts",
        Balancer: &kafka.LeastBytes{},
    }
    defer writer.Close()

    for _, alert := range alertData.Alerts {
        message := fmt.Sprintf("Alert: %s, Status: %s", alert.Labels["alertname"], alert.Status)
        err = writer.WriteMessages(r.Context(), kafka.Message{
            Value: []byte(message),
        })
        if err != nil {
            http.Error(w, "Chyba při odesílání do Kafka", http.StatusInternalServerError)
            return
        }
    }

    w.WriteHeader(http.StatusOK)
}

func main() {
    config := receiver.KafkaConfig{
        ClientCertPath: os.Getenv("KAFKA_CLIENT_CERT"),
        ClientKeyPath:  os.Getenv("KAFKA_CLIENT_KEY"),
        CACertPath:     os.Getenv("KAFKA_CA_CERT"),
        KafkaBrokerURL: os.Getenv("KAFKA_BROKER_URL"),
    }

    receiver.InitKafka(config)

    http.HandleFunc("/", alertHandler)
    log.Println("Start webserver on port 8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}