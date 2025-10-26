# Загрузка медиа файлов

## Обзор

MediaService обеспечивает прямую загрузку медиа файлов (изображения, видео, аудио, GIF) в S3-совместимое хранилище через gRPC streaming с валидацией и подтверждением.

## gRPC Service: MediaService

Proto файл: `proto/media.proto`

## Сущности

### RelationType

```protobuf
enum RelationType {
  MEDIA_TYPE_UNSPECIFIED        = 0
  MEDIA_TYPE_USER_AVATAR        = 1
  MEDIA_TYPE_USER_BANNER        = 2
  MEDIA_TYPE_COMMENT_ATTACHMENT = 3
  MEDIA_TYPE_PLATFORM_LOGO      = 4
  MEDIA_TYPE_PLATFORM_BANNER    = 5
  MEDIA_TYPE_COMMUNITY_LOGO     = 6
  MEDIA_TYPE_COMMUNITY_BANNER   = 7
}
```

### FileType

```protobuf
enum FileType {
  FILE_TYPE_UNSPECIFIED = 0
  FILE_TYPE_IMAGE       = 1  // max 10MB
  FILE_TYPE_VIDEO       = 2  // max 100MB
  FILE_TYPE_AUDIO       = 3  // max 20MB
  FILE_TYPE_GIF         = 4  // max 15MB
}
```

## Endpoints

### Upload

**RPC:** `Upload(UploadRequest) returns (UploadResponse)`  
**HTTP:** `POST /media/upload`  
**FR:** FR-141-144, FR-147-148

Прямая загрузка медиа файла через streaming.

**Request:**

```protobuf
message UploadRequest {
  bytes chunk                // данные файла (chunks)
  RelationType relation_type // тип использования
  string relation_id         // ID связанной сущности
}
```

**Response:**

```protobuf
message UploadResponse {
  string id  // уникальный ID загрузки
}
```

**Требования:**

- Принимает файл как streaming chunks для больших файлов (FR-143)
- relation_type указывает назначение файла (FR-142)
- Валидация permissions перед приемом (FR-144):
  - Только owner платформы может загружать platform_logo/banner
  - Только owner сообщества может загружать community_logo/banner
  - Владелец профиля может загружать user_avatar/banner
  - Верифицированные пользователи могут загружать comment_attachment
- Валидация размера во время загрузки (FR-147)
- Возврат уникального upload ID (FR-148)
- Сохранение в S3-совместимое хранилище

**Ошибки:**

- Недостаточно прав для типа загрузки
- Превышен лимит размера
- Невалидный формат файла
- Ошибка соединения с S3

---

### Confirm

**RPC:** `Confirm(ConfirmRequest) returns (ConfirmResponse)`  
**HTTP:** `POST /media/confirm`  
**FR:** FR-145-146

Подтверждение завершения загрузки и сохранение ссылки.

**Request:**

```protobuf
message ConfirmRequest {
  string id  // upload ID из Upload response
}
```

**Response:**

```protobuf
message ConfirmResponse {
  string message
}
```

**Требования:**

- Валидация существования файла в S3 (FR-146)
- Сохранение ссылки на файл в БД
- Связывание с соответствующей сущностью (user, community, comment)
- Атомарная операция

**Ошибки:**

- Upload ID не найден
- Файл не найден в S3
- Истек timeout загрузки

---

## Процесс загрузки

### Полный workflow

1. **Клиент инициирует Upload**
   - Отправляет файл chunks через gRPC streaming
   - Указывает relation_type и relation_id
2. **Backend валидирует**

   - Проверяет permissions
   - Проверяет размер по мере загрузки
   - Валидирует тип файла

3. **Backend сохраняет в S3**

   - Генерирует уникальное имя файла
   - Сохраняет в S3 bucket
   - Возвращает upload ID

4. **Клиент вызывает Confirm**

   - Передает полученный upload ID
   - Backend проверяет файл в S3

5. **Backend finalize**
   - Сохраняет S3 URL в БД
   - Связывает с соответствующей сущностью
   - Файл становится доступен

---

## Типы медиа

### USER_AVATAR

**Назначение:** Аватар пользователя

**Права:** Владелец профиля

**Рекомендации:**

- Квадратное изображение
- Минимум 200x200px
- Формат: JPEG, PNG, WebP

**Размер:** До 10MB (IMAGE)

---

### USER_BANNER

**Назначение:** Баннер профиля пользователя

**Права:** Владелец профиля

**Рекомендации:**

- Широкое изображение (например, 1500x500px)
- Формат: JPEG, PNG, WebP

**Размер:** До 10MB (IMAGE)

---

### COMMENT_ATTACHMENT

**Назначение:** Вложение к комментарию

**Права:** Верифицированные пользователи

**Типы:** Image, Video, Audio, GIF

**Лимиты:**

- Images: 10MB
- Video: 100MB
- Audio: 20MB
- GIF: 15MB
- Максимум 5 файлов на комментарий (проверяется в CommentService)

---

### PLATFORM_LOGO

**Назначение:** Логотип платформы

**Права:** Только платформенный owner (FR-144)

