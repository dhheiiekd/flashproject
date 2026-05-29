package main

import "time"

// User соответствует таблице users в базе данных
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// Deck соответствует таблице decks
type Deck struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	AuthorID    int       `json:"author_id"`
	IsPublic    bool      `json:"is_public"`
	CreatedAt   time.Time `json:"created_at"`
}

// Card соответствует таблице cards
type Card struct {
	ID        int       `json:"id"`
	DeckID    int       `json:"deck_id"`
	FrontText string    `json:"front_text"`
	BackText  string    `json:"back_text"`
	ImageURL  string    `json:"image_url"`
	CreatedAt time.Time `json:"created_at"`
}

// Category соответствует таблице categories
type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// DeckCategory соответствует связующей таблице deck_categories
type DeckCategory struct {
	DeckID     int `json:"deck_id"`
	CategoryID int `json:"category_id"`
}

// CardPerformance соответствует таблице card_performance
type CardPerformance struct {
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	CardID         int       `json:"card_id"`
	CorrectCount   int       `json:"correct_count"`
	IncorrectCount int       `json:"incorrect_count"`
	LastReviewed   time.Time `json:"last_reviewed"`
}

// StudySession соответствует таблице study_sessions
type StudySession struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	DeckID     int       `json:"deck_id"`
	DeckTitle  string    `json:"deck_title"`
	Score      int       `json:"score"`
	TotalCards int       `json:"total_cards"`
	CreatedAt  time.Time `json:"created_at"`
}

// DashboardResponse — общая структура ответа для фронтенда дашборда
type DashboardResponse struct {
	Username       string         `json:"username"`
	Decks          []Deck         `json:"decks"`
	RecentSessions []StudySession `json:"recent_sessions"`
}

// PassedDeckStat описывает агрегированную строку для списка пройденных колод
type PassedDeckStat struct {
	DeckTitle  string `json:"deck_title"`
	Percentage int    `json:"percentage"`
	Comment    string `json:"comment"`
}

// UserStatsResponse возвращает итоговую аналитику
type UserStatsResponse struct {
	TotalPassedDecks int              `json:"total_passed_decks"`
	OverallKnowledge int              `json:"overall_knowledge"`
	PassedDecks      []PassedDeckStat `json:"passed_decks"`
}

// Group описывает учебное сообщество
type Group struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatorID   int       `json:"creator_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// GroupMember описывает участника группы и его роль
type GroupMember struct {
	GroupID  int       `json:"group_id"`
	UserID   int       `json:"user_id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

// GroupProgress хранит данные о том, как член группы справился с колодой
type GroupProgress struct {
	Username   string `json:"username"`
	DeckTitle  string `json:"deck_title"`
	Score      int    `json:"score"`
	TotalCards int    `json:"total_cards"`
	Percentage int    `json:"percentage"`
	StatusText string `json:"status_text"`
}
