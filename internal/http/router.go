package httpx

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/dentdesk/dentdesk/internal/auth"
	"github.com/dentdesk/dentdesk/internal/http/handlers"
	"github.com/dentdesk/dentdesk/internal/http/middleware"
)

type Router struct {
	AuthSvc   *auth.Service
	Log       zerolog.Logger
	Origin    string
	AuthH     *handlers.AuthHandler
	AdminH    *handlers.AdminHandler
	CRMH      *handlers.CRMHandler
	ResourceH *handlers.ResourceHandler
	ScheduleH *handlers.SchedulingHandler
	WhatsApp  *handlers.WhatsAppHandler
	MacDent   *handlers.MacDentHandler
}

func (r *Router) Build() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	app := gin.New()

	app.Use(middleware.Recover(r.Log))
	app.Use(middleware.Logging(r.Log))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     splitOrigins(r.Origin),
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	app.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	app.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// ==== Public ====
	app.POST("/api/register", r.AdminH.Register)
	app.POST("/api/auth/login", r.AuthH.Login)
	app.GET("/webhook/whatsapp", r.WhatsApp.Verify)
	app.POST("/webhook/whatsapp", r.WhatsApp.Receive)
	app.POST("/webhook/macdent/:clinicID", r.MacDent.Receive)

	// ==== Protected ====
	authmw := middleware.AuthRequired(r.AuthSvc)
	api := app.Group("/api", authmw)
	{
		// Auth
		api.GET("/auth/me", r.AuthH.Me)
		api.POST("/auth/change-password", r.AdminH.ChangePassword)

		// Clinic
		api.GET("/clinic", r.AdminH.GetClinic)
		api.PUT("/clinic", r.AdminH.UpdateClinic)
		api.GET("/clinic/macdent/webhook-url", r.MacDent.GetWebhookURL)

		// Users
		api.GET("/users", r.AdminH.ListUsers)
		api.POST("/users", r.AdminH.CreateUser)
		api.GET("/users/:id", r.AdminH.GetUser)
		api.PUT("/users/:id", r.AdminH.UpdateUser)
		api.DELETE("/users/:id", r.AdminH.DeleteUser)

		// Dashboard
		api.GET("/stats", r.CRMH.Stats)

		// Chats
		api.GET("/chats", r.CRMH.ListChats)
		api.GET("/chats/:id", r.ScheduleH.GetConversation)
		api.GET("/chats/:id/messages", r.CRMH.ListMessages)
		api.POST("/chats/:id/send", r.CRMH.OperatorSend)
		api.POST("/chats/:id/release", r.CRMH.ReleaseHandoff)
		api.POST("/chats/:id/close", r.ScheduleH.CloseConversation)

		// Patients
		api.GET("/patients", r.ScheduleH.GetPatients)
		api.POST("/patients", r.ResourceH.CreatePatient)
		api.GET("/patients/:id", r.ScheduleH.GetPatient)
		api.PUT("/patients/:id", r.ResourceH.UpdatePatient)
		api.GET("/patients/:id/appointments", r.CRMH.PatientAppointments)

		// Doctors
		api.GET("/doctors", r.ScheduleH.GetDoctors)
		api.POST("/doctors", r.ResourceH.CreateDoctor)
		api.GET("/doctors/:id", r.ScheduleH.GetDoctor)
		api.PUT("/doctors/:id", r.ResourceH.UpdateDoctor)
		api.DELETE("/doctors/:id", r.ResourceH.DeactivateDoctor)

		// Chairs
		api.GET("/chairs", r.ResourceH.ListChairs)
		api.POST("/chairs", r.ResourceH.CreateChair)
		api.PUT("/chairs/:id", r.ResourceH.UpdateChair)
		api.DELETE("/chairs/:id", r.ResourceH.DeactivateChair)

		// Scheduling
		api.GET("/schedule/doctors", r.ScheduleH.GetDoctors)
		api.POST("/schedule/doctors/sync", r.ScheduleH.SyncDoctors)
		api.GET("/schedule/patients", r.ScheduleH.GetPatients)
		api.GET("/schedule/patients/:id", r.ScheduleH.GetPatient)
		api.GET("/slots", r.ScheduleH.GetSlots)
		api.POST("/appointments", r.ScheduleH.CreateAppointment)
		api.GET("/appointments/:id", r.ScheduleH.GetAppointment)
		api.PUT("/appointments/:id/status", r.ScheduleH.UpdateAppointmentStatus)
		api.GET("/calendar", r.CRMH.Calendar)

		// Realtime
		api.GET("/events", r.CRMH.SSE)
	}

	return app
}

func splitOrigins(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if origin := strings.TrimSpace(part); origin != "" {
			out = append(out, origin)
		}
	}
	return out
}
