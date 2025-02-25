package kafka

import (
	"crypto/tls"
	"crypto/x509"
	"os"

	"github.com/IBM/sarama"
	"github.com/cloud-for-you/alertmanager-webhook-server/internal/logger"
)

type KafkaClient struct {
	producer sarama.SyncProducer
	topic    string
}

func Client() *KafkaClient {
	brokerURL := os.Getenv("KAFKA_BROKER_URL")
	clientCertFile := os.Getenv("KAFKA_CLIENT_CERT")
	clientKeyFile := os.Getenv("KAFKA_CLIENT_KEY")
	caCertFile := os.Getenv("KAFKA_CA_CERT")
	topic := os.Getenv("KAFKA_TOPIC")

	// Load client cert
	clientCert, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
	if err != nil {
		logger.Log.Fatalf("Failed to load client certificate: %v", err)
	}

	// Load CA cert
	caCert, err := os.ReadFile(caCertFile)
	if err != nil {
		logger.Log.Fatalf("Failed to read CA certificate: %v", err)
	}

	// Create cert pool
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Create TLS config
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      caCertPool,
	}

	// Configure Sarama
	config := sarama.NewConfig()
	config.Net.TLS.Enable = true
	config.Net.TLS.Config = tlsConfig
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Metadata.Retry.Max = 5

	// Create producer
	producer, err := sarama.NewSyncProducer([]string{brokerURL}, config)
	if err != nil {
		logger.Log.Fatalf("Failed to create Kafka producer: %v", err)
	}

	return &KafkaClient{
		producer: producer,
		topic:    topic,
	}
}

func (r *KafkaClient) SendMessage(data []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: r.topic,
		Value: sarama.ByteEncoder(data),
	}

	partition, offset, err := r.producer.SendMessage(msg)
	if err != nil {
		logger.Log.Errorf("Failed to send message to Kafka: %v", err)
		return err
	}

	logger.Log.Infof("Message sent to Kafka topic %s: %s, partition: %d, offset: %d", r.topic, string(data), partition, offset)
	return nil
}
