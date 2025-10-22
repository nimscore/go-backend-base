# API Documentation

Документация REST API для Community сервиса с интерактивным Swagger UI.

## 🚀 Быстрый старт

### Запуск Gateway сервера

```bash
./community gateway
```

### Доступ к документации

- 🎨 **Swagger UI** (интерактивно): http://localhost:8090/swagger/
- 📄 **OpenAPI JSON**: http://localhost:8090/swagger/api.swagger.json
- 🔗 **Быстрый доступ**: http://localhost:8090/docs
- 💊 **Health check**: http://localhost:8090/health

---

## 📱 Swagger UI - Интерактивная документация

### Что такое Swagger UI?

Swagger UI — это интерактивный веб-интерфейс для тестирования API прямо из браузера.

**Возможности:**

- 📋 Список всех endpoints с описаниями
- 🔧 Интерактивное тестирование API без curl
- 📝 Полная документация параметров и ответов
- ✨ Автоматическая валидация запросов
- 🎯 Примеры запросов и ответов
- 🔐 Встроенная авторизация JWT

### Как использовать

1. Откройте http://localhost:8090/swagger/ в браузере
2. Выберите интересующий endpoint
3. Нажмите **"Try it out"**
4. Заполните параметры (если требуются)
5. Нажмите **"Execute"**
6. Просмотрите ответ сервера

### Пример: Создание сообщества через Swagger UI

1. Найдите `POST /communities` в списке
2. Нажмите **"Try it out"**
3. Заполните JSON в поле Request body:
   ```json
   {
   	"owner_id": "550e8400-e29b-41d4-a716-446655440000",
   	"slug": "golang",
   	"name": "Go Community",
   	"description": "A community for Go developers"
   }
   ```
4. Нажмите **"Execute"**
5. Увидите ответ с созданным ID в разделе Responses

### 🔐 Авторизация через Swagger UI

Для защищенных endpoints (требующих JWT токена):

1. Сначала выполните `POST /auth/login` для получения токена
2. Скопируйте значение `access_token` из ответа
3. Нажмите кнопку **"Authorize"** (замочек) вверху страницы
4. В поле введите: `Bearer <ваш_токен>`
5. Нажмите **"Authorize"** и закройте модальное окно
6. Теперь все запросы будут автоматически включать заголовок авторизации

---

## 📋 REST API Endpoints

### Authorization Service

| Метод | Endpoint         | Описание                        | Авторизация     |
| ----- | ---------------- | ------------------------------- | --------------- |
| POST  | `/auth/register` | Регистрация нового пользователя | ❌ Не требуется |
| POST  | `/auth/login`    | Вход пользователя               | ❌ Не требуется |
| POST  | `/auth/logout`   | Выход пользователя              | ✅ Требуется    |
| POST  | `/auth/refresh`  | Обновление access токена        | ✅ Требуется    |
| POST  | `/auth/validate` | Валидация токена                | ❌ Не требуется |

### Community Service

| Метод | Endpoint                            | Описание                     | Авторизация     |
| ----- | ----------------------------------- | ---------------------------- | --------------- |
| POST  | `/communities`                      | Создание сообщества          | ✅ Требуется    |
| GET   | `/communities/{id}`                 | Получение сообщества по UUID | ❌ Не требуется |
| GET   | `/communities/slug/{slug}`          | Получение сообщества по slug | ❌ Не требуется |
| GET   | `/communities?limit=40&cursor=<id>` | Список сообществ (пагинация) | ❌ Не требуется |

---

## 💻 Примеры curl запросов

### Authorization

#### Регистрация

```bash
curl -X POST http://localhost:8090/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "slug": "john_doe",
    "email": "john@example.com",
    "password": "securepass123"
  }'
```

#### Логин

```bash
curl -X POST http://localhost:8090/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "securepass123"
  }'
```

**Ответ:**

```json
{
	"user": {
		"id": "550e8400-e29b-41d4-a716-446655440000",
		"slug": "john_doe",
		"email": "john@example.com"
	},
	"access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
	"refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Обновление токена

```bash
curl -X POST http://localhost:8090/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

### Community

#### Создание сообщества

```bash
curl -X POST http://localhost:8090/communities \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "owner_id": "550e8400-e29b-41d4-a716-446655440000",
    "slug": "golang-community",
    "name": "Golang Community",
    "description": "Community for Go developers"
  }'
```

