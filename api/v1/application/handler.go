package application

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/madhatkul/CxWrapper-v2/util"
)

type ApplicationHandler struct {
	service *ApplicationService
	logger  util.Logger
}

func NewApplicationHandler(service *ApplicationService, logger util.Logger) *ApplicationHandler {
	return &ApplicationHandler{
		service: service,
		logger:  logger,
	}
}

func (h *ApplicationHandler) RegisterRoutes(v1 *gin.RouterGroup) {
	appGroup := v1.Group("/applications")
	{
		appGroup.POST("/projects", h.AssignProjectToApp)
	}
}

var validate = validator.New()

func (h *ApplicationHandler) AssignProjectToApp(c *gin.Context) {
	var req AssignProjectRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	h.logger.Infof("Attempting to assign project '%s' to application '%s'", req.ProjectName, req.AppName)

	if len(req.ProjectName) == 0 || req.ProjectName == "" || len(req.AppName) == 0 || req.AppName == "" {
		h.logger.Errorf("Project name or application name is empty")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project name and application name are required"})
		return
	}

	err := h.service.AssignProjectToApp(req.AppName, req.ProjectName)
	if err != nil {
		h.logger.Errorf("Failed to assign project '%s' to app '%s': %v", req.ProjectName, req.AppName, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign project", "details": err.Error()})
		return
	}

	h.logger.Infof("✅ Successfully assigned project '%s' to application '%s'", req.ProjectName, req.AppName)
	c.JSON(http.StatusOK, gin.H{
		"message":     "Project assigned to application successfully",
		"application": req.AppName,
		"project":     req.ProjectName,
	})
}