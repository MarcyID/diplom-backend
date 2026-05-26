# Технологии разработки Backend для дипломного проекта

## 1. Общие сведения о проекте

**Наименование проекта:** Diplom Backend  
**Тип приложения:** REST API сервер (прокси-сервис)  
**Назначение:** Предоставление данных о фильмах, актёрах и режиссёрах из Kinopoisk API Unofficial с возможностью кеширования и пользовательской аутентификации.

---

## 2. Стек технологий

### 2.1 Язык программирования

**Go (Golang) версии 1.21+**

**Обоснование выбора:**
- Высокая производительность благодаря компиляции в машинный код
- Встроенная поддержка многопоточности (горутины)
- Статическая типизация для предотвращения ошибок на этапе компиляции
- Минимальное потребление памяти по сравнению с интерпретируемыми языками
- Отличная поддержка работы с HTTP и JSON из коробки

**Пример объявления структуры на Go:**
```go
type KinopoiskFilm struct {
    KinopoiskID int64     `json:"kinopoiskId"`
    NameRU      *string   `json:"nameRu"`
    Year        *int      `json:"year"`
    RatingImdb  *float64  `json:"ratingImbd"`
}
```

---

### 2.2 Веб-фреймворк

**Gin Web Framework**

**Назначение:** Обработка HTTP-запросов, маршрутизация, middleware.

**Обоснование выбора:**
- Высокая производительность (до 40 раз быстрее net/http)
- Поддержка middleware для обработки запросов
- Встроенная валидация JSON
- Удобная работа с параметрами запроса

**Пример маршрутизатора:**
```go
api := r.Group("/api/v1")
{
    api.GET("/films/popular", movieHandler.GetPopularFilms)
    api.GET("/movies/:id", movieHandler.GetMovieByID)
    api.POST("/auth/login", authHandler.Login)
}
```

---

### 2.3 База данных

**PostgreSQL версии 15**

**Назначение:** Хранение данных пользователей, сессий аутентификации.

**Обоснование выбора:**
- Надёжность и ACID-транзакции
- Поддержка внешних ключей и каскадного удаления
- Индексы для ускорения поиска
- Триггеры для автоматического обновления полей

**Схема базы данных:**

**Таблица `users`:**
```sql
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(50) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100),
    avatar_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

**Таблица `sessions`:**
```sql
CREATE TABLE sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash VARCHAR(255) NOT NULL,
    user_agent TEXT,
    ip_address VARCHAR(45),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_revoked BOOLEAN DEFAULT FALSE
);
```

**Драйвер подключения:** `pgx/v5` (нативный PostgreSQL драйвер для Go)

---

### 2.4 Система кеширования

**Redis / In-Memory Cache**

**Назначение:** Кеширование ответов от Kinopoisk API для экономии лимитов.

**Обоснование выбора:**
- Redis: персистентное хранение, возможность масштабирования
- In-Memory: быстрое кеширование без внешних зависимостей
- TTL (Time To Live) — 24 часа для всех запросов

**Алгоритм кеширования:**
1. Поступает запрос к API
2. Генерируется ключ кеша: `kinopoisk:/endpoint:param=value`
3. Проверка наличия в кеше
4. При Cache HIT — возврат данных из кеша
5. При Cache MISS — запрос к Kinopoisk API, сохранение в кеш, возврат данных

**Пример ключа кеша:**
```
kinopoisk:/collections:type=TOP_POPULAR_MOVIES:page=1
```

---

### 2.5 Аутентификация и авторизация

**JWT (JSON Web Tokens)**

**Библиотека:** `golang-jwt/jwt/v5`

**Структура токена:**
```json
{
  "user_id": 1,
  "username": "john_doe",
  "exp": 1716624000,
  "iat": 1716623100
}
```

**Время жизни токенов:**
- Access Token: 15 минут (для доступа к API)
- Refresh Token: 7 дней (для обновления пары токенов)

**Хеширование паролей:** `bcrypt` (стоимость: 10)

**Схема аутентификации:**
1. Пользователь отправляет email/password
2. Сервер проверяет хеш пароля через `bcrypt.CompareHashAndPassword`
3. Генерируется пара JWT токенов
4. Refresh Token хешируется (SHA-256) и сохраняется в БД
5. Access Token возвращается клиенту для последующих запросов

---

### 2.6 Внешнее API

**Kinopoisk API Unofficial**

**Базовый URL:** `https://kinopoiskapiunofficial.tech`

