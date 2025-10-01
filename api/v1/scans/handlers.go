// api/v1/scans/handlers.go
package scans

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	cx1 "github.com/madhatkul/CxWrapper-v2/Cx1ClientGo"
	"github.com/madhatkul/CxWrapper-v2/util"
)

type ScanHandler struct {
	service *ScanService
	logger  util.Logger
}

func NewScanHandler(service *ScanService, logger util.Logger) *ScanHandler {
	return &ScanHandler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers all scan routes with the given router group
func (sh *ScanHandler) RegisterRoutes(v1 *gin.RouterGroup) {
	scans := v1.Group("/scans")
	{
		// Apply file size limit middleware to upload endpoints
		staticGroup := scans.Group("/static")
		//uploadGroup.Use(FileSizeLimitMiddleware(500 << 20))
		{
			staticGroup.POST("", sh.StartStaticScan)
			staticGroup.GET("", sh.ListScans)
			staticGroup.GET("/results", sh.GetScanResults)
			staticGroup.GET("/status", sh.GetScanStatus)
			staticGroup.POST("/cancel", sh.CancelScan)
			staticGroup.GET("/presets", sh.getPreset)
			// staticGroup.GET("/config", sh.getTempConfig)
			//staticGroup.GET("PollingStatus", sh.PollingStatus)
		}
	}
}

func (sh *ScanHandler) StartStaticScan(c *gin.Context) {
	sh.logger.Infof("🚀 Static scan handler reached")

	if err := c.Request.ParseMultipartForm(1 << 30); err != nil {
		sh.logger.Errorf("❌ Failed to parse multipart form: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Failed to parse multipart form", Details: err.Error()})
		return
	}

	// Read form values directly for reliability
	appName := c.Request.FormValue("app_name")
	projectName := c.Request.FormValue("project_name")
	branch := c.Request.FormValue("branch")
	commitID := c.Request.FormValue("commit_id")
	scanTypesStr := c.Request.FormValue("scan_types")
	tagsStr := c.Request.FormValue("tags")
	isFastScanStr := c.Request.FormValue("is_fast_scan")
	presetStr := c.Request.FormValue("preset")

	sh.logger.Infof("Received raw 'is_fast_scan' value from form: '%s'", isFastScanStr)

	if appName == "" || projectName == "" || branch == "" || commitID == "" || scanTypesStr == "" || isFastScanStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Missing required fields: app_name, project_name, branch, commit_id, scan_types,is_fast_scan "})
		return
	}

	file, fileHeader, err := c.Request.FormFile("zip_file")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "No zip file uploaded", Details: err.Error()})
		return
	}
	defer file.Close()

	// fileContents, err := io.ReadAll(file)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to read file contents", Details: err.Error()})
	// 	return
	// }

	scanTypes, _ := sh.parseScanTypes(scanTypesStr)
	tags, _ := sh.parseTags(tagsStr)
	tags["commit_id"] = commitID

	isFastScan := strings.ToLower(isFastScanStr) == "true"
	sh.logger.Infof("Parsed 'is_fast_scan' as: %v", isFastScan)

	req := StaticScanRequestWithFile{
		AppName:     appName,
		ProjectName: projectName,
		Branch:      branch,
		CommitID:    commitID,
		ScanTypes:   scanTypes,
		IsFastScan:  isFastScan,
		Preset:      presetStr,
		Tags:        tags,
		File:        file,
		FileSize:    fileHeader.Size,
		FileName:    fileHeader.Filename,
	}

	scan, err := sh.service.StartStaticScanWithFile(req)
	if err != nil {
		sh.logger.Errorf("❌ Failed to start static scan: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to start scan", Details: err.Error()})
		return
	}

	sh.logger.Infof("✅ Static scan initiated successfully: ScanID=%s", scan.ScanID)
	c.JSON(http.StatusOK, ScanResponse{
		ScanID:  scan.ScanID,
		Status:  "started",
		Message: "Static scan initiated successfully",
	})
}

