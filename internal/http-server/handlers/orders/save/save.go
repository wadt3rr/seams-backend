package save

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	resp "seams-backend/internal/lib/api/response"
	"seams-backend/internal/lib/logger/sl"
	services "seams-backend/internal/services/orderService"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Request = services.CreateOrderRequest

type Response struct {
	resp.Response
	OrderID   string `json:"order_id"`
	InvoiceID string `json:"invoice_id"`
	Total     int64  `json:"total"`
}

type OrderSaver interface {
	CreateOrderWithCustomerAndInvoice(
		ctx context.Context,
		req *services.CreateOrderRequest,
	) (*services.CreateOrderResponse, error)
}

func New(log *slog.Logger, saver OrderSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.orders.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

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

		result, err := saver.CreateOrderWithCustomerAndInvoice(r.Context(), &req)
		if err != nil {
			log.Error("failed to create order", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to create order"))
			return
		}

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, &Response{
			Response:  resp.OK(),
			OrderID:   result.OrderID.String(),
			InvoiceID: result.InvoiceID.String(),
			Total:     result.Total,
		})
	}

}