**Аутентификация:** API Key в заголовке `X-API-KEY`

**Лимиты бесплатного тарифа:** 1000 запросов в день

**Основные endpoints:**
- `GET /api/v2.2/films` — поиск фильмов
- `GET /api/v2.2/films/{id}` — детали фильма
- `GET /api/v2.2/films/collections` — коллекции (топы)
- `GET /api/v1/staff/{id}` — данные персоны
- `GET /api/v1/persons` — поиск персон

---

## 3. Архитектура приложения

### 3.1 Слоёная архитектура

```
┌─────────────────────────────────────────┐
│           HTTP Layer (Gin)              │
│         /internal/api/handlers/         │
├─────────────────────────────────────────┤
│         Service Layer                   │
│       /internal/service/                │
│  (бизнес-логика, кеширование, JWT)      │
├─────────────────────────────────────────┤
│       Repository Layer                  │
│      /internal/repository/              │
│    (работа с базой данных)              │
├─────────────────────────────────────────┤
│        Data Layer                       │
│   PostgreSQL / Redis / HTTP Client      │
└─────────────────────────────────────────┘
```

### 3.2 Структура проекта

```
diplom-backend/
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
│   │       └── cache.go         # Обработчики кеша
│   ├── cache/
│   │   ├── cache.go             # Интерфейс кеша
│   │   ├── memory_cache.go      # In-Memory реализация
│   │   ├── redis_cache.go       # Redis реализация
│   │   └── factory.go           # Фабрика кеша
│   ├── database/
│   │   ├── postgres.go          # Подключение к PostgreSQL
│   │   └── migrations/          # SQL миграции
│   ├── model/
│   │   ├── kinopoisk/           # Модели данных API
│   │   └── auth/                # Модели аутентификации
│   ├── repository/
│   │   ├── interfaces.go        # Интерфейсы
│   │   ├── user_repository.go   # Репозиторий пользователей
│   │   └── session_repository.go# Репозиторий сессий
│   └── service/
│       ├── kinopoisk_client.go  # Клиент API с кешированием
│       └── auth_service.go      # Сервис аутентификации
├── api/
│   ├── openapi/                 # OpenAPI спецификации
│   └── postman/                 # Postman коллекция
├── .env                         # Переменные окружения
├── go.mod                       # Зависимости Go
└── README.md                    # Документация
```

---

## 4. Детали реализации

### 4.1 Кеширование запросов

**Алгоритм работы:**

```go
func (c *KinopoiskClient) GetFilmByID(filmID int64) (*KinopoiskFilm, error) {
    ctx := context.Background()
    
    // 1. Генерация ключа
    cacheKey := c.cacheKey("/films/id", map[string]string{
        "id": strconv.FormatInt(filmID, 10),
    })
    
    // 2. Проверка кеша
    if cachedData, found := c.cacheGet(ctx, cacheKey); found {
        var film KinopoiskFilm
        json.Unmarshal(cachedData, &film)
        return &film, nil
    }
    
    // 3. Запрос к API
    resp, err := c.client.Do(req)
    // ... обработка ответа
    
    // 4. Сохранение в кеш
    c.cacheSet(ctx, cacheKey, jsonMustMarshal(&film))
    
    return &film, nil
}
```

**Экономия запросов:**
- Без кеширования: 50 пользователей × 20 запросов = 1000 запросов (лимит исчерпан)
- С кешированием: 20 запросов (первый запрос) + 0 (все остальные из кеша)
- **Экономия: 98%**

---

### 4.2 Аутентификация

**Регистрация пользователя:**

```go
func (s *AuthService) Register(ctx context.Context, req CreateUserRequest) (*User, error) {
    // 1. Проверка существования пользователя
    existingUser, _ := s.userRepo.GetByEmail(ctx, req.Email)
    if existingUser != nil {
        return nil, errors.New("user already exists")
    }
    
    // 2. Хеширование пароля
    hashedPassword, _ := bcrypt.GenerateFromPassword(
        []byte(req.Password), 
        bcrypt.DefaultCost
    )
    
    // 3. Создание в БД
    user := &User{
        Email:    req.Email,
        Username: req.Username,
        Password: string(hashedPassword),
    }
    return s.userRepo.Create(ctx, user)
}
```

