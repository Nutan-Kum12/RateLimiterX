package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Nutan-Kum12/RateLimiterX/internal/dto"
	"github.com/Nutan-Kum12/RateLimiterX/internal/service"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

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
		c.JSON(http.StatusConflict, dto.NewErrorResponse(
			"registration failed",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusCreated, dto.NewSuccessResponse("user registered successfully", resp))
}

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
