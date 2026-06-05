package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"seams-backend/internal/models"
	"seams-backend/internal/storage"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

type OrderFilter struct {
	CustomerID *uuid.UUID
	Status     *models.OrderStatus

	Limit  int
	Offset int
}

// NewStorage создает новое подключение к PostgreSQL и возвращает Storage
func NewStorage(ctx context.Context, dsn string, migrationsPath string) (*Storage, error) {
	const op = "postgres.NewStorage"
	// Создаем connection pool
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create connection pool: %w", op, err)
	}

	// Запускаем миграции
	if err := runMigrations(migrationsPath, dsn); err != nil {
		pool.Close()
		return nil, fmt.Errorf("%s: failed to run migrations: %w", op, err)
	}

	// Проверяем соединение
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("%s: failed to ping database: %w", op, err)
	}

	return &Storage{pool: pool}, nil
}

// Close закрывает соединение с БД
func (s *Storage) Close() {
	s.pool.Close()
}

// runMigrations применяет миграции из указанной папки
func runMigrations(migrationsPath, dsn string) error {
	// Преобразуем относительный путь в абсолютный
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to resolve migrations path: %w", err)
	}

	// Проверяем что папка существует
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("migrations path does not exist: %s: %w", absPath, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("migrations path is not a directory: %s", absPath)
	}

	log.Printf("Running migrations from: %s", absPath)

	// Формируем file:// URL для migrate
	sourceURL := "file://" + absPath

	// migrate.New принимает путь к источнику и URL базы
	m, err := migrate.New(sourceURL, dsn)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Применяем все миграции вверх
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration up failed: %w", err)
	}

	log.Println("Migrations applied successfully")
	return nil
}

// GetCategories возвращает все категории из базы данных
func (s *Storage) GetCategories(ctx context.Context) ([]*models.Category, error) {
	// Реализация метода для получения категорий из базы данных
	const op = "postgres.Storage.GetCategories"

	rows, err := s.pool.Query(ctx, "SELECT id, name, slug FROM categories ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get categories: %w", op, err)
	}
	defer rows.Close()

	var categories []*models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug); err != nil {
			return nil, fmt.Errorf("%s: failed to scan category: %w", op, err)
		}
		categories = append(categories, &c)
	}

	return categories, nil
}

func (s *Storage) GetCategoryBySlug(ctx context.Context, slug string) (*models.Category, error) {
	const op = "postgres.Storage.GetCategoryBySlug"
	var c models.Category

	err := s.pool.QueryRow(ctx, "SELECT id, name, slug FROM categories WHERE slug = $1", slug).Scan(&c.ID, &c.Name, &c.Slug)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrNotFound)
		}
		return nil, fmt.Errorf("%s: failed to get category by slug: %w", op, err)
	}
	return &c, nil
}

// GetCategoryByID возвращает категорию по ее ID из базы данных
func (s *Storage) GetCategoryByID(ctx context.Context, id uuid.UUID) (*models.Category, error) {
	// Реализация метода для получения категории по ID из базы данных
	const op = "postgres.Storage.GetCategoryByID"
	var c models.Category

	err := s.pool.QueryRow(ctx, "SELECT id, name, slug FROM categories WHERE id = $1", id).Scan(&c.ID, &c.Name, &c.Slug)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrNotFound)
		}
		return nil, fmt.Errorf("%s: failed to get category by ID: %w", op, err)
	}
	return &c, nil
}

// GetProducts возвращает все продукты из базы данных
func (s *Storage) GetProducts(ctx context.Context, limit, offset int) ([]*models.Product, error) {
	// Реализация метода для получения продуктов из базы данных
	const op = "postgres.Storage.GetProducts"

	rows, err := s.pool.Query(ctx, "SELECT id, name, sku, price, category_id, description FROM products ORDER BY name LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get products: %w", op, err)
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Sku, &p.Price, &p.CategoryID, &p.Description); err != nil {
			return nil, fmt.Errorf("%s: failed to scan product: %w", op, err)
		}
		products = append(products, &p)
	}

	return products, nil
}

