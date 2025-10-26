# Система жалоб (Reports)

## Обзор

ReportService предоставляет инструменты для пользователей жаловаться на нарушающий правила контент, и для модераторов обрабатывать эти жалобы.

## gRPC Service: ReportService

Proto файл: `proto/report.proto`

## Сущности

### ReportReason

```protobuf
enum ReportReason {
  REPORT_REASON_UNSPECIFIED      = 0
  REPORT_REASON_SPAM             = 1
  REPORT_REASON_HARASSMENT       = 2
  REPORT_REASON_MISINFORMATION   = 3
  REPORT_REASON_EXPLICIT_CONTENT = 4
  REPORT_REASON_VIOLENCE         = 5
  REPORT_REASON_HATE_SPEECH      = 6
  REPORT_REASON_OTHER            = 7
}
```

### ReportStatus

```protobuf
enum ReportStatus {
  REPORT_STATUS_UNSPECIFIED = 0
  REPORT_STATUS_PENDING     = 1
  REPORT_STATUS_RESOLVED    = 2
  REPORT_STATUS_DISMISSED   = 3
}
```

### ReportedContentType

```protobuf
enum ReportedContentType {
  REPORTED_CONTENT_TYPE_UNSPECIFIED = 0
  REPORTED_CONTENT_TYPE_POST        = 1
  REPORTED_CONTENT_TYPE_COMMENT     = 2
}
```

### Report

```protobuf
message Report {
  string id
  string reporter_id
  string reporter_username
  ReportedContentType content_type
  string content_id
  ReportReason reason
  string description
  ReportStatus status
  optional string resolver_id
  optional string resolver_username
  optional string resolution_note
  google.protobuf.Timestamp created_at
  optional google.protobuf.Timestamp resolved_at
}
```

## Endpoints

### Create

**RPC:** `Create(CreateRequest) returns (CreateResponse)`  
**HTTP:** `POST /reports`  
**FR:** FR-111-112, FR-149, FR-368-375, FR-496-503

Создание жалобы на контент.

**Request:**

```protobuf
message CreateRequest {
  ReportedContentType content_type  // post или comment
  string content_id
  ReportReason reason
  string description  // 10-1000 символов
}
```

**Response:**

```protobuf
message CreateResponse {
  Report report
}
```

**Требования:**

- Требуется аутентификация и report_content permission (FR-131, FR-369)
- Валидация существования контента (FR-370)
- Reason из предопределенного списка (FR-371)
- Description 10-1000 символов (FR-372)
- Статус устанавливается "pending" (FR-373)
- Reporter устанавливается текущий пользователь
- Возврат созданной жалобы с ID и timestamps (FR-374)
- Разрешены множественные жалобы на один контент (FR-375)

**Причины жалоб:**

- SPAM: нежелательная реклама, повторяющийся контент
- HARASSMENT: преследование, буллинг
- MISINFORMATION: ложная информация
- EXPLICIT_CONTENT: NSFW контент
- VIOLENCE: призывы к насилию, угрозы
- HATE_SPEECH: hate speech, дискриминация
- OTHER: другие нарушения

---

### Get

**RPC:** `Get(GetRequest) returns (GetResponse)`  
**HTTP:** `GET /reports/{report_id}`  
**FR:** FR-151

Получение детальной информации о жалобе.

**Request:**

```protobuf
message GetRequest {
  string report_id
}
```

**Response:**

```protobuf
message GetResponse {
  Report report
}
```

**Требования:**

- Требуется view_reports permission (FR-155)
- Возврат полной информации включая репортера, контент, статус, timestamps

---

### List

**RPC:** `List(ListRequest) returns (ListResponse)`  
**HTTP:** `GET /reports`  
**FR:** FR-150, FR-154, FR-155

Получение списка жалоб с фильтрацией.

**Request:**

```protobuf
message ListRequest {
  optional ReportStatus status_filter
  optional ReportedContentType content_type_filter
  string cursor
  int32 limit
}
```

**Response:**

```protobuf
message ListResponse {
  repeated Report reports
  string next_cursor
  bool has_more
}
```

**Требования:**

- Требуется view_reports permission (FR-155)
- Cursor-based пагинация (FR-154)
- Опциональная фильтрация по статусу
- Опциональная фильтрация по типу контента

**Фильтры:**

- status_filter: pending, resolved, dismissed
- content_type_filter: post, comment
- Без фильтров: все жалобы

---

### Resolve

**RPC:** `Resolve(ResolveRequest) returns (ResolveResponse)`  
**HTTP:** `POST /reports/{report_id}/resolve`  
**FR:** FR-152, FR-156

Разрешение жалобы (действие принято).

**Request:**

```protobuf
message ResolveRequest {
  string report_id
  optional string resolution_note
}
```

**Response:**

```protobuf
message ResolveResponse {
  Report report
}
```

**Требования:**

