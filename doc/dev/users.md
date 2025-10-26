# Управление пользователями и профилями

## Обзор

UserService предоставляет функционал для управления профилями пользователей, подписками, просмотра статистики и контента пользователей.

## gRPC Service: UserService

Proto файл: `proto/user.proto`

## Сущности

### UserProfile

```protobuf
message UserProfile {
  string id
  string username
  string avatar_url
  string banner_url
  string description
  int32 reputation
  int32 follower_count
  int32 following_count
  bool is_banned
  bool is_following         // следует ли текущий пользователь за этим
  google.protobuf.Timestamp created_at
}
```

### CurrentUserProfile

```protobuf
message CurrentUserProfile {
  string id
  string username
  string email
  bool email_verified
  string avatar_url
  string banner_url
  string description
  int32 reputation
  int32 follower_count
  int32 following_count
  int32 joined_communities_count
  int32 active_sessions_count
  google.protobuf.Timestamp created_at
}
```

### UserStatistics

```protobuf
message UserStatistics {
  int32 total_posts
  int32 total_comments
  int32 total_likes_received
  int32 post_likes
  int32 comment_likes
  int32 communities_created
  int32 communities_joined
}
```

## Endpoints

### Get

**RPC:** `Get(GetRequest) returns (GetResponse)`  
**HTTP:** `GET /users/{user_id}`  
**FR:** FR-303, FR-307, FR-313, FR-314

Получение публичного профиля любого пользователя.

**Request:**

```protobuf
message GetRequest {
  string user_id  // или username
}
```

**Response:**

```protobuf
message GetResponse {
  UserProfile user
}
```

**Требования:**

- Поддержка поиска по user_id или username
- Возврат публичного профиля (FR-307):
  - username, avatar, banner, description
  - reputation, follower_count, following_count
  - created_at, banned status
- Поле is_following указывает следует ли текущий пользователь (требует auth)
- Для забаненных пользователей включается banned status, скрываются приватные данные (FR-314)

**Ошибки:**

- Пользователь не найден (FR-313)
- Аккаунт удален

---

### GetCurrent

**RPC:** `GetCurrent(GetCurrentRequest) returns (GetCurrentResponse)`  
**HTTP:** `GET /users/me`  
**FR:** FR-304, FR-308

Получение полного профиля текущего аутентифицированного пользователя.

**Request:**

```protobuf
message GetCurrentRequest {}
```

**Response:**

```protobuf
message GetCurrentResponse {
  CurrentUserProfile user
}
```

**Требования:**

- Возврат полного профиля включая приватные данные (FR-308):
  - Все поля из GetUser
  - email, email_verified
  - joined_communities_count
  - active_sessions_count
- Требуется аутентификация

---

### UpdateProfile

**RPC:** `UpdateProfile(UpdateProfileRequest) returns (UpdateProfileResponse)`  
**HTTP:** `PATCH /users/me`  
**FR:** FR-054, FR-305, FR-309, FR-310, FR-311

Обновление профиля текущего пользователя.

**Request:**

```protobuf
message UpdateProfileRequest {
  optional string avatar_url   // S3 URL после загрузки
  optional string banner_url   // S3 URL после загрузки
  optional string description  // max 500 символов
}
```

**Response:**

```protobuf
message UpdateProfileResponse {
  CurrentUserProfile user
}
```

**Требования:**

- Все поля опциональны
- avatar_url и banner_url должны быть валидными S3 URLs из предварительной загрузки (FR-309)
- description максимум 500 символов (FR-310)
- НЕ может изменять (FR-311):
  - username
  - email
  - reputation
  - created_at
- Требуется аутентификация

**Ошибки:**

- description превышает 500 символов
- Невалидный S3 URL

---

### GetStatistics

**RPC:** `GetStatistics(GetStatisticsRequest) returns (GetStatisticsResponse)`  
**HTTP:** `GET /users/{user_id}/statistics`  
**FR:** FR-306, FR-312

Получение статистики активности пользователя.

**Request:**

```protobuf
message GetStatisticsRequest {
  string user_id
}
```

