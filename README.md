# alertmanager-webhook-server

Tento projekt poskytuje webhook server pro Prometheus Alertmanager, který přijímá alerty pomocí webhook a odesílá je do různých receiverů, jako je Kafka a Stdout.

## Typy receiverů
### Stdout
Data, která jsou přijata pomocí HTTP webhooku, jsou zalogovány na stdout

### Kafka
Data jsou přeposílány do Kafka topicu.

### Centreon
Data jsou do centreon posílána pomocí protokolu NRDP.

## Podporované ENV proměnné

Níže je uvedena tabulka s proměnnými prostředí, které jsou vyžadovány pro jednotlivé receivery.

| Proměnná                       | Popis                                                  | Default     | Receiver  |
|--------------------------------|--------------------------------------------------------|-------------|-----------|
| `APP_ENV` .                    | prostředí pro logování (production, development)       | development |           |
| `INSECURE_SKIP_VERIFY`         | Umožňuje vypnout verifikaci serverového certifikátu    | false       |           |
| `RECEIVER_TYPE`                | specifikace receiveru                                  | stdout      |           |
| `KAFKA_BROKER_URL`             | URL Kafka brokera včetně protokolu                     |             | kafka     |
| `KAFKA_TOPIC`                  | Název Kafka topicu                                     |             | kafka     |
| `KAFKA_CLIENT_CERT`            | Cesta k souboru s klientským certifikátem              |             | kafka     |
| `KAFKA_CLIENT_KEY`             | Cesta k souboru s klientským klíčem                    |             | kafka     |
| `KAFKA_CA_CERT`                | Cesta k souboru s CA certifikátem                      |             | kafka     |
| `CENTREON_NRDP_URL`            | URL Endpoint pro NRDP protokol                         |             | centreon  |
| `CENTREON_NRDP_TOKEN`          | Autentizační token                                     |             | centreon  |
| `CENTREON_MONITORING_HOSTNAME` | Hostname v centreonu, na který se budou alerty párovat |             | centreon  |


## Příklady hodnot proměnných prostředí

### Kafka Receiver

```yaml
env:
  - name: RECEIVER_TYPE
    value: "kafka"
  - name: KAFKA_BROKER_URL
    value: "tls://your-kafka-broker-url:9093"
  - name: KAFKA_TOPIC
    value: "topic-name"
  - name: KAFKA_CLIENT_CERT
    value: "/etc/kafka/certs/tls.crt"
  - name: KAFKA_CLIENT_KEY
    value: "/etc/kafka/certs/tls.key"
  - name: KAFKA_CA_CERT
    value: "/etc/kafka/certs/ca.crt"
```