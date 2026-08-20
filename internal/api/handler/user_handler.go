package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Nutan-Kum12/RateLimiterX.git/internal/dto"
	"github.com/Nutan-Kum12/RateLimiterX.git/internal/service"
)

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetProfile handles GET /api/v1/users/me
// Returns the authenticated user's profile.
func (h *UserHandler) GetProfile(c *gin.Context) {
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(
			"authentication required",
			"user ID not found in context",
		))
		return
	}

	profile, err := h.userService.GetProfile(c.Request.Context(), c.GetString("userID"))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewErrorResponse(
			"user not found",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("profile retrieved", profile))
}

// UpdateTier handles PATCH /api/v1/users/tier
// Updates a user's tier (admin operation).
// func (h *UserHandler) UpdateTier(c *gin.Context) {
// 	var req dto.UpdateTierRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
// 			"invalid request body",
// 			err.Error(),
// 		))
// 		return
// 	}

// 	profile, err := h.userService.UpdateTier(c.Request.Context(), req.UserID, req.Tier)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
// 			"tier update failed",
// 			err.Error(),
// 		))
// 		return
// 	}

// 	c.JSON(http.StatusOK, dto.NewSuccessResponse("tier updated successfully", profile))
// }
