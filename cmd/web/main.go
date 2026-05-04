package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"repair-crm/internal/models" // Проверь, что структура Order там есть

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "modernc.org/sqlite"
)

// Оставляем здесь, если в internal/models ее нет.
// Если есть — удали этот блок и используй models.Order
type Order struct {
	ID             int
	CustomerName   string
	DeviceName     string
	Description    string
	Status         string
	DeviceModel    string
	Appearance     string
	Password       string
	Equipment      string
	EstimatedPrice float64
}

type application struct {
	orders *models.OrderModel
}

var db *sql.DB // Глобальная переменная для БД

func main() {
	var err error
	// Инициализируем ГЛОБАЛЬНУЮ db (без :=)
	db, err = sql.Open("sqlite", "repair.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Проверка соединения
	if err = db.Ping(); err != nil {
		log.Fatal("БД недоступна:", err)
	}

	app := &application{
		orders: &models.OrderModel{DB: db},
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Статика
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	r.Handle("/static/*", http.StripPrefix("/static", fileServer))

	// МАРШРУТЫ
	r.Get("/", app.homeHandler)
	r.Post("/order/create", app.createOrderHandler)
	r.Get("/order/view", app.viewOrderHandler) // Теперь через r.Get
	r.Get("/order/status", app.updateStatusHandler)
	r.Get("/order/delete", app.deleteOrderHandler)

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
	if r.Method == http.MethodPost {
		// Читаем старые поля
		name := r.FormValue("customer_name")
		device := r.FormValue("device_name")
		desc := r.FormValue("description")

		// Читаем НОВЫЕ поля
		model := r.FormValue("device_model")
		appearance := r.FormValue("appearance")
		password := r.FormValue("password")
		equipment := r.FormValue("equipment")
		price := r.FormValue("estimated_price")

		// Записываем в базу (добавь новые колонки в свой SQL INSERT)
		query := `INSERT INTO orders 
                  (customer_name, device_name, description, status, device_model, appearance, password, equipment, estimated_price) 
                  VALUES (?, ?, ?, 'новый', ?, ?, ?, ?, ?)`

		_, err := db.Exec(query, name, device, desc, model, appearance, password, equipment, price)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// Редирект или HTMX ответ
		http.Redirect(w, r, "/", 303)
	}
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

func (app *application) viewOrderHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")

	var o Order
	query := `SELECT id, customer_name, device_name, description, status, 
                     device_model, appearance, password, equipment, estimated_price 
              FROM orders WHERE id = ?`

	// Используем db напрямую или app.orders.DB
	err := db.QueryRow(query, idStr).Scan(
		&o.ID, &o.CustomerName, &o.DeviceName, &o.Description, &o.Status,
		&o.DeviceModel, &o.Appearance, &o.Password, &o.Equipment, &o.EstimatedPrice,
	)

	if err != nil {
		log.Println("Ошибка при поиске заказа:", err)
		http.Error(w, "Заказ не найден", 404)
		return
	}

	ts, err := template.ParseFiles("./ui/html/order_view.html")
	if err != nil {
		http.Error(w, "Шаблон не найден", 500)
		return
	}
	ts.Execute(w, o)

	// Готовим даты
	now := time.Now()
	warrantyUntil := now.AddDate(0, 0, 14) // Плюс 2 недели

	// Передаем всё в шаблон через карту (map)
	data := map[string]interface{}{
		"Order":        o,
		"PrintDate":    now.Format("02.01.2006"),
		"WarrantyDate": warrantyUntil.Format("02.01.2006"),
	}

	tmpl, _ := template.ParseFiles("ui/html/order_view.html")
	tmpl.Execute(w, data)
}
