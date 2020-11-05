package domain

// Tenant settings
type Tenant struct {
	ID              string `json:"id"`
	MaxParallelJobs int    `json:"maxParallelJobs"`
}
