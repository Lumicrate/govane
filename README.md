# GoVane

This is an API Gateway called GoVane designed for routing and forwarding requests between clients and microservices in a microservice architecture. The gateway supports JWT-based authentication and authorization, custom middleware, and request forwarding to the downstream services. It is highly configurable through a structured `config.json` file, enabling flexible route and security configurations.

## Features

- **Routing**: Forward requests from upstream routes to downstream microservices.
- **Authentication & Authorization**: Supports JWT-based authentication and role-based access control. The roles are dynamically configurable in the `config.json` file.
- **Middleware Support**: Allows users to define custom middleware that can be applied on routes for pre-processing requests.
- **Configurable through JSON**: All route, middleware, and authentication configurations are handled in a single `config.json` file.
- **Path Parameter Handling**: Automatically maps and forwards requests with dynamic path parameters between the upstream and downstream services.

## Installation

### Prerequisites

- Go (Golang) version 1.23 or higher
- JSON configuration file (`config.json`)

### Step 1: Clone the Repository

```bash
go get github.com/Lumicrate/govane
```
### Step 2: Install Dependencies
```bash 
go mod tidy
```

### Step 3: Configuration
Create a `config.json` file to configure the routes, authentication, and middleware.

Example `config.json`

```bash
{
  "routes": [
    {
      "downstream": {
        "host": "127.0.0.1",
        "port": 8000,
        "route": "/blogs/{id}"
      },
      "upstream": "/api/service1/{id}",
      "methods": ["GET", "POST"],
      "metadata": {
        "service_name": "Service1",
        "service_id": 123
      },
      "middleware": ["LoggingMiddleware"],
      "auth": {
        "key": "mysecretkey",
        "roleClaimKey": "UserType",
        "allowedValues": ["admin", "user"]
      }
    }
  ]
}
```

### Step 4: Load the `config.json` file

```bash
func main() {
	// Load configuration
	config, err := gateway.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
}

```

### Step 5 : Setup Routes

```bash
func main() {
	// Load configuration
	config, err := gateway.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set up routes
	routes := gateway.SetupRoutes(config)

	// Start the server
	log.Println("GoVane running on port 8080...")
	if err := http.ListenAndServe(":8080", routes); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
```

### Step 6: Start the Gateway
Run the following command to start the gateway:
```bash 
go run main.go
```
This will start the gateway and it will begin listening on the configured port.

## Configuration Breakdown
### Routes Configuration
Each route in the `config.json` file contains the following fields:

- **downstream**: The details for the downstream service to which the request is forwarded.
    - `host`: The host of the downstream service.
    - `port`: The port of the downstream service.
    - `route`: The route on the downstream service (can contain path parameters).

- **upstream**: The route that clients will use to access the service via the API Gateway (can also contain path parameters).

- **methods**: An array of allowed HTTP methods (e.g., `GET`, `POST`).

- **metadata**: Optional metadata for the service, which can be used for custom middleware development.

- **middleware**: An array of custom middleware functions to be applied to this route.

- **auth**: Configuration for JWT-based authentication.

    - `key`: The secret key used to validate JWT tokens.
    - `roleClaimKey`: The key in the JWT claim that holds the role (e.g., UserType).
    - `allowedValues`: An array of roles that have access to this route.

### Example Route Configuration
The configuration below defines a route that forwards requests from `/api/service1/{id}` to `/blogs/{id}` in the downstream service.

```bash 
{
  "downstream": {
    "host": "127.0.0.1",
    "port": 8000,
    "route": "/blogs/{id}"
  },
  "upstream": "/api/service1/{id}",
  "methods": ["GET", "POST"],
  "metadata": {
    "service_name": "Service1",
    "service_id": 123
  },
  "middleware": ["LoggingMiddleware"],
  "auth": {
    "key": "mysecretkey",
    "roleClaimKey": "UserType",
    "allowedValues": ["admin", "user"]
  }
}
```
### Middleware
The gateway allows the use of custom middleware. Middleware functions can be defined and applied to one or more routes. Middleware functions can inspect or modify the request before it's forwarded to the downstream service.

### Example: LoggingMiddleware

#### Step 1: Add you custom middleware
```bash
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("Received request: %s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })
}
```

#### Step 2: Register it with a specific name
Registration requires 3 options:
 - `middleware name`: provide a unique name for the custom middleware.
 - `middleware function`: provide a middleware function. 
 - `applyToAll`: a boolean option for applying the middleware to all routes. 
    - `true`: it applies middleware to all routes and there is no need to be specified in the `config.json`.
    - `false`: middleware won't be applied to all the routes and it must be specified in the `config.json` file in the `middleware` section.

```bash
func main() {
	// Register user-defined middleware
	gateway.RegisterMiddleware("LoggingMiddleware", LoggingMiddleware, true) // Apply to all
}
```

#### Step 3: Assign it to the route
Assign the middleware to the route it `applyToAll` is `false`.
```bash
{
  "downstream": {
    "host": "127.0.0.1",
    "port": 8000,
    "route": "/blogs/{id}"
  },
  "upstream": "/api/service1/{id}",
  "methods": ["GET", "POST"],
  "middleware": ["LoggingMiddleware"],

}
```

### Authentication & Authorization
JWT-based authentication is supported in the gateway. The gateway will:

1. Validate the JWT token in the `Authorization` header.

2. Extract the role (or equivalent) from the token using the configurable `roleClaimKey`.

3. Check if the user's role matches one of the allowed roles defined in the `allowedValues` array.

If the JWT is invalid or the role does not match, the request will be rejected with a `401 Unauthorized` or `403 Forbidden` error.

### Path Parameter Mapping
The API Gateway supports dynamic path parameters in both the upstream and downstream routes. Path parameters in the `upstream` route (e.g., `/api/service1/{id}`) will be extracted and passed to the downstream route (e.g., `/blogs/{id}`).

## Example Use Case
### Upstream Request
A client makes a request to the API Gateway:
```bash
GET /api/service1/123
```

### Gateway Logic
1. The gateway will look up the route configuration.
2. It will extract the `id` path parameter from the upstream route (`/api/service1/{id}`).
3. The `id` will be passed to the downstream route `/blogs/{id}`.
4. The gateway will forward the request to the downstream service (`http://127.0.0.1:8000/blogs/123`).

### Response
The response from the downstream service will be sent back to the client.

## Error Handling
If any errors occur during the request forwarding or validation process, the API Gateway will return a suitable HTTP error:

 - `401 Unauthorized`: If JWT authentication fails.
 - `403 Forbidden`: If the user's role does not have access to the route.
 - `404 Not Found`: If the route is not found.
 - `500 Internal Server Error`: For any other server errors.

## Contributing
We welcome contributions to this project! If you'd like to contribute:

1. Fork the repository.
2. Create a new branch.
3. Implement your changes.
4. Submit a pull request with a detailed description of the changes.

## License
This project is licensed under the MIT License - see the LICENSE file for details.