**Вход (Login):**

```go
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
    // 1. Поиск пользователя
    user, _ := s.userRepo.GetByEmail(ctx, req.Email)
    
    // 2. Проверка пароля
    err := bcrypt.CompareHashAndPassword(
        []byte(user.Password), 
        []byte(req.Password)
    )
    
    // 3. Генерация JWT токенов
    accessToken, refreshToken := s.generateTokens(user)
    
    // 4. Создание сессии
    s.createSession(user.ID, refreshToken)
    
    return &AuthResponse{
        User: user.ToUserInfo(),
        AccessToken: accessToken,
        RefreshToken: refreshToken,
    }, nil
}
```

**Middleware для проверки токена:**

```go
func JWTAuth(authService *AuthService) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Получение токена из заголовка
        authHeader := c.GetHeader("Authorization")
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        
        // 2. Валидация токена
        claims, err := authService.ValidateToken(tokenString)
        if err != nil {
            c.JSON(401, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }
        
        // 3. Сохранение userID в контекст
        c.Set("userID", claims.UserID)
        c.Next()
    }
}
```

---

### 4.3 Обработка ошибок

**Формат ответа об ошибке:**
```json
{
  "error": "Invalid email or password",
  "code": "UNAUTHORIZED",
  "details": "Optional details"
}
```

**Коды ошибок:**
- `BAD_REQUEST` (400) — некорректный запрос
- `UNAUTHORIZED` (401) — требуется авторизация
- `FORBIDDEN` (403) — нет доступа
- `NOT_FOUND` (404) — ресурс не найден
- `CONFLICT` (409) — конфликт (email занят)
- `INTERNAL_SERVER_ERROR` (500) — внутренняя ошибка

---

## 5. Конфигурация и развёртывание

### 5.1 Переменные окружения

```env
# Kinopoisk API
KINOPOISK_API_KEY=ваш_ключ

# Сервер
PORT=5454

# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=diplom_db
DB_SSLMODE=disable

# JWT
JWT_SECRET=ваш_секретный_ключ

# CORS
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173

# Кеширование
CACHE_TYPE=memory
CACHE_TTL_HOURS=24

# Redis (опционально)
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
```

---

### 5.2 Запуск приложения

**Локальная разработка:**
```bash
# Установка зависимостей
go mod download

# Применение миграций
psql -d diplom_db -f internal/database/migrations/001_init_auth.up.sql

# Запуск сервера
go run cmd/server/main.go
```

**Production (Docker):**
```dockerfile
FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go build -o server ./cmd/server
EXPOSE 5454
CMD ["./server"]
```

---

### 5.3 Развёртывание проекта

**Важно:** Проект состоит из двух частей, которые развёртываются на разных платформах:

| Часть проекта | Платформа | Тип хостинга |
|---------------|-----------|--------------|
| **Frontend** (React/Vite) | GitHub Pages | Статический хостинг |
| **Backend** (Go API) | Render/Railway/Heroku | Cloud PaaS |

#### Развёртывание Frontend на GitHub Pages

**GitHub Pages** — бесплатный хостинг для статических веб-сайтов.

**Шаги развёртывания:**

1. **Сборка фронтенда:**
```bash
npm run build
# Создаётся папка dist/ со статическими файлами
```

2. **Настройка GitHub Actions:**
```yaml
# .github/workflows/deploy.yml
name: Deploy to GitHub Pages

on:
  push:
    branches: [ main ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
      - run: npm install
      - run: npm run build
      - uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./dist
```

3. **Настройка репозитория:**
   - Settings → Pages → Source: GitHub Actions
   - После деплоя сайт доступен по URL: `https://username.github.io/repo-name`

4. **Настройка API URL во фронтенде:**
```javascript
// .env.production
VITE_API_BASE_URL=https://diplom-backend.onrender.com/api/v1
```

#### Развёртывание Backend на Cloud PaaS

**Варианты хостинга для Go backend:**

