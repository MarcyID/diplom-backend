# Diplom Backend

Backend для дипломного проекта — API для поиска и рекомендаций фильмов.

## 🚀 Быстрый старт

### 1. Установка зависимостей

```bash
go mod download
```

### 2. Настройка переменных окружения

Создайте файл `.env` в корне проекта:

```env
KINOPOISK_API_KEY=ваш_api_ключ
PORT=5454

# PostgreSQL (опционально, для аутентификации)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=diplom_db
DB_SSLMODE=disable

# JWT Secret (обязательно для аутентификации!)
JWT_SECRET=your-super-secret-key-change-in-production

# CORS разрешённые origins (для фронтенда)
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
```

Получить API ключ можно на [kinopoiskapiunofficial.tech](https://kinopoiskapiunofficial.tech).

### 3. Настройка базы данных (для аутентификации)

**Вариант 1: Автоматически (скрипт)**

```bash
# macOS/Linux
./scripts/init_db.sh

# Windows (PowerShell)
.\scripts\init_db.ps1
```

Скрипт создаст:
- Базу данных `diplom_db`
- Пользователя `diplom` с паролем `diplom`
- Все таблицы (users, sessions, collections, collection_films, favorites)

**Вариант 2: Вручную**

```bash
# Создать базу данных
createdb diplom_db

# Применить все миграции
psql -d diplom_db -f internal/database/migrations/complete.sql
```

**Структура БД:**
- `users` — пользователи (email, username, password_hash, full_name, avatar_url, banner_url, genre_preferences[])
- `sessions` — сессии (refresh_token_hash, user_agent, ip_address, expires_at)
- `collections` — подборки (title, description, is_public)
- `collection_films` — фильмы в подборках
- `favorites` — избранное (фильмы и персоны)

### 4. Запуск сервера

```bash
go run cmd/server/main.go
```

Сервер запустится на порту `5454` (или указанном в `.env`).

## 🔗 Интеграция с фронтендом

### Health Check

Фронтенд может проверять доступность бэкенда:

```bash
# Базовая проверка
GET http://localhost:5454/health

# Расширенная проверка (БД, Kinopoisk API)
GET http://localhost:5454/api/v1/health
```

**Ответ:**
```json
{
  "status": "ok",
  "services": {
    "server": "ok",
    "database": "ok",
    "kinopoisk": "ok"
  }
}
```

### Root endpoint

```bash
GET http://localhost:5454/
```

**Ответ:**
```json
{
  "name": "Diplom Backend API",
  "version": "1.0.0",
  "description": "Proxy server for Kinopoisk API with user authentication",
  "endpoints": {
    "health": "/health",
    "api": "/api/v1",
    "api_docs": "/api/v1/health"
  }
}
```

### CORS настройка

В `.env` укажите origins вашего фронтенда:

```env
# React (Create React App)
ALLOWED_ORIGINS=http://localhost:3000

# Vite
ALLOWED_ORIGINS=http://localhost:5173

# Несколько origins (через запятую)
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
```

### 📦 Кеширование Kinopoisk API

Для экономии лимита запросов к Kinopoisk API реализовано кеширование.

**Варианты кеша:**

1. **In-Memory** (по умолчанию) — быстрое кеширование в памяти сервера
   ```env
   CACHE_TYPE=memory
   CACHE_TTL_HOURS=24
   ```

2. **Redis** — надёжное кеширование с персистентностью
   ```env
   CACHE_TYPE=redis
   REDIS_ADDR=localhost:6379
   REDIS_PASSWORD=
   REDIS_DB=0
   CACHE_TTL_HOURS=24
   ```

**Запуск Redis (Docker):**
```bash
docker run -d --name diplom-redis -p 6379:6379 redis:7
```

**Кешируемые endpoints:**

Все запросы к Kinopoisk API кешируются. TTL настраивается в `.env` (по умолчанию 24 часа):

| Endpoint | Описание |
|----------|----------|
| `GET /api/v1/films/popular` | Популярные фильмы |
| `GET /api/v1/films/upcoming` | Предстоящие премьеры |
| `GET /api/v1/films/premieres` | Премьеры за текущий и следующий год |
| `GET /api/v1/films/:id` | Детали фильма |
| `GET /api/v1/movies/:id/similar` | Похожие фильмы |
| `GET /api/v1/movies/:id/staff` | Актёры и режиссёры фильма |
| `GET /api/v1/persons/:id` | Данные персоны |
| `GET /api/v1/actors/search` | Поиск актёров |
| `GET /api/v1/movies/search` | Поиск фильмов |

**Преимущества:**
- Экономия лимита API Kinopoisk (1000 запросов/день для бесплатного тарифа)
- Ускорение ответа сервера (данные из кеша)
- Снижение нагрузки на внешнее API

### Пример запроса с фронтенда (React/Vite)

```javascript
// Конфигурация API
const API_BASE_URL = 'http://localhost:5454/api/v1';

// Вход
async function login(email, password) {
  const response = await fetch(`${API_BASE_URL}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  });
  
  const data = await response.json();
  
  if (!response.ok) {
    throw new Error(data.error);
  }
  
  // Сохраняем токены
  localStorage.setItem('access_token', data.data.access_token);
  localStorage.setItem('refresh_token', data.data.refresh_token);
  
  return data.data;
}