// GetProductByID возвращает продукт по его ID из базы данных
func (s *Storage) GetProductByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	// Реализация метода для получения продукта по ID из базы данных
	const op = "postgres.Storage.GetProductByID"
	var p models.Product

	err := s.pool.QueryRow(ctx, "SELECT id, name, sku, price, category_id, description FROM products WHERE id = $1", id).Scan(&p.ID, &p.Name, &p.Sku, &p.Price, &p.CategoryID, &p.Description)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrNotFound)
		}
		return nil, fmt.Errorf("%s: failed to get product by ID: %w", op, err)
	}
	return &p, nil
}

// GetProductsByCategoryID возвращает продукты по ID категории из базы данных
func (s *Storage) GetProductsByCategoryID(ctx context.Context, categoryID uuid.UUID, limit, offset int) ([]*models.Product, error) {
	// Реализация метода для получения продуктов по ID категории из базы данных
	const op = "postgres.Storage.GetProductsByCategoryID"

	rows, err := s.pool.Query(ctx, "SELECT id, name, sku, price, category_id, description FROM products WHERE category_id = $1 ORDER BY name LIMIT $2 OFFSET $3", categoryID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get products by category ID: %w", op, err)
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Sku, &p.Price, &p.CategoryID, &p.Description); err != nil {
			return nil, fmt.Errorf("%s: failed to scan product: %w", op, err)
		}
		products = append(products, &p)
	}

	return products, nil
}

// GetProductBySku возвращает продукт по его SKU из базы данных
func (s *Storage) GetProductBySku(ctx context.Context, sku string) (*models.Product, error) {
	const op = "postgres.Storage.GetProductBySku"
	var p models.Product

	err := s.pool.QueryRow(ctx, "SELECT id, name, sku, price, category_id, description FROM products WHERE sku = $1", sku).Scan(&p.ID, &p.Name, &p.Sku, &p.Price, &p.CategoryID, &p.Description)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrNotFound)
		}
		return nil, fmt.Errorf("%s: failed to get product by SKU: %w", op, err)
	}
	return &p, nil
}

// SaveProduct сохраняет продукт в базе данных
func (s *Storage) SaveProduct(ctx context.Context, product *models.Product) (uuid.UUID, error) {
	const op = "postgres.Storage.SaveProduct"

	err := s.pool.QueryRow(
		ctx,
		`INSERT INTO products (name, sku, price, category_id, description) 
		VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		product.Name,
		product.Sku,
		product.Price,
		product.CategoryID,
		product.Description,
	).Scan(&product.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: failed to save product: %w", op, err)
	}

	return product.ID, nil
}

// SearchProducts ищет продукты по запросу в базе данных
func (s *Storage) SearchProducts(ctx context.Context, query string, limit, offset int) ([]*models.Product, error) {
	const op = "postgres.Storage.SearchProducts"

	rows, err := s.pool.Query(
		ctx,
		`SELECT id, name, sku, price, category_id, description FROM products 
		WHERE name ILIKE $1 OR description ILIKE $1 OR sku ILIKE $1
		ORDER BY name LIMIT $2 OFFSET $3`,
		"%"+query+"%",
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to search products: %w", op, err)
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Sku, &p.Price, &p.CategoryID, &p.Description); err != nil {
			return nil, fmt.Errorf("%s: failed to scan product: %w", op, err)
		}
		products = append(products, &p)
	}

	return products, nil
}

// SaveOrder сохраняет заказ и его позиции в базе данных
func (s *Storage) SaveOrder(ctx context.Context, order *models.Order) error {
	const op = "postgres.Storage.SaveOrder"

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: begin transaction: %w", op, err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// считаем total на backend
	var total int64
	for _, item := range order.Items {
		if item.Quantity <= 0 {
			return fmt.Errorf("%s: invalid quantity", op)
		}
		total += item.Price * int64(item.Quantity)
	}
	order.Total = total

	// если ID не задан — генерируем
	if order.ID == uuid.Nil {
		order.ID = uuid.New()
	}

	// вставка заказа
	err = tx.QueryRow(
		ctx,
		`INSERT INTO orders (id, customer_id, status, total, payment_method, fullfillment_method)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING created_at, updated_at`,
		order.ID,
		order.CustomerID,
		order.Status,
		order.Total,
	).Scan(&order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return fmt.Errorf("%s: insert order: %w", op, err)
	}

	// вставка позиций заказа
	for _, item := range order.Items {
		itemID := item.ID
		if itemID == uuid.Nil {
			itemID = uuid.New()
		}

		_, err = tx.Exec(
			ctx,
			`INSERT INTO order_items (id, order_id, product_id, quantity, price, total)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			itemID,
			order.ID,
			item.ProductID,
			item.Quantity,
			item.Price,
			item.Total,
			order.PaymentMethod,
			order.FullfillmentMethod,
		)
		if err != nil {
			return fmt.Errorf("%s: insert order item: %w", op, err)
		}
	}

	// коммит
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: commit transaction: %w", op, err)
	}

	return nil
}

