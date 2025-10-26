# Система уведомлений

## Обзор

NotificationService предоставляет real-time уведомления о событиях на платформе через gRPC Server-Side Streaming с персонализированными настройками и историей.

## gRPC Service: NotificationService

Proto файл: `proto/notification.proto`

## Сущности

### NotificationType

```protobuf
enum NotificationType {
  NOTIFICATION_TYPE_UNSPECIFIED                 = 0
  NOTIFICATION_TYPE_NEW_POST_IN_COMMUNITY       = 1
  NOTIFICATION_TYPE_NEW_POST_FROM_FOLLOWED_USER = 2
  NOTIFICATION_TYPE_COMMENT_LIKED               = 3
  NOTIFICATION_TYPE_NEW_COMMENT_IN_POST         = 4
  NOTIFICATION_TYPE_COMMENT_REPLY               = 5
}
```

### NotificationContent

```protobuf
message NotificationContent {
  string actor_username
  string actor_avatar
  string target_type          // post или comment
  string target_id
  string action_type

  // Type-specific fields
  optional string community_name
  optional string post_title
  optional string post_id
  optional string comment_id
  optional string comment_text      // truncated 100 chars
  optional string original_comment_id
  optional string reply_text

  google.protobuf.Timestamp timestamp
}
```

### Notification

```protobuf
message Notification {
  string id
  string user_id
  NotificationType type
  NotificationContent content
  bool is_read
  google.protobuf.Timestamp created_at
}
```

### NotificationPreferences

```protobuf
message NotificationPreferences {
  bool new_post_in_community
  bool new_post_from_followed_user
  bool comment_liked
  bool new_comment_in_post
  bool comment_reply
}
```

## Real-time Streaming

### Stream

**RPC:** `Stream(StreamRequest) returns (stream Notification)`  
**gRPC Streaming:** Server-Side Streaming  
**FR:** FR-400, FR-408, FR-410, FR-430-432

Real-time поток уведомлений для текущего пользователя.

**Request:**

```protobuf
message StreamRequest {}
```

**Response Stream:**

```protobuf
stream Notification
```

**Требования:**

- Требуется аутентификация
- Доставка уведомлений в реальном времени (FR-408)
- Хранение в БД даже если пользователь offline (FR-409)
- Доставка пропущенных при переподключении (FR-410)
- Подключение активно до disconnect клиента (FR-430)
- Graceful reconnection с дедупликацией (FR-431)
- Максимум 10 одновременных подключений на пользователя (FR-432)
- Доставка в течение 500ms от события (SC-012)

**События:**
Клиент получает stream уведомлений по мере их создания.

---

## История уведомлений

### Get

**RPC:** `Get(GetRequest) returns (GetResponse)`  
**HTTP:** `GET /notifications`  
**FR:** FR-411, FR-421-423

Получение истории уведомлений.

**Request:**

```protobuf
message GetRequest {
  optional bool read_status_filter       // фильтр по прочитанности
  optional NotificationType type_filter
  string cursor
  int32 limit
}
```

**Response:**

```protobuf
message GetResponse {
  repeated Notification notifications
  string next_cursor
  bool has_more
}
```

**Требования:**

- Cursor-based пагинация
- Сортировка по created_at в обратном порядке (новые первые) (FR-421)
- Фильтрация по read status: all, unread, read (FR-422)
- Фильтрация по типу уведомления (FR-423)

---

### MarkAsRead

**RPC:** `MarkAsRead(MarkAsReadRequest) returns (MarkAsReadResponse)`  
**HTTP:** `POST /notifications/{notification_id}/read`  
**FR:** FR-412, FR-424

Пометка уведомления как прочитанного.

**Request:**

```protobuf
message MarkAsReadRequest {
  string notification_id
}
```

**Требования:**

- Установка is_read = true
- Обновление unread count

---

### MarkAllAsRead

**RPC:** `MarkAllAsRead(MarkAllAsReadRequest) returns (MarkAllAsReadResponse)`  
**HTTP:** `POST /notifications/read-all`  
**FR:** FR-413, FR-425

Пометка всех уведомлений как прочитанных.

**Request:**

```protobuf
message MarkAllAsReadRequest {}
```

**Response:**

```protobuf
message MarkAllAsReadResponse {
  string message
  int32 marked_count
}
```

**Требования:**

- Пометка всех уведомлений пользователя
- Возврат количества помеченных
- Unread count становится 0

---

### GetUnreadCount

**RPC:** `GetUnreadCount(GetUnreadCountRequest) returns (GetUnreadCountResponse)`  
**HTTP:** `GET /notifications/unread-count`  
**FR:** FR-414, FR-426

Получение количества непрочитанных уведомлений.

**Request:**

```protobuf
message GetUnreadCountRequest {}
```

**Response:**

```protobuf
message GetUnreadCountResponse {
  int32 unread_count
}
```

**Требования:**

- Возврат количества уведомлений с is_read = false

---

## Настройки уведомлений

### GetPreferences

**RPC:** `GetPreferences(GetPreferencesRequest) returns (GetPreferencesResponse)`  
**HTTP:** `GET /notifications/preferences`  
**FR:** FR-415, FR-428

Получение настроек уведомлений пользователя.

**Request:**

```protobuf
message GetPreferencesRequest {}
```

