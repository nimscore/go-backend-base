# Система наград (Badges)

## Обзор

BadgeService предоставляет функционал для создания, модерации и выдачи наград пользователям и сообществам. Награды делятся на платформенные (platform badges) и награды сообществ (community badges).

## gRPC Service: BadgeService

Proto файл: `proto/badge.proto`

## Сущности

### Badge

```protobuf
message Badge {
  string id
  string name
  string description
  string icon_url
  RarityLevel rarity
  BadgeType type                    // PLATFORM или COMMUNITY
  BadgeStatus status                // ACTIVE, PENDING_APPROVAL, REJECTED
  optional string community_id      // заполнено только для community badges
  string created_by                 // ID модератора создавшего награду
  google.protobuf.Timestamp created_at
  google.protobuf.Timestamp updated_at
}

enum RarityLevel {
  COMMON = 0;
  RARE = 1;
  EPIC = 2;
  LEGENDARY = 3;
}

enum BadgeType {
  PLATFORM = 0;
  COMMUNITY = 1;
}

enum BadgeStatus {
  ACTIVE = 0;
  PENDING_APPROVAL = 1;
  REJECTED = 2;
}
```

### BadgeAward (Выдача награды)

```protobuf
message BadgeAward {
  string id
  string badge_id
  Badge badge                       // полная информация о награде
  string recipient_id               // user_id или community_id
  RecipientType recipient_type      // USER или COMMUNITY
  string awarded_by                 // ID модератора выдавшего награду
  string awarded_by_username        // username модератора
  optional string reason            // причина выдачи
  google.protobuf.Timestamp awarded_at
}

enum RecipientType {
  USER = 0;
  COMMUNITY = 1;
}
```

## Типы наград

### Platform Badges (Платформенные награды)

- Создаются модераторами платформы с правом `create_platform_badges`
- НЕ требуют модерации - статус сразу `ACTIVE` (FR-472, FR-479)
- Могут быть выданы пользователям и сообществам (FR-469)
- Выдаются модераторами с правом `award_platform_badges`

### Community Badges (Награды сообществ)

- Создаются владельцем или модераторами сообщества с правом `create_community_badges` (FR-470, FR-478)
- ТРЕБУЮТ модерации платформой - статус `PENDING_APPROVAL` (FR-471, FR-480)
- Могут быть выданы только участникам этого сообщества (FR-494)
- Выдаются модераторами сообщества с правом `award_community_badges` (FR-495)
- Владелец сообщества всегда имеет все badge permissions для своего сообщества (FR-522)

## Endpoints

### CreatePlatformBadge

**RPC:** `CreatePlatformBadge(CreatePlatformBadgeRequest) returns (CreatePlatformBadgeResponse)`  
**HTTP:** `POST /badges/platform`  
**FR:** FR-473, FR-475, FR-477, FR-479, FR-481-483

Создание платформенной награды.

**Request:**

```protobuf
message CreatePlatformBadgeRequest {
  string name
  string description
  string icon_url
  RarityLevel rarity
}
```

**Response:**

```protobuf
message CreatePlatformBadgeResponse {
  Badge badge
}
```

**Требования:**

- Требуется право `create_platform_badges` (FR-477)
- Валидация:
  - Название: 3-100 символов (FR-481)
  - Описание: 10-500 символов (FR-482)
  - Уровень редкости: COMMON, RARE, EPIC, LEGENDARY (FR-483)
- Награда создается со статусом `ACTIVE` сразу (FR-479)
- Модерация НЕ требуется (FR-472)

**Ошибки:**

- Недостаточно прав
- Невалидные параметры (длина строк, неизвестный rarity)

---

### CreateCommunityBadge

**RPC:** `CreateCommunityBadge(CreateCommunityBadgeRequest) returns (CreateCommunityBadgeResponse)`  
**HTTP:** `POST /badges/community`  
**FR:** FR-474, FR-476, FR-478, FR-480, FR-481-483

