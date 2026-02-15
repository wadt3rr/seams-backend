package get

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	resp "seams-backend/internal/lib/api/response"
	"seams-backend/internal/lib/logger/sl"
	"seams-backend/internal/models"
	"seams-backend/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type Response struct {
	resp.Response
	Customer *models.Customer `json:"customer,omitempty"`
}

type CustomerGetter interface {
	GetCustomerByID(ctx context.Context, id uuid.UUID) (*models.Customer, error)
}

func New(log *slog.Logger, getter CustomerGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.customers.get.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request-id", middleware.GetReqID(r.Context())),
		)

		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			log.Error("missing customer id in request")

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("missing customer id"))

			return
		}

		id, err := uuid.Parse(idStr)
		if err != nil {
			log.Error("invalid customer id format", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("invalid customer id format"))

			return
		}

		customer, err := getter.GetCustomerByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				log.Error("customer not found", sl.Err(err))

				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, resp.Error("customer not found"))

				return
			}

			log.Error("failed to get customer", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal server error"))

			return
		}

		render.Status(r, http.StatusOK)
		responseOK(w, r, customer)
	}
}
func responseOK(w http.ResponseWriter, r *http.Request, customer *models.Customer) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Customer: customer,
	})
}
