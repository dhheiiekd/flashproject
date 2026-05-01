package main

import (
	"database/sql"
	"fmt"
	"log"
	"os" // Нужно для работы с переменными окружения

	"github.com/joho/godotenv" // Библиотека для загрузки .env
	_ "github.com/lib/pq"      // Драйвер для PostgreSQL
)

var DB *sql.DB

func InitDB() {
	// 1. Загружаем файл .env из корня проекта
	// "../.env" означает: выйти из папки backend и найти файл .env в корне
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("Предупреждение: файл .env не найден, используются системные переменные")
	}

	// 2. Получаем данные из переменных окружения
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")

	// 3. Собираем строку подключения
	// Если переменные не заданы в .env, здесь будут пустые строки,
	// поэтому убедись, что в .env всё написано правильно.
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, name)

	// 4. Открываем соединение
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Ошибка открытия базы: ", err)
	}

	// Проверяем соединение
	err = DB.Ping()
	if err != nil {
		log.Fatal("Ошибка подключения к базе (Ping): ", err)
	}

	fmt.Println("Успешное подключение к базе данных через .env!")
}

func CreateUser(username, email, password string) error {
	query := ` INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3)`
	_, err := DB.Exec(query, username, email, password)
	return err
}
