package list

import (
	"context"
	"log/slog"
	"net/http"
	resp "seams-backend/internal/lib/api/response"
	"seams-backend/internal/lib/logger/sl"
	"seams-backend/internal/models"
	"seams-backend/internal/storage/postgres"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Response struct {
	resp.Response
	Orders []*models.Order `json:"orders"`
}

type OrdersGetter interface {
	ListOrders(ctx context.Context, filter postgres.OrderFilter) ([]*models.Order, error)
}

func New(log *slog.Logger, getter OrdersGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.orders.list.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request-id", middleware.GetReqID(r.Context())),
		)

		orders, err := getter.ListOrders(r.Context(), postgres.OrderFilter{})
		if err != nil {
			log.Error("failed to list orders", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal server error"))

			return
		}

		render.Status(r, http.StatusOK)
		responseOK(w, r, orders)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, orders []*models.Order) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Orders:   orders,
	})
}
