# Система поиска

## Обзор

SearchService предоставляет унифицированный полнотекстовый поиск по постам, комментариям, пользователям и сообществам используя PostgreSQL Full-Text Search.

## gRPC Service: SearchService

Proto файл: `proto/search.proto`

## Сущности

### ContentType

```protobuf
enum ContentType {
  CONTENT_TYPE_UNSPECIFIED = 0
  CONTENT_TYPE_POSTS       = 1
  CONTENT_TYPE_COMMENTS    = 2
  CONTENT_TYPE_USERS       = 3
  CONTENT_TYPE_COMMUNITIES = 4
  CONTENT_TYPE_ALL         = 5
}
```

### SearchResult

```protobuf
message SearchResult {
  string result_type  // post, comment, user, community
  oneof result {
    Post post
    Comment comment
    UserProfile user
    Community community
  }
  float relevance_score
}
```

## Endpoint

### Search

**RPC:** `Search(SearchRequest) returns (SearchResponse)`  
**HTTP:** `GET /search`  
**FR:** FR-046-049, FR-090-092, FR-268-279

Унифицированный поиск по всем типам контента.

**Request:**

```protobuf
message SearchRequest {
  string query              // минимум 3 символа
  ContentType content_type
  string cursor
  int32 limit
}
```

**Response:**

```protobuf
message SearchResponse {
  repeated SearchResult results
  string next_cursor
  bool has_more
}
```

**Требования:**

- Cursor-based пагинация (FR-270)
- Минимальная длина запроса: 3 символа (FR-277)
- Поддержка exact phrase search в кавычках (FR-278)
- Исключение контента из забаненных пользователей и сообществ (FR-279)

---

## Типы поиска

### POSTS (FR-046, FR-271)

Поиск по заголовкам постов.

**Индексация:**

- title поля Post
- tsvector колонка с GIN индексом (FR-090)

**Возврат:**

- Полные объекты Post
- relevance_score на основе ts_rank (FR-091)

**Пример запроса:**

```
query: "golang backend"
content_type: POSTS
```

---

### COMMENTS (FR-047, FR-272)

Поиск по тексту комментариев.

**Индексация:**

- text поля Comment
- tsvector с GIN индексом

**Возврат:**

- Полные объекты Comment
- relevance_score

**Особенность:**
Результаты включают context поста для навигации.

---

### USERS (FR-048, FR-273)

Поиск по именам пользователей.

**Индексация:**

- username поля User
- tsvector для morphology support

**Возврат:**

- UserProfile объекты
- relevance_score

**Фильтрация:**

- Исключение забаненных пользователей (FR-279)

---

### COMMUNITIES (FR-049, FR-274)

Поиск по именам сообществ.

**Индексация:**

- name поля Community
- Опционально description

**Возврат:**

- Community объекты
- relevance_score

**Фильтрация:**

- Исключение забаненных сообществ для не-модераторов (FR-279)

---

### ALL (FR-275)

Смешанный поиск по всем типам контента.

**Результаты:**

- Комбинация всех типов
- Сортировка по relevance_score
- Указание content type для каждого результата (FR-276)

**Формат:**

```json
{
  "results": [
    {"result_type": "post", "post": {...}, "relevance_score": 0.95},
    {"result_type": "user", "user": {...}, "relevance_score": 0.87},
    {"result_type": "community", "community": {...}, "relevance_score": 0.82}
  ]
}
```

---

## PostgreSQL Full-Text Search

### Технология

- **tsvector**: текст преобразован в токены для индексации
- **GIN индексы**: быстрый поиск по tsvector колонкам (FR-090)
- **ts_rank**: функция ранжирования результатов (FR-091)

### Создание индексов

```sql
-- Пример для постов
ALTER TABLE posts ADD COLUMN title_tsv tsvector;
UPDATE posts SET title_tsv = to_tsvector('english', title);
CREATE INDEX posts_title_tsv_idx ON posts USING GIN(title_tsv);
```

### Запрос

```sql
-- Пример поиска
SELECT *, ts_rank(title_tsv, query) AS rank
FROM posts, to_tsquery('english', 'golang & backend') query
WHERE title_tsv @@ query
ORDER BY rank DESC;
```

---

## Morphology и Stemming

### Поддержка (FR-092)

PostgreSQL FTS поддерживает:

- Базовый stemming (word → root form)
- Stop words фильтрация
- Языковые конфигурации

### Языки

- Основной: English
- Опционально: Русский, другие
- Конфигурируется через text search config

### Примеры

```
Query: "running"
Matches: "run", "runs", "running", "ran"

Query: "develop"
Matches: "developer", "development", "developed"
```

---

## Exact Phrase Search

### Синтаксис (FR-278)

Кавычки для точной фразы:

```
"golang backend"
```

### Реализация

- Использование phraseto_tsquery вместо to_tsquery
- Сохранение порядка и позиций слов
- Более strict matching

### Примеры

```
Query: "golang backend"
Matches: "golang backend development"
NOT matches: "backend in golang"

Query: golang backend (без кавычек)
Matches: "backend in golang" (любой порядок)
```

---

## Ранжирование

### ts_rank функция (FR-091)

Факторы ранжирования:

- Частота токенов в документе
- Близость токенов друг к другу
- Позиция в документе (title важнее body)

### Нормализация

```sql
ts_rank(tsvector, query, normalization)
```

Опции normalization:

- 0: no normalization
- 1: divide by document length
- 2: divide by number of unique words
- 4: divide by harmonic distance between extents

### Приоритеты

- Title match выше чем content match
- Exact match выше чем partial
- Полные слова выше чем префиксы

---

## Фильтрация результатов

### Забаненные пользователи (FR-279)

Исключить из результатов:

- Посты от забаненных пользователей
- Комментарии от забаненных пользователей
- Сами забаненные пользователи (в user search)

### Забаненные сообщества (FR-279)

- Исключить для обычных пользователей
- Включить для модераторов с view_all_communities

### Неопубликованный контент

- Только published посты
- Drafts не индексируются

---

## Минимальная длина запроса

### Требование (FR-277)

Запросы короче 3 символов возвращают пустой результат.

**Причины:**

- Производительность (слишком широкие результаты)
- UX (слишком много нерелевантных результатов)
- Защита от abuse

**Реализация:**

```go
if len(query) < 3 {
    return &SearchResponse{
        Results: []SearchResult{},
        NextCursor: "",
        HasMore: false,
    }
}
```

---

## Cursor Pagination

### Реализация

Cursor кодирует:

- Последний relevance_score
- Последний ID результата
- Content type для фильтрации

### Стабильность

- Результаты могут меняться при новом контенте
- Cursor обеспечивает консистентную пагинацию
- Keyset pagination на основе (score, id)

---

## Производительность

### Целевые метрики (SC-005)

- 95% запросов < 2 секунды
- Быстрый поиск даже с большим объемом данных

### Оптимизации

**Индексы:**

- GIN индексы на всех tsvector колонках (FR-090)
- Composite индексы для фильтрации

**Кеширование:**

- Popular queries в Redis
- TTL 5-10 минут
- Cache warming для trending searches

**Limits:**

- Максимум 1000 результатов per query
- Cursor pagination для deep pagination

---

## Расширенные возможности

### Autocomplete (опционально)

- Prefix matching для suggestions
- Trigram индексы для fuzzy search
- Кеширование popular suggestions

### Faceted search (опционально)

Фильтры по:

- Community
- Date range
- User
- Content type

### Highlights

Подсветка найденных терминов:

```sql
ts_headline('english', document, query)
```

---

## Мониторинг

### Метрики

- Search query latency
- Empty result rate
- Popular queries
- Search type distribution

### Аналитика

- Trending search terms
- Failed searches (для улучшения)
- Query length distribution

---

## Ограничения

### PostgreSQL FTS

**Преимущества:**

- Встроен в PostgreSQL
- Не требует отдельного сервиса
- Хорошая производительность до среднего масштаба

**Недостатки:**

- Менее мощный чем ElasticSearch
- Ограниченная настройка ранжирования
- Не для очень больших объемов (>10M документов)

### Альтернативы

Для масштабирования:

- Elasticsearch
- MeiliSearch
- Typesense

---

## Примеры запросов

### Поиск постов

```
GET /search?query=golang&content_type=POSTS&limit=20
```

### Поиск пользователей

```
GET /search?query=john&content_type=USERS&limit=10
```

### Поиск всего

```
GET /search?query="web development"&content_type=ALL&limit=50
```

### С пагинацией

```
GET /search?query=backend&content_type=ALL&cursor=xyz123&limit=20
```
