package save

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	resp "seams-backend/internal/lib/api/response"
	"seams-backend/internal/lib/logger/sl"
	requestservice "seams-backend/internal/services/requestService"
	services "seams-backend/internal/services/requestService"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type Request = services.CreateRequestReq

type Response struct {
	resp.Response
	OrderID string `json:"order_id"`
}

type RequestSaver interface {
	CreateRequestWithCustomer(ctx context.Context, req *services.CreateRequestReq) (*services.CreateRequestResponse, error)
}

func New(log *slog.Logger, saver RequestSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.requests.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		err := r.ParseMultipartForm(10 << 20)

		name := r.FormValue("name")
		email := r.FormValue("email")
		phone := r.FormValue("phone")
		desc := r.FormValue("desc")

		file, handler, err := r.FormFile("file")

		var filePath string

		if err == nil {
			defer file.Close()

			id := uuid.New().String()
			ext := filepath.Ext(handler.Filename)

			filePath = "./uploads/" + id + ext

			dst, err := os.Create(filePath)
			if err != nil {
				http.Error(w, "cannot save file", http.StatusInternalServerError)
				return
			}
			defer dst.Close()

			io.Copy(dst, file)

		}

		req := &requestservice.CreateRequestReq{
			Name:        name,
			Email:       email,
			Phone:       phone,
			Description: desc,
			FilePath:    filePath,
		}

		result, err := saver.CreateRequestWithCustomer(r.Context(), req)
		if err != nil {
			log.Error("failed to create request", sl.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to create request"))
			return
		}

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, &Response{
			Response: resp.OK(),
			OrderID:  result.RequestID.String(),
		})
	}
}
