package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"wb-project/internal/models"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// Save - метод для сохранения order в БД.
func (r *OrderRepository) Save(ctx context.Context, order models.Order) error {
	// Начинаем транзакцию
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() { //при ошибке откатываем транзакцию
		if err := tx.Rollback(); err != nil {
			log.Printf("не удалось откатить транзакцию %v", err)
		}
	}()

	// Сначала добавляем заказ в бд
	_, err = tx.ExecContext(ctx,
		`INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shard_key, sm_id, date_created, oof_shard) 
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,
		order.CustomerID, order.DeliveryService, order.ShardKey, order.SmID, order.DateCreated, order.OofShard,
	)
	if err != nil {
		return fmt.Errorf("ошибка при добавлении сущности order в БД, error: %w", err)
	}

	// Добавляем сущность payments
	_, err = tx.ExecContext(ctx,
		`INSERT INTO payments (order_uid, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee) 
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		order.OrderUID, order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency, order.Payment.Provider,
		order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee,
	)
	if err != nil {
		return fmt.Errorf("ошибка при добавлении сущности payments в бд, error: %w", err)
	}

	// Добавляем сущность deliveries
	_, err = tx.ExecContext(ctx,
		`INSERT INTO deliveries (order_uid, name, phone, zip, city, address, region, email) 
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email,
	)
	if err != nil {
		return fmt.Errorf("ошибка при добавлении сущности delivery в бд, error: %w", err)
	}

	// Добавляем сущность Items
	for _, item := range order.Items {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status) 
             VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.Rid, item.Name,
			item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status,
		)
		if err != nil {
			return fmt.Errorf("ошибка при добавлении сущности item в бд, error: %w", err)
		}
	}

	// В случая успеха фиксируем наши изменения
	return tx.Commit()
}

// Get - метод получения, возвращает заказ и ошибку
func (r *OrderRepository) Get(ctx context.Context, uid string) (models.Order, error) {
	var order models.Order

	//orders
	err := r.db.QueryRowContext(ctx, "Select order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shard_key, sm_id, date_created, oof_shard  FROM orders Where order_uid=$1",
		uid).Scan(&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature, &order.CustomerID, &order.DeliveryService, &order.ShardKey, &order.SmID, &order.DateCreated, &order.OofShard)
	if err != nil {
		return models.Order{}, fmt.Errorf("error при получении orders: %v", err)
	}
	//payments
	err = r.db.QueryRowContext(ctx, `Select transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee FROM payments Where order_uid = $1`,
		uid).Scan(&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency, &order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDt, &order.Payment.Bank, &order.Payment.DeliveryCost, &order.Payment.GoodsTotal, &order.Payment.CustomFee)
	if err != nil {
		return models.Order{}, fmt.Errorf("error при получении payments: %v", err)
	}
	//deliveries
	err = r.db.QueryRowContext(ctx, "SELECT name, phone, zip, city, address, region, email FROM deliveries WHERE order_uid = $1",
		uid).Scan(&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip, &order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email)
	if err != nil {
		return models.Order{}, fmt.Errorf("error при получении deliveries: %v", err)
	}

	//items
	rows, err := r.db.QueryContext(ctx, "Select chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status FROM items where order_uid=$1", uid)
	if err != nil {
		return models.Order{}, fmt.Errorf("error при получении items : %v", err)
	}
	defer func() {
		if err = rows.Close(); err != nil {
			log.Printf("ошибка при закрытии rows: %v", err)
		}
	}()

	for rows.Next() {
		var item models.Items
		if err := rows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status); err != nil {
			return models.Order{}, fmt.Errorf("error при получении items: %v", err)
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}

// GetAll - возвращает массив заказов и ошибку
func (r *OrderRepository) GetAll(ctx context.Context) ([]models.Order, error) {
	var orders []models.Order

	rows, err := r.db.QueryContext(ctx, "SELECT order_uid from orders")
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении всех заказов")
	}
	defer func() {
		if err = rows.Close(); err != nil {
			log.Printf("ошибка при закрытии rows: %v", err)
		}
	}()

	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			continue
		}

		order, err := r.Get(ctx, uid)
		if err == nil {
			orders = append(orders, order)
		}

	}
	return orders, nil

}