// Запрос с токеном
async function getFilms(query) {
  const token = localStorage.getItem('access_token');
  
  const response = await fetch(`${API_BASE_URL}/movies/search?q=${query}`, {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  
  if (response.status === 401) {
    // Токен истёк - нужно обновить
    await refreshTokens();
    return getFilms(query); // Повторить запрос
  }
  
  return response.json();
}

// Обновление токенов
async function refreshTokens() {
  const refreshToken = localStorage.getItem('refresh_token');
  
  const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token: refreshToken })
  });
  
  const data = await response.json();
  
  if (!response.ok) {
    // Refresh token истёк - нужно разлогинить
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    window.location.href = '/login';
    throw new Error('Session expired');
  }
  
  localStorage.setItem('access_token', data.data.access_token);
  localStorage.setItem('refresh_token', data.data.refresh_token);
}
```

### Формат ответов API

**Успешный ответ:**
```json
{
  "data": { ... },
  "message": "Success message"
}
```

**Ошибка:**
```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": "Optional details"
}
```

**Коды ошибок:**
- `BAD_REQUEST` (400) — некорректный запрос
- `UNAUTHORIZED` (401) — требуется авторизация
- `FORBIDDEN` (403) — нет доступа
- `NOT_FOUND` (404) — ресурс не найден
- `CONFLICT` (409) — конфликт (например, email занят)
- `VALIDATION_ERROR` (422) — ошибка валидации
- `INTERNAL_SERVER_ERROR` (500) — внутренняя ошибка

## 🔐 Аутентификация

| Метод | Endpoint | Описание |
|-------|----------|----------|
| POST | `/api/v1/auth/register` | Регистрация нового пользователя |
| POST | `/api/v1/auth/login` | Вход в систему |
| POST | `/api/v1/auth/logout` | Выход из системы (требует токен) |
| GET | `/api/v1/auth/me` | Получить данные текущего пользователя (требует токен) |
| POST | `/api/v1/auth/refresh` | Обновить пару токенов |

## 👤 Профиль

| Метод | Endpoint | Описание |
|-------|----------|----------|
| GET | `/api/v1/profile/me` | Получить мой профиль (требует токен) |
| PUT | `/api/v1/profile` | Обновить профиль (требует токен) |

### Примеры запросов

#### Обновить профиль

```bash
PUT /api/v1/profile
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
  "full_name": "John Doe",
  "avatar_url": "https://example.com/avatar.jpg",
  "banner_url": "https://example.com/banner.jpg"
}
```

## 🎬 Подборки фильмов

Подборки — это публичные или приватные коллекции фильмов, которые пользователи могут создавать и делиться ими.

| Метод | Endpoint | Описание |
|-------|----------|----------|
| GET | `/api/v1/collections/:id` | Получить подборку по ID (публичная или владелец) |
| GET | `/api/v1/collections/my` | Получить все мои подборки (требует токен) |
| POST | `/api/v1/collections` | Создать новую подборку (требует токен) |
| PUT | `/api/v1/collections/:id` | Обновить подборку (требует токен, владелец) |
| DELETE | `/api/v1/collections/:id` | Удалить подборку (требует токен, владелец) |
| POST | `/api/v1/collections/:id/films` | Добавить фильм в подборку (требует токен, владелец) |
| DELETE | `/api/v1/collections/:id/films/:filmId` | Удалить фильм из подборки (требует токен, владелец) |
| PUT | `/api/v1/collections/:id/films/reorder` | Изменить порядок фильмов (требует токен, владелец) |
| GET | `/api/v1/users/:id/collections` | Получить публичные подборки пользователя |

### Примеры запросов

#### Создать подборку

```bash
POST /api/v1/collections
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
  "title": "Любимые комедии 90-х",
  "description": "Фильмы, которые заставляют смеяться",
  "is_public": true
}
```

#### Получить подборку по ID (публичный доступ)

```bash
GET /api/v1/collections/123
```

**Ответ:**
```json
{
  "collection": {
    "id": 123,
    "user_id": 1,
    "title": "Любимые комедии 90-х",
    "description": "Фильмы, которые заставляют смеяться",
    "is_public": true,
    "created_at": "2026-05-25T10:00:00Z",
    "updated_at": "2026-05-25T10:00:00Z",
    "films": [
      {
        "kinopoiskId": 12345,
        "nameRu": "Маска",
        "nameEn": "The Mask",
        "year": 1994,
        "posterUrl": "...",
        "ratingKinopoisk": 7.8,
        "type": "FILM"
      }
    ]
  }
}
```

#### Добавить фильм в подборку

```bash
POST /api/v1/collections/123/films
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
  "film_id": 12345,
  "position": 1
}
```

#### Получить публичные подборки пользователя

```bash
GET /api/v1/users/1/collections?page=1&page_size=20
```

## ❤️ Избранное

Избранное позволяет сохранять фильмы и персоны в личный список.

| Метод | Endpoint | Описание |
|-------|----------|----------|
| GET | `/api/v1/favorites` | Получить все избранное (требует токен) |
| POST | `/api/v1/favorites/film/:filmId` | Добавить фильм (требует токен) |
| DELETE | `/api/v1/favorites/film/:filmId` | Удалить фильм (требует токен) |
| POST | `/api/v1/favorites/person/:personId` | Добавить персону (требует токен) |
| DELETE | `/api/v1/favorites/person/:personId` | Удалить персону (требует токен) |
| POST | `/api/v1/favorites/toggle/film/:filmId` | Переключить фильм (требует токен) |
| POST | `/api/v1/favorites/toggle/person/:personId` | Переключить персону (требует токен) |

### Примеры запросов

#### Получить избранное

```bash
GET /api/v1/favorites?page=1&page_size=20
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Возвращает:**
```json
{
  "total": 10,
  "page": 1,
  "items": [
    {
      "object_type": "film",
      "object_id": 12345,
      "created_at": "2026-05-25T10:00:00Z",
      "film_data": {
        "kinopoiskId": 12345,
        "nameRu": "Маска",
        "nameEn": "The Mask",
        "posterUrl": "...",
        "year": 1994,
        "ratingKinopoisk": 7.8,
        "type": "FILM"
      }
    },
    {
      "object_type": "person",
      "object_id": 66539,
      "created_at": "2026-05-25T11:00:00Z",
      "person_data": {
        "personId": 66539,
        "nameRu": "Винс Гиллиган",
        "posterUrl": "...",
        "professionText": "Режиссер"
      }
    }
  ]
}
```

