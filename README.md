# rwb-test-task

Kafka consumer + HTTP API for top search requests in the last 5 minutes.

## API

`GET /api/v1/top-requests?n=10` returns the top N запросов за последние 5 минут,
используя timestamp сообщений Kafka как время события.

## Dependencies

Kafka, Redis.

## Environment

- `KAFKA_BROKER` (comma-separated)
- `KAFKA_TOPICS` (comma-separated)
- `KAFKA_GROUP_ID`
- `REDIS_ADDR`
- `REDIS_PASSWORD` (optional)
- `HTTP_ADDR` (optional, default `:8080`)
