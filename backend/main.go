package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
)

// Регулярное выражение для базовой проверки корректности формата email
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,4}$`)

// hashPassword преобразует исходный текстовый пароль в безопасный SHA-256 хэш
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// ==========================================================================
// MIDDLEWARE (ФУНКЦИИ КОНТРОЛЯ ДОСТУПА)
// ==========================================================================

// AuthRequiredForUI защищает HTML-страницы.
// Если пользователь не авторизован, его перенаправляет на страницу входа.
func AuthRequiredForUI(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Ищем куку session_user_id, которую ты создаешь в LoginHandler
		_, err := r.Cookie("session_user_id")
		if err != nil {
			// Куки нет — отправляем на форму входа
			http.Redirect(w, r, "/login.html", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

// AuthRequiredForAPI защищает данные (JSON/API-запросы).
// Если сессия отсутствует, возвращает статус 401 Unauthorized.
func AuthRequiredForAPI(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Ищем куку session_user_id
		_, err := r.Cookie("session_user_id")
		if err != nil {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Unauthorized: Требуется авторизация")
			return
		}
		next(w, r)
	}
}

// registerHandler обрабатывает регистрацию новых пользователей
func registerHandler(w http.ResponseWriter, r *http.Request) {
	// Допускаем только POST-запросы для отправки данных формы
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/register.html", http.StatusSeeOther)
		return
	}

	// Читаем и парсим данные, пришедшие из HTML-формы
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Ошибка чтения данных формы", http.StatusBadRequest)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	password := r.FormValue("password")

	// Проверяем, что пользователь заполнил все обязательные поля
	if username == "" || email == "" || password == "" {
		http.Error(w, "Все поля должны быть заполнены", http.StatusBadRequest)
		return
	}

	// 1. ПРОВЕРКА ФОРМАТА ПОЧТЫ (регулярное выражение)
	if !emailRegex.MatchString(email) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Ошибка: Неверный формат почты! Пример: user@example.com")
		return
	}

	// 2. ПРОВЕРКА В БД: Не занят ли уже данный никнейм
	var userExists bool
	err = DB.QueryRow("SELECT EXISTS(SELECT 1 FROM Users WHERE username=$1)", username).Scan(&userExists)
	if err != nil {
		log.Println("Ошибка при проверке уникальности никнейма:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	if userExists {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusConflict)
		fmt.Fprintf(w, "Ошибка: Никнейм %s уже занят!", username)
		return
	}

	// 3. ПРОВЕРКА В БД: Не занята ли уже эта почта (email)
	var emailExists bool
	err = DB.QueryRow("SELECT EXISTS(SELECT 1 FROM Users WHERE email=$1)", email).Scan(&emailExists)
	if err != nil {
		log.Println("Ошибка при проверке уникальности email:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	if emailExists {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusConflict)
		fmt.Fprintf(w, "Ошибка: Почта %s уже зарегистрирована в системе!", email)
		return
	}

	// Хэшируем пароль перед записью, чтобы не хранить его в открытом виде
	hashedPassword := hashPassword(password)

	// Создаем пользователя в БД, передавая зашифрованный пароль в функцию из database.go
	err = CreateUser(username, email, hashedPassword)
	if err != nil {
		log.Println("Ошибка при вызове CreateUser:", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Ошибка сервера при сохранении пользователя")
		return
	}

	// После успешной регистрации перенаправляем на страницу входа
	http.Redirect(w, r, "/login.html", http.StatusSeeOther)
}

func main() {
	// Инициализируем подключение к базе данных PostgreSQL
	InitDB()

	// --------------------------------------------------------------------------
	// 1. ПУБЛИЧНЫЕ МАРШРУТЫ API (Доступны всем без авторизации)
	// --------------------------------------------------------------------------
	http.HandleFunc("/api/register", registerHandler)
	http.HandleFunc("/api/login", LoginHandler)

	// --------------------------------------------------------------------------
	// 2. ЗАЩИЩЕННЫЕ МАРШРУТЫ API (Закрыты через AuthRequiredForAPI)
	// --------------------------------------------------------------------------

	// Наш новый хендлер для категорий, защищаем его твоей мидлварью
	http.HandleFunc("/api/categories", AuthRequiredForAPI(GetCategoriesHandler))

	http.HandleFunc("/api/dashboard", AuthRequiredForAPI(DashboardHandler))
	http.HandleFunc("/api/decks/create", AuthRequiredForAPI(CreateDeckHandler))
	http.HandleFunc("/api/cards/create", AuthRequiredForAPI(CreateCardHandler))
	http.HandleFunc("/api/logout", AuthRequiredForAPI(LogoutHandler))
	http.HandleFunc("/api/decks/my", AuthRequiredForAPI(GetUserDecksHandler))
	http.HandleFunc("/api/decks/search", AuthRequiredForAPI(SearchDecksHandler))

	// Статистика прохождения
	http.HandleFunc("/api/study-sessions", AuthRequiredForAPI(SaveSessionHandler))
	http.HandleFunc("/api/user/stats", AuthRequiredForAPI(GetUserStatsHandler))

	// Работа с параметрами в URL
	http.HandleFunc("/api/cards/", AuthRequiredForAPI(GetCardsHandler))
	http.HandleFunc("/api/decks/delete/", AuthRequiredForAPI(DeleteDeckHandler))
	http.HandleFunc("/api/cards/delete/", AuthRequiredForAPI(DeleteCardHandler))

	// Управление совместными группами (кланами)
	http.HandleFunc("/api/groups/my", AuthRequiredForAPI(GetUserGroupsHandler))
	http.HandleFunc("/api/groups/create", AuthRequiredForAPI(CreateGroupHandler))
	http.HandleFunc("/api/groups/invite", AuthRequiredForAPI(InviteToGroupHandler))
	http.HandleFunc("/api/groups/remove", AuthRequiredForAPI(RemoveMemberHandler))
	http.HandleFunc("/api/groups/details", AuthRequiredForAPI(GetGroupDetailsHandler))
	http.HandleFunc("/api/groups/members", AuthRequiredForAPI(GetGroupMembersHandler))
	http.HandleFunc("/api/groups/delete", AuthRequiredForAPI(DeleteGroupHandler))

	// Colony внутри групп (LMS-модуль)
	http.HandleFunc("/api/groups/decks", AuthRequiredForAPI(GetGroupDecksHandler))
	http.HandleFunc("/api/groups/decks/add", AuthRequiredForAPI(AddDeckToGroupHandler))
	http.HandleFunc("/api/groups/decks/remove", AuthRequiredForAPI(RemoveDeckFromGroupHandler))
	http.HandleFunc("/api/groups/decks/progress", AuthRequiredForAPI(GetDeckProgressHandler))

	// Профиль пользователя
	http.HandleFunc("/api/user/me", AuthRequiredForAPI(GetCurrentUserHandler))
	http.HandleFunc("/api/user/update-profile", AuthRequiredForAPI(UpdateProfileHandler))

	// --------------------------------------------------------------------------
	// 3. ЗАЩИЩЕННЫЕ СТРАНИЦЫ ИНТЕРФЕЙСА (Закрыты через AuthRequiredForUI)
	// --------------------------------------------------------------------------

	// Режим интерактивного прохождения колоды
	http.HandleFunc("/decks/play", AuthRequiredForUI(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/play_deck.html")
	}))

	// Раздел "Ваши колоды" (персональный список пользователя)
	http.HandleFunc("/decks/my", AuthRequiredForUI(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/my_decks.html")
	}))

	// Содержимое конкретной колоды (карточки)
	http.HandleFunc("/decks/view", AuthRequiredForUI(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/view_cards.html")
	}))

	// Главная страница панели управления с кнопками-меню
	http.HandleFunc("/dashboard", AuthRequiredForUI(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/dashboard.html")
	}))

	// Анимированная страница презентации только что созданной колоды
	http.HandleFunc("/decks/preview", AuthRequiredForUI(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/deck_preview.html")
	}))

	// Кузница карточек (добавление новых вопросов и ответов в колоду)
	http.HandleFunc("/decks/add-cards", AuthRequiredForUI(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/add_cards.html")
	}))

	// Конструктор создания новой колоды
	http.HandleFunc("/decks/create", AuthRequiredForUI(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/create_deck.html")
	}))

	// Раздел совместного доступа через Кузнечные Кланы (Общие группы)
	http.HandleFunc("/decks/public", AuthRequiredForUI(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/public_groups.html")
	}))

	// Страница детальной информации и состава конкретной группы
	http.HandleFunc("/decks/group-details", AuthRequiredForUI(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/group_details.html")
	}))

	// Поиск по глобальной библиотеке публичных колод
	http.HandleFunc("/decks/search", AuthRequiredForUI(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/search_deck.html")
	}))

	// Профиль пользователя
	http.HandleFunc("/profile", AuthRequiredForUI(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/profile.html")
	}))

	// Личная статистика прогресса обучения
	http.HandleFunc("/user/stats", AuthRequiredForUI(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/user_stats.html")
	}))

	// --------------------------------------------------------------------------
	// 4. РАЗДАЧА СТАТИЧЕСКИХ РЕСУРСОВ
	// --------------------------------------------------------------------------
	fs := http.FileServer(http.Dir("../frontend"))
	http.Handle("/", fs)

	// Запуск веб-сервера
	log.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