// GetOrderByID возвращает заказ по его ID из базы данных
func (s *Storage) GetOrderByID(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	const op = "postgres.Storage.GetOrderByID"
	var order models.Order

	// Получаем заказ
	err := s.pool.QueryRow(ctx, "SELECT id, customer_id, status, total, created_at, updated_at, payment_method, fullfillment_method FROM orders WHERE id = $1", id).
		Scan(&order.ID, &order.CustomerID, &order.Status, &order.Total, &order.CreatedAt, &order.UpdatedAt, &order.PaymentMethod, &order.FullfillmentMethod)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrNotFound)
		}
		return nil, fmt.Errorf("%s: failed to get order by ID: %w", op, err)
	}

	// Получаем позиции заказа
	rows, err := s.pool.Query(ctx, "SELECT id, product_id, quantity, price, total FROM order_items WHERE order_id = $1", order.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get order items: %w", op, err)
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ID, &item.ProductID, &item.Quantity, &item.Price, &item.Total); err != nil {
			return nil, fmt.Errorf("%s: failed to scan order item: %w", op, err)
		}
		items = append(items, item)
	}
	order.Items = items

	return &order, nil
}

// ListOrders возвращает все заказы из базы данных
func (s *Storage) ListOrders(ctx context.Context, filter OrderFilter) ([]*models.Order, error) {
	const op = "postgres.Storage.ListOrders"

	query := `
		SELECT
			id,
			customer_id,
			status,
			total,
			created_at,
			updated_at
		FROM orders
		WHERE
			($1::uuid IS NULL OR customer_id = $1)
			AND ($2::text IS NULL OR status = $2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}

	rows, err := s.pool.Query(ctx, query, filter.CustomerID, filter.Status, limit, filter.Offset)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get orders: %w", op, err)
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(&order.ID, &order.CustomerID, &order.Status, &order.Total, &order.CreatedAt, &order.UpdatedAt); err != nil {
			return nil, fmt.Errorf("%s: failed to scan order: %w", op, err)
		}
		orders = append(orders, &order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows error: %w", op, err)
	}

	return orders, nil
}

// UpdateOrder обновляет статус заказа в базе данных
func (s *Storage) UpdateOrder(ctx context.Context, orderID uuid.UUID, status string) (bool, error) {
	const op = "postgres.Storage.UpdateOrder"

	res, err := s.pool.Exec(ctx, `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`, status, orderID)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	if res.RowsAffected() == 0 {
		return false, nil
	}
	return true, nil
}

// SaveCustomer сохраняет клиента в базе данных
func (s *Storage) SaveCustomer(ctx context.Context, customer *models.Customer) (uuid.UUID, error) {
	const op = "postgres.Storage.CreateCustomer"

	var id uuid.UUID
	err := s.pool.QueryRow(ctx,
		`INSERT INTO customers (name, email, phone)
		 VALUES ($1, $2, $3) RETURNING id`,
		customer.Name,
		customer.Email,
		customer.Phone,
	).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: insert customer: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetCustomerByID(ctx context.Context, id uuid.UUID) (*models.Customer, error) {
	{
		const op = "postgres.Storage.GetCustomerByID"
		var c models.Customer

		err := s.pool.QueryRow(ctx, "SELECT id, name, email, phone FROM customers WHERE id = $1", id).Scan(&c.ID, &c.Name, &c.Email, &c.Phone)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf("%s: %w", op, storage.ErrNotFound)
			}
			return nil, fmt.Errorf("%s: failed to get customer by ID: %w", op, err)
		}
		return &c, nil
	}
}

// GetCustomerByEmail возвращает клиента по его email из базы данных
func (s *Storage) GetCustomerByEmail(ctx context.Context, email string) (*models.Customer, error) {
	const op = "postgres.Storage.GetCustomerByEmail"
	var c models.Customer

	err := s.pool.QueryRow(ctx, "SELECT id, name, email, phone FROM customers WHERE email = $1", email).Scan(&c.ID, &c.Name, &c.Email, &c.Phone)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrNotFound)
		}
		return nil, fmt.Errorf("%s: failed to get customer by email: %w", op, err)
	}
	return &c, nil
}

// ListCustomers возвращает всех клиентов из базы данных
func (s *Storage) ListCustomers(ctx context.Context) ([]*models.Customer, error) {
	const op = "postgres.Storage.ListCustomers"

	rows, err := s.pool.Query(ctx, "SELECT id, name, email, phone FROM customers")
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get customers: %w", op, err)
	}
	defer rows.Close()

	var customers []*models.Customer
	for rows.Next() {
		var c models.Customer
		if err := rows.Scan(&c.ID, &c.Name, &c.Email, &c.Phone); err != nil {
			return nil, fmt.Errorf("%s: failed to scan customer: %w", op, err)
		}
		customers = append(customers, &c)
	}

	return customers, nil
}

// SaveInvoice сохраняет счет в базе данных
func (s *Storage) SaveInvoice(ctx context.Context, invoice *models.Invoice) error {
	const op = "postgres.Storage.SaveInvoice"

	err := s.pool.QueryRow(ctx,
		`INSERT INTO invoices (order_id, number, amount, status, issued_at)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at`,
		invoice.OrderID,
		invoice.Number,
		invoice.Amount,
		invoice.Status,
		invoice.IssuedAt,
	).Scan(&invoice.ID, &invoice.CreatedAt)

	if err != nil {
		return fmt.Errorf("%s: insert invoice: %w", op, err)
	}
	return nil
}

// GetInvoiceByOrderID возвращает счет по ID заказа из базы данных
func (s *Storage) GetInvoiceByOrderID(ctx context.Context, orderID uuid.UUID) (*models.Invoice, error) {
	const op = "postgres.Storage.GetInvoiceByOrderID"
	var invoice models.Invoice

	err := s.pool.QueryRow(ctx, "SELECT id, order_id, number, amount, status, issued_at, created_at, updated_at FROM invoices WHERE order_id = $1", orderID).
		Scan(&invoice.ID, &invoice.OrderID, &invoice.Number, &invoice.Amount, &invoice.Status, &invoice.IssuedAt, &invoice.CreatedAt, &invoice.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrNotFound)
		}
		return nil, fmt.Errorf("%s: failed to get invoice by order ID: %w", op, err)
	}

	return &invoice, nil
}

// UpdateInvoiceStatus обновляет статус счета в базе данных
func (s *Storage) UpdateInvoiceStatus(ctx context.Context, invoiceID uuid.UUID, status models.InvoiceStatus) error {
	const op = "postgres.Storage.UpdateInvoiceStatus"

	_, err := s.pool.Exec(ctx,
		`UPDATE invoices SET status = $1, updated_at = now() WHERE id = $2`,
		status,
		invoiceID,
	)

	if err != nil {
		return fmt.Errorf("%s: update invoice status: %w", op, err)
	}
	return nil
}

// ListInvoices возвращает все счета из базы данных
func (s *Storage) ListInvoices(ctx context.Context) ([]*models.Invoice, error) {
	const op = "postgres.Storage.ListInvoices"

	rows, err := s.pool.Query(ctx, "SELECT id, order_id, number, amount, status, issued_at, created_at, updated_at FROM invoices")
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get invoices: %w", op, err)
	}
	defer rows.Close()

	var invoices []*models.Invoice
	for rows.Next() {
		var invoice models.Invoice
		if err := rows.Scan(&invoice.ID, &invoice.OrderID, &invoice.Number, &invoice.Amount, &invoice.Status, &invoice.IssuedAt, &invoice.CreatedAt, &invoice.UpdatedAt); err != nil {
			return nil, fmt.Errorf("%s: failed to scan invoice: %w", op, err)
		}
		invoices = append(invoices, &invoice)
	}

	return invoices, nil
}

func (s *Storage) SaveRequest(ctx context.Context, customer_id uuid.UUID, desc string, file_path string) (uuid.UUID, error) {
	const op = "postgres.Storage.SaveRequest"

	var id uuid.UUID

	err := s.pool.QueryRow(ctx, `INSERT INTO requests(customer_id, description, file_path) VALUES ($1, $2, $3) RETURNING id`, customer_id, desc, file_path).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}
