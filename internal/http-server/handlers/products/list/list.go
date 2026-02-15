package list

import (
	"context"
	"log/slog"
	"net/http"
	resp "seams-backend/internal/lib/api/response"
	"seams-backend/internal/lib/logger/sl"
	"seams-backend/internal/models"
	"strconv"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Response struct {
	resp.Response
	Products []*models.Product `json:"products,omitempty"`
}

type ProductsGetter interface {
	GetProducts(ctx context.Context, search string, page, limit int) ([]*models.Product, error)
}

func New(log *slog.Logger, getter ProductsGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.products.get.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request-id", middleware.GetReqID(r.Context())),
		)

		page := 1
		limit := 20

		if v := r.URL.Query().Get("page"); v != "" {
			p, err := strconv.Atoi(v)
			if err != nil || p < 1 {
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, resp.Error("invalid page"))
				return
			}
			page = p
		}

		if v := r.URL.Query().Get("limit"); v != "" {
			l, err := strconv.Atoi(v)
			if err != nil || l <= 0 || l > 100 {
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, resp.Error("invalid limit"))
				return
			}
			limit = l
		}

		search := r.URL.Query().Get("search")

		products, err := getter.GetProducts(r.Context(), search, page, limit)

		if err != nil {
			log.Error("failed to get products", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal server error"))

			return
		}

		render.Status(r, http.StatusOK)
		responseOK(w, r, products)

	}
}
func responseOK(w http.ResponseWriter, r *http.Request, products []*models.Product) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Products: products,
	})
}
