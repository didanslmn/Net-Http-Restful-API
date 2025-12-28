package routers

import (
	"net/http"
	"postgresDB/config"
	"postgresDB/internal/delivery/handler"
	"postgresDB/internal/delivery/middleware"
	"postgresDB/internal/domain/entities"
	"postgresDB/pkg/jwt"
)

// Router sets up all routes for the application
type Router struct {
	// Define router fields here
	mux            *http.ServeMux
	authHandler    *handler.AuthHandler
	userHandler    *handler.UserHandler
	productHandler *handler.ProductHandler
	orderHandler   *handler.OrderHandler
	jwtService     *jwt.JWTService
	cfg            *config.Config
}

// NewRouter creates a new Router instance
func NewRouter(
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	productHandler *handler.ProductHandler,
	orderHandler *handler.OrderHandler,
	jwtService *jwt.JWTService,
	cfg *config.Config,
) *Router {
	return &Router{
		mux:            http.NewServeMux(),
		authHandler:    authHandler,
		userHandler:    userHandler,
		productHandler: productHandler,
		orderHandler:   orderHandler,
		jwtService:     jwtService,
		cfg:            cfg,
	}
}

// setupRoutes configures all application routes
func (r *Router) SetupRoutes() http.Handler {
	// Health check route
	r.mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	// Auth routes (public)
	r.mux.HandleFunc("POST /api/v1/auth/register", r.authHandler.Register)
	r.mux.HandleFunc("POST /api/v1/auth/login", r.authHandler.Login)
	r.mux.HandleFunc("POST /api/v1/auth/refresh", r.authHandler.RefreshToken)
	// Auth routes (protected)
	r.mux.Handle("POST /api/v1/auth/logout", r.withAuth(http.HandlerFunc(r.authHandler.Logout)))
	r.mux.Handle("POST /api/v1/auth/revoke", r.withAuth(http.HandlerFunc(r.authHandler.RevokeAllSessions)))

	// User routes (protected)
	r.mux.Handle("GET /api/v1/users", r.withAuth(http.HandlerFunc(r.userHandler.GetProfile)))                           // GET all users (admin only)
	r.mux.Handle("GET /api/v1/users/{id}", r.withAuth(http.HandlerFunc(r.userHandler.GetProfile)))                      // GET user by ID
	r.mux.Handle("PUT /api/v1/users/{id}", r.withAuth(http.HandlerFunc(r.userHandler.UpdateUser)))                      // PUT/PATCH update user
	r.mux.Handle("POST /api/v1/users/{id}/change-password", r.withAuth(http.HandlerFunc(r.userHandler.ChangePassword))) // POST change password
	r.mux.Handle("DELETE /api/v1/users/{id}", r.withAuth(http.HandlerFunc(r.userHandler.DeleteUser)))                   // DELETE user

	// Admin management user (protected)
	//r.mux.Handle("PUT /api/v1/users/{id}", r.withAuthAndRole(http.HandlerFunc(r.userHandler.UpdateUser), entities.RoleAdmin))

	// Product routes (public)
	r.mux.HandleFunc("GET /api/v1/products", http.HandlerFunc(r.productHandler.List))
	r.mux.HandleFunc("GET /api/v1/products/{id}", http.HandlerFunc(r.productHandler.GetByID))

	// Product routes (protected)
	r.mux.Handle("POST /api/v1/products", r.withAuthAndRole(http.HandlerFunc(r.productHandler.CreateProduct), entities.RoleAdmin))
	r.mux.Handle("PUT /api/v1/products/{id}", r.withAuthAndRole(http.HandlerFunc(r.productHandler.Update), entities.RoleAdmin))
	r.mux.Handle("PATCH /api/products/{id}", r.withAuthAndRole(http.HandlerFunc(r.productHandler.Update), entities.RoleAdmin))
	r.mux.Handle("DELETE /api/v1/products/{id}", r.withAuthAndRole(http.HandlerFunc(r.productHandler.Delete), entities.RoleAdmin))

	// Order routes (protected)
	r.mux.Handle("GET /api/v1/orders", r.withAuth(http.HandlerFunc(r.orderHandler.ListOrders)))
	r.mux.Handle("GET /api/v1/orders/{id}", r.withAuth(http.HandlerFunc(r.orderHandler.GetOrderByID)))
	r.mux.Handle("POST /api/v1/orders", r.withAuthAndRole(http.HandlerFunc(r.orderHandler.CreateOrder), entities.RoleUser))
	r.mux.Handle("PATCH /api/v1/orders/{id}/status", r.withAuthAndRole(http.HandlerFunc(r.orderHandler.UpdateOrderStatus), entities.RoleAdmin))

	return nil
}

// withAuthMiddleware applies authentication middleware to protected routes
func (r *Router) withAuth(h http.Handler) http.Handler {
	return middleware.Auth(r.jwtService)(h)
}

// withAuthAndRole wraps a handler with authentication and role middleware
func (r *Router) withAuthAndRole(h http.Handler, roles ...entities.Role) http.Handler {
	return middleware.Auth(r.jwtService)(
		middleware.RequireRole(roles...)(h),
	)
}
