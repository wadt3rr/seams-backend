package update

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	resp "seams-backend/internal/lib/api/response"
	"seams-backend/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type Request struct {
	Status string `json:"status"`
}

type Response struct {
	resp.Response
	Updated bool `json:"updated"`
}

type OrderUpdater interface {
	UpdateOrder(ctx context.Context, orderID uuid.UUID, status string) (bool, error)
}

func New(log *slog.Logger, updater OrderUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.orders.update.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		idStr := chi.URLParam(r, "id")

		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {

			log.Error("request body is empty")

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("empty request"))

			return
		}
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}
		log.Info("request body decoded", slog.Any("request", req))

		orderID, err := uuid.Parse(idStr)
		if err != nil {
			log.Error("failed to parse orderID", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("failed to parse orderID"))
		}
		result, err := updater.UpdateOrder(r.Context(), orderID, req.Status)
		if err != nil {
			log.Error("failed to update order status", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal server error"))
			return
		}

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, &Response{
			Response: resp.OK(),
			Updated:  result,
		})
	}
}
