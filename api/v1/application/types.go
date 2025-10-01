package application

type AssignProjectRequest struct {
	AppName     string `json:"app_name" validate:"required"`
	ProjectName string `json:"project_name" validate:"required"`
}