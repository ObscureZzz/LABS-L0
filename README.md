# Demo Order Service

Проект `demo-order-service` — это сервис для приёма заказов, их хранения в кэше и базе данных PostgreSQL, а также публикации/подписки через NATS Streaming.

---

## 📦 Структура проекта

```
demo-order-service/
├─ service/             # Основной Go-сервис
├─ publisher/           # Go-публикатор заказов
├─ model.json           # Пример JSON заказа
├─ frontend/            # Интерфейс для поиска заказов
├─ docker-compose.yml   # Сборка всех контейнеров
└─ README.md            # Инструкция
```

---

## ⚡ Запуск проекта через Docker

### 1. Запуск всех контейнеров

```bash
docker-compose up -d
```

Проверяем, что контейнеры поднялись:

```bash
docker ps
```

Пример вывода:

```
CONTAINER ID   IMAGE                          PORTS
frontend       nginx:latest                   0.0.0.0:9090->80/tcp
publisher      demo-order-service-publisher   0.0.0.0:4000->4000/tcp
service        demo-order-service-service     0.0.0.0:3000->3000/tcp
postgres       postgres:15                    0.0.0.0:5432->5432/tcp
nats           nats-streaming:latest          0.0.0.0:4222->4222/tcp
```

### 2. Остановка контейнеров

```bash
docker-compose down
```

Или по отдельности:

```bash
docker stop service publisher frontend nats postgres
```

### 3. Перезапуск контейнеров

```bash
docker-compose restart
```

Или конкретного:

```bash
docker restart service
```

---

## 📝 Работа с сервисом

### 1. Публикация заказов через HTTP POST

Сервис `publisher` слушает порт **4000**:

```bash
POST http://localhost:4000/publish
Content-Type: application/json

{
  "order_uid": "b563feb7b2b84b6test",
  "track_number": "WBILMTESTTRACK",
  "entry": "WEB",
  "delivery": { ... },
  "payment": { ... },
  "items": [ ... ]
}
```

Можно использовать JSON-файл `model.json` с curl:

```bash
curl -X POST http://localhost:4000/publish -H "Content-Type: application/json" -d @model.json
```

### 2. Просмотр кэша

Сервис `service` слушает порт **3000**.

- Посмотреть конкретный заказ по `order_uid`:

```
curl http://localhost:3000/orders/<order_uid>
```

- Посмотреть весь кэш:

```
curl http://localhost:3000/cache
```

### 3. Просмотр логов контейнера

```bash
docker logs service --tail 200
```

### 4. Работа с базой данных PostgreSQL

#### Войти в контейнер PostgreSQL

```bash
docker exec -it postgres psql -U demo_user -d demo_db
```

#### Просмотр таблиц и схем

```sql
\dt   -- таблицы
\dn   -- схемы
```

#### Просмотр данных

```sql
SELECT * FROM orders;
SELECT * FROM delivery;
SELECT * FROM payment;
SELECT * FROM items;
```

#### Очистка таблиц

```sql
TRUNCATE TABLE items, payment, delivery, orders RESTART IDENTITY;
```

### 5. Очищение кэша сервиса

Для полной очистки кэша можно перезапустить сервис:

```bash
docker restart service
```

Кэш будет загружен заново из базы данных.

---


## 🌐 Frontend

- Открыть `http://localhost:9090`.
- Вводим `order_uid` для поиска заказа.

