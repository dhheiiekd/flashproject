package main

import "time"

// User соответствует таблице users в базе данных
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // "-" значит, что пароль не будет превращаться в JSON для безопасности
	CreatedAt    time.Time `json:"created_at"`
}

// Deck соответствует таблице decks
type Deck struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	AuthorID    int       `json:"author_id"`
	IsPrivate   bool      `json:"is_private"`
	CreatedAt   time.Time `json:"created_at"`
}