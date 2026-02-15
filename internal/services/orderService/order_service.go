package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"seams-backend/internal/models"
	"seams-backend/internal/storage"
	"seams-backend/internal/storage/postgres"

	"github.com/google/uuid"
)

type Mailer interface {
	SendOrderCreated(ctx context.Context, order *models.Order, customer *models.Customer) error
}

// OrderService содержит бизнес-логику для работы с заказами
type OrderService struct {
	storage Storage
}

// Storage интерфейс для работы с хранилищем
type Storage interface {
	GetProductByID(ctx context.Context, id uuid.UUID) (*models.Product, error)
	SaveCustomer(ctx context.Context, customer *models.Customer) (uuid.UUID, error)
	GetCustomerByEmail(ctx context.Context, email string) (*models.Customer, error)
	SaveOrder(ctx context.Context, order *models.Order) error
	SaveInvoice(ctx context.Context, invoice *models.Invoice) error
	ListOrders(ctx context.Context, filter postgres.OrderFilter) ([]*models.Order, error)
	GetOrderByID(ctx context.Context, orderID uuid.UUID) (*models.Order, error)
	GetCustomerByID(ctx context.Context, customerID uuid.UUID) (*models.Customer, error)
	GetInvoiceByOrderID(ctx context.Context, orderID uuid.UUID) (*models.Invoice, error)
}

