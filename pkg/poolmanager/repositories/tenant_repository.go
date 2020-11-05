package repositories

import "github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/domain"

// TenantRepository to access tenant settings and configurations
type TenantRepository interface {
	Add(tenant *domain.Tenant) (*domain.Tenant, error)
	Get(tenantID string) (*domain.Tenant, error)
	Update(tenant *domain.Tenant) (*domain.Tenant, error)
}