- Требуется resolve_reports permission (FR-156)
- Изменение статуса pending → resolved
- Установка resolver_id
- Установка resolution_note если указана
- Установка resolved_at timestamp
- Возврат обновленной жалобы

---

### Dismiss

**RPC:** `Dismiss(DismissRequest) returns (DismissResponse)`  
**HTTP:** `POST /reports/{report_id}/dismiss`  
**FR:** FR-153, FR-156

Отклонение жалобы (не требует действий).

**Request:**

```protobuf
message DismissRequest {
  string report_id
  optional string dismissal_reason
}
```

**Response:**

```protobuf
message DismissResponse {
  Report report
}
```

**Требования:**

- Требуется dismiss_reports permission (FR-156)
- Изменение статуса pending → dismissed
- Установка resolver_id
- Установка resolution_note с dismissal_reason
- Установка resolved_at timestamp

---

## Жизненный цикл жалобы

### 1. Создание (pending)

- Пользователь видит нарушающий контент
- Использует report_content permission (доступно @everyone)
- Заполняет причину и описание
- Жалоба создается со статусом pending

### 2. Рассмотрение

- Модератор с view_reports видит жалобу
- Изучает контент и context
- Принимает решение

### 3. Разрешение

**Опция A: Resolve (действие принято)**

- Модератор принимает меры (удаление, бан, etc.)
- Жалоба помечается resolved
- resolution_note описывает принятые меры

**Опция B: Dismiss (отклонение)**

- Модератор считает жалобу необоснованной
- Жалоба помечается dismissed
- dismissal_reason объясняет причину

---

## Разрешения

### report_content (FR-131)

- Доступно роли @everyone по умолчанию (FR-109, FR-110)
- Позволяет создавать жалобы
- Защита от abuse через rate limiting

### view_reports (FR-131, FR-155)

- Модераторы могут просматривать жалобы
- Доступ к списку и деталям

### resolve_reports (FR-131, FR-156)

- Модераторы могут разрешать жалобы
- Принятие корректирующих действий

### dismiss_reports (FR-131, FR-156)

- Модераторы могут отклонять жалобы
- Закрытие необоснованных жалоб

---

## Множественные жалобы

### Дубликаты разрешены (FR-375)

**Причины:**

- Разные пользователи могут жаловаться независимо
- Разные причины для одного контента
- Приоритизация по количеству жалоб

**Пример:**
Пост P1:

- User A: жалоба SPAM
- User B: жалоба HARASSMENT
- User C: жалоба SPAM

Все 3 жалобы существуют отдельно.

### Обработка дубликатов

- Модератор может resolve/dismiss все жалобы на контент
- Bulk actions для эффективности (опционально)
- Автоматическое закрытие при удалении контента (опционально)

---

## Связь с модерацией

### Workflow

1. Жалоба создана (pending)
2. Модератор рассматривает
3. Модератор принимает решение:
   - Delete контент → Resolve жалобу
   - Ban пользователя → Resolve жалобу
   - Mute пользователя → Resolve жалобу
   - Нет нарушения → Dismiss жалобу

### Действия модератора

После resolve жалобы модератор может:

- Удалить пост/комментарий (delete_any_post/comment)
- Забанить автора (ban_users)
- Замутить автора (mute_users)
- Другие корректирующие действия

---

## Валидация

### Description

- Минимум: 10 символов (FR-372)
- Максимум: 1000 символов (FR-372)
- Должна объяснять причину жалобы
- Помогает модератору принять решение

### Контент

- Пост или комментарий должен существовать (FR-370)
- Удаленный контент может вызвать ошибку
- Или: жалоба автоматически dismissed

---

## Статистика

### Для платформы (FR-324)

PlatformStatistics включает:

- pending_reports: количество ожидающих
- resolved_reports: разрешенных жалоб
- dismissed_reports: отклоненных жалоб

### Аналитика

Полезные метрики:

- Количество жалоб по типам
- Время рассмотрения (created_at → resolved_at)
- Ratio resolve/dismiss
- Топ репортеры (abuse detection)

---

## Производительность

### Индексы

- (status, created_at) для списков pending
- (content_type, content_id) для жалоб на контент
- (reporter_id, created_at) для истории репортера
- (resolver_id, resolved_at) для действий модератора

### Кеширование

- Счетчики pending reports
- Часто жалуемый контент (hot reports)

---

## Abuse prevention

### Rate limiting

Рекомендуется:

- Лимит на создание жалоб (например, 10 в час)
- Блокировка spam репортеров
- Анализ паттернов abuse

### Модерация репортеров

- Отслеживание dismissed жалоб
- Бан за злоупотребление системой
- False reports могут караться

---

## Приватность

### Анонимность репортера

Опции:

- Скрыть репортера от автора контента
- Показать модератору для context
- Защита от возмездия

### Публичная информация

Обычно НЕ публичны:

- Списки жалоб
- Детали жалоб
- Резолюции

Доступно только модераторам.
