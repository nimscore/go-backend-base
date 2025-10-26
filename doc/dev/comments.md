# Система комментариев

## Обзор

CommentService управляет созданием, редактированием комментариев с поддержкой вложенности, медиа вложений и real-time обновлений через gRPC streaming.

## gRPC Service: CommentService

Proto файл: `proto/comment.proto`

## Сущности

### Comment

```protobuf
message Comment {
  string id
  string post_id
  string author_id
  string author_username
  string author_avatar
  string text
  optional string parent_comment_id    // для вложенных ответов
  repeated MediaAttachment attachments
  int32 like_count
  bool is_liked_by_me
  bool is_edited
  google.protobuf.Timestamp created_at
  google.protobuf.Timestamp updated_at
}
```

### MediaAttachment

```protobuf
message MediaAttachment {
  string url        // S3 URL
  string type       // image, video, audio, gif
  int64 size_bytes
}
```

### CommentEventType

```protobuf
enum CommentEventType {
  COMMENT_EVENT_TYPE_UNSPECIFIED = 0
  COMMENT_EVENT_TYPE_CREATED     = 1
  COMMENT_EVENT_TYPE_UPDATED     = 2
  COMMENT_EVENT_TYPE_DELETED     = 3
}
```

### CommentEvent

```protobuf
message CommentEvent {
  CommentEventType event_type
  Comment comment
  google.protobuf.Timestamp timestamp
}
```

## Endpoints

### Create

**RPC:** `Create(CreateRequest) returns (CreateResponse)`  
**HTTP:** `POST /comments`  
**FR:** FR-025-028, FR-335-344

Создание нового комментария.

**Request:**

```protobuf
message CreateRequest {
  string post_id
  string text                       // 1-10000 символов
  optional string parent_comment_id // для ответов
  repeated string attachment_urls   // S3 URLs
}
```

**Response:**

```protobuf
message CreateResponse {
  Comment comment
}
```

**Требования:**

- Пользователь должен быть верифицирован (FR-009, FR-337)
- Пост должен существовать и быть опубликованным (FR-338)
- parent_comment_id должен существовать если указан (FR-339)
- Text 1-10000 символов (FR-340)
- attachment_urls опциональны, максимум 5 файлов (FR-065, FR-341)
- Валидация размеров вложений (FR-064, FR-342):
  - Images: 10MB
  - Video: 100MB
  - Audio: 20MB
  - GIF: 15MB
- Автор устанавливается в текущего пользователя (FR-343)
- Возврат созданного комментария (FR-344)
- Триггер real-time stream update (FR-031, FR-344)
- Создание уведомления автору поста или родительского комментария

**Ошибки:**

- Пользователь не верифицирован
- Пост не найден или не опубликован
- parent_comment_id не найден
- Text невалидной длины
- Превышено количество вложений
- Превышен размер вложения
- Невалидный тип медиа (FR-067)
- Пользователь забанен в сообществе

---

### Get

**RPC:** `Get(GetRequest) returns (GetResponse)`  
**HTTP:** `GET /comments/{comment_id}`  
**FR:** FR-216, FR-219

Получение комментария по ID.

**Request:**

```protobuf
message GetRequest {
  string comment_id
}
```

**Response:**

```protobuf
message GetResponse {
  Comment comment
}
```

**Требования:**

- Возврат полной информации (FR-219):
  - text
  - author info (id, username, avatar)
  - parent_comment_id если это ответ
  - media attachments
  - like_count
  - timestamps
- Поле is_liked_by_me требует аутентификации

**Ошибки:**

- Комментарий не найден
- Пост комментария был удален

---

### Update

**RPC:** `Update(UpdateRequest) returns (UpdateResponse)`  
**HTTP:** `PATCH /comments/{comment_id}`  
**FR:** FR-029, FR-384-390

Обновление существующего комментария.

**Request:**

```protobuf
message UpdateRequest {
  string comment_id
  string text  // 1-10000 символов
}
```

**Response:**

```protobuf
message UpdateResponse {
  Comment comment
}
```

**Требования:**

