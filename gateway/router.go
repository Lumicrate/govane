package gateway

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
)

func CreateReverseProxy(d Downstream, upstreamRoute string) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(&url.URL{Scheme: "http", Host: fmt.Sprintf("%s:%d", d.Host, d.Port)})

	proxy.Director = func(req *http.Request) {
		log.Printf("Incoming request path: %s", req.URL.Path)
		req.URL.Scheme = "http"
		req.URL.Host = fmt.Sprintf("%s:%d", d.Host, d.Port)
		req.URL.Path = handlePathParams(upstreamRoute, d.Route, req.URL.Path)
		log.Printf("Forwarding to: %s%s", req.URL.Host, req.URL.Path)
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Error during proxy: %v", err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
	}

	return proxy
}

func handlePathParams(upstreamRoute, downstreamRoute, requestPath string) string {
	paramPattern := `\{([^}]+)\}`
	re := regexp.MustCompile(paramPattern)

	upstreamSegments := strings.Split(upstreamRoute, "/")
	requestSegments := strings.Split(requestPath, "/")

	for i, segment := range upstreamSegments {
		if re.MatchString(segment) {
			paramName := segment[1 : len(segment)-1]
			if i < len(requestSegments) {
				downstreamRoute = strings.Replace(downstreamRoute, "{"+paramName+"}", requestSegments[i], -1)
				log.Printf("Extracted path parameter: %s=%s", paramName, requestSegments[i])
			}
		}
	}

	return downstreamRoute
}

func SetupRoutes(config *Config) *http.ServeMux {
	mux := http.NewServeMux()

	for _, route := range config.Routes {
		proxy := CreateReverseProxy(route.Downstream, route.Upstream) // Pass upstream route dynamically
		upstream := route.Upstream
		allowedMethods := map[string]bool{}
		for _, method := range route.Methods {
			allowedMethods[strings.ToUpper(method)] = true
		}

		baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !allowedMethods[r.Method] {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			proxy.ServeHTTP(w, r)
		})

		handlerWithMiddleware := ApplyMiddleware(baseHandler, route.Middleware, route.Metadata)

		if route.Auth != nil {
			handlerWithMiddleware = JWTMiddleware(handlerWithMiddleware, route.Auth)
		}

		mux.Handle(upstream, handlerWithMiddleware)
	}

	return mux
}