#### Переключить фильм в избранном

```bash
POST /api/v1/favorites/toggle/film/12345
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Возвращает:**
```json
{
  "message": "Film added to favorites",
  "in_favorites": true
}
```

### Примеры запросов

#### Регистрация

```bash
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "username": "john_doe",
  "password": "securepassword123",
  "full_name": "John Doe"
}
```

#### Вход

```bash
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Ответ:**
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "username": "john_doe",
    "full_name": "John Doe",
    "created_at": "2026-05-25T10:00:00Z"
  },
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Получение данных пользователя (с токеном)

```bash
GET /api/v1/auth/me
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### Выход

```bash
POST /api/v1/auth/logout
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

## 📡 API Endpoints

### 🔐 Аутентификация

| Метод | Endpoint | Описание | Auth |
|-------|----------|----------|------|
| POST | `/api/v1/auth/register` | Регистрация пользователя | ❌ |
| POST | `/api/v1/auth/login` | Вход (получение токенов) | ❌ |
| POST | `/api/v1/auth/refresh` | Обновление access токена | ❌ |
| POST | `/api/v1/auth/logout` | Выход (удаление сессии) | ✅ |
| GET | `/api/v1/auth/me` | Получить текущий профиль | ✅ |

### 👤 Профиль пользователя

| Метод | Endpoint | Описание | Auth |
|-------|----------|----------|------|
| GET | `/api/v1/profile/me` | Получить профиль | ✅ |
| PUT | `/api/v1/profile` | Обновить профиль (full_name, avatar_url, banner_url) | ✅ |
| GET | `/api/v1/profile/genres` | Получить жанровые предпочтения | ✅ |
| PUT | `/api/v1/profile/genres` | Обновить жанровые предпочтения | ✅ |
| POST | `/api/v1/profile/avatar` | Загрузить аватар (multipart/form-data) | ✅ |
| POST | `/api/v1/profile/banner` | Загрузить фон (multipart/form-data) | ✅ |
| DELETE | `/api/v1/profile/avatar` | Удалить аватар | ✅ |
| DELETE | `/api/v1/profile/banner` | Удалить фон | ✅ |