Создание награды сообщества (требует модерации).

**Request:**

```protobuf
message CreateCommunityBadgeRequest {
  string community_id
  string name
  string description
  string icon_url
  RarityLevel rarity
}
```

**Response:**

```protobuf
message CreateCommunityBadgeResponse {
  Badge badge
}
```

**Требования:**

- Требуется право `create_community_badges` для этого сообщества (FR-478)
- Владелец сообщества всегда имеет это право (FR-522)
- Валидация аналогична CreatePlatformBadge (FR-481-483)
- Награда создается со статусом `PENDING_APPROVAL` (FR-480)
- Отправляется уведомление модераторам платформы (FR-480)

**Ошибки:**

- Недостаточно прав
- Сообщество не найдено
- Невалидные параметры

---

### ApproveCommunityBadge

**RPC:** `ApproveCommunityBadge(ApproveCommunityBadgeRequest) returns (ApproveCommunityBadgeResponse)`  
**HTTP:** `POST /badges/community/{badge_id}/approve`  
**FR:** FR-484, FR-486, FR-489

Одобрение награды сообщества модератором платформы.

**Request:**

```protobuf
message ApproveCommunityBadgeRequest {
  string badge_id
}
```

**Response:**

```protobuf
message ApproveCommunityBadgeResponse {
  Badge badge
}
```

**Требования:**

- Требуется право `approve_community_badges` (FR-486)
- Награда меняет статус с `PENDING_APPROVAL` на `ACTIVE` (FR-489)
- Отправляется уведомление владельцу сообщества об одобрении (FR-489)
- После одобрения награду можно выдавать участникам

**Ошибки:**

- Недостаточно прав
- Награда не найдена
- Награда не в статусе PENDING_APPROVAL

---

### RejectCommunityBadge

**RPC:** `RejectCommunityBadge(RejectCommunityBadgeRequest) returns (RejectCommunityBadgeResponse)`  
**HTTP:** `POST /badges/community/{badge_id}/reject`  
**FR:** FR-485, FR-487, FR-488, FR-490

Отклонение награды сообщества с причиной.

**Request:**

```protobuf
message RejectCommunityBadgeRequest {
  string badge_id
  string reason       // обязательное поле
}
```

**Response:**

```protobuf
message RejectCommunityBadgeResponse {
  Badge badge
}
```

**Требования:**

- Требуется право `approve_community_badges` (FR-487)
- Причина отклонения обязательна (FR-487)
- Награда меняет статус на `REJECTED` (FR-488)
- Отправляется уведомление владельцу сообщества с причиной (FR-488)
- Отклоненную награду можно отредактировать и отправить на повторную модерацию (FR-490)

**Ошибки:**

- Недостаточно прав
- Награда не найдена
- Отсутствует причина отклонения

---

### AwardBadgeToUser

**RPC:** `AwardBadgeToUser(AwardBadgeToUserRequest) returns (AwardBadgeToUserResponse)`  
**HTTP:** `POST /badges/{badge_id}/award/user`  
**FR:** FR-491, FR-493-495, FR-497-498

Выдача награды пользователю.

**Request:**

```protobuf
message AwardBadgeToUserRequest {
  string badge_id
  string user_id
  optional string reason
}
```

**Response:**

```protobuf
message AwardBadgeToUserResponse {
  BadgeAward award
}
```

**Требования:**

**Для платформенных наград:**

- Требуется право `award_platform_badges` (FR-493)
- Можно выдать любому пользователю

**Для наград сообщества:**

- Награда должна быть в статусе `ACTIVE` (одобрена) (FR-494)
- Пользователь должен быть участником сообщества (FR-494)
- Требуется право `award_community_badges` для этого сообщества (FR-495)
- Владелец сообщества всегда имеет это право (FR-522)

**Общие требования:**

- Уникальность: одна награда один раз одному пользователю (FR-497)
- Записывается: badge_id, user_id, awarded_by, awarded_at, optional reason (FR-498)