#### Получение сообщества по ID

```bash
curl http://localhost:8090/communities/550e8400-e29b-41d4-a716-446655440000
```

#### Получение сообщества по slug

```bash
curl http://localhost:8090/communities/slug/golang-community
```

#### Список сообществ (с пагинацией)

```bash
# Первая страница (40 элементов)
curl http://localhost:8090/communities?limit=40

# Следующая страница (используйте next_cursor из предыдущего ответа)
curl "http://localhost:8090/communities?limit=40&cursor=550e8400-e29b-41d4-a716-446655440000"
```

**Ответ:**

```json
{
  "communities": [...],
  "next_cursor": "450e8400-e29b-41d4-a716-446655440001",
  "has_more": true
}
```

---

## 🛠 Технические детали

### Используемые технологии

- **gRPC-Gateway**: автоматическая генерация REST API из gRPC
- **Swagger UI**: `github.com/swaggo/http-swagger/v2`
- **OpenAPI 2.0**: спецификация API

### Структура файлов

```
api/
├── swagger/
│   └── api.swagger.json    # Автоматически сгенерированная OpenAPI спецификация
└── README.md               # Эта документация
```

### Конфигурация Gateway

```go
import httpSwagger "github.com/swaggo/http-swagger/v2"

// Swagger UI
httpMux.HandleFunc("/swagger/", httpSwagger.Handler(
    httpSwagger.URL("/swagger/api.swagger.json"),
))
```

---

## 🔄 Обновление документации

### Автоматическая генерация

Документация генерируется автоматически из proto файлов при выполнении:

```bash
make generate-proto
```

Эта команда:

1. Генерирует Go код из `.proto` файлов
2. Создает gRPC-Gateway код
3. Генерирует OpenAPI/Swagger спецификацию в `api/swagger/api.swagger.json`

### После обновления proto файлов

1. Обновите HTTP аннотации в `.proto` файлах:

   ```protobuf
   rpc Create(CreateRequest) returns (CreateResponse) {
     option (google.api.http) = {
       post: "/communities"
       body: "*"
     };
   }
   ```

2. Запустите генерацию:

   ```bash
   make generate-proto
   ```

3. Перезапустите gateway сервер:

   ```bash
   ./community gateway
   ```

4. Обновите страницу в браузере — документация обновится автоматически!

---

## 📦 Альтернативные способы работы со Swagger

### Через Docker

```bash
# Запуск Swagger UI в Docker (альтернативный вариант)
docker run -p 8080:8080 \
  -e SWAGGER_JSON=/swagger/api.swagger.json \
  -v $(pwd)/api/swagger:/swagger \
  swaggerapi/swagger-ui

# Откройте http://localhost:8080
```

### Через VS Code

1. Установите расширение **"OpenAPI (Swagger) Editor"**
2. Откройте файл `api/swagger/api.swagger.json`
3. Редактор покажет визуализацию API

### Через онлайн редактор

1. Скопируйте содержимое `api/swagger/api.swagger.json`
2. Откройте https://editor.swagger.io/
3. Вставьте JSON в редактор

---

## ✨ Преимущества Swagger UI

✅ **Интерактивность** — тестируйте API прямо из браузера без curl  
✅ **Актуальность** — документация всегда соответствует коду  
✅ **Удобство** — красивый UI, не нужно писать команды вручную  
✅ **Валидация** — автоматическая проверка параметров  
✅ **Авторизация** — встроенная поддержка JWT токенов  
✅ **Профессионализм** — UI как у крупных компаний

---

## 🐛 Troubleshooting

### Swagger UI не открывается

```bash
# Проверьте, что gateway запущен
netstat -tlnp | grep 8090

# Проверьте логи
./community gateway
```

### Swagger JSON не загружается

```bash
# Убедитесь, что файл существует
ls -la api/swagger/api.swagger.json

# Регенерируйте, если нужно
make generate-proto
```

### CORS ошибки

Если используете Swagger UI из другого источника, убедитесь, что CORS настроен правильно в gateway сервере.

---

**📚 Больше информации:** см. [GATEWAY.md](../GATEWAY.md) для полной документации по HTTP Gateway.