**Рекомендации:**

- Квадратное или близкое
- Минимум 512x512px
- Прозрачный фон (PNG)

**Размер:** До 10MB

---

### PLATFORM_BANNER

**Назначение:** Баннер платформы (главная страница)

**Права:** Только платформенный owner

**Размер:** До 10MB

---

### COMMUNITY_LOGO

**Назначение:** Логотип сообщества

**Права:** Community owner

**Размер:** До 10MB

---

### COMMUNITY_BANNER

**Назначение:** Баннер сообщества

**Права:** Community owner

**Размер:** До 10MB

---

## Лимиты размеров

### По типу файла (FR-064)

- **IMAGE**: 10MB
- **VIDEO**: 100MB
- **AUDIO**: 20MB
- **GIF**: 15MB

### Enforcement

- Валидация во время загрузки (FR-147)
- Прерывание загрузки при превышении
- Четкое сообщение об ошибке (FR-066)

---

## Валидация файлов

### Тип файла (FR-067)

**Проверка:**

- Magic bytes (file signature)
- MIME type
- File extension

**Примеры:**

```
JPEG: FF D8 FF
PNG: 89 50 4E 47
GIF: 47 49 46 38
MP4: 66 74 79 70
```

### Размер

- Continuous проверка во время streaming upload
- Abort при превышении лимита
- Return error с детализацией

### Формат

**Поддерживаемые форматы:**

- Images: JPEG, PNG, WebP
- Video: MP4, WebM
- Audio: MP3, WAV, OGG
- GIF: GIF89a, GIF87a

---

## S3 Storage

### Структура bucket

```
/media/
  /avatars/
    /user-{id}/
      {timestamp}-{random}.jpg
  /banners/
    /user-{id}/
      {timestamp}-{random}.jpg
  /comments/
    /{post-id}/
      {comment-id}-{n}.{ext}
  /platform/
    logo.png
    banner.jpg
  /communities/
    /{community-id}/
      logo.png
      banner.jpg
```

### URL формат

- Публичные URL для аватаров, баннеров
- Опционально signed URLs для приватного контента
- CDN перед S3 для производительности

---

## Хранение в БД

### Поля с медиа (FR-076)

Только S3 URLs/keys хранятся:

**User:**

- avatar_url: string
- banner_url: string

**Community:**

- logo_url: string (опционально)
- banner_url: string (опционально)

**PlatformSettings:**

- logo_url: string
- banner_url: string
- auth_banner_url: string

**MediaAttachment (для комментариев):**

- url: string
- type: string
- size_bytes: int64

### Разделение (FR-074-075)

- **S3**: Все медиа файлы
- **БД**: Структурированные данные, текст, JSON content

---

## Безопасность

### Permissions

**Строгая валидация (FR-144):**

- Platform logo/banner: только platform owner
- Community logo/banner: только community owner
- User avatar/banner: только владелец профиля
- Comment attachments: верифицированные пользователи

### Антивирус (опционально)

- Сканирование загруженных файлов
- Quarantine подозрительных
- Async проверка после загрузки

### Rate limiting

- Лимит на количество загрузок в час
- Защита от spam uploads
- Tracking по user_id

---

## Удаление файлов

### Orphaned files

Background job для очистки:

- Файлы без ссылок из БД
- Старые temporary uploads
- Deleted user/community медиа

### Cascade delete

При удалении:

- User → удалить avatar, banner
- Community → удалить logo, banner
- Comment → удалить attachments

---

## Производительность

### Streaming upload

- Chunk размер: 64KB - 256KB рекомендуется
- Concurrent chunks для больших файлов
- Resume capability (опционально)

### CDN

- CloudFront, CloudFlare перед S3
- Кеширование публичных медиа
- Geographic распределение

### Оптимизация изображений

- Автоматический resize аватаров
- Генерация thumbnails
- WebP конвертация для поддерживающих браузеров

---

## Мониторинг

### Метрики

- Загрузка latency
- Успешность uploads (success rate)
- Storage utilization
- Bandwidth usage
- Тип файлов distribution

### Alerts

- Высокая failure rate
- S3 connection issues
- Превышение storage quota
- Suspicious upload patterns

---

## Примеры использования

### Загрузка аватара

1. Client вызывает Upload с chunks и relation_type=USER_AVATAR
2. Server сохраняет в S3: `/media/avatars/user-123/1234567890-abc.jpg`
3. Server возвращает upload_id
4. Client вызывает Confirm с upload_id
5. Server сохраняет URL в user.avatar_url

### Загрузка вложения комментария

1. Client загружает файл с relation_type=COMMENT_ATTACHMENT
2. Получает S3 URL
3. Client создает комментарий с attachment_urls=[url]
4. CommentService валидирует и сохраняет

---

## Ограничения и рекомендации

### Общие лимиты

- 5 вложений на комментарий (FR-065)
- Размеры по типам (FR-064)
- Timeout загрузки: 5 минут для больших файлов

### Best practices

- Resize изображений на клиенте перед загрузкой
- Использовать WebP когда возможно
- Compress видео перед загрузкой
- Показывать progress bar пользователю
