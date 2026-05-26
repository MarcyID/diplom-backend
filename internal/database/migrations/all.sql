-- Все миграции для diplom_db
-- Выполнять от имени пользователя diplom

-- 001_init_auth.up.sql
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    avatar_url VARCHAR(500),
    banner_url VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token VARCHAR(500) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_refresh_token ON sessions(refresh_token);

-- 002_collections.up.sql
CREATE TABLE IF NOT EXISTS collections (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    is_public BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_collections_user_id ON collections(user_id);
CREATE INDEX idx_collections_is_public ON collections(is_public);
CREATE INDEX idx_collections_created_at ON collections(created_at DESC);

CREATE TABLE IF NOT EXISTS collection_films (
    id BIGSERIAL PRIMARY KEY,
    collection_id BIGINT NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    film_id BIGINT NOT NULL,
    position INTEGER NOT NULL DEFAULT 0,
    added_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(collection_id, film_id)
);

CREATE INDEX idx_collection_films_collection_id ON collection_films(collection_id);
CREATE INDEX idx_collection_films_position ON collection_films(collection_id, position);

-- 003_favorites.up.sql
CREATE TABLE IF NOT EXISTS favorites (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    object_type VARCHAR(10) NOT NULL CHECK (object_type IN ('film', 'person')),
    object_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, object_type, object_id)
);

CREATE INDEX idx_favorites_user_id ON favorites(user_id);
CREATE INDEX idx_favorites_object_type ON favorites(object_type);
CREATE INDEX idx_favorites_created_at ON favorites(created_at DESC);
CREATE INDEX idx_favorites_user_type ON favorites(user_id, object_type);
