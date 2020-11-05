package inmemory

import (
	"sync"

	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/repositories"

	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/domain"
)

type inMemoryTenantRepository struct {
	tenants map[string]*domain.Tenant
	lock    sync.Mutex
}

// NewInMemoryTenantRepository new instance
func NewInMemoryTenantRepository() repositories.TenantRepository {
	return &inMemoryTenantRepository{
		tenants: make(map[string]*domain.Tenant),
	}
}

func (repo *inMemoryTenantRepository) Add(tenant *domain.Tenant) (*domain.Tenant, error) {
	repo.lock.Lock()
	defer repo.lock.Unlock()

	repo.tenants[tenant.ID] = tenant
	return nil, nil
}

func (repo *inMemoryTenantRepository) Get(tenantID string) (*domain.Tenant, error) {
	repo.lock.Lock()
	defer repo.lock.Unlock()

	if t, ok := repo.tenants[tenantID]; ok {
		return t, nil
	}

	t := &domain.Tenant{
		ID:              tenantID,
		MaxParallelJobs: 1,
	}

	repo.tenants[tenantID] = t
	return t, nil
}

func (repo *inMemoryTenantRepository) Update(tenant *domain.Tenant) (*domain.Tenant, error) {
	repo.lock.Lock()
	defer repo.lock.Unlock()

	repo.tenants[tenant.ID] = tenant
	return tenant, nil
}
