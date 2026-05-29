package main

import (
	"encoding/json"
	"net/http"
)

// Структура для приема данных с фронтенда
type UpdateProfileRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// 1. Получение данных текущего пользователя
func GetCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	// ИСПРАВЛЕНО: Вызываем глобальную функцию GetCurrentUserID (с большой буквы) из auth.go
	userID := GetCurrentUserID(r)
	if userID == 0 {
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}

	var username string
	// Ищем в БД имя именно того пользователя, чей ID пришел из куки
	err := DB.QueryRow("SELECT username FROM Users WHERE id = $1", userID).Scan(&username)
	if err != nil {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"username": username})
}

// 2. Обновление профиля
func UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// ИСПРАВЛЕНО: Вызываем глобальную функцию GetCurrentUserID из auth.go
	userID := GetCurrentUserID(r)
	if userID == 0 {
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	if req.Username == "" {
		http.Error(w, "Никнейм не может быть пустым", http.StatusBadRequest)
		return
	}

	// Обновляем никнейм только для текущего авторизованного userID
	_, err := DB.Exec("UPDATE Users SET username = $1 WHERE id = $2", req.Username, userID)
	if err != nil {
		http.Error(w, "Этот никнейм уже занят другим кузнецом", http.StatusConflict)
		return
	}

	// Если пароль пришел не пустой — хэшируем и обновляем именно для текущего userID
	if req.Password != "" {
		hashedPassword := hashPassword(req.Password)

		_, err = DB.Exec("UPDATE Users SET password_hash = $1 WHERE id = $2", hashedPassword, userID)
		if err != nil {
			http.Error(w, "Ошибка сохранения пароля", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Профиль успешно перекован!"))
}
