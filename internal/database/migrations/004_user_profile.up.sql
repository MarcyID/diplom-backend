-- Миграция для добавления полей аватара и баннера в таблицу users

ALTER TABLE users
ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(500),
ADD COLUMN IF NOT EXISTS banner_url VARCHAR(500);
