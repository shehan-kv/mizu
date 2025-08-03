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
	projectStore := db.NewProjectStore()
	invoiceStore := db.NewInvoiceStore()

	// Default admin user when the database has no users
	auth.CreateDefaultAdminUser(userStore, logger)

	// Services
	authService := service.NewAuthService(logger, userStore, sessionStore)
	projectService := service.NewProjectService(logger, projectStore)
	invoiceService := service.NewInvoiceService(logger, invoiceStore)

	// Handler mux init
	authMux := v1.NewAuthHandler(authService).GetMux(logger)
	projectMux := v1.NewProjectHandler(projectService).GetMux(logger, sessionStore, userStore)
	invoiceMux := v1.NewInvoiceHandler(invoiceService).GetMux(logger, sessionStore, userStore)

	// Server routes
	mainMux := http.NewServeMux()
	mainMux.Handle("/api/v1/auth/", http.StripPrefix("/api/v1/auth", authMux))
	mainMux.Handle("/api/v1/project/", http.StripPrefix("/api/v1/project", projectMux))
	mainMux.Handle("/api/v1/invoice/", http.StripPrefix("/api/v1/invoice", invoiceMux))

	// Start server
	listenOn := os.Getenv("LISTEN_ON")
	if listenOn == "" {
		listenOn = ":8080"
	}

	logger.Info("server listening...", "port", "8080")
	http.ListenAndServe(listenOn, mainMux)
}