**Response:**

```protobuf
message GetStatisticsResponse {
  UserStatistics statistics
}
```

**Требования:**

- Возврат полной статистики (FR-312):
  - total_posts: количество всех постов
  - total_comments: количество всех комментариев
  - total_likes_received: сумма лайков на постах и комментариях
  - post_likes: лайки только на постах
  - comment_likes: лайки только на комментариях
  - communities_created: созданных сообществ
  - communities_joined: вступленных сообществ

---

### ListCommunities

**RPC:** `ListCommunities(ListCommunitiesRequest) returns (ListCommunitiesResponse)`  
**HTTP:** `GET /users/{user_id}/communities`  
**FR:** FR-227, FR-232

Получение списка сообществ пользователя.

**Request:**

```protobuf
message ListCommunitiesRequest {
  string user_id
  string cursor
  int32 limit
}
```

**Response:**

```protobuf
message ListCommunitiesResponse {
  repeated Community communities
  string next_cursor
  bool has_more
}
```

**Требования:**

- Cursor-based пагинация
- Сортировка по дате вступления в обратном порядке (новые первые) (FR-232)
- Возврат только сообществ где пользователь является участником

---

### ListPosts

**RPC:** `ListPosts(ListPostsRequest) returns (ListPostsResponse)`  
**HTTP:** `GET /users/{user_id}/posts`  
**FR:** FR-212, FR-214, FR-215

Получение списка постов пользователя.

**Request:**

```protobuf
message ListPostsRequest {
  string user_id
  optional PostStatus status_filter  // all, draft, published
  string cursor
  int32 limit
}
```

**Response:**

```protobuf
message ListPostsResponse {
  repeated Post posts
  string next_cursor
  bool has_more
}
```

**Требования:**

- Cursor-based пагинация
- Поддержка фильтрации по статусу (FR-214):
  - all: все посты
  - draft: только черновики
  - published: только опубликованные
- Сортировка по дате создания в обратном порядке (новые первые) (FR-215)
- Черновики видны только автору

---

### ListComments

**RPC:** `ListComments(ListCommentsRequest) returns (ListCommentsResponse)`  
**HTTP:** `GET /users/{user_id}/comments`  
**FR:** FR-218, FR-222, FR-223

Получение списка комментариев пользователя.

**Request:**

```protobuf
message ListCommentsRequest {
  string user_id
  string cursor
  int32 limit
}
```

**Response:**

```protobuf
message CommentWithPostInfo {
  Comment comment
  string post_id
  string post_title
}

message ListCommentsResponse {
  repeated CommentWithPostInfo comments
  string next_cursor
  bool has_more
}
```

**Требования:**

- Cursor-based пагинация
- Сортировка по дате создания в обратном порядке (новые первые) (FR-222)
- Включение контекста поста для каждого комментария (FR-223):
  - post_id
  - post_title

---

## Система подписок (Follow System)

### Follow

**RPC:** `Follow(FollowRequest) returns (FollowResponse)`  
**HTTP:** `POST /users/{user_id}/follow`  
**FR:** FR-186, FR-187, FR-191, FR-194, FR-202, FR-203

Подписка на пользователя.

**Request:**

```protobuf
message FollowRequest {
  string user_id  // пользователь на которого подписываемся
}
```

**Response:**

```protobuf
message FollowResponse {
  string message
}
```

**Требования:**

- Односторонняя подписка (Twitter-модель) (FR-186)
- Идемпотентность: повторная подписка возвращает success (FR-191)
- Запрет подписки на самого себя (FR-194)
- Лимит 5000 подписок на пользователя (FR-202)
- Обновление счетчиков follower_count и following_count (FR-193)
- Требуется аутентификация

**Ошибки:**

- Попытка подписаться на самого себя
- Достигнут лимит 5000 подписок (FR-203)
- Целевой пользователь не найден

---

### Unfollow

**RPC:** `Unfollow(UnfollowRequest) returns (UnfollowResponse)`  
**HTTP:** `DELETE /users/{user_id}/follow`  
**FR:** FR-188, FR-192, FR-204

