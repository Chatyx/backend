package middleware

import "net/http"

type Middleware func(next http.Handler) http.Handler

func Chain(handler http.Handler, ms ...Middleware) http.Handler {
	if len(ms) == 0 {
		return handler
	}

	for i := len(ms) - 1; i >= 0; i-- {
		handler = ms[i](handler)
	}
	return handler
}
