package http

import (
	"net/http"

	"BootDevClone/services/test-service/internal/service"
	"BootDevClone/services/test-service/pkg/types"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *service.TestService
}

func NewHandler(svc *service.TestService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/tests", h.getTests)
	r.GET("/tests/:testID", h.getTestByID)
	r.POST("/tests", h.createTest)
	r.PUT("/tests/:testID", h.updateTest)
	r.DELETE("/tests/:testID", h.deleteTest)
	r.POST("/tests/:testID/submit", h.submitTest)
	r.POST("/tests/:testID/run", h.runTest)
	r.GET("/tests/group/:groupID", h.getTestsByGroup)
}

func (h *Handler) getTests(c *gin.Context) {
	tests, err := h.svc.GetTests()
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.APIResponse{
			Error: &types.APIError{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, types.APIResponse{Data: tests})
}

func (h *Handler) getTestByID(c *gin.Context) {
	testID := c.Param("testID")
	test, err := h.svc.GetTestByID(testID)
	if err != nil {
		c.JSON(http.StatusNotFound, types.APIResponse{
			Error: &types.APIError{Code: "NOT_FOUND", Message: "test not found"},
		})
		return
	}
	c.JSON(http.StatusOK, types.APIResponse{Data: test})
}

func (h *Handler) getTestsByGroup(c *gin.Context) {
	groupID := c.Param("groupID")
	tests, err := h.svc.GetTestsByGroupID(groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.APIResponse{
			Error: &types.APIError{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, types.APIResponse{Data: tests})
}

func (h *Handler) submitTest(c *gin.Context) {
	testID := c.Param("testID")
	var req types.SubmitTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Error: &types.APIError{Code: "BAD_REQUEST", Message: "Invalid request body"},
		})
		return
	}
	resp, err := h.svc.SubmitTest(testID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.APIResponse{
			Error: &types.APIError{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, types.APIResponse{Data: resp})
}

func (h *Handler) runTest(c *gin.Context) {
	testID := c.Param("testID")
	var req types.SubmitTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Error: &types.APIError{Code: "BAD_REQUEST", Message: "Invalid request body"},
		})
		return
	}
	resp, err := h.svc.RunTest(testID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.APIResponse{
			Error: &types.APIError{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, types.APIResponse{Data: resp})
}

func (h *Handler) createTest(c *gin.Context) {
	var req types.CreateTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Error: &types.APIError{Code: "BAD_REQUEST", Message: "Invalid request body"},
		})
		return
	}
	resp, err := h.svc.CreateTest(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.APIResponse{
			Error: &types.APIError{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusCreated, types.APIResponse{Data: resp})
}

func (h *Handler) updateTest(c *gin.Context) {
	testID := c.Param("testID")
	var req types.UpdateTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.APIResponse{
			Error: &types.APIError{Code: "BAD_REQUEST", Message: "Invalid request body"},
		})
		return
	}
	if err := h.svc.UpdateTest(testID, req); err != nil {
		c.JSON(http.StatusInternalServerError, types.APIResponse{
			Error: &types.APIError{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, types.APIResponse{Data: map[string]interface{}{"success": true}})
}

func (h *Handler) deleteTest(c *gin.Context) {
	testID := c.Param("testID")
	if err := h.svc.DeleteTest(testID); err != nil {
		c.JSON(http.StatusInternalServerError, types.APIResponse{
			Error: &types.APIError{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, types.APIResponse{Data: map[string]interface{}{"success": true}})
}