Отписка от пользователя.

**Request:**

```protobuf
message UnfollowRequest {
  string user_id
}
```

**Response:**

```protobuf
message UnfollowResponse {
  string message
}
```

**Требования:**

- Идемпотентность: отписка от неподписанного возвращает success (FR-192)
- Работает даже при достижении лимита подписок (FR-204)
- Обновление счетчиков follower_count и following_count
- Требуется аутентификация

---

### ListFollowers

**RPC:** `ListFollowers(ListFollowersRequest) returns (ListFollowersResponse)`  
**HTTP:** `GET /users/{user_id}/followers`  
**FR:** FR-189, FR-193

Получение списка подписчиков пользователя.

**Request:**

```protobuf
message ListFollowersRequest {
  string user_id
  string cursor
  int32 limit
}
```

**Response:**

```protobuf
message ListFollowersResponse {
  repeated UserProfile users
  string next_cursor
  bool has_more
}
```

**Требования:**

- Cursor-based пагинация
- Возврат UserProfile для каждого подписчика
- Поле is_following показывает подписан ли текущий пользователь на этого подписчика

---

### ListFollowing

**RPC:** `ListFollowing(ListFollowingRequest) returns (ListFollowingResponse)`  
**HTTP:** `GET /users/{user_id}/following`  
**FR:** FR-190, FR-193

Получение списка пользователей на которых подписан пользователь.

**Request:**

```protobuf
message ListFollowingRequest {
  string user_id
  string cursor
  int32 limit
}
```

**Response:**

```protobuf
message ListFollowingResponse {
  repeated UserProfile users
  string next_cursor
  bool has_more
}
```

**Требования:**

- Cursor-based пагинация
- Возврат UserProfile для каждого пользователя
- Поле is_following показывает подписан ли текущий пользователь

---

## Репутация пользователя

### Расчет

Репутация = сумма всех лайков на постах и комментариях пользователя (FR-053, FR-312)

### Компоненты

- **post_likes**: количество лайков на всех постах
- **comment_likes**: количество лайков на всех комментариях
- **reputation**: post_likes + comment_likes

### Обновление

- Автоматическое обновление при получении/удалении лайка (FR-178, FR-397)
- Real-time отражение в профиле
- Используется для ранжирования и отображения авторитетности

## Забаненные пользователи

### Отображение профиля

- Публичный профиль показывает banned status (FR-062, FR-314)
- Email и другие приватные данные скрыты
- Существующий контент остается видимым (FR-060)

### Ограничения

Забаненные пользователи НЕ могут (FR-061, FR-063):

- Создавать новые посты
- Создавать комментарии
- Создавать сообщества
- Лайкать контент
- Добавлять закладки
- Любые другие взаимодействия

Забаненные пользователи МОГУТ:

- Просматривать контент (если не вышли из системы)

## Лимиты

### Подписки

- **Максимум:** 5000 подписок на пользователя (FR-202)
- **Цель:** Предотвращение злоупотреблений и проблем с производительностью
- **Обработка:** Ошибка при попытке превысить лимит (FR-203)
- **Отписка:** Разрешена даже при достижении лимита (FR-204)

### Производительность персонализированной ленты

- Оптимизирована для до 5000 подписок (FR-250)
- Пагинация обеспечивает производительность
- Дедупликация постов по post_id

## Приватность

### Публичные данные

Доступны всем пользователям:

- Username
- Avatar и banner
- Description
- Reputation
- Счетчики (followers, following)
- Created_at
- Banned status

### Приватные данные

Доступны только владельцу профиля:

- Email
- Email verification status
- Количество вступленных сообществ
- Количество активных сессий
- Черновики постов

## Удаление аккаунта

### Cascade эффекты

При удалении пользователя:

- Удаляются все подписки (follow relationships) (FR-249)
- Удаляется платформенная роль @everyone (FR-105)
- Контент может оставаться с маркировкой "deleted user"

### Передача владения

Перед удалением пользователь должен:

- Передать владение всеми созданными сообществами
- Передать владение платформой (если является владельцем)
