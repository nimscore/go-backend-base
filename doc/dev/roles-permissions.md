# Система ролей и разрешений

## Обзор

Система ролей и разрешений обеспечивает гранулярный контроль доступа с поддержкой платформенных и сообществен-специфичных ролей, автоматической роли @everyone и real-time обновлениями разрешений.

## gRPC Services

- **RoleService** - управление ролями (proto/role.proto)
- **PermissionService** - проверка и отслеживание разрешений (proto/permission.proto)

## Сущности

### RoleType

```protobuf
enum RoleType {
  ROLE_TYPE_UNSPECIFIED = 0
  ROLE_TYPE_PLATFORM    = 1
  ROLE_TYPE_COMMUNITY   = 2
}
```

### Permissions

```protobuf
message Permissions {
  // Moderation (FR-127)
  bool ban_users
  bool mute_users
  bool delete_any_post
  bool delete_any_comment
  bool unpublish_post
  bool view_moderation_logs

  // Content (FR-128)
  bool create_post
  bool edit_own_post
  bool delete_own_post
  bool create_comment
  bool edit_own_comment
  bool delete_own_comment
  bool like_content
  bool bookmark_content

  // Community (FR-129)
  bool create_community
  bool edit_community_settings
  bool delete_community
  bool transfer_community_ownership
  bool manage_community_members
  bool assign_community_roles
  bool create_community_roles
  bool edit_community_roles
  bool delete_community_roles

  // Platform (FR-130)
  bool edit_platform_settings
  bool transfer_platform_ownership
  bool manage_platform_users
  bool assign_platform_roles
  bool create_platform_roles
  bool edit_platform_roles
  bool delete_platform_roles
  bool view_all_communities
  bool view_analytics

  // Reports (FR-131)
  bool report_content
  bool view_reports
  bool resolve_reports
  bool dismiss_reports

  // Advanced (FR-132)
  bool pin_post
  bool unpin_post
  bool lock_thread
  bool unlock_thread
  bool feature_post
  bool edit_any_post
  bool edit_any_comment
}
```

### Role

```protobuf
message Role {
  string id
  string name
  string color                     // hex color code
  RoleType type
  optional string community_id     // только для community roles
  Permissions permissions
  int32 member_count
  bool is_everyone
  google.protobuf.Timestamp created_at
}
```

### UserRoleInfo

```protobuf
message UserRoleInfo {
  string role_id
  string role_name
  string role_color
  RoleType role_type
}
```

## RoleService Endpoints

### CreatePlatformRole

**RPC:** `CreatePlatformRole(CreatePlatformRoleRequest) returns (CreatePlatformRoleResponse)`  
**HTTP:** `POST /roles/platform`  
**FR:** FR-036, FR-355, FR-357, FR-359, FR-361, FR-363-367

Создание платформенной роли.

**Request:**

```protobuf
message CreatePlatformRoleRequest {
  string name              // 1-50 символов, не "@everyone"
  string color             // hex color code
  Permissions permissions
}
```

**Response:**

```protobuf
message CreatePlatformRoleResponse {
  Role role
}
```

**Требования:**

- Требуется create_platform_roles permission (FR-359)
- Имя 1-50 символов (FR-364)
- Имя уникально среди платформенных ролей (FR-361)
- Имя НЕ может быть "@everyone" (FR-363, FR-100)
- Color валидный hex код (FR-365)
- Permissions содержат только валидные флаги (FR-366)
- Возврат роли с member_count=0 (FR-367)

---

### CreateCommunityRole

**RPC:** `CreateCommunityRole(CreateCommunityRoleRequest) returns (CreateCommunityRoleResponse)`  
**HTTP:** `POST /communities/{community_id}/roles`  
**FR:** FR-037, FR-356, FR-358, FR-360, FR-362-367

Создание роли сообщества.

**Request:**

```protobuf
message CreateCommunityRoleRequest {
  string community_id
  string name
  string color
  Permissions permissions
}
```

**Response:**

```protobuf
message CreateCommunityRoleResponse {
  Role role
}
```

**Требования:**

- Требуется create_community_roles permission для сообщества (FR-360)
- Имя уникально среди ролей сообщества (FR-362)
- Остальное аналогично CreatePlatformRole

---

### GetRole

**RPC:** `GetRole(GetRoleRequest) returns (GetRoleResponse)`  
**HTTP:** `GET /roles/{role_id}`  
**FR:** FR-238, FR-245

Получение информации о роли.

**Response:**

```protobuf
message GetRoleResponse {
  Role role  // с полной информацией (FR-245)
}
```

---

### UpdateRole

**RPC:** `UpdateRole(UpdateRoleRequest) returns (UpdateRoleResponse)`  
**HTTP:** `PATCH /roles/{role_id}`  
**FR:** FR-239, FR-246

Обновление роли.

