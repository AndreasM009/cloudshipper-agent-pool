package repositories

import (
	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/domain"
)

// DeploymentRepository to access tenant's queued deployments
type DeploymentRepository interface {
	Add(deployment *domain.Deployment) (*domain.Deployment, error)
	GetAll(tenantID string) ([]*domain.Deployment, error)
	Get(tenantID string, deploymentID string) (*domain.Deployment, error)
	Update(deployment *domain.Deployment) (*domain.Deployment, error)
	Delete(tenantID string, deploymentID string) (*domain.Deployment, error)
}
