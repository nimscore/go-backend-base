# Управление платформой

## Обзор

PlatformService предоставляет функционал для управления глобальными настройками платформы, статистикой и передачей владения.

## gRPC Service: PlatformService

Proto файл: `proto/platform.proto`

## Сущности

### PlatformSettings

```protobuf
message AutomaticBadgeSetting {
  string badge_id
  bool enabled
}

message PlatformSettings {
  string name
  string description
  string rules
  string logo_url
  string banner_url
  string auth_banner_url
  string owner_id
  string owner_username
  google.protobuf.Timestamp created_at
  repeated AutomaticBadgeSetting badge_settings  // FR-539: настройки автоматических наград
}
```

**Automatic Badge Settings** (FR-539):

- Список всех автоматических наград платформы с их статусом (enabled/disabled)
- Владелец платформы может включать/выключать каждую награду индивидуально
- По умолчанию все автоматические награды включены при инициализации

### PlatformStatistics

```protobuf
message PlatformStatistics {
  int32 total_users
  int32 verified_users
  int32 total_communities
  int32 total_posts
  int32 total_comments

  int32 pending_reports
  int32 resolved_reports
  int32 dismissed_reports

  int32 active_users_24h
  int32 active_users_7d
  int32 active_users_30d

  google.protobuf.Timestamp calculated_at
}
```

## Endpoints

### GetSettings

**RPC:** `GetSettings(GetSettingsRequest) returns (GetSettingsResponse)`  
**HTTP:** `GET /platform/settings`  
**FR:** FR-315, FR-317-319

Получение настроек платформы.

**Request:**

```protobuf
message GetSettingsRequest {}
```

**Response:**

```protobuf
message GetSettingsResponse {
  PlatformSettings settings
}
```

**Требования:**

- Доступен всем пользователям включая анонимных (FR-319)
- Возврат полной информации (FR-318):
  - platform name, description, rules
  - logo URL, banner URL, auth banner URL
  - owner info (id, username)
  - created_at timestamp
  - automatic badge settings (FR-539): список badge_id с enabled/disabled статусом

---

### UpdateSettings

**RPC:** `UpdateSettings(UpdateSettingsRequest) returns (UpdateSettingsResponse)`  
**HTTP:** `PATCH /platform/settings`  
**FR:** FR-050-051, FR-316, FR-320-323

Обновление настроек платформы.

**Request:**

```protobuf
message UpdateSettingsRequest {
  optional string name                                 // 3-100 символов
  optional string description                          // max 1000 символов
  optional string rules
  optional string logo_url
  optional string banner_url
  optional string auth_banner_url
  repeated AutomaticBadgeSetting automatic_badge_settings  // FR-540: обновление настроек наград
}
```

**Response:**

```protobuf
message UpdateSettingsResponse {
  PlatformSettings settings
}
```

**Требования:**

- Требуется быть platform owner (FR-320)
- Все поля опциональны (FR-321)
- Name 3-100 символов если указано (FR-322)
- Description максимум 1000 символов если указано (FR-323)
- URLs должны быть валидными S3 URLs из MediaService
- Automatic badge settings (FR-540):
  - Можно обновить enabled/disabled статус для любой автоматической награды
  - Отключенные награды пропускаются daily cron job (FR-535)
  - Уже выданные награды остаются у пользователей при отключении (FR-536)
- Возврат обновленных настроек

**Ошибки:**

- Только platform owner может обновлять
- Name невалидной длины
- Description превышает лимит
- Невалидный URL

---

### GetStatistics

**RPC:** `GetStatistics(GetStatisticsRequest) returns (GetStatisticsResponse)`  
**HTTP:** `GET /platform/statistics`  
**FR:** FR-317, FR-324-326

Получение статистики платформы.

**Request:**

```protobuf
message GetStatisticsRequest {}
```

**Response:**

```protobuf
message GetStatisticsResponse {
  PlatformStatistics statistics
}
```

**Требования:**

- Требуется view_analytics permission (FR-325)
- Возврат полной статистики (FR-324):

**Пользователи:**

- total_users: все зарегистрированные
- verified_users: с верифицированным email

**Контент:**

