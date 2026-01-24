# WB Project

Проект — это сервис обработки заказов с использованием **Kafka**, **PostgreSQL**, кэширования и **HTTP API**.

---

## 📁 Структура проекта

```
wb-project/
├── cmd/app/                # main.go, точка входа
├── internal/
│   ├── app/                # HTTP сервер
│   ├── cache/              # Кэширование заказов
│   ├── config/             # Конфигурация
│   ├── db/
│   │   ├── conn/           # Подключение к БД
│   │   └── repository/     # Репозитории для работы с таблицами
│   ├── handler/            # HTTP Handlers
│   ├── kafka/              # Producer и Consumer Kafka
│   ├── metric/             # Метрики Prometheus
│   ├── models/             # Модели заказов и связанных структур
│   └── service/            # Бизнес-логика
├── testdata/               # JSON-примеры заказов для тестов
└── go.mod
```

---

## ⚙️ Установка и запуск

1. Клонируем проект:

```bash
git clone <repo_url>
cd wb-project
```

2. Устанавливаем зависимости:

```bash
go mod tidy
```
3. Миграция бд:

```bash
make migrate-up
```
4. Запуск сервиса:

```bash
go run cmd/app/
```

* HTTP сервер на `:8080`
* Kafka Producer и Consumer
* Подгрузка кэша из БД

---

## 🛠 Endpoints

### GET /orders/{order_uid}

Возвращает заказ по UID из кэша или БД.

**Пример запроса:**

```bash
curl http://localhost:8080/orders/123e4567-e89b-12d3-a456-426614174000
```

**Пример ответа:**

```json
{
  "order_uid": "123e4567-e89b-12d3-a456-426614174000",
  "track_number": "WB123456789",
  "entry": "WB-Entry",
  "delivery": {
    "name": "Иван Иванов",
    "phone": "+71234567890",
    "zip": "123456",
    "city": "Москва",
    "address": "ул. Ленина, 1",
    "region": "Москва",
    "email": "ivan@example.com"
  },
  "payment": {
    "transaction": "TX12345",
    "request_id": "REQ98765",
    "currency": "RUB",
    "provider": "SBER",
    "amount": 5000,
    "payment_dt": 1670000000,
    "bank": "Sberbank",
    "delivery_cost": 300,
    "goods_total": 4700,
    "custom_fee": 0
  },
  "items": [
    {
      "chrt_id": 123,
      "track_number": "WB123456789",
      "price": 4700,
      "rid": "RID123",
      "name": "Товар 1",
      "sale": 0,
      "size": "M",
      "total_price": 4700,
      "nm_id": 1,
      "brand": "Brand1",
      "status": 1
    }
  ],
  "locale": "ru",
  "customer_id": "customer_1",
  "delivery_service": "dpd",
  "shard_key": "shard_1",
  "sm_id": 1,
  "date_created": "2026-01-24T10:00:00Z",
  "oof_shard": "oof_1"
}
```

---

## 📊 Метрики Prometheus

* **Kafka**: `order_kafka_messages_received_total{status="success|error"}`
* **Database**:

  * `order_db_operations_total{operation="save|get", status="success|error"}`
  * `order_db_operation_duration_seconds{operation="save|get"}`
* **Cache**:

  * `order_cache_items_count` — текущее количество заказов в кэше
  * `order_cache_cof_items_count{result="hit|miss"}` — попадания/промахи
* **HTTP Requests**:

  * `order_http_request{status="200|404|500"}`

Пример записи метрики:

```
order_http_request{status="200"} 1
order_cache_cof_items_count{result="hit"} 2
```

---

## 🧪 Тесты

* Расположены в `internal/service/service_test.go`
* Используется **testify** и моки
* Проверяется:

  * Парсинг и обработка сообщений Kafka
  * Сохранение заказов в БД
  * Валидация заказов
  * Работа кэша
  * Метод `ReCache`

### Запуск тестов

```bash
go test ./internal/service -v -cover
```

**Покрытие:**

```
ok      wb-project/internal/service     coverage: 97.4% of statements
```

---

## 🧩 Архитектура

Сервис использует **чистую архитектуру**:

* **Handler** — HTTP endpoints (только GET)
* **Service** — бизнес-логика, работа с кэшем и БД
* **Repository** — интерфейс к БД
* **Cache** — временное хранение заказов
* **Kafka** — интеграция с брокером
* **Metrics** — Prometheus
* **Application** — оркестрация всех компонентов, graceful shutdown


---

## ⚡ Примеры запуска

```bash
# Сервис
go run cmd/app/

# Тесты
go test ./internal/service -v -cover

