package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Nutan-Kum12/RateLimiterX.git/internal/dto"
	"github.com/Nutan-Kum12/RateLimiterX.git/internal/service"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register handles POST /api/v1/auth/register
// Creates a new user with the "free" tier and returns JWT tokens.
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			"invalid request body",
			err.Error(),
		))
		return
	}

	resp, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		// Check if it's a duplicate email error
		c.JSON(http.StatusConflict, dto.NewErrorResponse(
			"registration failed",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusCreated, dto.NewSuccessResponse("user registered successfully", resp))
}

// Login handles POST /api/v1/auth/login
// Validates credentials and returns JWT tokens.
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			"invalid request body",
			err.Error(),
		))
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(
			"login failed",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("login successful", resp))
}
