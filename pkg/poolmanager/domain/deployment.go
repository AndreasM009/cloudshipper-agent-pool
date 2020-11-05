package domain

import "time"

// DeploymentState state of a deployment
type DeploymentState int

const (
	// Queued job is queued and ready to run
	Queued DeploymentState = iota
	// Running job is running
	Running
)

// Deployment a tenant deployment job
type Deployment struct {
	DeploymentID   string          `json:"deplyomentId"`
	DeploymentName string          `json:"deploymentName"`
	DefinitionID   string          `json:"definitionId"`
	TenantID       string          `json:"tenantId"`
	Timestamp      time.Time       `json:"timestamp"`
	LiveStream     string          `json:"liveStream"`
	State          DeploymentState `json:"state"`
}
