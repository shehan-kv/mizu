package server

import (
	"mizu/internal/auth"
	"mizu/internal/db"
	v1 "mizu/internal/handler/v1"
	"mizu/internal/logger"
	"mizu/internal/service"
	"mizu/internal/session"
	"net/http"
	"os"
)

// Starts the REST API server.
func RunServer() {

	// Initialize logger
	logger := logger.NewSlogLogger()

	// Initialize database and clean up
	db := db.New(logger)
	db.Init()
	defer db.Close()

	// Initialize session store
	sessionStore := session.GetSessionStore(logger)
	sessionStore.Init()
	defer sessionStore.Close()

	// Initialize database stores
	userStore := db.NewUserStore()

	// Default admin user when the database has no users
	auth.CreateDefaultAdminUser(userStore, logger)

	// Services
	authService := service.NewAuthService(logger, userStore, sessionStore)

	// Handler mux init
	authMux := v1.NewAuthHandler(authService).GetMux(logger)

	// Server routes
	mainMux := http.NewServeMux()
	mainMux.Handle("/api/v1/auth/", http.StripPrefix("/api/v1/auth", authMux))

	// Start server
	listenOn := os.Getenv("LISTEN_ON")
	if listenOn == "" {
		listenOn = ":8080"
	}

	logger.Info("server listening...", "port", "8080")
	http.ListenAndServe(listenOn, mainMux)
}
