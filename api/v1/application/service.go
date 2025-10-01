package application

import (
	"fmt"

	cx1 "github.com/madhatkul/CxWrapper-v2/Cx1ClientGo"

	"github.com/madhatkul/CxWrapper-v2/util"
)

type ApplicationService struct {
	cx1Client *cx1.Cx1Client
	logger    util.Logger
}

func NewApplicationService(cx1Client *cx1.Cx1Client, logger util.Logger) *ApplicationService {
	return &ApplicationService{
		cx1Client: cx1Client,
		logger:    logger,
	}
}

func (s *ApplicationService) AssignProjectToApp(appName string, projectName string) error {
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