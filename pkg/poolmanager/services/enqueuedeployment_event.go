package services

// EnqueueDeploymentEvent event for agent to run deployment
type EnqueueDeploymentEvent struct {
	TenantID       string            `json:"tenantId"`
	DeploymentName string            `json:"deploymentName"`
	ID             string            `json:"id"`
	DefinitionID   string            `json:"definitionId"`
	Yaml           string            `json:"yaml"`
	Parameters     map[string]string `json:"parameters"`
	LiveStreamName string            `json:"liveStreamName"`
}
