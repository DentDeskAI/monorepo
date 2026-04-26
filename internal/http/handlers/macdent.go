package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	neturl "net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"

	"github.com/dentdesk/dentdesk/internal/http/middleware"
	"github.com/dentdesk/dentdesk/internal/macdent"
)

type MacDentHandler struct {
	DB            *sqlx.DB
	Redis         *redis.Client
	Log           zerolog.Logger
	PublicBaseURL string
	WebhookToken  string
}

func (h *MacDentHandler) GetWebhookURL(c *gin.Context) {
	if !isOwnerOrAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	claims := middleware.ClaimsFrom(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	baseURL := h.publicBaseURL(c)
	if baseURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to build public url"})
		return
	}

	webhookURL := strings.TrimRight(baseURL, "/") + "/webhook/macdent/" + claims.ClinicID.String()
	if token := strings.TrimSpace(h.WebhookToken); token != "" {
		webhookURL += "?token=" + neturl.QueryEscape(token)
	}

	c.JSON(http.StatusOK, gin.H{
		"method":           http.MethodPost,
		"url":              webhookURL,
		"clinic_id":        claims.ClinicID,
		"token_configured": strings.TrimSpace(h.WebhookToken) != "",
	})
}

func (h *MacDentHandler) Receive(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("clinicID"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "clinic not found"})
		return
	}

	if !h.authorized(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid webhook token"})
		return
	}

	exists, err := h.clinicExists(c, clinicID)
	if err != nil {
		h.Log.Error().Err(err).Str("clinic_id", clinicID.String()).Msg("macdent webhook clinic lookup failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "clinic not found"})
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.Log.Warn().Err(err).Str("clinic_id", clinicID.String()).Msg("macdent webhook read failed")
		c.JSON(http.StatusOK, gin.H{"status": "accepted"})
		return
	}

	dedupHash := sha256.Sum256(body)
	dedupKey := "macdent:webhook:" + clinicID.String() + ":" + hex.EncodeToString(dedupHash[:])
	if h.Redis != nil {
		ok, err := h.Redis.SetNX(c.Request.Context(), dedupKey, "1", 15*time.Minute).Result()
		if err != nil {
			h.Log.Warn().Err(err).Str("clinic_id", clinicID.String()).Msg("macdent webhook dedup unavailable")
		} else if !ok {
			c.JSON(http.StatusOK, gin.H{"status": "duplicate"})
			return
		}
	}

	summary := macdent.SummarizeWebhook(body)
	event := h.Log.Info().
		Str("clinic_id", clinicID.String()).
		Int("bytes", len(body)).
		Str("dedup_key", dedupKey)
	if summary.Event != "" {
		event = event.Str("event", summary.Event)
	}
	if summary.Entity != "" {
		event = event.Str("entity", summary.Entity)
	}
	if summary.ObjectID != "" {
		event = event.Str("object_id", summary.ObjectID)
	}
	event.Msg("macdent webhook accepted")

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *MacDentHandler) authorized(c *gin.Context) bool {
	expected := strings.TrimSpace(h.WebhookToken)
	if expected == "" {
		return true
	}

	provided := strings.TrimSpace(c.Query("token"))
	if provided == "" {
		provided = strings.TrimSpace(c.GetHeader("X-MacDent-Token"))
	}

	return provided == expected
}

func (h *MacDentHandler) clinicExists(c *gin.Context, clinicID uuid.UUID) (bool, error) {
	var exists bool
	if err := h.DB.GetContext(c.Request.Context(), &exists,
		`SELECT EXISTS (SELECT 1 FROM clinics WHERE id = $1)`, clinicID); err != nil {
		return false, err
	}
	return exists, nil
}

func (h *MacDentHandler) publicBaseURL(c *gin.Context) string {
	if configured := strings.TrimSpace(h.PublicBaseURL); configured != "" {
		return strings.TrimRight(configured, "/")
	}

	scheme := strings.TrimSpace(c.GetHeader("X-Forwarded-Proto"))
	if scheme == "" {
		if c.Request.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	host := strings.TrimSpace(c.GetHeader("X-Forwarded-Host"))
	if host == "" {
		host = strings.TrimSpace(c.Request.Host)
	}
	if host == "" {
		return ""
	}

	return scheme + "://" + host
}
