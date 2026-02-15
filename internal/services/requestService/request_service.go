package requestservice

import (
	"context"
	"errors"
	"fmt"
	"seams-backend/internal/models"
	"seams-backend/internal/storage"

	"github.com/google/uuid"
)

type RequestService struct {
	storage Storage
}

type Storage interface {
	SaveCustomer(ctx context.Context, customer *models.Customer) (uuid.UUID, error)
	SaveRequest(ctx context.Context, customer_id uuid.UUID, desc string, file_path string) (uuid.UUID, error)
	GetCustomerByEmail(ctx context.Context, email string) (*models.Customer, error)
}

type CreateRequestReq struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`

	Description string `json:"desc"`
	FilePath    string `json:"file_path"`
	Status      string `json:"status"`
}

type CreateRequestResponse struct {
	RequestID uuid.UUID `json:"request_id"`
}

func New(storage Storage) *RequestService {
	return &RequestService{
		storage: storage,
	}
}

func (r *RequestService) CreateRequestWithCustomer(ctx context.Context, req *CreateRequestReq) (*CreateRequestResponse, error) {
	const op = "services.requestService.CreateRequestWithCustomer"

	if err := r.validateCreateRequestReq(req); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Ищем или создаем клиента
	var customerID uuid.UUID
	customer, err := r.storage.GetCustomerByEmail(ctx, req.Email)

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
		customerID, err = r.storage.SaveCustomer(ctx, customer)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to save customer: %w", op, err)
		}

	default:
		return nil, err
	}

	id, err := r.storage.SaveRequest(ctx, customerID, req.Description, req.FilePath)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to save request: %w", op, err)
	}

	return &CreateRequestResponse{
		RequestID: id,
	}, nil

}

func (r *RequestService) validateCreateRequestReq(req *CreateRequestReq) error {
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

	if req.Description == "" {
		return fmt.Errorf("description is required")
	}

	return nil
}
