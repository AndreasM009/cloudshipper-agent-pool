package inmemory

import (
	"fmt"
	"sync"

	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/repositories"

	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/domain"
)

type inMemoryDeploymentRepository struct {
	deployments map[string][]*domain.Deployment
	lock        sync.Mutex
}

// NewInMemoryDeploymentRepository new instance
func NewInMemoryDeploymentRepository() repositories.DeploymentRepository {
	return &inMemoryDeploymentRepository{
		deployments: make(map[string][]*domain.Deployment),
	}
}

func (repo *inMemoryDeploymentRepository) Add(deployment *domain.Deployment) (*domain.Deployment, error) {
	repo.lock.Lock()
	defer repo.lock.Unlock()

	if val, ok := repo.deployments[deployment.TenantID]; ok {
		repo.deployments[deployment.TenantID] = append(val, deployment)
		return deployment, nil
	}

	d := make([]*domain.Deployment, 1)
	d[0] = deployment
	repo.deployments[deployment.TenantID] = d
	return deployment, nil
}

func (repo *inMemoryDeploymentRepository) GetAll(tenantID string) ([]*domain.Deployment, error) {
	repo.lock.Lock()
	defer repo.lock.Unlock()

	if val, ok := repo.deployments[tenantID]; ok {
		return val, nil
	}

	return nil, fmt.Errorf("no deployment for tenant %s found", tenantID)
}

func (repo *inMemoryDeploymentRepository) Get(tenantID string, deploymentID string) (*domain.Deployment, error) {
	repo.lock.Lock()
	defer repo.lock.Unlock()

	if val, ok := repo.deployments[tenantID]; ok {
		for _, d := range val {
			if d.TenantID == tenantID && d.DeploymentID == deploymentID {
				return d, nil
			}
		}
	}

	return nil, fmt.Errorf("No deployment with id %s and tenant %s found", deploymentID, tenantID)
}

func (repo *inMemoryDeploymentRepository) Update(deployment *domain.Deployment) (*domain.Deployment, error) {
	repo.lock.Lock()
	defer repo.lock.Unlock()

	tenantID := deployment.TenantID
	deploymentID := deployment.DeploymentID

	if val, ok := repo.deployments[tenantID]; ok {
		for i, d := range val {
			if d.TenantID == tenantID && d.DeploymentID == deploymentID {
				val[i] = deployment
				return d, nil
			}
		}
	}

	return nil, fmt.Errorf("No deployment with id %s found for tenent %s", deploymentID, tenantID)
}

func (repo *inMemoryDeploymentRepository) Delete(tenantID string, deploymentID string) (*domain.Deployment, error) {
	index := -1
	var deployments []*domain.Deployment

	if val, ok := repo.deployments[tenantID]; ok {
		deployments = val
		for i, d := range val {
			if d.TenantID == tenantID && d.DeploymentID == deploymentID {
				index = i
				break
			}
		}
	}

	if index == -1 {
		return nil, fmt.Errorf("No deployment with id %s for tenant %s found", deploymentID, tenantID)
	}

	res := deployments[index]
	d := append(deployments[:index], deployments[index+1:]...)
	repo.deployments[tenantID] = d
	return res, nil
}