**Request:**

```protobuf
message UpdateRoleRequest {
  string role_id
  optional string name             // нельзя переименовать @everyone
  optional string color
  optional Permissions permissions
}
```

**Требования:**

- Роль @everyone НЕ может быть переименована (FR-101, FR-246)
- Разрешения @everyone могут редактироваться владельцами (FR-102, FR-103)

---

### DeleteRole

**RPC:** `DeleteRole(DeleteRoleRequest) returns (DeleteRoleResponse)`  
**HTTP:** `DELETE /roles/{role_id}`  
**FR:** FR-240, FR-247, FR-248

Удаление роли.

**Требования:**

- Роль @everyone НЕ может быть удалена (FR-100, FR-247)
- Автоматическое удаление роли у всех пользователей (FR-248)
- Триггер permissions stream update

---

### ListPlatformRoles

**RPC:** `ListPlatformRoles(ListPlatformRolesRequest) returns (ListPlatformRolesResponse)`  
**HTTP:** `GET /roles/platform`  
**FR:** FR-241, FR-251

Список всех платформенных ролей.

**Требования:**

- Cursor-based пагинация
- Сортировка по member_count в обратном порядке (FR-251)

---

### ListCommunityRoles

**RPC:** `ListCommunityRoles(ListCommunityRolesRequest) returns (ListCommunityRolesResponse)`  
**HTTP:** `GET /communities/{community_id}/roles`  
**FR:** FR-242, FR-252

Список ролей сообщества.

**Требования:**

- Cursor-based пагинация
- Сортировка по created_at в обратном порядке (FR-252)

---

### AssignRole

**RPC:** `AssignRole(AssignRoleRequest) returns (AssignRoleResponse)`  
**HTTP:** `POST /roles/{role_id}/assign`  
**FR:** FR-243, FR-249, FR-253

Назначение роли пользователю.

**Request:**

```protobuf
message AssignRoleRequest {
  string role_id
  string user_id
}
```

**Требования:**

- Требуется assign_platform_roles или assign_community_roles (FR-249)
- Триггер permissions stream update (FR-253)

---

### RemoveRole

**RPC:** `RemoveRole(RemoveRoleRequest) returns (RemoveRoleResponse)`  
**HTTP:** `POST /roles/{role_id}/remove`  
**FR:** FR-244, FR-250, FR-253

Удаление роли у пользователя.

**Требования:**

- Соответствующие права на удаление (FR-250)
- Триггер permissions stream update (FR-253)

---

## PermissionService Endpoints

### GetUserPermissions

**RPC:** `GetUserPermissions(GetUserPermissionsRequest) returns (GetUserPermissionsResponse)`  
**HTTP:** `GET /permissions/platform` или `GET /users/{user_id}/permissions/platform`  
**FR:** FR-115, FR-117-121

Получение платформенных разрешений пользователя.

**Request:**

```protobuf
message GetUserPermissionsRequest {
  optional string user_id  // если пусто - текущий пользователь
}
```

**Response:**

```protobuf
message UserPermissionsInfo {
  Permissions calculated_permissions      // union всех ролей
  repeated UserRoleInfo roles
  google.protobuf.Timestamp calculated_at
}
```

**Требования:**

- Расчет разрешений как union всех ролей (FR-119)
- Возврат списка ролей с именем и цветом (FR-118)
- Пользователь может проверять свои разрешения (FR-120)
- Модератор может проверять любые разрешения (FR-121)

---

### GetCommunityPermissions

**RPC:** `GetCommunityPermissions(GetCommunityPermissionsRequest) returns (GetCommunityPermissionsResponse)`  
**HTTP:** `GET /permissions/communities/{community_id}`  
**FR:** FR-116-121

Получение разрешений пользователя в сообществе.

**Request:**

```protobuf
message GetCommunityPermissionsRequest {
  string community_id
  optional string user_id
}
```

**Требования:**

- Комбинация платформенных и community-специфичных ролей
- Остальное аналогично GetUserPermissions

---

### StreamPermissions

**RPC:** `StreamPermissions(StreamPermissionsRequest) returns (stream PermissionChangeEvent)`  
**gRPC Streaming:** Server-Side Streaming  
**FR:** FR-122-125

Real-time обновления разрешений пользователя.

**Request:**

```protobuf
message StreamPermissionsRequest {
  optional string community_id  // если пусто - платформенные
}
```

**Response Stream:**

```protobuf
message PermissionChangeEvent {
  PermissionChangeType change_type
  UserPermissionsInfo updated_permissions
  google.protobuf.Timestamp timestamp
}

enum PermissionChangeType {
  PERMISSION_CHANGE_TYPE_ROLE_ASSIGNED
  PERMISSION_CHANGE_TYPE_ROLE_REMOVED
  PERMISSION_CHANGE_TYPE_ROLE_EDITED
  PERMISSION_CHANGE_TYPE_COMMUNITY_JOINED
  PERMISSION_CHANGE_TYPE_COMMUNITY_LEFT
}
```

