# rwb-test-task

Kafka consumer + HTTP API for top search requests in the last 5 minutes.

## Local запуск

1. `docker compose up -d --build`
2. Подать события в Kafka: `bash scripts/produce.sh`
3. Проверить API:
   - `GET http://localhost:8080/api/v1/top-requests?n=10`
   - `GET http://localhost:8080/api/v1/stoplist`
   - `GET http://localhost:8080/metrics`

## Примеры запросов к API

`GET /api/v1/top-requests?n=10`
```bash
curl "http://localhost:8080/api/v1/top-requests?n=10"
```

`GET /api/v1/stoplist`
```bash
curl "http://localhost:8080/api/v1/stoplist"
```

`POST /api/v1/stoplist`
```bash
curl -X POST "http://localhost:8080/api/v1/stoplist" \
  -H "Content-Type: application/json" \
  -d '{"word":"spam"}'
```

`DELETE /api/v1/stoplist`
```bash
curl -X DELETE "http://localhost:8080/api/v1/stoplist" \
  -H "Content-Type: application/json" \
  -d '{"word":"spam"}'
```

`GET /metrics`
```bash
curl "http://localhost:8080/metrics"
```

## Контракт данных (Kafka)

**Payload:**
```json
{"request":"iphone"}
```

**Почему эти поля:**
1. `request` — единственный обязательный атрибут для подсчёта топа.
2. Время события берётся из `msg.Timestamp` Kafka. Это позволяет держать окно в 5 минут по event-time, не раздувая payload.
3. Пользовательские идентификаторы не нужны для текущей бизнес-задачи и исключены по соображениям приватности и производительности.

Если в апстриме нет корректного event-time, допускается установка Kafka timestamp в момент формирования события. В противном случае сервис использует `time.Now()` как fallback.

## Обоснование архитектуры

1. **Топ запросов:** Redis ZSET `requests:counts` — быстрый `ZREVRANGE` для чтений (read-heavy), стабильное O(log N + M).
2. **Скользящее окно 5 минут:** секундные бакеты `requests:bucket:<unix>` (HASH) + индекс `requests:buckets` (ZSET). Истечение бакета удаляет вклад в общий счётчик через Lua-скрипт, сохраняя консистентность.
3. **Стоп-лист:** Redis SET `stopList` с O(1) проверкой. Запросы фильтруются по токенам (lowercase + split по не-буквам/цифрам).

## Trade-offs и бизнес-логика

1. **Pre-aggregation ради скорости чтения.** Запросы читаются чаще в 10–50 раз, поэтому вместо вычисления топа “на лету” поддерживается агрегированный ZSET. Это увеличивает стоимость записи, но обеспечивает максимальную скорость чтений.
2. **Гранулярность 1 секунда.** Бакеты по секундам дают точность “до секунды”, экономя память и упрощая очистку. Меньшая гранулярность увеличила бы память и стоимость очистки.
3. **Стоп-лист по токенам.** Токенизация выполняется на записи, а при чтении результаты фильтруются по стоп-словам. Исторические счётчики не пересчитываются при изменении стоп-листа — запретные слова остаются в агрегате, хотя клиент их не видит.
4. **Неопределённость event-time в постановке.** Решение — использовать `msg.Timestamp` Kafka как время события, с fallback на `time.Now()` если timestamp отсутствует.
5. **Неопределённый формат payload.** Контракт минимальный: `request`. Этого достаточно для цели задачи и снижает нагрузку на сеть и парсинг.
6. **Анти-накрутка в постановке не формализована.** Без полей источника (user_id/source) невозможно корректно ограничивать всплески; решение — оставлено как расширение контракта (например, добавить `source` для per-source rate limiting).

## Бенчмарк

```
$ hey -n 500000 -c 200 "http://localhost:8080/api/v1/top-requests?n=10"

Summary:
  Total:        37.9450 secs
  Slowest:      0.4642 secs
  Fastest:      0.0006 secs
  Average:      0.0151 secs
  Requests/sec: 13176.9786
  
  Total data:   174993705 bytes
  Size/request: 349 bytes

Response time histogram:
  0.001 [1]     |
  0.047 [499737]        |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.093 [74]    |
  0.140 [0]     |
  0.186 [14]    |
  0.232 [25]    |
  0.279 [56]    |
  0.325 [12]    |
  0.371 [64]    |
  0.418 [14]    |
  0.464 [3]     |

Latency distribution:
  10%% in 0.0088 secs
  25%% in 0.0115 secs
  50%% in 0.0146 secs
  75%% in 0.0180 secs
  90%% in 0.0216 secs
  95%% in 0.0240 secs
  99%% in 0.0296 secs

Details (average, fastest, slowest):
  DNS+dialup:   0.0000 secs, 0.0000 secs, 0.0304 secs
  DNS-lookup:   0.0000 secs, 0.0000 secs, 0.0282 secs
  req write:    0.0000 secs, 0.0000 secs, 0.0088 secs
  resp wait:    0.0149 secs, 0.0006 secs, 0.4641 secs
  resp read:    0.0001 secs, 0.0000 secs, 0.0116 secs

Status code distribution:
  [200] 500000 responses
```

## Dependencies

Kafka, Redis.
