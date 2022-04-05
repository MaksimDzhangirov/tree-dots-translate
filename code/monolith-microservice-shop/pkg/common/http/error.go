package http

import (
	"net/http"

	"github.com/go-chi/render"
)

type ErrResponse struct {
	Err            error `json:"-"` // низкоуровневая ошибка времени выполнения
	HTTPStatusCode int   `json:"-"` // код состояния http ответа

	AppCode   int64  `json:"code,omitempty"`  // код ошибки, зависящий от приложения
	ErrorText string `json:"error,omitempty"` // сообщение об ошибке в слое приложения
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInternal(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusInternalServerError,
		ErrorText:      err.Error(),
	}
}

func ErrBadRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusBadRequest,
		ErrorText:      err.Error(),
	}
}
