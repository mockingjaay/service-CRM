package main

import (
	"database/sql"
	"log"
	"net/http"
	"repair-crm/internal/models" // Замени на путь из своего go.mod
	"strconv"
	"text/template"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "modernc.org/sqlite"
)

// Создаем структуру для хранения зависимостей приложения
type application struct {
	orders *models.OrderModel
}

func main() {
	// 1. Подключаемся к базе данных
	db, err := sql.Open("sqlite", "repair.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 2. Инициализируем зависимости (нашу структуру application)
	app := &application{
		orders: &models.OrderModel{DB: db},
	}

	// 3. Настраиваем роутер
	r := chi.NewRouter()
	r.Use(middleware.Logger) // Добавляет логирование запросов в консоль

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	r.Handle("/static/*", http.StripPrefix("/static", fileServer))

	// 4. Описываем маршруты (routes)
	r.Get("/", app.homeHandler) // Показать главную с формой
	r.Get("/order/status", app.updateStatusHandler)
	r.Get("/order/delete", app.deleteOrderHandler)
	r.Post("/order/create", app.createOrderHandler) // Обработать отправку формы

	// 5. Запускаем сервер
	log.Println("Сервер запущен на http://localhost:8080")
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal(err)
	}
}

func (app *application) homeHandler(w http.ResponseWriter, r *http.Request) {
	// Достаем параметры из ссылки (например, /?status=новый&search_id=5)
	statusFilter := r.URL.Query().Get("status")
	searchID := r.URL.Query().Get("search_id")

	// Передаем эти фильтры в модель
	orders, err := app.orders.Latest(statusFilter, searchID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Ошибка БД", 500)
		return
	}

	ts, err := template.ParseFiles("./ui/html/index.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Файл не найден", 500)
		return
	}

	// Отправляем данные в шаблон
	err = ts.Execute(w, orders)
	if err != nil {
		log.Println(err)
	}
}

func (app *application) createOrderHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Парсим данные из формы
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", 400)
		return
	}

	// 2. Достаем значения по именам полей из HTML
	customer := r.PostForm.Get("customer_name")
	device := r.PostForm.Get("device_name")
	description := r.PostForm.Get("description")

	// 3. Сохраняем в базу данных через нашу модель
	_, err = app.orders.Insert(customer, device, description)
	if err != nil {
		http.Error(w, "Ошибка сохранения", 500)
		return
	}

	// 4. Перенаправляем пользователя обратно на главную страницу
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) updateStatusHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем ID из URL (например, /order/status?id=1&status=выполнен)
	idStr := r.URL.Query().Get("id")
	newStatus := r.URL.Query().Get("status")

	// Преобразуем строку ID в число
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID", 400)
		return
	}

	// Обновляем в базе
	err = app.orders.UpdateStatus(id, newStatus)
	if err != nil {
		http.Error(w, "Ошибка обновления", 500)
		return
	}

	// Возвращаемся на главную
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) deleteOrderHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID", 400)
		return
	}

	err = app.orders.Delete(id)
	if err != nil {
		http.Error(w, "Ошибка удаления", 500)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
