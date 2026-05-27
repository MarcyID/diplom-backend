# Postman Collection - Diplom Backend

Коллекция Postman для тестирования API дипломного проекта.

## 📦 Установка

1. Откройте Postman
2. Нажмите **Import** (в левом верхнем углу)
3. Выберите файл `Kinopoisk_API.postman_collection.json`
4. Коллекция появится в списке

## 🔧 Настройка

Коллекция использует переменные:
- `base_url` = `http://localhost:5454` (по умолчанию)
- `access_token` = JWT access token (устанавливается автоматически после login)
- `refresh_token` = JWT refresh token (устанавливается автоматически после login)

Измените `base_url` в настройках коллекции, если сервер запущен на другом порту.

## 📋 Доступные запросы

### 🔐 Authentication
| Запрос | Метод | Описание |
|--------|-------|----------|
| Register | POST | Регистрация нового пользователя |
| Login | POST | Вход в систему (возвращает токены) |
| Get Me | GET | Данные текущего пользователя (требует токен) |
| Logout | POST | Выход из системы (требует токен) |
| Refresh Token | POST | Обновление пары токенов |

### 👤 Profile
| Запрос | Метод | Описание |
|--------|-------|----------|
| Get My Profile | GET | Получить профиль текущего пользователя |
| Update Profile | PUT | Обновить full_name, avatar_url, banner_url |
| Get Genre Preferences | GET | Получить жанровые предпочтения (ID жанров) |
| Update Genre Preferences | PUT | Обновить жанровые предпочтения |
| Upload Avatar | POST | Загрузить аватар (multipart/form-data) |
| Upload Banner | POST | Загрузить фон (multipart/form-data) |
| Delete Avatar | DELETE | Удалить аватар |
| Delete Banner | DELETE | Удалить фон |

### ⭐ Favorites
| Запрос | Метод | Описание |
|--------|-------|----------|
| Get Favorites | GET | Получить избранное (фильмы и персоны) |
| Toggle Film Favorite | POST | Добавить/удалить фильм из избранного |
| Toggle Person Favorite | POST | Добавить/удалить персону из избранного |

### 📚 Collections
| Запрос | Метод | Описание |
|--------|-------|----------|
| Get My Collections | GET | Получить все подборки текущего пользователя |
| Create Collection | POST | Создать новую подборку |
| Get Collection by ID | GET | Получить подборку по ID с фильмами |
| Update Collection | PUT | Обновить подборку |
| Delete Collection | DELETE | Удалить подборку |
| Add Film to Collection | POST | Добавить фильм в подборку (`/collections/:id/films`) |
| Remove Film from Collection | DELETE | Удалить фильм из подборки (`/collections/:id/films/:filmId`) |
| Reorder Films | PUT | Изменить порядок фильмов в подборке (`/collections/:id/films/reorder`) |
| Get User's Public Collections | GET | Публичные подборки пользователя (без авторизации) |

### 🎬 Films
| Запрос | Метод | Описание |
|--------|-------|----------|
| Search Films | GET | Поиск фильмов по названию и фильтрам |
| Get Film by ID | GET | Подробная информация о фильме |
| Get Random Film | GET | Случайный фильм (с фильтром по жанру и рейтингу) |
| Get Similar Films by ID | GET | Похожие фильмы по ID |
| Find Similar by Title | GET | Похожие фильмы по названию |
| Get Popular Films | GET | Популярные фильмы (топ) |
| Get Upcoming Films | GET | Предстоящие премьеры |
| Get Premieres | GET | Премьеры за указанный год и месяц |
| Get Genres and Countries | GET | Список жанров и стран для фильтров |

### 🎭 Cast & Crew
| Запрос | Метод | Описание |
|--------|-------|----------|
| Get Staff by Film ID | GET | Актёры и режиссёры фильма |
| Get Person by ID | GET | Данные о персоне (актёр, режиссёр) |
| Search Actors | GET | Поиск актёров по имени |
| Get Actor by ID | GET | Данные актёра по ID (возвращает фильмографию в поле `films`) |

