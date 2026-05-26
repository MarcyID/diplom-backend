@echo off
REM Скрипт создания и настройки БД для diplom-backend (Windows)
REM Использование: init_db.bat

setlocal enabledelayedexpansion

REM Конфигурация
set DB_HOST=%DB_HOST:~%127.0.0.1%
set DB_PORT=%DB_PORT:~%5432%
set DB_USER=%DB_USER:~%diplom%
set DB_PASSWORD=%DB_PASSWORD:~%diplom%
set DB_NAME=%DB_NAME:~%diplom_db%
set SUPER_USER=%SUPER_USER:~%vi.v.zhuravlev%

echo.
echo === Инициализация БД для diplom-backend ===
echo Host: %DB_HOST%
echo Port: %DB_PORT%
echo Database: %DB_NAME%
echo User: %DB_USER%
echo.

REM Шаг 1: Создание базы данных
echo [1/5] Создание базы данных...
psql -h %DB_HOST% -U %SUPER_USER% -d postgres -c "CREATE DATABASE %DB_NAME%;" 2>nul
if %errorlevel% neq 0 (
    echo База данных уже существует или ошибка подключения
)

REM Шаг 2: Создание пользователя
echo [2/5] Создание пользователя...
psql -h %DB_HOST% -U %SUPER_USER% -d postgres -c "CREATE USER %DB_USER% WITH PASSWORD '%DB_PASSWORD%';" 2>nul
psql -h %DB_HOST% -U %SUPER_USER% -d postgres -c "ALTER USER %DB_USER% WITH PASSWORD '%DB_PASSWORD%';" 2>nul

REM Шаг 3: Предоставление прав
echo [3/5] Предоставление прав...
psql -h %DB_HOST% -U %SUPER_USER% -d postgres -c "GRANT ALL PRIVILEGES ON DATABASE %DB_NAME% TO %DB_USER%;"
psql -h %DB_HOST% -U %SUPER_USER% -d postgres -c "ALTER DATABASE %DB_NAME% OWNER TO %DB_USER%;"

REM Шаг 4: Создание таблиц
echo [4/5] Создание таблиц...
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

REM Шаг 5: Проверка
echo.
echo [5/5] Проверка таблиц...
psql -h %DB_HOST% -U %DB_USER% -d %DB_NAME% -c "\dt"

echo.
echo === Инициализация завершена! ===
echo.
echo Теперь запусти сервер:
echo   go run cmd/server/main.go
echo.
pause
