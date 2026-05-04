package models

import (
	"database/sql"
	"time"
)

// Order описывает структуру заказа в нашей CRM
type Order struct {
	ID             int
	CustomerName   string
	DeviceName     string
	Description    string
	Status         string
	CreatedAt      time.Time
	DeviceModel    string
	Appearance     string
	Password       string
	Equipment      string
	EstimatedPrice float64
}

// OrderModel определяет объект, который будет работать с БД
type OrderModel struct {
	DB *sql.DB
}

// Insert добавляет новый заказ в базу данных
func (m *OrderModel) Insert(customer, device, desc string) (int, error) {
	stmt := `INSERT INTO orders (customer_name, device_name, description, status, created_at)
    VALUES(?, ?, ?, 'новый', CURRENT_TIMESTAMP)`

	result, err := m.DB.Exec(stmt, customer, device, desc)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// Get достает один заказ по ID
func (m *OrderModel) Get(id int) (*Order, error) {
	return nil, nil // Это реализуем позже
}

// Latest возвращает последние 10 заказов
func (m *OrderModel) Latest(status string, searchID string) ([]*Order, error) {
	// Базовый запрос
	stmt := `SELECT id, customer_name, device_name, description, status, created_at FROM orders WHERE 1=1`
	var args []interface{}

	// Если передан статус, добавляем фильтр
	if status != "" {
		stmt += " AND status = ?"
		args = append(args, status)
	}

	// Если передан ID для поиска
	if searchID != "" {
		stmt += " AND id = ?"
		args = append(args, searchID)
	}

	stmt += " ORDER BY created_at DESC LIMIT 50"

	rows, err := m.DB.Query(stmt, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		o := &Order{}
		err = rows.Scan(&o.ID, &o.CustomerName, &o.DeviceName, &o.Description, &o.Status, &o.CreatedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

// UpdateStatus изменяет статус существующего заказа
func (m *OrderModel) UpdateStatus(id int, status string) error {
	stmt := `UPDATE orders SET status = ? WHERE id = ?`

	_, err := m.DB.Exec(stmt, status, id)
	return err
}

func (m *OrderModel) Delete(id int) error {
	stmt := `DELETE FROM orders WHERE id = ?`
	_, err := m.DB.Exec(stmt, id)
	return err
}
