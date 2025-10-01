// api/v1/scans/service.go
package scans

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	cx1 "github.com/madhatkul/CxWrapper-v2/Cx1ClientGo"
	"github.com/madhatkul/CxWrapper-v2/util"
)

type ScanService struct {
	cx1Client *cx1.Cx1Client
	logger    util.Logger
}

func NewScanService(client *cx1.Cx1Client, logger util.Logger) *ScanService {
	return &ScanService{
		cx1Client: client,
		logger:    logger,
	}
}

// Updated StartStaticScanWithFile to use client-provided configurations
func (ss *ScanService) StartStaticScanWithFile(req StaticScanRequestWithFile) (*cx1.Scan, error) {
	ss.logger.Infof("Starting static scan for project: %s on branch: %s", req.ProjectName, req.Branch)

	// Validate input
	if req.AppName == "" {
		return nil, fmt.Errorf("application name is required")
	}
	if req.ProjectName == "" {
		return nil, fmt.Errorf("project name is required")
	}
	if req.Branch == "" {
		return nil, fmt.Errorf("branch is required")
	}
	if req.CommitID == "" {
		return nil, fmt.Errorf("commit ID is required")
	}
	if req.FileSize == 0 {
		return nil, fmt.Errorf("file contents are empty")
	}

	if req.Preset == "" {
		return nil, fmt.Errorf("preset is required")
	}

	var project cx1.Project

	// Get project by name
	projects, err := ss.cx1Client.GetProjectsByName(req.ProjectName)
	if err != nil {
		return nil, fmt.Errorf("failed to get project '%s': %v", req.ProjectName, err)
	}
	if len(projects) == 0 {
		ss.logger.Infof("Project '%s' not found, creating a new one.", req.ProjectName)

		newProject, err := ss.cx1Client.CreateProject(req.ProjectName, []string{}, make(map[string]string))
		if err != nil {
			return nil, fmt.Errorf("failed to create new project '%s': %v", req.ProjectName, err)
		}
		project = newProject
		ss.logger.Infof("✅ New project created with ID: %s", project.ProjectID)
	} else {
		project = projects[0] // ใช้โปรเจกต์แรกที่เจอ
		ss.logger.Infof("✅ Project found with ID: %s", project.ProjectID)
	}

	projectID := project.ProjectID

	err = ss.AssignProjectToApp(req.AppName, project.Name)
	if err != nil {
		ss.logger.Errorf("Failed to assign project to application: %v", err)
		return nil, fmt.Errorf("failed to assign project to application: %v", err)
	}

	// Upload file contents
	uploadURL, err := ss.cx1Client.UploadStreamForProjectByID(projectID, req.File, req.FileSize)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to project %s: %v", projectID, err)
	}

	ss.logger.Infof("✅ File uploaded successfully, URL: %s File Size: %s", uploadURL, (req.FileSize))

	// Convert client configurations to cx1.ScanConfigurationSet
	defaultSettings, err := ss.cx1Client.GetScanConfigurationByProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get default scan configuration for project %s: %v", projectID, err)
	}
	ss.logger.Infof("✅ Retrieved %d default configuration settings for project", len(defaultSettings))

	// 2. Filter configurations based on the requested scan types
	configMap := make(map[string]map[string]string)
	requiredCategories := make(map[string]bool)
	for _, scanType := range req.ScanTypes {
		requiredCategories[scanType] = true
	}

	for _, setting := range defaultSettings {
		if requiredCategories[setting.Category] {
			if _, ok := configMap[setting.Category]; !ok {
				configMap[setting.Category] = make(map[string]string)
			}

			if setting.Value != "" {
				configMap[setting.Category][setting.Name] = setting.Value
			}
		}
	}

	ss.logger.Infof("configMap: %v", configMap)

	for _, scanType := range req.ScanTypes {
		if scanType == "microengines" {
			configMap["microengines"] = map[string]string{
				"scorecard": "true",
				"2ms":       "true",
			}
		}
	}

	// 3. If is_fast_scan is true, override the SAST configuration
	if req.IsFastScan {
		ss.logger.Infof("⚡ Fast scan requested. Overriding SAST configuration.")
		if _, ok := configMap["sast"]; !ok {
			configMap["sast"] = make(map[string]string)
		}
		// The key "fast scan mode" corresponds to the 'name' field from the get configuration response
		configMap["sast"]["fastScanMode"] = "true"
	}

	if req.Preset != "" {
		ss.logger.Infof("🎯 Applying preset: %s", req.Preset)

		if _, ok := configMap["sast"]; !ok {
			configMap["sast"] = make(map[string]string)
		}
		configMap["sast"]["presetName"] = req.Preset

		err := ss.cx1Client.UpdateProjectConfigurationByID(projectID, []cx1.ConfigurationSetting{
			{ // Added the struct type here
				Key:             "scan.config.sast.presetName",
				Name:            "presetName",
				Category:        "sast",
				OriginLevel:     "Project",
				Value:           req.Preset,
				ValueType:       "RESTList",
				ValueTypeParams: "{\"path\":\"/queries/presets\",\"fieldMap\":{\"id\":\"id\",\"value\":\"name\",\"label\":\"name\"}}",
				AllowOverride:   true,
			},
		})
		if err != nil {
			ss.logger.Errorf("Failed to update project configuration: %v", err)
		}
	}
	var finalScanConfigurations []cx1.ScanConfiguration
	for category, values := range configMap {
		finalScanConfigurations = append(finalScanConfigurations, cx1.ScanConfiguration{
			ScanType: category,
			Values:   values,
		})
	}

	configJSON, _ := json.Marshal(finalScanConfigurations)
	ss.logger.Infof("Prepared %d scan configurations. Details: %s", len(finalScanConfigurations), string(configJSON))

	// Prepare tags
	tags := req.Tags
	if tags == nil {
		tags = make(map[string]string)
	}

	tags["commit_id"] = req.CommitID

	// Trigger scan
	scan, err := ss.cx1Client.ScanProjectZipByID(projectID, uploadURL, req.Branch, finalScanConfigurations, tags)
	if err != nil {
		return nil, fmt.Errorf("failed to trigger scan for project %s: %v", projectID, err)
	}

	// Polling
	go ss.PollingStatus(&scan)

	ss.logger.Infof("✅ Scan triggered successfully with ID: %s for project ID: %s", scan.ScanID, projectID)

	return &scan, nil
}

