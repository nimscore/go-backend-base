# Система модерации

## Обзор

ModerationService обеспечивает инструменты для модерации пользователей на платформе и в сообществах, включая баны, муты и логирование действий.

## gRPC Service: ModerationService

Proto файл: `proto/moderation.proto`

## Сущности

### MuteDuration

```protobuf
enum MuteDuration {
  MUTE_DURATION_UNSPECIFIED = 0
  MUTE_DURATION_1_HOUR      = 1
  MUTE_DURATION_24_HOURS    = 2
  MUTE_DURATION_7_DAYS      = 3
  MUTE_DURATION_30_DAYS     = 4
  MUTE_DURATION_PERMANENT   = 5
}
```

### ModerationAction

```protobuf
message ModerationAction {
  string id
  string moderator_id
  string moderator_username
  string target_user_id
  string target_username
  string action_type              // ban, unban, mute, unmute
  string reason
  optional string community_id
  google.protobuf.Timestamp created_at
}
```

## Платформенная модерация

### BanUser

**RPC:** `BanUser(BanUserRequest) returns (BanUserResponse)`  
**HTTP:** `POST /moderation/users/{user_id}/ban`  
**FR:** FR-254, FR-260-262

Платформенный бан пользователя.

**Request:**

```protobuf
message BanUserRequest {
  string user_id
  string reason
}
```

**Требования:**

- Требуется ban_users permission (FR-254)
- Запрет создания постов, комментариев, сообществ (FR-061, FR-260)
- Запрет лайков, закладок, взаимодействий (FR-063, FR-261)
- Существующий контент остается видимым (FR-060, FR-262)
- Профиль показывает "banned" статус (FR-062)
- Логирование действия (FR-267)

---

### UnbanUser

**RPC:** `UnbanUser(UnbanUserRequest) returns (UnbanUserResponse)`  
**HTTP:** `POST /moderation/users/{user_id}/unban`  
**FR:** FR-255

Снятие платформенного бана.

**Требования:**

- Требуется ban_users permission
- Восстановление всех возможностей
- Логирование действия

---

## Модерация в сообществах

### BanUserInCommunity

**RPC:** `BanUserInCommunity(BanUserInCommunityRequest) returns (BanUserInCommunityResponse)`  
**HTTP:** `POST /moderation/communities/{community_id}/users/{user_id}/ban`  
**FR:** FR-256, FR-263

Бан пользователя в конкретном сообществе.

**Request:**

```protobuf
message BanUserInCommunityRequest {
  string user_id
  string community_id
  string reason
}
```

**Требования:**

- Требуется ban_users permission в сообществе
- Запрет постинга/комментирования ТОЛЬКО в этом сообществе (FR-263)
- Остальные сообщества не затронуты
- Логирование с community context

---

### UnbanUserInCommunity

**RPC:** `UnbanUserInCommunity(UnbanUserInCommunityRequest) returns (UnbanUserInCommunityResponse)`  
**HTTP:** `POST /moderation/communities/{community_id}/users/{user_id}/unban`  
**FR:** FR-257

Снятие бана в сообществе.

---

### MuteUserInCommunity

**RPC:** `MuteUserInCommunity(MuteUserInCommunityRequest) returns (MuteUserInCommunityResponse)`  
**HTTP:** `POST /moderation/communities/{community_id}/users/{user_id}/mute`  
**FR:** FR-258, FR-264-266

Временный мут пользователя в сообществе.

**Request:**

```protobuf
message MuteUserInCommunityRequest {
  string user_id
  string community_id
  MuteDuration duration
  string reason
}
```

**Response:**

```protobuf
message MuteUserInCommunityResponse {
  string message
  google.protobuf.Timestamp muted_until
  ModerationAction action
}
```

**Требования:**

- Требуется mute_users permission (FR-258)
- Duration: 1 час, 24 часа, 7 дней, 30 дней, permanent (FR-264)
- Автоматический unmute после окончания duration (FR-265)
- Мутнутый может читать, но не создавать контент (FR-266)
- Логирование действия

**Длительности:**

- MUTE_DURATION_1_HOUR: 1 час
- MUTE_DURATION_24_HOURS: 1 день
- MUTE_DURATION_7_DAYS: 1 неделя
- MUTE_DURATION_30_DAYS: 1 месяц
- MUTE_DURATION_PERMANENT: бессрочно

---

### UnmuteUserInCommunity

