package server

import (
	"mizu/internal/auth"
	"mizu/internal/db"
	"mizu/internal/email"
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

	// Initialize email sender
	emailSender := email.GetEmailSender()
	emailSender.Init()
	defer emailSender.Close()

	// Initialize database stores
	userStore := db.NewUserStore()
	projectStore := db.NewProjectStore()
	invoiceStore := db.NewInvoiceStore()
	messageStore := db.NewMessagetore()

	// Default admin user when the database has no users
	auth.CreateDefaultAdminUser(userStore, logger)

	// Services
	authService := service.NewAuthService(logger, userStore, sessionStore)
	projectService := service.NewProjectService(logger, projectStore)
	invoiceService := service.NewInvoiceService(logger, invoiceStore)
	userService := service.NewUserService(logger, userStore, emailSender)
	messageService := service.NewMessageService(logger, messageStore)

	// Handler mux init
	authMux := v1.NewAuthHandler(authService).GetMux(logger)
	projectMux := v1.NewProjectHandler(projectService).GetMux(logger, sessionStore, userStore)
	invoiceMux := v1.NewInvoiceHandler(invoiceService).GetMux(logger, sessionStore, userStore)
	userMux := v1.NewUserHandler(userService).GetMux(logger, sessionStore, userStore)
	messageMux := v1.NewMessageHandler(messageService).GetMux(logger, sessionStore, userStore)

	// Server routes
	mainMux := http.NewServeMux()
	mainMux.Handle("/api/v1/auth/", http.StripPrefix("/api/v1/auth", authMux))
	mainMux.Handle("/api/v1/projects/", http.StripPrefix("/api/v1/projects", projectMux))
	mainMux.Handle("/api/v1/invoices/", http.StripPrefix("/api/v1/invoices", invoiceMux))
	mainMux.Handle("/api/v1/users/", http.StripPrefix("/api/v1/users", userMux))
	mainMux.Handle("/api/v1/messages/", http.StripPrefix("/api/v1/messages", messageMux))

	// Start server
	listenOn := os.Getenv("LISTEN_ON")
	if listenOn == "" {
		listenOn = ":8080"
	}

	logger.Info("server listening...", "port", listenOn)
	http.ListenAndServe(listenOn, mainMux)
}
