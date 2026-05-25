-- 1. Таблица пользователей
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL, 
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. Таблица колодок
CREATE TABLE decks (
    id SERIAL PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    description TEXT,
    author_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    is_private BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. Cards — сами карточки
CREATE TABLE cards (
    id SERIAL PRIMARY KEY,
    deck_id INTEGER REFERENCES decks(id) ON DELETE CASCADE,
    front_text TEXT NOT NULL,
    back_text TEXT NOT NULL,
    image_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 4. Categories — тематические теги (Медицина, Языки, История)
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL
);

-- 5. Deck_Categories — связь «многие ко многим» между колодами и категориями
CREATE TABLE deck_categories (
    deck_id INTEGER REFERENCES decks(id) ON DELETE CASCADE,
    category_id INTEGER REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (deck_id, category_id) -- составной ключ, чтобы теги не дублировались в одной колоде
);

-- 6. Study_Sessions — логи каждой попытки прохождения сессии обучения
CREATE TABLE study_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    deck_id INTEGER REFERENCES decks(id) ON DELETE CASCADE,
    score INTEGER NOT NULL,          -- сколько раз ответил верно (например, 7)
    total_cards INTEGER NOT NULL,    -- всего карточек в сессии (например, 10)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 7. Card_Performance — детальная статистика по каждой карточке
CREATE TABLE card_performance (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    card_id INTEGER REFERENCES cards(id) ON DELETE CASCADE,
    correct_count INTEGER DEFAULT 0,   -- сколько раз пользователь ответил верно
    incorrect_count INTEGER DEFAULT 0, -- сколько раз пользователь ошибся
    last_reviewed TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_user_card UNIQUE (user_id, card_id) -- одна запись на пару пользователь-карточка
);

-- 8. Groups — учебные группы для совместного доступа
CREATE TABLE groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    description TEXT,
    creator_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 9. Group_Members — список пользователей, состоящих в конкретных группах
CREATE TABLE group_members (
    group_id INTEGER REFERENCES groups(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) DEFAULT 'member', -- роль: member, admin, moderator
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (group_id, user_id)   -- составной ключ, чтобы пользователь не вступал в группу дважды
);