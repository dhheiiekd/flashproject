package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

// LoginHandler обрабатывает вход пользователя по никнейму и хэшированному паролю
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Если это не POST, а обычный переход по ссылке — просто отдаем HTML-страницу входа
	if r.Method != http.MethodPost {
		http.ServeFile(w, r, "../frontend/login.html")
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	var userID int
	var dbPasswordHash string

	// Ищем ID и Хэш пароля в базе по никнейму (Используем Users с большой буквы, как в main.go)
	query := `SELECT id, password_hash FROM Users WHERE username = $1`
	err := DB.QueryRow(query, username).Scan(&userID, &dbPasswordHash)

	if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Ошибка: Пользователь с таким никнеймом не найден")
		} else {
			log.Println("Ошибка при поиске пользователя в БД:", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Ошибка сервера")
		}
		return
	}

	// Хэшируем введенный пароль (вызывается функция hashPassword из main.go) и сверяем его с хэшем из БД
	if dbPasswordHash == hashPassword(password) {
		// СЕССИЯ: Создаем Cookie и записываем туда ID вошедшего пользователя
		cookie := &http.Cookie{
			Name:     "session_user_id",
			Value:    strconv.Itoa(userID),
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour), // Сессия активна 1 день
			HttpOnly: true,                           // Защита от кражи кук через сторонние JS-скрипты (XSS)
		}
		http.SetCookie(w, cookie)

		// Перенаправляем на главный дашборд (маршрут /dashboard из main.go)
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	} else {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Ошибка: Неверный пароль")
	}
}

// GetCurrentUserID читает Cookie и возвращает ID авторизованного пользователя.
// Если куки нет или она неверна — возвращает 0.
func GetCurrentUserID(r *http.Request) int {
	cookie, err := r.Cookie("session_user_id")
	if err != nil {
		return 0
	}
	userID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		return 0
	}
	return userID
}

// LogoutHandler очищает куку сессии при выходе из аккаунта
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:     "session_user_id",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour), // Установка времени в прошлом принудительно удаляет куку в браузере
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)

	// Перенаправляем на страницу входа через статический файл
	http.Redirect(w, r, "/login.html", http.StatusSeeOther)
}
