package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("Предупреждение: файл .env не найден, используются системные переменные")
	}

	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, name)

	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Ошибка открытия базы: ", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal("Ошибка подключения к базе (Ping): ", err)
	}

	fmt.Println("Успешное подключение к базе данных!")
}

func CreateUser(username, email, password string) error {
	query := `INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3)`
	_, err := DB.Exec(query, username, email, password)
	return err
}

func GetUserDecks(userID int) ([]Deck, error) {
	query := `SELECT id, title, description, author_id, is_public, created_at FROM decks WHERE author_id = $1`
	rows, err := DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decks []Deck
	for rows.Next() {
		var d Deck
		err := rows.Scan(&d.ID, &d.Title, &d.Description, &d.AuthorID, &d.IsPublic, &d.CreatedAt)
		if err != nil {
			return nil, err
		}
		decks = append(decks, d)
	}
	return decks, nil
}

func GetRecentSessions(userID int) ([]StudySession, error) {
	query := `
		SELECT s.id, s.user_id, s.deck_id, d.title, s.score, s.total_cards, s.created_at 
		FROM study_sessions s
		JOIN decks d ON s.deck_id = d.id
		WHERE s.user_id = $1
		ORDER BY s.created_at DESC
		LIMIT 5`

	rows, err := DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []StudySession
	for rows.Next() {
		var s StudySession
		err := rows.Scan(&s.ID, &s.UserID, &s.DeckID, &s.DeckTitle, &s.Score, &s.TotalCards, &s.CreatedAt)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func CreateDeck(title, description string, authorID int, isPublic bool) (int, error) {
	var lastInsertId int

	query := `
		INSERT INTO decks (title, description, author_id, is_public, created_at) 
		VALUES ($1, $2, $3, $4, NOW()) 
		RETURNING id`

	err := DB.QueryRow(query, title, description, authorID, isPublic).Scan(&lastInsertId)
	if err != nil {
		return 0, err
	}

	return lastInsertId, nil
}

func AddDeckCategory(deckID, categoryID int) error {
	query := `INSERT INTO deck_categories (deck_id, category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := DB.Exec(query, deckID, categoryID)
	return err
}

func CreateCard(deckID int, question, answer string) error {
	query := `
		INSERT INTO cards (deck_id, question, answer, created_at)
		VALUES ($1, $2, $3, NOW())`

	_, err := DB.Exec(query, deckID, question, answer)
	return err
}

func SaveCardToDB(deckID int, frontText, backText string) error {
	query := `
		INSERT INTO cards (deck_id, front_text, back_text, created_at) 
		VALUES ($1, $2, $3, NOW())`

	_, err := DB.Exec(query, deckID, frontText, backText)
	return err
}

func GetCardsByDeckID(deckID int) ([]Card, error) {
	query := `
		SELECT id, deck_id, front_text, back_text, COALESCE(image_url, ''), created_at 
		FROM cards 
		WHERE deck_id = $1 
		ORDER BY id DESC`

	rows, err := DB.Query(query, deckID)
	if err != nil {
		log.Println("Ошибка выполнения SQL-запроса карт:", err)
		return nil, err
	}
	defer rows.Close()

	var cards []Card
	for rows.Next() {
		var d Card
		err := rows.Scan(&d.ID, &d.DeckID, &d.FrontText, &d.BackText, &d.ImageURL, &d.CreatedAt)
		if err != nil {
			log.Println("Ошибка при сканировании строки карточки (rows.Scan):", err)
			return nil, err
		}
		cards = append(cards, d)
	}

	if cards == nil {
		cards = []Card{}
	}

	return cards, nil
}

func GetDecksByAuthorID(authorID int) ([]Deck, error) {
	query := `SELECT id, title, description, author_id, is_public, created_at FROM decks WHERE author_id = $1 ORDER BY created_at DESC`

	rows, err := DB.Query(query, authorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decks []Deck
	for rows.Next() {
		var d Deck
		err := rows.Scan(&d.ID, &d.Title, &d.Description, &d.AuthorID, &d.IsPublic, &d.CreatedAt)
		if err != nil {
			return nil, err
		}
		decks = append(decks, d)
	}

	if decks == nil {
		decks = []Deck{}
	}
	return decks, nil
}

func DeleteDeckFromDB(deckID, authorID int) error {
	query := `DELETE FROM decks WHERE id = $1 AND author_id = $2`
	_, err := DB.Exec(query, deckID, authorID)
	return err
}

func DeleteCardFromDB(cardID int) error {
	query := `DELETE FROM cards WHERE id = $1`
	_, err := DB.Exec(query, cardID)
	return err
}