func (ss *ScanService) PollingStatus(scan *cx1.Scan) {

	ss.logger.Infof("🔄 Polling status for scan ID: %s", scan.ScanID)

	ss.logger.Infof("🔄 Starting scan polling process")

	updatedScan, err := ss.cx1Client.ScanPolling(scan)
	if err != nil {
		ss.logger.Errorf("❌ Error during scan polling: %v", err)
		return
	}

	ss.logger.Infof("✅ Scan polling completed successfully for scan ID: %s with status: %s", updatedScan.ScanID, updatedScan.Status)

	response, err := ss.GetScanResultsByScanID(updatedScan.ScanID)
	if err != nil {
		ss.logger.Errorf("❌ Error getting scan results for scan ID %s: %v", updatedScan.ScanID, err)
		return
	}

	ss.logger.Infof("✅ Scan results retrieved successfully for scan ID: %s", updatedScan.ScanID)

	// Log the actual response content (be careful with size)
	if response != nil {
		ss.logger.Debugf("📋 Full scan response details: %+v", response)
	}

	// Webhook section (currently commented out)
	webhookURL := os.Getenv("STATIC_WEBHOOK_URL")
	if err := ss.sendWebhook(webhookURL, &updatedScan); err != nil {
		ss.logger.Errorf("❌ Failed to send webhook: %v", err)
	} else {
		ss.logger.Infof("✅ Webhook sent successfully")
	}
}

