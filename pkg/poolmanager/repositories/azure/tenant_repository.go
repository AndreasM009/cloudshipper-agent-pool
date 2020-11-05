package azure

import (
	"encoding/json"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/domain"
	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/repositories"
)

const (
	tenantTableName = "poolmngrtenants"
)

type tenantRepository struct {
}

// NewAzureTenantRepository new instance
func NewAzureTenantRepository() (repositories.TenantRepository, error) {
	client := GetStorageAccountInstance().GetClient()
	svc := client.GetTableService()
	tbl := svc.GetTableReference(tenantTableName)

	repo := &tenantRepository{}

	// table exists? if not, create it
	if err := tbl.Get(10, storage.FullMetadata); err != nil {
		if err := tbl.Create(10, storage.EmptyPayload, nil); err != nil {
			return nil, err
		}
	}
	return repo, nil
}

func (repo *tenantRepository) Add(tenant *domain.Tenant) (*domain.Tenant, error) {
	client := GetStorageAccountInstance().GetClient()
	svc := client.GetTableService()
	tbl := svc.GetTableReference(tenantTableName)

	ety := tbl.GetEntityReference(tenant.ID, tenant.ID)
	props, err := getTenantProperties(tenant)
	if err != nil {
		return nil, err
	}

	ety.Properties = props

	if err := ety.Insert(storage.EmptyPayload, nil); err != nil {
		return nil, err
	}

	return tenant, nil
}

func (repo *tenantRepository) Get(tenantID string) (*domain.Tenant, error) {
	client := GetStorageAccountInstance().GetClient()
	svc := client.GetTableService()
	tbl := svc.GetTableReference(tenantTableName)

	ety := tbl.GetEntityReference(tenantID, tenantID)

	if err := ety.Get(10, storage.FullMetadata, nil); err != nil {
		if svcerror, ok := err.(storage.AzureStorageServiceError); ok {
			if svcerror.StatusCode == http.StatusNotFound {
				// maybe tenant must be created, so try to add tenant with default settings
				tenant := domain.Tenant{
					ID:              tenantID,
					MaxParallelJobs: 1,
				}
				return repo.Add(&tenant)
			}
		}

		return nil, err
	}

	tenant, err := fromtenantProperties(ety.Properties)
	if err != nil {
		return nil, err
	}

	return tenant, nil
}

func (repo *tenantRepository) Update(tenant *domain.Tenant) (*domain.Tenant, error) {
	client := GetStorageAccountInstance().GetClient()
	svc := client.GetTableService()
	tbl := svc.GetTableReference(tenantTableName)

	ety := tbl.GetEntityReference(tenant.ID, tenant.ID)
	if err := ety.Get(10, storage.FullMetadata, nil); err != nil {
		return nil, err
	}

	props, err := getTenantProperties(tenant)
	if err != nil {
		return nil, err
	}

	ety.Properties = props

	if err := ety.Update(false, nil); err != nil {
		return nil, err
	}

	return tenant, nil
}

func getTenantProperties(tenant *domain.Tenant) (map[string]interface{}, error) {
	data, err := json.Marshal(tenant)
	if err != nil {
		return nil, err
	}

	props := map[string]interface{}{}
	if err := json.Unmarshal(data, &props); err != nil {
		return nil, err
	}

	return props, nil
}

func fromtenantProperties(props map[string]interface{}) (*domain.Tenant, error) {
	data, err := json.Marshal(props)
	if err != nil {
		return nil, err
	}

	tenant := &domain.Tenant{}
	if err := json.Unmarshal(data, tenant); err != nil {
		return nil, err
	}

	return tenant, nil
}
