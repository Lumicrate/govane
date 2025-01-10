package gateway

import "net/http"

type MiddlewareFunc func(http.Handler, map[string]any) http.Handler

type MiddlewareEntry struct {
	Handler    MiddlewareFunc
	ApplyToAll bool
}

var MiddlewareRegistry = make(map[string]MiddlewareEntry)

func RegisterMiddleware(name string, middleware MiddlewareFunc, applyToAll bool) {
	MiddlewareRegistry[name] = MiddlewareEntry{
		Handler:    middleware,
		ApplyToAll: applyToAll,
	}
}

func ApplyMiddleware(handler http.Handler, middlewareNames []string, metadata map[string]any) http.Handler {
	for _, entry := range MiddlewareRegistry {
		if entry.ApplyToAll {
			handler = entry.Handler(handler, metadata)
		}
	}

	for _, name := range middlewareNames {
		if entry, exists := MiddlewareRegistry[name]; exists {
			handler = entry.Handler(handler, metadata)
		}
	}

	return handler
}
