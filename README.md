# rwb-test-task

Kafka consumer + HTTP API for top search requests in the last 5 minutes.

## API

`GET /api/v1/top-requests?n=10` returns the top N запросов за последние 5 минут.

## Dependencies

Kafka, Redis.

## Environment

- `KAFKA_BROKER`
- `KAFKA_TOPICS`
- `KAFKA_GROUP_ID`
