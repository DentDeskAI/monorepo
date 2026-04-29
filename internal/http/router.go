package httpx

import (
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
	AuthSvc    *auth.Service
	Log        zerolog.Logger
	Origin     string
	AuthH      *handlers.AuthHandler
	AdminH     *handlers.AdminHandler
	CRMH       *handlers.CRMHandler
	ResourceH  *handlers.ResourceHandler
	ScheduleH  *handlers.SchedulingHandler
	DashboardH *handlers.DashboardHandler
	WhatsApp   *handlers.WhatsAppHandler
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

	app.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	app.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// ==== Public ====
	app.POST("/api/register", r.AdminH.Register)
	app.POST("/api/auth/login", r.AuthH.Login)
	app.GET("/webhook/whatsapp", r.WhatsApp.Verify)
	app.POST("/webhook/whatsapp", r.WhatsApp.Receive)

	// ==== Protected ====
	authmw := middleware.AuthRequired(r.AuthSvc)
	api := app.Group("/api", authmw)
	{
		// Auth
		api.GET("/auth/me", r.AuthH.Me)
		api.POST("/auth/change-password", r.AdminH.ChangePassword)

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

		// Patients (read = scheduler/MacDent, write = local repo)
		api.GET("/patients", r.ScheduleH.GetPatients)
		api.POST("/patients", r.ResourceH.CreatePatient)
		api.GET("/patients/:id", r.ScheduleH.GetPatient)
		api.PUT("/patients/:id", r.ResourceH.UpdatePatient)
		api.GET("/patients/:id/appointments", r.CRMH.PatientAppointments)

		// Doctors (read = scheduler/MacDent, write = local repo)
		api.GET("/doctors", r.ScheduleH.GetDoctors)
		api.POST("/doctors", r.ResourceH.CreateDoctor)
		api.GET("/doctors/:id", r.ScheduleH.GetDoctor)
		api.PUT("/doctors/:id", r.ResourceH.UpdateDoctor)
		api.DELETE("/doctors/:id", r.ResourceH.DeactivateDoctor)
		api.POST("/doctors/sync", r.ScheduleH.SyncDoctors)

		// Clinic — local clinic record (settings, name set at registration)
		api.GET("/clinic", r.AdminH.GetClinic)
		api.PUT("/clinic", r.AdminH.UpdateClinic)

		// Chairs
		api.GET("/chairs", r.ResourceH.ListChairs)
		api.POST("/chairs", r.ResourceH.CreateChair)
		api.PUT("/chairs/:id", r.ResourceH.UpdateChair)
		api.DELETE("/chairs/:id", r.ResourceH.DeactivateChair)

		// Schedule — MacDent read/write (integer IDs, live data)
		api.GET("/schedule/doctors", r.ScheduleH.ListAppointments)
		api.GET("/schedule/appointments/:id", r.ScheduleH.GetScheduleAppointment)
		api.PUT("/schedule/appointments/:id", r.ScheduleH.UpdateScheduleAppointment)
		api.DELETE("/schedule/appointments/:id", r.ScheduleH.DeleteScheduleAppointment)
		api.GET("/schedule/patients/:id", r.ScheduleH.GetSchedulePatient)
		api.POST("/schedule/patients", r.ScheduleH.CreateSchedulePatient)
		api.POST("/schedule/appointments", r.ScheduleH.CreateScheduleAppointment)
		api.PUT("/schedule/appointments/:id/status", r.ScheduleH.SetScheduleAppointmentStatus)
		api.POST("/schedule/appointment-requests", r.ScheduleH.SendAppointmentRequest)

		// Appointments — local DB (UUID-keyed, operator-created)
		api.GET("/appointments/", r.ScheduleH.ListAppointments)
		api.POST("/appointments", r.ScheduleH.CreateAppointment)
		api.GET("/appointments/:id", r.ScheduleH.GetAppointment)
		api.PUT("/appointments/:id/status", r.ScheduleH.UpdateAppointmentStatus)
		api.GET("/calendar", r.CRMH.Calendar)

		// Dashboard analytics
		api.GET("/dashboard/today", r.DashboardH.Today)
		api.GET("/dashboard/stats", r.DashboardH.Stats)
		api.GET("/dashboard/revenue", r.DashboardH.Revenue)

		// History — past + current week appointments table
		api.GET("/history", r.ScheduleH.GetHistory)

		// Realtime
		api.GET("/events", r.CRMH.SSE)
	}

	return app
}
