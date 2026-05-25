#!/usr/bin/env bash
set -euo pipefail

RATE="${RATE:-100}"        # messages per second
DURATION="${DURATION:-60}" # seconds
TOPIC="${TOPIC:-search-requests}"

QUERIES=(
  "Ноутбук для учебы"
  "Купить ноутбук"
  "Летняя обувь"
  "Летняя одежда"
  "Гантели"
  "Игровой ноутбук"
  "Мышь для ноутбука"
  "Клавиатура для ноутбука"
  "Компьютерные аксессуары"
  "Смарт-часы для фитнеса"
)

interval=$(awk -v rate="$RATE" 'BEGIN { if (rate<=0) rate=1; printf "%.6f", 1/rate }')
end=$((SECONDS + DURATION))

producer_cmd=(docker compose exec -T kafka kafka-console-producer --bootstrap-server kafka:9092 --topic "$TOPIC")

while [ $SECONDS -lt $end ]; do
  q=${QUERIES[$RANDOM % ${#QUERIES[@]}]}
  printf '{"request":"%s"}\n' "$q"
  sleep "$interval"
done | "${producer_cmd[@]}"