// GetScanStatus gets scan status by commit_id
func (sh *ScanHandler) GetScanStatus(c *gin.Context) {
	commitID := c.Query("commit_id")
	projectName := c.Query("project_name") // Optional: to narrow down search

	// Validate required parameters
	if commitID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:     "Missing required parameter",
			Details:   "commit_id is required",
			Timestamp: time.Now().Format(time.RFC3339),
			Path:      c.Request.URL.Path,
		})
		return
	}

	status, err := sh.service.GetScanStatusByCommitID(commitID, projectName)
	if err != nil {
		// Determine appropriate HTTP status code based on error type
		statusCode := http.StatusInternalServerError
		if err.Error() == "scan not found" || err.Error() == "no scans found for commit_id" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, ErrorResponse{
			Error:     "Failed to get scan status",
			Details:   err.Error(),
			Timestamp: time.Now().Format(time.RFC3339),
			Path:      c.Request.URL.Path,
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetScanResults returns the results for a specific scan
func (sh *ScanHandler) GetScanResults(c *gin.Context) {
	commitID := c.Query("commit_id")
	projectName := c.Query("project_name") // Optional: to narrow down search

	// Validate required parameters
	if commitID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:     "Missing required parameter",
			Details:   "commit_id is required",
			Timestamp: time.Now().Format(time.RFC3339),
			Path:      c.Request.URL.Path,
		})
		return
	}

	results, err := sh.service.GetAllScanResultsByCommitID(commitID, projectName)
	if err != nil {
		// Determine appropriate HTTP status code based on error type
		statusCode := http.StatusInternalServerError
		//if contains(err.Error(), "scan not found") ||
		//	contains(err.Error(), "no scans found") {
		//	statusCode = http.StatusNotFound
		//} else if contains(err.Error(), "not completed") {
		//	statusCode = http.StatusPreconditionFailed // 412 - scan not ready
		//}

		c.JSON(statusCode, ErrorResponse{
			Error:     "Failed to get scan results",
			Details:   err.Error(),
			Timestamp: time.Now().Format(time.RFC3339),
			Path:      c.Request.URL.Path,
		})
		return
	}

	c.JSON(http.StatusOK, results)
}

// ListScans lists scans with filtering options
func (sh *ScanHandler) ListScans(c *gin.Context) {
	var req ListScansRequest

	// Parse query parameters
	if projectName := c.Query("project_name"); projectName != "" {
		req.ProjectName = projectName
	}

	if commitID := c.Query("commit_id"); commitID != "" {
		req.CommitID = commitID
	}

	// Parse pagination parameters
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			req.Limit = l
		}
	}
	if req.Limit <= 0 {
		req.Limit = 20 // Default limit
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); o >= 0 && err == nil {
			req.Offset = o
		}
	}

	// Call service method to get filtered scans
	response, err := sh.service.ListScansFiltered(req)
	if err != nil {
		sh.logger.Errorf("❌ Failed to list scans: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:     "Failed to list scans",
			Details:   err.Error(),
			Timestamp: time.Now().Format(time.RFC3339),
			Path:      c.Request.URL.Path,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CancelScan cancels a running scan
func (sh *ScanHandler) CancelScan(c *gin.Context) {
	commitID := c.Query("commit_id")
	projectName := c.Query("project_name") // Optional query parameter

	// Route would be: /api/scans/:commit_id/cancel?project_name=optional

	err := sh.service.CancelScan(commitID, projectName)
	if err != nil {
		if err.Error() == "scan not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:     err.Error(),
				Timestamp: time.Now().Format(time.RFC3339),
				Path:      c.Request.URL.Path,
			})
		} else {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:     err.Error(),
				Timestamp: time.Now().Format(time.RFC3339),
				Path:      c.Request.URL.Path,
			})
		}
		return
	}

	response := gin.H{
		"commit_id": commitID,
		"status":    "cancelled",
		"message":   "SUCCESS",
	}

	// Only include project_name in response if it was provided
	if projectName != "" {
		response["project_name"] = projectName
	}

	c.JSON(http.StatusOK, response)
}

