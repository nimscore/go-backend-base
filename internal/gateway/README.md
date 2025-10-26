# HTTP Gateway Module

HTTP/REST gateway сервер, который проксирует REST API запросы к gRPC сервисам используя [grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway).

## Обзор

Gateway модуль предоставляет:

- **REST API** - HTTP endpoints для всех gRPC сервисов
- **Swagger UI** - интерактивная документация API
- **Health checks** - проверка состояния сервера
- **Graceful shutdown** - корректное завершение работы

## Архитектура

```
Client (HTTP) → Gateway (HTTP:8090) → gRPC Server (gRPC:8080)
                   ↓
              Swagger UI
```

Gateway **не содержит бизнес-логики** - он только проксирует HTTP запросы к gRPC сервисам.

## Использование

### Запуск

```bash
# Из корня проекта
go run cmd/backend/main.go gateway

# Или через бинарник
./backend gateway
```

### Переменные окружения

```bash
# Gateway настройки
GATEWAY_HOST=0.0.0.0    # default: 0.0.0.0
GATEWAY_PORT=8090        # default: 8090

# gRPC backend настройки
GRPC_HOST=127.0.0.1      # default: 127.0.0.1
GRPC_PORT=8080           # default: 8080

# Debug mode
DEBUG=1                  # включить development logger
```

### Endpoints

После запуска доступны:

- **REST API**: `http://localhost:8090/`
- **Swagger UI (Docs)**: `http://localhost:8090/docs/`
- **Swagger JSON**: `http://localhost:8090/api/swagger.json`
- **Health Check**: `http://localhost:8090/health`

## Зарегистрированные сервисы

Текущие (2 из 15):

- ✅ AuthorizationService - `/auth/*`
- ✅ CommunityService - `/communities/*`

Требуют регистрации (13):

- ⏳ UserService
- ⏳ PostService
- ⏳ CommentService
- ⏳ FeedService
- ⏳ RoleService
- ⏳ PermissionService
- ⏳ PlatformService
- ⏳ ModerationService
- ⏳ ReportService
- ⏳ MediaService
- ⏳ NotificationService
- ⏳ SearchService
- ⏳ BadgeService

## Добавление нового сервиса

### 1. Убедитесь что proto файл содержит HTTP аннотации

```protobuf
service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse) {
    option (google.api.http) = {
      get: "/users/{user_id}"
    };
  }
}
```

### 2. Сгенерируйте gateway код

```bash
# Добавьте proto файл в Makefile:
make generate-proto
```

Это создаст `internal/proto/user.pb.gw.go`

### 3. Зарегистрируйте сервис в gateway

Отредактируйте `internal/gateway/gateway.go`:

```go
func (this *Gateway) registerServices(ctx context.Context, mux *runtime.ServeMux, opts []grpc.DialOption) error {
    // ... existing services ...

    // User Service
    err = protopkg.RegisterUserServiceHandlerFromEndpoint(ctx, mux, this.grpcEndpoint, opts)
    if err != nil {
        return fmt.Errorf("register user service: %w", err)
    }
    this.logger.Info("Registered UserService")

    return nil
}
```

### 4. Перезапустите gateway

```bash
go run cmd/backend/main.go gateway
```

## Структура кода

```
internal/gateway/
├── gateway.go          # Gateway тип и логика
└── README.md           # Документация

Методы Gateway:
├── NewGateway()        # Конструктор
├── Start()             # Запуск сервера
├── Stop()              # Остановка сервера
├── registerServices()  # Регистрация gRPC сервисов
└── setupRoutes()       # Настройка HTTP маршрутов
```

## Swagger UI

После запуска gateway перейдите на `http://localhost:8090/docs/` для интерактивной документации.

Swagger UI позволяет:

- Просматривать все доступные endpoints
- Тестировать API запросы прямо из браузера
- Видеть схемы request/response
- Копировать curl команды

## Health Check

```bash
curl http://localhost:8090/health
# {"status":"healthy"}
```

## Graceful Shutdown

Gateway поддерживает корректное завершение работы:

```bash
# SIGINT (Ctrl+C) или SIGTERM
kill -TERM <pid>
```

При получении сигнала:

1. Прекращает принимать новые запросы
2. Завершает обработку текущих запросов
3. Закрывает HTTP сервер
4. Логирует завершение

## Отличия от gRPC модуля

| Аспект        | gRPC Module        | Gateway Module                     |
| ------------- | ------------------ | ---------------------------------- |
| Протокол      | gRPC               | HTTP/REST                          |
| Порт          | 8080               | 8090                               |
| Клиенты       | gRPC клиенты       | Браузеры, curl, любые HTTP клиенты |
| Документация  | Proto файлы        | Swagger UI                         |
| Бизнес-логика | Да                 | Нет (только proxy)                 |
| Зависимость   | База данных, Kafka | gRPC сервер                        |

## Производительность

Gateway добавляет небольшой overhead:

- **Преобразование** HTTP → gRPC
- **Сериализация** JSON ↔ Protobuf
- **Сетевой hop** (если gRPC на другом хосте)

Для высоконагруженных сценариев рекомендуется использовать прямые gRPC клиенты.

### 404 на API endpoints

**Причина**: Сервис не зарегистрирован в gateway

**Решение**: Добавьте регистрацию сервиса в `registerServices()`

### Swagger JSON не загружается

**Причина**: Файл `api/swagger/api.swagger.json` не существует

**Решение**: Сгенерируйте Swagger документацию

```bash
make generate-proto
```

## Development

### Запуск в debug режиме

```bash
DEBUG=1 go run cmd/backend/main.go gateway
```

### Просмотр логов

Gateway логирует все важные события:

- Старт сервера
- Регистрация сервисов
- Настройка маршрутов
- Ошибки
- Shutdown

### Тестирование

```bash
# Unit тесты
go test ./internal/gateway/...

# Integration тесты
go test -tags=integration ./internal/gateway/...
```

## Best Practices

1. **Всегда запускайте gRPC сервер перед gateway**
2. **Используйте Swagger UI для тестирования API**
3. **Проверяйте health endpoint в мониторинге**
4. **Для production используйте reverse proxy (nginx) перед gateway**
5. **Настройте CORS если gateway используется из браузера**

## Связанные файлы

- `cmd/backend/cmd_gateway.go` - CLI команда запуска
- `proto/*.proto` - gRPC service definitions с HTTP аннотациями
- `internal/proto/*.pb.gw.go` - сгенерированный gateway код
- `api/swagger/api.swagger.json` - Swagger спецификация
- `Makefile` - команда `generate-proto`

## Ссылки

- [grpc-gateway documentation](https://grpc-ecosystem.github.io/grpc-gateway/)
- [gRPC HTTP mapping](https://cloud.google.com/endpoints/docs/grpc/transcoding)
- [Swagger UI](https://swagger.io/tools/swagger-ui/)
