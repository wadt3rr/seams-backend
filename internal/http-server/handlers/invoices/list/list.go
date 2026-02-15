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
	Invoices []*models.Invoice `json:"invoices"`
}

type InvoicesGetter interface {
	ListInvoices(ctx context.Context) ([]*models.Invoice, error)
}

func New(log *slog.Logger, getter InvoicesGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.invoices.list.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request-id", middleware.GetReqID(r.Context())),
		)

		invoices, err := getter.ListInvoices(r.Context())
		if err != nil {
			log.Error("failed to list invoices", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal server error"))

			return
		}

		render.Status(r, http.StatusOK)
		responseOK(w, r, invoices)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, invoices []*models.Invoice) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Invoices: invoices,
	})
}
