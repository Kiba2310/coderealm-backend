package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"BootDevClone/services/user-service/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc *service.UserService
}

func NewAuthHandler(svc *service.UserService, _ string) *AuthHandler {
	return &AuthHandler{
		svc: svc,
	}
}

func (h *AuthHandler) RegisterRoutes(r *gin.Engine) {
	auth := r.Group("/auth")
	{
		auth.POST("/sync", h.sync)
		auth.GET("/me", h.me)
		auth.GET("/me/lessons/completed", h.getCompletedLessons)
		auth.POST("/complete-lesson", h.completeLesson)
		auth.DELETE("/reset-progress", h.resetProgress)
		auth.PUT("/profile", h.updateProfile)
		auth.POST("/avatar", h.uploadAvatar)
		auth.POST("/achievements", h.addAchievement)
		auth.POST("/lessons/:lessonId/time", h.recordLessonTime)

		auth.GET("/leaderboard", h.getLeaderboard)
		auth.GET("/stats/daily", h.getDailyStats)
		auth.GET("/stats/public", h.adminGetStats)
		auth.GET("/achievements/stats", h.getAchievementStats)

		// Store endpoints
		auth.GET("/store", h.getStoreItems)
		auth.POST("/store/refresh", h.refreshStore)
		auth.POST("/store/buy/:id", h.buyStoreItem)
		auth.POST("/store/consume/helper", h.consumeHelper)

		// Public profile
		auth.GET("/users/:id/profile", h.getPublicProfile)

		// Daily login
		auth.POST("/daily-login/claim", h.claimDailyLogin)
		auth.GET("/daily-login/status", h.getDailyLoginStatus)

		// Inventory & Wardrobe endpoints
		auth.GET("/inventory", h.getUserInventory)
		auth.POST("/inventory/:itemId/toggle", h.toggleEquipItem)
		auth.DELETE("/inventory/:id", h.deleteInventoryItem)

		// Daily Challenge endpoints
		auth.GET("/daily-challenge/today", h.getTodayDailyChallenge)
		auth.POST("/daily-challenge/start", h.startDailyChallenge)
		auth.POST("/daily-challenge/run", h.runDailyChallengeCode)
		auth.POST("/daily-challenge/submit", h.submitDailyChallenge)
		auth.GET("/daily-challenge/leaderboard", h.getDailyChallengeLeaderboard)

		// Serve uploaded files statically
		auth.Static("/uploads", "./uploads")

		// Admin endpoints
		admin := auth.Group("/admin")
		{
			admin.GET("/users", h.adminListUsers)
			admin.GET("/users/:id", h.adminGetUser)
			admin.POST("/users/:id/ban", h.adminBanUser)
			admin.POST("/users/:id/unban", h.adminUnbanUser)
			admin.PUT("/users/:id/roles", h.adminSetUserRoles)
			admin.POST("/users/:id/reset-progress", h.adminResetProgress)
			admin.GET("/stats", h.adminGetStats)
			admin.GET("/stats/registrations", h.adminGetRegistrationTrend)
			admin.GET("/stats/completions", h.adminGetCompletionTrend)
			admin.GET("/users/:id/lessons/time", h.adminGetUserLessonTimes)
		}
	}
}

func (h *AuthHandler) sync(c *gin.Context) {
	keycloakID := c.GetHeader("X-User-Id")
	email := c.GetHeader("X-User-Email")
	rolesStr := c.GetHeader("X-User-Roles")

	if keycloakID == "" || email == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing user headers from API Gateway"})
		return
	}

	user, err := h.svc.SyncUser(keycloakID, email, rolesStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync user: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) me(c *gin.Context) {
	userID := c.GetHeader("X-User-Id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing user headers"})
		return
	}

	user, err := h.svc.SyncUser(userID, c.GetHeader("X-User-Email"), c.GetHeader("X-User-Roles"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

type completeLessonRequest struct {
	LessonID   string `json:"lesson_id" binding:"required"`
	LessonType string `json:"lesson_type"`
}

func (h *AuthHandler) completeLesson(c *gin.Context) {
	keycloakID := c.GetHeader("X-User-Id")
	if keycloakID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing user headers"})
		return
	}

	var req completeLessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lesson_id is required"})
		return
	}

	pointsAwarded, totalPoints, helperUsed, err := h.svc.CompleteLesson(keycloakID, req.LessonID, req.LessonType, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete lesson: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"points_awarded": pointsAwarded,
		"total_points":   totalPoints,
		"helper_used":    helperUsed,
	})
}

