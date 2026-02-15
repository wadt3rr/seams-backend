package list

import (
	"context"
	"log/slog"
	"net/http"
	resp "seams-backend/internal/lib/api/response"
	"seams-backend/internal/lib/logger/sl"
	"seams-backend/internal/models"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Response struct {
	resp.Response
	Customers []*models.Customer `json:"customers"`
}

type CustomersGetter interface {
	ListCustomers(ctx context.Context) ([]*models.Customer, error)
}

func New(log *slog.Logger, getter CustomersGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.customers.list.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request-id", middleware.GetReqID(r.Context())),
		)

		customers, err := getter.ListCustomers(r.Context())
		if err != nil {
			log.Error("failed to list customers", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal server error"))

			return
		}

		render.Status(r, http.StatusOK)
		responseOK(w, r, customers)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, customers []*models.Customer) {
	render.JSON(w, r, Response{
		Response:  resp.OK(),
		Customers: customers,
	})
}
