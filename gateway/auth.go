package gateway

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"log"
	"net/http"
	"strings"
)

func JWTMiddleware(next http.Handler, authConfig *AuthConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authConfig == nil {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		tokenStr := tokenParts[1]

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(authConfig.Key), nil
		})

		if err != nil || !token.Valid {
			log.Printf("JWT validation failed: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		roleClaim, ok := claims[authConfig.RoleClaimKey]
		if !ok {
			http.Error(w, fmt.Sprintf("'%s' not found in token", authConfig.RoleClaimKey), http.StatusForbidden)
			return
		}

		role, ok := roleClaim.(string)
		if !ok {
			http.Error(w, fmt.Sprintf("'%s' is not a string in token", authConfig.RoleClaimKey), http.StatusForbidden)
			return
		}

		isAuthorized := false
		for _, allowedValue := range authConfig.AllowedValues {
			if role == allowedValue {
				isAuthorized = true
				break
			}
		}

		if !isAuthorized {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