func (ss *ScanService) sendWebhook(webhookURL string, scan *cx1.Scan) (err error) {
	// Create payload
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("a panic occurred while sending webhook: %v", r)
		}
	}()

	// Construct the ScanResultResponse, similar to GetAllScanResultsByCommitID but for a single scan.
	resultsLink := fmt.Sprintf("https://sng.ast.checkmarx.net/projects/%s/scans?branch=%s&id=%s",
		scan.ProjectID, scan.Branch, scan.ScanID)

	scanResponse := &ScanResultResponse{
		Link:      resultsLink,
		ScanID:    scan.ScanID,
		CommitID:  scan.Tags["commit_id"], // Assumes commit_id is in tags
		ProjectID: scan.ProjectID,
		Branch:    scan.Branch,
		Status:    scan.Status,
		CreatedAt: scan.CreatedAt,
		UpdatedAt: scan.UpdatedAt,
		Tags:      scan.Tags,
	}

	if scan.Status == "Completed" {
		results, err := ss.cx1Client.GetAllScanResultsByID(scan.ScanID)
		if err != nil {
			ss.logger.Errorf("Failed to get results for scan ID %s: %v", scan.ScanID, err)
			errorMsg := fmt.Sprintf("Failed to get results: %v", err)
			scanResponse.Error = &errorMsg
		} else {
			scanResponse.Results = results
			scanResponse.Summary = Summary{TotalResults: int(results.Count())}
			ss.logger.Debugf("Retrieved %d results for scan ID %s", results.Count(), scan.ScanID)
		}

		config, err := ss.cx1Client.GetScanConfigurationByID(scan.ProjectID, scan.ScanID)
		if err != nil {
			ss.logger.Warnf("Failed to get scan configuration for scan ID %s: %v. Assuming full scan.", scan.ScanID, err)
			scanResponse.IsFastScan = false
		} else {
			scanResponse.IsFastScan = ss.isFastScanMode(config)
		}

		breakbuild, policyErr := ss.cx1Client.RetrievePolicyViolationInfo(scan.ProjectID, scan.ScanID)
		if policyErr != nil {
			ss.logger.Warnf("Failed to retrieve policy violation info for scan ID %s: %v. Continuing without breakbuild status.", scan.ScanID, policyErr)
			warning := fmt.Sprintf("Policy violation info unavailable: %v", policyErr)
			scanResponse.PolicyWarning = &warning
			scanResponse.BreakBuild = false
		} else {
			scanResponse.BreakBuild = breakbuild
		}
	} else {
		scanResponse.IsFastScan = false
		scanResponse.BreakBuild = false
		statusMsg := fmt.Sprintf("Scan not completed (status: %s)", scan.Status)
		scanResponse.StatusMessage = &statusMsg
	}

	// Marshal payload to JSON
	jsonData, err := json.Marshal(scanResponse)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	// Create HTTP request
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "CX1-ScanService/1.0")

	// Send request
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			// This single line tells the client to use HTTP_PROXY, HTTPS_PROXY,
			// and NO_PROXY from your environment variables.
			Proxy: http.ProxyFromEnvironment,

			// A default TLSClientConfig is secure and enforces certificate validation.
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-success status: %d", resp.StatusCode)
	}

	return nil
}

func (ss *ScanService) GetScanResultsByScanID(scanID string) (interface{}, error) {

	// Get filtered scans
	scan, err := ss.cx1Client.GetScanByID(scanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get scans: %v", err)
	}

	ss.logger.Debugf("Found scan ID: %s with status: %s", scan.ScanID, scan.Status)

	// Check if scan is completed
	if scan.Status != "Completed" {
		return nil, fmt.Errorf("scan is not completed yet (current status: %s). Scan ID: %s", scan.Status, scan.ScanID)
	}

	// Get detailed results using the actual scan ID
	results, err := ss.cx1Client.GetAllScanResultsByID(scan.ScanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get results for scan %s: %v", scan.ScanID, err)
	}

	ss.logger.Infof("Successfully retrieved scan results for scan ID: %s with %d results", scan.ScanID, results.Count())

	// Create a structured response with scan metadata and results
	resultsLink := fmt.Sprintf("https://sng.ast.checkmarx.net/projects/%s/scans?branch=%s&id=%s",
		scan.ProjectID, scan.Branch, scan.ScanID)

	var breakbuild bool
	var policyErr error

	breakbuild, policyErr = ss.cx1Client.RetrievePolicyViolationInfo(scan.ProjectID, scan.ScanID)
	if policyErr != nil {

		ss.logger.Warnf("Failed to retrieve policy violation info for scan ID %s: %v. Continuing without breakbuild status.", scan.ScanID, policyErr)

		// Set breakbuild to false as default when policy info is unavailable
		breakbuild = true
	}

	response := &ScanResultResponse{
		Link:       resultsLink,
		BreakBuild: breakbuild,
		ScanID:     scan.ScanID,
		ProjectID:  scan.ProjectID,
		Branch:     scan.Branch,
		Status:     scan.Status,
		CreatedAt:  scan.CreatedAt,
		UpdatedAt:  scan.UpdatedAt,
		Tags:       scan.Tags,
		Results:    results,
		Summary: Summary{
			TotalResults: int(results.Count()),
		},
	}

	if policyErr != nil {
		warning := fmt.Sprintf("Policy violation info unavailable: %v", policyErr)
		response.PolicyWarning = &warning
	}

	return response, nil
}

