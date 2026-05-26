-- Откат миграции для таблиц подборок

-- Удаляем триггер и функцию
DROP TRIGGER IF EXISTS update_collections_updated_at ON collections;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Удаляем таблицы
DROP TABLE IF EXISTS collection_films;
DROP TABLE IF EXISTS collections;