- Требуется быть автором или иметь edit_any_comment permission (FR-132, FR-386)
- Text 1-10000 символов (FR-387)
- НЕ может изменять (FR-388):
  - author
  - post_id
  - parent_comment_id
  - created_at
  - attachments (нельзя добавить/удалить после создания)
- Автоматическое обновление updated_at (FR-389)
- Установка флага is_edited (FR-389)
- Возврат обновленного Comment (FR-390)
- Триггер real-time stream update (FR-031, FR-390)

**Ошибки:**

- Недостаточно прав
- Text невалидной длины
- Комментарий не найден

---

### Delete

**RPC:** `Delete(DeleteRequest) returns (DeleteResponse)`  
**HTTP:** `DELETE /comments/{comment_id}`  
**FR:** FR-030

Удаление комментария.

**Request:**

```protobuf
message DeleteRequest {
  string comment_id
}
```

**Response:**

```protobuf
message DeleteResponse {
  string message
}
```

**Требования:**

- Может удалить (FR-030):
  - Автор комментария
  - Автор поста (модератор поста)
  - Модератор сообщества с delete_any_comment
  - Платформенный модератор с delete_any_comment
- Cascade удаление дочерних комментариев (или soft delete)
- Удаление всех лайков
- Триггер real-time stream update (FR-034)

**Ошибки:**

- Недостаточно прав
- Комментарий не найден

---

### Like

**RPC:** `Like(LikeRequest) returns (LikeResponse)`  
**HTTP:** `POST /comments/{comment_id}/like`  
**FR:** FR-173, FR-175, FR-177, FR-178

Лайк комментария.

**Request:**

```protobuf
message LikeRequest {
  string comment_id
}
```

**Response:**

```protobuf
message LikeResponse {
  string message
  int32 new_like_count
}
```

**Требования:**

- Пользователь должен быть верифицирован
- Идемпотентность: повторный лайк возвращает success (FR-175)
- Немедленное обновление like_count (FR-177)
- Обновление репутации автора (FR-178, FR-053)
- Создание уведомления автору комментария

**Ошибки:**

- Пользователь не верифицирован
- Пользователь забанен (FR-063)
- Комментарий не найден

---

### Unlike

**RPC:** `Unlike(UnlikeRequest) returns (UnlikeResponse)`  
**HTTP:** `DELETE /comments/{comment_id}/like`  
**FR:** FR-174, FR-176

Удаление лайка с комментария.

**Request:**

```protobuf
message UnlikeRequest {
  string comment_id
}
```

**Response:**

```protobuf
message UnlikeResponse {
  string message
  int32 new_like_count
}
```

**Требования:**

- Идемпотентность: удаление несуществующего лайка возвращает success (FR-176)
- Немедленное обновление like_count (FR-177)
- Обновление репутации автора (FR-178)

---

## Real-time Streaming

### Stream

**RPC:** `Stream(StreamRequest) returns (stream CommentEvent)`  
**gRPC Streaming:** Server-Side Streaming  
**FR:** FR-031-034, FR-081

Подписка на real-time обновления комментариев.

**Request:**

```protobuf
message StreamRequest {
  optional string post_id  // если пусто, глобальный поток
}
```

**Response Stream:**

```protobuf
stream CommentEvent {
  CommentEventType event_type  // CREATED, UPDATED, DELETED
  Comment comment
  google.protobuf.Timestamp timestamp
}
```

**Требования:**

- Если post_id указан: stream только комментариев к этому посту (FR-032)
- Если post_id пуст: stream всех комментариев платформы (FR-033)
- События (FR-034):
  - CREATED: новый комментарий
  - UPDATED: отредактированный комментарий
  - DELETED: удаленный комментарий
- Подключение остается открытым до отключения клиента
- Нет HTTP mapping (только gRPC)
- Поддержка тысяч одновременных подключений (SC-004)
- Доставка обновлений в течение 1 секунды (SC-003)

**События:**

**CREATED:**

- Новый комментарий добавлен
- Полный объект Comment в event

**UPDATED:**

- Комментарий отредактирован
- is_edited = true
- Обновленный текст

**DELETED:**

- Комментарий удален
- Только ID комментария (остальные поля пустые/null)

---

## Вложенная структура (Threading)

### Родительские и дочерние комментарии

Комментарии поддерживают древовидную структуру через parent_comment_id:

```
Comment 1 (parent_comment_id = null)
├─ Comment 2 (parent_comment_id = Comment 1)
│  └─ Comment 3 (parent_comment_id = Comment 2)
└─ Comment 4 (parent_comment_id = Comment 1)
```

### Глубина вложенности

- Теоретически неограниченная глубина
- Рекомендуется до 10 уровней для UX (SC-010)
- Клиент должен правильно отображать иерархию

### Сортировка

- Корневые комментарии: по created_at по возрастанию (FR-220)
- Дочерние комментарии: обычно также по created_at
- Клиент может реализовать альтернативную сортировку

---

## Медиа вложения

### Поддерживаемые типы

- **image**: JPEG, PNG, WebP
- **video**: MP4, WebM
- **audio**: MP3, WAV, OGG
- **gif**: анимированные GIF

### Лимиты размеров

- Images: максимум 10MB (FR-064)
- Video: максимум 100MB (FR-064)
- Audio: максимум 20MB (FR-064)
- GIF: максимум 15MB (FR-064)

### Количество

- Максимум 5 файлов на комментарий (FR-065)
- Можно комбинировать разные типы

### Процесс загрузки

1. Клиент загружает файлы через MediaService.Upload
2. Получает S3 URLs
3. Передает URLs в CreateComment.attachment_urls
4. Backend валидирует:
   - Файлы существуют в S3 (FR-146)
   - Размеры соответствуют лимитам (FR-066)
   - Типы соответствуют заявленным (FR-067)

### Валидация

- Проверка размера при загрузке (FR-147)
- Проверка типа файла по magic bytes
- Отклонение превышающих лимит с четкой ошибкой (FR-066)

---

## Модерация комментариев

### Права на удаление

Удалить комментарий могут (FR-030):

1. Автор комментария
2. Автор поста (владелец треда)
3. Модератор сообщества с delete_any_comment
4. Платформенный модератор с delete_any_comment

### Редактирование

Редактировать могут (FR-386):

1. Автор комментария
2. Модератор с edit_any_comment permission

### Индикация редактирования

- Флаг is_edited устанавливается при первом редактировании
- updated_at обновляется при каждом редактировании
- История изменений не хранится (опционально для будущего)

---

## Репутация автора

При получении/удалении лайка на комментарии:

- Обновляется reputation автора (FR-053, FR-178)
- comment_likes в статистике пользователя
- total_likes_received в статистике пользователя

---

## Уведомления

Комментарии создают уведомления:

1. **NEW_COMMENT_IN_POST** - автору поста при новом комментарии (FR-406)
2. **COMMENT_REPLY** - автору комментария при ответе (FR-407)
3. **COMMENT_LIKED** - автору комментария при лайке (FR-405)

---

## Производительность

### Real-time обновления

- Целевая задержка: < 1 секунда (SC-003)
- Поддержка 10,000+ одновременных подключений (SC-004)
- Эффективная рассылка через pub/sub механизм

### Индексы

Рекомендуемые индексы:

- (post_id, created_at) для списков комментариев
- (parent_comment_id) для дочерних комментариев
- (author_id, created_at) для комментариев пользователя

### Кеширование

Рекомендуется кешировать:

- Счетчики (like_count)
- Hot threads (часто комментируемые посты)
- Данные автора

---

## Ограничения

### Создание

- Только верифицированные пользователи (FR-009)
- Только на опубликованные посты
- Запрещено забаненным пользователям (FR-061)

### Текст

- Минимум: 1 символ (FR-340)
- Максимум: 10000 символов (FR-340)
- Поддержка Unicode
- Может содержать Markdown (обработка на клиенте)

### Медиа

- Максимум 5 файлов (FR-065)
- Размеры по типам (FR-064)
- Нельзя изменить после создания

---

## Каскадное удаление

При удалении комментария:

- Опция 1: Cascade удаление всех дочерних комментариев
- Опция 2: Soft delete (замена текста на "[deleted]")
- Удаление всех лайков

При удалении поста:

- Удаляются все комментарии (с каскадом выше)

При удалении пользователя:

- Опция: комментарии остаются с "[deleted user]"
- Или: cascade удаление всех комментариев