func (h *AuthHandler) getStoreItems(c *gin.Context) {
	keycloakID := c.GetHeader("X-User-Id")
	if keycloakID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing user headers"})
		return
	}

	user, err := h.svc.SyncUser(keycloakID, c.GetHeader("X-User-Email"), c.GetHeader("X-User-Roles"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	items, err := h.svc.GetStoreItems(user.ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get store items"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *AuthHandler) refreshStore(c *gin.Context) {
	keycloakID := c.GetHeader("X-User-Id")
	if keycloakID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing user headers"})
		return
	}

	user, err := h.svc.SyncUser(keycloakID, c.GetHeader("X-User-Email"), c.GetHeader("X-User-Roles"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	items, err := h.svc.RefreshStore(user.ID.String(), 50)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to refresh store: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items, "message": "Store refreshed successfully"})
}

func (h *AuthHandler) buyStoreItem(c *gin.Context) {
	keycloakID := c.GetHeader("X-User-Id")
	if keycloakID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing user headers"})
		return
	}
	itemID := c.Param("id")

	user, err := h.svc.SyncUser(keycloakID, c.GetHeader("X-User-Email"), c.GetHeader("X-User-Roles"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	err = h.svc.PurchaseStoreItem(user.ID.String(), itemID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Purchase failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Purchase successful"})
}

func (h *AuthHandler) resetProgress(c *gin.Context) {
	keycloakID := c.GetHeader("X-User-Id")
	if keycloakID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing user headers"})
		return
	}

	if err := h.svc.ResetProgress(keycloakID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset progress: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Progress reset successfully"})
}

type updateProfileRequest struct {
	Nickname  *string `json:"nickname"`
	Bio       *string `json:"bio"`
	AvatarURL *string `json:"avatar_url"`
}

func (h *AuthHandler) updateProfile(c *gin.Context) {
	keycloakID := c.GetHeader("X-User-Id")
	if keycloakID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing user headers"})
		return
	}

	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	user, err := h.svc.UpdateProfile(keycloakID, req.Nickname, req.Bio, req.AvatarURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) uploadAvatar(c *gin.Context) {
	keycloakID := c.GetHeader("X-User-Id")
	if keycloakID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing user headers"})
		return
	}

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded or invalid form-data"})
		return
	}
	defer file.Close()

	// 2MB limit
	const maxUploadSize = 2 << 20
	if header.Size > maxUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds 2MB limit"})
		return
	}

	// Validate extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only .jpg, .jpeg, .png, and .webp files are allowed"})
		return
	}

	// Read magic bytes for MIME validation
	buff := make([]byte, 512)
	if _, err := file.Read(buff); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	// Reset file pointer
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset file pointer"})
		return
	}

	mimeType := http.DetectContentType(buff)
	if !strings.HasPrefix(mimeType, "image/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file format"})
		return
	}

	// Save file
	newFilename := fmt.Sprintf("%s_%d%s", keycloakID, time.Now().Unix(), ext)
	savePath := filepath.Join("uploads", "avatars", newFilename)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(savePath), 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	out, err := os.Create(savePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file on server (dir check: ok)"})
		return
	}
	defer out.Close()

	if _, err = io.Copy(out, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file content"})
		return
	}
	// Explicitly close to flush to disk and release file lock before frontend requests it
	out.Close()

	// Store relative URL instead of absolute to avoid domain hardcoding
	avatarURL := fmt.Sprintf("/auth/uploads/avatars/%s", newFilename)

	// Update user profile with new avatar URL
	user, err := h.svc.UpdateProfile(keycloakID, nil, nil, &avatarURL)
	if err != nil {
		// Clean up the uploaded file since DB update failed
		_ = os.Remove(savePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user, "avatar_url": avatarURL})
}

type addAchievementRequest struct {
	AchievementID string `json:"achievement_id" binding:"required"`
}

