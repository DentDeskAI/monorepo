// DentDesk — SaaS platform for dental clinics.
// Entry point: wires config → DB → repositories → services → handlers → routes.

// @title DentDesk API
// @version 1.0
// @description API documentation
// @host localhost:18080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/dentdesk/backend/internal/config"
	"github.com/dentdesk/backend/internal/handler"
	"github.com/dentdesk/backend/internal/llm"
	"github.com/dentdesk/backend/internal/repository"
	"github.com/dentdesk/backend/internal/routes"
	"github.com/dentdesk/backend/internal/service"
	"github.com/dentdesk/backend/internal/whatsapp"
	"github.com/dentdesk/backend/migrations"

	_ "github.com/dentdesk/backend/cmd/docs"
)

func main() {
	// ─── Logger ───────────────────────────────────────────────────────────────
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	zerolog.TimeFieldFormat = time.RFC3339

	// ─── Config ───────────────────────────────────────────────────────────────
	// Load .env for local development; silently ignored if file is absent.
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load configuration")
	}

	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// ─── Database ─────────────────────────────────────────────────────────────
	db, err := connectDB(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	if err := migrations.Run(db); err != nil {
		log.Fatal().Err(err).Msg("database migration failed")
	}

	// ─── Repositories ─────────────────────────────────────────────────────────
	clinicRepo := repository.NewClinicRepository(db)
	userRepo := repository.NewUserRepository(db)
	patientRepo := repository.NewPatientRepository(db)
	apptRepo := repository.NewAppointmentRepository(db)
	msgRepo := repository.NewMessageLogRepository(db)

	// ─── External clients ─────────────────────────────────────────────────────
	waClient := whatsapp.NewClient(
		cfg.WhatsApp.APIURL,
		cfg.WhatsApp.PhoneNumberID,
		cfg.WhatsApp.AccessToken,
	)

	llmClient := llm.NewClient(
		cfg.LLM.BaseURL,
		cfg.LLM.APIKey,
		cfg.LLM.Model,
		cfg.LLM.Provider,
	)

	// ─── Services ─────────────────────────────────────────────────────────────
	authSvc := service.NewAuthService(userRepo, clinicRepo, cfg.JWT)
	patientSvc := service.NewPatientService(patientRepo)
	apptSvc := service.NewAppointmentService(apptRepo)
	waSvc := service.NewWhatsAppService(msgRepo, patientRepo, clinicRepo, waClient, llmClient)

	// ─── Handlers ─────────────────────────────────────────────────────────────
	authH := handler.NewAuthHandler(authSvc)
	patientH := handler.NewPatientHandler(patientSvc)
	apptH := handler.NewAppointmentHandler(apptSvc)
	waH := handler.NewWhatsAppHandler(waSvc, cfg.WhatsApp.VerifyToken)

	// ─── HTTP server ──────────────────────────────────────────────────────────
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestLogger())

	routes.Setup(r, cfg.JWT.Secret, authH, patientH, apptH, waH)

	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ─── Graceful shutdown ────────────────────────────────────────────────────
	go func() {
		log.Info().Str("port", cfg.App.Port).Msg("DentDesk API starting")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("server forced shutdown")
	}
	log.Info().Msg("server stopped")
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func connectDB(cfg *config.Config) (*gorm.DB, error) {
	gormCfg := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	}

	db, err := gorm.Open(postgres.Open(cfg.DB.DSN()), gormCfg)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Info().Str("host", cfg.DB.Host).Str("db", cfg.DB.Name).Msg("database connected")
	return db, nil
}

// requestLogger returns a Gin middleware that logs each request with zerolog.
func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Dur("latency", time.Since(start)).
			Str("ip", c.ClientIP()).
			Msg("request")
	}
}