**RPC:** `UnmuteUserInCommunity(UnmuteUserInCommunityRequest) returns (UnmuteUserInCommunityResponse)`  
**HTTP:** `POST /moderation/communities/{community_id}/users/{user_id}/unmute`  
**FR:** FR-259

Снятие мута в сообществе.

**Требования:**

- Требуется mute_users permission
- Немедленное восстановление возможности постить

---

## Логи модерации

### ListModerationLogs

**RPC:** `ListModerationLogs(ListModerationLogsRequest) returns (ListModerationLogsResponse)`  
**HTTP:** `GET /moderation/logs`  
**FR:** FR-267, FR-446-451

Просмотр истории действий модерации.

**Request:**

```protobuf
message ListModerationLogsRequest {
  optional string community_id  // если пусто - platform-wide
  string cursor
  int32 limit
}
```

**Response:**

```protobuf
message ListModerationLogsResponse {
  repeated ModerationAction actions
  string next_cursor
  bool has_more
}
```

**Требования:**

- Требуется view_moderation_logs permission (FR-449)
- Cursor-based пагинация (FR-448)
- Без community_id: все действия платформы (FR-447)
- С community_id: только действия в сообществе
- Сортировка по created_at в обратном порядке (FR-450)
- Каждое действие включает (FR-451):
  - Информация о модераторе
  - Информация о целевом пользователе
  - Тип действия
  - Причина
  - Community context (если применимо)
  - Timestamp

---

## Типы действий

### action_type значения:

- "ban" - платформенный или community бан
- "unban" - снятие бана
- "mute" - временный мут в сообществе
- "unmute" - снятие мута

---

## Эффекты модерации

### Платформенный бан

**Запрещено (FR-061, FR-063):**

- Создавать посты
- Создавать комментарии
- Создавать сообщества
- Лайкать контент
- Добавлять закладки
- Подписываться на пользователей
- Любые другие взаимодействия

**Разрешено:**

- Просматривать контент (если в сессии)
- Logout

**Видимость:**

- Существующий контент остается видимым (FR-060)
- Профиль показывает "banned" индикатор (FR-062)
- Email и приватные данные скрыты

---

### Бан в сообществе

**Запрещено (FR-263):**

- Создавать посты в этом сообществе
- Комментировать в этом сообществе

**Разрешено:**

- Все действия в других сообществах
- Просмотр контента забаненного сообщества
- Все остальные взаимодействия на платформе

---

### Мут в сообществе

**Запрещено (FR-266):**

- Создавать посты в этом сообществе
- Комментировать в этом сообществе

**Разрешено:**

- Читать контент сообщества
- Лайкать в сообществе
- Все действия в других сообществах

**Особенности:**

- Временное ограничение (FR-264)
- Автоматическое снятие (FR-265)

---

## Логирование

### Обязательная информация (FR-267)

Каждое действие модерации должно логироваться с:

- ID модератора и username
- ID целевого пользователя и username
- Тип действия
- Причина (обязательна для ban/mute)
- Community ID если применимо
- Timestamp

### Использование логов

- Аудит действий модераторов
- Разрешение споров
- Аналитика модерации
- Выявление паттернов abuse

---

## Разрешения модерации

### ban_users (FR-127)

- Бан и unbан на платформе (требует платформенной роли)
- Бан и unbан в сообществе (требует community роли)

### mute_users (FR-127)

- Мут и unmut в сообществах

### view_moderation_logs (FR-127)

- Просмотр логов модерации
- Аудит действий других модераторов

---

## Каскадные эффекты

### При бане пользователя

- Завершаются все активные сессии (опционально)
- Existing content помечается как "banned user"
- Уведомления пользователю (опционально)

### При анбане

- Восстановление всех прав
- Удаление индикаторов "banned"

---

## Автоматизация

### Автоматический unmute

- Background job проверяет muted_until timestamp
- Автоматическое снятие по истечении (FR-265)
- Логирование автоматического действия

### Rate limiting

Рекомендуется лимит на модерацию:

- Предотвращение mass ban abuse
- Требование подтверждения для bulk действий

---

## Уведомления

### Целевому пользователю (опционально)

- Email о бане с причиной
- Длительность мута
- Контакты для appeal

### Другим модераторам

- Уведомления о значимых действиях
- Логи в модерационном канале

---

## Производительность

### Индексы

- (target_user_id, created_at) для истории действий по пользователю
- (moderator_id, created_at) для действий модератора
- (community_id, created_at) для логов сообщества

### Кеширование

- Banned status пользователей
- Muted until timestamps
- Часто проверяемые разрешения
