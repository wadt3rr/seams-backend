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
	Invoice *models.Invoice `json:"invoice,omitempty"`
}

type InvoiceGetter interface {
	GetInvoiceByOrderID(ctx context.Context, orderID uuid.UUID) (*models.Invoice, error)
}

func New(log *slog.Logger, getter InvoiceGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.invoices.get.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request-id", middleware.GetReqID(r.Context())),
		)

		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			log.Error("missing order id in request")

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("missing order id"))

			return
		}

		id, err := uuid.Parse(idStr)
		if err != nil {
			log.Error("invalid order id format", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("invalid order id format"))

			return
		}

		invoice, err := getter.GetInvoiceByOrderID(r.Context(), id)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				log.Error("invoice not found", sl.Err(err))

				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, resp.Error("invoice not found"))

				return
			}

			log.Error("failed to get invoice", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal server error"))

			return
		}

		render.Status(r, http.StatusOK)
		responseOK(w, r, invoice)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, invoice *models.Invoice) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Invoice:  invoice,
	})
}