func (h *AuthHandler) addAchievement(c *gin.Context) {
	keycloakID := c.GetHeader("X-User-Id")
	if keycloakID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing user headers"})
		return
	}

	var req addAchievementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	user, err := h.svc.AddAchievement(keycloakID, req.AchievementID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add achievement: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) getDailyStats(c *gin.Context) {
	keycloakID := c.GetHeader("X-User-Id")
	if keycloakID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing user headers"})
		return
	}

	stats, err := h.svc.GetDailyStats(keycloakID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get daily stats: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *AuthHandler) getAchievementStats(c *gin.Context) {
	stats, err := h.svc.GetAchievementStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stats: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *AuthHandler) getLeaderboard(c *gin.Context) {
	limit := 100
	users, err := h.svc.GetLeaderboard(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get leaderboard: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

// ==================== Admin Handlers ====================

func (h *AuthHandler) adminListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	result, err := h.svc.ListAllUsers(page, limit, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list users: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *AuthHandler) adminGetUser(c *gin.Context) {
	userID := c.Param("id")
	user, err := h.svc.AdminGetUser(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) adminBanUser(c *gin.Context) {
	userID := c.Param("id")
	if err := h.svc.BanUser(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to ban user: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User banned successfully"})
}

func (h *AuthHandler) adminUnbanUser(c *gin.Context) {
	userID := c.Param("id")
	if err := h.svc.UnbanUser(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unban user: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User unbanned successfully"})
}

func (h *AuthHandler) adminResetProgress(c *gin.Context) {
	userID := c.Param("id")
	if err := h.svc.AdminResetProgress(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset progress: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Progress reset successfully"})
}

func (h *AuthHandler) adminSetUserRoles(c *gin.Context) {
	userID := c.Param("id")
	var req struct {
		Roles []string `json:"roles" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	if err := h.svc.AdminSetUserRoles(userID, req.Roles); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set user roles: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Roles updated successfully"})
}

func (h *AuthHandler) adminGetStats(c *gin.Context) {
	stats, err := h.svc.GetPlatformStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stats: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *AuthHandler) adminGetRegistrationTrend(c *gin.Context) {
	period := c.DefaultQuery("period", "daily")
	trend, err := h.svc.GetRegistrationTrend(period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get registration trend: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, trend)
}

func (h *AuthHandler) adminGetCompletionTrend(c *gin.Context) {
	period := c.DefaultQuery("period", "daily")
	trend, err := h.svc.GetCompletionTrend(period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get completion trend: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, trend)
}

func (h *AuthHandler) recordLessonTime(c *gin.Context) {
	keycloakID := c.GetHeader("X-User-Id")
	lessonID := c.Param("lessonId")
	if keycloakID == "" || lessonID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user ID or lesson ID"})
		return
	}

	var req struct {
		Seconds int `json:"seconds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Seconds <= 0 {
		c.Status(http.StatusOK) // Ignore 0 or negative time
		return
	}

	if err := h.svc.RecordLessonTime(keycloakID, lessonID, req.Seconds); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record time"})
		return
	}
	c.Status(http.StatusOK)
}

func (h *AuthHandler) adminGetUserLessonTimes(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user ID"})
		return
	}

	stats, err := h.svc.AdminGetUserLessonTimes(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch lesson times"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *AuthHandler) getCompletedLessons(c *gin.Context) {
	keycloakID := c.GetHeader("X-User-Id")
	if keycloakID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing user ID"})
		return
	}

	lessons, err := h.svc.GetCompletedLessons(keycloakID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch completed lessons"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"completed_lessons": lessons})
}

func (h *AuthHandler) consumeHelper(c *gin.Context) {
	keycloakID := c.GetHeader("X-User-Id")
	if keycloakID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing user headers"})
		return
	}

	var req struct {
		LessonID string `json:"lessonId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	user, err := h.svc.SyncUser(keycloakID, c.GetHeader("X-User-Email"), c.GetHeader("X-User-Roles"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	consumed, err := h.svc.ConsumeHelperItem(user.ID, req.LessonID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !consumed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nincs helper tárgyad!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Helper tárgy aktiválva"})
}

func (h *AuthHandler) getPublicProfile(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user ID"})
		return
	}

	profile, err := h.svc.GetPublicProfile(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *AuthHandler) claimDailyLogin(c *gin.Context) {
	keycloakID := c.GetHeader("X-User-Id")
	if keycloakID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing user headers"})
		return
	}

	user, err := h.svc.SyncUser(keycloakID, c.GetHeader("X-User-Email"), c.GetHeader("X-User-Roles"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	result, err := h.svc.ClaimDailyLogin(user.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *AuthHandler) getDailyLoginStatus(c *gin.Context) {
	keycloakID := c.GetHeader("X-User-Id")
	if keycloakID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing user headers"})
		return
	}

	user, err := h.svc.SyncUser(keycloakID, c.GetHeader("X-User-Email"), c.GetHeader("X-User-Roles"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	year := time.Now().Year()
	month := int(time.Now().Month())

	status, err := h.svc.GetDailyLoginStatus(user.ID, year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

func (h *AuthHandler) getUserInventory(c *gin.Context) {
	keycloakID := c.GetHeader("X-User-Id")
	if keycloakID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing user headers"})
		return
	}

	user, err := h.svc.SyncUser(keycloakID, c.GetHeader("X-User-Email"), c.GetHeader("X-User-Roles"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	items, err := h.svc.GetUserInventory(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, items)
}

func (h *AuthHandler) toggleEquipItem(c *gin.Context) {
	keycloakID := c.GetHeader("X-User-Id")
	if keycloakID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing user headers"})
		return
	}

	itemID := c.Param("itemId")
	if itemID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing item ID"})
		return
	}

	user, err := h.svc.SyncUser(keycloakID, c.GetHeader("X-User-Email"), c.GetHeader("X-User-Roles"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	result, err := h.svc.ToggleEquipItem(user.ID, itemID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *AuthHandler) deleteInventoryItem(c *gin.Context) {
	keycloakID := c.GetHeader("X-User-Id")
	if keycloakID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing user headers"})
		return
	}

	inventoryID := c.Param("id")
	if inventoryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing inventory ID"})
		return
	}

	user, err := h.svc.SyncUser(keycloakID, c.GetHeader("X-User-Email"), c.GetHeader("X-User-Roles"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	result, err := h.svc.DeleteInventoryItem(user.ID, inventoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Daily Challenge Handlers

func (h *AuthHandler) getTodayDailyChallenge(c *gin.Context) {
	keycloakID := c.GetHeader("X-User-Id")
	if keycloakID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing user headers"})
		return
	}

	user, err := h.svc.SyncUser(keycloakID, c.GetHeader("X-User-Email"), c.GetHeader("X-User-Roles"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	resp, err := h.svc.GetTodayChallenge(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

type startDailyChallengeReq struct {
	Language string `json:"language"`
}

func (h *AuthHandler) startDailyChallenge(c *gin.Context) {
	keycloakID := c.GetHeader("X-User-Id")
	if keycloakID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing user headers"})
		return
	}

	var req startDailyChallengeReq
	_ = c.ShouldBindJSON(&req)

	user, err := h.svc.SyncUser(keycloakID, c.GetHeader("X-User-Email"), c.GetHeader("X-User-Roles"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	resp, err := h.svc.StartDailyChallenge(user.ID, req.Language)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

type runDailyChallengeReq struct {
	Language string `json:"language" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

func (h *AuthHandler) runDailyChallengeCode(c *gin.Context) {
	var req runDailyChallengeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Hiányzó nyelv vagy forráskód!"})
		return
	}

	resp, err := h.svc.RunDailyChallengeCode(req.Language, req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) submitDailyChallenge(c *gin.Context) {
	keycloakID := c.GetHeader("X-User-Id")
	if keycloakID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing user headers"})
		return
	}

	var req runDailyChallengeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Hiányzó nyelv vagy forráskód!"})
		return
	}

	user, err := h.svc.SyncUser(keycloakID, c.GetHeader("X-User-Email"), c.GetHeader("X-User-Roles"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	resp, err := h.svc.SubmitDailyChallenge(user.ID, req.Language, req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) getDailyChallengeLeaderboard(c *gin.Context) {
	leaderboard, err := h.svc.GetDailyChallengeLeaderboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, leaderboard)
}