func (ss *ScanService) GetAllScanResultsByCommitID(commitID string, projectName string) (interface{}, error) {
	// Create filter to find scans by commit_id
	filter := cx1.ScanFilter{}

	// Add commit_id filter
	filter.TagKeys = append(filter.TagKeys, "commit_id")
	filter.TagValues = append(filter.TagValues, commitID)

	// Optionally filter by project if provided
	if projectName != "" {
		projects, err := ss.cx1Client.GetProjectsByName(projectName)
		if err != nil {
			return nil, fmt.Errorf("failed to find project '%s': %v", projectName, err)
		}
		if len(projects) == 0 {
			return nil, fmt.Errorf("project '%s' not found", projectName)
		}
		filter.ProjectID = projects[0].ProjectID
		ss.logger.Debugf("Filtering by project '%s' (ID: %s) and commit_id '%s' for results", projectName, filter.ProjectID, commitID)
	} else {
		ss.logger.Debugf("Filtering by commit_id '%s' only for results", commitID)
	}

	// Get filtered scans (assuming they are sorted newest to oldest)
	scans, err := ss.cx1Client.GetLastScansFiltered(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get scans: %v", err)
	}

	if len(scans) == 0 {
		return nil, fmt.Errorf("no scans found for commit_id: %s", commitID)
	}

	// Pointers to hold the latest fast and full scans found.
	var latestFastScan *ScanResultResponse
	var latestFullScan *ScanResultResponse

	// Process scans to find the latest of each type
	for _, scan := range scans {
		// Optimization: if we've found both, we can stop processing.
		if latestFastScan != nil && latestFullScan != nil {
			break
		}

		ss.logger.Debugf("Processing scan ID: %s with status: %s for commit_id: %s", scan.ScanID, scan.Status, commitID)

		// Initialize scan response
		resultsLink := fmt.Sprintf("https://sng.ast.checkmarx.net/projects/%s/scans?branch=%s&id=%s",
			scan.ProjectID, scan.Branch, scan.ScanID)

		scanResponse := ScanResultResponse{
			Link:      resultsLink,
			ScanID:    scan.ScanID,
			CommitID:  commitID,
			ProjectID: scan.ProjectID,
			Branch:    scan.Branch,
			Status:    scan.Status,
			CreatedAt: scan.CreatedAt,
			UpdatedAt: scan.UpdatedAt,
			Tags:      scan.Tags,
		}

		// Only get detailed results for completed scans
		if scan.Status == "Completed" {
			results, err := ss.cx1Client.GetAllScanResultsByID(scan.ScanID)
			if err != nil {
				ss.logger.Errorf("Failed to get results for scan ID %s: %v", scan.ScanID, err)
				errorMsg := fmt.Sprintf("Failed to get results: %v", err)
				scanResponse.Error = &errorMsg
			} else {
				scanResponse.Results = results
				scanResponse.Summary = Summary{TotalResults: int(results.Count())}
				ss.logger.Debugf("Retrieved %d results for scan ID %s", results.Count(), scan.ScanID)
			}

			// Determine if the scan is a fast scan
			config, err := ss.cx1Client.GetScanConfigurationByID(scan.ProjectID, scan.ScanID)
			if err != nil {
				ss.logger.Warnf("Failed to get scan configuration for scan ID %s: %v. Assuming full scan.", scan.ScanID, err)
				scanResponse.IsFastScan = false
			} else {
				scanResponse.IsFastScan = ss.isFastScanMode(config)
			}

			ss.logger.Infof("Is fast scan: ", config)

			// Get policy violation info
			breakbuild, policyErr := ss.cx1Client.RetrievePolicyViolationInfo(scan.ProjectID, scan.ScanID)
			if policyErr != nil {
				ss.logger.Warnf("Failed to retrieve policy violation info for scan ID %s: %v. Continuing without breakbuild status.", scan.ScanID, policyErr)
				warning := fmt.Sprintf("Policy violation info unavailable: %v", policyErr)
				scanResponse.PolicyWarning = &warning
				scanResponse.BreakBuild = false
			} else {
				scanResponse.BreakBuild = breakbuild
			}
		} else {
			// For non-completed scans, set defaults
			scanResponse.IsFastScan = false
			scanResponse.BreakBuild = false
			statusMsg := fmt.Sprintf("Scan not completed (status: %s)", scan.Status)
			scanResponse.StatusMessage = &statusMsg
		}

		// Assign the processed scan to the correct category if it's the first one we've found
		if scanResponse.IsFastScan {
			if latestFastScan == nil {
				// Create a new variable to hold the response and assign its address
				s := scanResponse
				latestFastScan = &s
			}
		} else { // It's a full scan
			if latestFullScan == nil {
				s := scanResponse
				latestFullScan = &s
			}
		}
	}

	// Create the final response object with the categorized scans
	response := &AllScansResponse{
		CommitID:    commitID,
		ProjectName: projectName,
		Scans: CategorizedScans{
			Fast: latestFastScan,
			Full: latestFullScan,
		},
	}

	// Add summary counts based on the found scans
	completedScans := 0
	breakBuildCount := 0
	totalScansInResponse := 0

	if latestFastScan != nil {
		totalScansInResponse++
		if latestFastScan.Status == "Completed" {
			completedScans++
			if latestFastScan.BreakBuild {
				breakBuildCount++
			}
		}
	}
	if latestFullScan != nil {
		totalScansInResponse++
		if latestFullScan.Status == "Completed" {
			completedScans++
			if latestFullScan.BreakBuild {
				breakBuildCount++
			}
		}
	}

	response.TotalScans = totalScansInResponse
	response.Summary = AllScansSummary{
		CompletedScans:  completedScans,
		BreakBuildCount: breakBuildCount,
	}

	ss.logger.Infof("✅ Successfully processed and categorized latest scans for commit_id: %s. Total scans found: %d, Completed: %d, BreakBuild: %d", commitID, totalScansInResponse, completedScans, breakBuildCount)

	return response, nil
}

