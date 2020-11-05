package domain

// QueueState state of a Tenant's deployment queue
type QueueState struct {
	TenantID      string   `json:"tenantId"`
	RunningJobs   int      `json:"runningJobs"`
	ETag          string   `json:"etag"`
	DeploymentIDs []string `json:"deploymenIds"`
}
