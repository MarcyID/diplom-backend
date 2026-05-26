-- Откат миграции 001: Удаление таблиц users и sessions
-- Дата: 2026-05-25

-- Удаляем триггер
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Удаляем таблицы (сначала sessions из-за FK)
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS users CASCADE;
