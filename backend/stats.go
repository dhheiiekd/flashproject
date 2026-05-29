package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// SaveSessionHandler отвечает за запись результатов сессии в БД
func SaveSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// ИСПРАВЛЕНО: Достаем реальный ID авторизованного кузнеца
	userID := GetCurrentUserID(r)
	if userID == 0 {
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}

	var session StudySession
	err := json.NewDecoder(r.Body).Decode(&session)
	if err != nil || session.TotalCards == 0 {
		http.Error(w, "Неверные входные данные", http.StatusBadRequest)
		return
	}

	// Записываем сессию именно для того юзера, который её проходил
	_, err = DB.Exec(`
		INSERT INTO study_sessions (user_id, deck_id, score, total_cards) 
		VALUES ($1, $2, $3, $4)`,
		userID, session.DeckID, session.Score, session.TotalCards,
	)
	if err != nil {
		log.Println("Ошибка保存сессии в БД:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetUserStatsHandler собирает уникальную статистику по колодам для дашборда
func GetUserStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// ИСПРАВЛЕНО: Достаем реальный ID авторизованного кузнеца
	userID := GetCurrentUserID(r)
	if userID == 0 {
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}

	// Вытягиваем статистику только для конкретного userID
	rows, err := DB.Query(`
		SELECT DISTINCT ON (s.deck_id) d.title, s.score, s.total_cards 
		FROM study_sessions s
		JOIN decks d ON s.deck_id = d.id
		WHERE s.user_id = $1
		ORDER BY s.deck_id, s.created_at DESC`,
		userID,
	)
	if err != nil {
		log.Println("Ошибка получения статистики из БД:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var passedDecks []PassedDeckStat
	var totalPercentageSum int

	for rows.Next() {
		var title string
		var score, totalCards int

		if err := rows.Scan(&title, &score, &totalCards); err != nil {
			continue
		}

		percentage := int((float64(score) / float64(totalCards)) * 100)

		comment := "Отличный результат! ⚡"
		if percentage < 70 {
			comment = "Стоит повторить 📚"
		} else if percentage == 100 {
			comment = "Идеально выковано! 🔥"
		}

		passedDecks = append(passedDecks, PassedDeckStat{
			DeckTitle:  title,
			Percentage: percentage,
			Comment:    comment,
		})

		totalPercentageSum += percentage
	}

	overallKnowledge := 0
	if len(passedDecks) > 0 {
		overallKnowledge = totalPercentageSum / len(passedDecks)
	}

	response := UserStatsResponse{
		TotalPassedDecks: len(passedDecks),
		OverallKnowledge: overallKnowledge,
		PassedDecks:      passedDecks,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
