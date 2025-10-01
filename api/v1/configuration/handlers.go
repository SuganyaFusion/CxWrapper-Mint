package configuration

// ConfigurationHandlers holds the configuration service for HTTP operations
// type ConfigurationHandlers struct {
// 	configService *ConfigurationService
// 	logger        util.Logger
// }

// // NewConfigurationHandlers creates a new configuration handler
// func NewConfigurationHandlers(service *ConfigurationService, logger util.Logger) *ConfigurationHandlers {
// 	return &ConfigurationHandlers{
// 		configService: service,
// 		logger:        logger,
// 	}
// }

// // GetProjectParameters handles GET /v1/configuration/project
// func (h *ConfigurationHandlers) GetProjectParameters(c *gin.Context) {
// 	projectID := c.Query("project-id")
// 	if projectID == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "project-id query parameter is required",
// 		})
// 		return
// 	}

// 	h.logger.Infof("Getting parameters for project: %s", projectID)

// 	response, err := h.configService.GetProjectParameters(projectID)
// 	if err != nil {
// 		h.logger.Errorf("Failed to get project parameters for %s: %v", projectID, err)
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error":   "Failed to retrieve project parameters",
// 			"details": err.Error(),
// 		})
// 		return
// 	}
// 	c.JSON(http.StatusOK, response)
// }

// // SetProjectParameters handles PATCH /v1/configuration/project
// func (h *ConfigurationHandlers) SetProjectParameters(c *gin.Context) {
// 	projectID := c.Query("project-id")
// 	if projectID == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "project-id query parameter is required",
// 		})
// 		return
// 	}

// 	const projectIDRegex = `^[a-zA-Z0-9-_]{1,50}$`

// 	match, err := regexp.MatchString(projectIDRegex, projectID)
// 	if err != nil || !match {
// 		h.logger.Errorf("Invalid 'project-id' provided: %s", projectID)
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "Invalid project-id format",
// 		})
// 		return
// 	}

// 	var req []cx1.ConfigurationSetting

// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		h.logger.Errorf("Invalid request body for project parameters: %v", err)
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error":   "Invalid request body",
// 			"details": err.Error(),
// 		})
// 		return
// 	}

// 	h.logger.Infof("Setting %d parameters for project: %s", len(req), projectID)

// 	if err := h.configService.SetProjectParameters(projectID, req); err != nil {
// 		h.logger.Errorf("Failed to set project parameters for %s: %v", projectID, err)
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error":   "Failed to set project parameters",
// 			"details": err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusNoContent, nil)
// }

// // DeleteProjectParameters handles DELETE /v1/configuration/project
// func (h *ConfigurationHandlers) DeleteProjectParameters(c *gin.Context) {
// 	projectID := c.Query("project-id")
// 	if projectID == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "project-id query parameter is required",
// 		})
// 		return
// 	}

// 	configKeysParam := c.Query("config-keys")
// 	if configKeysParam == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "config-keys query parameter is required",
// 		})
// 		return
// 	}

// 	//configKeys := strings.Split(configKeysParam, ",")
// 	//for i := range configKeys {
// 	//	configKeys[i] = strings.TrimSpace(configKeys[i])
// 	//}

// 	h.logger.Infof("Deleting parameters for project %s: %v", projectID, configKeysParam)

// 	if err := h.configService.DeleteProjectParameters(projectID, configKeysParam); err != nil {
// 		h.logger.Errorf("Failed to delete project parameters for %s: %v", projectID, err)
// 		if strings.Contains(err.Error(), "not found") {
// 			c.JSON(http.StatusNotFound, gin.H{
// 				"error":   "Project or parameters not found",
// 				"details": err.Error(),
// 			})
// 		} else {
// 			c.JSON(http.StatusInternalServerError, gin.H{
// 				"error":   "Failed to delete project parameters",
// 				"details": err.Error(),
// 			})
// 		}
// 		return
// 	}

// 	c.JSON(http.StatusNoContent, nil)
// }

// // Scan Configuration Handlers

// // GetScanParameters handles GET /v1/configuration/scan
// func (h *ConfigurationHandlers) GetScanParameters(c *gin.Context) {
// 	projectID := c.Query("project-id")
// 	if projectID == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "project-id query parameter is required",
// 		})
// 		return
// 	}

// 	scanID := c.Query("scan-id")
// 	if scanID == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "scan-id query parameter is required",
// 		})
// 		return
// 	}

// 	h.logger.Infof("Getting scan parameters for project: %s, scan: %s", projectID, scanID)

// 	response, err := h.configService.GetScanParameters(projectID, scanID)
// 	if err != nil {
// 		h.logger.Errorf("Failed to get scan parameters for project %s, scan %s: %v", projectID, scanID, err)
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error":   "Failed to retrieve scan parameters",
// 			"details": err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, response)
// }

// // RegisterRoutes registers all configuration routes with the given router group
// func (h *ConfigurationHandlers) RegisterRoutes(v1 *gin.RouterGroup) {
// 	config := v1.Group("/configuration")
// 	{
// 		// Tenant configuration routes
// 		//tenant := config.Group("/tenant")
// 		//{
// 		//tenant.GET("", h.GetTenantParameters)
// 		//tenant.PATCH("", h.SetTenantParameters)
// 		//tenant.DELETE("", h.DeleteTenantParameters)
// 		//}

// 		project := config.Group("/project")
// 		{
// 			project.GET("", h.GetProjectParameters)
// 			project.PATCH("", h.SetProjectParameters)
// 			project.DELETE("", h.DeleteProjectParameters)
// 		}

// 		scan := config.Group("/scan")
// 		{
// 			scan.GET("", h.GetScanParameters)
// 		}
// 	}
// }