package azure

import (
	"encoding/json"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/domain"
	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/repositories"
)

const (
	queueStateTableName = "poolmngrqueuestate"
)

type queueStateRepository struct {
}

// NewAzureQueueStateRepository new instance
func NewAzureQueueStateRepository() (repositories.QueueStateRepository, error) {
	client := GetStorageAccountInstance().GetClient()
	svc := client.GetTableService()
	tbl := svc.GetTableReference(queueStateTableName)

	repo := &queueStateRepository{}

	// table exists? if not, create it
	if err := tbl.Get(10, storage.FullMetadata); err != nil {
		if err := tbl.Create(10, storage.EmptyPayload, nil); err != nil {
			return nil, err
		}
	}
	return repo, nil
}

func (repo *queueStateRepository) Update(state *domain.QueueState) (*domain.QueueState, error) {
	client := GetStorageAccountInstance().GetClient()
	svc := client.GetTableService()
	tbl := svc.GetTableReference(queueStateTableName)

	ety := tbl.GetEntityReference(state.TenantID, state.TenantID)

	err := ety.Get(10, storage.FullMetadata, nil)

	if err := ety.Get(10, storage.FullMetadata, nil); err != nil {
		if svcerr, ok := err.(storage.AzureStorageServiceError); ok {
			if svcerr.StatusCode == http.StatusNotFound {
				// try to add state
				props, err := getStateProperties(state)
				if err != nil {
					return nil, err
				}

				ety = tbl.GetEntityReference(state.TenantID, state.TenantID)
				ety.Properties = props

				if err := ety.Insert(storage.FullMetadata, nil); err != nil {
					// try to load it again on error, maybe another instance has already added it
					ety = tbl.GetEntityReference(state.TenantID, state.TenantID)
					if err := ety.Get(10, storage.FullMetadata, nil); err != nil {
						return nil, err
					}
				}
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// update state
	props, err := getStateProperties(state)
	if err != nil {
		return nil, err
	}

	ety.Properties = props

	if err := ety.Update(false, nil); err != nil {
		if svcerr, ok := err.(storage.AzureStorageServiceError); ok {
			if svcerr.StatusCode == http.StatusPreconditionFailed {
				return nil, repositories.NewPreconditionFailedError(svcerr)
			}

			return nil, err
		}
	}
	return state, nil
}

func (repo *queueStateRepository) Get(tenantID string) (*domain.QueueState, error) {
	client := GetStorageAccountInstance().GetClient()
	svc := client.GetTableService()
	tbl := svc.GetTableReference(queueStateTableName)

	ety := tbl.GetEntityReference(tenantID, tenantID)

	if err := ety.Get(10, storage.FullMetadata, nil); err != nil {
		if svcerr, ok := err.(storage.AzureStorageServiceError); ok {
			if svcerr.StatusCode == http.StatusNotFound {
				// try to add default state
				state := &domain.QueueState{
					TenantID:      tenantID,
					RunningJobs:   0,
					ETag:          "",
					DeploymentIDs: []string{},
				}
				props, err := getStateProperties(state)
				if err != nil {
					return nil, err
				}

				ety = tbl.GetEntityReference(tenantID, tenantID)
				ety.Properties = props

				if err := ety.Insert(storage.FullMetadata, nil); err != nil {
					// try to load it again on error, maybe another instance has already added it
					ety = tbl.GetEntityReference(state.TenantID, state.TenantID)
					if err := ety.Get(10, storage.FullMetadata, nil); err != nil {
						return nil, err
					}
				}
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	state, err := fromStateProperties(ety.Properties)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func getStateProperties(state *domain.QueueState) (map[string]interface{}, error) {
	depidData, err := json.Marshal(state.DeploymentIDs)
	if err != nil {
		return nil, err
	}

	wrapper := struct {
		TenantID      string `json:"tenantId"`
		RunningJobs   int    `json:"runningJobs"`
		DeploymentIDs []byte `json:"deploymenIds"`
	}{
		TenantID:      state.TenantID,
		RunningJobs:   state.RunningJobs,
		DeploymentIDs: depidData,
	}

	data, err := json.Marshal(wrapper)
	if err != nil {
		return nil, err
	}

	props := map[string]interface{}{}
	if err := json.Unmarshal(data, &props); err != nil {
		return nil, err
	}

	return props, nil
}

func fromStateProperties(props map[string]interface{}) (*domain.QueueState, error) {
	data, err := json.Marshal(props)
	if err != nil {
		return nil, err
	}

	wrapper := struct {
		TenantID      string `json:"tenantId"`
		RunningJobs   int    `json:"runningJobs"`
		DeploymentIDs []byte `json:"deploymenIds"`
	}{}

	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}

	depids := []string{}
	err = json.Unmarshal(wrapper.DeploymentIDs, &depids)
	if err != nil {
		return nil, err
	}

	state := &domain.QueueState{
		TenantID:      wrapper.TenantID,
		RunningJobs:   wrapper.RunningJobs,
		ETag:          "",
		DeploymentIDs: depids,
	}
	return state, nil
}