**Ошибки:**

- Недостаточно прав
- Награда не найдена или не активна
- Пользователь не является участником сообщества (для community badges)
- Награда уже выдана этому пользователю

---

### AwardBadgeToCommunity

**RPC:** `AwardBadgeToCommunity(AwardBadgeToCommunityRequest) returns (AwardBadgeToCommunityResponse)`  
**HTTP:** `POST /badges/{badge_id}/award/community`  
**FR:** FR-492, FR-496, FR-497-498

Выдача награды сообществу (только платформенные награды).

**Request:**

```protobuf
message AwardBadgeToCommunityRequest {
  string badge_id
  string community_id
  optional string reason
}
```

**Response:**

```protobuf
message AwardBadgeToCommunityResponse {
  BadgeAward award
}
```

**Требования:**

- Только платформенные награды могут быть выданы сообществам (FR-496)
- Требуется право `award_platform_badges` (FR-496)
- Уникальность: одна награда один раз одному сообществу (FR-497)
- Записывается: badge_id, community_id, awarded_by, awarded_at, optional reason (FR-498)

**Ошибки:**

- Недостаточно прав
- Награда не найдена или не является платформенной
- Сообщество не найдено
- Награда уже выдана этому сообществу

---

### RevokeBadgeFromUser

**RPC:** `RevokeBadgeFromUser(RevokeBadgeFromUserRequest) returns (RevokeBadgeFromUserResponse)`  
**HTTP:** `DELETE /badges/{badge_id}/award/user/{user_id}`  
**FR:** FR-499, FR-501, FR-503

Отзыв награды у пользователя.

**Request:**

```protobuf
message RevokeBadgeFromUserRequest {
  string badge_id
  string user_id
  optional string reason
}
```

**Response:**

```protobuf
message RevokeBadgeFromUserResponse {
  string message
}
```

**Требования:**

- Требуются те же права что и для выдачи (FR-501):
  - `award_platform_badges` для платформенных наград
  - `award_community_badges` для наград сообщества
- Операция логируется с moderator info, reason, timestamp (FR-503)

**Ошибки:**

- Недостаточно прав
- Награда не выдана этому пользователю

---

### RevokeBadgeFromCommunity

**RPC:** `RevokeBadgeFromCommunity(RevokeBadgeFromCommunityRequest) returns (RevokeBadgeFromCommunityResponse)`  
**HTTP:** `DELETE /badges/{badge_id}/award/community/{community_id}`  
**FR:** FR-500, FR-502, FR-503

Отзыв награды у сообщества.

**Request:**

```protobuf
message RevokeBadgeFromCommunityRequest {
  string badge_id
  string community_id
  optional string reason
}
```

**Response:**

```protobuf
message RevokeBadgeFromCommunityResponse {
  string message
}
```

**Требования:**

- Требуется право `award_platform_badges` (FR-502)
- Операция логируется с moderator info, reason, timestamp (FR-503)

**Ошибки:**

- Недостаточно прав
- Награда не выдана этому сообществу

---

### ListUserBadges

**RPC:** `ListUserBadges(ListUserBadgesRequest) returns (ListUserBadgesResponse)`  
**HTTP:** `GET /users/{user_id}/badges`  
**FR:** FR-504, FR-507, FR-510-511

Получение списка наград пользователя.

**Request:**

```protobuf
message ListUserBadgesRequest {
  string user_id
  string cursor
  int32 limit
}
```

**Response:**

```protobuf
message ListUserBadgesResponse {
  repeated BadgeAward awards
  string next_cursor
  bool has_more
}
```

**Требования:**

- Cursor-based пагинация (FR-504)
- Сортировка по awarded_at в обратном порядке (новые первые) (FR-507)
- Каждая запись включает (FR-510):
  - Полная информация о награде (name, description, icon, rarity)
  - awarded_at timestamp
  - awarded_by info (username модератора)
  - optional reason
