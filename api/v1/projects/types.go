package projects

// // Request Models
// type CreateProjectRequest struct {
// 	Name        string            `json:"name" binding:"required"`
// 	Groups      []string          `json:"groups,omitempty"`
// 	Tags        map[string]string `json:"tags,omitempty"`
// 	Criticality int               `json:"criticality,omitempty"`
// 	Origin      string            `json:"origin,omitempty"`
// }

// type UpdateProjectRequest struct {
// 	Name        *string            `json:"name,omitempty"`
// 	Groups      *[]string          `json:"groups,omitempty"`
// 	Tags        *map[string]string `json:"tags,omitempty"`
// 	Criticality *int               `json:"criticality,omitempty"`
// }

// // Response Models
// type ProjectResponse struct {
// 	Project cx1.Project `json:"project"`
// 	Message string      `json:"message"`
// }

// type ProjectListResponse struct {
// 	Projects []cx1.Project `json:"projects"`
// 	Total    int           `json:"total"`
// 	Limit    int           `json:"limit"`
// 	Offset   int           `json:"offset"`
// }

// type ProjectScansResponse struct {
// 	Scans     []cx1.Scan `json:"scans"`
// 	Total     int        `json:"total"`
// 	Limit     int        `json:"limit"`
// 	Offset    int        `json:"offset"`
// 	ProjectID string     `json:"project_id"`
// }

// // Error Response
// type ErrorResponse struct {
// 	Error   string `json:"error"`
// 	Details string `json:"details,omitempty"`
// }