func (sh *ScanHandler) parseScanConfigurations(configStr string, scanTypes []string) ([]cx1.ScanConfiguration, error) {
	var configurations []cx1.ScanConfiguration

	// If no config provided, generate default configurations
	if configStr == "" {
		sh.logger.Infof("No scan configuration provided, generating defaults")
		return sh.generateDefaultConfigurations(scanTypes), nil
	}

	// Parse JSON configuration
	if err := json.Unmarshal([]byte(configStr), &configurations); err != nil {
		return nil, fmt.Errorf("invalid JSON format in config: %v", err)
	}

	// Validate configurations
	if err := sh.validateScanConfigurations(configurations, scanTypes); err != nil {
		return nil, err
	}

	sh.logger.Infof("Using provided scan configurations: %+v", configurations)
	return configurations, nil
}

// validateScanConfigurations validates that provided configurations match requested scan types
func (sh *ScanHandler) validateScanConfigurations(configurations []cx1.ScanConfiguration, scanTypes []string) error {
	// Create a map of valid scan types
	validScanTypes := map[string]bool{
		"sast":         true,
		"sca":          true,
		"secrets":      true,
		"kics":         true,
		"containersec": true,
		"apisec":       true,
	}

	// Create a map of requested scan types
	requestedTypes := make(map[string]bool)
	for _, scanType := range scanTypes {
		requestedTypes[strings.ToLower(scanType)] = true
	}

	// Validate each configuration
	configTypes := make(map[string]bool)
	for i, config := range configurations {
		// Check if scan type is valid
		if !validScanTypes[strings.ToLower(config.ScanType)] {
			return fmt.Errorf("invalid scan type in configuration[%d]: %s", i, config.ScanType)
		}

		// Check if scan type was requested
		if !requestedTypes[strings.ToLower(config.ScanType)] {
			return fmt.Errorf("configuration provided for unrequested scan type: %s", config.ScanType)
		}

		// Check for duplicate configurations
		lowerType := strings.ToLower(config.ScanType)
		if configTypes[lowerType] {
			return fmt.Errorf("duplicate configuration for scan type: %s", config.ScanType)
		}
		configTypes[lowerType] = true

		// Validate configuration values based on scan type
		if err := sh.validateConfigurationValues(config); err != nil {
			return fmt.Errorf("invalid configuration for %s: %v", config.ScanType, err)
		}
	}

	return nil
}

// validateConfigurationValues validates configuration values for specific scan types
func (sh *ScanHandler) validateConfigurationValues(config cx1.ScanConfiguration) error {
	switch strings.ToLower(config.ScanType) {
	case "sast":
		// Validate SAST-specific configurations
		if preset, exists := config.Values["presetName"]; exists && preset == "" {
			return fmt.Errorf("presetName cannot be empty")
		}
		if incremental, exists := config.Values["incremental"]; exists {
			if incremental != "true" && incremental != "false" {
				return fmt.Errorf("incremental must be 'true' or 'false', got: %s", incremental)
			}
		}
		if verbose, exists := config.Values["engineVerbose"]; exists {
			if verbose != "true" && verbose != "false" {
				return fmt.Errorf("engineVerbose must be 'true' or 'false', got: %s", verbose)
			}
		}

	case "sca":
		// Validate SCA-specific configurations
		if exploitable, exists := config.Values["exploitablePath"]; exists {
			if exploitable != "true" && exploitable != "false" {
				return fmt.Errorf("exploitablePath must be 'true' or 'false', got: %s", exploitable)
			}
		}

	// Add validation for other scan types as needed
	case "secrets", "kics", "containersec", "apisec":
		// These typically don't require specific validation for now
		break
	}

	return nil
}

