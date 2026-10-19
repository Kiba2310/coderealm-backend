package api

import (
	"net/http"
	"strings"

	"BootDevClone/services/group-service/internal/service"
	"BootDevClone/services/group-service/pkg/types"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *service.GroupService
}

func NewHandler(svc *service.GroupService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "group-service"})
	})

	api := r.Group("/groups")
	{
		api.POST("", h.createGroup)
		api.POST("/join", h.joinGroup)
		api.GET("/teacher", h.listTeacherGroups)
		api.GET("/student", h.listStudentGroups)
		api.GET("/admin/all", h.adminListAllGroups)
		api.GET("/:id", h.getGroup)
		api.GET("/:id/members", h.getGroupMembers)
		api.DELETE("/:id", h.deleteGroup)
		api.DELETE("/:id/members/:memberId", h.removeMember)
	}
}

func (h *Handler) createGroup(c *gin.Context) {
	userID := c.GetHeader("X-User-Id")
	userRoles := c.GetHeader("X-User-Roles")

	// Ensure the user is a teacher or admin
	if !strings.Contains(userRoles, "teacher") && !strings.Contains(userRoles, "admin") {
		c.JSON(http.StatusForbidden, types.APIResponse{
			Error: &types.APIError{Code: "FORBIDDEN", Message: "Only teachers can create groups"},
		})
		return
	}

	var req types.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Error: &types.APIError{Code: "BAD_REQUEST", Message: "Invalid request payload"},
		})
		return
	}

	group, err := h.svc.CreateGroup(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.APIResponse{
			Error: &types.APIError{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusCreated, types.APIResponse{Data: group})
}

func (h *Handler) joinGroup(c *gin.Context) {
	userID := c.GetHeader("X-User-Id")
	
	var req types.JoinGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Error: &types.APIError{Code: "BAD_REQUEST", Message: "Invalid request payload"},
		})
		return
	}

	group, err := h.svc.JoinGroup(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Error: &types.APIError{Code: "BAD_REQUEST", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, types.APIResponse{Data: group})
}

func (h *Handler) listTeacherGroups(c *gin.Context) {
	userID := c.GetHeader("X-User-Id")
	groups, err := h.svc.ListTeacherGroups(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.APIResponse{
			Error: &types.APIError{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, types.APIResponse{Data: groups})
}

func (h *Handler) listStudentGroups(c *gin.Context) {
	userID := c.GetHeader("X-User-Id")
	groups, err := h.svc.ListStudentGroups(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.APIResponse{
			Error: &types.APIError{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, types.APIResponse{Data: groups})
}

func (h *Handler) adminListAllGroups(c *gin.Context) {
	userRoles := c.GetHeader("X-User-Roles")
	if !strings.Contains(userRoles, "admin") {
		c.JSON(http.StatusForbidden, types.APIResponse{
			Error: &types.APIError{Code: "FORBIDDEN", Message: "Only admins can list all groups"},
		})
		return
	}

	groups, err := h.svc.AdminListAllGroups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.APIResponse{
			Error: &types.APIError{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, types.APIResponse{Data: groups})
}

func (h *Handler) getGroup(c *gin.Context) {
	groupID := c.Param("id")
	group, err := h.svc.GetGroupByID(groupID)
	if err != nil {
		c.JSON(http.StatusNotFound, types.APIResponse{
			Error: &types.APIError{Code: "NOT_FOUND", Message: "Group not found"},
		})
		return
	}
	c.JSON(http.StatusOK, types.APIResponse{Data: group})
}

func (h *Handler) getGroupMembers(c *gin.Context) {
	groupID := c.Param("id")
	members, err := h.svc.GetGroupMembers(groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.APIResponse{
			Error: &types.APIError{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, types.APIResponse{Data: members})
}

func (h *Handler) deleteGroup(c *gin.Context) {
	groupID := c.Param("id")
	userID := c.GetHeader("X-User-Id")
	
	if err := h.svc.DeleteGroup(groupID, userID); err != nil {
		c.JSON(http.StatusForbidden, types.APIResponse{
			Error: &types.APIError{Code: "FORBIDDEN", Message: err.Error()},
		})
		return
	}
	
	c.JSON(http.StatusOK, types.APIResponse{Data: map[string]interface{}{"success": true}})
}

func (h *Handler) removeMember(c *gin.Context) {
	groupID := c.Param("id")
	memberID := c.Param("memberId")
	userID := c.GetHeader("X-User-Id")
	userRoles := c.GetHeader("X-User-Roles")

	if err := h.svc.RemoveMember(groupID, memberID, userID, userRoles); err != nil {
		c.JSON(http.StatusForbidden, types.APIResponse{
			Error: &types.APIError{Code: "FORBIDDEN", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, types.APIResponse{Data: map[string]interface{}{"success": true}})
}
