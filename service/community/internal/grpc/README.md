# gRPC-Gateway Integration

Проект интегрирован с **grpc-gateway** для автоматической генерации REST API и Swagger документации из gRPC proto файлов.

## 🚀 Быстрый старт

### 1. Запуск gRPC сервера

```bash
# Установите переменные окружения
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export KAFKA_HOST=localhost
export KAFKA_PORT=9092

# Запустите gRPC сервер
./community server
# Сервер запустится на порту 8080
```

### 2. Запуск HTTP Gateway

```bash
# В другом терминале
./community gateway
# Gateway запустится на порту 8090
```

### 3. Доступ к документации

🎨 Swagger UI: http://localhost:8090/swagger/
📚 Документация: http://localhost:8090/docs
💊 Health: http://localhost:8090/health

Swagger UI предоставляет интерактивный интерфейс для тестирования всех endpoints прямо из браузера!

Вы увидите красивую страницу с:

- 📄 Ссылкой на Swagger спецификацию
- 📋 Списком всех доступных endpoints
- 💊 Health check endpoint

## 📚 Документация API

### Swagger/OpenAPI

```bash
# Скачать Swagger спецификацию
curl http://localhost:8090/swagger/ > api.swagger.json

# Или просмотреть в Swagger UI
docker run -p 8081:8080 \
  -e SWAGGER_JSON=/swagger/api.swagger.json \
  -v $(pwd)/api/swagger:/swagger \
  swaggerapi/swagger-ui

# Откройте http://localhost:8081
```

## 🔧 REST API Endpoints

### Authorization Service

```bash
# Регистрация
curl -X POST http://localhost:8090/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "slug": "john_doe",
    "email": "john@example.com",
    "password": "securepass123"
  }'

# Логин
curl -X POST http://localhost:8090/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "securepass123"
  }'

# Обновление токена
curl -X POST http://localhost:8090/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "your-refresh-token"
  }'
```

### Community Service

```bash
# Создание сообщества
curl -X POST http://localhost:8090/communities \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "owner_id": "550e8400-e29b-41d4-a716-446655440000",
    "slug": "golang-community",
    "name": "Golang Community",
    "description": "A community for Go developers"
  }'

# Получение сообщества по ID
curl http://localhost:8090/communities/550e8400-e29b-41d4-a716-446655440000

# Получение сообщества по slug
curl http://localhost:8090/communities/slug/golang-community

# Список сообществ (первая страница, 40 элементов)
curl http://localhost:8090/communities?limit=40

# Следующая страница
curl "http://localhost:8090/communities?limit=40&cursor=550e8400-e29b-41d4-a716-446655440000"
```

## 🛠 Разработка

### Добавление нового endpoint

1. **Обновите proto файл** с HTTP аннотациями:

```protobuf
service CommunityService {
  rpc UpdateCommunity(UpdateCommunityRequest) returns (UpdateCommunityResponse) {
    option (google.api.http) = {
      put: "/communities/{id}"
      body: "*"
    };
  }
}
```

2. **Регенерируйте код:**

```bash
make generate-proto
```

3. **Перезапустите сервисы:**

```bash
# Перезапустите gRPC сервер
./community server

# Перезапустите gateway (в другом терминале)
./community gateway
```

### Структура файлов

```
service/community/
├── proto/                          # Proto определения
│   ├── authorization.proto         # + HTTP аннотации
│   └── community.proto             # + HTTP аннотации
├── internal/proto/                 # Сгенерированный код
│   ├── *.pb.go                     # Protobuf messages
│   ├── *_grpc.pb.go                # gRPC service
│   └── *.pb.gw.go                  # Gateway reverse proxy (НОВОЕ!)
├── api/swagger/                    # Swagger документация (НОВОЕ!)
│   └── api.swagger.json            # OpenAPI спецификация
└── cmd/community/
    ├── cmd_server.go               # gRPC сервер
    └── cmd_gateway.go              # HTTP gateway (НОВОЕ!)
```

## 🌐 Переменные окружения

### Gateway

```bash
GATEWAY_HOST=0.0.0.0      # Адрес HTTP gateway (по умолчанию 0.0.0.0)
GATEWAY_PORT=8090         # Порт HTTP gateway (по умолчанию 8090)
GRPC_HOST=127.0.0.1       # Адрес gRPC сервера
GRPC_PORT=8080            # Порт gRPC сервера
```

### gRPC Server

```bash
GRPC_HOST=127.0.0.1       # Адрес gRPC сервера
GRPC_PORT=8080            # Порт gRPC сервера
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
KAFKA_HOST=localhost
KAFKA_PORT=9092
DEBUG=1                   # Development mode
```

## 🐳 Docker

### Dockerfile обновлений не требуется

Существующий `Dockerfile.microservice` работает как для gRPC, так и для gateway:

```bash
# Сборка образа
docker build -f Dockerfile.microservice -t community:latest .

# Запуск gRPC сервера
docker run -p 8080:8080 community:latest server

# Запуск HTTP gateway
docker run -p 8090:8090 \
  -e GRPC_HOST=host.docker.internal \
  community:latest gateway
```

## 📊 Мониторинг

### Health checks

```bash
# Gateway health
curl http://localhost:8090/health

# Возвращает:
# {"status":"healthy"}
```

### Метрики

Gateway автоматически проксирует метрики с gRPC сервера. Prometheus метрики доступны через gRPC.

## 🔍 Отладка

### Логирование

```bash
# Development mode с детальными логами
DEBUG=1 ./community gateway
```

### Проверка gRPC→HTTP маппинга

```bash
# Используйте curl с verbose
curl -v http://localhost:8090/communities

# Проверьте заголовки
curl -I http://localhost:8090/health
```

## 🎯 Best Practices

1. **Версионирование API:** Все endpoints начинаются с `/`
2. **Cursor пагинация:** Используйте `limit` и `cursor` параметры
3. **HTTP методы:**
   - `GET` для чтения
   - `POST` для создания
   - `PUT` для обновления
   - `DELETE` для удаления
4. **Коды ответов:**
   - `200 OK` - успех
   - `400 Bad Request` - неверные данные
   - `401 Unauthorized` - не авторизован
   - `404 Not Found` - не найдено
   - `500 Internal Server Error` - внутренняя ошибка

## 📦 Зависимости

```bash
# Основные
github.com/grpc-ecosystem/grpc-gateway/v2  # Gateway runtime
google.golang.org/genproto/googleapis/api  # HTTP аннотации

# Инструменты (только для разработки)
protoc-gen-grpc-gateway                     # Генератор gateway кода
protoc-gen-openapiv2                        # Генератор Swagger
```

## 🚨 Troubleshooting

### Gateway не может подключиться к gRPC

```bash
# Проверьте, что gRPC сервер запущен
netstat -tlnp | grep 8080

# Проверьте переменные окружения
echo $GRPC_HOST
echo $GRPC_PORT
```

### Swagger не генерируется

```bash
# Проверьте установку плагинов
which protoc-gen-openapiv2

# Переустановите
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
```

### 404 на endpoints

```bash
# Проверьте, что routes зарегистрированы
curl http://localhost:8090/

# Должна открыться страница с документацией
```

## 📚 Дополнительные ресурсы

- [grpc-gateway Documentation](https://grpc-ecosystem.github.io/grpc-gateway/)
- [Google API HTTP Annotations](https://github.com/googleapis/googleapis/blob/master/google/api/http.proto)
- [OpenAPI Specification](https://swagger.io/specification/)
