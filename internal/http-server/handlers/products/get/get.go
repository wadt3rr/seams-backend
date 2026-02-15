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

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type Response struct {
	resp.Response
	Product *models.Product `json:"product,omitempty"`
}

type ProductGetter interface {
	GetProductByID(ctx context.Context, id uuid.UUID) (*models.Product, error)
}

func New(log *slog.Logger, getter ProductGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.product.get.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request-id", middleware.GetReqID(r.Context())),
		)

		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			log.Error("missing product id in request")

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("missing product id"))

			return
		}
		id, err := uuid.Parse(idStr)
		if err != nil {
			log.Error("invalid product id format", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("invalid product id format"))

			return
		}

		prod, err := getter.GetProductByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				log.Error("product not found", sl.Err(err))

				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, resp.Error("product not found"))

				return
			}

			log.Error("failed to get product by id", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal server error"))

			return
		}
		render.Status(r, http.StatusOK)
		responseOK(w, r, prod)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, product *models.Product) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Product:  product,
	})
}
