package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "modernc.org/sqlite" // Импорт драйвера (обязательно через нижнее подчеркивание)
)

func main() {
	// 1. Открываем (или создаем) файл базы данных
	db, err := sql.Open("sqlite", "repair.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 2. Создаем таблицу, если её еще нет
	statement := `
    CREATE TABLE IF NOT EXISTS orders (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        customer_name TEXT,
        device_name TEXT,
        description TEXT,
        status TEXT DEFAULT 'новый',
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`

	_, err = db.Exec(statement)
	if err != nil {
		log.Fatalf("Ошибка при создании таблицы: %q: %s\n", err, statement)
	}

	// Твой текущий роутер chi
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// Тут будет логика получения данных из db
		w.Write([]byte("База данных готова, таблица orders создана!"))
	})

	log.Println("Сервер запущен на http://localhost:8080")
	http.ListenAndServe(":8080", r)
}
