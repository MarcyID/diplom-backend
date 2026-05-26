#!/bin/bash

# Скрипт создания и настройки БД для diplom-backend
# Использование: ./scripts/init_db.sh

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Конфигурация
DB_HOST="${DB_HOST:-127.0.0.1}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-diplom}"
DB_PASSWORD="${DB_PASSWORD:-diplom}"
DB_NAME="${DB_NAME:-diplom_db}"
SUPER_USER="${SUPER_USER:-vi.v.zhuravlev}"

echo -e "${YELLOW}=== Инициализация БД для diplom-backend ===${NC}"
echo "Host: $DB_HOST"
echo "Port: $DB_PORT"
echo "Database: $DB_NAME"
echo "User: $DB_USER"
echo ""

# Функция для выполнения SQL от имени суперпользователя
run_as_super() {
    PGPASSWORD="" psql -h "$DB_HOST" -U "$SUPER_USER" -d postgres -t -c "$1" 2>/dev/null
}

# Функция для выполнения SQL от имени пользователя БД
run_as_user() {
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -t -c "$1"
}

# Шаг 1: Проверка подключения
echo -e "${YELLOW}[1/5] Проверка подключения...${NC}"
if ! run_as_super "SELECT 1;" > /dev/null 2>&1; then
    echo -e "${RED}❌ Не удалось подключиться к PostgreSQL от имени суперпользователя '$SUPER_USER'${NC}"
    echo "Попробуй указать правильного суперпользователя:"
    echo "  SUPER_USER=your_user $0"
    exit 1
fi
echo -e "${GREEN}✓ Подключение к PostgreSQL успешно${NC}"

# Шаг 2: Создание базы данных (если не существует)
echo -e "${YELLOW}[2/5] Создание базы данных...${NC}"
run_as_super "CREATE DATABASE $DB_NAME;" 2>/dev/null || echo "База данных уже существует"
echo -e "${GREEN}✓ База данных $DB_NAME готова${NC}"

# Шаг 3: Создание пользователя (если не существует)
echo -e "${YELLOW}[3/5] Проверка пользователя...${NC}"
run_as_super "CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';" 2>/dev/null || echo "Пользователь уже существует"
run_as_super "ALTER USER $DB_USER WITH PASSWORD '$DB_PASSWORD';"
echo -e "${GREEN}✓ Пользователь $DB_USER готов${NC}"

# Шаг 4: Предоставление прав
echo -e "${YELLOW}[4/5] Предоставление прав...${NC}"
run_as_super "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;"
run_as_super "ALTER DATABASE $DB_NAME OWNER TO $DB_USER;"
run_as_super "\\c $DB_NAME" "GRANT ALL ON SCHEMA public TO $DB_USER;"
echo -e "${GREEN}✓ Права предоставлены${NC}"

# Шаг 5: Создание таблиц
echo -e "${YELLOW}[5/5] Создание таблиц...${NC}"

run_as_user "
-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
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
CREATE TABLE IF NOT EXISTS sessions (
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
CREATE TABLE IF NOT EXISTS collections (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    is_public BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Таблица фильмов в подборках
CREATE TABLE IF NOT EXISTS collection_films (
    id BIGSERIAL PRIMARY KEY,
    collection_id BIGINT NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    film_id BIGINT NOT NULL,
    position INTEGER NOT NULL DEFAULT 0,
    added_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(collection_id, film_id)
);

-- Таблица избранного
CREATE TABLE IF NOT EXISTS favorites (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    object_type VARCHAR(10) NOT NULL CHECK (object_type IN ('film', 'person')),
    object_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, object_type, object_id)
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_refresh_token ON sessions(refresh_token_hash);
CREATE INDEX IF NOT EXISTS idx_collections_user_id ON collections(user_id);
CREATE INDEX IF NOT EXISTS idx_collection_films_collection_id ON collection_films(collection_id);
CREATE INDEX IF NOT EXISTS idx_favorites_user_id ON favorites(user_id);
"

echo -e "${GREEN}✓ Таблицы созданы${NC}"

# Проверка результата
echo ""
echo -e "${YELLOW}Проверка таблиц...${NC}"
run_as_user "\\dt"

echo ""
echo -e "${GREEN}=== Инициализация завершена успешно! ===${NC}"
echo ""
echo "Теперь запусти сервер:"
echo "  go run cmd/server/main.go"
echo ""
echo "И зарегистрируйся через API или фронтенд."