// CreateOrderRequest запрос на создание заказа
type CreateOrderRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`

	PaymentMethod      string `json:"payment_method"`
	FullfillmentMethod string `json:"fullfillment_method"`

	Items []CreateOrderItemRequest `json:"items"`
}

// CreateOrderItemRequest элемент заказа в запросе
type CreateOrderItemRequest struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}

// CreateOrderResponse ответ при создании заказа
type CreateOrderResponse struct {
	OrderID   uuid.UUID `json:"order_id"`
	InvoiceID uuid.UUID `json:"invoice_id"`
	Total     int64     `json:"total"`
}

// NewOrderService создает новый сервис для работы с заказами
func NewOrderService(storage Storage) *OrderService {
	return &OrderService{storage: storage}
}

// CreateOrderWithCustomerAndInvoice создает заказ, сохраняет клиента и выстав инвойс
func (s *OrderService) CreateOrderWithCustomerAndInvoice(
	ctx context.Context,
	req *CreateOrderRequest,
) (*CreateOrderResponse, error) {
	const op = "services.OrderService.CreateOrderWithCustomerAndInvoice"

	// Валидируем запрос
	if err := s.validateCreateOrderRequest(req); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Ищем или создаем клиента
	var customerID uuid.UUID
	customer, err := s.storage.GetCustomerByEmail(ctx, req.Email)

	switch {
	case err == nil:
		customerID = customer.ID

	case errors.Is(err, storage.ErrNotFound):
		customer = &models.Customer{
			ID:    uuid.New(),
			Name:  req.Name,
			Email: req.Email,
			Phone: req.Phone,
		}
		customerID, err = s.storage.SaveCustomer(ctx, customer)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to save customer: %w", op, err)
		}

	default:
		return nil, err
	}

	// Расчитываем сумму и валидируем товары
	var totalAmount int64
	var orderItems []models.OrderItem

	for _, reqItem := range req.Items {
		// Проверяем что товар существует и получаем его цену
		product, err := s.storage.GetProductByID(ctx, reqItem.ProductID)
		if err != nil {
			return nil, fmt.Errorf("%s: product not found: %s: %w", op, reqItem.ProductID, err)
		}

		if reqItem.Quantity <= 0 {
			return nil, fmt.Errorf("%s: invalid quantity for product %s", op, reqItem.ProductID)
		}

		// Считаем стоимость позиции
		itemTotal := product.Price * int64(reqItem.Quantity)
		totalAmount += itemTotal

		// Добавляем позицию в заказ
		orderItems = append(orderItems, models.OrderItem{
			ID:        uuid.New(),
			ProductID: reqItem.ProductID,
			Quantity:  reqItem.Quantity,
			Price:     product.Price,
			Total:     itemTotal,
		})
	}

	// Создаем заказ
	order := &models.Order{
		ID:                 uuid.New(),
		CustomerID:         customerID,
		Items:              orderItems,
		Total:              totalAmount,
		PaymentMethod:      models.PaymentMethod(req.PaymentMethod),
		FullfillmentMethod: models.FullfillmentMethod(req.FullfillmentMethod),
		Status:             models.OrderNew,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.storage.SaveOrder(ctx, order); err != nil {
		return nil, fmt.Errorf("%s: failed to save order: %w", op, err)
	}

	// Создаем инвойс для заказа
	invoice := &models.Invoice{
		ID:        uuid.New(),
		OrderID:   order.ID,
		Number:    s.generateInvoiceNumber(),
		Amount:    totalAmount,
		Status:    models.InvoiceIssued,
		IssuedAt:  time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.storage.SaveInvoice(ctx, invoice); err != nil {
		return nil, fmt.Errorf("%s: failed to save invoice: %w", op, err)
	}

	return &CreateOrderResponse{
		OrderID:   order.ID,
		InvoiceID: invoice.ID,
		Total:     totalAmount,
	}, nil
}

func (s *OrderService) ListOrdersByUserEmail(
	ctx context.Context,
	email string,
) ([]*models.Order, error) {
	const op = "services.OrderService.ListOrdersByUserEmail"

	if email == "" {
		return nil, fmt.Errorf("%s: email is required", op)
	}

	// 1. ищем клиента
	customer, err := s.storage.GetCustomerByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return []*models.Order{}, nil
		}
		return nil, fmt.Errorf("%s: failed to get customer: %w", op, err)
	}

	// 2. собираем фильтр
	filter := postgres.OrderFilter{
		CustomerID: &customer.ID,
	}

	// 3. получаем заказы
	orders, err := s.storage.ListOrders(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to list orders: %w", op, err)
	}

	return orders, nil
}

func (s *OrderService) GetOrderDetails(
	ctx context.Context,
	orderID uuid.UUID,
) (*models.OrderDetailsDTO, error) {
	const op = "services.OrderService.GetOrderDetails"

	// 1. заказ
	order, err := s.storage.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// 2. покупатель
	customer, err := s.storage.GetCustomerByID(ctx, order.CustomerID)
	if err != nil {
		return nil, err
	}

	// 3. товары
	items := make([]models.OrderItemDTO, 0, len(order.Items))

	for _, item := range order.Items {
		product, err := s.storage.GetProductByID(ctx, item.ProductID)
		if err != nil {
			return nil, err
		}

		items = append(items, models.OrderItemDTO{
			ProductID:   item.ProductID,
			ProductName: product.Name,
			Quantity:    item.Quantity,
			Price:       item.Price,
			Total:       item.Total,
		})
	}

	// 4. инвойс

	invoice, err := s.storage.GetInvoiceByOrderID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	// 5. собираем DTO
	return &models.OrderDetailsDTO{
		ID:                 order.ID,
		CustomerName:       customer.Name,
		Status:             order.Status,
		PaymentMethod:      order.PaymentMethod,
		FullfillmentMethod: order.FullfillmentMethod,
		Items:              items,
		Total:              order.Total,
		CreatedAt:          order.CreatedAt,
		Invoice:            invoice,
	}, nil
}

// validateCreateOrderRequest валидирует запрос на создание заказа
func (s *OrderService) validateCreateOrderRequest(req *CreateOrderRequest) error {
	if req == nil {
		return fmt.Errorf("request is nil")
	}

	if req.Name == "" {
		return fmt.Errorf("customer name is required")
	}

	if req.Email == "" {
		return fmt.Errorf("customer email is required")
	}

	if req.Phone == "" {
		return fmt.Errorf("customer phone is required")
	}

	if req.FullfillmentMethod == "" {
		return fmt.Errorf("customer fullfillment_method is required")
	}

	if req.PaymentMethod == "" {
		return fmt.Errorf("customer payment_method is required")
	}

	if len(req.Items) == 0 {
		return fmt.Errorf("order must contain at least one item")
	}

	for i, item := range req.Items {
		if item.ProductID == uuid.Nil {
			return fmt.Errorf("product_id at index %d is required", i)
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("quantity at index %d must be greater than 0", i)
		}
	}

	return nil
}

// generateInvoiceNumber генерирует номер инвойса
func (s *OrderService) generateInvoiceNumber() string {
	now := time.Now()
	year := now.Year()
	return fmt.Sprintf("INV-%d-%d", year, time.Now().UnixNano()%100000)
}
