package httpx

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/dentdesk/dentdesk/internal/auth"
	"github.com/dentdesk/dentdesk/internal/http/handlers"
	"github.com/dentdesk/dentdesk/internal/http/middleware"
)

type Router struct {
	AuthSvc  *auth.Service
	Log      zerolog.Logger
	Origin   string
	AuthH    *handlers.AuthHandler
	CRMH     *handlers.CRMHandler
	WhatsApp *handlers.WhatsAppHandler
}

func (r *Router) Build() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	app := gin.New()

	app.Use(middleware.Recover(r.Log))
	app.Use(middleware.Logging(r.Log))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{r.Origin},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	app.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// ==== Public ====
	app.POST("/api/auth/login", r.AuthH.Login)

	// WhatsApp webhook (без JWT, защищает verify_token + идемпотентность)
	app.GET("/webhook/whatsapp", r.WhatsApp.Verify)
	app.POST("/webhook/whatsapp", r.WhatsApp.Receive)

	// ==== Protected ====
	authmw := middleware.AuthRequired(r.AuthSvc)
	api := app.Group("/api", authmw)
	{
		api.GET("/auth/me", r.AuthH.Me)
		api.GET("/stats", r.CRMH.Stats)

		api.GET("/chats", r.CRMH.ListChats)
		api.GET("/chats/:id/messages", r.CRMH.ListMessages)
		api.POST("/chats/:id/send", r.CRMH.OperatorSend)
		api.POST("/chats/:id/release", r.CRMH.ReleaseHandoff)

		api.GET("/patients", r.CRMH.ListPatients)
		api.GET("/patients/:id/appointments", r.CRMH.PatientAppointments)

		api.GET("/calendar", r.CRMH.Calendar)
		api.GET("/doctors", r.CRMH.ListDoctors)

		api.GET("/events", r.CRMH.SSE)
	}

	return app
}
