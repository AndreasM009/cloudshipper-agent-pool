package services

import (
	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/domain"
)

// EnqueueDeploymentService interface for events to enqueue agent deployments
type EnqueueDeploymentService interface {
	Enqueue(deployment *domain.Deployment, yaml, parameters []byte) error
}