- Отображается в профиле пользователя (FR-511)

---

### ListCommunityBadges

**RPC:** `ListCommunityBadges(ListCommunityBadgesRequest) returns (ListCommunityBadgesResponse)`  
**HTTP:** `GET /communities/{community_id}/badges`  
**FR:** FR-505, FR-508, FR-510, FR-512

Получение списка наград сообщества.

**Request:**

```protobuf
message ListCommunityBadgesRequest {
  string community_id
  string cursor
  int32 limit
}
```

**Response:**

```protobuf
message ListCommunityBadgesResponse {
  repeated BadgeAward awards
  string next_cursor
  bool has_more
}
```

**Требования:**

- Cursor-based пагинация (FR-505)
- Сортировка по awarded_at в обратном порядке (новые первые) (FR-508)
- Структура записи аналогична ListUserBadges (FR-510)
- Отображается в профиле сообщества (FR-512)

---

### ListPendingBadges

**RPC:** `ListPendingBadges(ListPendingBadgesRequest) returns (ListPendingBadgesResponse)`  
**HTTP:** `GET /badges/pending`  
**FR:** FR-506, FR-509

Получение списка наград ожидающих модерации (только для модераторов платформы).

**Request:**

```protobuf
message ListPendingBadgesRequest {
  string cursor
  int32 limit
}
```

**Response:**

```protobuf
message ListPendingBadgesResponse {
  repeated Badge badges
  string next_cursor
  bool has_more
}
```

**Требования:**

- Требуется право `approve_community_badges` (FR-509)
- Возвращает только community badges в статусе `PENDING_APPROVAL`
- Cursor-based пагинация

**Ошибки:**

- Недостаточно прав

---

### GetBadge

**RPC:** `GetBadge(GetBadgeRequest) returns (GetBadgeResponse)`  
**HTTP:** `GET /badges/{badge_id}`  
**FR:** FR-513

Получение информации о награде.

**Request:**

```protobuf
message GetBadgeRequest {
  string badge_id
}
```

**Response:**

```protobuf
message GetBadgeResponse {
  Badge badge
  int32 awarded_count   // сколько раз выдана
}
```

**Требования:**

- Доступно всем пользователям
- Возвращает полную информацию о награде
- Включает счетчик выдач

---

### UpdateBadge

**RPC:** `UpdateBadge(UpdateBadgeRequest) returns (UpdateBadgeResponse)`  
**HTTP:** `PUT /badges/{badge_id}`  
**FR:** FR-514-517

Обновление информации о награде.

**Request:**

```protobuf
message UpdateBadgeRequest {
  string badge_id
  optional string name
  optional string description
  optional string icon_url
  optional RarityLevel rarity
}
```

**Response:**

```protobuf
message UpdateBadgeResponse {
  Badge badge
}
```

**Требования:**

**Для платформенных наград:**

- Требуется право `edit_platform_badges` (FR-515)
- Обновляется напрямую

**Для наград сообщества:**

- Требуется право `edit_community_badges` для этого сообщества (FR-516)
- Владелец сообщества всегда имеет это право (FR-522)
- Если награда была одобрена (`ACTIVE`), статус сбрасывается на `PENDING_APPROVAL` (FR-517)
- Требует повторной модерации

**Ошибки:**

- Недостаточно прав
- Награда не найдена
- Невалидные параметры

---

### DeleteBadge

**RPC:** `DeleteBadge(DeleteBadgeRequest) returns (DeleteBadgeResponse)`  
**HTTP:** `DELETE /badges/{badge_id}`  
**FR:** FR-518-520

Удаление награды.

**Request:**

```protobuf
message DeleteBadgeRequest {
  string badge_id
}
```

**Response:**

```protobuf
message DeleteBadgeResponse {
  string message
}
```

**Требования:**

**Для платформенных наград:**

- Требуется право `delete_platform_badges` (FR-520)

