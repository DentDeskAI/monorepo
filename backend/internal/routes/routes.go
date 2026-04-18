// Package routes wires together all HTTP routes for the DentDesk API.
package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/dentdesk/backend/internal/handler"
	"github.com/dentdesk/backend/internal/middleware"
)

// Setup registers all routes on the provided Gin engine.
func Setup(
	r *gin.Engine,
	jwtSecret string,
	authH      *handler.AuthHandler,
	patientH   *handler.PatientHandler,
	apptH      *handler.AppointmentHandler,
	waH        *handler.WhatsAppHandler,
) {
	// ─── Health check ─────────────────────────────────────────────────────────
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// ─── WhatsApp webhook (public — no JWT) ───────────────────────────────────
	webhook := r.Group("/webhook")
	{
		webhook.GET("/whatsapp", waH.Verify)   // Meta verification challenge
		webhook.POST("/whatsapp", waH.Receive) // Inbound events
	}

	// ─── Auth (public) ────────────────────────────────────────────────────────
	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/register", authH.Register)
		auth.POST("/login", authH.Login)
	}

	// ─── Protected routes (require valid JWT + tenant resolution) ─────────────
	api := r.Group("/api/v1")
	api.Use(middleware.TenantMiddleware(jwtSecret))
	{
		// Current user
		api.GET("/auth/me", authH.Me)

		// Patients
		patients := api.Group("/patients")
		{
			patients.GET("", patientH.List)
			patients.POST("", patientH.Create)
			patients.GET("/:id", patientH.Get)
			patients.PUT("/:id", patientH.Update)
			patients.DELETE("/:id", patientH.Delete)
		}

		// Appointments
		appointments := api.Group("/appointments")
		{
			appointments.GET("", apptH.List)
			appointments.POST("", apptH.Create)
			appointments.GET("/:id", apptH.Get)
			appointments.PUT("/:id", apptH.Update)
			appointments.DELETE("/:id", apptH.Delete)
		}

		// Whatsapp — send manual message (staff-initiated)
		api.POST("/messages/send", waH.SendManual)
	}
}