**Response:**

```protobuf
message GetPreferencesResponse {
  NotificationPreferences preferences
}
```

---

### UpdatePreferences

**RPC:** `UpdatePreferences(UpdatePreferencesRequest) returns (UpdatePreferencesResponse)`  
**HTTP:** `PUT /notifications/preferences`  
**FR:** FR-416, FR-427-428

Обновление настроек уведомлений.

**Request:**

```protobuf
message UpdatePreferencesRequest {
  NotificationPreferences preferences
}
```

**Response:**

```protobuf
message UpdatePreferencesResponse {
  NotificationPreferences preferences
}
```

**Требования:**

- Пользователь может enable/disable каждый тип независимо (FR-427)
- По умолчанию все типы enabled для новых пользователей (FR-428)

---

## Типы уведомлений

### NEW_POST_IN_COMMUNITY

**Триггер:** Новый пост опубликован в сообществе где пользователь состоит (FR-403)

**Content (FR-434):**

- community_name
- post_title
- post_id
- author info (actor_username, actor_avatar)

**Пример:**
"[Author] опубликовал пост '[Title]' в [Community]"

---

### NEW_POST_FROM_FOLLOWED_USER

**Триггер:** Подписанный пользователь опубликовал новый пост (FR-404)

**Content (FR-435):**

- author info (actor_username, actor_avatar)
- post_title
- post_id
- community_name

**Пример:**
"[Author], на которого вы подписаны, опубликовал '[Title]' в [Community]"

---

### COMMENT_LIKED

**Триггер:** Чей-то лайк на комментарий пользователя (FR-405)

**Content (FR-436):**

- liker info (actor_username, actor_avatar)
- comment_id
- comment_text (обрезан до 100 символов)

**Пример:**
"[User] лайкнул ваш комментарий: '[text...]'"

---

### NEW_COMMENT_IN_POST

**Триггер:** Новый комментарий на пост пользователя (FR-406)

**Content (FR-437):**

- commenter info (actor_username, actor_avatar)
- post_title
- post_id
- comment_text (обрезан)

**Пример:**
"[User] прокомментировал ваш пост '[Title]': '[text...]'"

---

### COMMENT_REPLY

**Триггер:** Ответ на комментарий пользователя (FR-407)

**Content (FR-438):**

- replier info (actor_username, actor_avatar)
- original_comment_id
- reply_text (обрезан)
- post context (post_id, post_title)

**Пример:**
"[User] ответил на ваш комментарий: '[text...]'"

---

## Создание уведомлений

### Правила создания

- Уважать preferences пользователя (FR-419)
- НЕ создавать если тип отключен в настройках
- НЕ создавать если пользователь взаимодействует с собственным контентом (FR-420)
- Хранить в БД независимо от online статуса (FR-409)

### Батчинг (FR-439)

- Уведомления батчатся для эффективности
- Максимальная задержка batching: 100ms
- Доставка в реальном времени после batch

### Идемпотентность (FR-440)

- Уникальные constraints для предотвращения дубликатов
- Одно событие = одно уведомление

---

## Жизненный цикл

### 1. Создание

- Событие происходит (новый пост, лайк, etc.)
- Backend проверяет preferences
- Создает уведомление в БД
- is_read = false

### 2. Доставка

- Если пользователь online: push через stream
- Если offline: ожидает в БД

### 3. Прочтение

- Клиент вызывает MarkAsRead
- is_read = true
- unread_count уменьшается

### 4. Удаление

- Автоматически через 90 дней (FR-429, FR-057)
- Background job clean up

---

## Real-time доставка

### Производительность

- Целевая задержка: 500ms (SC-012)
- Поддержка тысяч одновременных streams
- Эффективный pub/sub механизм

### Reconnection

- Graceful handling disconnects (FR-431)
- Доставка missed notifications при reconnect (FR-410)
- Дедупликация для избежания дублей

### Лимиты

- Максимум 10 одновременных подключений (FR-432)
- Защита от resource exhaustion
- Старые connections могут force close

---

## Персонализация

### Preferences

Каждый пользователь может включать/отключать:

- new_post_in_community
- new_post_from_followed_user
- comment_liked
- new_comment_in_post
- comment_reply

### Defaults (FR-428)

Новые пользователи получают все типы enabled.

### Granularity

- Per-type control (FR-427)
- Без per-community control (MVP)
- Без per-user control

---

## Хранение

### База данных (FR-401)

Каждое уведомление хранится с:

- user_id (recipient)
- type
- content (JSON)
- is_read
- created_at

### Retention (FR-429)

- Автоматическое удаление через 90 дней
- Background job периодически очищает
- Балансирует хранение и доступ к истории

---

## Производительность

### Индексы

- (user_id, created_at) для истории
- (user_id, is_read) для unread count
- (created_at) для cleanup job

### Кеширование

- unread_count в Redis
- Recent notifications для fast access
- Stream connections в memory

### Масштабирование

- Pub/sub для распределения streams
- Горизонтальное масштабирование workers
- Партиционирование по user_id

---

## Мониторинг

### Метрики

- Доставка latency
- Stream connection count
- Unread count distribution
- Notification type frequency
- Reconnection rate

### Alerts

- High latency доставки
- Превышение connection limit
- Высокий unread count (engagement issue)
