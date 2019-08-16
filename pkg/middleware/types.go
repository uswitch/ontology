package middleware

import (
	"net/http"
)

type Middleware interface {
	Middleware(http.Handler) http.Handler
}

func Wrap(middleware []Middleware, handler http.Handler) http.Handler {
	h := handler

	for i := len(middleware) - 1; i >= 0; i = i - 1 {
		h = middleware[i].Middleware(h)
	}

	return h
}

type MiddlewareFunc func(http.Handler) http.Handler

func (fn MiddlewareFunc) Middleware(r http.Handler) http.Handler {
	return fn(r)
}

var PassThru = MiddlewareFunc(func(h http.Handler) http.Handler {
	return h
})
