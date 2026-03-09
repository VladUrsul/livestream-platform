package handler

import (
	"errors"
	"net/http"

	"github.com/VladUrsul/livestream-platform/services/user-service/internal/domain"
	"github.com/VladUrsul/livestream-platform/services/user-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct{ svc service.UserService }

func New(svc service.UserService) *Handler { return &Handler{svc} }

func (h *Handler) Register(api *gin.RouterGroup, jwtSecret string) {
	// Public
	api.GET("/search", h.Search)
	api.GET("/internal/:userID/follower-ids", h.GetFollowerIDs)
	api.GET("/:username", h.GetProfile)

	// Authenticated
	a := api.Group("")
	a.Use(AuthMiddleware(jwtSecret))
	a.GET("/me", h.Me)
	a.PUT("/me", h.UpdateProfile)
	a.GET("/:username/follow", h.IsFollowing)
	a.POST("/:username/follow", h.Follow)
	a.DELETE("/:username/follow", h.Unfollow)
	a.GET("/me/following", h.GetFollowing)
}

// GET /api/v1/users/search?q=vlad
func (h *Handler) Search(c *gin.Context) {
	q := c.Query("q")
	results, err := h.svc.Search(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": results})
}

func (h *Handler) GetFollowerIDs(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	ids, err := h.svc.GetFollowerIDs(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"follower_ids": ids})
}

// GET /api/v1/users/:username/follow
func (h *Handler) IsFollowing(c *gin.Context) {
	followerID, ok := callerID(c)
	if !ok {
		return
	}
	target, err := h.svc.GetProfile(c.Request.Context(), c.Param("username"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	isFollowing, err := h.svc.IsFollowing(c.Request.Context(), followerID, target.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"following": isFollowing})
}

// GET /api/v1/users/me
func (h *Handler) Me(c *gin.Context) {
	userID, ok := callerID(c)
	if !ok {
		return
	}
	p, err := h.svc.GetProfileByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, p)
}

// GET /api/v1/users/:username
func (h *Handler) GetProfile(c *gin.Context) {
	p, err := h.svc.GetProfile(c.Request.Context(), c.Param("username"))
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusOK, p)
}

// PUT /api/v1/users/me
func (h *Handler) UpdateProfile(c *gin.Context) {
	userID, ok := callerID(c)
	if !ok {
		return
	}
	var input domain.UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.svc.UpdateProfile(c.Request.Context(), userID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}
	c.JSON(http.StatusOK, p)
}

// POST /api/v1/users/:username/follow
func (h *Handler) Follow(c *gin.Context) {
	followerID, ok := callerID(c)
	if !ok {
		return
	}
	target, err := h.svc.GetProfile(c.Request.Context(), c.Param("username"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if err := h.svc.Follow(c.Request.Context(), followerID, target.UserID); err != nil {
		if errors.Is(err, service.ErrCannotFollow) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot follow yourself"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "following"})
}

// GET /api/v1/users/me/following
func (h *Handler) GetFollowing(c *gin.Context) {
	userID, ok := callerID(c)
	if !ok {
		return
	}
	results, err := h.svc.GetFollowing(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": results})
}

// DELETE /api/v1/users/:username/follow
func (h *Handler) Unfollow(c *gin.Context) {
	followerID, ok := callerID(c)
	if !ok {
		return
	}
	target, err := h.svc.GetProfile(c.Request.Context(), c.Param("username"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	h.svc.Unfollow(c.Request.Context(), followerID, target.UserID)
	c.JSON(http.StatusOK, gin.H{"message": "unfollowed"})
}
