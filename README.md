![Покрытие тестами](.badges/coverage.svg)

# sys-metrics

Сервис для сбора и хранения метрик с машины, на которой запущен агент. Агент периодически снимает показатели (память, CPU, счётчики опроса и т.д.) и отправляет их на сервер. Сервер сохраняет метрики и отдаёт их через HTTP или gRPC.

Проект из курса «Go-разработчик» (Яндекс Практикум), трек «Сервер сбора метрик и алертинга».

## Как это устроено

```
┌─────────┐     HTTP или gRPC      ┌──────────────────┐
│  agent  │ ──────────────────────► │  server /        │
│         │                         │  server_grpc     │
└─────────┘                         └────────┬─────────┘
     │                                       │
     │ опрашивает                            │ пишет в
     │ runtime.Metrics                       ▼
     ▼                              память / файл / PostgreSQL
  локальная машина
```

- **Агент** — фоновый процесс на машине-источнике. Раз в несколько секунд собирает метрики и раз в N секунд отправляет пачку на сервер.
- **Сервер** (`cmd/server`) — один процесс: HTTP (REST API) и gRPC (protobuf) на разных портах.
- **Хранилище** — по умолчанию в памяти; опционально файл-бэкап или PostgreSQL (через `DATABASE_DSN`).

## Быстрый старт

### Сборка

```bash
make build
```

Собираются `cmd/server/server` и `cmd/agent/agent`.

### Запуск (самый простой сценарий)

Терминал 1 — сервер (HTTP + gRPC):

```bash
./cmd/server/server -a localhost:8080 -a-grpc localhost:9090 -m development
```

Терминал 2 — агент (HTTP):

```bash
./cmd/agent/agent -a localhost:8080 -m development
```

Или агент с gRPC (указываешь адрес gRPC-порта):

```bash
./cmd/agent/agent -a localhost:9090 -report-transport grpc -m development
```

По умолчанию, если `-a-grpc` не задан: gRPC слушает тот же host, порт `9090`.

## Настройка

Конфигурация читается в порядке приоритета: **флаги** → **переменные окружения** → **JSON-файл** (`-c path/to/config.json`).

### Сервер (HTTP и gRPC)

| Флаг | Переменная | По умолчанию | Зачем |
|------|------------|--------------|-------|
| `-a` | `ADDRESS` | `localhost:8080` | HTTP: `host:port` |
| `-a-grpc` | `ADDRESS_GRPC` | `localhost:9090` | gRPC: `host:port` (host как у HTTP, если не задан) |
| `-m` | `MODE` | `production` | `development` — подробные логи |
| `-i` | `STORE_INTERVAL` | `300` | Интервал сброса на диск (сек) |
| `-f` | `STORE_FILE` | `./backups` | Путь к файлу бэкапа |
| `-d` | `DATABASE_DSN` | — | PostgreSQL вместо файла |
| `-r` | `RESTORE` | `true` | Восстановить метрики при старте |
| `-k` | `KEY` | — | Ключ для подписи запросов (SHA-256) |
| `-t` | `TRUSTED_SUBNET` | — | Доверенная подсеть для gRPC (CIDR) |
| `-crypto-key` | `CRYPTO_KEY` | — | PEM-ключ для расшифровки тела запроса |

### Агент

| Флаг | Переменная | По умолчанию | Зачем |
|------|------------|--------------|-------|
| `-a` | `ADDRESS` | `localhost:8080` | Куда слать метрики |
| `-p` | `POLL_INTERVAL` | `2` | Как часто опрашивать систему (сек) |
| `-r` | `REPORT_INTERVAL` | `10` | Как часто отправлять на сервер (сек) |
| `-report-transport` | `REPORT_TRANSPORT` | `http` | `http` или `grpc` |
| `-m` | `MODE` | `production` | Режим логирования |
| `-l` | `RATE_LIMIT` | `1` | Параллельность отправки |
| `-k` | `KEY` | — | Ключ подписи (если включён на сервере) |
| `-crypto-key` | `CRYPTO_KEY` | — | Публичный ключ для шифрования тела |

Пример JSON-конфига агента:

```json
{
  "address": "localhost:9090",
  "poll_interval": "2s",
  "report_interval": "10s",
  "report_transport": "grpc"
}
```

## Что где лежит в проекте

```
sys-metrics/
├── cmd/                    # Точки входа — исполняемые программы
│   ├── agent/              # Агент сбора метрик
│   ├── server/             # HTTP + gRPC в одном процессе
│   ├── server_grpc/        # устарело, используй cmd/server
│   ├── keygenerator/       # Утилита генерации RSA-ключей
│   └── ...
│
├── internal/               # Основная логика (не импортируется снаружи)
│   ├── agent/              # Сбор метрик, отправка на сервер
│   ├── agent/sender/       # HTTP- и gRPC-отправители
│   ├── app/                # Общий bootstrap сервера: БД, бэкап, сервисы
│   ├── handler/            # HTTP-обработчики (/update, /value, /ping, …)
│   ├── grpchandler/        # gRPC UpdateMetrics
│   ├── Interceptors/       # gRPC: x-real-ip, проверка trusted subnet
│   ├── service/metrics/    # Бизнес-логика обновления метрик
│   ├── repository/         # Хранение: memory, file, postgres
│   ├── config/             # Парсинг флагов, env, JSON
│   ├── proto/              # Сгенерированный код из protobuf
│   └── ...
│
├── api/                    # Исходные .proto-файлы
├── migrations/             # SQL-миграции для PostgreSQL
├── pkg/                    # Переиспользуемые утилиты (сжатие HTTP, пулы)
└── Makefile                # Сборка, тесты, автотесты курса
```

Если нужно разобраться в коде — начни с `cmd/*/main.go`, дальше `internal/app` (сервер) или `internal/agent` (агент).

## HTTP API

| Метод | Путь | Назначение |
|-------|------|------------|
| POST | `/update/{type}/{name}/{value}` | Одна метрика |
| POST | `/update/` | Одна метрика (JSON) |
| POST | `/updates/` | Пачка метрик (JSON) |
| GET | `/value/{type}/{name}` | Значение метрики |
| GET | `/` | HTML со списком всех метрик |
| GET | `/ping` | Проверка живости (и БД, если настроена) |

Типы метрик: `gauge` (float) и `counter` (int).

## gRPC

Сервис `Metrics`, метод `UpdateMetrics` — принимает массив метрик в protobuf. Агент при `-report-transport grpc` шлёт батч через этот метод. В metadata передаётся `x-real-ip`; сервер может отклонить запрос, если IP не из доверенной подсети (`-t` / `TRUSTED_SUBNET`).

## Зависимости

- Go 1.26+
- PostgreSQL — только если используешь `-d` / `DATABASE_DSN`
- Docker — для integration-тестов и автотестов с БД
