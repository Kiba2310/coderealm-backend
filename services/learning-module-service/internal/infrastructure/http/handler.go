package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"BootDevClone/services/learning-module-service/internal/infrastructure/repository"
	"BootDevClone/services/learning-module-service/internal/service"
	"BootDevClone/services/learning-module-service/pkg/types"
	"BootDevClone/shared/contracts"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *service.LearningService
}

func NewHandler(svc *service.LearningService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/languages", h.getLanguages)
	r.GET("/courses", h.listCourses)
	r.GET("/courses/:courseID", h.getCourse)
	r.GET("/courses/:courseID/chapters/:chapterID/lessons/:lessonID", h.getLesson)
	r.POST("/courses/:courseID/chapters/:chapterID/lessons/:lessonID/run", h.runCode)
	r.POST("/courses/:courseID/chapters/:chapterID/lessons/:lessonID/submit", h.submitAnswer)
	r.POST("/courses/:courseID/chapters/:chapterID/lessons/:lessonID/help", h.requestHelp)
	
	r.POST("/courses", h.createCourse)
	r.POST("/courses/:courseID/chapters", h.createChapter)
	r.POST("/courses/:courseID/chapters/:chapterID/lessons", h.createLesson)

	// Update/Delete routes
	r.PUT("/courses/:courseID", h.updateCourse)
	r.DELETE("/courses/:courseID", h.deleteCourse)
	r.PUT("/courses/:courseID/chapters/:chapterID", h.updateChapter)
	r.DELETE("/courses/:courseID/chapters/:chapterID", h.deleteChapter)
	r.PUT("/courses/:courseID/chapters/:chapterID/lessons/:lessonID", h.updateLesson)
	r.DELETE("/courses/:courseID/chapters/:chapterID/lessons/:lessonID", h.deleteLesson)

	// Document uploads
	r.POST("/courses/uploads/documents", h.uploadDocument)
	r.Static("/courses/uploads/documents", "./uploads/documents")
}

func (h *Handler) getLanguages(c *gin.Context) {
	langs, err := h.svc.GetLanguages()
	if err != nil {
		fmt.Printf("Error fetching languages: %v\n", err)
		c.JSON(http.StatusInternalServerError, contracts.APIResponse{
			Error: &contracts.APIError{Code: "INTERNAL_ERROR", Message: "Failed to get languages"},
		})
		return
	}
	c.JSON(http.StatusOK, contracts.APIResponse{Data: langs})
}

func (h *Handler) listCourses(c *gin.Context) {
	lang := c.Query("lang")
	courses, err := h.svc.ListCourses(lang)
	if err != nil {
		c.JSON(http.StatusInternalServerError, contracts.APIResponse{
			Error: &contracts.APIError{Code: "INTERNAL_ERROR", Message: "Internal server error"},
		})
		return
	}
	c.JSON(http.StatusOK, contracts.APIResponse{Data: courses})
}

func (h *Handler) getCourse(c *gin.Context) {
	courseID := c.Param("courseID")
	lang := c.Query("lang")
	detail, err := h.svc.GetCourseDetail(courseID, lang)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, contracts.APIResponse{
				Error: &contracts.APIError{Code: "NOT_FOUND", Message: "Course not found"},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, contracts.APIResponse{
			Error: &contracts.APIError{Code: "INTERNAL_ERROR", Message: "Internal server error"},
		})
		return
	}
	c.JSON(http.StatusOK, contracts.APIResponse{Data: detail})
}

func (h *Handler) getLesson(c *gin.Context) {
	courseID := c.Param("courseID")
	chapterID := c.Param("chapterID")
	lessonID := c.Param("lessonID")
	lang := c.Query("lang")

	lesson, err := h.svc.GetLesson(courseID, chapterID, lessonID, lang)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, contracts.APIResponse{
				Error: &contracts.APIError{Code: "NOT_FOUND", Message: "Lesson not found"},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, contracts.APIResponse{
			Error: &contracts.APIError{Code: "INTERNAL_ERROR", Message: "Internal server error"},
		})
		return
	}
	c.JSON(http.StatusOK, contracts.APIResponse{Data: lesson})
}

