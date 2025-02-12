package asres

import (
	"api/internal/http/aserr"
	"encoding/json"
	"net/http"
)

type Response[T any] struct {
	Data           T       `json:"data,omitempty"`
	HTTPStatusCode int     `json:"-"`
	Redirect       *string `json:"redirect,omitempty"`
}

func NewResponse[T any](data T, statusCode int) *Response[T] {
	return &Response[T]{Data: data, HTTPStatusCode: statusCode}
}

func NewRedirect(redirect string, statusCode int) *Response[string] {
	return &Response[string]{Redirect: &redirect, HTTPStatusCode: statusCode}
}

type HandlerFunc[T any] func(http.ResponseWriter, *http.Request) (*Response[T], error)

func ToHandlerFunc[T any](f HandlerFunc[T]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := f(w, r)
		if data != nil && data.Redirect != nil {
			http.Redirect(w, r, *data.Redirect, data.HTTPStatusCode)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		if err != nil {
			asErr, ok := err.(*aserr.ASError)
			if ok {
				w.WriteHeader(asErr.HTTPStatusCode)
				json.NewEncoder(w).Encode(asErr)
				return
			}

			w.WriteHeader(aserr.ErrInternalServerError.HTTPStatusCode)
			json.NewEncoder(w).Encode(aserr.ErrInternalServerError)
			return
		}

		w.WriteHeader(data.HTTPStatusCode)
		json.NewEncoder(w).Encode(data.Data)
	}
}