**Для наград сообщества:**

- Требуется право `delete_community_badges` для этого сообщества (FR-520)
- Владелец сообщества всегда имеет это право (FR-522)

**Важно:**

- Автоматически отзывает награду у всех пользователей/сообществ кто ее имеет (FR-519)
- Каскадное удаление всех BadgeAward записей
- Необратимая операция

**Ошибки:**

- Недостаточно прав
- Награда не найдена

---

## Workflow модерации наград сообществ

### 1. Создание

```
Владелец/модератор сообщества → CreateCommunityBadge
  ↓
Статус: PENDING_APPROVAL
  ↓
Уведомление модераторам платформы
```

### 2. Модерация

**Вариант A: Одобрение**

```
Модератор платформы → ApproveCommunityBadge
  ↓
Статус: ACTIVE
  ↓
Уведомление владельцу сообщества
  ↓
Награду можно выдавать участникам
```

**Вариант B: Отклонение**

```
Модератор платформы → RejectCommunityBadge + reason
  ↓
Статус: REJECTED
  ↓
Уведомление владельцу с причиной
  ↓
Владелец может отредактировать → UpdateBadge
  ↓
Статус: PENDING_APPROVAL (повторная модерация)
```

### 3. Выдача

```
Модератор сообщества → AwardBadgeToUser
  ↓
Проверки:
  - Награда в статусе ACTIVE
  - Пользователь - участник сообщества
  - Награда еще не выдана этому пользователю
  ↓
BadgeAward создан
  ↓
Отображается в профиле пользователя
```

## Права доступа (Permissions)

### Платформенные права

- **create_platform_badges**: Создание платформенных наград
- **edit_platform_badges**: Редактирование платформенных наград
- **delete_platform_badges**: Удаление платформенных наград
- **award_platform_badges**: Выдача платформенных наград пользователям/сообществам
- **approve_community_badges**: Модерация (одобрение/отклонение) наград сообществ

### Права сообществ

- **create_community_badges**: Создание наград сообщества
- **edit_community_badges**: Редактирование наград сообщества
- **delete_community_badges**: Удаление наград сообщества
- **award_community_badges**: Выдача наград сообщества участникам

### Владелец сообщества

Владелец сообщества **всегда** имеет все badge permissions для своего сообщества (FR-522):

- create_community_badges
- edit_community_badges
- delete_community_badges
- award_community_badges

## Уникальность выдачи

- Одна награда может быть выдана один раз одному пользователю (FR-497)
- Одна награда может быть выдана один раз одному сообществу (FR-497)
- Уникальный constraint: `(badge_id, user_id)` или `(badge_id, community_id)`
- Попытка повторной выдачи возвращает ошибку
- Для повторной выдачи нужно сначала отозвать награду

## Отображение в профилях

### Профиль пользователя (FR-511)

```json
{
  "user": {
    "id": "user_123",
    "username": "john_doe",
    ...
  },
  "badges": [
    {
      "badge": {
        "id": "badge_1",
        "name": "Early Adopter",
        "description": "One of the first 100 users",
        "icon_url": "https://...",
        "rarity": "LEGENDARY"
      },
      "awarded_at": "2025-01-15T10:30:00Z",
      "awarded_by": "admin",
      "reason": "Platform launch participant"
    }
  ]
}
```

### Профиль сообщества (FR-512)

```json
{
  "community": {
    "id": "community_456",
    "name": "Golang Developers",
    ...
  },
  "badges": [
    {
      "badge": {
        "id": "badge_2",
        "name": "Community of the Month",
        "description": "Most active community",
        "icon_url": "https://...",
        "rarity": "EPIC"
      },
      "awarded_at": "2025-02-01T00:00:00Z",
      "awarded_by": "platform_admin",
      "reason": "Exceptional growth and engagement"
    }
  ]
}
```

## Лимиты

