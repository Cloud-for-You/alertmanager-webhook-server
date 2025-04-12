# alertmanager-webhook-server

Tento projekt poskytuje webhook server pro Prometheus Alertmanager, který přijímá alerty pomocí webhook a odesílá je do různých receiverů, jako je Kafka a Stdout.

## Typy receiverů
### Stdout
Data, která jsou přijata pomocí HTTP webhooku, jsou zalogovány na stdout

### Kafka
Data jsou přeposílány do Kafka topicu.

### MSTeams
Data jsou přeposílány do MSTeams s podporou Power Automate. 

### Podporované ENV proměnné

Níže je uvedena tabulka s proměnnými prostředí, které jsou vyžadovány pro jednotlivé receivery.

| Proměnná              | Popis                                            | Default     | Receiver  |
|-----------------------|--------------------------------------------------|-------------|-----------|
| `APP_ENV` .           | prostředí pro logování (production, development) | development |           |
| `RECEIVER_TYPE`       | specifikace receiveru (stdout, kafka, msteams)   | stdout      |           |
| `KAFKA_BROKER_URL`    | URL Kafka brokera včetně protokolu               |             | kafka     |
| `KAFKA_TOPIC`         | Název Kafka topicu                               |             | kafka     |
| `KAFKA_CLIENT_CERT`   | Cesta k souboru s klientským certifikátem        |             | kafka     |
| `KAFKA_CLIENT_KEY`    | Cesta k souboru s klientským klíčem              |             | kafka     |
| `KAFKA_CA_CERT`       | Cesta k souboru s CA certifikátem                |             | kafka     |
| `MSTEAMS_WEBHOOK_URL` | URL k webhooku do MS Teams kanálu                |             | msteams   |