// generateDefaultConfigurations creates default configurations for scan types
func (sh *ScanHandler) generateDefaultConfigurations(scanTypes []string) []cx1.ScanConfiguration {
	var configurations []cx1.ScanConfiguration

	for _, scanType := range scanTypes {
		switch strings.ToLower(scanType) {
		case "sast":
			configurations = append(configurations, cx1.ScanConfiguration{
				ScanType: "sast",
				Values: map[string]string{
					"incremental":   "false",
					"presetName":    "Checkmarx Default",
					"engineVerbose": "false",
				},
			})

		case "sca":
			configurations = append(configurations, cx1.ScanConfiguration{
				ScanType: "sca",
				Values: map[string]string{
					"lastSastScanTime": "",
					"exploitablePath":  "false",
				},
			})

		case "secrets":
			configurations = append(configurations, cx1.ScanConfiguration{
				ScanType: "secrets",
				Values:   map[string]string{},
			})

		case "kics":
			configurations = append(configurations, cx1.ScanConfiguration{
				ScanType: "kics",
				Values:   map[string]string{},
			})

		case "containersec":
			configurations = append(configurations, cx1.ScanConfiguration{
				ScanType: "container",
				Values:   map[string]string{},
			})

		case "apisec":
			configurations = append(configurations, cx1.ScanConfiguration{
				ScanType: "apisec",
				Values:   map[string]string{},
			})
		}
	}

	return configurations
}

// Helper method to parse and validate scan types (unchanged)
func (sh *ScanHandler) parseScanTypes(scanTypesStr string) ([]string, error) {
	var scanTypes []string

	if scanTypesStr != "" {
		scanTypes = strings.Split(strings.ReplaceAll(scanTypesStr, " ", ""), ",")
	} else {
		scanTypes = []string{"sast", "sca"} // default
	}

	// Validate scan types
	validScanTypes := map[string]bool{
		"sast":         true,
		"sca":          true,
		"microengines": true,
		"kics":         true,
		"containers":   true,
		"apisec":       true,
	}

	for _, scanType := range scanTypes {
		if !validScanTypes[strings.ToLower(scanType)] {
			return nil, fmt.Errorf("invalid scan type: %s. Valid types: sast,sca,secrets,kics,containersec,apisec", scanType)
		}
	}

	return scanTypes, nil
}

// Helper method to parse tags from string format (unchanged)
func (sh *ScanHandler) parseTags(tagsStr string) (map[string]string, error) {
	tags := make(map[string]string)

	if tagsStr == "" {
		return tags, nil
	}

	tagPairs := strings.Split(tagsStr, ",")
	for _, pair := range tagPairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		kv := strings.SplitN(pair, ":", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid tag format: %s. Expected format: key:value", pair)
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		if key == "" {
			return nil, fmt.Errorf("empty tag key in: %s", pair)
		}

		tags[key] = value
	}

	return tags, nil
}

// Helper method to validate zip file (unchanged)
func (sh *ScanHandler) isValidZipFile(filename string) bool {
	return strings.HasSuffix(strings.ToLower(filename), ".zip")
}

func (sh *ScanHandler) getPreset(c *gin.Context) {
	response := gin.H{
		"presets": []string{"K-Web", "K-API", "K-Mobile"},
	}

	c.JSON(http.StatusOK, response)
}

// func (sh *ScanHandler) getTempConfig(c *gin.Context) {
// 	projectID := c.Query("project_id")

// 	data, err := sh.service.TempGetConfig(projectID)
// 	if err != nil {
// 		// Determine appropriate HTTP status code based on error type
// 		statusCode := http.StatusInternalServerError
// 		if err.Error() == "scan not found" || err.Error() == "no scans found for commit_id" {
// 			statusCode = http.StatusNotFound
// 		}

// 		c.JSON(statusCode, ErrorResponse{
// 			Error:     "Failed to get scan status",
// 			Details:   err.Error(),
// 			Timestamp: time.Now().Format(time.RFC3339),
// 			Path:      c.Request.URL.Path,
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, data)
// }
