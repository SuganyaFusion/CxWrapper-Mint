package projects

// type ProjectHandlers struct {
// 	projectService *ProjectService
// 	logger         util.Logger
// }

// func NewProjectHandlers(service *ProjectService, logger util.Logger) *ProjectHandlers {
// 	return &ProjectHandlers{
// 		projectService: service,
// 		logger:         logger,
// 	}
// }

// // ListProjects handles GET /v1/projects
// func (ph *ProjectHandlers) ListProjects(c *gin.Context) {
// 	// Parse query parameters
// 	limit := 20
// 	offset := 0
// 	nameFilter := c.Query("name")

// 	if l := c.Query("limit"); l != "" {
// 		if parsedLimit, err := strconv.Atoi(l); err == nil && parsedLimit > 0 {
// 			limit = parsedLimit
// 		}
// 	}

// 	if o := c.Query("offset"); o != "" {
// 		if parsedOffset, err := strconv.Atoi(o); err == nil && parsedOffset >= 0 {
// 			offset = parsedOffset
// 		}
// 	}

// 	// Validate limit (max 100)
// 	if limit > 100 {
// 		limit = 100
// 	}

// 	// Get projects from service
// 	response, err := ph.projectService.ListProjects(limit, offset, nameFilter)
// 	if err != nil {
// 		ph.logger.Errorf("Failed to list projects: %v", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error":   "Failed to retrieve projects",
// 			"details": err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, response)
// }

// // CreateProject handles POST /v1/projects
// func (ph *ProjectHandlers) CreateProject(c *gin.Context) {
// 	var req CreateProjectRequest

// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		ph.logger.Errorf("Invalid request body: %v", err)
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error":   "Invalid request body",
// 			"details": err.Error(),
// 		})
// 		return
// 	}

// 	// Create project via service
// 	response, err := ph.projectService.CreateProject(&req)
// 	if err != nil {
// 		ph.logger.Errorf("Failed to create project: %v", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error":   "Failed to create project",
// 			"details": err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusCreated, response)
// }

// // GetProject handles GET /v1/projects/{id}
// func (ph *ProjectHandlers) GetProject(c *gin.Context) {
// 	projectID := c.Param("id")

// 	if projectID == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "Project ID is required",
// 		})
// 		return
// 	}

// 	// Get project via service
// 	response, err := ph.projectService.GetProjectByID(projectID)
// 	if err != nil {
// 		ph.logger.Errorf("Failed to get project %s: %v", projectID, err)
// 		c.JSON(http.StatusNotFound, gin.H{
// 			"error":   "Project not found",
// 			"details": err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, response)
// }

// // UpdateProject handles PUT /v1/projects/{id}
// func (ph *ProjectHandlers) UpdateProject(c *gin.Context) {
// 	projectID := c.Param("id")

// 	if projectID == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "Project ID is required",
// 		})
// 		return
// 	}

// 	var req UpdateProjectRequest

// 	// Use ShouldBindJSON only with structs that have exported fields and proper binding tags
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		ph.logger.Errorf("Invalid request body: %v", err)
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error":   "Invalid request body",
// 			"details": err.Error(),
// 		})
// 		return
// 	}

// 	// Update project via service
// 	response, err := ph.projectService.UpdateProject(projectID, &req)
// 	if err != nil {
// 		ph.logger.Errorf("Failed to update project %s: %v", projectID, err)
// 		if err.Error() == "project not found" {
// 			c.JSON(http.StatusNotFound, gin.H{
// 				"error":   "Project not found",
// 				"details": err.Error(),
// 			})
// 		} else {
// 			c.JSON(http.StatusInternalServerError, gin.H{
// 				"error":   "Failed to update project",
// 				"details": err.Error(),
// 			})
// 		}
// 		return
// 	}

// 	c.JSON(http.StatusOK, response)
// }

// // DeleteProject handles DELETE /v1/projects/{id}
// func (ph *ProjectHandlers) DeleteProject(c *gin.Context) {
// 	projectID := c.Param("id")

// 	if projectID == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "Project ID is required",
// 		})
// 		return
// 	}

// 	// Delete project via service
// 	err := ph.projectService.DeleteProject(projectID)
// 	if err != nil {
// 		ph.logger.Errorf("Failed to delete project %s: %v", projectID, err)
// 		if err.Error() == "project not found" {
// 			c.JSON(http.StatusNotFound, gin.H{
// 				"error":   "Project not found",
// 				"details": err.Error(),
// 			})
// 		} else {
// 			c.JSON(http.StatusInternalServerError, gin.H{
// 				"error":   "Failed to delete project",
// 				"details": err.Error(),
// 			})
// 		}
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "Project deleted successfully",
// 	})
// }

// // GetProjectScans handles GET /v1/projects/{id}/scans
// func (ph *ProjectHandlers) GetProjectLastScan(c *gin.Context) {
// 	projectID := c.Param("id")

// 	if projectID == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "Project ID is required",
// 		})
// 		return
// 	}

// 	// Parse query parameters
// 	limit := 20
// 	offset := 0

// 	if l := c.Query("limit"); l != "" {
// 		if parsedLimit, err := strconv.Atoi(l); err == nil && parsedLimit > 0 {
// 			limit = parsedLimit
// 		}
// 	}

// 	if o := c.Query("offset"); o != "" {
// 		if parsedOffset, err := strconv.Atoi(o); err == nil && parsedOffset >= 0 {
// 			offset = parsedOffset
// 		}
// 	}

// 	// Validate limit (max 100)
// 	if limit > 100 {
// 		limit = 100
// 	}

// 	// Get project scans via service
// 	response, err := ph.projectService.GetProjectLastScan(projectID, limit, offset)
// 	if err != nil {
// 		ph.logger.Errorf("Failed to get scans for project %s: %v", projectID, err)
// 		if err.Error() == "project not found" {
// 			c.JSON(http.StatusNotFound, gin.H{
// 				"error":   "Project not found",
// 				"details": err.Error(),
// 			})
// 		} else {
// 			c.JSON(http.StatusInternalServerError, gin.H{
// 				"error":   "Failed to retrieve project scans",
// 				"details": err.Error(),
// 			})
// 		}
// 		return
// 	}

// 	c.JSON(http.StatusOK, response)
// }

// // RegisterRoutes registers all project routes with the given router group
// func (ph *ProjectHandlers) RegisterRoutes(v1 *gin.RouterGroup) {
// 	projects := v1.Group("/projects")
// 	{
// 		projects.GET("", ph.ListProjects)
// 		projects.POST("", ph.CreateProject)
// 		projects.GET("/:id", ph.GetProject)
// 		projects.PUT("/:id", ph.UpdateProject)
// 		projects.DELETE("/:id", ph.DeleteProject)
// 		projects.GET("/:id/scans", ph.GetProjectLastScan)
// 	}
// }