package projects

// type ProjectService struct {
// 	cx1Client *cx1.Cx1Client
// 	logger    util.Logger
// }

// func NewProjectService(client *cx1.Cx1Client, logger util.Logger) *ProjectService {
// 	return &ProjectService{
// 		cx1Client: client,
// 		logger:    logger,
// 	}
// }

// // ListProjects retrieves projects with pagination and optional name filtering
// func (ps *ProjectService) ListProjects(limit, offset int, nameFilter string) (*ProjectListResponse, error) {
// 	ps.logger.Debugf("Listing projects: limit=%d, offset=%d, nameFilter=%s", limit, offset, nameFilter)

// 	// Get projects from Checkmarx
// 	projects, err := ps.cx1Client.GetProjects(uint64(limit))
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get projects: %w", err)
// 	}

// 	// Filter by name if provided
// 	if nameFilter != "" {
// 		filteredProjects := make([]cx1.Project, 0)
// 		for _, project := range projects {
// 			if strings.Contains(strings.ToLower(project.Name), strings.ToLower(nameFilter)) {
// 				filteredProjects = append(filteredProjects, project)
// 			}
// 		}
// 		projects = filteredProjects
// 	}

// 	// Apply pagination
// 	totalCount := len(projects)
// 	start := offset
// 	end := offset + limit

// 	if start > totalCount {
// 		start = totalCount
// 	}
// 	if end > totalCount {
// 		end = totalCount
// 	}

// 	paginatedProjects := projects[start:end]

// 	return &ProjectListResponse{
// 		Projects: paginatedProjects,
// 		Total:    totalCount,
// 		Limit:    limit,
// 		Offset:   offset,
// 	}, nil
// }

// // CreateProject creates a new project
// func (ps *ProjectService) CreateProject(req *CreateProjectRequest) (*ProjectResponse, error) {
// 	ps.logger.Debugf("Creating project: %s", req.Name)

// 	if req.Name == "" {
// 		return nil, fmt.Errorf("project name is required")
// 	}

// 	// Set default values if not provided
// 	groups := req.Groups
// 	if groups == nil {
// 		groups = []string{}
// 	}

// 	tags := req.Tags
// 	if tags == nil {
// 		tags = make(map[string]string)
// 	}

// 	// Create project in Checkmarx
// 	project, err := ps.cx1Client.CreateProject(req.Name, groups, tags)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create project: %w", err)
// 	}

// 	return &ProjectResponse{
// 		Project: project,
// 		Message: "Project created successfully",
// 	}, nil
// }

// // GetProjectByID retrieves a specific project by ID
// func (ps *ProjectService) GetProjectByID(projectID string) (*ProjectResponse, error) {
// 	ps.logger.Debugf("Getting project by ID: %s", projectID)

// 	if projectID == "" {
// 		return nil, fmt.Errorf("project ID is required")
// 	}

// 	project, err := ps.cx1Client.GetProjectByID(projectID)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get project: %w", err)
// 	}

// 	return &ProjectResponse{
// 		Project: project,
// 		Message: "Project retrieved successfully",
// 	}, nil
// }

// // UpdateProject updates an existing project
// func (ps *ProjectService) UpdateProject(projectID string, req *UpdateProjectRequest) (*ProjectResponse, error) {
// 	ps.logger.Debugf("Updating project: %s", projectID)

// 	if projectID == "" {
// 		return nil, fmt.Errorf("project ID is required")
// 	}

// 	// Get existing project first
// 	existingProject, err := ps.cx1Client.GetProjectByID(projectID)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get existing project: %w", err)
// 	}

// 	// Update project fields if provided
// 	if req.Name != nil {
// 		existingProject.Name = *req.Name
// 	}

// 	if req.Tags != nil {
// 		existingProject.Tags = *req.Tags
// 	}

// 	if req.Groups != nil {
// 		existingProject.Groups = *req.Groups
// 	}

// 	// Update project in Checkmarx
// 	err = ps.cx1Client.UpdateProject(&existingProject)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to update project: %w", err)
// 	}

// 	// Get the updated project to return the latest state
// 	updatedProject, err := ps.cx1Client.GetProjectByID(projectID)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get updated project: %w", err)
// 	}

// 	return &ProjectResponse{
// 		Project: updatedProject,
// 		Message: "Project updated successfully",
// 	}, nil
// }

// // DeleteProject deletes a project by ID
// func (ps *ProjectService) DeleteProject(projectID string) error {
// 	ps.logger.Debugf("Deleting project: %s", projectID)

// 	if projectID == "" {
// 		return fmt.Errorf("project ID is required")
// 	}

// 	// Check if project exists first
// 	_, err := ps.cx1Client.GetProjectByID(projectID)
// 	if err != nil {
// 		return fmt.Errorf("project not found: %w", err)
// 	}

// 	project := &cx1.Project{
// 		ProjectID: projectID,
// 	}
// 	// Delete project
// 	err = ps.cx1Client.DeleteProject(project)
// 	if err != nil {
// 		return fmt.Errorf("failed to delete project: %w", err)
// 	}

// 	ps.logger.Infof("Project %s deleted successfully", projectID)
// 	return nil
// }

// // GetProjectScans retrieves scans for a specific project
// func (ps *ProjectService) GetProjectLastScan(projectID string, limit, offset int) (*ProjectScansResponse, error) {
// 	ps.logger.Debugf("Getting scans for project: %s", projectID)

// 	if projectID == "" {
// 		return nil, fmt.Errorf("project ID is required")
// 	}

// 	// Check if project exists
// 	_, err := ps.cx1Client.GetProjectByID(projectID)
// 	if err != nil {
// 		return nil, fmt.Errorf("project not found: %w", err)
// 	}

// 	// Get scans for the project
// 	scans, err := ps.cx1Client.GetLastScansByID(projectID, 10)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get project scans: %w", err)
// 	}

// 	// Apply pagination
// 	totalCount := len(scans)
// 	start := offset
// 	end := offset + limit

// 	if start > totalCount {
// 		start = totalCount
// 	}
// 	if end > totalCount {
// 		end = totalCount
// 	}

// 	paginatedScans := scans[start:end]

// 	return &ProjectScansResponse{
// 		Scans:     paginatedScans,
// 		Total:     totalCount,
// 		Limit:     limit,
// 		Offset:    offset,
// 		ProjectID: projectID,
// 	}, nil
// }