### 📦 Cache
| Запрос | Метод | Описание |
|--------|-------|----------|
| Get Cache Stats | GET | Статистика кеша (для отладки) |
| Clear Cache | POST | Очистка всего кеша (для тестирования) |

### ⚙️ System
| Запрос | Метод | Описание |
|--------|-------|----------|
| Root Endpoint | GET | Базовая информация о сервере |
| Health Check | GET | Базовая проверка здоровья |
| API Health Check | GET | Расширенная проверка (БД, API) |
| API Info | GET | Список всех endpoints |

### Examples (готовые примеры)
- Popular Films 2023 — фильмы 2023 года с рейтингом > 7.0
- Comedy Films — комедии
- Top Rated Sci-Fi — лучшая фантастика (рейтинг > 8.0)

## 🚀 Быстрый старт

1. Запустите сервер:
   ```bash
   go run cmd/server/main.go
   ```

2. Проверьте работу API:
   - Откройте коллекцию в Postman
   - Выполните запрос **Health Check**
   - Должен вернуться статус "ok"

3. Протестируйте аутентификацию:
   - Выполните **Register** для создания пользователя
   - Выполните **Login** для получения токенов
   - Токены автоматически сохранятся в переменные коллекции
   - Выполните **Get Me** для проверки авторизации

4. Протестируйте кеширование:
   - Выполните **Get Popular Films** (первый запрос - Cache MISS)
   - Выполните повторно (второй запрос - Cache HIT)
   - Проверьте логи сервера

## 💡 Советы

- **Аутентификация:** После Login токены автоматически подставляются в запросы
- **Кеширование:** Все запросы к Kinopoisk API кешируются на 24 часа
- **Очистка кеша:** Используйте **Clear Cache** для тестирования без кеша

## 📝 Параметры поиска

**Для поиска фильмов (`/api/v1/movies/search`):**

| Параметр | Описание | Пример |
|----------|----------|--------|
| `q` | Поисковый запрос | `матрица` |
| `genre` | ID жанра (можно несколько) | `4` (комедия), `1` (боевик) |
| `country` | ID страны | `1` (США), `2` (Россия) |
| `year_from` | Год от | `2020` |
| `year_to` | Год до | `2024` |
| `rating_min` | Мин. рейтинг | `7.5` |
| `rating_max` | Макс. рейтинг | `10` |
| `page` | Номер страницы | `1` |

**Для случайного фильма (`/api/v1/movies/random`):**

| Параметр | Описание | Пример |
|----------|----------|--------|
| `genre` | ID жанра (можно несколько) | `15` (мультфильм), `4` (комедия) |
| `min_rating` | Минимальный рейтинг | `7.5` |

**Примеры:**
- `/api/v1/movies/random?genre=15` — случайный мультфильм
- `/api/v1/movies/random?genre=4&min_rating=7.0` — случайная комедия с рейтингом от 7.0
- `/api/v1/movies/random?genre=17&genre=13` — фантастика или триллер

**ID популярных жанров:**
- `1` — Триллер
- `2` — Драма
- `3` — Криминал
- `4` — Мелодрама
- `5` — Детектив
- `6` — Фантастика
- `11` — Боевик
- `13` — Комедия
- `15` — История
- `17` — Ужасы
- `18` — Мультфильм
- `19` — Семейный
- `22` — Документальный

**Полный список:** `GET /api/v1/genres`

## 🔑 API Key

API ключ Kinopoisk должен быть указан в файле `.env` сервера:
```
KINOPOISK_API_KEY=ваш_ключ
```

## 📚 Документация

- [Kinopoisk API Unofficial Docs](https://kinopoiskapiunofficial.tech/documentation/api)
- OpenAPI spec: `/api/openapi/kinopoisk/openapi.json`
- Технологии: `/TECHNOLOGIES.md`
- Краткая справка: `/DIPLOMA_SUMMARY.md`

## 🛠 Технологии проекта

- **Backend:** Go 1.21 + Gin Framework
- **Database:** PostgreSQL 15
- **Cache:** Redis/In-Memory (TTL 24 часа)
- **Auth:** JWT (Access 15 мин, Refresh 7 дней)
- **Deployment:** GitHub Pages (frontend), Render (backend)
