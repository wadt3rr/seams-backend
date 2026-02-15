package bysku

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
)

type Response struct {
	resp.Response
	Product *models.Product `json:"product,omitempty"`
}

type ProductGetter interface {
	GetProductBySku(ctx context.Context, sku string) (*models.Product, error)
}

func New(log *slog.Logger, getter ProductGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.products.bySku.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request-id", middleware.GetReqID(r.Context())),
		)

		log.Debug("request path", slog.String("path", r.URL.Path))
		rc := chi.RouteContext(r.Context())
		if rc != nil {
			for i := 0; i < len(rc.URLParams.Keys); i++ {
				log.Debug("route param", slog.String(rc.URLParams.Keys[i], rc.URLParams.Values[i]))
			}
		}

		sku := chi.URLParam(r, "sku")
		log.Debug("sku parameter", slog.String("sku", sku))
		if sku == "" {
			log.Error("missing product sku in request")

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("missing product sku"))

			return
		}

		prod, err := getter.GetProductBySku(r.Context(), sku)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				log.Error("product not found", sl.Err(err))

				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, resp.Error("product not found"))

				return
			}

			render.Status(r, http.StatusInternalServerError)
			log.Error("failed to get product by sku", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))

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
