package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/dentdesk/dentdesk/internal/auth"
)

const (
	CtxClaims = "auth_claims"
)

func AuthRequired(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := ""
		if h := c.GetHeader("Authorization"); strings.HasPrefix(h, "Bearer ") {
			token = strings.TrimPrefix(h, "Bearer ")
		} else if q := c.Query("token"); q != "" {
			// для SSE: браузерный EventSource не умеет кастомные хедеры
			token = q
		}
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		claims, err := svc.Parse(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set(CtxClaims, claims)
		c.Next()
	}
}

func Logging(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Dur("took", time.Since(start)).
			Str("ip", c.ClientIP()).
			Msg("http")
	}
}

func Recover(log zerolog.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.Error().Interface("panic", recovered).Msg("panic recovered")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal"})
	})
}

// ClaimsFrom — утилита для handlers.
func ClaimsFrom(c *gin.Context) *auth.Claims {
	v, ok := c.Get(CtxClaims)
	if !ok {
		return nil
	}
	cl, _ := v.(*auth.Claims)
	return cl
}
