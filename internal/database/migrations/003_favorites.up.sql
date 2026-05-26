-- Миграция для создания таблицы избранного (фильмы и персоны)

-- Таблица избранного
CREATE TABLE IF NOT EXISTS favorites (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    object_type VARCHAR(10) NOT NULL CHECK (object_type IN ('film', 'person')),
    object_id BIGINT NOT NULL,  -- kinopoiskId фильма или personId персоны
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, object_type, object_id)  -- один объект может быть в избранном только один раз
);

-- Индексы для ускорения поиска
CREATE INDEX idx_favorites_user_id ON favorites(user_id);
CREATE INDEX idx_favorites_object_type ON favorites(object_type);
CREATE INDEX idx_favorites_created_at ON favorites(created_at DESC);
CREATE INDEX idx_favorites_user_type ON favorites(user_id, object_type);