### 📚 Подборки (Collections)

| Метод | Endpoint | Описание | Auth |
|-------|----------|----------|------|
| GET | `/api/v1/collections/my` | Мои подборки | ✅ |
| POST | `/api/v1/collections/my` | Создать подборку | ✅ |
| GET | `/api/v1/collections/:id` | Получить подборку по ID | ✅ |
| PUT | `/api/v1/collections/:id` | Обновить подборку | ✅ |
| DELETE | `/api/v1/collections/:id` | Удалить подборку | ✅ |
| POST | `/api/v1/collections/:id/films/:filmId` | Добавить фильм | ✅ |
| DELETE | `/api/v1/collections/:id/films/:filmId` | Удалить фильм | ✅ |
| PUT | `/api/v1/collections/:id/films/reorder` | Изменить порядок фильмов | ✅ |
| GET | `/api/v1/collections/user/:userId` | Публичные подборки пользователя | ❌ |

### ⭐ Избранное (Favorites)

| Метод | Endpoint | Описание | Auth |
|-------|----------|----------|------|
| GET | `/api/v1/favorites` | Получить избранное (фильмы и персоны) | ✅ |
| POST | `/api/v1/favorites/toggle/film/:filmId` | Добавить/удалить фильм | ✅ |
| POST | `/api/v1/favorites/toggle/person/:personId` | Добавить/удалить персону | ✅ |

### 🎬 Фильмы

