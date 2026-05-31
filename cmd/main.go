package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"mizu/internal/application/eventbus"
	"mizu/internal/application/filestore"
	"mizu/internal/application/lifecycle"
	"mizu/internal/application/logger"
	"mizu/internal/application/mailer"
	"mizu/internal/application/session"
	"mizu/internal/application/uow"
	"mizu/internal/infrastructure/auth"
	"mizu/internal/infrastructure/db/sqlite"
	"mizu/internal/infrastructure/email"
	"mizu/internal/infrastructure/eventbus/externalbus"
	"mizu/internal/infrastructure/eventbus/internalbus"
	filestoreInfra "mizu/internal/infrastructure/filestore"
	loggerInfra "mizu/internal/infrastructure/logger"
	sessionInfra "mizu/internal/infrastructure/session"
	"mizu/internal/infrastructure/uuid"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"mizu/internal/domain/billing"
	"mizu/internal/domain/contract"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/message"
	"mizu/internal/domain/project"
	"mizu/internal/domain/task"
	"mizu/internal/domain/verification"

	authzApp "mizu/internal/application/authz"
	billingApp "mizu/internal/application/billing"
	contractApp "mizu/internal/application/contract"
	iamApp "mizu/internal/application/iam"
	messageApp "mizu/internal/application/message"
	projectApp "mizu/internal/application/project"
	taskApp "mizu/internal/application/task"

	messageIntgEvt "mizu/internal/application/message/integration"

	iamEvtHdl "mizu/internal/application/iam/eventhandler"
	messageEvtHdl "mizu/internal/application/message/eventhandler"

	"mizu/internal/presentation/http/cookie"
	billingHandler "mizu/internal/presentation/http/rest/billing"
	contractHandler "mizu/internal/presentation/http/rest/contract"
	iamHandler "mizu/internal/presentation/http/rest/iam"
	messageHandler "mizu/internal/presentation/http/rest/message"
	"mizu/internal/presentation/http/rest/middleware"
	projectHandler "mizu/internal/presentation/http/rest/project"
	"mizu/internal/presentation/http/rest/response"
	taskHandler "mizu/internal/presentation/http/rest/task"
	"mizu/internal/presentation/http/sse"

	ui "mizu/web"

	"github.com/redis/go-redis/v9"
)

