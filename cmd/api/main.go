package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dentdesk/dentdesk/internal/appointments"
	"github.com/dentdesk/dentdesk/internal/auth"
	"github.com/dentdesk/dentdesk/internal/conversations"
	"github.com/dentdesk/dentdesk/internal/doctors"
	httpx "github.com/dentdesk/dentdesk/internal/http"
	"github.com/dentdesk/dentdesk/internal/http/handlers"
	"github.com/dentdesk/dentdesk/internal/llm"
	"github.com/dentdesk/dentdesk/internal/patients"
	"github.com/dentdesk/dentdesk/internal/platform/config"
	"github.com/dentdesk/dentdesk/internal/platform/db"
	"github.com/dentdesk/dentdesk/internal/platform/logger"
	redisx "github.com/dentdesk/dentdesk/internal/platform/redis"
	"github.com/dentdesk/dentdesk/internal/realtime"
	"github.com/dentdesk/dentdesk/internal/scheduler"
	"github.com/dentdesk/dentdesk/internal/whatsapp"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.LogLevel)

	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()

	// --- DB ---
	database, err := db.Connect(rootCtx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("db connect")
	}
	defer database.Close()

	if migDir := firstExisting("./migrations", "/app/migrations"); migDir != "" {
		if err := db.RunMigrations(rootCtx, database, migDir); err != nil {
			log.Fatal().Err(err).Msg("migrations")
		}
		log.Info().Str("dir", migDir).Msg("migrations applied")
	} else {
		log.Warn().Msg("migrations dir not found, skipping")
	}

	// --- Redis ---
	redisClient, err := redisx.Connect(rootCtx, cfg.RedisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("redis connect")
	}
	defer redisClient.Close()

	// --- Repositories ---
	patientsRepo := patients.NewRepo(database)
	convRepo := conversations.NewRepo(database)
	apptRepo := appointments.NewRepo(database)
	doctorsRepo := doctors.NewRepo(database)

	// --- Scheduler: выбираем реализацию на старте (для MVP — по env).
	var sched scheduler.Scheduler
	switch cfg.SchedulerDefault {
	case "mock":
		sched = scheduler.NewMockAdapter(database)
		log.Info().Msg("scheduler: mock")
	case "macdent":
		sched = scheduler.NewMacDentAdapter("", "") // placeholder; в проде — per-clinic
		log.Info().Msg("scheduler: macdent")
	default:
		sched = scheduler.NewLocalAdapter(database)
		log.Info().Msg("scheduler: local")
	}

	// --- LLM ---
	var llmClient llm.Client
	switch cfg.LLMProvider {
	case "anthropic":
		if cfg.AnthropicAPIKey == "" {
			log.Warn().Msg("ANTHROPIC_API_KEY not set — falling back to mock LLM")
			llmClient = llm.NewMock()
		} else {
			llmClient = llm.NewAnthropic(cfg.AnthropicAPIKey, cfg.AnthropicModel)
			log.Info().Str("model", cfg.AnthropicModel).Msg("llm: anthropic")
		}
	case "groq":
		if cfg.GroqAPIKey == "" {
			log.Warn().Msg("GROQ_API_KEY not set — falling back to mock LLM")
			llmClient = llm.NewMock()
		} else {
			llmClient = llm.NewGroq(cfg.GroqAPIKey, cfg.GroqModel)
			log.Info().Str("model", cfg.GroqModel).Msg("llm: groq")
		}
	default:
		llmClient = llm.NewMock()
		log.Info().Msg("llm: mock")
	}

	orchestrator := llm.NewOrchestrator(llmClient, sched, log)

	// --- WhatsApp ---
	waClient := whatsapp.NewClient(cfg.WhatsAppToken, cfg.WhatsAppPhoneNumberID, cfg.WhatsAppAPIVersion)

	// --- Realtime hub ---
	hub := realtime.NewHub()

	// --- Auth ---
	authSvc := auth.NewService(database, cfg.JWTSecret)

	// --- Handlers ---
	authH := &handlers.AuthHandler{Svc: authSvc}
	crmH := &handlers.CRMHandler{
		DB:            database,
		Patients:      patientsRepo,
		Conversations: convRepo,
		Appointments:  apptRepo,
		Doctors:       doctorsRepo,
		Hub:           hub,
		WhatsApp:      waClient,
	}
	waH := &handlers.WhatsAppHandler{
		DB:            database,
		Redis:         redisClient,
		Log:           log,
		VerifyToken:   cfg.WhatsAppVerifyToken,
		WhatsApp:      waClient,
		Patients:      patientsRepo,
		Conversations: convRepo,
		Orchestrator:  orchestrator,
		Scheduler:     sched,
		Hub:           hub,
	}

	router := (&httpx.Router{
		AuthSvc:  authSvc,
		Log:      log,
		Origin:   cfg.CRMOrigin,
		AuthH:    authH,
		CRMH:     crmH,
		WhatsApp: waH,
	}).Build()

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Info().Str("port", cfg.HTTPPort).Msg("http listening")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("listen")
		}
	}()

	// --- Graceful shutdown ---
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Info().Msg("shutting down")

	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Error().Err(err).Msg("shutdown")
	}
}

func firstExisting(paths ...string) string {
	for _, p := range paths {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			return p
		}
	}
	return ""
}
