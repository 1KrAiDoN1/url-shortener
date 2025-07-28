package save

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"url-shortener/internal/lib/api/random"
	"url-shortener/internal/lib/api/response"
	resp "url-shortener/internal/lib/api/response"
	slogger "url-shortener/internal/lib/logger/slog"
	"url-shortener/internal/service"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

const aliasLength = 6

type Handlers struct {
	service service.ServiceInterface
}

func NewHandlers(service service.ServiceInterface) *Handlers {
	return &Handlers{
		service: service,
	}
}

type HandlersInterface interface {
	New(ctx context.Context, urlToSave string, alias string) (int64, error)
	GetURL(ctx context.Context, alias string) (string, error)
	DeleteURl(ctx context.Context, alias string) error
}

func (h *Handlers) New(ctx context.Context, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.url.save.New"
		log = log.With(slog.String("op", op), slog.String("request_id", middleware.GetReqID(r.Context())))

		var req Request
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			log.Error("failed to decode request body", slogger.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to decode request body"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			log.Error("invalid request", slogger.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.ValidationError(err.(validator.ValidationErrors)))
			return
		}

		res, err := h.service.URLExists(ctx, req.URL)
		if err != nil {
			log.Error("failed to check url existence", slogger.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to check url existence"))
			return
		}
		if res {
			log.Info("url already exists", slog.Any("url", req.URL))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("url already exists"))
			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}

		id, err := h.service.SaveURL(ctx, req.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Info("url already exists", slog.String("url", req.URL))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("url already exists"))
			return
		}
		if err != nil {
			log.Error("failed to add url", slogger.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to add url"))
			return
		}

		log.Info("url added", slog.Int64("id", id))
		w.WriteHeader(http.StatusOK)
		responseOK(w, r, id, alias)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, id int64, alias string) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Id:       id,
		Alias:    alias,
	})
}

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Id    int64  `json:"id"`
	Alias string `json:"alias,omitempty"`
}
