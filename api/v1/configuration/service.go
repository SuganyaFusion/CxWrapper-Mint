package configuration

// type ConfigurationService struct {
// 	cx1Client *cx1.Cx1Client
// }

// // NewConfigurationService creates a new configuration service instance
// func NewConfigurationService(client *cx1.Cx1Client) *ConfigurationService {
// 	return &ConfigurationService{
// 		cx1Client: client,
// 	}
// }

// func (s *ConfigurationService) GetProjectParameters(projectID string) ([]cx1.ConfigurationSetting, error) {
// 	response, err := s.cx1Client.GetProjectConfigurationByID(projectID)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get projects: %w", err)
// 	}
// 	return response, nil
// }

// func (s *ConfigurationService) SetProjectParameters(projectID string, req []cx1.ConfigurationSetting) error {
// 	err := s.cx1Client.UpdateProjectConfigurationByID(projectID, req)
// 	if err != nil {
// 		return fmt.Errorf("failed to update projects: %w", err)
// 	}
// 	return nil
// }

// func (s *ConfigurationService) DeleteProjectParameters(projectID string, configKey string) error {
// 	err := s.cx1Client.DeleteProjectConfiguration(projectID, configKey)
// 	if err != nil {
// 		return fmt.Errorf("failed to delete project configuration: %w", err)
// 	}
// 	return nil
// }

// func (s *ConfigurationService) GetScanParameters(projectID string, scanID string) ([]cx1.ConfigurationSetting, error) {
// 	response, err := s.cx1Client.GetScanConfigurationByID(projectID, scanID)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get scan configuration: %w", err)
// 	}
// 	return response, nil
// }