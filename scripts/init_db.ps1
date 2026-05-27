# Скрипт создания и настройки БД для diplom-backend (PowerShell)
# Использование: .\scripts\init_db.ps1

$ErrorActionPreference = "Stop"

# Конфигурация
$DB_HOST = $env:DB_HOST ?? "127.0.0.1"
$DB_PORT = $env:DB_PORT ?? "5432"
$DB_USER = $env:DB_USER ?? "diplom"
$DB_PASSWORD = $env:DB_PASSWORD ?? "diplom"
$DB_NAME = $env:DB_NAME ?? "diplom_db"
$SUPER_USER = $env:SUPER_USER ?? "vi.v.zhuravlev"

Write-Host "`n=== Инициализация БД для diplom-backend ===" -ForegroundColor Yellow
Write-Host "Host: $DB_HOST"
Write-Host "Port: $DB_PORT"
Write-Host "Database: $DB_NAME"
Write-Host "User: $DB_USER`n"

# Функция для выполнения SQL
function Invoke-SQL {
    param($User, $Database, $Query)
    $env:PGPASSWORD = $DB_PASSWORD
    & psql -h $DB_HOST -U $User -d $Database -t -c $Query 2>&1
}

# Шаг 1: Создание БД
Write-Host "[1/5] Создание базы данных..." -ForegroundColor Yellow
Invoke-SQL -User $SUPER_USER -Database postgres -Query "CREATE DATABASE $DB_NAME;" | Out-Null
Write-Host "✓ База данных $DB_NAME создана" -ForegroundColor Green

# Шаг 2: Создание пользователя
Write-Host "[2/5] Создание пользователя..." -ForegroundColor Yellow
Invoke-SQL -User $SUPER_USER -Database postgres -Query "CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';" | Out-Null
Invoke-SQL -User $SUPER_USER -Database postgres -Query "ALTER USER $DB_USER WITH PASSWORD '$DB_PASSWORD';" | Out-Null
Write-Host "✓ Пользователь $DB_USER создан" -ForegroundColor Green

# Шаг 3: Предоставление прав
Write-Host "[3/5] Предоставление прав..." -ForegroundColor Yellow
Invoke-SQL -User $SUPER_USER -Database postgres -Query "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;" | Out-Null
Invoke-SQL -User $SUPER_USER -Database postgres -Query "ALTER DATABASE $DB_NAME OWNER TO $DB_USER;" | Out-Null
Write-Host "✓ Права предоставлены" -ForegroundColor Green

# Шаг 4: Создание таблиц
Write-Host "[4/5] Создание таблиц..." -ForegroundColor Yellow

# Удаляем старые таблицы
Invoke-SQL -User $DB_USER -Database $DB_NAME -Query "DROP TABLE IF EXISTS favorites CASCADE;" | Out-Null
Invoke-SQL -User $DB_USER -Database $DB_NAME -Query "DROP TABLE IF EXISTS collection_films CASCADE;" | Out-Null
Invoke-SQL -User $DB_USER -Database $DB_NAME -Query "DROP TABLE IF EXISTS collections CASCADE;" | Out-Null
Invoke-SQL -User $DB_USER -Database $DB_NAME -Query "DROP TABLE IF EXISTS sessions CASCADE;" | Out-Null
Invoke-SQL -User $DB_USER -Database $DB_NAME -Query "DROP TABLE IF EXISTS users CASCADE;" | Out-Null

# Создаём новые
$SQL = @"
-- Таблица пользователей
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    avatar_url VARCHAR(500),
    banner_url VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Таблица сессий
CREATE TABLE sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash VARCHAR(500) NOT NULL UNIQUE,
    user_agent TEXT,
    ip_address VARCHAR(50),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_revoked BOOLEAN DEFAULT FALSE
);

-- Таблица подборок
CREATE TABLE collections (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    is_public BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Таблица фильмов в подборках
CREATE TABLE collection_films (
    id BIGSERIAL PRIMARY KEY,
    collection_id BIGINT NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    film_id BIGINT NOT NULL,
    position INTEGER NOT NULL DEFAULT 0,
    added_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(collection_id, film_id)
);

-- Таблица избранного
CREATE TABLE favorites (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    object_type VARCHAR(10) NOT NULL CHECK (object_type IN ('film', 'person')),
    object_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, object_type, object_id)
);

-- Индексы
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_refresh_token ON sessions(refresh_token_hash);
CREATE INDEX idx_collections_user_id ON collections(user_id);
CREATE INDEX idx_collection_films_collection_id ON collection_films(collection_id);
CREATE INDEX idx_favorites_user_id ON favorites(user_id);
"@

Invoke-SQL -User $DB_USER -Database $DB_NAME -Query $SQL | Out-Null
Write-Host "✓ Таблицы созданы" -ForegroundColor Green

# Шаг 5: Проверка
Write-Host "`n[5/5] Проверка таблиц..." -ForegroundColor Yellow
Invoke-SQL -User $DB_USER -Database $DB_NAME -Query "\dt"

Write-Host "`n=== Инициализация завершена! ===" -ForegroundColor Green
Write-Host "`nТеперь запусти сервер:"
Write-Host "  go run cmd/server/main.go`n"
