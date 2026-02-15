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
	Categories []*models.Category `json:"categories"`
}

type CategoriesGetter interface {
	GetCategories(ctx context.Context) ([]*models.Category, error)
}

func New(log *slog.Logger, getter CategoriesGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.categories.get.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request-id", middleware.GetReqID(r.Context())),
		)

		categories, err := getter.GetCategories(r.Context())
		if err != nil {
			log.Error("failed to get categories", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal server error"))

			return
		}

		render.Status(r, http.StatusOK)
		responseOK(w, r, categories)

	}
}

func responseOK(w http.ResponseWriter, r *http.Request, categories []*models.Category) {
	render.JSON(w, r, Response{
		Response:   resp.OK(),
		Categories: categories,
	})
}
