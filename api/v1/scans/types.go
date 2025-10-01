package scans

import (
	"io"
	"time"

	cx1 "github.com/madhatkul/CxWrapper-v2/Cx1ClientGo"
)

// Base scan request structure
type ScanRequest struct {
	Branch      string `json:"branch" binding:"required"`
	CommitID    string `json:"commit_id" binding:"required"`
	ProjectName string `json:"project_name" binding:"required"`
}

// Form request for static scans (matches OpenAPI spec)
type StaticScanFormRequest struct {
	ProjectName string `form:"project_name" binding:"required"`
	ScanTypes   string `form:"scan_types"` // comma-separated: sast,sca,secrets,kics,containersec,apisec
	Branch      string `form:"branch" binding:"required"`
	// Config      string `form:"config"`
	IsFastScan string `form:"is_fast_scan"`
	CommitID   string `form:"commit_id" binding:"required"`
	Tags       string `form:"tags"`
	// ZipFile is handled by multipart form, not included in struct
}

// Internal request structure with file contents
type StaticScanRequestWithFile struct {
	AppName     string
	ProjectName string
	Branch      string
	CommitID    string
	ScanTypes   []string
	IsFastScan  bool
	Preset      string
	Tags        map[string]string
	// FileContents []byte
	File     io.Reader
	FileSize int64
	FileName string
}

// Response structures for API
type ScanResponse struct {
	ScanID  string `json:"scan_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error     string `json:"error"`
	Details   string `json:"details,omitempty"`
	Timestamp string `json:"timestamp"`
	Path      string `json:"path"`
}

type ScanLogsResponse struct {
	ScanID string   `json:"scan_id"`
	Logs   []string `json:"logs"`
}

// Health check response
type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ScanListRequest represents the request parameters for listing scans
type ScanListRequest struct {
	ProjectID   string     `json:"project_id,omitempty"`
	ProjectName string     `json:"project_name,omitempty"`
	ScanType    string     `json:"scan_type,omitempty"`
	Status      []string   `json:"status,omitempty"`
	Limit       int        `json:"limit,omitempty"`
	Offset      int        `json:"offset,omitempty"`
	FromDate    *time.Time `json:"from_date,omitempty"`
	ToDate      *time.Time `json:"to_date,omitempty"`
}

// ScanListResponse represents the response for listing scans
type ScanListResponse struct {
	Scans  interface{} `json:"scans"` // Can be []cx1.Scan or []ScanStatusInternal
	Total  int         `json:"total"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
}

type ListScansRequest struct {
	ProjectName string `json:"project_name,omitempty" form:"project_name"`
	CommitID    string `json:"commit_id,omitempty" form:"commit_id"`
	Limit       int    `json:"limit,omitempty" form:"limit"`
	Offset      int    `json:"offset,omitempty" form:"offset"`
}

type ListScansResponse struct {
	Scans  []cx1.Scan `json:"scans"`
	Total  int        `json:"total"`
	Limit  int        `json:"limit"`
	Offset int        `json:"offset"`
}

type SimpleScanStatus struct {
	ScanID string `json:"scan_id"`
	Status string `json:"status"`
}

// WebhookPayload represents the data sent to the webhook
type WebhookPayload struct {
	ScanID    string                 `json:"scan_id"`
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// WebhookConfig holds webhook configuration
type WebhookConfig struct {
	URL     string
	Timeout time.Duration
	Headers map[string]string
}
type AllScansResponse struct {
	CommitID    string           `json:"commit_id"`
	ProjectName string           `json:"project_name,omitempty"`
	TotalScans  int              `json:"total_scans"`
	Summary     AllScansSummary  `json:"summary"`
	Scans       CategorizedScans `json:"scans"`
}

type CategorizedScans struct {
	Fast *ScanResultResponse `json:"fast,omitempty"`
	Full *ScanResultResponse `json:"full,omitempty"`
}

type AllScansSummary struct {
	CompletedScans  int `json:"completed_scans"`
	BreakBuildCount int `json:"break_build_count"`
}

// Update ScanResultResponse to include optional fields
type ScanResultResponse struct {
	Link          string      `json:"link"`
	IsFastScan    bool        `json:"is_fast_scan"`
	BreakBuild    bool        `json:"is_policy_blocked"`
	ScanID        string      `json:"scan_id"`
	CommitID      string      `json:"commit_id"`
	ProjectID     string      `json:"project_id"`
	Branch        string      `json:"branch"`
	Status        string      `json:"status"`
	CreatedAt     string      `json:"created_at"`
	UpdatedAt     string      `json:"updated_at"`
	Tags          interface{} `json:"tags"`
	Results       interface{} `json:"results"`
	Summary       Summary     `json:"summary"`
	PolicyWarning *string     `json:"policy_warning,omitempty"`
	Error         *string     `json:"error,omitempty"`
	StatusMessage *string     `json:"status_message,omitempty"`
}

// Summary represents the summary section of the scan response
type Summary struct {
	TotalResults int `json:"total_results"`
}
