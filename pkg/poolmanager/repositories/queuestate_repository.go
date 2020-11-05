package repositories

import (
	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/domain"
)

// QueueStateRepository to access QueueState of a tenant
type QueueStateRepository interface {
	Update(state *domain.QueueState) (*domain.QueueState, error)
	Get(tenantID string) (*domain.QueueState, error)
}
