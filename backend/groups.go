package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// Проверка: является ли текущий пользователь создателем группы
func isGroupCreator(userID int, groupID int) bool {
	var creatorID int
	err := DB.QueryRow("SELECT creator_id FROM groups WHERE id = $1", groupID).Scan(&creatorID)
	if err != nil {
		return false
	}
	return creatorID == userID
}

// Получить базовую информацию о конкретной группе (Название, Описание)
func GetGroupDetailsHandler(w http.ResponseWriter, r *http.Request) {
	groupID, _ := strconv.Atoi(r.URL.Query().Get("group_id"))

	type GroupDetails struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	var gd GroupDetails
	err := DB.QueryRow("SELECT id, name, description FROM groups WHERE id = $1", groupID).Scan(&gd.ID, &gd.Name, &gd.Description)
	if err != nil {
		http.Error(w, "Группа не найдена", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gd)
}

// Получить список групп текущего пользователя
func GetUserGroupsHandler(w http.ResponseWriter, r *http.Request) {
	// ИСПРАВЛЕНО: Берем ID из реальной сессии/куки
	currentUserID := GetCurrentUserID(r)
	if currentUserID == 0 {
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}

	query := `
		SELECT g.id, g.name, g.description 
		FROM groups g
		JOIN group_members gm ON g.id = gm.group_id
		WHERE gm.user_id = $1`

	rows, err := DB.Query(query, currentUserID)
	if err != nil {
		http.Error(w, "Ошибка получения групп", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type GroupInfo struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	var groups []GroupInfo
	for rows.Next() {
		var g GroupInfo
		if err := rows.Scan(&g.ID, &g.Name, &g.Description); err == nil {
			groups = append(groups, g)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}

// Создать новую группу
func CreateGroupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// ИСПРАВЛЕНО: Берем ID из реальной сессии/куки
	currentUserID := GetCurrentUserID(r)
	if currentUserID == 0 {
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")

	if name == "" {
		http.Error(w, "Название группы не может быть пустым", http.StatusBadRequest)
		return
	}

	var groupID int
	err := DB.QueryRow("INSERT INTO groups (name, description, creator_id) VALUES ($1, $2, $3) RETURNING id", name, description, currentUserID).Scan(&groupID)
	if err != nil {
		http.Error(w, "Ошибка создания группы в БД", http.StatusInternalServerError)
		return
	}

	_, err = DB.Exec("INSERT INTO group_members (group_id, user_id, role) VALUES ($1, $2, 'admin')", groupID, currentUserID)
	if err != nil {
		http.Error(w, "Ошибка добавления создателя в группу", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Группа успешно создана!"))
}

// Добавить участника в группу по никнейму
func InviteToGroupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// ИСПРАВЛЕНО: Защищаем эндпоинт авторизацией (приглашать может только авторизованный член клана)
	currentUserID := GetCurrentUserID(r)
	groupID, _ := strconv.Atoi(r.FormValue("group_id"))

	if currentUserID == 0 {
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}

	// Дополнительная проверка безопасности: проверяем, состоит ли сам приглашающий в этой группе
	var isMember bool
	err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id=$1 AND user_id=$2)", groupID, currentUserID).Scan(&isMember)
	if err != nil || !isMember {
		http.Error(w, "Вы не состоите в этом клане, чтобы приглашать туда людей", http.StatusForbidden)
		return
	}

	username := r.FormValue("username")

	var targetUserID int
	err = DB.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&targetUserID)
	if err != nil {
		http.Error(w, "Пользователь с таким никнеймом не найден", http.StatusNotFound)
		return
	}

	_, err = DB.Exec("INSERT INTO group_members (group_id, user_id, role) VALUES ($1, $2, 'member') ON CONFLICT DO NOTHING", groupID, targetUserID)
	if err != nil {
		http.Error(w, "Ошибка добавления в группу", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Пользователь успешно добавлен в группу!"))
}

// Получить список КОЛОД, добавленных конкретно в эту группу
func GetGroupDecksHandler(w http.ResponseWriter, r *http.Request) {
	groupID, _ := strconv.Atoi(r.URL.Query().Get("group_id"))

	query := `
		SELECT d.id, d.title, d.description 
		FROM decks d
		JOIN group_decks gd ON d.id = gd.deck_id
		WHERE gd.group_id = $1`

	rows, err := DB.Query(query, groupID)
	if err != nil {
		http.Error(w, "Ошибка БД", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type SimpleDeck struct {
		ID          int    `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	var decks []SimpleDeck
	for rows.Next() {
		var d SimpleDeck
		if err := rows.Scan(&d.ID, &d.Title, &d.Description); err == nil {
			decks = append(decks, d)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(decks)
}

// Добавить колоду в группу
func AddDeckToGroupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Берем ID из реальной сессии/куки
	currentUserID := GetCurrentUserID(r)
	if currentUserID == 0 {
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}

	groupID, _ := strconv.Atoi(r.FormValue("group_id"))
	deckID, _ := strconv.Atoi(r.FormValue("deck_id"))

	var isMember bool
	err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id=$1 AND user_id=$2)", groupID, currentUserID).Scan(&isMember)
	if err != nil || !isMember {
		http.Error(w, "У вас нет прав для добавления колод в эту группу", http.StatusForbidden)
		return
	}

	// ИСПРАВЛЕНО: Меняем устаревшее is_private=false на актуальное is_public=true
	var isAccessible bool
	query := `SELECT EXISTS(SELECT 1 FROM decks WHERE id=$1 AND (is_public=true OR author_id=$2))`
	err = DB.QueryRow(query, deckID, currentUserID).Scan(&isAccessible)
	if err != nil || !isAccessible {
		http.Error(w, "Колода не найдена или у вас нет к ней доступа (она приватная)", http.StatusBadRequest)
		return
	}

	_, err = DB.Exec("INSERT INTO group_decks (group_id, deck_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", groupID, deckID)
	if err != nil {
		http.Error(w, "Ошибка привязки колоды к группе", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Колода успешно привязана к группе!"))
}

// RemoveDeckFromGroupHandler удаляет/отвязывает колоду из группы (клана)
func RemoveDeckFromGroupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// ИСПРАВЛЕНО: Берем ID из реальной сессии/куки
	currentUserID := GetCurrentUserID(r)
	if currentUserID == 0 {
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}

	groupID, _ := strconv.Atoi(r.FormValue("group_id"))
	deckID, _ := strconv.Atoi(r.FormValue("deck_id"))

	// Удалять колоды может только создатель клана
	if !isGroupCreator(currentUserID, groupID) {
		http.Error(w, "Только создатель клана может отвязывать колоды!", http.StatusForbidden)
		return
	}

	_, err := DB.Exec("DELETE FROM group_decks WHERE group_id = $1 AND deck_id = $2", groupID, deckID)
	if err != nil {
		http.Error(w, "Ошибка при удалении колоды из группы", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Колода успешно убрана из группы."))
}

// Отдельная статистика КТО ПРОШЕЛ КОНКРЕТНУЮ КОЛОДУ
func GetDeckProgressHandler(w http.ResponseWriter, r *http.Request) {
	groupID, _ := strconv.Atoi(r.URL.Query().Get("group_id"))
	deckID, _ := strconv.Atoi(r.URL.Query().Get("deck_id"))

	query := `
		SELECT 
			u.username,
			COALESCE(s.score, 0) as score,
			COALESCE(s.total_cards, 0) as total_cards,
			CASE WHEN s.id IS NOT NULL THEN true ELSE false END as completed
		FROM group_members gm
		JOIN users u ON gm.user_id = u.id
		LEFT JOIN LATERAL (
			SELECT id, score, total_cards 
			FROM study_sessions 
			WHERE user_id = u.id AND deck_id = $2
			ORDER BY score DESC LIMIT 1
		) s ON TRUE
		WHERE gm.group_id = $1`

	rows, err := DB.Query(query, groupID, deckID)
	if err != nil {
		http.Error(w, "Ошибка расчета статистики", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type UserProgress struct {
		Username   string `json:"username"`
		Score      int    `json:"score"`
		TotalCards int    `json:"total_cards"`
		Completed  bool   `json:"completed"`
	}

	var progressList []UserProgress
	for rows.Next() {
		var p UserProgress
		if err := rows.Scan(&p.Username, &p.Score, &p.TotalCards, &p.Completed); err == nil {
			progressList = append(progressList, p)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(progressList)
}

// Список участников группы с флагом проверки прав создателя
func GetGroupMembersHandler(w http.ResponseWriter, r *http.Request) {
	groupID, _ := strconv.Atoi(r.URL.Query().Get("group_id"))

	// ИСПРАВЛЕНО: Берем ID из реальной сессии/куки
	currentUserID := GetCurrentUserID(r)
	if currentUserID == 0 {
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}

	isCreator := isGroupCreator(currentUserID, groupID)

	query := `
		SELECT u.id, u.username, gm.role 
		FROM group_members gm
		JOIN users u ON gm.user_id = u.id
		WHERE gm.group_id = $1`

	rows, err := DB.Query(query, groupID)
	if err != nil {
		http.Error(w, "Ошибка получения участников", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type MemberInfo struct {
		UserID   int    `json:"user_id"`
		Username string `json:"username"`
		Role     string `json:"role"`
	}

	var members []MemberInfo
	for rows.Next() {
		var m MemberInfo
		if err := rows.Scan(&m.UserID, &m.Username, &m.Role); err == nil {
			members = append(members, m)
		}
	}

	response := map[string]interface{}{
		"is_creator": isCreator,
		"members":    members,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Жесткое исключение участника
func RemoveMemberHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// ИСПРАВЛЕНО: Берем ID из реальной сессии/куки
	currentUserID := GetCurrentUserID(r)
	if currentUserID == 0 {
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}

	groupID, _ := strconv.Atoi(r.FormValue("group_id"))
	targetUserID, _ := strconv.Atoi(r.FormValue("user_id"))

	if !isGroupCreator(currentUserID, groupID) {
		http.Error(w, "У вас нет прав администратора! Только создатель группы может исключать участников.", http.StatusForbidden)
		return
	}

	if currentUserID == targetUserID {
		http.Error(w, "Вы не можете удалить самого себя из группы", http.StatusBadRequest)
		return
	}

	_, err := DB.Exec("DELETE FROM group_members WHERE group_id = $1 AND user_id = $2", groupID, targetUserID)
	if err != nil {
		http.Error(w, "Ошибка удаления участника", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Пользователь успешно исключен из клана."))
}

// DeleteGroupHandler полностью удаляет группу (клан) из базы данных
func DeleteGroupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// ИСПРАВЛЕНО: Берем ID из реальной сессии/куки
	currentUserID := GetCurrentUserID(r)
	if currentUserID == 0 {
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}

	groupID, _ := strconv.Atoi(r.FormValue("group_id"))

	if groupID == 0 {
		http.Error(w, "Отсутствует параметр group_id", http.StatusBadRequest)
		return
	}

	// Проверяем, является ли пользователь создателем группы
	if !isGroupCreator(currentUserID, groupID) {
		http.Error(w, "Только создатель клана может полностью удалить его!", http.StatusForbidden)
		return
	}

	// Начинаем транзакцию, чтобы безопасно удалить группу и все её связи
	tx, err := DB.Begin()
	if err != nil {
		log.Println("Ошибка начала транзакции удаления группы:", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Удаляем связи с колодами внутри этой группы
	_, err = tx.Exec("DELETE FROM group_decks WHERE group_id = $1", groupID)
	if err != nil {
		log.Println("Ошибка удаления связей group_decks:", err)
		http.Error(w, "Ошибка при очистке колод группы", http.StatusInternalServerError)
		return
	}

	// Удаляем участников этой группы
	_, err = tx.Exec("DELETE FROM group_members WHERE group_id = $1", groupID)
	if err != nil {
		log.Println("Ошибка удаления участников group_members:", err)
		http.Error(w, "Ошибка при очистке участников группы", http.StatusInternalServerError)
		return
	}

	// Удаляем саму группу
	_, err = tx.Exec("DELETE FROM groups WHERE id = $1", groupID)
	if err != nil {
		log.Println("Ошибка удаления группы из таблицы groups:", err)
		http.Error(w, "Ошибка при удалении группы", http.StatusInternalServerError)
		return
	}

	// Фиксируем изменения в БД
	if err := tx.Commit(); err != nil {
		log.Println("Ошибка коммита транзакции удаления группы:", err)
		http.Error(w, "Ошибка сервера при сохранении изменений", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Клан успешно распущен и удален."))
}