**Требования:**

- Уведомление при назначении/удалении роли (FR-123)
- Уведомление при редактировании разрешений роли (FR-123)
- Уведомление при вступлении/выходе из сообщества (FR-125)
- Включение обновленных разрешений и ролей (FR-124)
- Доставка обновлений в течение 1 секунды (SC-011)

---

## Специальная роль @everyone

### Платформенная @everyone

**Создание:**

- Автоматически при инициализации платформы (FR-093)

**Назначение:**

- Автоматически всем зарегистрированным пользователям (FR-094)

**Разрешения по умолчанию:**

- report_content (FR-109)

**Управление:**

- Не может быть удалена (FR-100)
- Не может быть переименована (FR-101)
- Разрешения могут редактироваться платформенным владельцем (FR-102)
- Не удаляется при активном аккаунте (FR-105)

### Сообщественная @everyone

**Создание:**

- Автоматически при создании сообщества (FR-096)

**Назначение:**

- Создателю сообщества (FR-098)
- Всем вступающим в сообщество (FR-099)

**Разрешения по умолчанию:**

- report_content в рамках сообщества (FR-110)

**Управление:**

- Не может быть удалена (FR-100)
- Не может быть переименована (FR-101)
- Разрешения могут редактироваться владельцем сообщества (FR-103)
- Автоматически удаляется при выходе из сообщества (FR-104)

---

## Расчет эффективных разрешений

### Алгоритм

```
effective_permissions = union(permissions from all user roles)
```

Если пользователь имеет:

- Platform Role A: {create_post, edit_own_post}
- Community Role B: {delete_any_post}
- @everyone: {report_content}

Эффективные разрешения:

```
{create_post, edit_own_post, delete_any_post, report_content}
```

### Особенности

- Нет иерархии или отрицательных разрешений
- Любая роль с разрешением дает его пользователю
- Без лимита на количество ролей (FR-113)
- Эффективный расчет для сотен ролей (FR-248)

### Real-time расчет

- Разрешения рассчитываются в реальном времени (FR-139)
- НЕ кешируются для целей авторизации
- Клиентские разрешения только для UI/UX (FR-136, FR-182)

---

## Валидация на сервере

### Принцип Zero Trust

- Сервер независимо валидирует разрешения для КАЖДОЙ операции (FR-135)
- Клиент НЕ должен доверяться для security (FR-137)
- Клиентские разрешения только для show/hide кнопок (FR-136)

### Реализация

```go
// Псевдокод валидации
func ValidatePermission(ctx, action) error {
    userID := getUserFromContext(ctx)
    communityID := getCommunityFromContext(ctx)

    permissions := calculateEffectivePermissions(userID, communityID)

    if !permissions.Has(action) {
        return grpc.PermissionDenied // код 7
    }

    return nil
}
```

### Логирование

Все попытки без прав логируются (FR-140):

- user_id
- action attempted
- required permission
- timestamp

---

## Категории разрешений

### Moderation (6 флагов) - FR-127

Управление контентом и пользователями на платформе.

### Content (8 флагов) - FR-128

Создание и управление собственным контентом.

### Community (9 флагов) - FR-129

Управление сообществами и их участниками.

### Platform (9 флагов) - FR-130

Управление глобальными настройками платформы.

### Reports (4 флага) - FR-131

Система жалоб и модерации контента.

### Advanced (7 флагов) - FR-132

Расширенные возможности модерации контента.

---

## Владение платформой

### Первый пользователь

- Первый зарегистрированный становится владельцем (FR-095)
- Записывается в platform settings

### Передача владения

- Владелец может передать права верифицированному пользователю (FR-106)
- Требуется подтверждение от нового владельца (FR-108)
- Старый владелец становится обычным пользователем (FR-107)

### Права владельца

- Полный доступ ко всем настройкам платформы
- Может создавать и назначать платформенные роли
- Может редактировать платформенную @everyone

---

## Производительность

### Масштабирование

- Система должна эффективно работать с неограниченным количеством ролей (FR-113-114)
- Оптимизация для пользователей с сотнями ролей (FR-248)

### Рекомендации

- Индексы на role assignments
- Кеширование данных ролей (не разрешений!)
- Эффективные JOIN запросы
- Bitset для permissions в памяти

---

## Ошибки

### gRPC коды

- **PermissionDenied (7)**: Недостаточно прав (FR-138)
- **Unauthenticated (16)**: Невалидный/отсутствующий JWT
- **InvalidArgument (3)**: Невалидные параметры роли

### Обработка

- Четкие сообщения об ошибках
- Указание какое именно разрешение требуется
- Логирование для аудита