| Метод | Endpoint | Описание | Auth |
|-------|----------|----------|------|
| GET | `/api/v1/movies/search` | Поиск фильмов | ❌ |
| GET | `/api/v1/movies/:id` | Фильм по ID | ❌ |
| GET | `/api/v1/movies/random` | Случайный фильм | ❌ |
| GET | `/api/v1/movies/:id/similar` | Похожие фильмы по ID | ❌ |
| GET | `/api/v1/movies/similar/by-title` | Похожие по названию | ❌ |
| GET | `/api/v1/films/popular` | Популярные фильмы (топ) | ❌ |
| GET | `/api/v1/films/upcoming` | Предстоящие премьеры | ❌ |
| GET | `/api/v1/films/premieres` | Премьеры за текущий и следующий год | ❌ |

### 🎭 Актёры и режиссёры

| Метод | Endpoint | Описание | Auth |
|-------|----------|----------|------|
| GET | `/api/v1/actors/search` | Поиск актёров по имени | ❌ |
| GET | `/api/v1/actors/:id` | Данные об актёре по ID | ❌ |
| GET | `/api/v1/actors/:id/filmography` | Фильмография актёра | ❌ |
| GET | `/api/v1/movies/:id/staff` | Актёры и режиссёры фильма | ❌ |
| GET | `/api/v1/persons/:id` | Данные о персоне (алиас) | ❌ |

### 📋 Справочная информация

| Метод | Endpoint | Описание | Auth |
|-------|----------|----------|------|
| GET | `/api/v1/genres` | Жанры и страны для фильтров | ❌ |
| GET | `/api/v1/health` | Проверка здоровья сервера | ❌ |

### Примеры запросов

#### Поиск фильмов

```bash
GET /api/v1/movies/search?q=матрица&page=1
```

**Параметры:**
- `q` — поисковый запрос
- `genre` — ID жанра (можно несколько)
- `country` — ID страны
- `year_from` / `year_to` — диапазон лет
- `rating_min` / `rating_max` — диапазон рейтинга
- `page` — номер страницы

#### Фильм по ID

```bash
GET /api/v1/movies/301
```

#### Похожие фильмы

```bash
GET /api/v1/movies/301/similar
```

### Актёры и режиссёры

#### Актёры фильма

```bash
GET /api/v1/movies/301/staff
```

#### Данные о персоне

```bash
GET /api/v1/persons/119448
```

Примеры ID персон:
- `119448` — Киану Ривз
- `66539` — Винс Гиллиган (режиссёр)

#### Поиск актёров

```bash
GET /api/v1/actors/search?q=Киану&page=1
```

**Параметры:**
- `q` — поисковый запрос (имя актёра)
- `page` — номер страницы (макс. 2)

#### Фильмография актёра

```bash
GET /api/v1/actors/119448/filmography
```

#### Популярные фильмы

```bash
GET /api/v1/films/popular?page=1
```

#### Предстоящие премьеры

```bash
GET /api/v1/films/upcoming?page=1
```

#### Премьеры за указанный год и месяц

```bash
GET /api/v1/films/premieres?year=2025&month=JANUARY
```

**Параметры:**
- `year` - год (обязательно)
- `month` - месяц на английском (обязательно): JANUARY, FEBRUARY, MARCH, APRIL, MAY, JUNE, JULY, AUGUST, SEPTEMBER, OCTOBER, NOVEMBER, DECEMBER

**Возвращает:**
- `total` - общее количество премьер
- `items` - массив премьер с полями:
  - `kinopoiskId` - ID фильма
  - `nameRu`, `nameEn` - названия
  - `year` - год
  - `posterUrl`, `posterUrlPreview` - постеры
  - `countries` - страны
  - `genres` - жанры
  - `duration` - длительность
  - `premiereRu` - дата российской премьеры

## 🧪 Тестирование через Postman

Готовая коллекция запросов находится в [`/api/postman/`](./api/postman/).