- total_communities: все сообщества
- total_posts: все посты
- total_comments: все комментарии

**Жалобы:**

- pending_reports: ожидающие рассмотрения
- resolved_reports: разрешенные
- dismissed_reports: отклоненные

**Активность:**

- active_users_24h: активных за последние 24 часа
- active_users_7d: за последние 7 дней
- active_users_30d: за последние 30 дней

**Расчет активности (FR-326):**
На основе last_activity или last login timestamp.

---

### TransferOwnership

**RPC:** `TransferOwnership(TransferOwnershipRequest) returns (TransferOwnershipResponse)`  
**HTTP:** `POST /platform/transfer-ownership`  
**FR:** FR-106-108

Инициация передачи владения платформой.

**Request:**

```protobuf
message TransferOwnershipRequest {
  string new_owner_id
}
```

**Response:**

```protobuf
message TransferOwnershipResponse {
  string message
}
```

**Требования:**

- Требуется быть текущим platform owner (FR-106)
- new_owner_id должен быть верифицированным пользователем (FR-106)
- Требуется подтверждение от нового владельца (FR-108)
- Генерация токена подтверждения
- Отправка уведомления новому владельцу
- Возврат сообщения о необходимости подтверждения

**Ошибки:**

- Только owner может передать владение
- Новый owner не верифицирован
- Новый owner не найден
- Попытка передать самому себе

---

### ConfirmOwnership

**RPC:** `ConfirmOwnership(ConfirmOwnershipRequest) returns (ConfirmOwnershipResponse)`  
**HTTP:** `POST /platform/confirm-ownership`  
**FR:** FR-108, FR-441-445

Подтверждение получения владения платформой.

**Request:**

```protobuf
message ConfirmOwnershipRequest {
  string token  // токен из уведомления
}
```

**Response:**

```protobuf
message ConfirmOwnershipResponse {
  string message
  PlatformSettings settings
}
```

**Требования:**

- Токен должен быть валидным (FR-443)
- Рекомендуется 24-часовое окно действия токена (FR-443)
- Атомарная передача владения (FR-444):
  - owner_id устанавливается в нового владельца
  - owner_username обновляется
  - Старый владелец становится обычным пользователем (FR-107)
- Возврат обновленных platform settings (FR-445)

**Ошибки:**

- Токен недействителен
- Токен истек
- Передача уже подтверждена или отменена

---

## Инициализация платформы

### Первый запуск

При первом запуске платформы:

1. **Seed @everyone роли (FR-093)**

   - Создается платформенная роль @everyone
   - Default permissions: report_content (FR-109)

2. **Первый пользователь (FR-095)**

   - Первый зарегистрированный пользователь
   - Автоматически становится platform owner
   - Записывается в PlatformSettings

3. **Default настройки**
   - Platform name: "Multiblog Platform"
   - Description: пустая или default
   - Rules: пустые или template
   - created_at: текущий timestamp

---

## Владение платформой

### Права владельца (FR-050-051)

Platform owner может:

- Обновлять название платформы
- Обновлять описание и правила
- Загружать логотип платформы
- Загружать баннеры (основной и auth page)
- Создавать и управлять платформенными ролями
- Редактировать платформенную роль @everyone
- Передать владение другому пользователю
- Доступ ко всей статистике

### Ограничения

- Только один owner в любой момент времени
- Owner должен быть верифицирован
- Owner не может удалить собственный аккаунт без передачи владения

---

## Передача владения

### Процесс (FR-106-108)

**1. Инициация:**

- Текущий owner вызывает TransferOwnership
- Указывает нового владельца (верифицированного)

**2. Подтверждение:**

- Генерируется secure токен
- Отправляется уведомление новому владельцу
- Токен действителен 24 часа (рекомендуется)

**3. Финализация:**

- Новый владелец вызывает ConfirmOwnership с токеном
- Атомарная передача прав
- Старый owner теряет привилегии

**4. Результат:**

- Новый owner в PlatformSettings (FR-107)
- Старый owner становится обычным пользователем
- Все платформенные роли остаются

### Безопасность

- Двухфакторное подтверждение (токен)
- Защита от случайной передачи
- Возможность отмены до подтверждения (опционально)

