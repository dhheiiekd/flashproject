-- 1. Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL, 
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. Таблица колод (ОБНОВЛЕНО: теперь поле называется is_public)
CREATE TABLE IF NOT EXISTS decks (
    id SERIAL PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    description TEXT,
    author_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    is_public BOOLEAN DEFAULT true, -- true = публичная, false = приватная
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. Cards — сами карточки
CREATE TABLE IF NOT EXISTS cards (
    id SERIAL PRIMARY KEY,
    deck_id INTEGER REFERENCES decks(id) ON DELETE CASCADE,
    front_text TEXT NOT NULL,
    back_text TEXT NOT NULL,
    image_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 4. Categories — тематические теги (Медицина, Языки, История)
CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL
);

-- 5. Deck_Categories — связь «многие ко многим» между колодами и категориями
CREATE TABLE IF NOT EXISTS deck_categories (
    deck_id INTEGER REFERENCES decks(id) ON DELETE CASCADE,
    category_id INTEGER REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (deck_id, category_id)
);

-- 6. Study_Sessions — логи каждой попытки прохождения сессии обучения
CREATE TABLE IF NOT EXISTS study_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    deck_id INTEGER REFERENCES decks(id) ON DELETE CASCADE,
    score INTEGER NOT NULL,          
    total_cards INTEGER NOT NULL,    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 7. Card_Performance — детальная статистика по каждой карточке
CREATE TABLE IF NOT EXISTS card_performance (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    card_id INTEGER REFERENCES cards(id) ON DELETE CASCADE,
    correct_count INTEGER DEFAULT 0,   
    incorrect_count INTEGER DEFAULT 0, 
    last_reviewed TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_user_card UNIQUE (user_id, card_id)
);

-- 8. Groups — учебные группы для совместного доступа
CREATE TABLE IF NOT EXISTS groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    description TEXT,
    creator_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 9. Group_Members — список пользователей, состоящих в конкретных группах
CREATE TABLE IF NOT EXISTS group_members (
    group_id INTEGER REFERENCES groups(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) DEFAULT 'member', 
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (group_id, user_id)   
);

-- 10. Таблица для точечной привязки колод к группам (Многие-ко-многим)
CREATE TABLE IF NOT EXISTS group_decks (
    group_id INTEGER REFERENCES groups(id) ON DELETE CASCADE,
    deck_id INTEGER REFERENCES decks(id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, deck_id)
);