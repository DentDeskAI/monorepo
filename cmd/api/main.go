// @title           DentDesk API
// @version         1.0
// @description     Dental CRM — WhatsApp bot + operator workspace.
// @host            localhost:8082
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
// @description     Format: "Bearer <token>"

package main

import (
	_ "github.com/dentdesk/dentdesk/docs"

	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dentdesk/dentdesk/internal/appointments"
	"github.com/dentdesk/dentdesk/internal/auth"
	"github.com/dentdesk/dentdesk/internal/chairs"
	"github.com/dentdesk/dentdesk/internal/clinics"
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
	"github.com/dentdesk/dentdesk/internal/services"
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
	clinicsRepo := clinics.NewRepo(database)
	chairsRepo := chairs.NewRepo(database)
	patientsRepo := patients.NewRepo(database)
	convRepo := conversations.NewRepo(database)
	apptRepo := appointments.NewRepo(database)
	doctorsRepo := doctors.NewRepo(database)

	// --- Scheduler ---
	var sched scheduler.Scheduler
	switch cfg.SchedulerDefault {
	case "mock":
		sched = scheduler.NewMockAdapter(database)
		log.Info().Msg("scheduler: mock")
	case "macdent":
		sched = scheduler.NewMacDentAdapter(database)
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
	adminSvc := services.NewAdminService(authSvc, clinicsRepo)
	resourceSvc := services.NewResourceService(doctorsRepo, chairsRepo, patientsRepo)
	schedulingSvc := services.NewSchedulingService(apptRepo, convRepo, sched)
	crmSvc := services.NewCRMService(database, patientsRepo, convRepo, apptRepo, doctorsRepo, hub, waClient)

	// --- Handlers ---
	authH := &handlers.AuthHandler{Svc: authSvc}
	adminH := &handlers.AdminHandler{Svc: adminSvc}
	crmH := &handlers.CRMHandler{Svc: crmSvc}
	resourceH := &handlers.ResourceHandler{Svc: resourceSvc}
	scheduleH := &handlers.SchedulingHandler{Svc: schedulingSvc}
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
		AuthSvc:   authSvc,
		Log:       log,
		Origin:    cfg.CRMOrigin,
		AuthH:     authH,
		AdminH:    adminH,
		CRMH:      crmH,
		ResourceH: resourceH,
		ScheduleH: scheduleH,
		WhatsApp:  waH,
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