func main() {

	// Utils
	// ----------------------------------------------------------------------------
	log := loggerInfra.NewSlogLogger()
	pwHasher := auth.NewDefaultPasswordHasher()
	idgen := uuid.NewUUIDGenerator()
	authCookie := cookie.NewAuthCookie()

	// Environment
	// ----------------------------------------------------------------------------
	publicURL := env("PUBLIC_URL")
	if publicURL == "" {
		log.Warn("public url not specified, required for email links")
	}

	serverAddr := env("SERVER_ADDR")
	if serverAddr == "" {
		log.Info("server address not set, using :8080")
		serverAddr = ":8080"
	}
	emailDriver := env("EMAIL_DRIVER")
	if emailDriver == "" {
		log.Info("email driver not specified, using no-op driver")
		emailDriver = "noop"
	}

	if emailDriver != "noop" && publicURL == "" {
		log.Fatal("public url is required for email drivers other than no-op")
	}

	emailTemplateDir := env("EMAIL_TEMPLATE_DIR")

	fileStorageDriver := env("FILE_STORAGE_DRIVER")
	if fileStorageDriver == "" {
		log.Info("file storage driver not specified, using disk")
		fileStorageDriver = "disk"
	}

	fileConnString := env("FILE_CONNECTION_STRING")
	if fileStorageDriver != "disk" && fileConnString == "" {
		log.Fatal("file storage connection string is required")
	}

	messageQueueDriver := env("MESSAGE_QUEUE_DRIVER")
	if messageQueueDriver == "" {
		log.Info("message queue driver not specified, using in-memory queue")
		messageQueueDriver = "memory"
	}

	messageQueueConnString := env("MESSAGE_QUEUE_CONNECTION_STRING")
	if messageQueueDriver != "memory" && messageQueueConnString == "" {
		log.Fatal("message queue connection string is required")
	}

	sessionDriver := env("SESSION_DRIVER")
	if sessionDriver == "" {
		log.Info("session driver not specified, using in-memory sessions")
		sessionDriver = "memory"
	}

	sessionConnString := env("SESSION_CONNECTION_STRING")
	if sessionDriver != "memory" && sessionConnString == "" {
		log.Fatal("message queue connection string is required")
	}

	dbDriver := env("DB_DRIVER")
	if dbDriver == "" {
		log.Info("database driver not specified, using sqlite")
		dbDriver = "sqlite"
	}
	dbConnString := env("DB_CONNECTION_STRING")
	if dbDriver == "sqlite" && dbConnString == "" {
		log.Info("sqlite database name not specified, using mizu.db")
		dbConnString = "mizu.db"
	}
	if dbDriver != "sqlite" && dbConnString == "" {
		log.Fatal("database connection string is required")
	}

	eventBufferSizeStr := env("EVENT_BUFFER_SIZE")
	if eventBufferSizeStr == "" {
		log.Info("event buffer size not specified, using 1024 as default")
		eventBufferSizeStr = "1024"
	}
	eventBufferSize, err := strconv.Atoi(eventBufferSizeStr)
	if err != nil {
		log.Fatal("invalid event buffer size")
	}

	maxUploadSizeMBStr := env("MAX_UPLOAD_SIZE_MB")
	if maxUploadSizeMBStr == "" {
		log.Info("max upload size not specified, using 100MB as default")
		maxUploadSizeMBStr = "100"
	}
	maxUploadSizeMB, err := strconv.ParseInt(maxUploadSizeMBStr, 10, 64)
	if err != nil {
		log.Fatal("invalid max upload size", "err", err)
	}

	// SMTP configuration
	// ----------------------------------------------------------------------------
	var smtpConfig email.SMTPConfig
	if emailDriver == "smtp" {
		smtpConfig = email.SMTPConfig{
			Host:     requireEnv(log, "SMTP_HOST"),
			Port:     requireInt(log, "SMTP_PORT"),
			Username: requireEnv(log, "SMTP_USERNAME"),
			Password: requireEnv(log, "SMTP_PASSWORD"),
			From:     requireEnv(log, "SMTP_FROM"),
			BaseURL:  publicURL,
		}
	}

	// FileStore
	// ----------------------------------------------------------------------------
	var fileStore filestore.Store
	switch fileStorageDriver {
	case "disk":
		localStore, err := filestoreInfra.NewLocalStore()
		if err != nil {
			log.Fatal("write error message here", "err", err)
		}
		fileStore = localStore

	default:
		log.Fatal("unsupported file storage driver", "driver", fileStorageDriver)
	}

	// Email
	// ----------------------------------------------------------------------------
	emailTemplateMgr, err := email.NewTemplateManager(emailTemplateDir)
	if err != nil {
		log.Fatal("failed to initialize email template manager", "err", err)
	}

	var mlr mailer.Mailer
	switch emailDriver {
	case "smtp":
		mlr = email.NewSMTPMailer(smtpConfig, emailTemplateMgr, eventBufferSize)

	case "noop":
		mlr = email.NewNoOpMailer(log)

	default:
		log.Fatal("unsupported email driver", "driver", emailDriver)
	}

	// Session
	// ----------------------------------------------------------------------------
	var sessionStore session.Store

	switch sessionDriver {
	case "memory":
		sessionStore = sessionInfra.NewInMemoryStore()

	case "redis":
		opts, err := redis.ParseURL(sessionConnString)
		if err != nil {
			log.Fatal("invalid redis connection string", "err", err)
		}

		client := redis.NewClient(opts)

		if err := client.Ping(context.Background()).Err(); err != nil {
			log.Fatal("redis connection failed", "err", err)
		}

		sessionStore = sessionInfra.NewRedisStore(client, authCookie.Validity())

		log.Info("redis session store connected")

	default:
		log.Fatal("unsupported session driver", "driver", sessionDriver)
	}

	// Repositories
	// ----------------------------------------------------------------------------
	var iamRepo iam.Repository
	var verificationRepo verification.Repository
	var projectRepo project.Repository
	var taskRepo task.Repository
	var billingRepo billing.Repository
	var contractRepo contract.Repository
	var messageRepo message.MessageRepository
	var channelRepo message.ChannelRepository
	var fileRepo message.FileRepository
	var uow uow.UnitOfWork

	switch dbDriver {
	case "sqlite":
		db, err := sqlite.Connect(dbConnString)
		if err != nil {
			log.Fatal("sqlite connection failed", "err", err)
		}

		log.Info("migrating sqlite")
		if err := sqlite.RunMigrations(db); err != nil {
			log.Fatal("failed to migrate sqlite", "err", err)
		}

		defer db.Close()

		iamRepo = sqlite.NewIAMRepository(db)
		verificationRepo = sqlite.NewVerificationRepository(db)
		projectRepo = sqlite.NewProjectRepository(db)
		taskRepo = sqlite.NewTaskRepository(db)
		billingRepo = sqlite.NewBillingRepository(db)
		contractRepo = sqlite.NewContractRepository(db)
		messageRepo = sqlite.NewMessageRepository(db)
		channelRepo = sqlite.NewChannelRepository(db)
		fileRepo = sqlite.NewFileRepository(db)
		uow = sqlite.NewUnitOfWork(db)

	default:
		log.Fatal("unrecognized database driver", "driver", dbDriver)
	}

	// External event bus
	// ----------------------------------------------------------------------------

	var extBus eventbus.ExternalBus
	switch messageQueueDriver {
	case "memory":
		extBus = externalbus.NewInMemoryBus(log, eventBufferSize)

	case "redis":
		opts, err := redis.ParseURL(sessionConnString)
		if err != nil {
			log.Fatal("invalid redis message bus connection string", "err", err)
		}

		client := redis.NewClient(opts)

		if err := client.Ping(context.Background()).Err(); err != nil {
			log.Fatal("redis message bus connection failed", "err", err)
		}

		redisBus := externalbus.NewRedisBus(client, log)

		redisBus.RegisterEvent(messageIntgEvt.EventMessageBroadcast, func() eventbus.Event {
			return messageIntgEvt.MessageBroadcast{}
		})

		extBus = redisBus

	default:
		log.Fatal("unsupported message queue driver", "driver", messageQueueDriver)
	}

	// Internal event bus
	// ----------------------------------------------------------------------------

	systemMsgPub := messageEvtHdl.NewSystemMessagePublisher(channelRepo, messageRepo, extBus, uow, idgen)

	verificationEmailEvtHdl := iamEvtHdl.NewSendVerificationEmail(mlr, iamRepo)
	verifiedEmailEvtHdl := iamEvtHdl.NewSendVerifiedEmail(mlr, iamRepo)
	contractCreatedEvtHdl := messageEvtHdl.NewAddContractCreatedMessage(systemMsgPub)
	contractStatusEvtHdl := messageEvtHdl.NewAddContractStatusChangedMessage(iamRepo, systemMsgPub)
	fileUploadEvtHdl := messageEvtHdl.NewAddFileUploadedMessage(messageRepo, iamRepo, extBus, idgen)
	invoiceConvertedEvtHdl := messageEvtHdl.NewAddInvoiceConvertedMessage(systemMsgPub)
	invoiceCreatedEvtHdl := messageEvtHdl.NewAddInvoiceCreatedMessage(systemMsgPub)
	invoiceStatusEvtHdl := messageEvtHdl.NewAddInvoiceStatusChangedMessage(systemMsgPub)
	createDefaultChannelEvtHdl := messageEvtHdl.NewCreateDefaultChannel(mlr, channelRepo, idgen)
	createProjectChannelEvtHdl := messageEvtHdl.NewCreateProjectChannel(channelRepo, idgen, uow)

	intBus := internalbus.NewInMemoryBus(log, eventBufferSize)

	intBus.Subscribe(verification.EventTypeVerificationCreated, verificationEmailEvtHdl.Handle)
	intBus.Subscribe(contract.EventTypeContractCreated, contractCreatedEvtHdl.Handle)
	intBus.Subscribe(contract.EventTypeContractStatusChanged, contractStatusEvtHdl.Handle)
	intBus.Subscribe(message.EventTypeFileCreated, fileUploadEvtHdl.Handle)
	intBus.Subscribe(billing.EventTypeInvoiceConverted, invoiceConvertedEvtHdl.Handle)
	intBus.Subscribe(billing.EventTypeInvoiceCreated, invoiceCreatedEvtHdl.Handle)
	intBus.Subscribe(billing.EventTypeInvoiceStatusChanged, invoiceStatusEvtHdl.Handle)
	intBus.Subscribe(iam.EventTypeUserCreated, createDefaultChannelEvtHdl.Handle)
	intBus.Subscribe(iam.EventTypeUserVerified, verifiedEmailEvtHdl.Handle)
	intBus.Subscribe(project.EventTypeProjectCreated, createProjectChannelEvtHdl.Handle)

	// Domain services
	// ----------------------------------------------------------------------------
	iamService := iam.NewService(pwHasher, iamRepo)
	projectService := project.NewService(iamRepo, projectRepo)
	taskService := task.NewService(iamRepo, projectRepo)
	contractService := contract.NewService()
	messageService := message.NewService()

	// Services
	// ----------------------------------------------------------------------------
	authzAppService := authzApp.NewService(iamRepo)
	iamAppService := iamApp.NewService(
		iamRepo,
		verificationRepo,
		projectRepo,
		uow,
		iamService,
		authzAppService,
		projectService,
		intBus,
		idgen,
		sessionStore,
		log,
		mlr,
		pwHasher,
	)
	projectAppService := projectApp.NewService(
		iamRepo,
		projectRepo,
		taskRepo,
		billingRepo,
		contractRepo,
		fileRepo,
		uow,
		authzAppService,
		projectService,
		idgen,
		log,
	)
	taskAppService := taskApp.NewService(
		iamRepo,
		taskRepo,
		projectRepo,
		taskService,
		authzAppService,
		intBus,
		idgen,
		log,
	)
	billingAppService := billingApp.NewService(
		billingRepo,
		projectRepo,
		authzAppService,
		intBus,
		idgen,
		log,
	)
	contractAppService := contractApp.NewService(
		iamRepo,
		contractRepo,
		projectRepo,
		contractService,
		authzAppService,
		intBus,
		idgen,
		log,
	)
	messageAppService := messageApp.NewService(
		iamRepo,
		channelRepo,
		messageRepo,
		fileRepo,
		projectRepo,
		idgen,
		intBus,
		extBus,
		authzAppService,
		messageService,
		fileStore,
		log,
	)

	sseSender := sse.NewSender()

	// Handlers
	// ----------------------------------------------------------------------------
	iamHdl := iamHandler.NewIAMHandler(iamAppService, authCookie, log)
	projectHdl := projectHandler.NewProjectHandler(projectAppService, log)
	contractHdl := contractHandler.NewContractHandler(contractAppService, log)
	billingHdl := billingHandler.NewBillingHandler(billingAppService, log)
	messageHdl := messageHandler.NewMessageHandler(messageAppService, log, maxUploadSizeMB)
	taskHdl := taskHandler.NewTaskHandler(taskAppService, log)
	sseHdl := sse.NewHandler(sseSender)

	authMiddleware := middleware.Authenticated(sessionStore, authCookie)

	iamMux := iamHdl.NewMux(authMiddleware)
	projectMux := projectHdl.NewMux(authMiddleware)
	contractMux := contractHdl.NewMux(authMiddleware)
	billingMux := billingHdl.NewMux(authMiddleware)
	messageMux := messageHdl.NewMux(authMiddleware)
	taskMux := taskHdl.NewMux(authMiddleware)
	sseMux := sseHdl.NewMux(authMiddleware)

	mainMux := http.NewServeMux()

	mainMux.Handle("/api/v1/users", http.StripPrefix("/api/v1", iamMux))
	mainMux.Handle("/api/v1/users/", http.StripPrefix("/api/v1", iamMux))

	mainMux.Handle("/api/v1/projects", http.StripPrefix("/api/v1", projectMux))
	mainMux.Handle("/api/v1/projects/", http.StripPrefix("/api/v1", projectMux))

	mainMux.Handle("/api/v1/contracts", http.StripPrefix("/api/v1", contractMux))
	mainMux.Handle("/api/v1/contracts/", http.StripPrefix("/api/v1", contractMux))

	mainMux.Handle("/api/v1/invoices", http.StripPrefix("/api/v1", billingMux))
	mainMux.Handle("/api/v1/invoices/", http.StripPrefix("/api/v1", billingMux))

	mainMux.Handle("/api/v1/messages/", http.StripPrefix("/api/v1", messageMux))

	mainMux.Handle("/api/v1/tasks/", http.StripPrefix("/api/v1", taskMux))

	mainMux.Handle("/api/v1/events", http.StripPrefix("/api/v1", sseMux))
	mainMux.Handle("/api/v1/events/", http.StripPrefix("/api/v1", sseMux))

	// Static file serving
	// ----------------------------------------------------------------------------
	staticFS, err := fs.Sub(ui.WebUiFS, "ui/build")
	if err != nil {
		log.Fatal("failed to create sub filesystem for static UI", "err", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("failed to get working directory", "err", err)
	}

	uploadsDir := filepath.Join(cwd, "uploads")
	if err := os.MkdirAll(uploadsDir, os.ModePerm); err != nil {
		log.Fatal("failed to create uploads directory", "path", uploadsDir, "err", err)
	}

	uploadsFS := os.DirFS(uploadsDir)

	mainMux.Handle("GET /_app/", http.FileServerFS(staticFS))
	mainMux.Handle("GET /assets/", http.FileServerFS(staticFS))
	mainMux.Handle("GET /uploads/", http.StripPrefix("/uploads/", http.FileServerFS(uploadsFS)))
	mainMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			response.WriteError(w, http.StatusNotFound, "not found")
			return
		}

		if r.Method != http.MethodGet {
			response.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		http.ServeFileFS(w, r, staticFS, "index.html")
	})

	// Server setup
	// ----------------------------------------------------------------------------
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      mainMux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())

	startIfAble(ctx, mlr)
	startIfAble(ctx, fileStore)
	startIfAble(ctx, sessionStore)
	startIfAble(ctx, intBus)
	startIfAble(ctx, extBus)

	// Default administrator
	// ----------------------------------------------------------------------------
	if err := iamAppService.EnsureDefaultAdminExists(ctx); err != nil {
		log.Fatal("failed to create default administrator", "err", err)
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Info("server started", "addr", serverAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- fmt.Errorf("server error: %w", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Info("shutdown signal received", "signal", sig)
	case err := <-serverErr:
		log.Error("server encountered a fatal error", "err", err)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("server shutdown did not complete cleanly", "err", err)
	}

	cancel()
	stopIfAble(mlr)
	stopIfAble(fileStore)
	stopIfAble(sessionStore)
	stopIfAble(intBus)
	stopIfAble(extBus)

	log.Info("shutdown complete")
}

// env reads an environment variable, trims whitespace, and lowercases the result.
func env(key string) string {
	return strings.ToLower(strings.TrimSpace(os.Getenv(key)))
}

// requireEnv reads a required environment variable and fatally logs if it is empty.
func requireEnv(log logger.Logger, key string) string {
	val := env(key)
	if val == "" {
		log.Fatal("required environment variable is not set", "key", key)
	}
	return val
}

// requireInt reads a required integer environment variable
// and fatally logs if it is missing or invalid.
func requireInt(log logger.Logger, key string) int {
	val := requireEnv(log, key)
	n, err := strconv.Atoi(val)
	if err != nil {
		log.Fatal("environment variable must be a valid integer", "key", key, "value", val)
	}
	return n
}

// startIfAble starts v if it implements lifecycle.Starter.
// Components that do not require a background goroutine are silently skipped.
func startIfAble(ctx context.Context, v any) {
	if s, ok := v.(lifecycle.Starter); ok {
		s.Start(ctx)
	}
}

// stopIfAble stops v if it implements lifecycle.Stopper.
// Components that do not require graceful shutdown are silently skipped.
func stopIfAble(v any) {
	if s, ok := v.(lifecycle.Stopper); ok {
		s.Stop()
	}
}