func (ss *ScanService) isFastScanMode(config interface{}) bool {
	// The config is a slice of a specific struct type, not a generic interface.
	// We perform a type assertion to the correct type from the cx1 library.
	if configSlice, ok := config.([]cx1.ConfigurationSetting); ok {
		for _, setting := range configSlice {
			// Access the struct's 'Key' field directly.
			if setting.Key == "scan.config.sast.fastScanMode" {

				ss.logger.Infof("Found fastScanMode setting, checking value: %s (type: %T)", setting.Value, setting.Value)

				// Access the 'Value' field and check its type (string or bool).
				// setting.Value is of type string, so compare directly
				return setting.Value == "true"
			}
		}
	} else {
		// Add a log to show if the initial type assertion fails.
		ss.logger.Warnf("isFastScanMode received an unexpected config type: %T", config)
	}

	// Default to false if not found or if the type assertion fails.
	return false
}

func (ss *ScanService) ListScansFiltered(req ListScansRequest) (*ListScansResponse, error) {
	// Create ScanFilter based on the request
	filter := cx1.ScanFilter{}

	// Handle project_name - need to resolve to project_id
	if req.ProjectName != "" {
		projects, err := ss.cx1Client.GetProjectsByName(req.ProjectName)
		if err != nil {
			return nil, fmt.Errorf("failed to find project '%s': %v", req.ProjectName, err)
		}
		if len(projects) == 0 {
			return nil, fmt.Errorf("project '%s' not found", req.ProjectName)
		}
		// Use the first matching project
		filter.ProjectID = projects[0].ProjectID

		ss.logger.Debugf("Resolved project name '%s' to ID '%s'", req.ProjectName, filter.ProjectID)
	}

	// Handle commit_id - add it to tag values for filtering
	if req.CommitID != "" {
		filter.TagKeys = append(filter.TagKeys, "commit_id")
		filter.TagValues = append(filter.TagValues, req.CommitID)
		ss.logger.Debugf("Added commit_id filter: %s", req.CommitID)
	}

	// Get filtered scans using the cx1Client method
	scans, err := ss.cx1Client.GetLastScansFiltered(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get filtered scans: %v", err)
	}

	ss.logger.Debugf("Retrieved filtered scans, total count: %d", len(scans))

	// Apply client-side pagination since the API might not support it directly
	total := len(scans)
	start := req.Offset
	end := req.Offset + req.Limit

	// Validate pagination bounds
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	var paginatedScans []cx1.Scan
	if start < total {
		paginatedScans = scans[start:end]
	}

	return &ListScansResponse{
		Scans:  paginatedScans,
		Total:  total,
		Limit:  req.Limit,
		Offset: req.Offset,
	}, nil
}