| Платформа | Бесплатный тариф | Особенности |
|-----------|------------------|-------------|
| **Render** | Да (с ограничениями) | Автоматический деплой из GitHub |
| **Railway** | $5 кредитов/мес | Простая настройка PostgreSQL |
| **Heroku** | Нет (только платно) | Надёжный, но дорогой |
| **Fly.io** | Да (до 3 VM) | Близко к пользователю |

**Пример развёртывания на Render:**

1. **Создать `render.yaml`:**
```yaml
services:
  - type: web
    name: diplom-backend
    env: go
    buildCommand: go build -o server ./cmd/server
    startCommand: ./server
    envVars:
      - key: KINOPOISK_API_KEY
        sync: false
      - key: JWT_SECRET
        generateValue: true
      - key: CACHE_TYPE
        value: memory
```

2. **Подключить PostgreSQL:**
   - Render → New Database → PostgreSQL
   - Скопировать CONNECTION_URL
   - Добавить в переменные окружения

3. **Настроить CORS:**
```env
ALLOWED_ORIGINS=https://username.github.io
```

**Итоговый URL API:**
```
https://diplom-backend.onrender.com/api/v1
```

---

## 6. Тестирование

### 6.1 Postman коллекция

Все endpoints документированы в Postman коллекции:
- `/api/postman/Kinopoisk_API.postman_collection.json`
- 25+ запросов с примерами
- Переменные для токенов

### 6.2 Health Check

```bash
GET /health
# Ответ: {"status": "ok"}

GET /api/v1/health
# Ответ: {
#   "status": "ok",
#   "services": {
#     "server": "ok",
#     "database": "ok",
#     "kinopoisk": "ok"
#   }
# }
```

---

## 7. Безопасность

### 7.1 Защита данных

- Пароли хешируются через `bcrypt` (необратимо)
- JWT токены подписываются секретным ключом (HMAC-SHA256)
- Refresh Token хешируются перед сохранением в БД (SHA-256)
- HTTPS для production (настраивается на уровне reverse proxy)

### 7.2 CORS

- Настройка разрешённых origins через `.env`
- Поддержка credentials для cookies
- Preflight requests (OPTIONS) обрабатываются автоматически

---

## 8. Производительность

### 8.1 Метрики

- Время ответа из кеша: ~1-5 мс
- Время ответа от Kinopoisk API: ~100-500 мс
- RPS (Requests Per Second): до 1000 (без БД)

### 8.2 Оптимизации

- Connection pooling для PostgreSQL (pgxpool)
- Keep-Alive для HTTP соединений с Kinopoisk API
- In-Memory кеш для статических данных
- Индексы в БД для поиска по email/username

---

## 9. Используемые библиотеки Go

| Библиотека | Назначение | Версия |
|------------|------------|--------|
| `gin-gonic/gin` | Веб-фреймворк | v1.9.1 |
| `jackc/pgx/v5` | PostgreSQL драйвер | v5.5.0 |
| `golang-jwt/jwt/v5` | JWT токены | v5.2.0 |
| `golang.org/x/crypto/bcrypt` | Хеширование | v0.17.0 |
| `redis/go-redis/v9` | Redis клиент | v9.3.0 |
| `joho/godotenv` | Загрузка .env | v1.5.1 |
| `spf13/viper` | Конфигурация | v1.18.0 |

---

## 10. Выводы

В ходе разработки дипломного проекта были применены следующие технологии и подходы:

1. **Язык Go** — обеспечил высокую производительность и надёжность
2. **Gin Framework** — упростил создание REST API
3. **PostgreSQL** — надёжное хранение данных пользователей
4. **Redis/In-Memory Cache** — экономия лимитов API на 98%
5. **JWT** — безопасная аутентификация без сессий на сервере
6. **Слоёная архитектура** — разделение ответственности между компонентами
7. **12-Factor App** — конфигурация через переменные окружения
8. **GitHub Pages** — бесплатный хостинг для фронтенда
9. **Render/Railway** — облачный хостинг для backend на Go

**Архитектура развёртывания:**
- **Frontend:** GitHub Pages (статические файлы через GitHub Actions)
- **Backend:** Cloud PaaS (Render/Railway с PostgreSQL)
- **CORS:** Настроен для взаимодействия между доменами

Данный стек технологий позволил создать масштабируемое, производительное и безопасное backend-приложение, готовое к эксплуатации в production-среде с минимальными затратами на инфраструктуру.
