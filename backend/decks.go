package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// GetCategoriesHandler возвращает JSON-список всех категорий из БД
func GetCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	rows, err := DB.Query("SELECT id, name FROM categories ORDER BY id")
	if err != nil {
		log.Println("Ошибка получения категорий:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			continue
		}
		list = append(list, c)
	}

	if list == nil {
		list = []Category{}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(list)
}

// CreateDeckHandler обрабатывает отправку формы создания новой колоды с учетом категории
func CreateDeckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/decks/create", http.StatusSeeOther)
		return
	}

	authorID := GetCurrentUserID(r)
	if authorID == 0 {
		http.Redirect(w, r, "/login.html", http.StatusSeeOther)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Ошибка чтения формы", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")

	isPrivate := r.FormValue("is_private") == "on"
	isPublic := !isPrivate

	categoryIDStr := r.FormValue("category_id")

	if title == "" {
		http.Error(w, "Название колоды обязательно", http.StatusBadRequest)
		return
	}

	deckID, err := CreateDeck(title, description, authorID, isPublic)
	if err != nil {
		log.Println("Ошибка при сохранении колоды в БД:", err)
		http.Error(w, "Ошибка сервера при создании колоды", http.StatusInternalServerError)
		return
	}

	if categoryIDStr != "" {
		categoryID, err := strconv.Atoi(categoryIDStr)
		if err == nil && categoryID > 0 {
			err = AddDeckCategory(deckID, categoryID)
			if err != nil {
				log.Println("Предупреждение: не удалось связать категорию с колодой:", err)
			}
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/decks/preview?deck_id=%d", deckID), http.StatusSeeOther)
}

// CreateCardHandler обрабатывает отправку формы создания новой карточки
func CreateCardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Ошибка чтения формы", http.StatusBadRequest)
		return
	}

	deckIDStr := r.FormValue("deck_id")
	question := r.FormValue("question")
	answer := r.FormValue("answer")

	deckID, err := strconv.Atoi(deckIDStr)
	if err != nil || deckID == 0 {
		http.Error(w, "Неверный ID колоды", http.StatusBadRequest)
		return
	}

	if question == "" || answer == "" {
		http.Error(w, "Поля вопроса и ответа не могут быть пустыми", http.StatusBadRequest)
		return
	}

	err = SaveCardToDB(deckID, question, answer)
	if err != nil {
		log.Println("Ошибка при сохранении карточки в БД:", err)
		http.Error(w, "Ошибка сервера при сохранении карточки", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/decks/preview?deck_id="+deckIDStr, http.StatusSeeOther)
}

// GetCardsHandler отдает JSON со списком всех карточек в колоде
func GetCardsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	deckIDStr := r.URL.Query().Get("deck_id")
	deckID, err := strconv.Atoi(deckIDStr)
	if err != nil || deckID == 0 {
		http.Error(w, "Неверный ID колоды", http.StatusBadRequest)
		return
	}

	cards, err := GetCardsByDeckID(deckID)
	if err != nil {
		log.Println("Ошибка внутри GetCardsHandler:", err)
		http.Error(w, "Ошибка сервера при получении карточек", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(cards)
}

// GetUserDecksHandler возвращает JSON со всеми колодами авторизованного юзера
func GetUserDecksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	authorID := GetCurrentUserID(r)
	if authorID == 0 {
		http.Error(w, "Неавторизован", http.StatusUnauthorized)
		return
	}

	decks, err := GetDecksByAuthorID(authorID)
	if err != nil {
		http.Error(w, "Ошибка сервера при получении колод", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(decks)
}

// DeleteDeckHandler обрабатывает запрос на удаление колоды
func DeleteDeckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	authorID := GetCurrentUserID(r)
	if authorID == 0 {
		http.Error(w, "Неавторизован", http.StatusUnauthorized)
		return
	}

	deckIDStr := r.URL.Query().Get("deck_id")
	deckID, err := strconv.Atoi(deckIDStr)
	if err != nil || deckID == 0 {
		http.Error(w, "Неверный ID колоды", http.StatusBadRequest)
		return
	}

	err = DeleteDeckFromDB(deckID, authorID)
	if err != nil {
		http.Error(w, "Ошибка при удалении колоды", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteCardHandler обрабатывает DELETE запросы для удаления одной карточки
func DeleteCardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	if GetCurrentUserID(r) == 0 {
		http.Error(w, "Неавторизован", http.StatusUnauthorized)
		return
	}

	cardIDStr := r.URL.Query().Get("card_id")
	cardID, err := strconv.Atoi(cardIDStr)
	if err != nil || cardID == 0 {
		http.Error(w, "Неверный ID карточки", http.StatusBadRequest)
		return
	}

	err = DeleteCardFromDB(cardID)
	if err != nil {
		http.Error(w, "Ошибка сервера при удалении карточки", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// SearchDecksHandler выполняет поиск по публичным колодам с учетом выбранной категории
func SearchDecksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	searchQuery := r.URL.Query().Get("query")
	categoryParam := r.URL.Query().Get("category_id")

	var rows *sql.Rows
	var err error

	if categoryParam != "" {
		query := `
			SELECT d.id, d.title, d.description, d.author_id, d.is_public, d.created_at 
			FROM decks d
			JOIN deck_categories dc ON d.id = dc.deck_id
			WHERE d.is_public = true 
			  AND dc.category_id = $1 
			  AND (d.title ILIKE $2 OR d.description ILIKE $2)
			ORDER BY d.created_at DESC`

		rows, err = DB.Query(query, categoryParam, "%"+searchQuery+"%")
	} else {
		query := `
			SELECT id, title, description, author_id, is_public, created_at 
			FROM decks 
			WHERE is_public = true AND (title ILIKE $1 OR description ILIKE $1)
			ORDER BY created_at DESC`

		rows, err = DB.Query(query, "%"+searchQuery+"%")
	}

	if err != nil {
		http.Error(w, "Ошибка базы данных: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var foundDecks []Deck

	for rows.Next() {
		var d Deck
		err := rows.Scan(&d.ID, &d.Title, &d.Description, &d.AuthorID, &d.IsPublic, &d.CreatedAt)
		if err != nil {
			http.Error(w, "Ошибка чтения данных: "+err.Error(), http.StatusInternalServerError)
			return
		}
		foundDecks = append(foundDecks, d)
	}

	if foundDecks == nil {
		foundDecks = []Deck{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(foundDecks)
}
