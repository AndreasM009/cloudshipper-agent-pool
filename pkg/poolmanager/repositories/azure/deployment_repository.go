package azure

import (
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/domain"
	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/repositories"
)

const (
	deploymentTableName = "poolmngrdeployments"
)

type deploymentRepository struct {
}

// NewAzureDeploymentRepository new instance
func NewAzureDeploymentRepository() (repositories.DeploymentRepository, error) {
	client := GetStorageAccountInstance().GetClient()
	svc := client.GetTableService()
	tbl := svc.GetTableReference(deploymentTableName)

	repo := &deploymentRepository{}

	// table exists? if not, create it
	if err := tbl.Get(10, storage.FullMetadata); err != nil {
		if err := tbl.Create(10, storage.EmptyPayload, nil); err != nil {
			return nil, err
		}
	}
	return repo, nil
}

func (repo *deploymentRepository) Add(deployment *domain.Deployment) (*domain.Deployment, error) {
	client := GetStorageAccountInstance().GetClient()
	svc := client.GetTableService()
	tbl := svc.GetTableReference(deploymentTableName)

	ety := tbl.GetEntityReference(deployment.TenantID, deployment.DeploymentID)

	props, err := getProperties(deployment)
	if err != nil {
		return nil, err
	}

	ety.Properties = props

	if err := ety.Insert(storage.EmptyPayload, nil); err != nil {
		return nil, err
	}

	return deployment, nil
}

func (repo *deploymentRepository) GetAll(tenantID string) ([]*domain.Deployment, error) {
	client := GetStorageAccountInstance().GetClient()
	svc := client.GetTableService()
	tbl := svc.GetTableReference(deploymentTableName)

	query := storage.QueryOptions{
		Filter: fmt.Sprintf("PartitionKey eq '%s'", tenantID),
	}

	entities, err := tbl.QueryEntities(30, storage.FullMetadata, &query)
	if err != nil {
		return nil, err
	}

	if len(entities.Entities) <= 0 {
		return []*domain.Deployment{}, nil
	}

	deployments := make([]*domain.Deployment, len(entities.Entities))

	for i, e := range entities.Entities {
		deployment, err := fromProperties(e.Properties)
		if err != nil {
			return nil, err
		}
		deployments[i] = deployment
	}

	return deployments, nil
}

func (repo *deploymentRepository) Get(tenantID string, deploymentID string) (*domain.Deployment, error) {
	client := GetStorageAccountInstance().GetClient()
	svc := client.GetTableService()
	tbl := svc.GetTableReference(deploymentTableName)

	ety := tbl.GetEntityReference(tenantID, deploymentID)

	if err := ety.Get(10, storage.FullMetadata, nil); err != nil {
		return nil, err
	}

	deployment, err := fromProperties(ety.Properties)
	if err != nil {
		return nil, err
	}

	return deployment, nil
}

func (repo *deploymentRepository) Update(deployment *domain.Deployment) (*domain.Deployment, error) {
	client := GetStorageAccountInstance().GetClient()
	svc := client.GetTableService()
	tbl := svc.GetTableReference(deploymentTableName)

	ety := tbl.GetEntityReference(deployment.TenantID, deployment.DeploymentID)
	props, err := getProperties(deployment)
	if err != nil {
		return nil, err
	}

	ety.Properties = props

	if err := ety.Update(true, nil); err != nil {
		return nil, err
	}

	return deployment, nil
}

func (repo *deploymentRepository) Delete(tenantID string, deploymentID string) (*domain.Deployment, error) {
	client := GetStorageAccountInstance().GetClient()
	svc := client.GetTableService()
	tbl := svc.GetTableReference(deploymentTableName)

	ety := tbl.GetEntityReference(tenantID, deploymentID)

	if err := ety.Get(10, storage.FullMetadata, nil); err != nil {
		return nil, err
	}

	deployment, err := fromProperties(ety.Properties)
	if err != nil {
		return nil, err
	}

	if err := ety.Delete(false, nil); err != nil {
		return nil, err
	}

	return deployment, nil
}

func getProperties(deployment *domain.Deployment) (map[string]interface{}, error) {
	data, err := json.Marshal(deployment)
	if err != nil {
		return nil, err
	}

	props := map[string]interface{}{}
	if err := json.Unmarshal(data, &props); err != nil {
		return nil, err
	}

	return props, nil
}

func fromProperties(props map[string]interface{}) (*domain.Deployment, error) {
	data, err := json.Marshal(props)
	if err != nil {
		return nil, err
	}

	deployment := &domain.Deployment{}
	if err := json.Unmarshal(data, deployment); err != nil {
		return nil, err
	}

	return deployment, nil
}