---

## Брендинг платформы

### Логотип

**Использование:**

- Навигационное меню
- Главная страница
- Email уведомления

**Загрузка:**
MediaService.Upload с MEDIA_TYPE_PLATFORM_LOGO

**Требования:**

- Только platform owner может загружать
- Рекомендуется квадратное изображение
- Минимум 512x512px

### Баннер

**banner_url:**

- Главная страница платформы
- Marketing материалы

**auth_banner_url:**

- Страница регистрации/входа
- Приветственная страница

**Загрузка:**
MediaService.Upload с MEDIA_TYPE_PLATFORM_BANNER

---

## Правила платформы

### rules поле

**Назначение:**

- Кодекс поведения
- Запрещенный контент
- Последствия нарушений

**Формат:**

- Markdown рекомендуется
- Структурированный список
- Ссылки на детальные политики

**Отображение:**

- Отдельная страница /rules
- Ссылки при регистрации
- В footer сайта

---

## Статистика

### Real-time vs cached

**Real-time (дорогие операции):**

- Пересчитываются по запросу
- Или: кешируются с коротким TTL (5-10 минут)

**Cached (рекомендуется):**

- Background job обновляет счетчики
- Хранятся в таблице statistics
- Быстрый доступ без пересчета

### Активные пользователи (FR-326)

**Определение:**

- Пользователь с активностью в заданном периоде
- last_activity или last_login timestamp

**Расчет:**

```sql
-- Активные за 24 часа
SELECT COUNT(*)
FROM users
WHERE last_activity >= NOW() - INTERVAL '24 hours'
  AND is_banned = false;
```

### Использование

- Dashboard для администраторов
- Публичная страница "О платформе"
- Аналитика для принятия решений
- Marketing материалы

---

## Разрешения

### edit_platform_settings (FR-130)

- Обновление настроек платформы
- Обычно только у platform owner
- Может быть делегировано через роль

### view_analytics (FR-130, FR-325)

- Просмотр статистики платформы
- Для администраторов и аналитиков
- Разные уровни детализации (опционально)

### transfer_platform_ownership (FR-130)

- Передача владения
- Критическое разрешение
- Только у текущего owner

---

## Мониторинг

### Метрики

- Запросы к GetSettings (публичный endpoint)
- Частота обновлений настроек
- Рост пользовательской базы
- Engagement метрики (active users)

### Alerts

- Резкое падение активных пользователей
- Высокий процент непроверенных регистраций
- Spike в pending reports

---

## Примеры использования

### Получение настроек для клиента

```
GET /platform/settings
→ Возвращает название, описание, логотип для отображения
```

### Обновление названия платформы

```
PATCH /platform/settings
Body: {"name": "My Awesome Platform"}
→ Требует platform owner права
```

### Просмотр статистики

```
GET /platform/statistics
→ Требует view_analytics permission
→ Возвращает все счетчики
```

### Передача владения

```
1. POST /platform/transfer-ownership
   Body: {"new_owner_id": "user-123"}

2. Новый owner получает уведомление с токеном

3. POST /platform/confirm-ownership
   Body: {"token": "xyz..."}

4. Владение передано
```

---

## Производительность

### Кеширование

**GetSettings:**

- Public endpoint
- Кешировать агрессивно (1 час)
- Invalidate при UpdateSettings

**GetStatistics:**

- Тяжелые агрегации
- Materialize counts в отдельную таблицу
- Background job обновляет каждые 5-15 минут

### Индексы

- (is_banned, last_activity) для активных пользователей
- Счетчики в отдельной таблице для быстрого доступа

---

## Безопасность

### Критические операции

- UpdateSettings: только owner
- TransferOwnership: только owner + подтверждение
- GetStatistics: только с view_analytics

### Audit log

Логирование:

- Всех изменений настроек
- Попыток передачи владения
- Доступа к статистике

---

## Первоначальная настройка

### Checklist для нового owner

1. Обновить platform name
2. Заполнить description
3. Написать rules
4. Загрузить logo
5. Загрузить banners
6. Создать модераторские роли
7. Назначить первых модераторов
8. Создать первые сообщества
