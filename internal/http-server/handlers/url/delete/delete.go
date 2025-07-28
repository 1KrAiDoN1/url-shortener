package delete

import (
	"context"
	"log/slog"
	"net/http"
	"url-shortener/internal/lib/api/response"
	slogger "url-shortener/internal/lib/logger/slog"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

func New(ctx context.Context, log *slog.Logger, storage *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.New"
		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")

		if alias == "" {
			log.Info("empty alias")
			render.JSON(w, r, response.Error("empty alias"))
			return
		}

		err := storage.DeleteURl(ctx, alias)
		if err != nil {
			log.Error("failed to delete url", slogger.Err(err), slog.String("alias", alias))
			render.JSON(w, r, response.Error("internal error"))
			return
		}

		log.Info("url deleted", slog.String("alias", alias))

		render.JSON(w, r, response.OK())
	}
}
