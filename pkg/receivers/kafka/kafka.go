package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"

	"github.com/segmentio/kafka-go"
)

var Writer *kafka.Writer

type KafkaConfig struct {
    ClientCertPath  string `json:"clientCertPath"`
    ClientKeyPath   string `json:"clientKeyPath"`
    CACertPath      string `json:"caCertPath"`
    KafkaBrokerURL  string `json:"kafkaBrokerURL"`
    KafkaTopic      string `json:"kafkaTopic"`
}

func InitKafka(config KafkaConfig) {
    // Načtení TLS certifikátů pro Kafka klienta
    cert, err := tls.LoadX509KeyPair(config.ClientCertPath, config.ClientKeyPath)
    if err != nil {
        log.Fatalf("Chyba při načítání Kafka certifikátu: %v", err)
    }

    caCert, err := os.ReadFile(config.CACertPath)
    if err != nil {
        log.Fatalf("Chyba při načítání Kafka CA certifikátu: %v", err)
    }

    caCertPool := x509.NewCertPool()
    caCertPool.AppendCertsFromPEM(caCert)

    // Kafka TLS konfigurace
    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
        RootCAs:      caCertPool,
    }

    // Kafka Writer
    Writer = &kafka.Writer{
        Addr:     kafka.TCP(config.KafkaBrokerURL),
        Topic:    config.KafkaTopic,
        Balancer: &kafka.LeastBytes{},
        Transport: &kafka.Transport{
            TLS: tlsConfig,
        },
    }
}

func SendMessage(ctx context.Context, message string) error {
    err := Writer.WriteMessages(ctx, kafka.Message{
        Value: []byte(message),
    })
    return err
}