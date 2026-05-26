-- Откат миграции для полей аватара и баннера

ALTER TABLE users
DROP COLUMN IF EXISTS avatar_url,
DROP COLUMN IF EXISTS banner_url;
