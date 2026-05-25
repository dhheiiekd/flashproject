package main

import (
	"encoding/json"
	"net/http"
)

// DashboardHandler обрабатывает запросы к эндпоинту /api/dashboard
func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Устанавливаем заголовок, что возвращаем JSON
	w.Header().Set("Content-Type", "application/json")

	// Получаем реальный ID вошедшего пользователя из куки (функция из твоего auth.go)
	userID := GetCurrentUserID(r)
	if userID == 0 {
		http.Error(w, `{"error": "Пользователь не авторизован"}`, http.StatusUnauthorized)
		return
	}

	username := ""

	// 1. Получаем колоды пользователя
	decks, err := GetUserDecks(userID)
	if err != nil {
		http.Error(w, ` {"error": "Не удалось загрузить колоды"}`, http.StatusInternalServerError)
		return
	}

	// 2. Получаем последние сессии
	sessions, err := GetRecentSessions(userID)
	if err != nil {
		http.Error(w, `{"error": "Не удалось загрузить статистику"} `, http.StatusInternalServerError)
		return
	}

	// 3. Формируем единый ответ
	response := DashboardResponse{
		Username:       username,
		Decks:          decks,
		RecentSessions: sessions,
	}

	// 4. Отправляем JSON фронтенду
	json.NewEncoder(w).Encode(response)
}
