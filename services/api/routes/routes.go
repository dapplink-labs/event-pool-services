package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v4"

	"github.com/multimarket-labs/event-pod-services/services/api/service"
)

var (
	jwtSecret  = []byte("my_token_secret_123456") // todo: replace more security token
	CapitalKey = "%s:capital"
)

type Claims struct {
	BusinessId string `json:"business_id"`
	jwt.RegisteredClaims
}

type Routes struct {
	router *chi.Mux
	svc    service.Service
}

// NewRoutes ... Construct a new route handler instance
func NewRoutes(r *chi.Mux, svc service.Service) Routes {
	rs := Routes{
		router: r,
		svc:    svc,
	}

	// Register predict event route (Dify integration)
	r.Post("/api/v1/admin/predict-event", rs.PredictEventHandler)

	// Register event routes
	r.Post("/api/v1/events", rs.CreateEventHandler)
	r.Get("/api/v1/events", rs.ListEventsHandler)

	return rs
}

func (rs *Routes) validateTokenStr(uid string, tokenRaw string) bool {
	return false
}

func (rs *Routes) JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing Authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "invalid Authorization header", http.StatusUnauthorized)
			return
		}
		tokenStr := parts[1]

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		if !rs.validateTokenStr(claims.BusinessId, tokenStr) {
			http.Error(w, "invalid or expired token,not same", http.StatusUnauthorized)
			return
		}
		return
	})
}
