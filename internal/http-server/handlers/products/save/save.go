package save

import (
	"context"
	"log/slog"
	"net/http"
	resp "seams-backend/internal/lib/api/response"
	"seams-backend/internal/lib/logger/sl"
	"seams-backend/internal/models"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type Request struct {
	Name        string  `json:"name"`
	SKU         string  `json:"sku"`
	Price       int64   `json:"price"`
	CategoryID  string  `json:"category_id"`
	Description *string `json:"description,omitempty"`
}

type Response struct {
	resp.Response
	ID string `json:"id,omitempty"`
}

type ProductSaver interface {
	SaveProduct(ctx context.Context, product *models.Product) (uuid.UUID, error)
}

func New(log *slog.Logger, saver ProductSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.products.save.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request-id", middleware.GetReqID(r.Context())),
		)

		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("invalid request body"))

			return
		}

		categoryID, err := uuid.Parse(req.CategoryID)
		if err != nil {
			log.Error("invalid category ID", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("invalid category ID"))

			return
		}

		if req.Name == "" || req.SKU == "" || req.Price <= 0 {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("invalid product data"))
			return
		}

		product := &models.Product{
			Name:        req.Name,
			Sku:         req.SKU,
			Price:       req.Price,
			CategoryID:  categoryID,
			Description: req.Description,
		}

		id, err := saver.SaveProduct(r.Context(), product)
		if err != nil {
			log.Error("failed to save product", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal server error"))

			return
		}

		render.Status(r, http.StatusCreated)
		responseOK(w, r, id)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		ID:       id.String(),
	})
}
