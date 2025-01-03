package router

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"ledger-link/internal/handlers"
	"ledger-link/pkg/httputil"
	"ledger-link/pkg/middleware"
)

func getIDFromPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func NewRouter(
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	transactionHandler *handlers.TransactionHandler,
	balanceHandler *handlers.BalanceHandler,
	authMiddleware *middleware.AuthMiddleware,
	rbacMiddleware *middleware.RBACMiddleware,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("/api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("/api/v1/auth/refresh", authHandler.RefreshToken)

	mux.HandleFunc("/api/v1/users", func(w http.ResponseWriter, r *http.Request) {
		authMiddleware.Authenticate(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					userHandler.GetUsers(w, r)
				default:
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			}),
		).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/v1/users/", func(w http.ResponseWriter, r *http.Request) {
		id := getIDFromPath(r.URL.Path)
		if id == "" {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), httputil.PathParamsKey, map[string]string{"id": id})
		r = r.WithContext(ctx)

		authMiddleware.Authenticate(
			rbacMiddleware.RequireOwnerOrAdmin(func(r *http.Request) uint {
				idStr := httputil.GetPathParam(r.Context(), "id")
				id, _ := strconv.ParseUint(idStr, 10, 32)
				return uint(id)
			})(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.Method {
					case http.MethodGet:
						userHandler.GetUser(w, r)
					case http.MethodPut:
						userHandler.UpdateUser(w, r)
					case http.MethodDelete:
						userHandler.DeleteUser(w, r)
					default:
						http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
					}
				}),
			),
		).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/v1/transactions/history", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		authMiddleware.Authenticate(
			http.HandlerFunc(transactionHandler.HandleGetTransactionHistory),
		).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/v1/transactions/credit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		authMiddleware.Authenticate(
			http.HandlerFunc(transactionHandler.HandleCredit),
		).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/v1/transactions/debit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		authMiddleware.Authenticate(
			http.HandlerFunc(transactionHandler.HandleDebit),
		).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/v1/transactions/transfer", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		authMiddleware.Authenticate(
			http.HandlerFunc(transactionHandler.HandleTransfer),
		).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/v1/transactions/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id := getIDFromPath(r.URL.Path)
		if id == "" {
			http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), httputil.PathParamsKey, map[string]string{"id": id})
		r = r.WithContext(ctx)

		authMiddleware.Authenticate(
			http.HandlerFunc(transactionHandler.HandleGetTransaction),
		).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/v1/balances/current", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		authMiddleware.Authenticate(
			http.HandlerFunc(balanceHandler.GetCurrentBalance),
		).ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/v1/balances/history", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		authMiddleware.Authenticate(
			http.HandlerFunc(balanceHandler.GetBalanceHistory),
		).ServeHTTP(w, r)
	})

	return mux
}