- **Количество наград**: НЕТ лимита на создание наград (FR-521)
- **Количество выдач**: Одна награда один раз одному получателю
- **Уровни редкости**: 4 уровня (COMMON, RARE, EPIC, LEGENDARY)

## Логирование

### Операции требующие логирования (FR-503)

- **Отзыв награды**: RevokeBadgeFromUser / RevokeBadgeFromCommunity
  - Moderator info (ID, username)
  - Reason (опционально, но рекомендуется)
  - Timestamp
  - Target (user/community ID)
  - Badge info

### Рекомендуемые логи

- Создание наград
- Модерация (одобрение/отклонение)
- Выдача наград
- Все операции с указанием actor (кто выполнил)

## Edge Cases

### Удаление награды (FR-519)

```
DeleteBadge → badge_id
  ↓
Автоматически отзываются все выдачи:
  - BadgeAward записи удаляются
  - Пользователи/сообщества теряют награду
  ↓
Необратимо
```

### Редактирование одобренной награды сообщества (FR-517)

```
UpdateBadge → approved community badge
  ↓
Статус сбрасывается: ACTIVE → PENDING_APPROVAL
  ↓
Требуется повторная модерация
  ↓
Выданные награды остаются, но новые выдать нельзя до одобрения
```

### Пользователь покидает сообщество

- Награды сообщества остаются у пользователя
- Но модераторы сообщества могут отозвать награды у бывших участников
- Рекомендуется автоматический отзыв при выходе (опционально)

### Бан пользователя

- Забаненный пользователь сохраняет свои награды (видны в профиле)
- Модераторы могут отозвать награды у забаненных пользователей

## Реализация

### База данных

**Таблица: badges**

```sql
CREATE TABLE badges (
  id UUID PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  description VARCHAR(500) NOT NULL,
  icon_url TEXT NOT NULL,
  rarity VARCHAR(20) NOT NULL CHECK (rarity IN ('common', 'rare', 'epic', 'legendary')),
  type VARCHAR(20) NOT NULL CHECK (type IN ('platform', 'community')),
  status VARCHAR(20) NOT NULL CHECK (status IN ('active', 'pending_approval', 'rejected')),
  community_id UUID REFERENCES communities(id) ON DELETE CASCADE, -- nullable
  created_by UUID NOT NULL REFERENCES users(id),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_badges_status ON badges(status) WHERE status = 'pending_approval';
CREATE INDEX idx_badges_community ON badges(community_id) WHERE community_id IS NOT NULL;
CREATE INDEX idx_badges_type ON badges(type);
```

**Таблица: badge_awards**

```sql
CREATE TABLE badge_awards (
  id UUID PRIMARY KEY,
  badge_id UUID NOT NULL REFERENCES badges(id) ON DELETE CASCADE,
  recipient_id UUID NOT NULL,  -- user_id или community_id
  recipient_type VARCHAR(20) NOT NULL CHECK (recipient_type IN ('user', 'community')),
  awarded_by UUID NOT NULL REFERENCES users(id),
  reason TEXT,
  awarded_at TIMESTAMP NOT NULL DEFAULT NOW(),

  -- Уникальность: одна награда один раз одному получателю
  UNIQUE(badge_id, recipient_id, recipient_type)
);

CREATE INDEX idx_badge_awards_recipient ON badge_awards(recipient_id, recipient_type);
CREATE INDEX idx_badge_awards_badge ON badge_awards(badge_id);
CREATE INDEX idx_badge_awards_awarded_at ON badge_awards(awarded_at DESC);
```

### Кэширование

- Списки наград пользователей: кэш на 5-10 минут
- Списки наград сообществ: кэш на 5-10 минут
- Информация о награде: кэш на 1 час
- Инвалидация при любых изменениях

### Уведомления

- Создание community badge → уведомление модераторам платформы
- Одобрение → уведомление владельцу сообщества
- Отклонение → уведомление владельцу с причиной
- Выдача награды → уведомление получателю (опционально)
