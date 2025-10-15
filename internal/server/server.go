package server

import (
	"mizu/internal/auth"
	"mizu/internal/db"
	"mizu/internal/email"
	"mizu/internal/event"
	"mizu/internal/file"
	v1 "mizu/internal/handler/v1"
	"mizu/internal/logger"
	"mizu/internal/service"
	"mizu/internal/session"
	"mizu/internal/sse"
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

	// Initialize file storage
	fileStorage := file.GetFileStorage(logger)
	defer fileStorage.Close()

	// Initialize email sender
	emailSender := email.GetEmailSender()
	emailSender.Init()
	defer emailSender.Close()

	// Initialize database stores
	userStore := db.NewUserStore()
	projectStore := db.NewProjectStore()
	invoiceStore := db.NewInvoiceStore()
	messageStore := db.NewMessagetore()
	contractStore := db.NewContractStore()
	changeReqStore := db.NewChangeRequestStore()
	fileStore := db.NewFileStore()

	// Default admin user when the database has no users
	auth.CreateDefaultAdminUser(userStore, logger)

	sseSender := sse.NewSseSender()
	eventSender := event.NewEventSender(sseSender)

	// Services
	authService := service.NewAuthService(logger, userStore, sessionStore)
	projectService := service.NewProjectService(
		logger,
		projectStore,
		invoiceStore,
		contractStore,
		changeReqStore,
		fileStore,
	)
	invoiceService := service.NewInvoiceService(logger, invoiceStore)
	userService := service.NewUserService(logger, userStore, emailSender)
	messageService := service.NewMessageService(logger, eventSender, messageStore)
	contractService := service.NewContractService(logger, eventSender, contractStore, emailSender)
	changeReqService := service.NewChangeRequestService(logger, eventSender, changeReqStore)
	fileService := service.NewFileService(logger, eventSender, fileStore, fileStorage)

	// Handler mux init
	authMux := v1.NewAuthHandler(authService).GetMux(logger)
	projectMux := v1.NewProjectHandler(projectService).GetMux(logger, sessionStore, userStore)
	invoiceMux := v1.NewInvoiceHandler(invoiceService).GetMux(logger, sessionStore, userStore)
	userMux := v1.NewUserHandler(userService).GetMux(logger, sessionStore, userStore)
	messageMux := v1.NewMessageHandler(messageService).GetMux(logger, sessionStore, userStore)
	contractMux := v1.NewContractHandler(contractService).GetMux(logger, sessionStore, userStore)
	changeReqMux := v1.NewChangeRequestHandler(changeReqService).GetMux(logger, sessionStore, userStore)
	fileMux := v1.NewFileHandler(fileService).GetMux(logger, sessionStore, userStore)
	sseMux := v1.NewSseHandler(sseSender).GetMux(logger, sessionStore, userStore)

	// Server routes
	mainMux := http.NewServeMux()
	mainMux.Handle("/api/v1/auth/", http.StripPrefix("/api/v1/auth", authMux))
	mainMux.Handle("/api/v1/projects/", http.StripPrefix("/api/v1/projects", projectMux))
	mainMux.Handle("/api/v1/invoices/", http.StripPrefix("/api/v1/invoices", invoiceMux))
	mainMux.Handle("/api/v1/users/", http.StripPrefix("/api/v1/users", userMux))
	mainMux.Handle("/api/v1/messages/", http.StripPrefix("/api/v1/messages", messageMux))
	mainMux.Handle("/api/v1/contracts/", http.StripPrefix("/api/v1/contracts", contractMux))
	mainMux.Handle("/api/v1/change-requests/", http.StripPrefix("/api/v1/change-requests", changeReqMux))
	mainMux.Handle("/api/v1/files/", http.StripPrefix("/api/v1/files", fileMux))
	mainMux.Handle("/api/v1/events/", http.StripPrefix("/api/v1/events", sseMux))

	// Start server
	listenOn := os.Getenv("LISTEN_ON")
	if listenOn == "" {
		listenOn = ":8080"
	}

	logger.Info("server listening...", "port", listenOn)
	http.ListenAndServe(listenOn, mainMux)
}
