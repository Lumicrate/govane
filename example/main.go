package main

import (
	"github.com/Lumicrate/govane/gateway"
	"log"
	"net/http"
)

// Example Middleware
func LoggingMiddleware(next http.Handler, metadata map[string]any) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if serviceName, ok := metadata["service_name"].(string); ok {
			log.Printf("Request for service: %s, Method: %s, Path: %s", serviceName, r.Method, r.URL.Path)
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Register user-defined middleware
	gateway.RegisterMiddleware("LoggingMiddleware", LoggingMiddleware, true) // Apply to all

	// Load configuration
	config, err := gateway.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set up routes
	routes := gateway.SetupRoutes(config)

	// Start the server
	log.Println("API Gateway running on port 8080...")
	if err := http.ListenAndServe(":8080", routes); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
