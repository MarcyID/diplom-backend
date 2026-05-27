@echo off
REM Скрипт создания и настройки БД для diplom-backend (Windows)
REM Использование: init_db.bat

setlocal enabledelayedexpansion

REM Конфигурация из .env или значения по умолчанию
if "%DB_HOST%"=="" set DB_HOST=localhost
if "%DB_PORT%"=="" set DB_PORT=5432
if "%DB_USER%"=="" set DB_USER=diplom
if "%DB_PASSWORD%"=="" set DB_PASSWORD=diplom
if "%DB_NAME%"=="" set DB_NAME=diplom_db
if "%SUPER_USER%"=="" set SUPER_USER=postgres

REM Устанавливаем PGPASSWORD для аутентификации без ввода пароля
set PGPASSWORD=%DB_PASSWORD%

echo.
echo === Инициализация БД для diplom-backend ===
echo Host: %DB_HOST%
echo Port: %DB_PORT%
echo Database: %DB_NAME%
echo User: %DB_USER%
echo.

REM Шаг 1: Создание базы данных
echo [1/4] Создание базы данных...
psql -h %DB_HOST% -U %SUPER_USER% -d postgres -c "CREATE DATABASE %DB_NAME%;" 2>nul
if %errorlevel% neq 0 (
    echo База данных уже существует или ошибка подключения
)

REM Шаг 2: Создание пользователя и предоставление прав
echo [2/4] Создание пользователя и предоставление прав...
psql -h %DB_HOST% -U %SUPER_USER% -d postgres -c "CREATE USER %DB_USER% WITH PASSWORD '%DB_PASSWORD%';" 2>nul
psql -h %DB_HOST% -U %SUPER_USER% -d postgres -c "ALTER USER %DB_USER% WITH PASSWORD '%DB_PASSWORD%';" 2>nul
psql -h %DB_HOST% -U %SUPER_USER% -d postgres -c "GRANT ALL PRIVILEGES ON DATABASE %DB_NAME% TO %DB_USER%;"
psql -h %DB_HOST% -U %SUPER_USER% -d postgres -c "ALTER DATABASE %DB_NAME% OWNER TO %DB_USER%;"

REM Шаг 3: Создание таблиц
echo [3/4] Создание таблиц...
psql -h %DB_HOST% -U %DB_USER% -d %DB_NAME% -c "DROP TABLE IF EXISTS favorites CASCADE;"
psql -h %DB_HOST% -U %DB_USER% -d %DB_NAME% -c "DROP TABLE IF EXISTS collection_films CASCADE;"
psql -h %DB_HOST% -U %DB_USER% -d %DB_NAME% -c "DROP TABLE IF EXISTS collections CASCADE;"
psql -h %DB_HOST% -U %DB_USER% -d %DB_NAME% -c "DROP TABLE IF EXISTS sessions CASCADE;"
psql -h %DB_HOST% -U %DB_USER% -d %DB_NAME% -c "DROP TABLE IF EXISTS users CASCADE;"

psql -h %DB_HOST% -U %DB_USER% -d %DB_NAME% -c ^
"CREATE TABLE users (^
    id BIGSERIAL PRIMARY KEY,^
    email VARCHAR(255) NOT NULL UNIQUE,^
    username VARCHAR(100) NOT NULL UNIQUE,^
    password_hash VARCHAR(255) NOT NULL,^
    full_name VARCHAR(255),^
    avatar_url VARCHAR(500),^
    banner_url VARCHAR(500),^
    genre_preferences BIGINT[] DEFAULT '{}',^
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,^
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP^
);"

psql -h %DB_HOST% -U %DB_USER% -d %DB_NAME% -c ^
"CREATE TABLE sessions (^
    id BIGSERIAL PRIMARY KEY,^
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,^
    refresh_token_hash VARCHAR(500) NOT NULL UNIQUE,^
    user_agent TEXT,^
    ip_address VARCHAR(50),^
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,^
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,^
    is_revoked BOOLEAN DEFAULT FALSE^
);"

psql -h %DB_HOST% -U %DB_USER% -d %DB_NAME% -c ^
"CREATE TABLE collections (^
    id BIGSERIAL PRIMARY KEY,^
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,^
    title VARCHAR(255) NOT NULL,^
    description TEXT,^
    is_public BOOLEAN DEFAULT TRUE,^
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,^
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP^
);"

psql -h %DB_HOST% -U %DB_USER% -d %DB_NAME% -c ^
"CREATE TABLE collection_films (^
    id BIGSERIAL PRIMARY KEY,^
    collection_id BIGINT NOT NULL REFERENCES collections(id) ON DELETE CASCADE,^
    film_id BIGINT NOT NULL,^
    position INTEGER NOT NULL DEFAULT 0,^
    added_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,^
    UNIQUE(collection_id, film_id)^
);"

psql -h %DB_HOST% -U %DB_USER% -d %DB_NAME% -c ^
"CREATE TABLE favorites (^
    id BIGSERIAL PRIMARY KEY,^
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,^
    object_type VARCHAR(10) NOT NULL CHECK (object_type IN ('film', 'person')),^
    object_id BIGINT NOT NULL,^
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,^
    UNIQUE(user_id, object_type, object_id)^
);"

REM Шаг 4: Создание индексов
echo [4/4] Создание индексов...
psql -h %DB_HOST% -U %DB_USER% -d %DB_NAME% -c "CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);"
psql -h %DB_HOST% -U %DB_USER% -d %DB_NAME% -c "CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);"
psql -h %DB_HOST% -U %DB_USER% -d %DB_NAME% -c "CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);"
psql -h %DB_HOST% -U %DB_USER% -d %DB_NAME% -c "CREATE INDEX IF NOT EXISTS idx_sessions_refresh_token_hash ON sessions(refresh_token_hash);"
psql -h %DB_HOST% -U %DB_USER% -d %DB_NAME% -c "CREATE INDEX IF NOT EXISTS idx_collections_user_id ON collections(user_id);"
psql -h %DB_HOST% -U %DB_USER% -d %DB_NAME% -c "CREATE INDEX IF NOT EXISTS idx_collections_created_at ON collections(created_at DESC);"
psql -h %DB_HOST% -U %DB_USER% -d %DB_NAME% -c "CREATE INDEX IF NOT EXISTS idx_collection_films_collection_id ON collection_films(collection_id);"
psql -h %DB_HOST% -U %DB_USER% -d %DB_NAME% -c "CREATE INDEX IF NOT EXISTS idx_favorites_user_id ON favorites(user_id);"
psql -h %DB_HOST% -U %DB_USER% -d %DB_NAME% -c "CREATE INDEX IF NOT EXISTS idx_favorites_user_type ON favorites(user_id, object_type);"

REM Проверка
echo.
echo === Проверка таблиц... ===
psql -h %DB_HOST% -U %DB_USER% -d %DB_NAME% -c "\dt"

echo.
echo === Инициализация завершена! ===
echo.
echo Теперь запусти сервер:
echo   go run cmd/server/main.go
echo.
pause
