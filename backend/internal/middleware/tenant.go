// Package middleware provides Gin middleware for DentDesk.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const (
	// CtxClinicID is the context key used to propagate the resolved tenant ID.
	CtxClinicID = "clinic_id"
	// CtxUserID is the context key for the authenticated user.
	CtxUserID = "user_id"
	// CtxUserRole is the context key for the authenticated user's role.
	CtxUserRole = "user_role"
)

// DentDeskClaims extends standard JWT claims with tenant info.
type DentDeskClaims struct {
	ClinicID uuid.UUID `json:"clinic_id"`
	UserID   uuid.UUID `json:"user_id"`
	Role     string    `json:"role"`
	jwt.RegisteredClaims
}

// TenantMiddleware resolves the current tenant (ClinicID) from the JWT token.
// In MVP we use JWT-embedded clinic_id; subdomain mode can be added later.
func TenantMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := extractBearerToken(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing or malformed authorization token",
			})
			return
		}

		claims, err := parseJWT(token, jwtSecret)
		if err != nil {
			log.Warn().Err(err).Msg("invalid JWT token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		if claims.ClinicID == uuid.Nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "token does not contain a valid clinic_id",
			})
			return
		}

		// Inject tenant context — all repository calls will use this value.
		c.Set(CtxClinicID, claims.ClinicID)
		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxUserRole, claims.Role)

		log.Debug().
			Str("clinic_id", claims.ClinicID.String()).
			Str("user_id", claims.UserID.String()).
			Msg("tenant resolved")

		c.Next()
	}
}

// ClinicIDFromCtx is a helper to retrieve the tenant ID from a Gin context.
// Panics if TenantMiddleware was not applied — this is intentional (misconfiguration).
func ClinicIDFromCtx(c *gin.Context) uuid.UUID {
	v, exists := c.Get(CtxClinicID)
	if !exists {
		panic("clinic_id not found in context: TenantMiddleware must be applied")
	}
	return v.(uuid.UUID)
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func extractBearerToken(c *gin.Context) (string, error) {
	header := c.GetHeader("Authorization")
	if header == "" {
		return "", ErrMissingToken
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", ErrMalformedToken
	}
	return parts[1], nil
}

func parseJWT(tokenStr, secret string) (*DentDeskClaims, error) {
	claims := &DentDeskClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(_ *jwt.Token) (any, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// sentinel errors
var (
	ErrMissingToken  = &tokenError{"authorization header is missing"}
	ErrMalformedToken = &tokenError{"authorization header format must be Bearer {token}"}
)

type tokenError struct{ msg string }

func (e *tokenError) Error() string { return e.msg }