// GetScanStatusByCommitID gets scan status by commit_id, optionally filtered by project_name
func (ss *ScanService) GetScanStatusByCommitID(commitID string, projectName string) (*SimpleScanStatus, error) {
	// Create filter to find scans by commit_id
	filter := cx1.ScanFilter{}

	// Add commit_id filter
	filter.TagKeys = append(filter.TagKeys, "commit_id")
	filter.TagValues = append(filter.TagValues, commitID)

	// Optionally filter by project if provided
	if projectName != "" {
		projects, err := ss.cx1Client.GetProjectsByName(projectName)
		if err != nil {
			return nil, fmt.Errorf("failed to find project '%s': %v", projectName, err)
		}
		if len(projects) == 0 {
			return nil, fmt.Errorf("project '%s' not found", projectName)
		}
		filter.ProjectID = projects[0].ProjectID

		ss.logger.Debugf("Filtering by project '%s' (ID: %s) and commit_id '%s'", projectName, filter.ProjectID, commitID)
	} else {
		ss.logger.Debugf("Filtering by commit_id '%s' only", commitID)
	}

	// Get filtered scans
	scans, err := ss.cx1Client.GetLastScansFiltered(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get scans: %v", err)
	}

	if len(scans) == 0 {
		return nil, fmt.Errorf("no scans found for commit_id: %s", commitID)
	}

	// Get the most recent scan (first one since GetLastScansFiltered sorts by created_at desc)
	scan := scans[0]

	return &SimpleScanStatus{
		ScanID: scan.ScanID,
		Status: scan.Status,
	}, nil
}

func (ss *ScanService) CancelScan(commitID string, projectName string) error {
	filter := cx1.ScanFilter{}

	// Add commit_id filter
	filter.TagKeys = append(filter.TagKeys, "commit_id")
	filter.TagValues = append(filter.TagValues, commitID)

	// Add project_name filter if provided
	if projectName != "" {
		filter.TagKeys = append(filter.TagKeys, "project_name")
		filter.TagValues = append(filter.TagValues, projectName)
	}

	scans, err := ss.cx1Client.GetLastScansFiltered(filter)
	if err != nil {
		return fmt.Errorf("failed to get scans: %v", err)
	}

	if len(scans) == 0 {
		if projectName != "" {
			return fmt.Errorf("no scans found for commit_id: %s and project_name: %s", commitID, projectName)
		}
		return fmt.Errorf("no scans found for commit_id: %s", commitID)
	}

	// Get the most recent scan (first one since GetLastScansFiltered sorts by created_at desc)
	scan := scans[0]

	err = ss.cx1Client.CancelScanByID(scan.ScanID)
	if err != nil {
		return fmt.Errorf("failed to cancel scan: %v", err)
	}

	return nil
}

// func (ss *ScanService) TempGetConfig(project_Id string) ([]cx1.ConfigurationSetting, error) {
//  config, err := ss.cx1Client.GetScanConfigurationByProjectID(project_Id)
//  if err != nil {
//      return nil, fmt.Errorf("failed to get config: %v", err)
//  }

//  return config, nil
// }

func (s *ScanService) AssignProjectToApp(appName string, projectName string) error {
	// Get application by name.
	application, err := s.cx1Client.GetApplicationByName(appName)
	if err != nil {
		s.logger.Infof("Application '%s' not found, creating a new one.", appName)
		// Create the application if it doesn't exist.
		newApplication, createErr := s.cx1Client.CreateApplication(appName)
		if createErr != nil {
			s.logger.Errorf("Error creating application '%s': %v", appName, createErr)
			return createErr
		}
		application = newApplication
	} else {
		s.logger.Infof("Application fetched: %v", application.Name)
	}

	// Get project by name.
	projects, err := s.cx1Client.GetProjectsByName(projectName)
	if err != nil {
		return fmt.Errorf("failed to get project '%s': %v", projectName, err)
	}

	var project cx1.Project
	if len(projects) == 0 {
		s.logger.Infof("Project '%s' not found, creating a new one.", projectName)
		// Create the project if it doesn't exist.
		newProject, createErr := s.cx1Client.CreateProject(projectName, []string{}, make(map[string]string))
		if createErr != nil {
			return fmt.Errorf("failed to create new project '%s': %v", projectName, createErr)
		}
		project = newProject
		s.logger.Infof("✅ New project created with ID: %s", project.ProjectID)
	} else {
		project = projects[0] // Use the first project found.
		s.logger.Infof("✅ Project found with ID: %s", project.ProjectID)
	}

	// Assign the project to the application.
	application.AssignProject(&project)

	// Update the application to save the changes.
	if err := s.cx1Client.UpdateApplication(&application); err != nil {
		s.logger.Errorf("Error updating application '%s': %v", appName, err)
		return err
	}

	s.logger.Infof("✅ Successfully assigned project '%s' to application '%s'", projectName, appName)
	return nil
}
