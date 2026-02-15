package update

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"seams-backend/internal/lib/logger/sl"
	"seams-backend/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type InvoiceUpdater interface {
	UpdateInvoiceStatus(ctx context.Context, id uuid.UUID, status models.InvoiceStatus) error
}

type InvoiceUpdateRequest struct {
	Status string `json:"status,omitempty"`
}

type InvoiceUpdateResponse struct {
	Success bool `json:"success"`
}

func New(log *slog.Logger, updater InvoiceUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.invoices.update.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request-id", chi.URLParam(r, "id")),
		)

		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			log.Error("missing invoice ID in URL parameters")

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, InvoiceUpdateResponse{Success: false})

			return
		}

		id, err := uuid.Parse(idStr)
		if err != nil {
			log.Error("failed to parse invoice ID", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, InvoiceUpdateResponse{Success: false})

			return
		}

		var req InvoiceUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, InvoiceUpdateResponse{Success: false})

			return
		}

		status := models.InvoiceStatus(req.Status)

		if err := updater.UpdateInvoiceStatus(r.Context(), id, status); err != nil {
			log.Error("failed to update invoice", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, InvoiceUpdateResponse{Success: false})

			return
		}

		log.Info("invoice updated successfully")

		render.JSON(w, r, InvoiceUpdateResponse{Success: true})
	}
}