**Коллекция включает:**
- 🔐 Аутентификация (register, login, logout, refresh)
- 👤 Профиль (аватары, баннеры, жанровые предпочтения)
- ⭐ Избранное (фильмы и персоны)
- 📚 Подборки (CRUD, фильмы, порядок)
- 🎬 Фильмы (поиск, детали, похожие, премьеры)
- 🎭 Актёры и режиссёры

**Использование:**
1. Откройте Postman
2. Импортируйте `Kinopoisk_API.postman_collection.json` через **Import**
3. Выполните `Login` для получения токенов
4. Тестируйте endpoints

📖 Подробнее в [`api/postman/README.md`](./api/postman/README.md).

## 📁 Структура проекта

```
.
├── cmd/
│   └── server/
│       └── main.go              # Точка входа
├── internal/
│   ├── api/
│   │   ├── router.go            # Маршрутизация
│   │   ├── middleware/
│   │   │   └── auth.go          # JWT middleware
│   │   └── handlers/
│   │       ├── movie.go         # Обработчики фильмов
│   │       ├── auth.go          # Обработчики аутентификации
│   │       ├── profile.go       # Профиль (avatar, banner, genres)
│   │       ├── collection.go    # Подборки
│   │       └── favorite.go      # Избранное
│   ├── database/
│   │   ├── postgres.go          # Подключение к PostgreSQL
│   │   └── migrations/
│   │       ├── 001_init_auth.up.sql
│   │       ├── 002_collections.up.sql
│   │       ├── 003_favorites.up.sql
│   │       ├── 004_user_profile.up.sql
│   │       └── complete.sql     # Все миграции вместе
│   ├── model/
│   │   ├── kinopoisk/           # Модели данных Kinopoisk API
│   │   │   ├── common.go
│   │   │   ├── film.go
│   │   │   └── person.go
│   │   ├── auth/
│   │   │   ├── user.go          # User, UserInfo, JWTClaims
│   │   │   └── session.go
│   │   ├── collection/
│   │   │   └── collection.go    # Collection, CollectionFilm
│   │   └── favorite/
│   │       └── favorite.go      # Favorite (film/person)
│   ├── repository/
│   │   ├── interfaces.go
│   │   ├── user_repository.go
│   │   ├── session_repository.go
│   │   ├── collection_repository.go
│   │   ├── collection_postgres_repository.go
│   │   └── favorite_repository.go
│   └── service/
│       ├── kinopoisk_client.go  # Клиент Kinopoisk API
│       ├── auth_service.go      # Сервис аутентификации
│       ├── user_service.go      # Профиль пользователя
│       ├── file_upload_service.go # Загрузка файлов
│       ├── collection_service.go
│       └── favorite_service.go
├── scripts/
│   ├── init_db.sh               # Скрипт инициализации БД (macOS/Linux)
│   └── init_db.ps1              # Скрипт инициализации БД (Windows)
├── uploads/
│   ├── avatars/                 # Аватары пользователей
│   └── banners/                 # Фоны профилей
├── api/
│   ├── openapi/kinopoisk/
│   │   └── openapi.json
│   └── postman/
│       ├── Kinopoisk_API.postman_collection.json
│       └── README.md
├── .env
├── .env.example
├── go.mod
└── README.md
```

## 🔧 Технологии

- **Go** — язык программирования
- **Gin** — веб-фреймворк
- **PostgreSQL** — база данных (pgx/v5)
- **Redis** — кеширование (go-redis/v9, опционально)
- **JWT** — аутентификация (golang-jwt/jwt/v5)
- **bcrypt** — хеширование паролей
- **Kinopoisk API Unofficial** — источник данных о фильмах

## 📚 Документация

- [Kinopoisk API Unofficial](https://kinopoiskapiunofficial.tech/documentation/api)
- [Gin Framework](https://gin-gonic.com/docs/)
- OpenAPI спецификация: [`/api/openapi/kinopoisk/openapi.json`](./api/openapi/kinopoisk/openapi.json)

## 📝 Лицензия

Учебный проект для дипломной работы.
