package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Nutan-Kum12/RateLimiterX/internal/auth"
	"github.com/Nutan-Kum12/RateLimiterX/internal/dto"
)

// AuthMiddleware creates a Gin middleware that validates JWT access tokens.
func AuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewErrorResponse(
				"authentication required",
				"missing Authorization header",
			))
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewErrorResponse(
				"authentication required",
				"invalid Authorization header format, expected 'Bearer <token>'",
			))
			return
		}

		tokenString := parts[1]

		// Parse and validate the access token
		claims, err := jwtManager.ParseAccessToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewErrorResponse(
				"authentication failed",
				err.Error(),
			))
			return
		}

		// Set user info in context for downstream handlers and middleware
		c.Set("userID", claims.UserID)
		// c.Set("email", claims.Email)
		c.Set("tier", claims.Tier)

		c.Next()
	}
}
