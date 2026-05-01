package main

import (
	"fmt"
	"log"
	"net/http"
)

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/register.html", http.StatusSeeOther)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Ошибка чтения данных формы", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if username == "" || email == "" || password == "" {
		http.Error(w, "Все поля должны быть заполнены", http.StatusBadRequest)
		return
	}

	err = CreateUser(username, email, password)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusConflict)
		fmt.Fprintf(w, "Ошибка: Пользователь с почтой %s уже существует!", email)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "Пользователь %s успешно зарегистрирован!", username)
}

func main() {
	// 1. Инициализация БД
	InitDB()

	// 2. Раздача статики (HTML, CSS, Картинки)
	fs := http.FileServer(http.Dir("../frontend"))
	http.Handle("/", fs)

	// 3. Обработчик API
	http.HandleFunc("/api/register", registerHandler)

	fmt.Println("Сервер запущен на http://localhost:8080")
	fmt.Println("Для регистрации перейди: http://localhost:8080/register.html")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Ошибка запуска сервера: ", err)
	}
}