func (h *Handler) runCode(c *gin.Context) {
	courseID := c.Param("courseID")
	chapterID := c.Param("chapterID")
	lessonID := c.Param("lessonID")

	var req types.RunCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, contracts.APIResponse{
			Error: &contracts.APIError{Code: "BAD_REQUEST", Message: "Invalid request body"},
		})
		return
	}

	lang := c.Query("lang")
	resp, err := h.svc.RunCode(courseID, chapterID, lessonID, req, lang)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, contracts.APIResponse{
				Error: &contracts.APIError{Code: "NOT_FOUND", Message: "Lesson not found"},
			})
			return
		}
		c.JSON(http.StatusBadRequest, contracts.APIResponse{
			Error: &contracts.APIError{Code: "BAD_REQUEST", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, contracts.APIResponse{Data: resp})
}

func (h *Handler) submitAnswer(c *gin.Context) {
	courseID := c.Param("courseID")
	chapterID := c.Param("chapterID")
	lessonID := c.Param("lessonID")

	var req types.SubmitAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, contracts.APIResponse{
			Error: &contracts.APIError{Code: "BAD_REQUEST", Message: "Invalid request body"},
		})
		return
	}

	lang := c.Query("lang")
	userID := c.GetHeader("X-User-Id")
	resp, err := h.svc.SubmitAnswer(courseID, chapterID, lessonID, userID, req, lang)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, contracts.APIResponse{
				Error: &contracts.APIError{Code: "NOT_FOUND", Message: "Lesson not found"},
			})
			return
		}
		c.JSON(http.StatusBadRequest, contracts.APIResponse{
			Error: &contracts.APIError{Code: "BAD_REQUEST", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, contracts.APIResponse{Data: resp})
}

func (h *Handler) requestHelp(c *gin.Context) {
	courseID := c.Param("courseID")
	chapterID := c.Param("chapterID")
	lessonID := c.Param("lessonID")
	userID := c.GetHeader("X-User-Id")
	lang := c.Query("lang")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, contracts.APIResponse{
			Error: &contracts.APIError{Code: "UNAUTHORIZED", Message: "User not authenticated"},
		})
		return
	}

	resp, err := h.svc.RequestHelp(courseID, chapterID, lessonID, userID, lang)
	if err != nil {
		c.JSON(http.StatusBadRequest, contracts.APIResponse{
			Error: &contracts.APIError{Code: "BAD_REQUEST", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, contracts.APIResponse{Data: resp})
}

func (h *Handler) createCourse(c *gin.Context) {
	var req types.CreateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, contracts.APIResponse{
			Error: &contracts.APIError{Code: "BAD_REQUEST", Message: "Invalid request body"},
		})
		return
	}

	resp, err := h.svc.CreateCourse(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, contracts.APIResponse{
			Error: &contracts.APIError{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusCreated, contracts.APIResponse{Data: resp})
}

func (h *Handler) createChapter(c *gin.Context) {
	courseID := c.Param("courseID")
	var req types.CreateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, contracts.APIResponse{
			Error: &contracts.APIError{Code: "BAD_REQUEST", Message: "Invalid request body"},
		})
		return
	}

	resp, err := h.svc.CreateChapter(courseID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, contracts.APIResponse{
			Error: &contracts.APIError{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusCreated, contracts.APIResponse{Data: resp})
}

func (h *Handler) createLesson(c *gin.Context) {
	courseID := c.Param("courseID")
	chapterID := c.Param("chapterID")

	var req types.CreateLessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, contracts.APIResponse{
			Error: &contracts.APIError{Code: "BAD_REQUEST", Message: "Invalid request body"},
		})
		return
	}

	resp, err := h.svc.CreateLesson(courseID, chapterID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, contracts.APIResponse{
			Error: &contracts.APIError{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusCreated, contracts.APIResponse{Data: resp})
}

// ==================== Update/Delete Handlers ====================

func (h *Handler) updateCourse(c *gin.Context) {
	courseID := c.Param("courseID")
	var req types.CreateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, contracts.APIResponse{
			Error: &contracts.APIError{Code: "BAD_REQUEST", Message: "Invalid request body"},
		})
		return
	}
	if err := h.svc.UpdateCourse(courseID, req); err != nil {
		c.JSON(http.StatusInternalServerError, contracts.APIResponse{
			Error: &contracts.APIError{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, contracts.APIResponse{Data: map[string]interface{}{"success": true}})
}

func (h *Handler) deleteCourse(c *gin.Context) {
	courseID := c.Param("courseID")
	if err := h.svc.DeleteCourse(courseID); err != nil {
		c.JSON(http.StatusInternalServerError, contracts.APIResponse{
			Error: &contracts.APIError{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, contracts.APIResponse{Data: map[string]interface{}{"success": true}})
}

func (h *Handler) updateChapter(c *gin.Context) {
	courseID := c.Param("courseID")
	chapterID := c.Param("chapterID")
	var req types.CreateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, contracts.APIResponse{
			Error: &contracts.APIError{Code: "BAD_REQUEST", Message: "Invalid request body"},
		})
		return
	}
	if err := h.svc.UpdateChapter(courseID, chapterID, req); err != nil {
		c.JSON(http.StatusInternalServerError, contracts.APIResponse{
			Error: &contracts.APIError{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, contracts.APIResponse{Data: map[string]interface{}{"success": true}})
}

func (h *Handler) deleteChapter(c *gin.Context) {
	courseID := c.Param("courseID")
	chapterID := c.Param("chapterID")
	if err := h.svc.DeleteChapter(courseID, chapterID); err != nil {
		c.JSON(http.StatusInternalServerError, contracts.APIResponse{
			Error: &contracts.APIError{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, contracts.APIResponse{Data: map[string]interface{}{"success": true}})
}

func (h *Handler) updateLesson(c *gin.Context) {
	courseID := c.Param("courseID")
	chapterID := c.Param("chapterID")
	lessonID := c.Param("lessonID")
	_ = courseID // validated by URL routing
	var req types.CreateLessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, contracts.APIResponse{
			Error: &contracts.APIError{Code: "BAD_REQUEST", Message: "Invalid request body"},
		})
		return
	}
	if err := h.svc.UpdateLesson(courseID, chapterID, lessonID, req); err != nil {
		c.JSON(http.StatusInternalServerError, contracts.APIResponse{
			Error: &contracts.APIError{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, contracts.APIResponse{Data: map[string]interface{}{"success": true}})
}

func (h *Handler) deleteLesson(c *gin.Context) {
	courseID := c.Param("courseID")
	chapterID := c.Param("chapterID")
	lessonID := c.Param("lessonID")
	if err := h.svc.DeleteLesson(courseID, chapterID, lessonID); err != nil {
		c.JSON(http.StatusInternalServerError, contracts.APIResponse{
			Error: &contracts.APIError{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, contracts.APIResponse{Data: map[string]interface{}{"success": true}})
}

func (h *Handler) uploadDocument(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, contracts.APIResponse{
			Error: &contracts.APIError{Code: "BAD_REQUEST", Message: "No file uploaded or invalid form-data"},
		})
		return
	}
	defer file.Close()

	const maxUploadSize = 10 << 20 // 10 MB
	if header.Size > maxUploadSize {
		c.JSON(http.StatusBadRequest, contracts.APIResponse{
			Error: &contracts.APIError{Code: "BAD_REQUEST", Message: "File too large (max 10MB)"},
		})
		return
	}

	ext := filepath.Ext(header.Filename)
	if ext != ".pdf" && ext != ".doc" && ext != ".docx" && ext != ".ppt" && ext != ".pptx" {
		c.JSON(http.StatusBadRequest, contracts.APIResponse{
			Error: &contracts.APIError{Code: "BAD_REQUEST", Message: "Only PDF, Word, and PPT files are allowed"},
		})
		return
	}

	// Create uploads directory if not exists
	uploadDir := "./uploads/documents"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, contracts.APIResponse{
			Error: &contracts.APIError{Code: "INTERNAL_ERROR", Message: "Could not create upload directory"},
		})
		return
	}

	newFilename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	savePath := filepath.Join(uploadDir, newFilename)

	out, err := os.Create(savePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, contracts.APIResponse{
			Error: &contracts.APIError{Code: "INTERNAL_ERROR", Message: "Could not save file"},
		})
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, contracts.APIResponse{
			Error: &contracts.APIError{Code: "INTERNAL_ERROR", Message: "Error writing file"},
		})
		return
	}

	fileURL := fmt.Sprintf("/courses/uploads/documents/%s", newFilename)

	c.JSON(http.StatusOK, contracts.APIResponse{Data: map[string]interface{}{
		"url": fileURL,
	}})
}


